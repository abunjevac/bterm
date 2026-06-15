package ui

import (
	"context"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/config"
)

// Run starts the GTK application and blocks until the window closes.
func Run(ctx context.Context, bundle *config.Bundle) {
	app := gtk.NewApplication("io.github.abunjevac.bterm", gio.ApplicationNonUnique)

	app.ConnectActivate(func() {
		w := newWindow(ctx, app, bundle)

		w.Present()
	})

	os.Exit(app.Run(os.Args[:1]))
}
