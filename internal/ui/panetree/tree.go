package panetree

// Tree is a tab's pane layout.
type Tree struct {
	root    *node
	focused *node
}

// New creates a tree with a single leaf.
func New(id int) *Tree {
	r := &node{id: id}

	return &Tree{root: r, focused: r}
}

// Empty reports whether the tree has no panes.
func (t *Tree) Empty() bool { return t.root == nil }

// Focused returns the focused leaf id (0 if empty).
func (t *Tree) Focused() int {
	if t.focused == nil {
		return 0
	}

	return t.focused.id
}

// SetFocus focuses the leaf with the given id if present.
func (t *Tree) SetFocus(id int) {
	if n := t.find(t.root, id); n != nil {
		t.focused = n
	}
}

// Leaves returns all leaf ids (in-order, left/top first).
func (t *Tree) Leaves() []int {
	var out []int
	var walk func(n *node)

	walk = func(n *node) {
		if n == nil {
			return
		}

		if n.isLeaf() {
			out = append(out, n.id)

			return
		}

		walk(n.a)
		walk(n.b)
	}

	walk(t.root)

	return out
}

// Split replaces the focused leaf with a split: old leaf on the left/top (a) + new leaf (b).
// Focus moves to the new leaf.
func (t *Tree) Split(o Orientation, newID int) {
	target := t.focused

	if target == nil || !target.isLeaf() {
		return
	}

	oldLeaf := &node{id: target.id, parent: target}
	newLeaf := &node{id: newID, parent: target}

	target.id = 0
	target.orient = o
	target.a = oldLeaf
	target.b = newLeaf

	t.focused = newLeaf
}

// Close removes the leaf with id and collapses its parent split to the sibling.
func (t *Tree) Close(id int) {
	n := t.find(t.root, id)

	if n == nil {
		return
	}

	if n.parent == nil { // root leaf
		t.root = nil
		t.focused = nil

		return
	}

	p := n.parent
	sib := p.a

	if sib == n {
		sib = p.b
	}

	// promote sibling into parent's slot
	*p = *sib

	if p.a != nil {
		p.a.parent = p
	}

	if p.b != nil {
		p.b.parent = p
	}

	// fix focused: if it was the closed node or is gone, move to first leaf under p
	if t.focused == n || t.find(t.root, t.focused.id) == nil {
		t.focused = t.firstLeaf(p)
	}
}

// Neighbor returns the id of the focused leaf's neighbor in dir, or 0 if none.
func (t *Tree) Neighbor(dir Direction) int {
	if t.focused == nil {
		return 0
	}

	wantOrient := LeftRight

	if dir == DirUp || dir == DirDown {
		wantOrient = TopBottom
	}

	cur := t.focused

	for cur.parent != nil {
		p := cur.parent

		if p.orient == wantOrient {
			if (dir == DirLeft || dir == DirUp) && p.b == cur {
				return t.firstLeaf(p.a).id
			}

			if (dir == DirRight || dir == DirDown) && p.a == cur {
				return t.firstLeaf(p.b).id
			}
		}

		cur = p
	}

	return 0
}

func (t *Tree) find(n *node, id int) *node {
	if n == nil {
		return nil
	}

	if n.isLeaf() {
		if n.id == id {
			return n
		}

		return nil
	}

	if r := t.find(n.a, id); r != nil {
		return r
	}

	return t.find(n.b, id)
}

func (t *Tree) firstLeaf(n *node) *node {
	for !n.isLeaf() {
		n = n.a
	}

	return n
}
