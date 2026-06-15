package theme

import (
	"embed"
	"fmt"
	"sort"
)

//go:embed themes/*.toml
var builtinFS embed.FS

// BuiltinNames returns the names of all embedded themes.
func BuiltinNames() []string {
	entries, err := builtinFS.ReadDir("themes")
	if err != nil {
		panic(err) // embedded FS is always readable
	}

	names := make([]string, 0, len(entries))

	for _, e := range entries {
		names = append(names, e.Name()[:len(e.Name())-len(".toml")])
	}

	sort.Strings(names)

	return names
}

// Builtin parses an embedded theme by name.
func Builtin(name string) (*Palette, error) {
	data, err := builtinFS.ReadFile("themes/" + name + ".toml")
	if err != nil {
		return nil, fmt.Errorf("unknown builtin theme %q: %w", name, err)
	}

	return Parse(string(data))
}

// BuiltinSource returns the raw TOML for an embedded theme (used to scaffold files).
func BuiltinSource(name string) ([]byte, error) {
	data, err := builtinFS.ReadFile("themes/" + name + ".toml")
	if err != nil {
		return nil, fmt.Errorf("unknown builtin theme %q: %w", name, err)
	}

	return data, nil
}

// Load returns a theme by name: user file first, then builtin, then "default" fallback.
func Load(dir, name string) *Palette {
	if p, err := loadUserTheme(dir, name); err == nil {
		return p
	}

	if p, err := Builtin(name); err == nil {
		return p
	}

	p, _ := Builtin("default")

	return p
}
