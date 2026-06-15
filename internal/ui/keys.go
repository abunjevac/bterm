package ui

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/keymap"
)

// installKeys attaches a capture-phase key controller to the window.
func (w *window) installKeys() {
	ctl := gtk.NewEventControllerKey()

	ctl.SetPropagationPhase(gtk.PhaseCapture)
	ctl.ConnectKeyPressed(func(keyval, _ uint, state gdk.ModifierType) (ok bool) {
		binding := buildBinding(keyval, state)
		act := w.keys.Lookup(binding)

		if act == keymap.ActionUnknown {
			return false
		}

		w.dispatch(act)

		return true
	})

	w.win.AddController(ctl)
}

// buildBinding converts a GTK key event into a normalized binding string
// matching the keymap format, e.g. "ctrl+shift+o".
func buildBinding(keyval uint, state gdk.ModifierType) string {
	// canonical modifier order matches keymap.modifierOrder: ctrl, shift, alt, super
	var parts []string

	if state&gdk.ControlMask != 0 {
		parts = append(parts, "ctrl")
	}

	if state&gdk.ShiftMask != 0 {
		parts = append(parts, "shift")
	}

	if state&gdk.AltMask != 0 {
		parts = append(parts, "alt")
	}

	name := strings.ToLower(gdk.KeyvalName(gdk.KeyvalToLower(keyval)))

	parts = append(parts, name)

	return strings.Join(parts, "+")
}
