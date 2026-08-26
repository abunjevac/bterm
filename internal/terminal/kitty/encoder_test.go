package kitty

import (
	"strings"
	"testing"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
)

// --- Text keys ---

func TestEncodeKeyPlainLetterFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_a, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain 'a', got %q", r.Bytes())
	}
}

func TestEncodeKeyShiftLetterFallsThrough(t *testing.T) {
	// Shift+A without ctrl/alt should fall through (it produces text "A").
	r := EncodeKey(gdk.KEY_a, gdk.ShiftMask)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for shift+a, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlA(t *testing.T) {
	r := EncodeKey(gdk.KEY_a, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[97;5u"
	if got != want {
		t.Fatalf("ctrl+a: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlShiftA(t *testing.T) {
	r := EncodeKey(gdk.KEY_a, gdk.ControlMask|gdk.ShiftMask)
	got := string(r.Bytes())
	want := "\x1b[97;6u"
	if got != want {
		t.Fatalf("ctrl+shift+a: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyAltA(t *testing.T) {
	r := EncodeKey(gdk.KEY_a, gdk.AltMask)
	got := string(r.Bytes())
	want := "\x1b[97;3u"
	if got != want {
		t.Fatalf("alt+a: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlC(t *testing.T) {
	r := EncodeKey(gdk.KEY_c, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[99;5u"
	if got != want {
		t.Fatalf("ctrl+c: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlSpace(t *testing.T) {
	r := EncodeKey(gdk.KEY_space, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[32;5u"
	if got != want {
		t.Fatalf("ctrl+space: expected %q, got %q", want, got)
	}
}

func TestEncodeKeySpaceFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_space, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain space, got %q", r.Bytes())
	}
}

func TestEncodeKeyNumberFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_5, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain '5', got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlNumber(t *testing.T) {
	r := EncodeKey(gdk.KEY_5, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[53;5u"
	if got != want {
		t.Fatalf("ctrl+5: expected %q, got %q", want, got)
	}
}

// --- Safety exception keys (Enter, Tab, Backspace) ---

func TestEncodeKeyEnterFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Return, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain Enter, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlEnter(t *testing.T) {
	r := EncodeKey(gdk.KEY_Return, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[13;5u"
	if got != want {
		t.Fatalf("ctrl+enter: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyShiftEnter(t *testing.T) {
	r := EncodeKey(gdk.KEY_Return, gdk.ShiftMask)
	got := string(r.Bytes())
	want := "\x1b[13;2u"
	if got != want {
		t.Fatalf("shift+enter: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyTabFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Tab, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain Tab, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlTab(t *testing.T) {
	r := EncodeKey(gdk.KEY_Tab, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[9;5u"
	if got != want {
		t.Fatalf("ctrl+tab: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyBackspaceFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_BackSpace, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain Backspace, got %q", r.Bytes())
	}
}

// --- Escape ---

func TestEncodeKeyEscape(t *testing.T) {
	// Escape is always encoded (0x1b is ambiguous with CSI start).
	r := EncodeKey(gdk.KEY_Escape, 0)
	got := string(r.Bytes())
	want := "\x1b[27u"
	if got != want {
		t.Fatalf("escape: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlEscape(t *testing.T) {
	r := EncodeKey(gdk.KEY_Escape, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[27;5u"
	if got != want {
		t.Fatalf("ctrl+escape: expected %q, got %q", want, got)
	}
}

// --- Arrow keys ---

func TestEncodeKeyArrowUpFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Up, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain Up, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlUp(t *testing.T) {
	r := EncodeKey(gdk.KEY_Up, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[1;5A"
	if got != want {
		t.Fatalf("ctrl+up: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlShiftUp(t *testing.T) {
	r := EncodeKey(gdk.KEY_Up, gdk.ControlMask|gdk.ShiftMask)
	got := string(r.Bytes())
	want := "\x1b[1;6A"
	if got != want {
		t.Fatalf("ctrl+shift+up: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyArrowDownFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Down, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain Down, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlDown(t *testing.T) {
	r := EncodeKey(gdk.KEY_Down, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[1;5B"
	if got != want {
		t.Fatalf("ctrl+down: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlLeft(t *testing.T) {
	r := EncodeKey(gdk.KEY_Left, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[1;5D"
	if got != want {
		t.Fatalf("ctrl+left: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlRight(t *testing.T) {
	r := EncodeKey(gdk.KEY_Right, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[1;5C"
	if got != want {
		t.Fatalf("ctrl+right: expected %q, got %q", want, got)
	}
}

// --- Home, End, Insert, Delete, PageUp, PageDown ---

func TestEncodeKeyHomeFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Home, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain Home, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlHome(t *testing.T) {
	r := EncodeKey(gdk.KEY_Home, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[1;5H"
	if got != want {
		t.Fatalf("ctrl+home: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlEnd(t *testing.T) {
	r := EncodeKey(gdk.KEY_End, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[1;5F"
	if got != want {
		t.Fatalf("ctrl+end: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyInsertFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Insert, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain Insert, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlInsert(t *testing.T) {
	r := EncodeKey(gdk.KEY_Insert, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[2;5~"
	if got != want {
		t.Fatalf("ctrl+insert: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlDelete(t *testing.T) {
	r := EncodeKey(gdk.KEY_Delete, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[3;5~"
	if got != want {
		t.Fatalf("ctrl+delete: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlPageUp(t *testing.T) {
	r := EncodeKey(gdk.KEY_Page_Up, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[5;5~"
	if got != want {
		t.Fatalf("ctrl+page_up: expected %q, got %q", want, got)
	}
}

// --- F-keys ---

func TestEncodeKeyF1FallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_F1, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for plain F1, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlF1(t *testing.T) {
	r := EncodeKey(gdk.KEY_F1, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[1;5P"
	if got != want {
		t.Fatalf("ctrl+F1: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyCtrlF5(t *testing.T) {
	r := EncodeKey(gdk.KEY_F5, gdk.ControlMask)
	got := string(r.Bytes())
	want := "\x1b[15;5~"
	if got != want {
		t.Fatalf("ctrl+F5: expected %q, got %q", want, got)
	}
}

func TestEncodeKeyF13AlwaysEncoded(t *testing.T) {
	// F13+ uses PUA codes and should always be encoded.
	r := EncodeKey(gdk.KEY_F13, 0)
	got := string(r.Bytes())
	want := "\x1b[57376u"
	if got != want {
		t.Fatalf("F13: expected %q, got %q", want, got)
	}
}

// --- Modifier-only keys ---

func TestEncodeKeyShiftAloneFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Shift_L, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for Shift alone, got %q", r.Bytes())
	}
}

func TestEncodeKeyCtrlAloneFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Control_L, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for Ctrl alone, got %q", r.Bytes())
	}
}

func TestEncodeKeyAltAloneFallsThrough(t *testing.T) {
	r := EncodeKey(gdk.KEY_Alt_L, 0)
	if r.Bytes() != nil {
		t.Fatalf("expected nil for Alt alone, got %q", r.Bytes())
	}
}

// --- Modifier encoding ---

func TestEncodeModifiersAll(t *testing.T) {
	m := encodeModifiers(gdk.ShiftMask | gdk.ControlMask | gdk.AltMask | gdk.SuperMask)
	// 1 + shift(1) + alt(2) + ctrl(4) + super(8) = 16
	if m != 16 {
		t.Fatalf("expected 16, got %d", m)
	}
}

func TestEncodeModifiersNone(t *testing.T) {
	m := encodeModifiers(0)
	if m != 1 {
		t.Fatalf("expected 1, got %d", m)
	}
}

// --- All functional keys produce valid CSI ---

func TestEncodeKeyAllFunctionalKeysWithMods(t *testing.T) {
	for keyval := range functionalKeys {
		r := EncodeKey(keyval, gdk.ControlMask)
		if r.Bytes() == nil {
			t.Fatalf("keyval %d with ctrl produced nil bytes", keyval)
		}

		got := string(r.Bytes())
		if !strings.HasPrefix(got, "\x1b[") {
			t.Fatalf("keyval %d: expected CSI prefix, got %q", keyval, got)
		}
	}
}
