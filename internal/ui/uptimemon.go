package ui

import (
	"fmt"
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// uptimeMonitor is a header-bar label that shows elapsed time since launch,
// refreshed at memRefreshInterval. Hours are prefixed only once elapsed
// reaches one hour; before that the format is mm:ss.
type uptimeMonitor struct {
	label *gtk.Label
	start time.Time
}

// newUptimeMonitor creates a dimmed label and starts a recurring refresh. The
// label is packed into the header bar by the caller.
func newUptimeMonitor() *uptimeMonitor {
	u := &uptimeMonitor{
		label: gtk.NewLabel(""),
		start: time.Now(),
	}

	u.label.AddCSSClass("bterm-memmon")
	u.label.SetTooltipText("Elapsed time since launch")
	u.label.SetVisible(true)

	u.update()

	glib.TimeoutAdd(memRefreshInterval, func() bool {
		u.update()

		return true
	})

	return u
}

// update sets the label text to the current elapsed duration.
func (u *uptimeMonitor) update() {
	u.label.SetText(formatUptime(time.Since(u.start)))
}

// formatUptime renders d as [hh]:mm:ss — hours prefixed only when >= 1h.
func formatUptime(d time.Duration) string {
	totalSec := int64(d.Seconds())

	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}

	return fmt.Sprintf("%02d:%02d", m, s)
}
