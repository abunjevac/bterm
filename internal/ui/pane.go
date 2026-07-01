package ui

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/config"
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
	// half the current terminal's size instead of GTK's default position=0.
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
	cfg := pa.win.bundle.Config
	shell := config.InferShell(cfg.Shell, os.Getenv("SHELL"))

	t.SetFont(pa.win.fontFamily, pa.win.fontSize)
	t.SetColors(pa.win.palette)
	t.SetScrollback(cfg.Scrollback)

	pa.win.installTerminalNotifications(t)

	t.Spawn(workingDir, shell, shellArgs(cfg), func(_ int, _ error) {})
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

	// Capture the focused terminal's current size before rebuild clears it.
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
	if w := pa.widgets[pa.tree.Focused()]; w != nil {
		gtk.BaseWidget(w).GrabFocus()
	}
}

// rebuild replaces the child of pa.root with a freshly-built widget tree
// that mirrors the current panetree state.
func (pa *paneArea) rebuild() {
	// Unparent all terminal widgets first so they can be re-parented into
	// the new GtkPaned tree. Without this, GTK rejects SetStartChild/SetEndChild
	// on widgets that still have a parent from the previous layout.
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

	// prevent panes from collapsing to zero when GTK computes initial positions
	paned.SetShrinkStartChild(false)
	paned.SetShrinkEndChild(false)

	if a := pa.buildFromDesc(d.A); a != nil {
		paned.SetStartChild(a)
	}

	if b := pa.buildFromDesc(d.B); b != nil {
		paned.SetEndChild(b)
	}

	// If this is the GtkPaned created by the most recent split, position it at
	// half the focused terminal's pre-split size so both halves start equal.
	// The new terminal's ID uniquely identifies this paned in the tree.
	if pa.splitHintNewID != 0 && pa.splitHintPos > 0 && (pa.isHintLeaf(d.A) || pa.isHintLeaf(d.B)) {
		paned.SetPosition(pa.splitHintPos)

		pa.splitHintNewID = 0
	}

	return paned
}
