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
	"os"
	"sync"
	"sync/atomic"
	"time"
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

	frontend *os.File
	backend  *os.File

	writeMu   sync.Mutex
	closeOnce sync.Once

	columns int
	rows    int

	onTitle        func(string)
	onNotification func(title, message string)
	onExit         func(int)
}

//nolint:gochecknoglobals
var (
	reg   = make(map[int]*Terminal)
	regMu sync.Mutex

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
		ptr:     ptr,
		widget:  w,
		id:      int(nextID.Add(1)),
		columns: 80,
		rows:    24,
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
	argv := buildArgv(shell, args)
	defer freeArgv(argv)

	var cwd *C.char

	if workingDir != "" {
		cwd = C.CString(workingDir)
		defer C.free(unsafe.Pointer(cwd)) //nolint:nlreturn // cgo deferred free, not a return
	}

	spawn := C.vteSpawnProxy(t.ptr, cwd, &argv[0], C.int(t.columns), C.int(t.rows))
	if spawn.err_msg != nil {
		err := errors.New(C.GoString(spawn.err_msg))
		C.vteFreeError(spawn.err_msg)
		cb(-1, err)

		return
	}

	t.frontend = os.NewFile(uintptr(spawn.frontend_slave_fd), "bterm-vte-frontend")
	t.backend = os.NewFile(uintptr(spawn.backend_master_fd), "bterm-shell-pty")

	go t.copyFrontendToBackend()
	go t.copyBackendToFrontend()
	go t.syncBackendSize()

	cb(int(spawn.pid), nil)
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
	t.columns = columns
	t.rows = rows

	C.vteSetSize(t.ptr, C.int(columns), C.int(rows))

	if t.backend != nil {
		C.vteSetPtySize(C.int(t.backend.Fd()), C.int(columns), C.int(rows))
	}
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

// FeedChild writes data to the terminal's child process as if typed.
// Used to inject bytes (e.g. a newline) that VTE's legacy key encoding
// cannot represent, such as Shift+Enter.
func (t *Terminal) FeedChild(data []byte) {
	if len(data) == 0 {
		return
	}

	if t.backend != nil {
		_ = t.writeBackend(data)

		return
	}

	// vte_terminal_feed_child copies the buffer synchronously, so passing a
	// pointer into Go memory is safe — C does not retain it after the call.
	C.vteFeedChild(t.ptr, (*C.char)(unsafe.Pointer(&data[0])), C.int(len(data)))
}

// Copy copies the selected text to the clipboard.
func (t *Terminal) Copy() { C.vteCopyClipboard(t.ptr) }

// Paste pastes clipboard contents into the terminal.
func (t *Terminal) Paste() { C.vtePasteClipboard(t.ptr) }

// OnTitleChanged sets the callback invoked when the terminal title changes.
func (t *Terminal) OnTitleChanged(f func(string)) { t.onTitle = f }

// OnNotification sets the callback invoked when the shell requests a terminal notification.
func (t *Terminal) OnNotification(f func(title, message string)) { t.onNotification = f }

// OnChildExited sets the callback invoked when the shell process exits.
func (t *Terminal) OnChildExited(f func(int)) { t.onExit = f }

func (t *Terminal) copyFrontendToBackend() {
	buf := make([]byte, 32*1024)

	for {
		n, err := t.frontend.Read(buf)
		if n > 0 {
			_ = t.writeBackend(buf[:n])
		}

		if err != nil {
			return
		}
	}
}

func (t *Terminal) copyBackendToFrontend() {
	var parser oscParser
	buf := make([]byte, 32*1024)

	for {
		n, err := t.backend.Read(buf)
		if n > 0 {
			out, notes := parser.Filter(buf[:n])
			for _, note := range notes {
				if t.onNotification != nil {
					t.onNotification(note.Title, note.Message)
				}
			}

			if len(out) > 0 {
				_, _ = t.frontend.Write(out)
			}
		}

		if err != nil {
			return
		}
	}
}

func (t *Terminal) syncBackendSize() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	lastColumns := t.columns
	lastRows := t.rows

	for range ticker.C {
		if t.frontend == nil || t.backend == nil {
			return
		}

		var columns C.int
		var rows C.int
		if C.vteGetPtySize(C.int(t.frontend.Fd()), &columns, &rows) == 0 {
			return
		}

		goColumns := int(columns)
		goRows := int(rows)
		if goColumns <= 0 || goRows <= 0 || (goColumns == lastColumns && goRows == lastRows) {
			continue
		}

		lastColumns = goColumns
		lastRows = goRows
		C.vteSetPtySize(C.int(t.backend.Fd()), columns, rows)
	}
}

func (t *Terminal) writeBackend(data []byte) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	for len(data) > 0 {
		n, err := t.backend.Write(data)
		if err != nil {
			return err
		}

		data = data[n:]
	}

	return nil
}

func (t *Terminal) closeProxy() {
	t.closeOnce.Do(func() {
		if t.frontend != nil {
			_ = t.frontend.Close()
		}

		if t.backend != nil {
			_ = t.backend.Close()
		}
	})
}

// --- C-exported callbacks ---

//export goVteChildExited
func goVteChildExited(termID C.int, status C.int) {
	regMu.Lock()
	t := reg[int(termID)]
	regMu.Unlock()

	if t == nil {
		return
	}

	t.closeProxy()

	if t.onExit == nil {
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
