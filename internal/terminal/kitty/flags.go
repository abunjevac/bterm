// Package kitty implements the kitty keyboard protocol state machine and
// key encoder. See https://sw.kovidgoyal.net/kitty/keyboard-protocol/.
package kitty

// Progressive enhancement flags.
const (
	// FlagDisambiguate causes Esc, alt+key, ctrl+key, ctrl+alt+key,
	// shift+alt+key, and modified Enter/Tab/Backspace to be reported as
	// CSI u sequences instead of legacy encodings.
	FlagDisambiguate = 1
	// FlagReportEvents reports press, repeat, and release events.
	FlagReportEvents = 2
	// FlagReportAlternate includes shifted and base-layout key variants.
	FlagReportAlternate = 4
	// FlagReportAll encodes all keys as CSI u, even text-producing keys.
	FlagReportAll = 8
	// FlagReportText includes the associated text as unicode codepoints.
	FlagReportText = 16
)

// Modifier bit values as defined by the kitty keyboard protocol.
const (
	ModShift   = 1
	ModAlt     = 2
	ModCtrl    = 4
	ModSuper   = 8
	ModHyper   = 16
	ModMeta    = 32
	ModCapsLoc = 64
	ModNumLoc  = 128
)
