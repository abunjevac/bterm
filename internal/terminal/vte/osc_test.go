package vte

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOSCParserStrips777Notification(t *testing.T) {
	var p oscParser

	out, notes := p.Filter([]byte("before\x1b]777;notify;Title;Message\x07after"))

	require.Equal(t, "beforeafter", string(out))
	require.Equal(t, []terminalNotification{{Title: "Title", Message: "Message"}}, notes)
}

func TestOSCParserStripsOSC9Notification(t *testing.T) {
	var p oscParser

	out, notes := p.Filter([]byte("\x1b]9;Build complete\x1b\\"))

	require.Empty(t, out)
	require.Equal(t, []terminalNotification{{Message: "Build complete"}}, notes)
}

func TestOSCParserLeavesProgressAndUnknownOSC(t *testing.T) {
	var p oscParser

	input := "\x1b]9;4;1;50\x07\x1b]0;title\x07"
	out, notes := p.Filter([]byte(input))

	require.Equal(t, input, string(out))
	require.Empty(t, notes)
}

func TestOSCParserHandlesSplitSequence(t *testing.T) {
	var p oscParser

	out, notes := p.Filter([]byte("a\x1b]777;notify;Ti"))
	require.Equal(t, "a", string(out))
	require.Empty(t, notes)

	out, notes = p.Filter([]byte("tle;Body\x07b"))
	require.Equal(t, "b", string(out))
	require.Equal(t, []terminalNotification{{Title: "Title", Message: "Body"}}, notes)
}
