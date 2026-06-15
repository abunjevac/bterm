package panetree_test

import (
	"testing"

	"github.com/abunjevac/bterm/internal/ui/panetree"
	"github.com/stretchr/testify/require"
)

func TestSingleLeafTree(t *testing.T) {
	tr := panetree.New(1)

	require.Equal(t, []int{1}, tr.Leaves())
	require.Equal(t, 1, tr.Focused())
}

func TestSplitCreatesSibling(t *testing.T) {
	tr := panetree.New(1)

	tr.Split(panetree.LeftRight, 2)

	require.ElementsMatch(t, []int{1, 2}, tr.Leaves())
	require.Equal(t, 2, tr.Focused()) // focus moves to new pane
}

func TestCloseCollapsesToSibling(t *testing.T) {
	tr := panetree.New(1)

	tr.Split(panetree.LeftRight, 2)
	tr.Close(2)

	require.Equal(t, []int{1}, tr.Leaves())
	require.Equal(t, 1, tr.Focused())
}

func TestCloseLastLeafEmptiesTree(t *testing.T) {
	tr := panetree.New(1)

	tr.Close(1)

	require.Empty(t, tr.Leaves())
	require.True(t, tr.Empty())
}

func TestFocusNeighborDirectional(t *testing.T) {
	tr := panetree.New(1)

	tr.Split(panetree.LeftRight, 2) // 1 | 2, focus=2

	require.Equal(t, 1, tr.Neighbor(panetree.DirLeft))

	tr.SetFocus(1)

	require.Equal(t, 2, tr.Neighbor(panetree.DirRight))
	require.Equal(t, 0, tr.Neighbor(panetree.DirUp)) // none — wrong orientation
}

func TestNestedSplitNeighbor(t *testing.T) {
	tr := panetree.New(1)

	tr.Split(panetree.LeftRight, 2) // 1 | 2, focus=2
	tr.Split(panetree.TopBottom, 3) // 1 | (2 / 3), focus=3

	require.Equal(t, 2, tr.Neighbor(panetree.DirUp))

	tr.SetFocus(1)

	require.Equal(t, 2, tr.Neighbor(panetree.DirRight)) // first leaf of right subtree
}

func TestNeighborAfterClose(t *testing.T) {
	tr := panetree.New(1)

	tr.Split(panetree.LeftRight, 2) // 1 | 2, focus=2
	tr.Close(2)                     // back to single pane

	// Before fix, Close left p.parent pointing to itself, causing an infinite
	// loop in Neighbor. Verify it terminates and returns 0.
	require.Equal(t, 0, tr.Neighbor(panetree.DirLeft))
	require.Equal(t, 0, tr.Neighbor(panetree.DirRight))
	require.Equal(t, 1, tr.Focused())
}

func TestNeighborAfterCloseNested(t *testing.T) {
	tr := panetree.New(1)

	tr.Split(panetree.LeftRight, 2) // 1 | 2, focus=2
	tr.Split(panetree.TopBottom, 3) // 1 | (2 / 3), focus=3
	tr.Close(3)                     // back to 1 | 2

	tr.SetFocus(2)

	require.Equal(t, 1, tr.Neighbor(panetree.DirLeft))
	require.Equal(t, 0, tr.Neighbor(panetree.DirRight))
}

func TestCloseMiddleNodeFocusesRemaining(t *testing.T) {
	tr := panetree.New(1)

	tr.Split(panetree.LeftRight, 2)
	tr.Split(panetree.LeftRight, 3) // tree is 1 | (2 | 3), focused=3

	require.ElementsMatch(t, []int{1, 2, 3}, tr.Leaves())

	tr.Close(3)

	require.ElementsMatch(t, []int{1, 2}, tr.Leaves())

	// focused should be some valid leaf
	f := tr.Focused()

	require.True(t, f == 1 || f == 2)
}
