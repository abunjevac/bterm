package ui

import (
	_ "embed"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed testdata/format_input.tsv
var formatInputCases string

func TestFormatInput(t *testing.T) {
	t.Parallel()

	for line := range strings.SplitSeq(strings.TrimSpace(formatInputCases), "\n") {
		hexInput, want, ok := strings.Cut(line, "\t")
		require.True(t, ok, line)

		data, err := hex.DecodeString(hexInput)
		require.NoError(t, err)

		require.Equal(t, want, formatInput(data), hexInput)
	}
}

func TestFormatByteCount(t *testing.T) {
	t.Parallel()

	require.Equal(t, "0B", formatByteCount(0))
	require.Equal(t, "1023B", formatByteCount(1023))
	require.Equal(t, "1.0K", formatByteCount(1024))
	require.Equal(t, "1.5M", formatByteCount(1536*1024))
	require.Equal(t, "2048.0G", formatByteCount(2<<40))
}
