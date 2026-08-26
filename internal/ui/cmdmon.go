package ui

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/terminal"
)

// cmdPollInterval is how often the command monitor polls the PTY for command
// start/end. Kept short so completion appears near-instantly; the cost is a
// single TIOCGPGRP ioctl per open terminal per tick.
const cmdPollInterval = 200

type commandKey struct {
	tab    *tab
	paneID int
}

type commandState struct {
	running      bool
	start        time.Time
	lastDuration time.Duration
}

// cmdMonitor is a header-bar label that shows the duration of the latest
// command in the focused terminal. It tracks every open terminal so switching
// tabs or panes preserves running and completed command durations.
type cmdMonitor struct {
	label  *gtk.Label
	win    *window
	states map[commandKey]*commandState
	live   map[commandKey]struct{}
}

// newCmdMonitor creates a dimmed label and starts a recurring refresh. The
// label is packed into the header bar by the caller.
func newCmdMonitor(w *window) *cmdMonitor {
	m := &cmdMonitor{
		label:  gtk.NewLabel("—"),
		win:    w,
		states: make(map[commandKey]*commandState),
		live:   make(map[commandKey]struct{}),
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

// update polls every open terminal, removes state for closed terminals, and
// displays the focused terminal's current or latest command duration.
func (m *cmdMonitor) update() {
	now := time.Now()

	clear(m.live)

	for _, tab := range m.win.tabs {
		for paneID, term := range tab.area.terms {
			key := commandKey{tab: tab, paneID: paneID}

			m.live[key] = struct{}{}

			m.poll(key, term, now)
		}
	}

	for key := range m.states {
		if _, ok := m.live[key]; !ok {
			delete(m.states, key)
		}
	}

	m.renderFocused(now)
}

// poll updates one terminal's state from its foreground process group.
func (m *cmdMonitor) poll(key commandKey, term terminal.Terminal, now time.Time) {
	shellPID := term.ShellPID()

	if shellPID == 0 {
		return
	}

	foregroundPID, err := term.ForegroundPGID()
	if err != nil {
		return
	}

	state := m.states[key]

	if state == nil {
		state = &commandState{}
		m.states[key] = state
	}

	advanceCommandState(state, foregroundPID != shellPID, now)
}

// renderFocused displays the focused pane's state without changing it.
func (m *cmdMonitor) renderFocused(now time.Time) {
	key, ok := m.focusedKey()
	if !ok {
		m.label.SetText("—")

		return
	}

	state := m.states[key]

	switch {
	case state == nil:
		m.label.SetText("—")
	case state.running:
		m.label.SetText(formatCmdDuration(now.Sub(state.start)))
	case state.lastDuration > 0:
		m.label.SetText(formatCmdDuration(state.lastDuration))
	default:
		m.label.SetText("—")
	}
}

// focusedKey returns the active tab's focused pane key.
func (m *cmdMonitor) focusedKey() (commandKey, bool) {
	if len(m.win.tabs) == 0 {
		return commandKey{}, false
	}

	tab := m.win.tabs[m.win.active]
	paneID := tab.area.tree.Focused()

	if paneID == 0 {
		return commandKey{}, false
	}

	return commandKey{tab: tab, paneID: paneID}, true
}

// advanceCommandState applies a command start or completion transition.
func advanceCommandState(state *commandState, running bool, now time.Time) {
	switch {
	case running && !state.running:
		state.running = true
		state.start = now
	case !running && state.running:
		state.running = false
		state.lastDuration = now.Sub(state.start)
	}
}

// formatCmdDuration renders a duration like Go's Duration.String() but rounded
// to millisecond precision to avoid excessive decimal places.
func formatCmdDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}

	return d.Round(time.Millisecond).String()
}
