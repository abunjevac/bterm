package ui

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/keymap"
)

type selectionTerminal interface {
	HasSelection() bool
}

type contextMenuEntry struct {
	label     string
	action    keymap.Action
	sensitive bool
}

func (pa *paneArea) installTerminalContextMenu(id int, widget gtk.Widgetter) {
	gesture := gtk.NewGestureClick()

	gesture.SetButton(3)
	gesture.ConnectPressed(func(_ int, x, y float64) {
		pa.tree.SetFocus(id)
		pa.grabFocus()
		pa.showContextMenu(id, widget, x, y)
	})

	gtk.BaseWidget(widget).AddController(gesture)
}

// showFocusedContextMenu opens the context menu for the focused pane,
// positioned at the center of the terminal widget.
func (pa *paneArea) showFocusedContextMenu() {
	id := pa.tree.Focused()

	if id == 0 {
		return
	}

	widget := pa.widgets[id]

	if widget == nil {
		return
	}

	w := gtk.BaseWidget(widget)

	pa.showContextMenu(id, widget, float64(w.Width())/2, float64(w.Height())/2)
}

func (pa *paneArea) showContextMenu(id int, widget gtk.Widgetter, x, y float64) {
	var openSubmenu *gtk.Popover

	t := pa.terms[id]

	if t == nil {
		return
	}

	popover := gtk.NewPopover()
	box := gtk.NewBox(gtk.OrientationVertical, 0)

	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.SetMarginStart(4)
	box.SetMarginEnd(4)
	box.Append(contextAction(pa.win, popover, "Copy", keymap.ActionCopy, hasSelection(t)))
	box.Append(contextAction(pa.win, popover, "Paste", keymap.ActionPaste, true))
	box.Append(gtk.NewSeparator(gtk.OrientationHorizontal))
	box.Append(contextAction(pa.win, popover, "Clear", keymap.ActionClear, true))
	box.Append(contextAction(pa.win, popover, "Reset", keymap.ActionReset, true))
	box.Append(gtk.NewSeparator(gtk.OrientationHorizontal))
	box.Append(contextSubmenu(pa.win, popover, &openSubmenu, "Open", []contextMenuEntry{
		{label: "Editor", action: keymap.ActionOpenEditor, sensitive: true},
		{label: "File Browser", action: keymap.ActionOpenFileBrowser, sensitive: true},
	}))
	box.Append(gtk.NewSeparator(gtk.OrientationHorizontal))
	box.Append(contextSubmenu(pa.win, popover, &openSubmenu, "Split", []contextMenuEntry{
		{label: "Split Right", action: keymap.ActionSplitLeftRight, sensitive: true},
		{label: "Split Down", action: keymap.ActionSplitTopBottom, sensitive: true},
		{label: "Close Pane", action: keymap.ActionClosePane, sensitive: true},
	}))
	box.Append(contextSubmenu(pa.win, popover, &openSubmenu, "Tab", []contextMenuEntry{
		{label: "New Tab", action: keymap.ActionNewTabEnd, sensitive: true},
		{label: "New Tab After", action: keymap.ActionNewTabAfter, sensitive: true},
		{label: "Close Tab", action: keymap.ActionCloseTab, sensitive: true},
	}))
	box.Append(contextSubmenu(pa.win, popover, &openSubmenu, "Window", []contextMenuEntry{
		{label: "New Window", action: keymap.ActionNewWindow, sensitive: true},
		{label: "Close Window", action: keymap.ActionCloseWindow, sensitive: true},
		{label: "Preferences", action: keymap.ActionOpenConfig, sensitive: true},
	}))

	rect := gdk.NewRectangle(int(x), int(y), 1, 1)

	popover.SetChild(box)
	popover.SetParent(widget)
	popover.SetPointingTo(&rect)
	popover.Popup()
}

func hasSelection(t any) bool {
	selected, ok := t.(selectionTerminal)
	if !ok {
		return true
	}

	return selected.HasSelection()
}

func contextAction(w *window, root *gtk.Popover, label string, action keymap.Action, sensitive bool) gtk.Widgetter {
	row := contextRow(label, formatBinding(w.keys.BindingFor(action)), false)

	row.SetSensitive(sensitive)

	gesture := gtk.NewGestureClick()

	gesture.SetButton(0)
	gesture.ConnectReleased(func(_ int, _, _ float64) {
		root.Popdown()
		w.dispatch(action)
	})

	row.AddController(gesture)

	return row
}

func contextSubmenu(w *window, root *gtk.Popover, openSubmenu **gtk.Popover, label string, entries []contextMenuEntry) gtk.Widgetter {
	row := contextRow(label, "", true)
	box := gtk.NewBox(gtk.OrientationVertical, 0)

	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.SetMarginStart(4)
	box.SetMarginEnd(4)

	submenu := gtk.NewPopover()

	for _, entry := range entries {
		box.Append(contextSubmenuAction(w, root, submenu, entry))
	}

	submenu.SetChild(box)
	submenu.SetParent(row)
	submenu.SetPosition(gtk.PosRight)

	installSubmenuOpenControllers(row, submenu, openSubmenu)

	return row
}

func installSubmenuOpenControllers(row *gtk.Box, submenu *gtk.Popover, openSubmenu **gtk.Popover) {
	motion := gtk.NewEventControllerMotion()

	motion.ConnectEnter(func(_, _ float64) {
		openContextSubmenu(submenu, openSubmenu)
	})
	row.AddController(motion)

	gesture := gtk.NewGestureClick()

	gesture.SetButton(0)
	gesture.ConnectReleased(func(_ int, _, _ float64) {
		openContextSubmenu(submenu, openSubmenu)
	})
	row.AddController(gesture)
}

func openContextSubmenu(submenu *gtk.Popover, openSubmenu **gtk.Popover) {
	if *openSubmenu != nil && *openSubmenu != submenu {
		(*openSubmenu).Popdown()
	}

	*openSubmenu = submenu

	submenu.Popup()
}

func contextSubmenuAction(w *window, root, submenu *gtk.Popover, entry contextMenuEntry) gtk.Widgetter {
	row := contextRow(entry.label, formatBinding(w.keys.BindingFor(entry.action)), false)

	row.SetSensitive(entry.sensitive)

	gesture := gtk.NewGestureClick()

	gesture.SetButton(0)
	gesture.ConnectReleased(func(_ int, _, _ float64) {
		submenu.Popdown()
		root.Popdown()
		w.dispatch(entry.action)
	})

	row.AddController(gesture)

	return row
}

func contextRow(label, shortcut string, submenu bool) *gtk.Box {
	row := gtk.NewBox(gtk.OrientationHorizontal, 12)
	name := gtk.NewLabel(label)

	row.AddCSSClass("bterm-context-row")
	row.SetMarginTop(2)
	row.SetMarginBottom(2)
	row.SetMarginStart(8)
	row.SetMarginEnd(8)
	name.SetXAlign(0)
	name.SetHExpand(true)
	row.Append(name)

	if shortcut != "" {
		shortcutLabel := gtk.NewLabel(shortcut)

		shortcutLabel.AddCSSClass("bterm-menu-shortcut")
		row.Append(shortcutLabel)
	}

	if submenu {
		arrow := gtk.NewImageFromIconName("go-next-symbolic")

		row.Append(arrow)
	}

	return row
}
