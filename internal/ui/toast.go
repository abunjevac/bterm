package ui

import (
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// toaster overlays a short-lived message label over a content widget.
// Only one message is shown at a time; a new show() restarts the timer.
type toaster struct {
	overlay *gtk.Overlay
	label   *gtk.Label
	timerID glib.SourceHandle
}

func newToaster(content gtk.Widgetter) *toaster {
	t := &toaster{}

	t.label = gtk.NewLabel("")

	t.label.AddCSSClass("bterm-toast")
	t.label.SetHAlign(gtk.AlignEnd)
	t.label.SetVAlign(gtk.AlignStart)
	t.label.SetMarginTop(8)
	t.label.SetMarginEnd(8)
	t.label.SetVisible(false)

	t.overlay = gtk.NewOverlay()

	t.overlay.SetChild(content)
	t.overlay.AddOverlay(t.label)

	return t
}

// show displays msg for ~1.5 s then hides it. Calling show again while a
// message is visible resets the timer.
func (t *toaster) show(msg string) {
	if t.timerID != 0 {
		glib.SourceRemove(t.timerID)

		t.timerID = 0
	}

	t.label.SetText(msg)
	t.label.SetVisible(true)

	t.timerID = glib.TimeoutAdd(1500, func() bool {
		t.label.SetVisible(false)

		t.timerID = 0

		return false
	})
}
