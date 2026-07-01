// Package config loads and persists bterm settings, themes, and keymap files
// from the config directory (default ~/.config/bterm).
package config

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	TerminalNotificationDBus = "dbus"
	TerminalNotificationOff  = "off"
)

// Config holds top-level bterm settings (config.toml).
type Config struct {
	Font                       string   `toml:"font"`
	FontSize                   float64  `toml:"font_size"`
	Theme                      string   `toml:"theme"`
	Shell                      string   `toml:"shell"`
	ShellArgs                  []string `toml:"shell_args"`
	Scrollback                 int      `toml:"scrollback"`
	WindowColumns              int      `toml:"window_columns"`
	WindowRows                 int      `toml:"window_rows"`
	Title                      string   `toml:"title"`
	TerminalNotificationMethod string   `toml:"terminal_notification_method"`
}

// Parse decodes config.toml content, rejecting unknown keys, then applies defaults.
func Parse(data string) (*Config, error) {
	var cfg Config

	meta, err := toml.Decode(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, 0, len(undecoded))

		for _, k := range undecoded {
			keys = append(keys, k.String())
		}

		return nil, fmt.Errorf("unknown config keys: %s", strings.Join(keys, ", "))
	}

	applyDefaults(&cfg)

	if err := validateTerminalNotificationMethod(cfg.TerminalNotificationMethod); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	cfg.Font = cmp.Or(cfg.Font, "Google Sans Code")
	cfg.FontSize = cmp.Or(cfg.FontSize, 16.0)
	cfg.Theme = cmp.Or(cfg.Theme, "ayu")
	cfg.Scrollback = cmp.Or(cfg.Scrollback, 5000)
	cfg.WindowColumns = cmp.Or(cfg.WindowColumns, 180)
	cfg.WindowRows = cmp.Or(cfg.WindowRows, 40)
	cfg.Title = cmp.Or(cfg.Title, "bterm")

	cfg.TerminalNotificationMethod = cmp.Or(cfg.TerminalNotificationMethod, TerminalNotificationDBus)
}

func validateTerminalNotificationMethod(method string) error {
	switch method {
	case TerminalNotificationDBus, TerminalNotificationOff:
		return nil
	default:
		return fmt.Errorf(
			"invalid terminal_notification_method %q (expected %q or %q)",
			method,
			TerminalNotificationDBus,
			TerminalNotificationOff,
		)
	}
}

// LoadFile reads and parses a config.toml at path.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return Parse(string(data))
}

// Save writes cfg to path as TOML.
func Save(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config: %w", err)
	}

	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	return nil
}

// ResolveDir returns the config dir, honoring an explicit override or ~/.config/bterm.
func ResolveDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}

	return filepath.Join(base, "bterm"), nil
}

// Load ensures the config dir is scaffolded, then loads config.toml into a Bundle.
func Load(dir string) (*Bundle, error) {
	if err := Scaffold(dir); err != nil {
		return nil, err
	}

	cfg, err := LoadFile(filepath.Join(dir, "config.toml"))
	if err != nil {
		return nil, err
	}

	return &Bundle{Config: cfg, Dir: dir}, nil
}
