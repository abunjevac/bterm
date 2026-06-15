// Package terminal abstracts a terminal widget so the backend (VTE today) is swappable.
package terminal

import (
	"github.com/abunjevac/bterm/internal/theme"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// SpawnCallback reports the result of starting a shell.
type SpawnCallback func(pid int, err error)

// Terminal is one shell-backed terminal widget.
type Terminal interface {
	Widget() gtk.Widgetter
	Spawn(workingDir, shell string, args []string, cb SpawnCallback)
	SetFont(family string, size float64)
	SetColors(p *theme.Palette)
	SetScrollback(lines int)
	SetSize(columns, rows int)
	CurrentDir() string
	Copy()
	Paste()
	OnTitleChanged(fn func(title string))
	OnChildExited(fn func(status int))
}

// Factory creates new terminals. The VTE implementation provides one.
type Factory func() Terminal
