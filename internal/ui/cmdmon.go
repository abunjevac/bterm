package ui

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/terminal"
)

// cmdPollInterval is how often the command monitor polls the PTY for command
// start/end. Kept short so completion appears near-instantly; the cost is a
// single TIOCGPGRP ioctl per tick.
const cmdPollInterval = 200

// cmdMonitor is a header-bar label that shows the duration of the latest
// command in the focused terminal. While a command is running it ticks at
// cmdPollInterval; when the command finishes it shows the final duration
// rounded to millisecond precision.
//
// Command detection uses TIOCGPGRP: when the foreground process group differs
// from the shell PID, a command is running. When focus moves to a different
// terminal the state resets — the timer only tracks the focused pane.
type cmdMonitor struct {
	label *gtk.Label
	win   *window

	running      bool
	start        time.Time
	lastDuration time.Duration
	lastTerm     terminal.Terminal
}

// newCmdMonitor creates a dimmed label and starts a recurring refresh. The
// label is packed into the header bar by the caller.
func newCmdMonitor(w *window) *cmdMonitor {
	m := &cmdMonitor{
		label: gtk.NewLabel("—"),
		win:   w,
	}

	m.label.AddCSSClass("bterm-memmon")
	m.label.SetTooltipText("Duration of latest command in focused pane")
	m.label.SetVisible(true)

	glib.TimeoutAdd(cmdPollInterval, func() bool {
		m.update()

		return true
	})

	return m
}

// update polls the focused terminal, detects command start/end transitions,
// and updates the label text.
func (m *cmdMonitor) update() {
	t := m.win.focusedTerminal()

	if t == nil {
		m.reset()

		return
	}

	if m.lastTerm != nil && t != m.lastTerm {
		m.reset()
	}

	m.lastTerm = t

	shellPID := t.ShellPID()

	if shellPID == 0 {
		m.label.SetText("—")

		return
	}

	fg, err := t.ForegroundPGID()
	if err != nil {
		m.label.SetText("—")

		return
	}

	m.advance(fg != shellPID)
}

// advance applies a state transition based on whether a command is running.
func (m *cmdMonitor) advance(cmdRunning bool) {
	switch {
	case cmdRunning && !m.running:
		m.running = true
		m.start = time.Now()
		m.label.SetText(formatCmdDuration(0))
	case !cmdRunning && m.running:
		m.running = false
		m.lastDuration = time.Since(m.start)
		m.label.SetText(formatCmdDuration(m.lastDuration))
	case cmdRunning:
		m.label.SetText(formatCmdDuration(time.Since(m.start)))
	case m.lastDuration > 0:
		m.label.SetText(formatCmdDuration(m.lastDuration))
	default:
		m.label.SetText("—")
	}
}

// reset clears the monitor state, used when focus changes or no terminal exists.
func (m *cmdMonitor) reset() {
	m.running = false
	m.start = time.Time{}
	m.lastDuration = 0
	m.lastTerm = nil

	m.label.SetText("—")
}

// formatCmdDuration renders a duration like Go's Duration.String() but rounded
// to millisecond precision to avoid excessive decimal places.
func formatCmdDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	return d.Round(time.Millisecond).String()
}
