package theme

import (
	"fmt"
	"regexp"

	"github.com/BurntSushi/toml"
)

// hexColorRe matches the color formats gdk_rgba_parse accepts:
// #rgb, #rgba, #rrggbb, #rrggbbaa.
var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{3,4}$|^#[0-9a-fA-F]{6}([0-9a-fA-F]{2})?$`)

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

	for _, c := range []string{p.Foreground, p.Background, p.Cursor, p.Accent} {
		if !hexColorRe.MatchString(c) {
			return nil, fmt.Errorf("theme color %q is not a valid hex color (expected #rrggbb or #rrggbbaa)", c)
		}
	}

	for i, c := range p.Palette {
		if !hexColorRe.MatchString(c) {
			return nil, fmt.Errorf("theme palette[%d] %q is not a valid hex color (expected #rrggbb or #rrggbbaa)", i, c)
		}
	}

	return &p, nil
}
