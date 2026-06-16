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

	workingDir string

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

func newWindow(_ context.Context, app *gtk.Application, bundle *config.Bundle, workingDir string) *gtk.ApplicationWindow {
	cfg := bundle.Config

	w := &window{
		app:             app,
		bundle:          bundle,
		keys:            bundle.LoadKeymap(),
		workingDir:      workingDir,
		fontFamily:      cfg.Font,
		fontSize:        cfg.FontSize,
		defaultFontSize: cfg.FontSize,
		palette:         theme.Load(bundle.Dir, cfg.Theme),
		newTerm:         func() terminal.Terminal { return vtepkg.New() },
	}

	w.win = gtk.NewApplicationWindow(app)

	w.win.SetTitle(cfg.Title)
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

	if len(w.tabs) == 0 {
		t.SetSize(cfg.WindowColumns, cfg.WindowRows)
	}

	t.Spawn(workingDir, shell, shellArgs(cfg), func(_ int, _ error) {})
}

// applyNewConfig applies live changes from a preferences save. Font and theme
// are updated immediately; other settings take effect on next launch.
func (w *window) applyNewConfig(old, next config.Config) {
	fontChanged := next.Font != old.Font || next.FontSize != old.FontSize

	if fontChanged {
		w.fontFamily = next.Font
		w.defaultFontSize = next.FontSize
		w.fontSize = next.FontSize
	}

	if next.Theme != old.Theme {
		w.palette = theme.Load(w.bundle.Dir, next.Theme)

		applyStyle(w.palette)
	}

	for _, tab := range w.tabs {
		for _, t := range tab.area.terms {
			if fontChanged {
				t.SetFont(w.fontFamily, w.fontSize)
			}

			t.SetColors(w.palette)
		}
	}
}

// activeCWD returns the working directory of the active tab's focused terminal,
// or an empty string when unavailable (callers fall back to $HOME via spawnTerm).
func (w *window) activeCWD() string {
	if len(w.tabs) == 0 {
		return w.workingDir
	}

	if ft := w.tabs[w.active].area.focusedTerminal(); ft != nil {
		return ft.CurrentDir()
	}

	return w.workingDir
}

func shellArgs(cfg *config.Config) []string {
	if len(cfg.ShellArgs) > 0 {
		return cfg.ShellArgs
	}

	return []string{"-l"}
}
