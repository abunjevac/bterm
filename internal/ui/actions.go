package ui

import (
	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/ui/panetree"
)

// dispatch executes the action a, routing tab-level actions before pane actions.
func (w *window) dispatch(a keymap.Action) {
	switch a {
	case keymap.ActionNewTabEnd:
		w.newTabEnd()

		return
	case keymap.ActionNewTabAfter:
		w.newTabAfter()

		return
	case keymap.ActionCloseTab:
		if len(w.tabs) > 0 {
			w.closeTab(w.tabs[w.active])
		}

		return
	}

	if a >= keymap.ActionTab1 && a <= keymap.ActionTab9 {
		w.selectTab(int(a - keymap.ActionTab1))

		return
	}

	pa := w.current()

	if pa == nil {
		return
	}

	switch a {
	case keymap.ActionSplitLeftRight:
		pa.split(panetree.LeftRight)
	case keymap.ActionSplitTopBottom:
		pa.split(panetree.TopBottom)
	case keymap.ActionClosePane:
		pa.closeFocused()
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
	case keymap.ActionCopy:
		if t := pa.focusedTerminal(); t != nil {
			t.Copy()
		}
	case keymap.ActionPaste:
		if t := pa.focusedTerminal(); t != nil {
			t.Paste()
		}
	}
}

// current returns the active tab's paneArea, or nil when there are no tabs.
func (w *window) current() *paneArea {
	if len(w.tabs) == 0 {
		return nil
	}

	return w.tabs[w.active].area
}
