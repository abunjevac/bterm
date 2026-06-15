package theme_test

import (
	_ "embed"
	"testing"

	"github.com/abunjevac/bterm/internal/theme"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/valid_palette.toml
var validPaletteTOML string

//go:embed testdata/short_palette.toml
var shortPaletteTOML string

func TestBuiltinsLoad(t *testing.T) {
	names := theme.BuiltinNames()

	require.ElementsMatch(t,
		[]string{"dracula", "catppuccin-mocha", "tokyo-night", "bterm-neon", "default"},
		names)
}

func TestParsePalette(t *testing.T) {
	p, err := theme.Parse(validPaletteTOML)
	require.NoError(t, err)

	require.Equal(t, "#282a36", p.Background)
	require.Len(t, p.Palette, 16)
	require.False(t, p.UseSystemDefault)
}

func TestParseRejectsWrongPaletteLength(t *testing.T) {
	_, err := theme.Parse(shortPaletteTOML)
	require.Error(t, err)
	require.ErrorContains(t, err, "16")
}

func TestDefaultThemeUsesSystem(t *testing.T) {
	p, err := theme.Builtin("default")
	require.NoError(t, err)

	require.True(t, p.UseSystemDefault)
}
