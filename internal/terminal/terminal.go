package terminal

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/theme"
)

// SpawnCallback reports the result of starting a shell.
type SpawnCallback func(pid int, err error)

// Notification is a terminal notification extracted from an OSC sequence.
type Notification struct {
	Title   string
	Message string
}

// Terminal is one shell-backed terminal widget.
//
//nolint:interfacebloat // intentionally comprehensive; splitting would fragment a cohesive abstraction
type Terminal interface {
	Widget() gtk.Widgetter
	Spawn(workingDir, shell string, args []string, cb SpawnCallback)
	SetFont(family string, size float64)
	SetColors(p *theme.Palette)
	SetScrollback(lines int)
	SetScrollbar(visible bool)
	SetSize(columns, rows int)
	CurrentDir() string
	Copy()
	Paste()
	FeedChild(data []byte)
	OnTitleChanged(fn func(title string))
	OnNotification(fn func(title, message string))
	OnClipboardCopy(fn func(text string))
	OnChildExited(fn func(status int))
	ShellPID() int
	ForegroundPGID() (int, error)
}

// Factory creates new terminals. The VTE implementation provides one.
type Factory func() Terminal
