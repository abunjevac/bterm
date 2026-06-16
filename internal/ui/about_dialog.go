package ui

import (
	"fmt"
	"os"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/assets"
	"github.com/abunjevac/bterm/internal/version"
)

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
