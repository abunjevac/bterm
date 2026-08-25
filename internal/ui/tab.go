package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

const tabDetachDragThreshold = 72

// tabReorderCrossFraction is the fraction of a neighboring tab's width the
// pointer must cross before it swaps places with the dragged tab. Without
// any slide animation, each swap is an instant re-layout of the tab strip;
// requiring most of the neighbor's width to be crossed (rather than just
// half) gives dragging a dead zone that absorbs ordinary hand jitter, so
// swaps fire noticeably less often during a drag.
const tabReorderCrossFraction = 0.9

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
	selectBtn.SetCursorFromName("grab")
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

	// consumedX tracks how much of the gesture's cumulative offset has
	// already been applied as tab swaps, since GestureDrag reports offsets
	// relative to the drag start point rather than as per-update deltas.
	var consumedX float64

	drag.SetButton(1)
	drag.SetExclusive(true)
	drag.SetPropagationPhase(gtk.PhaseCapture)
	drag.ConnectDragBegin(func(_, _ float64) {
		drag.SetState(gtk.EventSequenceClaimed)

		consumedX = 0

		setDragSurfaceCursor(selectBtn, "move")

		selectBtn.AddCSSClass("bterm-tab-dragging")

		t.label.AddCSSClass("bterm-tab-dragging")
	})
	drag.ConnectDragUpdate(func(offsetX, _ float64) {
		setDragSurfaceCursor(selectBtn, "move")

		consumedX = reorderTabOnDrag(w, t, offsetX, consumedX)
	})
	drag.ConnectDragEnd(func(_, offsetY float64) {
		setDragSurfaceCursor(selectBtn, "")

		selectBtn.SetCursorFromName("grab")
		selectBtn.RemoveCSSClass("bterm-tab-dragging")

		t.label.RemoveCSSClass("bterm-tab-dragging")

		if offsetY > -tabDetachDragThreshold && offsetY < tabDetachDragThreshold {
			return
		}

		w.detachTab(t)
	})

	selectBtn.AddController(drag)
}

// reorderTabOnDrag swaps t with a neighboring tab whenever the cumulative
// drag offsetX has crossed that neighbor's midpoint, repeating until no
// further swap is warranted. consumedX is the portion of offsetX already
// applied by prior swaps; the returned value updates it for the next call.
func reorderTabOnDrag(w *window, t *tab, offsetX, consumedX float64) float64 {
	for {
		idx := w.tabIndex(t)

		if idx < 0 {
			return consumedX
		}

		delta := offsetX - consumedX

		switch {
		case delta > 0 && idx+1 < len(w.tabs):
			neighborWidth := float64(w.tabs[idx+1].label.Width())

			if delta <= neighborWidth*tabReorderCrossFraction {
				return consumedX
			}

			w.swapAdjacentTabs(idx)

			consumedX += neighborWidth
		case delta < 0 && idx > 0:
			neighborWidth := float64(w.tabs[idx-1].label.Width())

			if -delta <= neighborWidth*tabReorderCrossFraction {
				return consumedX
			}

			w.swapAdjacentTabs(idx - 1)

			consumedX -= neighborWidth
		default:
			return consumedX
		}
	}
}

func setDragSurfaceCursor(widget gtk.Widgetter, name string) {
	native := gtk.BaseWidget(widget).Native()

	if native == nil {
		return
	}

	surface := native.Surface()

	if surface == nil {
		return
	}

	var cursor *gdk.Cursor

	if name != "" {
		cursor = gdk.NewCursorFromName(name, nil)
	}

	gdk.BaseSurface(surface).SetCursor(cursor)
}
