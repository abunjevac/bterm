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

	tabs   []*tab
	active int
	stack  *gtk.Stack
	tabBox *gtk.Box
	toast  *toaster

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

	w.buildTabBar()

	w.toast = newToaster(w.stack)

	w.win.SetChild(w.toast.overlay)

	w.newTabEnd()
	w.installKeys()

	return w.win
}

// spawnTerm configures and spawns a shell in t. An empty workingDir defaults to $HOME.
func (w *window) spawnTerm(t terminal.Terminal, workingDir string) {
	cfg := w.bundle.Config
	shell := config.InferShell(cfg.Shell, os.Getenv("SHELL"))

	if workingDir == "" {
		workingDir, _ = os.UserHomeDir()
	}

	t.SetFont(w.fontFamily, w.fontSize)
	t.SetColors(w.palette)
	t.SetScrollback(cfg.Scrollback)
	t.Spawn(workingDir, shell, shellArgs(cfg), func(_ int, _ error) {})
}

func shellArgs(cfg *config.Config) []string {
	if len(cfg.ShellArgs) > 0 {
		return cfg.ShellArgs
	}

	return []string{"-l"}
}
