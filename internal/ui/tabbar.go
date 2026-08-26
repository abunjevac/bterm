package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/keymap"
)

// buildTabBar creates the GtkHeaderBar with a tab-label box and installs a
// GtkStack as the window child. Must be called once before any tabs are added.
func (w *window) buildTabBar() {
	w.tabBox = gtk.NewBox(gtk.OrientationHorizontal, 2)

	header := gtk.NewHeaderBar()

	header.SetShowTitleButtons(true)

	gtk.IconThemeGetForDisplay(gdk.DisplayGetDefault()).AddResourcePath("/io/github/abunjevac/bterm/icons")

	settingsBtn := gtk.NewButton()

	settingsBtn.SetIconName("preferences-system-symbolic")
	settingsBtn.SetTooltipText("Preferences")
	settingsBtn.AddCSSClass("flat")
	settingsBtn.ConnectClicked(func() { showConfigDialog(w.win, w) })

	header.PackStart(settingsBtn)

	splitLeftRightBtn := gtk.NewButton()

	splitLeftRightBtn.SetIconName("bterm-split-left-right-symbolic")
	splitLeftRightBtn.SetTooltipText(headerActionTooltip(w.keys, keymap.ActionSplitLeftRight, "Split left/right"))
	splitLeftRightBtn.AddCSSClass("flat")
	splitLeftRightBtn.ConnectClicked(func() { w.dispatch(keymap.ActionSplitLeftRight) })

	header.PackStart(splitLeftRightBtn)

	splitTopBottomBtn := gtk.NewButton()

	splitTopBottomBtn.SetIconName("bterm-split-top-bottom-symbolic")
	splitTopBottomBtn.SetTooltipText(headerActionTooltip(w.keys, keymap.ActionSplitTopBottom, "Split top/bottom"))
	splitTopBottomBtn.AddCSSClass("flat")
	splitTopBottomBtn.ConnectClicked(func() { w.dispatch(keymap.ActionSplitTopBottom) })

	header.PackStart(splitTopBottomBtn)

	addBtn := gtk.NewButton()

	addBtn.SetIconName("list-add-symbolic")
	addBtn.SetTooltipText(headerActionTooltip(w.keys, keymap.ActionNewTabEnd, "New tab"))
	addBtn.AddCSSClass("flat")
	addBtn.ConnectClicked(func() { w.newTabEnd() })

	header.PackStart(addBtn)
	header.PackStart(w.tabBox)

	w.memMon = newMemMonitor()
	w.uptimeMon = newUptimeMonitor()
	header.PackEnd(w.buildMenuButton())
	header.PackEnd(w.memMon.label)
	header.PackEnd(gtk.NewLabel("·"))
	header.PackEnd(w.uptimeMon.label)

	w.win.SetTitlebar(header)

	w.stack = gtk.NewStack()

	w.stack.SetVExpand(true)
	w.stack.SetHExpand(true)
	w.stack.SetTransitionType(gtk.StackTransitionTypeNone)
}

// addTab creates a new tab with the given cwd, appends it to the tab list.
// It does not select it; call selectTab separately.
func (w *window) addTab(cwd string) {
	t := &tab{}

	term := w.newTerm()

	w.spawnTerm(term, cwd)

	t.area = newPaneArea(w, term)

	w.attachTab(t)
}

func (w *window) attachTab(t *tab) {
	w.bindTab(t)

	w.stack.AddChild(t.area.root)
	w.tabBox.Append(t.label)

	w.tabs = append(w.tabs, t)
}

func (w *window) bindTab(t *tab) {
	t.area.win = w

	t.area.onEmpty = func() { w.closeTab(t) }

	t.buildLabel(w, len(w.tabs)+1)

	t.area.onTitleChanged = func(title string) {
		t.title = title

		if t.titleLabel != nil {
			t.titleLabel.SetText(title)
		}

		if len(w.tabs) > 0 && w.tabs[w.active] == t {
			w.win.SetTitle(title)
		}
	}
}

func (w *window) detachTab(t *tab) {
	idx := w.tabIndex(t)

	if idx < 0 || len(w.tabs) < 2 {
		return
	}

	w.stack.Remove(t.area.root)
	w.tabBox.Remove(t.label)

	w.tabs = slices.Delete(w.tabs, idx, idx+1)

	if w.active > idx {
		w.active--
	} else if w.active >= len(w.tabs) {
		w.active = len(w.tabs) - 1
	}

	w.renumber()
	w.selectTab(w.active)

	newWin := newEmptyWindow(w.app, w.bundle, w.workingDir)

	newWin.attachTab(t)
	newWin.renumber()
	newWin.selectTab(0)

	newWin.win.Present()
}

// newTabEnd opens a new tab at the end of the tab list and selects it.
func (w *window) newTabEnd() {
	w.addTab(w.activeCWD())
	w.renumber()
	w.selectTab(len(w.tabs) - 1)
}

// newTabAfter opens a new tab immediately after the active tab and selects it.
func (w *window) newTabAfter() {
	w.addTab(w.activeCWD())

	insertIdx := w.active + 1
	endIdx := len(w.tabs) - 1

	if insertIdx < endIdx {
		t := w.tabs[endIdx]

		copy(w.tabs[insertIdx+1:], w.tabs[insertIdx:endIdx])

		w.tabs[insertIdx] = t

		w.tabBox.ReorderChildAfter(t.label, w.tabs[w.active].label)
	}

	w.renumber()
	w.selectTab(insertIdx)
}

// selectTab makes tab i visible and transfers keyboard focus to its pane area.
func (w *window) selectTab(i int) {
	if i < 0 || i >= len(w.tabs) {
		return
	}

	if w.active < len(w.tabs) {
		w.tabs[w.active].label.RemoveCSSClass("bterm-tab-active")
	}

	w.active = i

	t := w.tabs[i]

	t.label.AddCSSClass("bterm-tab-active")
	w.stack.SetVisibleChild(t.area.root)

	title := t.title

	if title == "" {
		title = w.bundle.Config.Title
	}

	w.win.SetTitle(title)
	t.area.grabFocus()
}

// closeTab removes t from the tab list. Closes the window when it was the last tab.
func (w *window) closeTab(t *tab) {
	idx := w.tabIndex(t)

	if idx < 0 {
		return
	}

	w.stack.Remove(t.area.root)
	w.tabBox.Remove(t.label)

	w.tabs = slices.Delete(w.tabs, idx, idx+1)

	if len(w.tabs) == 0 {
		w.win.Close()

		return
	}

	if w.active > idx {
		w.active--
	} else if w.active >= len(w.tabs) {
		w.active = len(w.tabs) - 1
	}

	w.renumber()
	w.selectTab(w.active)
}

// swapAdjacentTabs swaps the tabs at indices i and i+1, keeping w.tabs,
// w.tabBox child order, w.active and the number badges in sync.
func (w *window) swapAdjacentTabs(i int) {
	j := i + 1

	w.tabs[i], w.tabs[j] = w.tabs[j], w.tabs[i]

	w.tabBox.ReorderChildAfter(w.tabs[j].label, w.tabs[i].label)

	switch w.active {
	case i:
		w.active = j
	case j:
		w.active = i
	}

	w.renumber()
}

// tabIndex returns the position of t in w.tabs, or -1 if not found.
func (w *window) tabIndex(t *tab) int {
	for i, tab := range w.tabs {
		if tab == t {
			return i
		}
	}

	return -1
}

// renumber refreshes the number badge on every tab label.
func (w *window) renumber() {
	for i, t := range w.tabs {
		t.numLabel.SetText(fmt.Sprintf("%d", i+1))
	}
}

// buildMenuButton returns the header menu button that opens the main popover.
func (w *window) buildMenuButton() *gtk.MenuButton {
	menuBtn := gtk.NewMenuButton()

	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.AddCSSClass("flat")
	menuBtn.SetPopover(w.buildMenuPopover())

	return menuBtn
}

// buildMenuPopover builds the popover attached to the hamburger menu button.
func (w *window) buildMenuPopover() *gtk.Popover {
	box := gtk.NewBox(gtk.OrientationVertical, 0)

	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.SetMarginStart(4)
	box.SetMarginEnd(4)

	popover := gtk.NewPopover()

	prefsBtn := menuItem("preferences-system-symbolic", "Preferences")

	prefsBtn.ConnectClicked(func() {
		popover.Popdown()

		showConfigDialog(w.win, w)
	})

	box.Append(prefsBtn)

	shortcutsBtn := menuItem("input-keyboard-symbolic", "Keyboard Shortcuts")

	shortcutsBtn.ConnectClicked(func() {
		popover.Popdown()

		showShortcutsDialog(w.win, w.keys)
	})

	box.Append(shortcutsBtn)

	aboutBtn := menuItem("help-about-symbolic", "About bterm")

	aboutBtn.ConnectClicked(func() {
		popover.Popdown()

		showAboutDialog(w.win)
	})

	box.Append(aboutBtn)

	popover.SetChild(box)

	return popover
}

// formatBinding converts a normalized binding like "ctrl+shift+t" to a
// human-readable form like "Ctrl+Shift+T".
func formatBinding(b string) string {
	if b == "" {
		return ""
	}

	parts := strings.Split(b, "+")

	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}

	return strings.Join(parts, "+")
}

// headerActionTooltip returns a label with the action's current shortcut when bound.
func headerActionTooltip(keys *keymap.Layout, action keymap.Action, label string) string {
	binding := formatBinding(keys.BindingFor(action))

	if binding == "" {
		return label
	}

	return label + " (" + binding + ")"
}

// menuItem returns a flat button with a leading symbolic icon and a text label.
func menuItem(iconName, label string) *gtk.Button {
	row := gtk.NewBox(gtk.OrientationHorizontal, 8)

	img := gtk.NewImageFromIconName(iconName)
	lbl := gtk.NewLabel(label)

	lbl.SetHExpand(true)
	lbl.SetXAlign(0)

	row.Append(img)
	row.Append(lbl)

	btn := gtk.NewButton()

	btn.SetChild(row)
	btn.AddCSSClass("flat")

	return btn
}
