// Package assets embeds application assets used at runtime.
package assets

import _ "embed"

// IconPNG is the full-size bterm application icon.
//
//go:embed io.github.abunjevac.bterm.png
var IconPNG []byte
