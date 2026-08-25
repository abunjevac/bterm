package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/ui/panetree"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
)

// bracketedNewline is a literal LF wrapped in bracketed-paste markers
// (xterm mode 2004: ESC[200~ ... ESC[201~). Apps that enable bracketed paste
// treat everything between the markers as a single atomic block of literal
// text rather than keystrokes, so the embedded LF is inserted rather than
// submitting the line — unlike a bare LF, which some apps (e.g. Codebuff)
// treat the same as a plain Enter outside of a paste block.
var bracketedNewline = []byte("\x1b[200~\n\x1b[201~") //nolint:gochecknoglobals

type clearTerminal interface {
	Clear()
}

type resetTerminal interface {
	Reset()
}

// dispatch routes the action: window-level first, then active pane area.
func (w *window) dispatch(a keymap.Action) {
	if w.dispatchWindow(a) {
		return
	}

	if pa := w.current(); pa != nil {
		w.dispatchPane(pa, a)
	}
}

// dispatchWindow handles tab and font actions. Returns true when the action
// was consumed so dispatch does not fall through to pane handling.
func (w *window) dispatchWindow(a keymap.Action) bool {
	switch a {
	case keymap.ActionNewTabEnd, keymap.ActionNewTabAfter:
		w.newTab(a)
	case keymap.ActionCloseTab:
		if len(w.tabs) > 0 {
			w.closeTab(w.tabs[w.active])
		}
	case keymap.ActionFontInc, keymap.ActionFontDec, keymap.ActionFontReset:
		w.applyFontAction(a)
	case keymap.ActionOpenConfig:
		showConfigDialog(w.win, w)
	case keymap.ActionNewWindow:
		w.openNewWindow()
	case keymap.ActionCloseWindow:
		w.win.Close()
	case keymap.ActionOpenEditor, keymap.ActionOpenFileBrowser:
		w.openConfiguredApplication(a)
	default:
		return w.dispatchTabSelect(a)
	}

	return true
}

func (w *window) newTab(action keymap.Action) {
	if action == keymap.ActionNewTabEnd {
		w.newTabEnd()

		return
	}

	w.newTabAfter()
}

func (w *window) openConfiguredApplication(action keymap.Action) {
	if action == keymap.ActionOpenEditor {
		w.openApplication("Opening editor...", w.bundle.Config.Editor, w.bundle.Config.EditorArgs)

		return
	}

	w.openApplication("Opening file browser...", w.bundle.Config.FileBrowser, w.bundle.Config.FileBrowserArgs)
}

// openApplication starts a configured application with the active terminal directory.
func (w *window) openApplication(message, command string, args []string) {
	command = strings.TrimSpace(command)
	if command == "" {
		w.toast.show("Application command is not configured")

		return
	}

	cwd := w.activeCWD()
	if cwd == "" {
		w.toast.show("Current directory is unavailable")

		return
	}

	expandedArgs := expandCWDArgs(args, cwd)

	w.toast.show(message)

	if err := exec.CommandContext(context.Background(), command, expandedArgs...).Start(); err != nil {
		w.toast.show("Could not open application")

		_, _ = fmt.Fprintf(os.Stderr, "bterm: launch %s: %v\n", command, err)
	}
}

func expandCWDArgs(args []string, cwd string) []string {
	expanded := make([]string, len(args))

	for i, arg := range args {
		expanded[i] = strings.ReplaceAll(arg, "{cwd}", cwd)
	}

	return expanded
}

// openNewWindow opens another application window using the active terminal directory.
func (w *window) openNewWindow() {
	win := newWindow(w.app, w.bundle, w.activeCWD())

	win.Present()
}

// dispatchTabSelect activates a numbered tab (Tab1–Tab9). Returns true when consumed.
func (w *window) dispatchTabSelect(a keymap.Action) bool {
	if a >= keymap.ActionTab1 && a <= keymap.ActionTab9 {
		w.selectTab(int(a - keymap.ActionTab1))

		return true
	}

	return false
}

// applyFontAction adjusts fontSize for zoom/reset actions and applies the new
// value to the currently focused terminal. New panes/tabs pick it up via spawnTerm.
func (w *window) applyFontAction(a keymap.Action) {
	switch a {
	case keymap.ActionFontInc:
		w.fontSize++
	case keymap.ActionFontDec:
		if w.fontSize > 4 {
			w.fontSize--
		}
	case keymap.ActionFontReset:
		w.fontSize = w.defaultFontSize
	default:
	}

	if pa := w.current(); pa != nil {
		if t := pa.focusedTerminal(); t != nil {
			t.SetFont(w.fontFamily, w.fontSize)
		}
	}
}

// dispatchPane routes pane-level actions to the appropriate paneArea method.
func (w *window) dispatchPane(pa *paneArea, a keymap.Action) { //nolint:cyclop
	switch a {
	case keymap.ActionSplitLeftRight:
		pa.split(panetree.LeftRight)
	case keymap.ActionSplitTopBottom:
		pa.split(panetree.TopBottom)
	case keymap.ActionClosePane:
		pa.closeFocused()
	case keymap.ActionCopy:
		if pa.copyFromFocused() {
			w.reownClipboard()
			w.toast.show("⧉ Copied")
		}
	case keymap.ActionPaste:
		if pa.pasteToFocused() {
			w.toast.show("⧉ Pasted")
		}
	case keymap.ActionClear:
		pa.clearFocused()
	case keymap.ActionReset:
		pa.resetFocused()
	case keymap.ActionSendNewline:
		pa.sendNewlineToFocused()
	case keymap.ActionSendNewlinePlain:
		pa.sendNewlinePlainToFocused()
	case keymap.ActionShowContextMenu:
		pa.showFocusedContextMenu()
	default:
		pa.dispatchDir(a)
	}
}

// dispatchDir routes focus and resize actions by direction.
func (pa *paneArea) dispatchDir(a keymap.Action) {
	switch a {
	case keymap.ActionFocusLeft:
		pa.focusDir(panetree.DirLeft)
	case keymap.ActionFocusRight:
		pa.focusDir(panetree.DirRight)
	case keymap.ActionFocusUp:
		pa.focusDir(panetree.DirUp)
	case keymap.ActionFocusDown:
		pa.focusDir(panetree.DirDown)
	case keymap.ActionResizeLeft:
		pa.resizeFocused(panetree.DirLeft)
	case keymap.ActionResizeRight:
		pa.resizeFocused(panetree.DirRight)
	case keymap.ActionResizeUp:
		pa.resizeFocused(panetree.DirUp)
	case keymap.ActionResizeDown:
		pa.resizeFocused(panetree.DirDown)
	default:
	}
}

// copyFromFocused copies the selection from the focused terminal.
// Returns false when there is no selection to copy.
func (pa *paneArea) copyFromFocused() bool {
	t := pa.focusedTerminal()
	if t == nil {
		return false
	}

	st, ok := t.(selectionTerminal)
	if !ok || !st.HasSelection() {
		return false
	}

	t.Copy()

	return true
}

// pasteToFocused pastes clipboard contents into the focused terminal.
// Returns false when the clipboard has no text content to paste.
func (pa *paneArea) pasteToFocused() bool {
	t := pa.focusedTerminal()
	if t == nil {
		return false
	}

	clip := pa.win.clipboard()
	if clip == nil {
		return false
	}

	// Check if the clipboard currently offers any text format before pasting.
	// Different applications report text with different MIME types
	// (text/plain, text/plain;charset=utf-8, UTF8_STRING, etc.).
	if !clipboardHasText(clip) {
		return false
	}

	t.Paste()

	return true
}

// clipboardHasText reports whether the clipboard offers any text-like format.
func clipboardHasText(clip *gdk.Clipboard) bool {
	for _, mime := range clip.Formats().MIMETypes() {
		if strings.HasPrefix(mime, "text/") || mime == "UTF8_STRING" {
			return true
		}
	}

	return false
}

// clearFocused clears the focused terminal's screen and scrollback.
func (pa *paneArea) clearFocused() {
	if t, ok := pa.focusedTerminal().(clearTerminal); ok {
		t.Clear()
	}
}

// resetFocused resets the focused terminal and clears scrollback.
func (pa *paneArea) resetFocused() {
	if t, ok := pa.focusedTerminal().(resetTerminal); ok {
		t.Reset()
	}
}

// sendNewlineToFocused feeds a bracketed-paste-wrapped newline to the focused
// terminal's child. Bound to Ctrl+Enter, a fallback for apps (e.g. Codebuff)
// that treat a bare LF the same as Enter outside of a bracketed-paste block.
func (pa *paneArea) sendNewlineToFocused() {
	if t := pa.focusedTerminal(); t != nil {
		t.FeedChild(bracketedNewline)
	}
}

// sendNewlinePlainToFocused feeds a bare LF to the focused terminal's child,
// with no bracketed-paste wrapping. Bound to Shift+Enter, the binding most
// apps (Claude Code, Codex, Copilot CLI) expect for "insert newline" instead
// of an Enter keystroke.
func (pa *paneArea) sendNewlinePlainToFocused() {
	if t := pa.focusedTerminal(); t != nil {
		t.FeedChild([]byte{'\n'})
	}
}

// current returns the active tab's paneArea, or nil when there are no tabs.
func (w *window) current() *paneArea {
	if len(w.tabs) == 0 {
		return nil
	}

	return w.tabs[w.active].area
}
