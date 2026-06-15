package config

import (
	"os"
	"path/filepath"

	"github.com/abunjevac/bterm/internal/keymap"
)

// Bundle is everything ui.Run needs: settings plus the config dir for theme/keymap loads.
type Bundle struct {
	Config *Config
	Dir    string
}

// LoadKeymap reads keymap.toml from the config dir. If the file is missing or
// unparseable, the built-in default layout is returned so the application
// always has a working keymap.
func (b *Bundle) LoadKeymap() *keymap.Layout {
	data, err := os.ReadFile(filepath.Join(b.Dir, "keymap.toml"))
	if err != nil {
		return keymap.Default()
	}

	layout, err := keymap.Parse(string(data))
	if err != nil {
		return keymap.Default()
	}

	return layout
}
