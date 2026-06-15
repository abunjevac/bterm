package ui

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/theme"
)

//go:embed bterm.css.tmpl
var cssTmpl string

// applyStyle installs a CSS provider deriving chrome colors from the theme accent.
// For the system-default theme it installs nothing (keeps host GTK theme).
func applyStyle(p *theme.Palette) {
	if p == nil || p.UseSystemDefault {
		return
	}

	t := template.Must(template.New("css").Parse(cssTmpl))

	var buf bytes.Buffer

	if err := t.Execute(&buf, p); err != nil {
		return
	}

	provider := gtk.NewCSSProvider()

	provider.LoadFromString(buf.String())

	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		provider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)
}
