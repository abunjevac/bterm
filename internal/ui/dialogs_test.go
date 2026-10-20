package ui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDialogSpecs(t *testing.T) {
	t.Parallel()

	require.Equal(t, dialogSpec{width: 900, height: 460, fixed: true}, preferencesDialogSpec())
	require.Equal(t, dialogSpec{width: 760, height: 720, fixed: true}, shortcutsDialogSpec())
	require.Equal(t, dialogSpec{width: 460, height: 440, fixed: true}, aboutDialogSpec())
	require.Equal(t, dialogSpec{width: 720, height: 480, fixed: false}, logDialogSpec())
}

func TestAboutCopyright(t *testing.T) {
	t.Parallel()

	require.Equal(t, "© 2026 Alan Bunjevac", aboutCopyright)
}
