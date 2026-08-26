package vte

import (
	"testing"

	"github.com/abunjevac/bterm/internal/terminal"
	"github.com/stretchr/testify/require"
)

func TestOSCParserStrips777Notification(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("before\x1b]777;notify;Title;Message\x07after"))

	require.Equal(t, "beforeafter", string(result.out))
	require.Equal(t, []terminal.Notification{{Title: "Title", Message: "Message"}}, result.notes)
	require.Empty(t, result.clipboardText)
}

func TestOSCParserStripsOSC9Notification(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("\x1b]9;Build complete\x1b\\"))

	require.Empty(t, result.out)
	require.Equal(t, []terminal.Notification{{Message: "Build complete"}}, result.notes)
	require.Empty(t, result.clipboardText)
}

func TestOSCParserLeavesProgressAndUnknownOSC(t *testing.T) {
	var p oscParser

	input := "\x1b]9;4;1;50\x07\x1b]0;title\x07"
	result := p.Filter([]byte(input))

	require.Equal(t, input, string(result.out))
	require.Empty(t, result.notes)
	require.Empty(t, result.clipboardText)
}

func TestOSCParserHandlesSplitSequence(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("a\x1b]777;notify;Ti"))
	require.Equal(t, "a", string(result.out))
	require.Empty(t, result.notes)

	result = p.Filter([]byte("tle;Body\x07b"))
	require.Equal(t, "b", string(result.out))
	require.Equal(t, []terminal.Notification{{Title: "Title", Message: "Body"}}, result.notes)
}

func TestOSCParserDecodesClipboardCopy(t *testing.T) {
	var p oscParser

	// OSC 52 with base64 "Hello" — should be stripped and decoded.
	result := p.Filter([]byte("before\x1b]52;c;SGVsbG8=\x07after"))

	require.Equal(t, "beforeafter", string(result.out))
	require.Equal(t, "Hello", result.clipboardText)
	require.Empty(t, result.notes)
}

func TestOSCParserDecodesClipboardCopyWithSTTerminator(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("\x1b]52;c;SGVsbG8=\x1b\\"))
	require.Equal(t, "Hello", result.clipboardText)
	require.Empty(t, result.out)
}

func TestOSCParserIgnoresClipboardQuery(t *testing.T) {
	var p oscParser

	// OSC 52 with "?" — a query, not a copy. Should pass through.
	result := p.Filter([]byte("\x1b]52;c?\x07"))

	require.Empty(t, result.clipboardText)
	require.Equal(t, "\x1b]52;c?\x07", string(result.out))
}
