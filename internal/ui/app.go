package ui

import (
	"context"
	"fmt"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/assets"
	"github.com/abunjevac/bterm/internal/config"
)

// Run starts the GTK application and blocks until the window closes.
func Run(ctx context.Context, bundle *config.Bundle, workingDir string) {
	_ = ctx

	iconResource, err := gio.NewResourceFromData(glib.NewBytes(assets.IconsGResource))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "bterm: load icon resources: %v\n", err)
	} else {
		gio.ResourcesRegister(iconResource)
	}

	app := gtk.NewApplication("io.github.abunjevac.bterm", gio.ApplicationNonUnique)

	app.ConnectActivate(func() { //nolint:contextcheck // GTK activation callback has no context parameter; downstream notification subprocess uses background context.
		w := newWindow(app, bundle, workingDir)

		w.Present()
	})

	os.Exit(app.Run(os.Args[:1]))
}
