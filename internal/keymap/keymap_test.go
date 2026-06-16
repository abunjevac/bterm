package keymap_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/abunjevac/bterm/internal/keymap"
)

//go:embed testdata/valid_keymap.toml
var validKeymapTOML string

//go:embed testdata/conflict_keymap.toml
var conflictKeymapTOML string

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Ctrl+Shift+O", "ctrl+shift+o"},
		{"ctrl+shift+o", "ctrl+shift+o"},
		{"Shift+Ctrl+O", "ctrl+shift+o"}, // canonical modifier order
		{"ALT+LEFT", "alt+left"},
		{"Return", "return"},
		{"Escape", "escape"},
		{"ctrl+kp_add", "ctrl+kp_add"},
		{"super+alt+ctrl+x", "ctrl+alt+super+x"}, // reorder modifiers
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := keymap.Normalize(tc.input)
			require.NoError(t, err)

			require.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeErrors(t *testing.T) {
	_, err := keymap.Normalize("")
	require.Error(t, err)

	// modifiers-only — no key component
	_, err = keymap.Normalize("ctrl+shift")
	require.Error(t, err)
}

func TestParseAndLookup(t *testing.T) {
	layout, err := keymap.Parse(validKeymapTOML)
	require.NoError(t, err)

	require.Equal(t, keymap.ActionCopy, layout.Lookup("ctrl+shift+c"))
	require.Equal(t, keymap.ActionCopy, layout.Lookup("ctrl+insert"))
	require.Equal(t, keymap.ActionPaste, layout.Lookup("ctrl+shift+v"))
	require.Equal(t, keymap.ActionNewTabEnd, layout.Lookup("ctrl+shift+t"))
	require.Equal(t, keymap.ActionCloseTab, layout.Lookup("ctrl+shift+q"))

	// unmapped binding
	require.Equal(t, keymap.ActionUnknown, layout.Lookup("ctrl+shift+z"))
}

func TestConflicts(t *testing.T) {
	layout, err := keymap.Parse(conflictKeymapTOML)
	require.NoError(t, err)

	conflicts := layout.Conflicts()

	require.Equal(t, []string{"ctrl+shift+c"}, conflicts)
}

func TestDefaultLayout(t *testing.T) {
	layout := keymap.Default()

	require.Equal(t, keymap.ActionSplitLeftRight, layout.Lookup("ctrl+shift+o"))
	require.Equal(t, keymap.ActionSplitTopBottom, layout.Lookup("ctrl+shift+e"))
	require.Equal(t, keymap.ActionCopy, layout.Lookup("ctrl+shift+c"))
	require.Equal(t, keymap.ActionCopy, layout.Lookup("ctrl+insert"))
	require.Equal(t, keymap.ActionPaste, layout.Lookup("ctrl+shift+v"))
	require.Equal(t, keymap.ActionPaste, layout.Lookup("shift+insert"))
	require.Equal(t, keymap.ActionTab1, layout.Lookup("alt+1"))
	require.Equal(t, keymap.ActionTab9, layout.Lookup("alt+9"))
	require.Equal(t, keymap.ActionFontInc, layout.Lookup("ctrl+kp_add"))
	require.Equal(t, keymap.ActionFontDec, layout.Lookup("ctrl+kp_subtract"))
	require.Equal(t, keymap.ActionSendNewline, layout.Lookup("shift+return"))
	require.Equal(t, keymap.ActionSendNewline, layout.Lookup("ctrl+return"))
}

func TestSendNewlineActionName(t *testing.T) {
	require.Equal(t, "send_newline", keymap.ActionSendNewline.String())
}

func TestDefaultLayoutNoConflicts(t *testing.T) {
	layout := keymap.Default()

	require.Empty(t, layout.Conflicts())
}
