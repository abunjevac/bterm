//go:build !nokitty

package vte

// kittyEnabled controls whether the kitty keyboard protocol is active.
// The default build has kitty enabled. Build with -tags nokitty to disable.
const kittyEnabled = true
