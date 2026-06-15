package ui

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/config"
	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/terminal"
	vtepkg "github.com/abunjevac/bterm/internal/terminal/vte"
	"github.com/abunjevac/bterm/internal/theme"
)

type window struct {
	app    *gtk.Application
	win    *gtk.ApplicationWindow
	bundle *config.Bundle
	keys   *keymap.Layout

	fontFamily      string
	fontSize        float64
	defaultFontSize float64
	palette         *theme.Palette
	newTerm         terminal.Factory
}

func newWindow(_ context.Context, app *gtk.Application, bundle *config.Bundle) *gtk.ApplicationWindow {
	cfg := bundle.Config

	w := &window{
		app:             app,
		bundle:          bundle,
		keys:            bundle.LoadKeymap(),
		fontFamily:      cfg.Font,
		fontSize:        cfg.FontSize,
		defaultFontSize: cfg.FontSize,
		palette:         theme.Load(bundle.Dir, cfg.Theme),
		newTerm:         func() terminal.Terminal { return vtepkg.New() },
	}

	w.win = gtk.NewApplicationWindow(app)

	w.win.SetTitle(cfg.Title)
	w.win.SetDefaultSize(cfg.WindowWidth, cfg.WindowHeight)
	w.win.SetIconName("io.github.abunjevac.bterm")

	applyStyle(w.palette)

	term := w.newTerm()

	term.SetFont(w.fontFamily, w.fontSize)
	term.SetColors(w.palette)
	term.SetScrollback(cfg.Scrollback)

	shell := config.InferShell(cfg.Shell, os.Getenv("SHELL"))

	home, _ := os.UserHomeDir()

	term.Spawn(home, shell, shellArgs(cfg), func(_ int, _ error) {})
	term.OnChildExited(func(_ int) { w.win.Close() })

	w.win.SetChild(term.Widget())

	return w.win
}

func shellArgs(cfg *config.Config) []string {
	if len(cfg.ShellArgs) > 0 {
		return cfg.ShellArgs
	}

	return []string{"-l"}
}
