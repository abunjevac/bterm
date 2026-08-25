package terminal

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/theme"
)

// SpawnCallback reports the result of starting a shell.
type SpawnCallback func(pid int, err error)

// Terminal is one shell-backed terminal widget.
//
//nolint:interfacebloat // terminal abstraction requires many methods
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
	// KittyDisambiguate reports whether the kitty keyboard protocol
	// disambiguate mode is active for this terminal.
	KittyDisambiguate() bool
	OnTitleChanged(fn func(title string))
	OnNotification(fn func(title, message string))
	OnClipboardCopy(fn func(text string))
	OnChildExited(fn func(status int))
}

// Factory creates new terminals. The VTE implementation provides one.
type Factory func() Terminal
