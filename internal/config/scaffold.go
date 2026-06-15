package config

import (
	"fmt"
	"os"
)

// Scaffold ensures dir exists. Default-file writing is added in the theme/keymap tasks.
func Scaffold(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("scaffold config dir: %w", err)
	}

	return nil
}
