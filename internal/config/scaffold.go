package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/theme"
)

// Scaffold creates the config dir, a default config.toml, and the builtin theme files
// if they do not already exist. Existing files are never overwritten.
func Scaffold(dir string) error {
	if err := os.MkdirAll(filepath.Join(dir, "themes"), 0o755); err != nil {
		return fmt.Errorf("scaffold config dir: %w", err)
	}

	if err := writeDefaultConfig(dir); err != nil {
		return err
	}

	if err := writeBuiltinThemes(dir); err != nil {
		return err
	}

	if err := writeDefaultKeymap(dir); err != nil {
		return err
	}

	return nil
}

func writeDefaultConfig(dir string) error {
	cfgPath := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
		return nil
	}

	def, _ := Parse("")

	return Save(cfgPath, def)
}

func writeBuiltinThemes(dir string) error {
	for _, name := range theme.BuiltinNames() {
		dst := filepath.Join(dir, "themes", name+".toml")
		if _, err := os.Stat(dst); !os.IsNotExist(err) {
			continue
		}

		src, err := theme.BuiltinSource(name)
		if err != nil {
			return fmt.Errorf("read builtin theme %s: %w", name, err)
		}

		if err := os.WriteFile(dst, src, 0o644); err != nil { //nolint:gosec
			return fmt.Errorf("write theme %s: %w", name, err)
		}
	}

	return nil
}

func writeDefaultKeymap(dir string) error {
	dst := filepath.Join(dir, "keymap.toml")
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		return nil
	}

	var content = `# bterm keymap — action = "binding" (or = ["b1","b2"])` + "\n" + keymap.DefaultKeymapTOML

	if err := os.WriteFile(dst, []byte(content), 0o644); err != nil { //nolint:gosec
		return fmt.Errorf("write keymap: %w", err)
	}

	return nil
}
