// Package panetree is a GTK-free binary tree modeling a tab's split layout.
// Leaves carry an opaque int ID the UI layer maps to a terminal+widget.
package panetree

// Orientation of a split.
type Orientation int

const (
	LeftRight Orientation = iota // children side by side
	TopBottom                    // children stacked
)

// Direction for focus navigation.
type Direction int

const (
	DirLeft Direction = iota
	DirRight
	DirUp
	DirDown
)

type node struct {
	// leaf fields
	id int
	// split fields (id == 0 when split)
	orient Orientation
	a, b   *node
	parent *node
}

func (n *node) isLeaf() bool { return n.a == nil && n.b == nil }
