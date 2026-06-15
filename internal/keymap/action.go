// Package keymap maps keyboard bindings to terminal actions.
package keymap

// Action represents a named user action that can be bound to a key combination.
type Action int

// Known actions. ActionUnknown is returned when a binding is not mapped.
const (
	ActionUnknown Action = iota
	ActionSplitLeftRight
	ActionSplitTopBottom
	ActionResizeLeft
	ActionResizeRight
	ActionResizeUp
	ActionResizeDown
	ActionFocusLeft
	ActionFocusRight
	ActionFocusUp
	ActionFocusDown
	ActionNewTabEnd
	ActionNewTabAfter
	ActionClosePane
	ActionCloseTab
	ActionCopy
	ActionPaste
	ActionFontInc
	ActionFontDec
	ActionFontReset
	ActionOpenConfig
	ActionTab1
	ActionTab2
	ActionTab3
	ActionTab4
	ActionTab5
	ActionTab6
	ActionTab7
	ActionTab8
	ActionTab9
)

// actionNames maps each Action to its TOML key (snake_case).
//
//nolint:gochecknoglobals
var actionNames = map[Action]string{
	ActionSplitLeftRight: "split_left_right",
	ActionSplitTopBottom: "split_top_bottom",
	ActionResizeLeft:     "resize_left",
	ActionResizeRight:    "resize_right",
	ActionResizeUp:       "resize_up",
	ActionResizeDown:     "resize_down",
	ActionFocusLeft:      "focus_left",
	ActionFocusRight:     "focus_right",
	ActionFocusUp:        "focus_up",
	ActionFocusDown:      "focus_down",
	ActionNewTabEnd:      "new_tab_end",
	ActionNewTabAfter:    "new_tab_after",
	ActionClosePane:      "close_pane",
	ActionCloseTab:       "close_tab",
	ActionCopy:           "copy",
	ActionPaste:          "paste",
	ActionFontInc:        "font_inc",
	ActionFontDec:        "font_dec",
	ActionFontReset:      "font_reset",
	ActionOpenConfig:     "open_config",
	ActionTab1:           "tab_1",
	ActionTab2:           "tab_2",
	ActionTab3:           "tab_3",
	ActionTab4:           "tab_4",
	ActionTab5:           "tab_5",
	ActionTab6:           "tab_6",
	ActionTab7:           "tab_7",
	ActionTab8:           "tab_8",
	ActionTab9:           "tab_9",
}

// nameToAction is the reverse of actionNames, built once at init.
//
//nolint:gochecknoglobals
var nameToAction map[string]Action

//nolint:gochecknoinits
func init() {
	nameToAction = make(map[string]Action, len(actionNames))

	for a, name := range actionNames {
		nameToAction[name] = a
	}
}

// String returns the snake_case TOML key for a, or "unknown" for ActionUnknown.
func (a Action) String() string {
	if name, ok := actionNames[a]; ok {
		return name
	}

	return "unknown"
}
