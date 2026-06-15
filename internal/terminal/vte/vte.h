#pragma once
#include <vte/vte.h>

extern void goVteSpawnDone(int callbackID, int pid, int ptyFd, char *errMsg);
extern void goVteChildExited(int termID, int status);
extern void goVteTitleChanged(int termID);

void vteSpawnAsync(VteTerminal *terminal, const char *workingDir, char **argv, int callbackID);
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
