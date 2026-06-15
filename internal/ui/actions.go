package ui

import (
	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/ui/panetree"
)

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
	case keymap.ActionNewTabEnd:
		w.newTabEnd()
	case keymap.ActionNewTabAfter:
		w.newTabAfter()
	case keymap.ActionCloseTab:
		if len(w.tabs) > 0 {
			w.closeTab(w.tabs[w.active])
		}
	case keymap.ActionFontInc, keymap.ActionFontDec, keymap.ActionFontReset:
		w.applyFontAction(a)
	case keymap.ActionOpenConfig:
		showConfigDialog(w.win, w)
	default:
		return w.dispatchTabSelect(a)
	}

	return true
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
func (w *window) dispatchPane(pa *paneArea, a keymap.Action) {
	switch a {
	case keymap.ActionSplitLeftRight:
		pa.split(panetree.LeftRight)
	case keymap.ActionSplitTopBottom:
		pa.split(panetree.TopBottom)
	case keymap.ActionClosePane:
		pa.closeFocused()
	case keymap.ActionCopy:
		pa.copyFromFocused()
		w.toast.show("⧉ Copied")
	case keymap.ActionPaste:
		pa.pasteToFocused()
		w.toast.show("⧉ Pasted")
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
func (pa *paneArea) copyFromFocused() {
	if t := pa.focusedTerminal(); t != nil {
		t.Copy()
	}
}

// pasteToFocused pastes clipboard contents into the focused terminal.
func (pa *paneArea) pasteToFocused() {
	if t := pa.focusedTerminal(); t != nil {
		t.Paste()
	}
}

// current returns the active tab's paneArea, or nil when there are no tabs.
func (w *window) current() *paneArea {
	if len(w.tabs) == 0 {
		return nil
	}

	return w.tabs[w.active].area
}
