package ui

import (
	"fmt"
	"slices"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/keymap"
)

// buildTabBar creates the GtkHeaderBar with a tab-label box and installs a
// GtkStack as the window child. Must be called once before any tabs are added.
func (w *window) buildTabBar() {
	w.tabBox = gtk.NewBox(gtk.OrientationHorizontal, 2)

	header := gtk.NewHeaderBar()

	header.SetShowTitleButtons(true)

	settingsBtn := gtk.NewButton()

	settingsBtn.SetIconName("preferences-system-symbolic")
	settingsBtn.SetTooltipText("Preferences")
	settingsBtn.AddCSSClass("flat")
	settingsBtn.ConnectClicked(func() { showConfigDialog(w.win, w) })

	header.PackStart(settingsBtn)
	header.PackStart(w.tabBox)

	addBtn := gtk.NewButton()

	addBtn.SetIconName("list-add-symbolic")
	addBtn.SetTooltipText("New tab (" + formatBinding(w.keys.BindingFor(keymap.ActionNewTabEnd)) + ")")
	addBtn.AddCSSClass("flat")
	addBtn.ConnectClicked(func() { w.newTabEnd() })

	header.PackEnd(addBtn)

	menuBtn := gtk.NewMenuButton()

	menuBtn.SetIconName("open-menu-symbolic")
	menuBtn.AddCSSClass("flat")
	menuBtn.SetPopover(w.buildMenuPopover())

	header.PackEnd(menuBtn)

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

	t.area.onEmpty = func() { w.closeTab(t) }

	t.buildLabel(w, len(w.tabs)+1)

	t.area.onTitleChanged = func(title string) {
		t.title = title
		t.titleLabel.SetText(title)

		if len(w.tabs) > 0 && w.tabs[w.active] == t {
			w.win.SetTitle(title)
		}
	}

	w.stack.AddChild(t.area.root)
	w.tabBox.Append(t.label)

	w.tabs = append(w.tabs, t)
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

// activeCWD returns the working directory of the active tab's focused terminal,
// or an empty string when unavailable (callers fall back to $HOME via spawnTerm).
func (w *window) activeCWD() string {
	if len(w.tabs) == 0 {
		return ""
	}

	if ft := w.tabs[w.active].area.focusedTerminal(); ft != nil {
		return ft.CurrentDir()
	}

	return ""
}
