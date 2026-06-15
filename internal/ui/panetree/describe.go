package panetree

// NodeDesc describes a node for widget-building by the UI layer.
type NodeDesc struct {
	IsLeaf bool
	ID     int         // set when IsLeaf == true
	Orient Orientation // set when IsLeaf == false
	A, B   *NodeDesc   // children when IsLeaf == false
}

// Describe returns a structural snapshot of the tree for widget building.
func (t *Tree) Describe() *NodeDesc {
	return describeNode(t.root)
}

func describeNode(n *node) *NodeDesc {
	if n == nil {
		return nil
	}

	if n.isLeaf() {
		return &NodeDesc{IsLeaf: true, ID: n.id}
	}

	return &NodeDesc{
		IsLeaf: false,
		Orient: n.orient,
		A:      describeNode(n.a),
		B:      describeNode(n.b),
	}
}
