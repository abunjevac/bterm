package keymap

import (
	_ "embed"
	"fmt"
	"sort"

	"github.com/BurntSushi/toml"
)

// DefaultKeymapTOML is the canonical binding set, kept in sync with
// config.writeDefaultKeymap.
//
//go:embed default_keymap.toml
var DefaultKeymapTOML string

// Layout holds the resolved binding → action mapping.
type Layout struct {
	// bindings maps normalized binding string → Action
	bindings map[string]Action

	// conflicts stores bindings that were claimed by more than one action
	conflicts []string
}

// Lookup returns the Action for the given normalized binding string,
// or ActionUnknown if no binding matches.
func (l *Layout) Lookup(binding string) Action {
	return l.bindings[binding]
}

// LayoutEntry pairs an action with the key bindings assigned to it.
type LayoutEntry struct {
	Action Action
	Keys   []string
}

// Entries returns all bound actions as a slice sorted by action name, each with
// its key bindings in sorted order. Used for displaying the shortcuts dialog.
func (l *Layout) Entries() []LayoutEntry {
	byAction := make(map[Action][]string, len(l.bindings))

	for binding, action := range l.bindings {
		byAction[action] = append(byAction[action], binding)
	}

	for action := range byAction {
		sort.Strings(byAction[action])
	}

	entries := make([]LayoutEntry, 0, len(byAction))

	for action, keys := range byAction {
		entries = append(entries, LayoutEntry{Action: action, Keys: keys})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Action.String() < entries[j].Action.String()
	})

	return entries
}

// Conflicts returns all binding strings that are bound to more than one action,
// in sorted order. Returns nil when there are no conflicts.
func (l *Layout) Conflicts() []string {
	if len(l.conflicts) == 0 {
		return nil
	}

	return l.conflicts
}

// rawKeymap is an intermediate TOML representation where each value is
// either a string or a list of strings. BurntSushi/toml decodes into
// any when the type is ambiguous, so we handle both cases manually.
type rawKeymap map[string]any

// Parse parses a TOML keymap and returns a Layout.
// Unknown action names are silently ignored so future versions stay compatible.
func Parse(data string) (*Layout, error) {
	var raw rawKeymap

	if _, err := toml.Decode(data, &raw); err != nil {
		return nil, fmt.Errorf("parse keymap: %w", err)
	}

	return buildLayout(raw)
}

// Default returns a Layout built from the same bindings that Scaffold writes
// to keymap.toml on first run.
func Default() *Layout {
	layout, err := Parse(DefaultKeymapTOML)
	if err != nil {
		// DefaultKeymapTOML is a compile-time constant — panic is appropriate
		panic(fmt.Sprintf("keymap: invalid default keymap: %v", err))
	}

	return layout
}

// buildLayout converts raw TOML map into a Layout, normalizing all bindings
// and recording any that are claimed by more than one action.
func buildLayout(raw rawKeymap) (*Layout, error) {
	// first pass: collect all (binding → []action) claims
	claims := make(map[string][]Action)

	for key, val := range raw {
		action, ok := nameToAction[key]
		if !ok {
			continue // unknown action name — skip
		}

		bindings, err := toBindingSlice(key, val)
		if err != nil {
			return nil, err
		}

		for _, b := range bindings {
			norm, err := Normalize(b)
			if err != nil {
				return nil, fmt.Errorf("keymap action %q binding %q: %w", key, b, err)
			}

			claims[norm] = append(claims[norm], action)
		}
	}

	// second pass: build final map and collect conflicts
	bindings := make(map[string]Action, len(claims))

	var conflicts []string

	for norm, actions := range claims {
		if len(actions) > 1 {
			conflicts = append(conflicts, norm)

			continue
		}

		bindings[norm] = actions[0]
	}

	sort.Strings(conflicts)

	return &Layout{bindings: bindings, conflicts: conflicts}, nil
}

// toBindingSlice coerces a TOML value (string or []any) into []string.
func toBindingSlice(key string, val any) ([]string, error) {
	switch v := val.(type) {
	case string:
		return []string{v}, nil
	case []any:
		out := make([]string, 0, len(v))

		for i, elem := range v {
			s, ok := elem.(string)
			if !ok {
				return nil, fmt.Errorf("keymap action %q: binding[%d] is not a string", key, i)
			}

			out = append(out, s)
		}

		return out, nil
	default:
		return nil, fmt.Errorf("keymap action %q: expected string or array, got %T", key, val)
	}
}
