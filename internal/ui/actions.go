package ui

import (
	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/ui/panetree"
)

// dispatch routes the action to the tab layer or, for pane-level actions,
// to the active pane area.
func (w *window) dispatch(a keymap.Action) {
	switch a {
	case keymap.ActionNewTabEnd:
		w.newTabEnd()
	case keymap.ActionNewTabAfter:
		w.newTabAfter()
	case keymap.ActionCloseTab:
		if len(w.tabs) > 0 {
			w.closeTab(w.tabs[w.active])
		}
	default:
		if a >= keymap.ActionTab1 && a <= keymap.ActionTab9 {
			w.selectTab(int(a - keymap.ActionTab1))

			return
		}

		if pa := w.current(); pa != nil {
			w.dispatchPane(pa, a)
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
