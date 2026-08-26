//go:build nokitty

package vte

// kittyEnabled controls whether the kitty keyboard protocol is active.
// Build with -tags nokitty to disable the kitty keyboard protocol entirely:
// negotiation sequences pass through to VTE (which ignores them) and the
// key encoder never fires. This is a const so the compiler eliminates the
// dead branch at build time.
const kittyEnabled = false
