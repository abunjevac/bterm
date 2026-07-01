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

	require.Equal(t, "Google Sans Code", cfg.Font)
	require.InEpsilon(t, 16.0, cfg.FontSize, 0.001)
	require.Equal(t, "ayu", cfg.Theme)
	require.Equal(t, 5000, cfg.Scrollback)
	require.Equal(t, 180, cfg.WindowColumns)
	require.Equal(t, 40, cfg.WindowRows)
	require.Equal(t, "dbus", cfg.TerminalNotificationMethod)
}

func TestParseOverrides(t *testing.T) {
	cfg, err := config.Parse(testConfig)
	require.NoError(t, err)

	require.Equal(t, "JetBrains Mono", cfg.Font)
	require.InEpsilon(t, 14.0, cfg.FontSize, 0.001)
	require.Equal(t, "tokyo-night", cfg.Theme)
	require.Equal(t, 10000, cfg.Scrollback)
	require.Equal(t, "off", cfg.TerminalNotificationMethod)
}

func TestParseRejectsUnknownKeys(t *testing.T) {
	_, err := config.Parse(`notakey = true`)
	require.ErrorContains(t, err, "notakey")
}

func TestParseRejectsInvalidTerminalNotificationMethod(t *testing.T) {
	_, err := config.Parse(`terminal_notification_method = "notify-send"`)
	require.ErrorContains(t, err, "invalid terminal_notification_method")
}

func TestParseRejectsOldTerminalNotificationsKey(t *testing.T) {
	_, err := config.Parse(`terminal_notifications = false`)
	require.ErrorContains(t, err, "terminal_notifications")
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
