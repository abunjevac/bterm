package keymap

import (
	"fmt"
	"slices"
	"strings"
)

// modifierOrder defines canonical modifier precedence (lower index = first).
//
//nolint:gochecknoglobals
var modifierOrder = []string{"ctrl", "shift", "alt", "super"}

// modifierSet is a fast lookup for known modifier names.
//
//nolint:gochecknoglobals
var modifierSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(modifierOrder))

	for _, mod := range modifierOrder {
		m[mod] = struct{}{}
	}

	return m
}()

// Normalize returns the canonical form of a binding string.
// Modifier order: ctrl, shift, alt, super. The key name is lowercased.
// Returns an error if the binding is empty or contains no non-modifier component.
func Normalize(binding string) (string, error) {
	if binding == "" {
		return "", fmt.Errorf("normalize: binding is empty")
	}

	parts := strings.Split(strings.ToLower(binding), "+")

	var (
		mods []string
		key  string
	)

	for _, p := range parts {
		if _, isMod := modifierSet[p]; isMod {
			mods = append(mods, p)
		} else {
			if key != "" {
				// two non-modifier tokens — treat the first as a modifier-like part
				// (shouldn't happen with well-formed bindings, surface the input as-is)
				return "", fmt.Errorf("normalize: ambiguous binding %q", binding)
			}

			key = p
		}
	}

	if key == "" {
		return "", fmt.Errorf("normalize: binding %q has no key component", binding)
	}

	// sort modifiers into canonical order, preserving only those present
	var ordered []string

	for _, mod := range modifierOrder {
		if slices.Contains(mods, mod) {
			ordered = append(ordered, mod)
		}
	}

	ordered = append(ordered, key)

	return strings.Join(ordered, "+"), nil
}
