// Package logbuf provides a small process-wide ring buffer for diagnostic
// events, viewable in-app via the Diagnostics dialog.
package logbuf

import (
	"fmt"
	"sync"
	"time"
)

// capacity bounds the number of buffered lines; oldest lines are evicted.
const capacity = 500

//nolint:gochecknoglobals
var (
	mu    sync.Mutex
	lines []string
)

// Logf appends a timestamped, formatted line to the ring. Safe for
// concurrent use from any goroutine.
func Logf(format string, args ...any) {
	line := time.Now().Format("15:04:05.000") + " " + fmt.Sprintf(format, args...)

	mu.Lock()
	defer mu.Unlock()

	if len(lines) == capacity {
		copy(lines, lines[1:])

		lines = lines[:capacity-1]
	}

	lines = append(lines, line)
}

// Lines returns a snapshot copy of the buffered lines, oldest first.
func Lines() []string {
	mu.Lock()
	defer mu.Unlock()

	return append([]string(nil), lines...)
}
