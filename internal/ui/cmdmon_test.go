package ui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdvanceCommandStatePreservesDurationAcrossPolling(t *testing.T) {
	t.Parallel()

	start := time.Unix(100, 0)
	state := &commandState{}

	advanceCommandState(state, true, start)

	require.True(t, state.running)
	require.Equal(t, start, state.start)

	advanceCommandState(state, true, start.Add(time.Second))

	require.True(t, state.running)
	require.Equal(t, start, state.start)

	advanceCommandState(state, false, start.Add(1234*time.Millisecond))

	require.False(t, state.running)
	require.Equal(t, 1234*time.Millisecond, state.lastDuration)

	advanceCommandState(state, false, start.Add(2*time.Second))

	require.Equal(t, 1234*time.Millisecond, state.lastDuration)
}

func TestFormatCmdDuration(t *testing.T) {
	t.Parallel()

	require.Equal(t, "1.235s", formatCmdDuration(1234567890*time.Nanosecond))
	require.Equal(t, "01:02", formatCmdDuration(time.Minute+2345678900*time.Nanosecond))
	require.Equal(t, "01:00", formatCmdDuration(time.Minute))
	require.Equal(t, "59.999s", formatCmdDuration(time.Minute-time.Millisecond))
}
