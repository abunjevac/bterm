package logbuf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // tests share the process-global ring buffer
func TestLogfAppendsTimestampedLines(t *testing.T) {
	clearForTest(t)

	Logf("term %d: %s", 7, "spawned")

	got := Lines()

	require.Len(t, got, 1)
	require.Contains(t, got[0], "term 7: spawned")
	require.Regexp(t, `^\d{2}:\d{2}:\d{2}\.\d{3} `, got[0])
}

//nolint:paralleltest // tests share the process-global ring buffer
func TestEvictsOldestAtCapacity(t *testing.T) {
	clearForTest(t)

	for i := range capacity + 10 {
		Logf("line %d", i)
	}

	got := Lines()

	require.Len(t, got, capacity)
	require.Contains(t, got[0], "line 10")
	require.Contains(t, got[capacity-1], "line 509")
}

//nolint:paralleltest // tests share the process-global ring buffer
func TestLinesReturnsSnapshotCopy(t *testing.T) {
	clearForTest(t)

	Logf("alpha")

	got := Lines()
	got[0] = "mutated"

	require.Contains(t, Lines()[0], "alpha")
}

// clearForTest empties the ring and registers cleanup to leave it empty.
func clearForTest(t *testing.T) {
	t.Helper()

	mu.Lock()
	lines = nil
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()
		lines = nil
		mu.Unlock()
	})
}
