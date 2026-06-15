package theme

import (
	"os"
	"path/filepath"
)

func loadUserTheme(dir, name string) (*Palette, error) {
	data, err := os.ReadFile(filepath.Join(dir, "themes", name+".toml"))
	if err != nil {
		return nil, err //nolint:wrapcheck // sentinel for fallback chain
	}

	return Parse(string(data))
}
