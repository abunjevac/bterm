package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

const tabDetachDragThreshold = 72

// tab represents one terminal tab: a pane layout, its current title, and the
// header-bar label widget.
type tab struct {
	area       *paneArea
	label      *gtk.Box   // outer label container shown in the header bar
	numLabel   *gtk.Label // number badge (.bterm-tab-num)
	titleLabel *gtk.Label // terminal OSC title or fallback
	title      string     // last title reported by the focused pane
}

// buildLabel constructs the header-bar label widget for t.
// idx is the 1-based display number for the badge.
// Must be called after t.area is initialised.
func (t *tab) buildLabel(w *window, idx int) {
	t.numLabel = gtk.NewLabel(fmt.Sprintf("%d", idx))

	t.numLabel.AddCSSClass("bterm-tab-num")

	title := t.title

	if title == "" {
		title = "Terminal"
	}

	t.titleLabel = gtk.NewLabel(title)

	t.titleLabel.SetEllipsize(pango.EllipsizeEnd)
	t.titleLabel.SetMaxWidthChars(20)

	innerBox := gtk.NewBox(gtk.OrientationHorizontal, 4)

	innerBox.Append(t.numLabel)
	innerBox.Append(t.titleLabel)

	// selectBtn selects this tab when clicked.
	selectBtn := gtk.NewButton()

	selectBtn.SetChild(innerBox)
	selectBtn.AddCSSClass("flat")
	selectBtn.ConnectClicked(func() {
		if i := w.tabIndex(t); i >= 0 {
			w.selectTab(i)
		}
	})

	installTabDetachDrag(selectBtn, w, t)

	// closeBtn is a sibling of selectBtn, not nested inside it.
	closeBtn := gtk.NewButton()

	closeBtn.SetIconName("window-close-symbolic")
	closeBtn.AddCSSClass("flat")
	closeBtn.SetTooltipText("Close tab")
	closeBtn.ConnectClicked(func() {
		w.closeTab(t)
	})

	t.label = gtk.NewBox(gtk.OrientationHorizontal, 0)

	t.label.SetMarginTop(2)
	t.label.SetMarginBottom(2)
	t.label.Append(selectBtn)
	t.label.Append(closeBtn)
}

func installTabDetachDrag(selectBtn *gtk.Button, w *window, t *tab) {
	drag := gtk.NewGestureDrag()

	drag.SetButton(1)
	drag.SetExclusive(true)
	drag.SetPropagationPhase(gtk.PhaseCapture)
	drag.ConnectDragBegin(func(_, _ float64) {
		drag.SetState(gtk.EventSequenceClaimed)
	})
	drag.ConnectDragEnd(func(_, offsetY float64) {
		if offsetY > -tabDetachDragThreshold && offsetY < tabDetachDragThreshold {
			return
		}

		w.detachTab(t)
	})
	selectBtn.AddController(drag)
}
