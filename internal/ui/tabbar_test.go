package ui

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/abunjevac/bterm/internal/keymap"
)

func TestHeaderActionTooltipIncludesConfiguredShortcut(t *testing.T) {
	t.Parallel()

	keys := keymap.Default()

	require.Equal(t, "Split left/right (Ctrl+Shift+O)", headerActionTooltip(keys, keymap.ActionSplitLeftRight, "Split left/right"))
	require.Equal(t, "Split top/bottom (Ctrl+Shift+E)", headerActionTooltip(keys, keymap.ActionSplitTopBottom, "Split top/bottom"))
}
