package vte

import "testing"

func TestKittyParserSetFlagsReplace(t *testing.T) {
	t.Parallel()

	var p kittyParser

	p.Filter([]byte("\x1b[=1;1u"))

	if !p.Disambiguate() {
		t.Fatal("expected disambiguate flag after CSI = 1;1 u")
	}
}

func TestKittyParserSetFlagsUnion(t *testing.T) {
	t.Parallel()

	var p kittyParser

	p.Filter([]byte("\x1b[=1;1u"))
	p.Filter([]byte("\x1b[=2;2u"))

	if p.Flags() != 3 {
		t.Fatalf("expected flags=3 after union, got %d", p.Flags())
	}
}

func TestKittyParserSetFlagsSubtract(t *testing.T) {
	t.Parallel()

	var p kittyParser

	p.Filter([]byte("\x1b[=3;1u"))
	p.Filter([]byte("\x1b[=1;3u"))

	if p.Flags() != 2 {
		t.Fatalf("expected flags=2 after subtract, got %d", p.Flags())
	}
}

func TestKittyParserQueryResponse(t *testing.T) {
	t.Parallel()

	var p kittyParser

	p.Filter([]byte("\x1b[=1;1u"))
	result := p.Filter([]byte("\x1b[?u"))

	if string(result.response) != "\x1b[?1u" {
		t.Fatalf("expected \\x1b[?1u, got %q", string(result.response))
	}
}

func TestKittyParserResponseDoesNotLoop(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// A response sequence (CSI ? <flags> u) should be stripped without
	// generating a new response, preventing an echo loop.
	result := p.Filter([]byte("\x1b[?0u"))

	if len(result.response) != 0 {
		t.Fatalf("expected no response for CSI ? 0 u, got %q", string(result.response))
	}

	if string(result.out) != "" {
		t.Fatalf("expected empty output, got %q", string(result.out))
	}
}

func TestKittyParserPushPop(t *testing.T) {
	t.Parallel()

	var p kittyParser

	p.Filter([]byte("\x1b[=1;1u"))
	p.Filter([]byte("\x1b[>2u"))

	if p.Flags() != 2 {
		t.Fatalf("expected flags=2 after push, got %d", p.Flags())
	}

	p.Filter([]byte("\x1b[<1u"))

	if p.Flags() != 1 {
		t.Fatalf("expected flags=1 after pop, got %d", p.Flags())
	}
}

func TestKittyParserStripsSequences(t *testing.T) {
	t.Parallel()

	var p kittyParser

	result := p.Filter([]byte("before\x1b[=1;1uafter"))

	if string(result.out) != "beforeafter" {
		t.Fatalf("expected 'beforeafter', got %q", string(result.out))
	}
}

func TestKittyParserPassesThroughNonKittyCSI(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// Regular CSI sequence (cursor forward) should pass through.
	input := "\x1b[3C"
	result := p.Filter([]byte(input))

	if string(result.out) != input {
		t.Fatalf("expected %q, got %q", input, string(result.out))
	}
}

func TestKittyParserPassesThroughCSIUWithoutPrefix(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// CSI u without kitty prefix should pass through (not a kitty sequence).
	input := "\x1b[97u"
	result := p.Filter([]byte(input))

	if string(result.out) != input {
		t.Fatalf("expected %q, got %q", input, string(result.out))
	}
}

func TestKittyParserHandlesSplitSequence(t *testing.T) {
	t.Parallel()

	var p kittyParser

	result := p.Filter([]byte("a\x1b[=1;1"))
	if string(result.out) != "a" {
		t.Fatalf("expected 'a', got %q", string(result.out))
	}

	result = p.Filter([]byte("ub"))
	if string(result.out) != "b" {
		t.Fatalf("expected 'b', got %q", string(result.out))
	}

	if !p.Disambiguate() {
		t.Fatal("expected disambiguate flag after split sequence")
	}
}

func TestKittyParserMultipleSequences(t *testing.T) {
	t.Parallel()

	var p kittyParser

	result := p.Filter([]byte("\x1b[=1;1u\x1b[?u"))

	if string(result.out) != "" {
		t.Fatalf("expected empty output, got %q", string(result.out))
	}

	if string(result.response) != "\x1b[?1u" {
		t.Fatalf("expected query response, got %q", string(result.response))
	}

	if !p.Disambiguate() {
		t.Fatal("expected disambiguate flag")
	}
}

func TestKittyParserDefaultModeIsReplace(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// Without explicit mode, default is 1 (replace).
	p.Filter([]byte("\x1b[=3;1u"))
	p.Filter([]byte("\x1b[=1u"))

	if p.Flags() != 1 {
		t.Fatalf("expected flags=1 with default mode, got %d", p.Flags())
	}
}

func TestKittyParserPopDefaultCount(t *testing.T) {
	t.Parallel()

	var p kittyParser

	p.Filter([]byte("\x1b[=1;1u"))
	p.Filter([]byte("\x1b[>2u"))
	p.Filter([]byte("\x1b[>4u"))

	// CSI < u without count should pop 1.
	p.Filter([]byte("\x1b[<u"))

	if p.Flags() != 2 {
		t.Fatalf("expected flags=2 after default pop, got %d", p.Flags())
	}
}

func TestKittyParserPassesThroughDECPrivateModes(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// DEC private mode sequences use CSI ? prefix but end with h/l,
	// not u. They must pass through to VTE.
	inputs := []string{
		"\x1b[?1000h", // mouse tracking
		"\x1b[?2004h", // bracketed paste
		"\x1b[?1049h", // alternate screen
		"\x1b[?1000l", // disable mouse tracking
		"\x1b[?2004l", // disable bracketed paste
		"\x1b[?1049l", // disable alternate screen
		"\x1b[?25h",   // show cursor
		"\x1b[?25l",   // hide cursor
	}

	for _, input := range inputs {
		result := p.Filter([]byte(input))
		if string(result.out) != input {
			t.Fatalf("expected %q to pass through, got %q", input, string(result.out))
		}
	}
}

func TestKittyParserPassesThroughSGRMouse(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// SGR mouse events use CSI < prefix but end with M/m, not u.
	input := "\x1b[<0;10;20M"
	result := p.Filter([]byte(input))

	if string(result.out) != input {
		t.Fatalf("expected SGR mouse to pass through, got %q", string(result.out))
	}
}

func TestKittyParserPassesThroughXtermModifiers(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// xterm modifier reports use CSI > prefix but end with c, not u.
	input := "\x1b[>0c"
	result := p.Filter([]byte(input))

	if string(result.out) != input {
		t.Fatalf("expected xterm modifier report to pass through, got %q", string(result.out))
	}
}

func TestKittyParserPassesThroughMixedContent(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// Regular output + DEC private mode + kitty negotiation + more output.
	input := "hello\x1b[?1049hworld\x1b[=1;1uend"
	result := p.Filter([]byte(input))

	want := "hello\x1b[?1049hworldend"
	if string(result.out) != want {
		t.Fatalf("expected %q, got %q", want, string(result.out))
	}

	if !p.Disambiguate() {
		t.Fatal("expected disambiguate flag after mixed content")
	}
}

func TestKittyParserPassesThroughSGRColors(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// SGR color sequences (CSI ... m) must pass through intact.
	inputs := []string{
		"\x1b[0m",              // reset
		"\x1b[31m",             // red
		"\x1b[1;32m",           // bold green
		"\x1b[38;5;208m",       // 256-color orange
		"\x1b[38;2;255;128;0m", // truecolor
	}

	for _, input := range inputs {
		result := p.Filter([]byte(input))
		if string(result.out) != input {
			t.Fatalf("expected %q to pass through, got %q", input, string(result.out))
		}
	}
}

func TestKittyParserPassesThroughCursorMovement(t *testing.T) {
	t.Parallel()

	var p kittyParser

	// Cursor movement sequences must pass through intact.
	inputs := []string{
		"\x1b[H",      // cursor home
		"\x1b[10;20H", // cursor to row 10, col 20
		"\x1b[2J",     // clear screen
		"\x1b[K",      // clear line
		"\x1b[5C",     // cursor forward 5
		"\x1b[3A",     // cursor up 3
	}

	for _, input := range inputs {
		result := p.Filter([]byte(input))
		if string(result.out) != input {
			t.Fatalf("expected %q to pass through, got %q", input, string(result.out))
		}
	}
}
