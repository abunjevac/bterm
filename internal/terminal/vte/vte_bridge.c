#include "vte.h"
#include <pango/pango.h>
#include <stdlib.h>
#include <string.h>

/* --- spawn --- */

static void vteSpawnBridge(VteTerminal *terminal, GPid pid, GError *error, gpointer user_data) {
    int callbackID = (int)(intptr_t)user_data;
    int ptyFd = -1;
    char *errMsg = NULL;

    if (error != NULL) {
        errMsg = error->message;
    } else {
        VtePty *pty = vte_terminal_get_pty(terminal);
        if (pty != NULL) {
            ptyFd = vte_pty_get_fd(pty);
        }
    }

    goVteSpawnDone(callbackID, (int)pid, ptyFd, errMsg);
}

void vteSpawnAsync(VteTerminal *terminal, const char *workingDir,
                   char **argv, int callbackID) {
    vte_terminal_spawn_async(
        terminal, VTE_PTY_DEFAULT, workingDir, argv,
        NULL, G_SPAWN_SEARCH_PATH,
        NULL, NULL, NULL, -1, NULL,
        vteSpawnBridge, (gpointer)(intptr_t)callbackID
    );
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
