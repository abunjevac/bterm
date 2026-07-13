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
// The system-default theme keeps host GTK colors while still enabling app CSS classes.
func applyStyle(p *theme.Palette) {
	if p == nil || p.UseSystemDefault {
		p = &theme.Palette{
			Foreground: "@theme_fg_color",
			Background: "@theme_bg_color",
			Accent:     "@theme_selected_bg_color",
		}
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
