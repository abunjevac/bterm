package vte

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOSCParserStrips777Notification(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("before\x1b]777;notify;Title;Message\x07after"))

	require.Equal(t, "beforeafter", string(result.out))
	require.Equal(t, []terminalNotification{{Title: "Title", Message: "Message"}}, result.notes)
	require.False(t, result.clipboardCopied)
}

func TestOSCParserStripsOSC9Notification(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("\x1b]9;Build complete\x1b\\"))

	require.Empty(t, result.out)
	require.Equal(t, []terminalNotification{{Message: "Build complete"}}, result.notes)
	require.False(t, result.clipboardCopied)
}

func TestOSCParserLeavesProgressAndUnknownOSC(t *testing.T) {
	var p oscParser

	input := "\x1b]9;4;1;50\x07\x1b]0;title\x07"
	result := p.Filter([]byte(input))

	require.Equal(t, input, string(result.out))
	require.Empty(t, result.notes)
	require.False(t, result.clipboardCopied)
}

func TestOSCParserHandlesSplitSequence(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("a\x1b]777;notify;Ti"))
	require.Equal(t, "a", string(result.out))
	require.Empty(t, result.notes)

	result = p.Filter([]byte("tle;Body\x07b"))
	require.Equal(t, "b", string(result.out))
	require.Equal(t, []terminalNotification{{Title: "Title", Message: "Body"}}, result.notes)
}

func TestOSCParserDetectsClipboardCopy(t *testing.T) {
	var p oscParser

	// OSC 52 with base64 data — should pass through and set clipboardCopied.
	result := p.Filter([]byte("before\x1b]52;c;SGVsbG8=\x07after"))

	require.Equal(t, "before\x1b]52;c;SGVsbG8=\x07after", string(result.out))
	require.True(t, result.clipboardCopied)
	require.Empty(t, result.notes)
}

func TestOSCParserDetectsClipboardCopyWithSTTerminator(t *testing.T) {
	var p oscParser

	result := p.Filter([]byte("\x1b]52;c;SGVsbG8=\x1b\\"))
	require.True(t, result.clipboardCopied)
	require.Equal(t, "\x1b]52;c;SGVsbG8=\x1b\\", string(result.out))
}

func TestOSCParserIgnoresClipboardQuery(t *testing.T) {
	var p oscParser

	// OSC 52 with no data field — a query or clear, not a copy.
	result := p.Filter([]byte("\x1b]52;c?\x07"))

	require.False(t, result.clipboardCopied)
	require.Equal(t, "\x1b]52;c?\x07", string(result.out))
}
