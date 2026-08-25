package assets

import _ "embed"

// IconPNG is the full-size bterm application icon.
//
//go:embed io.github.abunjevac.bterm.png
var IconPNG []byte

// IconsGResource contains the application-specific symbolic icons.
//
//go:generate glib-compile-resources icons.gresource.xml --target=icons.gresource --sourcedir=.
//go:embed icons.gresource
var IconsGResource []byte
