#pragma once
#include <vte/vte.h>

typedef struct {
    int pid;
    int frontend_slave_fd;
    int backend_master_fd;
    char *err_msg;
} BtermProxySpawn;

extern void goVteChildExited(int termID, int status);
extern void goVteTitleChanged(int termID);

BtermProxySpawn vteSpawnProxy(VteTerminal *terminal, const char *workingDir,
                              char **argv, int columns, int rows);
void vteFreeError(char *errMsg);
int vteGetPtySize(int fd, int *columns, int *rows);
void vteSetPtySize(int fd, int columns, int rows);
void vteConnectChildExited(VteTerminal *terminal, int termID);
void vteConnectTitleChanged(VteTerminal *terminal, int termID);
void vteFeedChild(VteTerminal *terminal, const char *data, int len);
void vteSetFont(VteTerminal *terminal, const char *desc_str);
void vteSetScrollback(VteTerminal *terminal, long lines);
void vteSetColors(VteTerminal *terminal, const char *fg, const char *bg,
                  const char *cursor, const char **palette, int paletteLen);
const char *vteGetCurrentDirUri(VteTerminal *terminal);
char *vteGetWindowTitle(VteTerminal *terminal);
void vteCopyClipboard(VteTerminal *terminal);
void vtePasteClipboard(VteTerminal *terminal);
void vteSetSize(VteTerminal *terminal, int columns, int rows);
