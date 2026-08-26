package ui

import (
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/terminal/kitty"
)

// f10Sequence is the xterm-compatible escape sequence VTE would normally
// send to the child process for an unmodified F10 key press.
const f10Sequence = "\x1b[21~"

// kittyDisambiguateTerminal is an optional capability for terminals that
// implement the kitty keyboard protocol disambiguate mode.
type kittyDisambiguateTerminal interface {
	KittyDisambiguate() bool
}

// installKeys attaches a capture-phase key controller to the window.
func (w *window) installKeys() {
	ctl := gtk.NewEventControllerKey()

	ctl.SetPropagationPhase(gtk.PhaseCapture)
	ctl.ConnectKeyPressed(func(keyval, _ uint, state gdk.ModifierType) bool {
		if keyval == gdk.KEY_F10 {
			// xkeyboard-config reports plain F10 with a phantom Shift
			// modifier, which GTK reads as "activate the header bar" and
			// highlights it instead of forwarding the key. Reclaim F10 for
			// the terminal so apps like Midnight Commander can use it
			w.feedFocusedTerminal(f10Sequence)

			return true
		}

		binding := buildBinding(keyval, state)
		act := w.keys.Lookup(binding)

		if act != keymap.ActionUnknown {
			w.dispatch(act)

			return true
		}

		// no keymap binding matched. If the focused terminal has the
		// kitty keyboard protocol disambiguate mode active, encode the
		// key as a CSI u sequence and send it directly to the shell
		if ft := w.focusedTerminal(); ft != nil {
			if kt, ok := ft.(kittyDisambiguateTerminal); ok && kt.KittyDisambiguate() {
				if r := kitty.EncodeKey(keyval, state); r.Bytes() != nil {
					ft.FeedChild(r.Bytes())

					return true
				}
			}
		}

		return false
	})

	w.win.AddController(ctl)
}

// feedFocusedTerminal writes seq to the active tab's focused terminal, as if typed.
func (w *window) feedFocusedTerminal(seq string) {
	if len(w.tabs) == 0 {
		return
	}

	if ft := w.tabs[w.active].area.focusedTerminal(); ft != nil {
		ft.FeedChild([]byte(seq))
	}
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

	if state&gdk.SuperMask != 0 {
		parts = append(parts, "super")
	}

	name := strings.ToLower(gdk.KeyvalName(gdk.KeyvalToLower(keyval)))

	parts = append(parts, name)

	return strings.Join(parts, "+")
}
