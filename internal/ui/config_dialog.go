package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"

	"github.com/abunjevac/bterm/internal/config"
	"github.com/abunjevac/bterm/internal/theme"
)

// configForm holds all editable widgets in the Preferences dialog.
type configForm struct {
	fontBtn                 *gtk.FontDialogButton
	uiFontBtn               *gtk.FontDialogButton
	themeDD                 *gtk.DropDown
	themes                  []string
	shellEntry              *gtk.Entry
	shellArgsEntry          *gtk.Entry
	scrollbackSpin          *gtk.SpinButton
	scrollbarSwitch         *gtk.Switch
	widthSpin               *gtk.SpinButton
	heightSpin              *gtk.SpinButton
	titleEntry              *gtk.Entry
	terminalNotificationsDD *gtk.DropDown
	editorEntry             *gtk.Entry
	editorArgsEntry         *gtk.Entry
	fileBrowserEntry        *gtk.Entry
	fileBrowserArgsEntry    *gtk.Entry
}

// collect reads the current widget values into a new Config, using base as the
// starting point so fields not represented in the form keep their current values.
func (f *configForm) collect(base config.Config) config.Config {
	next := base

	if desc := f.fontBtn.FontDesc(); desc != nil {
		if fam := desc.Family(); fam != "" {
			next.Font = fam
		}

		if sz := desc.Size(); sz > 0 {
			next.FontSize = float64(sz) / pango.SCALE
		}
	}

	if desc := f.uiFontBtn.FontDesc(); desc != nil {
		if fam := desc.Family(); fam != "" {
			next.UIFont = fam
		}

		if sz := desc.Size(); sz > 0 {
			next.UIFontSize = float64(sz) / pango.SCALE
		}
	}

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
	next.ShowScrollbar = f.scrollbarSwitch.Active()
	next.WindowColumns = int(f.widthSpin.Value())
	next.WindowRows = int(f.heightSpin.Value())
	next.Title = f.titleEntry.Text()
	next.Editor = f.editorEntry.Text()
	next.EditorArgs = splitCommandArgs(f.editorArgsEntry.Text())
	next.FileBrowser = f.fileBrowserEntry.Text()
	next.FileBrowserArgs = splitCommandArgs(f.fileBrowserArgsEntry.Text())

	if f.terminalNotificationsDD.Selected() == 0 {
		next.TerminalNotificationMethod = config.TerminalNotificationDBus
	} else {
		next.TerminalNotificationMethod = config.TerminalNotificationOff
	}

	return next
}

func splitCommandArgs(text string) []string {
	if text = strings.TrimSpace(text); text != "" {
		return strings.Fields(text)
	}

	return nil
}

// updateSpinners commits any pending text input in spin buttons so that
// collect() reads the latest typed values.
func (f *configForm) updateSpinners() {
	f.scrollbackSpin.Update()
	f.widthSpin.Update()
	f.heightSpin.Update()
}

// buildConfigForm constructs the Preferences content along with a form that
// holds references to each widget.
func buildConfigForm(cfg config.Config) (*gtk.Box, configForm) { //nolint:funlen
	var f configForm

	// Typography
	fontDlg := gtk.NewFontDialog()

	f.fontBtn = gtk.NewFontDialogButton(fontDlg)

	f.fontBtn.SetLevel(gtk.FontLevelFont)
	f.fontBtn.SetHExpand(true)

	initFontDesc := pango.NewFontDescription()

	initFontDesc.SetFamily(cfg.Font)
	initFontDesc.SetSize(int(cfg.FontSize * pango.SCALE))

	f.fontBtn.SetFontDesc(initFontDesc)

	typographyGrid := cfgGrid()

	cfgAttach(typographyGrid, 0, "Font", f.fontBtn)

	// UI font (optional — leave unset to use system default)
	uiFontDlg := gtk.NewFontDialog()

	f.uiFontBtn = gtk.NewFontDialogButton(uiFontDlg)

	f.uiFontBtn.SetLevel(gtk.FontLevelFont)
	f.uiFontBtn.SetHExpand(true)

	if cfg.UIFont != "" || cfg.UIFontSize > 0 {
		uiFontDesc := pango.NewFontDescription()

		if cfg.UIFont != "" {
			uiFontDesc.SetFamily(cfg.UIFont)
		}

		if cfg.UIFontSize > 0 {
			uiFontDesc.SetSize(int(cfg.UIFontSize * pango.SCALE))
		}

		f.uiFontBtn.SetFontDesc(uiFontDesc)
	}

	cfgAttach(typographyGrid, 1, "UI font", f.uiFontBtn)

	// Appearance
	f.themes = theme.BuiltinNames()
	f.themeDD = gtk.NewDropDownFromStrings(f.themes)

	f.themeDD.SetSelected(cfgThemeIndex(f.themes, cfg.Theme))
	f.themeDD.SetHExpand(true)

	appearanceGrid := cfgGrid()

	cfgAttach(appearanceGrid, 0, "Theme", f.themeDD)

	// Shell
	f.shellEntry = cfgEntry(cfg.Shell, "auto-detect from $SHELL")
	f.shellArgsEntry = cfgEntry(strings.Join(cfg.ShellArgs, " "), "-l")

	shellGrid := cfgGrid()

	cfgAttach(shellGrid, 0, "Shell", f.shellEntry)
	cfgAttach(shellGrid, 1, "Args", f.shellArgsEntry)

	// Terminal
	f.scrollbackSpin = cfgSpin(100, 200000, 100, float64(cfg.Scrollback))
	f.scrollbarSwitch = gtk.NewSwitch()

	f.scrollbarSwitch.SetActive(cfg.ShowScrollbar)
	f.scrollbarSwitch.SetHAlign(gtk.AlignStart)

	f.terminalNotificationsDD = gtk.NewDropDownFromStrings([]string{"D-Bus", "Off"})

	if cfg.TerminalNotificationMethod == config.TerminalNotificationDBus {
		f.terminalNotificationsDD.SetSelected(0)
	} else {
		f.terminalNotificationsDD.SetSelected(1)
	}

	f.terminalNotificationsDD.SetHExpand(true)

	terminalGrid := cfgGrid()

	cfgAttach(terminalGrid, 0, "Scrollback lines", f.scrollbackSpin)
	cfgAttach(terminalGrid, 1, "Show scrollbar", f.scrollbarSwitch)
	cfgAttach(terminalGrid, 2, "Terminal notifications", f.terminalNotificationsDD)

	// Window
	f.widthSpin = cfgSpin(40, 500, 1, float64(cfg.WindowColumns))
	f.heightSpin = cfgSpin(10, 200, 1, float64(cfg.WindowRows))
	f.titleEntry = cfgEntry(cfg.Title, "bterm")

	windowGrid := cfgGrid()

	cfgAttach(windowGrid, 0, "Columns", f.widthSpin)
	cfgAttach(windowGrid, 1, "Rows", f.heightSpin)
	cfgAttach(windowGrid, 2, "Title", f.titleEntry)

	// Applications
	f.editorEntry = cfgEntry(cfg.Editor, "zed")
	f.editorArgsEntry = cfgEntry(strings.Join(cfg.EditorArgs, " "), "{cwd}")
	f.fileBrowserEntry = cfgEntry(cfg.FileBrowser, "dolphin")
	f.fileBrowserArgsEntry = cfgEntry(strings.Join(cfg.FileBrowserArgs, " "), "{cwd}")

	applicationsGrid := cfgGrid()

	cfgAttach(applicationsGrid, 0, "Editor", f.editorEntry)
	cfgAttach(applicationsGrid, 1, "Editor args", f.editorArgsEntry)
	cfgAttach(applicationsGrid, 2, "File browser", f.fileBrowserEntry)
	cfgAttach(applicationsGrid, 3, "File browser args", f.fileBrowserArgsEntry)

	// Assemble
	box := gtk.NewBox(gtk.OrientationVertical, 0)

	sections := []struct {
		title string
		grid  *gtk.Grid
	}{
		{"Typography", typographyGrid},
		{"Appearance", appearanceGrid}, //nolint:goconst // dialog section labels intentionally remain local.
		{"Shell", shellGrid},
		{"Terminal", terminalGrid},
		{"Window", windowGrid},
		{"Applications", applicationsGrid},
	}

	for _, s := range sections {
		box.Append(cfgSectionLabel(s.title))
		box.Append(s.grid)
	}

	return box, f
}

func cfgSectionLabel(title string) *gtk.Label {
	lbl := gtk.NewLabel("")

	lbl.SetMarkup("<b>" + title + "</b>")
	lbl.SetXAlign(0)
	lbl.SetMarginTop(12)
	lbl.SetMarginBottom(4)
	lbl.SetMarginStart(12)

	return lbl
}

func cfgGrid() *gtk.Grid {
	g := gtk.NewGrid()

	g.SetRowSpacing(8)
	g.SetColumnSpacing(12)
	g.SetMarginStart(12)
	g.SetMarginEnd(12)
	g.SetMarginBottom(4)

	return g
}

func cfgAttach(g *gtk.Grid, row int, label string, w gtk.Widgetter) {
	lbl := gtk.NewLabel(label)

	lbl.SetHAlign(gtk.AlignEnd)

	g.Attach(lbl, 0, row, 1, 1)
	g.Attach(w, 1, row, 1, 1)
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
	s.SetNumeric(true)

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

func showConfigDialog(parent *gtk.ApplicationWindow, w *window) { //nolint:funlen
	cfg := *w.bundle.Config

	scroll, form := buildConfigForm(cfg)

	var win *gtk.Window

	cancelBtn := gtk.NewButtonWithLabel("Cancel")

	cancelBtn.ConnectClicked(func() { win.Close() })

	saveBtn := gtk.NewButtonWithLabel("Save")

	saveBtn.AddCSSClass("suggested-action")
	saveBtn.ConnectClicked(func() {
		form.updateSpinners()

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

	applyDialogSpec(win, preferencesDialogSpec())

	win.SetChild(mainBox)

	ctl := gtk.NewEventControllerKey()

	ctl.SetPropagationPhase(gtk.PhaseCapture)
	ctl.ConnectKeyPressed(func(keyval, _ uint, _ gdk.ModifierType) bool {
		switch keyval {
		case gdk.KEY_Return, gdk.KEY_KP_Enter:
			saveBtn.Activate()

			return true
		case gdk.KEY_Escape:
			win.Close()

			return true
		}

		return false
	})

	win.AddController(ctl)

	win.Present()
}
