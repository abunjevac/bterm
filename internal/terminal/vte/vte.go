// Package vte implements terminal.Terminal via the VTE cgo bridge.
package vte

// #cgo pkg-config: vte-2.91-gtk4
// #include "vte.h"
// #include <stdlib.h>
import "C"

import (
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/abunjevac/bterm/internal/terminal"
	"github.com/abunjevac/bterm/internal/theme"
	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// Terminal implements terminal.Terminal via the VTE cgo bridge.
type Terminal struct {
	ptr    *C.VteTerminal
	widget gtk.Widgetter
	id     int

	onTitle func(string)
	onExit  func(int)
}

//nolint:gochecknoglobals
var (
	reg   = make(map[int]*Terminal)
	regMu sync.Mutex

	spawnMu       sync.Mutex
	spawnRegistry = make(map[int]func(int, error))
	nextSpawnID   atomic.Int64

	nextID atomic.Int64
)

// New creates a VTE terminal widget and returns a *Terminal implementing terminal.Terminal.
func New() *Terminal {
	raw := C.vte_terminal_new()
	ptr := (*C.VteTerminal)(unsafe.Pointer(raw))

	obj := coreglib.Take(unsafe.Pointer(raw))

	result := obj.WalkCast(func(o coreglib.Objector) bool {
		_, ok := o.(gtk.Widgetter)

		return ok
	})

	w, ok := result.(gtk.Widgetter)
	if !ok {
		panic("vte: VTE widget does not implement gtk.Widgetter")
	}

	t := &Terminal{
		ptr:    ptr,
		widget: w,
		id:     int(nextID.Add(1)),
	}

	regMu.Lock()
	reg[t.id] = t
	regMu.Unlock()

	C.vteConnectTitleChanged(ptr, C.int(t.id))
	C.vteConnectChildExited(ptr, C.int(t.id))

	return t
}

// Widget returns the underlying GTK widget for embedding in a layout.
func (t *Terminal) Widget() gtk.Widgetter { return t.widget }

// Spawn starts shell asynchronously. cb is called on the GTK main thread.
func (t *Terminal) Spawn(workingDir, shell string, args []string, cb terminal.SpawnCallback) {
	id := int(nextSpawnID.Add(1))

	spawnMu.Lock()
	spawnRegistry[id] = cb
	spawnMu.Unlock()

	argv := buildArgv(shell, args)
	defer freeArgv(argv)

	var cwd *C.char

	if workingDir != "" {
		cwd = C.CString(workingDir)
		defer C.free(unsafe.Pointer(cwd)) //nolint:nlreturn // cgo deferred free, not a return
	}

	C.vteSpawnAsync(t.ptr, cwd, &argv[0], C.int(id))
}

// SetFont applies a Pango font description (e.g. family="Monospace", size=12).
func (t *Terminal) SetFont(family string, size float64) {
	desc := fmt.Sprintf("%s %g", family, size)

	cs := C.CString(desc)
	defer C.free(unsafe.Pointer(cs)) //nolint:nlreturn // cgo deferred free, not a return

	C.vteSetFont(t.ptr, cs)
}

// SetScrollback configures the scrollback buffer size in lines.
func (t *Terminal) SetScrollback(lines int) {
	C.vteSetScrollback(t.ptr, C.long(lines))
}

// SetSize sets the terminal's preferred size in character columns and rows.
// Call before the window is presented so GTK sizes the window to fit.
func (t *Terminal) SetSize(columns, rows int) {
	C.vteSetSize(t.ptr, C.int(columns), C.int(rows))
}

// SetColors applies the palette to the terminal. A nil or UseSystemDefault palette is a no-op.
func (t *Terminal) SetColors(p *theme.Palette) {
	if p == nil || p.UseSystemDefault {
		return
	}

	if len(p.Palette) == 0 {
		return
	}

	fg := C.CString(p.Foreground)
	bg := C.CString(p.Background)
	cur := C.CString(p.Cursor)
	defer C.free(unsafe.Pointer(fg))  //nolint:nlreturn
	defer C.free(unsafe.Pointer(bg))  //nolint:nlreturn
	defer C.free(unsafe.Pointer(cur)) //nolint:nlreturn

	cpal := make([]*C.char, len(p.Palette))
	for i, c := range p.Palette {
		cpal[i] = C.CString(c)
	}

	defer func() {
		for _, c := range cpal {
			C.free(unsafe.Pointer(c))
		}
	}()

	C.vteSetColors(t.ptr, fg, bg, cur, &cpal[0], C.int(len(cpal)))
}

// CurrentDir returns the shell's current working directory, decoded from the OSC 7 URI.
// Returns "" when the terminal has not yet reported a directory.
func (t *Terminal) CurrentDir() string {
	uri := C.vteGetCurrentDirUri(t.ptr) //nolint:nlreturn

	if uri == nil {
		return ""
	}

	u, err := url.Parse(C.GoString(uri))
	if err != nil {
		return ""
	}

	return u.Path
}

// Copy copies the selected text to the clipboard.
func (t *Terminal) Copy() { C.vteCopyClipboard(t.ptr) }

// Paste pastes clipboard contents into the terminal.
func (t *Terminal) Paste() { C.vtePasteClipboard(t.ptr) }

// OnTitleChanged sets the callback invoked when the terminal title changes.
func (t *Terminal) OnTitleChanged(f func(string)) { t.onTitle = f }

// OnChildExited sets the callback invoked when the shell process exits.
func (t *Terminal) OnChildExited(f func(int)) { t.onExit = f }

// --- C-exported callbacks ---

//export goVteSpawnDone
func goVteSpawnDone(callbackID C.int, pid C.int, _ C.int, errMsg *C.char) {
	id := int(callbackID)

	spawnMu.Lock()
	cb, ok := spawnRegistry[id]
	delete(spawnRegistry, id)
	spawnMu.Unlock()

	if !ok {
		return
	}

	var goErr error

	if errMsg != nil {
		goErr = errors.New(C.GoString(errMsg))
	}

	cb(int(pid), goErr)
}

//export goVteChildExited
func goVteChildExited(termID C.int, status C.int) {
	regMu.Lock()
	t := reg[int(termID)]
	regMu.Unlock()

	if t == nil || t.onExit == nil {
		return
	}

	t.onExit(int(status))
}

//export goVteTitleChanged
func goVteTitleChanged(termID C.int) {
	regMu.Lock()
	t := reg[int(termID)]
	regMu.Unlock()

	if t == nil || t.onTitle == nil {
		return
	}

	ctitle := C.vteGetWindowTitle(t.ptr) //nolint:nlreturn

	title := ""
	if ctitle != nil {
		title = C.GoString(ctitle)
		C.free(unsafe.Pointer(ctitle))
	}

	t.onTitle(title)
}

// --- helpers ---

func buildArgv(shell string, args []string) []*C.char {
	argv := make([]*C.char, 0, len(args)+2)

	argv = append(argv, C.CString(shell))

	for _, a := range args {
		argv = append(argv, C.CString(a))
	}

	argv = append(argv, nil)

	return argv
}

func freeArgv(argv []*C.char) {
	for _, p := range argv {
		if p != nil {
			C.free(unsafe.Pointer(p))
		}
	}
}
