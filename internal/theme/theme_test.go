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
	t.Parallel()

	names := theme.BuiltinNames()

	require.ElementsMatch(t,
		[]string{"ayu", "dracula", "catppuccin-mocha", "tokyo-night", "bterm-neon", "default"},
		names)
}

func TestParsePalette(t *testing.T) {
	t.Parallel()

	p, err := theme.Parse(validPaletteTOML)
	require.NoError(t, err)

	require.Equal(t, "#282a36", p.Background)
	require.Len(t, p.Palette, 16)
	require.False(t, p.UseSystemDefault)
}

func TestParseRejectsWrongPaletteLength(t *testing.T) {
	t.Parallel()

	_, err := theme.Parse(shortPaletteTOML)
	require.Error(t, err)
	require.ErrorContains(t, err, "16")
}

func TestDefaultThemeUsesSystem(t *testing.T) {
	t.Parallel()

	p, err := theme.Builtin("default")
	require.NoError(t, err)

	require.True(t, p.UseSystemDefault)
}

func TestParseRejectsInvalidHexColor(t *testing.T) {
	t.Parallel()

	_, err := theme.Parse(`foreground = "notacolor"` + "\n" +
		`background = "#000000"` + "\n" +
		`cursor = "#ffffff"` + "\n" +
		`accent = "#ffffff"` + "\n" +
		`palette = ["#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000"]`)
	require.Error(t, err)
	require.ErrorContains(t, err, "not a valid hex color")
}

func TestParseRejectsInvalidPaletteColor(t *testing.T) {
	t.Parallel()

	_, err := theme.Parse(`foreground = "#ffffff"` + "\n" +
		`background = "#000000"` + "\n" +
		`cursor = "#ffffff"` + "\n" +
		`accent = "#ffffff"` + "\n" +
		`palette = ["#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","nope"]`)
	require.Error(t, err)
	require.ErrorContains(t, err, "not a valid hex color")
}

func TestParseAcceptsShortHex(t *testing.T) {
	t.Parallel()

	p, err := theme.Parse(`foreground = "#fff"` + "\n" +
		`background = "#000"` + "\n" +
		`cursor = "#fff"` + "\n" +
		`accent = "#fff"` + "\n" +
		`palette = ["#000","#000","#000","#000","#000","#000","#000","#000","#000","#000","#000","#000","#000","#000","#000","#000"]`)
	require.NoError(t, err)
	require.Equal(t, "#fff", p.Foreground)
}

func TestParseAcceptsHexWithAlpha(t *testing.T) {
	t.Parallel()

	p, err := theme.Parse(`foreground = "#ffffff80"` + "\n" +
		`background = "#000000ff"` + "\n" +
		`cursor = "#ffffff"` + "\n" +
		`accent = "#ffffff"` + "\n" +
		`palette = ["#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000","#000000"]`)
	require.NoError(t, err)
	require.Equal(t, "#ffffff80", p.Foreground)
}
