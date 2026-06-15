package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/assets"
	"github.com/abunjevac/bterm/internal/config"
	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/theme"
	"github.com/abunjevac/bterm/internal/version"
)

func showShortcutsDialog(parent *gtk.ApplicationWindow, layout *keymap.Layout) {
	grid := gtk.NewGrid()

	grid.SetRowSpacing(4)
	grid.SetColumnSpacing(16)
	grid.SetMarginTop(8)
	grid.SetMarginBottom(8)
	grid.SetMarginStart(8)
	grid.SetMarginEnd(8)

	for i, entry := range layout.Entries() {
		nameLabel := gtk.NewLabel(humanizeAction(entry.Action))

		nameLabel.SetHAlign(gtk.AlignEnd)

		keysLabel := gtk.NewLabel(strings.Join(entry.Keys, ", "))

		keysLabel.SetHAlign(gtk.AlignStart)
		keysLabel.AddCSSClass("bterm-shortcut-key")

		grid.Attach(nameLabel, 0, i, 1, 1)
		grid.Attach(keysLabel, 1, i, 1, 1)
	}

	scroll := gtk.NewScrolledWindow()

	scroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scroll.SetChild(grid)

	win := gtk.NewWindow()

	win.SetTitle("Keyboard Shortcuts")
	win.SetTransientFor(&parent.Window)
	win.SetModal(true)
	win.SetDefaultSize(420, 520)
	win.SetChild(scroll)
	win.Present()
}

func humanizeAction(a keymap.Action) string {
	words := strings.Split(a.String(), "_")

	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}

	return strings.Join(words, " ")
}

// configForm holds all editable widgets in the Preferences dialog.
type configForm struct {
	fontEntry      *gtk.Entry
	fontSizeSpin   *gtk.SpinButton
	themeDD        *gtk.DropDown
	themes         []string
	shellEntry     *gtk.Entry
	shellArgsEntry *gtk.Entry
	scrollbackSpin *gtk.SpinButton
	widthSpin      *gtk.SpinButton
	heightSpin     *gtk.SpinButton
	titleEntry     *gtk.Entry
}

// collect reads the current widget values into a new Config, using base as the
// starting point so fields not represented in the form keep their zero values.
func (f *configForm) collect(base config.Config) config.Config {
	next := base

	next.Font = f.fontEntry.Text()
	next.FontSize = f.fontSizeSpin.Value()

	if sel := f.themeDD.Selected(); int(sel) < len(f.themes) {
		next.Theme = f.themes[sel]
	}

	next.Shell = f.shellEntry.Text()

	if text := strings.TrimSpace(f.shellArgsEntry.Text()); text != "" {
		next.ShellArgs = strings.Fields(text)
	} else {
		next.ShellArgs = nil
	}

	next.Scrollback = int(f.scrollbackSpin.Value())
	next.WindowColumns = int(f.widthSpin.Value())
	next.WindowRows = int(f.heightSpin.Value())
	next.Title = f.titleEntry.Text()

	return next
}

// buildConfigForm constructs the Preferences grid and returns the scroll
// container along with a form that holds references to each widget.
func buildConfigForm(cfg config.Config) (*gtk.ScrolledWindow, configForm) {
	grid := gtk.NewGrid()

	grid.SetRowSpacing(8)
	grid.SetColumnSpacing(12)
	grid.SetMarginTop(12)
	grid.SetMarginBottom(12)
	grid.SetMarginStart(12)
	grid.SetMarginEnd(12)

	row := 0

	attach := func(label string, w gtk.Widgetter) {
		lbl := gtk.NewLabel(label)

		lbl.SetHAlign(gtk.AlignEnd)
		grid.Attach(lbl, 0, row, 1, 1)
		grid.Attach(w, 1, row, 1, 1)

		row++
	}

	var f configForm

	f.fontEntry = cfgEntry(cfg.Font, "")

	attach("Font", f.fontEntry)

	f.fontSizeSpin = gtk.NewSpinButtonWithRange(4, 72, 0.5)
	f.fontSizeSpin.SetValue(cfg.FontSize)
	f.fontSizeSpin.SetDigits(1)

	attach("Font size", f.fontSizeSpin)

	f.themes = theme.BuiltinNames()
	f.themeDD = gtk.NewDropDownFromStrings(f.themes)
	f.themeDD.SetSelected(cfgThemeIndex(f.themes, cfg.Theme))

	attach("Theme", f.themeDD)

	f.shellEntry = cfgEntry(cfg.Shell, "auto-detect from $SHELL")

	attach("Shell", f.shellEntry)

	f.shellArgsEntry = cfgEntry(strings.Join(cfg.ShellArgs, " "), "-l")

	attach("Shell args", f.shellArgsEntry)

	f.scrollbackSpin = cfgSpin(100, 200000, 100, float64(cfg.Scrollback))

	attach("Scrollback lines", f.scrollbackSpin)

	f.widthSpin = cfgSpin(40, 500, 1, float64(cfg.WindowColumns))

	attach("Window columns", f.widthSpin)

	f.heightSpin = cfgSpin(10, 200, 1, float64(cfg.WindowRows))

	attach("Window rows", f.heightSpin)

	f.titleEntry = cfgEntry(cfg.Title, "bterm")

	attach("Default title", f.titleEntry)

	return cfgScroll(grid), f
}

func cfgScroll(child gtk.Widgetter) *gtk.ScrolledWindow {
	s := gtk.NewScrolledWindow()

	s.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	s.SetVExpand(true)
	s.SetChild(child)

	return s
}

func cfgEntry(text, placeholder string) *gtk.Entry {
	e := gtk.NewEntry()

	e.SetText(text)
	e.SetHExpand(true)

	if placeholder != "" {
		e.SetPlaceholderText(placeholder)
	}

	return e
}

func cfgSpin(lo, hi, step, value float64) *gtk.SpinButton {
	s := gtk.NewSpinButtonWithRange(lo, hi, step)

	s.SetValue(value)

	return s
}

func cfgThemeIndex(themes []string, name string) uint {
	for i, n := range themes {
		if n == name {
			return uint(i)
		}
	}

	return 0
}

func showConfigDialog(parent *gtk.ApplicationWindow, w *window) {
	cfg := *w.bundle.Config
	scroll, form := buildConfigForm(cfg)

	var win *gtk.Window

	cancelBtn := gtk.NewButtonWithLabel("Cancel")

	cancelBtn.ConnectClicked(func() { win.Close() })

	saveBtn := gtk.NewButtonWithLabel("Save")

	saveBtn.AddCSSClass("suggested-action")
	saveBtn.ConnectClicked(func() {
		newCfg := form.collect(cfg)

		if err := config.Save(filepath.Join(w.bundle.Dir, "config.toml"), &newCfg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "bterm: save config: %v\n", err)

			return
		}

		w.applyNewConfig(cfg, newCfg)

		*w.bundle.Config = newCfg

		win.Close()
	})

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 8)

	btnBox.SetHAlign(gtk.AlignEnd)
	btnBox.SetMarginTop(8)
	btnBox.SetMarginBottom(12)
	btnBox.SetMarginEnd(12)
	btnBox.Append(cancelBtn)
	btnBox.Append(saveBtn)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 0)

	mainBox.Append(scroll)
	mainBox.Append(btnBox)

	win = gtk.NewWindow()

	win.SetTitle("Preferences")
	win.SetTransientFor(&parent.Window)
	win.SetModal(true)
	win.SetDefaultSize(400, 500)
	win.SetChild(mainBox)
	win.Present()
}

func showAboutDialog(parent *gtk.ApplicationWindow) {
	d := gtk.NewAboutDialog()

	d.SetProgramName("bterm")
	d.SetVersion(version.Version)
	d.SetComments("An opinionated GTK4 terminal emulator")
	d.SetLicenseType(gtk.LicenseMITX11)
	d.SetWebsite("https://github.com/abunjevac/bterm")
	d.SetTransientFor(&parent.Window)
	d.SetModal(true)

	logo, err := gdk.NewTextureFromBytes(glib.NewBytes(assets.IconPNG))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "bterm: load about logo: %v\n", err)
	} else {
		d.SetLogo(logo)
	}

	d.Present()
}
