// Package theme defines color palettes for bterm and loads them from TOML.
package theme

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Palette is a full terminal color scheme. When UseSystemDefault is true, the
// terminal keeps VTE's built-in palette and the host GTK theme (the "default" theme).
type Palette struct {
	Foreground       string   `toml:"foreground"`
	Background       string   `toml:"background"`
	Cursor           string   `toml:"cursor"`
	Accent           string   `toml:"accent"`
	Palette          []string `toml:"palette"` // exactly 16 ANSI colors
	UseSystemDefault bool     `toml:"use_system_default"`
}

// Parse decodes a theme TOML. A theme with use_system_default = true may omit colors;
// any other theme must define a 16-entry palette.
func Parse(data string) (*Palette, error) {
	var p Palette

	if _, err := toml.Decode(data, &p); err != nil {
		return nil, fmt.Errorf("decode theme: %w", err)
	}

	if p.UseSystemDefault {
		return &p, nil
	}

	if len(p.Palette) != 16 {
		return nil, fmt.Errorf("theme palette must have 16 colors, got %d", len(p.Palette))
	}

	return &p, nil
}
