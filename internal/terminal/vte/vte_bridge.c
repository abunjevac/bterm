#include "vte.h"
#include <errno.h>
#include <fcntl.h>
#include <pango/pango.h>
#include <pty.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/wait.h>
#include <termios.h>
#include <unistd.h>

/* --- proxy spawn --- */

static void close_if_valid(int fd) {
    if (fd >= 0) {
        close(fd);
    }
}

static char *dup_errno_message(const char *prefix) {
    const char *msg = strerror(errno);
    size_t len = strlen(prefix) + strlen(msg) + 3;
    char *result = malloc(len);
    if (result == NULL) { return NULL; }

    snprintf(result, len, "%s: %s", prefix, msg);
    return result;
}

static void set_raw(int fd) {
    struct termios tio;
    if (tcgetattr(fd, &tio) != 0) { return; }

    cfmakeraw(&tio);
    tcsetattr(fd, TCSANOW, &tio);
}

BtermProxySpawn vteSpawnProxy(VteTerminal *terminal, const char *workingDir,
                              char **argv, int columns, int rows) {
    BtermProxySpawn result = { .pid = -1, .frontend_slave_fd = -1, .backend_master_fd = -1, .err_msg = NULL };

    int frontend_master = -1;
    int frontend_slave = -1;

    struct winsize ws = {0};
    ws.ws_col = columns > 0 ? (unsigned short)columns : 80;
    ws.ws_row = rows > 0 ? (unsigned short)rows : 24;

    if (openpty(&frontend_master, &frontend_slave, NULL, NULL, &ws) != 0) {
        result.err_msg = dup_errno_message("open frontend pty");
        return result;
    }

    set_raw(frontend_slave);

    GError *error = NULL;
    VtePty *frontend_pty = vte_pty_new_foreign_sync(frontend_master, NULL, &error);
    if (frontend_pty == NULL) {
        result.err_msg = error != NULL ? strdup(error->message) : strdup("create VTE pty");
        if (error != NULL) { g_error_free(error); }
        close_if_valid(frontend_master);
        close_if_valid(frontend_slave);
        return result;
    }

    vte_terminal_set_pty(terminal, frontend_pty);
    g_object_unref(frontend_pty);

    int backend_master = -1;
    pid_t pid = forkpty(&backend_master, NULL, NULL, &ws);
    if (pid < 0) {
        result.err_msg = dup_errno_message("fork pty");
        close_if_valid(frontend_slave);
        return result;
    }

    if (pid == 0) {
        if (workingDir != NULL && workingDir[0] != '\0') {
            if (chdir(workingDir) != 0) { _exit(127); }
        }

        execvp(argv[0], argv);
        _exit(127);
    }

    vte_terminal_watch_child(terminal, (GPid)pid);

    result.pid = (int)pid;
    result.frontend_slave_fd = frontend_slave;
    result.backend_master_fd = backend_master;

    return result;
}

void vteFreeError(char *errMsg) {
    free(errMsg);
}

int vteGetPtySize(int fd, int *columns, int *rows) {
    if (fd < 0) { return 0; }

    struct winsize ws = {0};
    if (ioctl(fd, TIOCGWINSZ, &ws) != 0) { return 0; }

    if (columns != NULL) { *columns = ws.ws_col; }
    if (rows != NULL) { *rows = ws.ws_row; }

    return 1;
}

void vteSetPtySize(int fd, int columns, int rows) {
    if (fd < 0) { return; }

    struct winsize ws = {0};
    ws.ws_col = columns > 0 ? (unsigned short)columns : 80;
    ws.ws_row = rows > 0 ? (unsigned short)rows : 24;

    ioctl(fd, TIOCSWINSZ, &ws);
}

/* --- child-exited --- */

static void vteChildExitedBridge(VteTerminal *terminal, int status, gpointer user_data) {
    (void)terminal;
    goVteChildExited((int)(intptr_t)user_data, status);
}

void vteConnectChildExited(VteTerminal *terminal, int termID) {
    g_signal_connect(G_OBJECT(terminal), "child-exited",
                     G_CALLBACK(vteChildExitedBridge),
                     (gpointer)(intptr_t)termID);
}

/* --- title-changed --- */

static void on_title_changed(VteTerminal *t, gpointer user_data) {
    (void)t;
    goVteTitleChanged((int)(intptr_t)user_data);
}

void vteConnectTitleChanged(VteTerminal *terminal, int termID) {
    g_signal_connect(terminal, "window-title-changed",
                     G_CALLBACK(on_title_changed), (gpointer)(intptr_t)termID);
}

/* --- feed / font --- */

void vteFeedChild(VteTerminal *terminal, const char *data, int len) {
    vte_terminal_feed_child(terminal, data, (gssize)len);
}

void vteSetFont(VteTerminal *terminal, const char *desc_str) {
    PangoFontDescription *desc = pango_font_description_from_string(desc_str);
    if (desc == NULL) { return; }
    vte_terminal_set_font(terminal, desc);
    pango_font_description_free(desc);
}

/* --- scrollback --- */

void vteSetScrollback(VteTerminal *terminal, long lines) {
    vte_terminal_set_scrollback_lines(terminal, lines);
}

/* --- colors --- */

static void parse_rgba(const char *hex, GdkRGBA *out) {
    gdk_rgba_parse(out, hex);
}

void vteSetColors(VteTerminal *terminal, const char *fg, const char *bg,
                  const char *cursor, const char **palette, int paletteLen) {
    GdkRGBA fgc, bgc, curc;

    parse_rgba(fg, &fgc);
    parse_rgba(bg, &bgc);
    parse_rgba(cursor, &curc);

    GdkRGBA pal[16];
    for (int i = 0; i < paletteLen && i < 16; i++) {
        parse_rgba(palette[i], &pal[i]);
    }

    vte_terminal_set_colors(terminal, &fgc, &bgc, pal, paletteLen);
    vte_terminal_set_color_cursor(terminal, &curc);
}

/* --- directory uri --- */

const char *vteGetCurrentDirUri(VteTerminal *terminal) {
    return vte_terminal_get_current_directory_uri(terminal);
}

/* --- window title --- */

char *vteGetWindowTitle(VteTerminal *terminal) {
    const char *t = vte_terminal_get_window_title(terminal);
    return t ? g_strdup(t) : NULL;
}

/* --- clipboard --- */

void vteCopyClipboard(VteTerminal *terminal) {
    vte_terminal_copy_clipboard_format(terminal, VTE_FORMAT_TEXT);
}

void vtePasteClipboard(VteTerminal *terminal) {
    vte_terminal_paste_clipboard(terminal);
}

/* --- size --- */

void vteSetSize(VteTerminal *terminal, int columns, int rows) {
    vte_terminal_set_size(terminal, (glong)columns, (glong)rows);
}
