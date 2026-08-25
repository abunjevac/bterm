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

const aboutCopyright = "© 2026 exobyte.org"

//nolint:funlen // keeps the complete About dialog layout together.
func showAboutDialog(parent *gtk.ApplicationWindow) {
	const description = "An opinionated GTK4 terminal emulator"

	content := gtk.NewBox(gtk.OrientationVertical, 8)

	content.SetMarginTop(24)
	content.SetMarginBottom(16)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)

	title := gtk.NewLabel("bterm")

	title.SetHAlign(gtk.AlignCenter)
	title.AddCSSClass("bterm-about-title")

	versionLabel := gtk.NewLabel("Version " + version.Version)

	versionLabel.SetHAlign(gtk.AlignCenter)
	versionLabel.AddCSSClass("bterm-about-version")

	descriptionLabel := gtk.NewLabel(description)

	descriptionLabel.SetHAlign(gtk.AlignCenter)
	descriptionLabel.SetWrap(true)
	descriptionLabel.SetJustify(gtk.JustifyCenter)

	licenseLabel := gtk.NewLabel("MIT License")

	licenseLabel.SetHAlign(gtk.AlignCenter)

	copyrightLabel := gtk.NewLabel(aboutCopyright)

	copyrightLabel.SetHAlign(gtk.AlignCenter)

	website := gtk.NewLinkButton("https://github.com/abunjevac/bterm")

	website.SetHAlign(gtk.AlignCenter)

	content.Append(title)
	content.Append(versionLabel)

	logo, err := gdk.NewTextureFromBytes(glib.NewBytes(assets.IconPNG))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "bterm: load about logo: %v\n", err)
	} else {
		picture := gtk.NewPictureForPaintable(logo)

		picture.SetCanShrink(true)
		picture.SetHAlign(gtk.AlignCenter)
		picture.SetSizeRequest(96, 96)

		content.Append(picture)
	}

	content.Append(descriptionLabel)
	content.Append(licenseLabel)
	content.Append(copyrightLabel)
	content.Append(website)

	closeBtn := gtk.NewButtonWithLabel("Close")

	closeBtn.AddCSSClass("suggested-action")

	footer := gtk.NewBox(gtk.OrientationHorizontal, 8)

	footer.SetHAlign(gtk.AlignEnd)
	footer.SetMarginTop(8)
	footer.Append(closeBtn)

	content.Append(footer)

	win := gtk.NewWindow()

	win.SetTitle("About bterm")
	win.SetTransientFor(&parent.Window)
	win.SetModal(true)

	applyDialogSpec(win, aboutDialogSpec())

	win.SetChild(content)

	closeBtn.ConnectClicked(func() { win.Close() })

	addEscapeClose(win)

	win.Present()
}
