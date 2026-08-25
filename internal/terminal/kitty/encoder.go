package kitty

import (
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
)

// EncodeResult holds the encoded key event.
type EncodeResult struct {
	bytes []byte // CSI sequence to send to the shell, or nil if the key should be handled by the terminal widget (legacy encoding)
}

// Bytes returns the encoded bytes, or nil if the key should fall through
// to the terminal widget.
func (e EncodeResult) Bytes() []byte { return e.bytes }

// keyEncoding describes how to encode a functional key.
type keyEncoding struct {
	number int  // CSI parameter number
	final  byte // CSI final byte: 'u', '~', or a letter (A,B,C,D,H,F,P,Q,S)
}

// functionalKeys maps GDK keyvals to kitty keyboard protocol encodings.
// Values from https://sw.kovidgoyal.net/kitty/keyboard-protocol/ functional
// key definitions table.
//
//nolint:gochecknoglobals // static lookup table
var functionalKeys = map[uint]keyEncoding{
	gdk.KEY_Escape:       {27, 'u'},
	gdk.KEY_Return:       {13, 'u'},
	gdk.KEY_KP_Enter:     {13, 'u'},
	gdk.KEY_Tab:          {9, 'u'},
	gdk.KEY_ISO_Left_Tab: {9, 'u'},
	gdk.KEY_BackSpace:    {127, 'u'},
	gdk.KEY_Insert:       {2, '~'},
	gdk.KEY_Delete:       {3, '~'},
	gdk.KEY_Left:         {1, 'D'},
	gdk.KEY_Right:        {1, 'C'},
	gdk.KEY_Up:           {1, 'A'},
	gdk.KEY_Down:         {1, 'B'},
	gdk.KEY_Page_Up:      {5, '~'},
	gdk.KEY_Page_Down:    {6, '~'},
	gdk.KEY_Home:         {1, 'H'},
	gdk.KEY_End:          {1, 'F'},
	gdk.KEY_F1:           {1, 'P'},
	gdk.KEY_F2:           {1, 'Q'},
	gdk.KEY_F3:           {13, '~'},
	gdk.KEY_F4:           {1, 'S'},
	gdk.KEY_F5:           {15, '~'},
	gdk.KEY_F6:           {17, '~'},
	gdk.KEY_F7:           {18, '~'},
	gdk.KEY_F8:           {19, '~'},
	gdk.KEY_F9:           {20, '~'},
	gdk.KEY_F10:          {21, '~'},
	gdk.KEY_F11:          {23, '~'},
	gdk.KEY_F12:          {24, '~'},
	gdk.KEY_F13:          {57376, 'u'},
	gdk.KEY_F14:          {57377, 'u'},
	gdk.KEY_F15:          {57378, 'u'},
	gdk.KEY_F16:          {57379, 'u'},
	gdk.KEY_F17:          {57380, 'u'},
	gdk.KEY_F18:          {57381, 'u'},
	gdk.KEY_F19:          {57382, 'u'},
	gdk.KEY_F20:          {57383, 'u'},
	gdk.KEY_F21:          {57384, 'u'},
	gdk.KEY_F22:          {57385, 'u'},
	gdk.KEY_F23:          {57386, 'u'},
	gdk.KEY_F24:          {57387, 'u'},
	gdk.KEY_KP_0:         {57399, 'u'},
	gdk.KEY_KP_1:         {57400, 'u'},
	gdk.KEY_KP_2:         {57401, 'u'},
	gdk.KEY_KP_3:         {57402, 'u'},
	gdk.KEY_KP_4:         {57403, 'u'},
	gdk.KEY_KP_5:         {57404, 'u'},
	gdk.KEY_KP_6:         {57405, 'u'},
	gdk.KEY_KP_7:         {57406, 'u'},
	gdk.KEY_KP_8:         {57407, 'u'},
	gdk.KEY_KP_9:         {57408, 'u'},
	gdk.KEY_KP_Decimal:   {57409, 'u'},
	gdk.KEY_KP_Divide:    {57410, 'u'},
	gdk.KEY_KP_Multiply:  {57411, 'u'},
	gdk.KEY_KP_Subtract:  {57412, 'u'},
	gdk.KEY_KP_Add:       {57413, 'u'},
	gdk.KEY_KP_Equal:     {57415, 'u'},
	gdk.KEY_KP_Separator: {57416, 'u'},
}

// EncodeKey encodes a key press as a kitty keyboard protocol sequence.
// Returns nil bytes when the key should fall through to VTE's legacy
// encoding.
//
// Only keys that need disambiguation are encoded:
//   - Esc (with or without modifiers)
//   - Text keys with ctrl, alt, or ctrl+shift modifiers
//   - Functional keys with modifiers
//   - F13+ and non-text keypad keys (VTE may not handle these)
//
// Keys that fall through to VTE:
//   - Plain text keys without ctrl/alt
//   - Enter, Tab, Backspace without modifiers (safety exception)
//   - Plain functional keys without modifiers (arrows, F1-F12, etc.)
//   - Modifier-only presses (Shift, Ctrl, Alt by themselves)
//
// Parameters:
//   - keyval: GDK keyval of the pressed key
//   - state: GDK modifier state
func EncodeKey(keyval uint, state gdk.ModifierType) EncodeResult {
	// never encode modifier-only key presses
	if isModifierKey(keyval) {
		return EncodeResult{}
	}

	mods := encodeModifiers(state)
	hasCtrlAlt := state&(gdk.ControlMask|gdk.AltMask) != 0

	// check if it's a functional key with a known encoding
	if enc, ok := functionalKeys[keyval]; ok {
		// plain functional keys without modifiers fall through to VTE,
		// except Esc and F13+ which always need encoding
		if mods == 1 && !alwaysEncode(keyval) {
			return EncodeResult{}
		}

		return EncodeResult{bytes: encodeSequence(enc, mods)}
	}

	// text keys: encode only when ctrl or alt is pressed
	if !hasCtrlAlt {
		return EncodeResult{}
	}

	code := int(gdk.KeyvalToLower(keyval))

	if code < 0x20 || code == 0x7f {
		// control codes (including space=32 which is fine, and
		// Backspace=127 handled above). Don't encode.
		if keyval == gdk.KEY_space {
			code = 32
		} else {
			return EncodeResult{}
		}
	}

	return EncodeResult{bytes: encodeSequence(keyEncoding{code, 'u'}, mods)}
}

// encodeSequence builds a CSI sequence from a key encoding and modifier
// value. Format depends on the final byte:
//   - Letter form (A,B,C,D,H,F,P,Q,S): "ESC[1;mods{final}" with mods,
//     "ESC[{final}" without mods (omit the 1 per spec).
//   - Tilde form (~): "ESC[{number};mods~" with mods, "ESC[{number}~" without.
//   - U form (u): "ESC[{number};modsu" with mods, "ESC[{number}u" without.
func encodeSequence(enc keyEncoding, mods int) []byte {
	var seq strings.Builder

	seq.WriteByte(0x1b)
	seq.WriteByte('[')

	switch enc.final {
	case 'u':
		seq.WriteString(strconv.Itoa(enc.number))

		if mods > 1 {
			seq.WriteByte(';')
			seq.WriteString(strconv.Itoa(mods))
		}

		seq.WriteByte('u')
	case '~':
		seq.WriteString(strconv.Itoa(enc.number))

		if mods > 1 {
			seq.WriteByte(';')
			seq.WriteString(strconv.Itoa(mods))
		}

		seq.WriteByte('~')
	default:
		// letter form: A, B, C, D, H, F, P, Q, S
		if mods > 1 {
			seq.WriteByte('1')
			seq.WriteByte(';')
			seq.WriteString(strconv.Itoa(mods))
		}

		seq.WriteByte(enc.final)
	}

	return []byte(seq.String())
}

// encodeModifiers converts GDK modifier state to the kitty modifier value
// (1 + bitfield). Uses addition because the base value 1 overlaps with
// ModShift's bit value of 1.
func encodeModifiers(state gdk.ModifierType) int {
	m := 1 // default: no modifiers

	if state&gdk.ShiftMask != 0 {
		m += ModShift
	}

	if state&gdk.AltMask != 0 {
		m += ModAlt
	}

	if state&gdk.ControlMask != 0 {
		m += ModCtrl
	}

	if state&gdk.SuperMask != 0 {
		m += ModSuper
	}

	if state&gdk.HyperMask != 0 {
		m += ModHyper
	}

	if state&gdk.MetaMask != 0 {
		m += ModMeta
	}

	return m
}

// isModifierKey reports whether keyval is a modifier key (Shift, Ctrl,
// Alt, Super, etc.). These should never be encoded as key events.
func isModifierKey(keyval uint) bool {
	switch keyval {
	case gdk.KEY_Shift_L, gdk.KEY_Shift_R,
		gdk.KEY_Control_L, gdk.KEY_Control_R,
		gdk.KEY_Alt_L, gdk.KEY_Alt_R,
		gdk.KEY_Super_L, gdk.KEY_Super_R,
		gdk.KEY_Hyper_L, gdk.KEY_Hyper_R,
		gdk.KEY_Meta_L, gdk.KEY_Meta_R,
		gdk.KEY_ISO_Level3_Shift, gdk.KEY_ISO_Level5_Shift,
		gdk.KEY_Caps_Lock, gdk.KEY_Scroll_Lock, gdk.KEY_Num_Lock:
		return true
	default:
		return false
	}
}

// alwaysEncode reports whether a functional key should always be encoded,
// even without modifiers. Esc needs encoding because 0x1b is ambiguous.
// F13+ and non-text keypad keys need encoding because VTE may not handle
// them correctly.
func alwaysEncode(keyval uint) bool {
	if keyval == gdk.KEY_Escape {
		return true
	}

	if enc, ok := functionalKeys[keyval]; ok {
		// PUA-encoded keys (F13+, keypad, media, etc.) always need encoding
		return enc.final == 'u' && enc.number >= 57344
	}

	return false
}

// FormatQueryResponse builds the CSI ? flags u response for a capability query.
func FormatQueryResponse(flags int) string {
	return "\x1b[?" + strconv.Itoa(flags) + "u"
}
