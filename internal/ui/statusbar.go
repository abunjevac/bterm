package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"github.com/abunjevac/bterm/internal/legacy"
	"github.com/abunjevac/bterm/internal/terminal"
)

// statusRefreshInterval is the delay between statusbar updates, in milliseconds.
const statusRefreshInterval = 1000

// statFieldTPGID is the index of tpgid (the tty's foreground process
// group) in /proc/<pid>/stat once the pid and comm fields are stripped off:
// state ppid pgrp session tty_nr tpgid.
const statFieldTPGID = 5

// paneStatus is the thin strip under one pane showing live terminal state.
type paneStatus struct {
	label *gtk.Label
	term  terminal.Terminal
	view  *gtk.Widget
}

// newPaneStatus creates the strip for term and starts its refresh timer. The
// timer ends once alive reports false, i.e. after the pane is closed.
func newPaneStatus(term terminal.Terminal, alive func() bool, visible bool) *paneStatus {
	s := &paneStatus{
		label: gtk.NewLabel(""),
		term:  term,
		view:  gtk.BaseWidget(term.Widget()),
	}

	s.label.AddCSSClass("bterm-statusbar")
	s.label.SetXAlign(0)
	s.label.SetHExpand(true)
	s.label.SetEllipsize(pango.EllipsizeEnd)
	s.label.SetVisible(visible)

	s.view.ConnectMap(s.update)

	glib.TimeoutAdd(statusRefreshInterval, func() bool {
		if !alive() {
			return false
		}

		s.update()

		return true
	})

	return s
}

// setVisible shows or hides the strip, refreshing it when shown.
func (s *paneStatus) setVisible(visible bool) {
	s.label.SetVisible(visible)

	s.update()
}

// update rewrites the strip text. Hidden or unmapped strips are skipped so
// background tabs cost nothing.
func (s *paneStatus) update() {
	if !s.label.Visible() || !s.view.Mapped() {
		return
	}

	st := s.term.Stats()

	parts := []string{
		"pid " + strconv.Itoa(s.term.ShellPID()),
		"fg " + foregroundProgram(s.term.ShellPID()),
		fmt.Sprintf("%dx%d", st.Columns, st.Rows),
		fmt.Sprintf("view %dx%d mapped=%v", s.view.Width(), s.view.Height(), s.view.Mapped()),
		fmt.Sprintf("in %s out %s", formatByteCount(st.BytesIn), formatByteCount(st.BytesOut)),
		"key " + formatLastInput(st),
		shortenHome(s.term.CurrentDir()),
	}

	s.label.SetText(strings.Join(parts, " · "))
}

// foregroundProgram names the process in the foreground of the shell's tty,
// or "?" when it cannot be determined.
func foregroundProgram(shellPID int) string {
	if shellPID <= 0 {
		return "?"
	}

	data, err := os.ReadFile("/proc/" + strconv.Itoa(shellPID) + "/stat")
	if err != nil {
		return "?"
	}

	// comm may contain spaces and parentheses; the remaining fields follow the last ')'
	text := string(data)

	_, tail, ok := strings.CutLast(text, ")")
	if !ok {
		return "?"
	}

	fields := strings.Fields(tail)

	if len(fields) <= statFieldTPGID {
		return "?"
	}

	pgid, err := strconv.Atoi(fields[statFieldTPGID])
	if err != nil || pgid <= 0 {
		return "?"
	}

	name, err := legacy.ProgramName(pgid)
	if err != nil {
		return "?"
	}

	return name
}

// formatLastInput renders the latest chunk sent to the shell as escaped text,
// or "-" when nothing was sent yet.
func formatLastInput(st terminal.Stats) string {
	if len(st.LastInput) == 0 {
		return "-"
	}

	text := formatInput(st.LastInput)

	if st.LastInputLen > len(st.LastInput) {
		text += fmt.Sprintf("…(%d bytes)", st.LastInputLen)
	}

	return text
}

// formatInput writes data with control bytes as escape sequences: \e for ESC,
// \r, \n, \t, and \xNN for other non-printable bytes.
func formatInput(data []byte) string {
	var b strings.Builder

	for _, c := range data {
		switch {
		case c == 0x1b:
			b.WriteString(`\e`)
		case c == '\r':
			b.WriteString(`\r`)
		case c == '\n':
			b.WriteString(`\n`)
		case c == '\t':
			b.WriteString(`\t`)
		case c == '\\':
			b.WriteString(`\\`)
		case c < 0x20 || c >= 0x7f:
			fmt.Fprintf(&b, `\x%02x`, c)
		default:
			b.WriteByte(c)
		}
	}

	return b.String()
}

// formatByteCount renders n as a short human-readable size.
func formatByteCount(n uint64) string {
	const unit = 1024

	if n < unit {
		return strconv.FormatUint(n, 10) + "B"
	}

	value := float64(n)

	for _, suffix := range []string{"K", "M", "G"} {
		value /= unit

		if value < unit || suffix == "G" {
			return fmt.Sprintf("%.1f%s", value, suffix)
		}
	}

	return strconv.FormatUint(n, 10) + "B"
}

// shortenHome replaces a leading home directory in path with "~".
func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	if path == home {
		return "~"
	}

	if rest, ok := strings.CutPrefix(path, home+"/"); ok {
		return "~/" + rest
	}

	return path
}
