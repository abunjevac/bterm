package ui

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/terminal"
	"github.com/abunjevac/bterm/internal/ui/panetree"
)

// paneArea owns the pane-tree model and the GTK widget tree that mirrors it.
type paneArea struct {
	tree    *panetree.Tree
	terms   map[int]terminal.Terminal
	widgets map[int]gtk.Widgetter
	root    *gtk.Box
	nextID  int

	win            *window
	onEmpty        func()
	onTitleChanged func(string)

	// splitHintNewID / splitHintPos carry a one-shot hint from split() to
	// buildFromDesc: set this pixel position on the GtkPaned that contains the
	// newly-created terminal (identified by its ID) so the new pane starts at
	// half the current terminal's size instead of GTK's default position=0
	splitHintNewID int
	splitHintPos   int
}

// newPaneArea creates a paneArea seeded with an already-configured, already-spawned terminal.
func newPaneArea(w *window, firstTerm terminal.Terminal) *paneArea {
	pa := &paneArea{
		terms:   make(map[int]terminal.Terminal),
		widgets: make(map[int]gtk.Widgetter),
		nextID:  1,
		win:     w,
	}

	id := pa.allocID()

	pa.tree = panetree.New(id)

	pa.registerTerm(id, firstTerm)

	pa.root = gtk.NewBox(gtk.OrientationVertical, 0)

	pa.root.SetVExpand(true)
	pa.root.SetHExpand(true)

	pa.rebuild()

	return pa
}

func (pa *paneArea) allocID() int {
	id := pa.nextID

	pa.nextID++

	return id
}

// registerTerm stores a terminal and wires its lifecycle callbacks.
func (pa *paneArea) registerTerm(id int, t terminal.Terminal) {
	pa.terms[id] = t

	w := t.Widget()

	gtk.BaseWidget(w).SetVExpand(true)
	gtk.BaseWidget(w).SetHExpand(true)

	pa.widgets[id] = w

	pa.installTerminalContextMenu(id, w)
	pa.installFontScroll(id, w)
	pa.installFocusSync(id, w)

	t.OnChildExited(func(_ int) {
		pa.closeID(id)
	})

	t.OnTitleChanged(func(title string) {
		if pa.tree.Focused() == id && pa.onTitleChanged != nil {
			pa.onTitleChanged(title)
		}
	})
}

// spawnInTerm configures and spawns a shell in t, using workingDir as the cwd.
func (pa *paneArea) spawnInTerm(t terminal.Terminal, workingDir string) {
	pa.win.configureAndSpawn(t, workingDir)
}

// split splits the focused pane with a new terminal inheriting the focused pane's cwd.
func (pa *paneArea) split(o panetree.Orientation) {
	var cwd string

	if focused := pa.terms[pa.tree.Focused()]; focused != nil {
		cwd = focused.CurrentDir()
	}

	if cwd == "" {
		cwd, _ = os.UserHomeDir()
	}

	// capture the focused terminal's current size before rebuild clears it
	// The new GtkPaned will be positioned at half this value.
	var splitPos int

	if fw := pa.widgets[pa.tree.Focused()]; fw != nil {
		base := gtk.BaseWidget(fw)

		if o == panetree.LeftRight {
			splitPos = base.Width() / 2
		} else {
			splitPos = base.Height() / 2
		}
	}

	id := pa.allocID()
	t := pa.win.newTerm()

	pa.registerTerm(id, t)
	pa.spawnInTerm(t, cwd)

	pa.splitHintNewID = id
	pa.splitHintPos = splitPos

	pa.tree.Split(o, id)

	pa.rebuild()
	pa.grabFocus()
}

// closeFocused closes the currently focused pane.
func (pa *paneArea) closeFocused() {
	pa.closeID(pa.tree.Focused())
}

// closeID removes the pane with id, updating the model and widget tree.
func (pa *paneArea) closeID(id int) {
	pa.tree.Close(id)

	delete(pa.terms, id)
	delete(pa.widgets, id)

	if pa.tree.Empty() {
		if pa.onEmpty != nil {
			pa.onEmpty()
		}

		return
	}

	pa.rebuild()
	pa.grabFocus()
}

// focusDir moves focus to the neighbor in direction d.
func (pa *paneArea) focusDir(d panetree.Direction) {
	neighbor := pa.tree.Neighbor(d)

	if neighbor == 0 {
		return
	}

	pa.tree.SetFocus(neighbor)

	pa.grabFocus()
}

// resizeFocused adjusts the nearest enclosing GtkPaned in the appropriate
// axis by a fixed step. It is a best-effort operation.
func (pa *paneArea) resizeFocused(d panetree.Direction) {
	const step = 40

	w := pa.widgets[pa.tree.Focused()]

	if w == nil {
		return
	}

	current := gtk.BaseWidget(w).Parent()

	for current != nil {
		if paned, ok := current.(*gtk.Paned); ok {
			pos := paned.Position()

			switch d {
			case panetree.DirLeft, panetree.DirUp:
				paned.SetPosition(pos - step)
			case panetree.DirRight, panetree.DirDown:
				paned.SetPosition(pos + step)
			}

			return
		}

		current = gtk.BaseWidget(current).Parent()
	}
}

// focusedTerminal returns the terminal for the focused pane, or nil.
func (pa *paneArea) focusedTerminal() terminal.Terminal {
	return pa.terms[pa.tree.Focused()]
}

// grabFocus gives keyboard focus to the focused pane's widget.
func (pa *paneArea) grabFocus() {
	w := pa.widgets[pa.tree.Focused()]
	if w == nil {
		return
	}

	target := gtk.BaseWidget(w)

	if sw, ok := w.(*gtk.ScrolledWindow); ok {
		if child := sw.Child(); child != nil {
			target = gtk.BaseWidget(child)
		}
	}

	grabFocusWhenMapped(target)
}

// grabFocusWhenMapped grabs keyboard focus on w, deferring to its "map"
// signal if w is not mapped yet. GTK requires a widget to have gone through
// a size-allocate pass (i.e. be mapped) before it can actually receive
// native keyboard/IM input: calling GrabFocus() before that updates GTK's
// logical focus-widget pointer but silently fails to move real input focus,
// leaving whichever widget was previously focused still receiving keys.
// This is what happens right after rebuild() reparents a freshly-split pane,
// since the new GtkPaned hasn't been through layout yet.
func grabFocusWhenMapped(w *gtk.Widget) {
	if w.Mapped() {
		w.GrabFocus()

		return
	}

	var handle glib.SignalHandle

	handle = w.ConnectMap(func() {
		w.GrabFocus()
		w.HandlerDisconnect(handle)
	})
}

// installFocusSync keeps the pane-tree model's focused id in sync with GTK's
// real keyboard focus. All keymap-bound actions (copy, paste, close, split,
// ...) are dispatched from a window-level key controller and read the
// focused pane through the model, not GTK's actual focus widget — so
// without this, clicking into a different pane (which GTK focuses on its
// own) leaves those actions still targeting whichever pane was last focused
// via split, close, focusDir, or the context menu.
func (pa *paneArea) installFocusSync(id int, widget gtk.Widgetter) {
	focus := gtk.NewEventControllerFocus()

	focus.ConnectEnter(func() {
		pa.tree.SetFocus(id)
	})

	gtk.BaseWidget(widget).AddController(focus)
}

// installFontScroll attaches a scroll controller that adjusts the font size
// when Ctrl is held while scrolling.
func (pa *paneArea) installFontScroll(id int, widget gtk.Widgetter) {
	scroll := gtk.NewEventControllerScroll(gtk.EventControllerScrollBothAxes)

	scroll.SetPropagationPhase(gtk.PhaseCapture)
	scroll.ConnectScroll(func(_ float64, dy float64) bool {
		if scroll.CurrentEventState()&gdk.ControlMask == 0 {
			return false
		}

		t := pa.terms[id]

		if t == nil {
			return false
		}

		if dy < 0 {
			pa.win.fontSize++
		} else if pa.win.fontSize > 4 {
			pa.win.fontSize--
		}

		t.SetFont(pa.win.fontFamily, pa.win.fontSize)

		return true
	})

	gtk.BaseWidget(widget).AddController(scroll)
}

// rebuild replaces the child of pa.root with a freshly-built widget tree
// that mirrors the current panetree state.
func (pa *paneArea) rebuild() {
	// unparent all terminal widgets first so they can be re-parented into
	// the new GtkPaned tree. Without this, GTK rejects SetStartChild/SetEndChild
	// on widgets that still have a parent from the previous layout
	for _, w := range pa.widgets {
		gtk.BaseWidget(w).Unparent()
	}

	if child := pa.root.FirstChild(); child != nil {
		pa.root.Remove(child)
	}

	desc := pa.tree.Describe()

	if desc == nil {
		pa.splitHintNewID = 0
		pa.splitHintPos = 0

		return
	}

	child := pa.buildFromDesc(desc)

	pa.splitHintNewID = 0 // consumed or unused — clear after build
	pa.splitHintPos = 0

	if child != nil {
		pa.root.Append(child)
	}
}

// isHintLeaf reports whether d is the newly-created pane from the last split,
// used to identify which GtkPaned should receive the explicit position hint.
func (pa *paneArea) isHintLeaf(d *panetree.NodeDesc) bool {
	return d != nil && d.IsLeaf && d.ID == pa.splitHintNewID
}

// buildFromDesc recursively converts a NodeDesc into a GTK widget subtree.
func (pa *paneArea) buildFromDesc(d *panetree.NodeDesc) gtk.Widgetter {
	if d == nil {
		return nil
	}

	if d.IsLeaf {
		return pa.widgets[d.ID]
	}

	orient := gtk.OrientationHorizontal

	if d.Orient == panetree.TopBottom {
		orient = gtk.OrientationVertical
	}

	paned := gtk.NewPaned(orient)

	paned.SetVExpand(true)
	paned.SetHExpand(true)

	// use a wide handle so the resize grab area matches the visible separator
	// exactly. with the default narrow handle GTK adds an invisible
	// HANDLE_EXTRA_SIZE (6px) padding around the line, which steals drags from
	// adjacent terminal text and blocks mouse selection near the split.
	paned.SetWideHandle(true)

	// prevent panes from collapsing to zero when GTK computes initial positions
	paned.SetShrinkStartChild(false)
	paned.SetShrinkEndChild(false)

	if a := pa.buildFromDesc(d.A); a != nil {
		paned.SetStartChild(a)
	}

	if b := pa.buildFromDesc(d.B); b != nil {
		paned.SetEndChild(b)
	}

	// if this is the GtkPaned created by the most recent split, position it at
	// half the focused terminal's pre-split size so both halves start equal.
	// The new terminal's ID uniquely identifies this paned in the tree.
	if pa.splitHintNewID != 0 && pa.splitHintPos > 0 && (pa.isHintLeaf(d.A) || pa.isHintLeaf(d.B)) {
		paned.SetPosition(pa.splitHintPos)

		pa.splitHintNewID = 0
	}

	return paned
}
