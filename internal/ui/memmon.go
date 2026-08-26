package ui

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// memRefreshInterval is the delay between memory display updates.
const memRefreshInterval = 5000

// memMonitor is a header-bar label that periodically shows the process RSS
// and the Go runtime's live heap allocation.
type memMonitor struct {
	label *gtk.Label
}

// newMemMonitor creates a dimmed label and starts a recurring refresh. The
// label is packed into the header bar by the caller; the refresh timer keeps
// the monitor alive for as long as it runs.
func newMemMonitor() *memMonitor {
	m := &memMonitor{
		label: gtk.NewLabel(""),
	}

	m.label.AddCSSClass("bterm-memmon")
	m.label.SetTooltipText("Process RSS · Go heap allocation")
	m.label.SetVisible(true)

	m.update()

	glib.TimeoutAdd(memRefreshInterval, func() bool {
		m.update()

		return true
	})

	return m
}

// update reads the current RSS and heap values and sets the label text.
func (m *memMonitor) update() {
	rss := readRSS()
	heap := readHeap()

	m.label.SetText(fmt.Sprintf("RSS %s · Heap %s", formatMemMB(rss), formatMemMB(heap)))
}

// readRSS returns the process resident set size in bytes, or 0 when the
// /proc/self/status VmRSS line is unavailable (non-Linux or parse failure).
func readRSS() uint64 {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 2 {
			return 0
		}

		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0
		}

		return kb * 1024
	}

	return 0
}

// readHeap returns the Go runtime's current HeapAlloc in bytes.
func readHeap() uint64 {
	var ms runtime.MemStats

	runtime.ReadMemStats(&ms)

	return ms.HeapAlloc
}

// formatMemMB renders bytes as a compact megabyte value, keeping one decimal
// below 10 MB to avoid jitter on small allocations.
func formatMemMB(bytes uint64) string {
	mb := float64(bytes) / 1024 / 1024

	if mb < 10 {
		return fmt.Sprintf("%.1fM", mb)
	}

	return fmt.Sprintf("%dM", int64(mb))
}
