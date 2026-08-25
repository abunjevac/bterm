package ui

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
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
	uiFontFamily    string
	uiFontSize      float64
	palette         *theme.Palette
	newTerm         terminal.Factory
}

func newWindow(app *gtk.Application, bundle *config.Bundle, workingDir string) *gtk.ApplicationWindow {
	w := newEmptyWindow(app, bundle, workingDir)

	w.newTabEnd()

	return w.win
}

func newEmptyWindow(app *gtk.Application, bundle *config.Bundle, workingDir string) *window {
	cfg := bundle.Config

	w := &window{
		app:             app,
		bundle:          bundle,
		keys:            bundle.LoadKeymap(),
		workingDir:      workingDir,
		fontFamily:      cfg.Font,
		fontSize:        cfg.FontSize,
		defaultFontSize: cfg.FontSize,
		uiFontFamily:    cfg.UIFont,
		uiFontSize:      cfg.UIFontSize,
		palette:         theme.Load(bundle.Dir, cfg.Theme),
		newTerm:         func() terminal.Terminal { return vtepkg.New() },
	}

	w.win = gtk.NewApplicationWindow(app)

	w.win.SetTitle(cfg.Title)
	w.win.SetIconName("io.github.abunjevac.bterm")

	applyStyle(w.palette, w.uiFontFamily, w.uiFontSize)

	w.buildTabBar()

	w.toast = newToaster(w.stack)

	w.win.SetChild(w.toast.overlay)

	w.installKeys()

	return w
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
	t.SetScrollbar(cfg.ShowScrollbar)

	w.installTerminalNotifications(t)
	w.installClipboardDetection(t)

	if len(w.tabs) == 0 {
		t.SetSize(cfg.WindowColumns, cfg.WindowRows)
	}

	t.Spawn(workingDir, shell, shellArgs(cfg), func(_ int, _ error) {})
}

// applyNewConfig applies live changes from a preferences save. Font and theme
// are updated immediately; other settings take effect on next launch.
func (w *window) applyNewConfig(old, next config.Config) { //nolint:cyclop // live config application needs per-field checks
	fontChanged := next.Font != old.Font || next.FontSize != old.FontSize

	if fontChanged {
		w.fontFamily = next.Font
		w.defaultFontSize = next.FontSize
		w.fontSize = next.FontSize
	}

	uiFontChanged := next.UIFont != old.UIFont || next.UIFontSize != old.UIFontSize

	if uiFontChanged {
		w.uiFontFamily = next.UIFont
		w.uiFontSize = next.UIFontSize
	}

	if next.Theme != old.Theme || uiFontChanged {
		if next.Theme != old.Theme {
			w.palette = theme.Load(w.bundle.Dir, next.Theme)
		}

		applyStyle(w.palette, w.uiFontFamily, w.uiFontSize)
	}

	for _, tab := range w.tabs {
		for _, t := range tab.area.terms {
			if fontChanged {
				t.SetFont(w.fontFamily, w.fontSize)
			}

			t.SetColors(w.palette)

			if next.ShowScrollbar != old.ShowScrollbar {
				t.SetScrollbar(next.ShowScrollbar)
			}

			if next.Scrollback != old.Scrollback {
				t.SetScrollback(next.Scrollback)
			}
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

// focusedTerminal returns the focused terminal of the active tab, or nil.
func (w *window) focusedTerminal() terminal.Terminal {
	if len(w.tabs) == 0 {
		return nil
	}

	return w.tabs[w.active].area.focusedTerminal()
}

func shellArgs(cfg *config.Config) []string {
	if len(cfg.ShellArgs) > 0 {
		return cfg.ShellArgs
	}

	return []string{"-l"}
}

// clipboard returns the GDK system clipboard for this window's display.
func (w *window) clipboard() *gdk.Clipboard {
	if w.win == nil {
		return nil
	}

	return gtk.BaseWidget(w.win).Display().Clipboard()
}

// reownClipboard reads the current clipboard text and writes it back with
// SetText, transferring ownership from the VTE widget to the application.
// Without this, closing the terminal that owned the clipboard clears it.
func (w *window) reownClipboard() {
	clip := w.clipboard()

	if clip == nil {
		return
	}

	// ReadTextAsync must be called from the GTK main thread.
	glib.IdleAdd(func() bool {
		clip.ReadTextAsync(context.Background(), func(res gio.AsyncResulter) {
			text, err := clip.ReadTextFinish(res)
			if err != nil || text == "" {
				return
			}

			// Re-set the text so the application owns the clipboard content,
			// independent of the VTE widget's lifetime.
			clip.SetText(text)
		})

		return false
	})
}
