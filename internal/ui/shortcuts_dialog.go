package ui

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/keymap"
)

func showShortcutsDialog(parent *gtk.ApplicationWindow, layout *keymap.Layout) {
	grid := gtk.NewGrid()

	grid.SetRowSpacing(4)
	grid.SetColumnSpacing(16)
	grid.SetMarginTop(8)
	grid.SetMarginBottom(8)
	grid.SetMarginStart(8)
	grid.SetMarginEnd(8)

	for i, entry := range layout.Entries() {
		nameLabel := gtk.NewLabel(humanizeAction(entry.Action))

		nameLabel.SetHAlign(gtk.AlignEnd)

		keysLabel := gtk.NewLabel(strings.Join(entry.Keys, ", "))

		keysLabel.SetHAlign(gtk.AlignStart)
		keysLabel.AddCSSClass("bterm-shortcut-key")

		grid.Attach(nameLabel, 0, i, 1, 1)
		grid.Attach(keysLabel, 1, i, 1, 1)
	}

	scroll := gtk.NewScrolledWindow()

	scroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scroll.SetChild(grid)

	win := gtk.NewWindow()

	win.SetTitle("Keyboard Shortcuts")
	win.SetTransientFor(&parent.Window)
	win.SetModal(true)
	win.SetDefaultSize(420, 520)
	win.SetChild(scroll)
	win.Present()
}

func humanizeAction(a keymap.Action) string {
	words := strings.Split(a.String(), "_")

	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}

	return strings.Join(words, " ")
}
