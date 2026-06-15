package config_test

import (
	"path/filepath"
	"testing"

	_ "embed"

	"github.com/abunjevac/bterm/internal/config"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/config.toml
var testConfig string

func TestParseAppliesDefaults(t *testing.T) {
	cfg, err := config.Parse("")
	require.NoError(t, err)

	require.Equal(t, "Monospace", cfg.Font)
	require.InEpsilon(t, 12.0, cfg.FontSize, 0.001)
	require.Equal(t, "dracula", cfg.Theme)
	require.Equal(t, 5000, cfg.Scrollback)
	require.Equal(t, 1200, cfg.WindowWidth)
	require.Equal(t, 800, cfg.WindowHeight)
}

func TestParseOverrides(t *testing.T) {
	cfg, err := config.Parse(testConfig)
	require.NoError(t, err)

	require.Equal(t, "JetBrains Mono", cfg.Font)
	require.InEpsilon(t, 14.0, cfg.FontSize, 0.001)
	require.Equal(t, "tokyo-night", cfg.Theme)
	require.Equal(t, 10000, cfg.Scrollback)
}

func TestParseRejectsUnknownKeys(t *testing.T) {
	_, err := config.Parse(`notakey = true`)
	require.ErrorContains(t, err, "notakey")
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()

	cfg, err := config.Parse(`font = "Fira Code"`)
	require.NoError(t, err)

	path := filepath.Join(dir, "config.toml")

	require.NoError(t, config.Save(path, cfg))

	reloaded, err := config.LoadFile(path)
	require.NoError(t, err)

	require.Equal(t, "Fira Code", reloaded.Font)
}
