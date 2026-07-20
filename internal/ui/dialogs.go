package ui

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type dialogSpec struct {
	width  int
	height int
	fixed  bool
}

func preferencesDialogSpec() dialogSpec {
	return dialogSpec{width: 500, height: 760, fixed: true}
}

func shortcutsDialogSpec() dialogSpec {
	return dialogSpec{width: 760, height: 720, fixed: true}
}

func aboutDialogSpec() dialogSpec {
	return dialogSpec{width: 460, height: 440, fixed: true}
}

func applyDialogSpec(win *gtk.Window, spec dialogSpec) {
	win.SetDefaultSize(spec.width, spec.height)
	win.SetResizable(!spec.fixed)
}

func addEscapeClose(win *gtk.Window) {
	ctl := gtk.NewEventControllerKey()

	ctl.SetPropagationPhase(gtk.PhaseCapture)
	ctl.ConnectKeyPressed(func(keyval, _ uint, _ gdk.ModifierType) bool {
		if keyval != gdk.KEY_Escape {
			return false
		}

		win.Close()

		return true
	})

	win.AddController(ctl)
}
