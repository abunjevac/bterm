package ui

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/abunjevac/bterm/internal/logbuf"
)

// showLogDialog opens a window listing the terminal bridge's diagnostic
// events recorded by the process-wide logbuf ring.
func showLogDialog(parent *gtk.ApplicationWindow, w *window) {
	win := gtk.NewWindow()

	win.SetTitle("Diagnostics")
	win.SetTransientFor(&parent.Window)

	applyDialogSpec(win, logDialogSpec())

	content, copyBtn, closeBtn := logDialogContent(w)

	win.SetChild(content)

	copyBtn.ConnectClicked(func() {
		w.copyText(logText(w))
		w.toast.show("⧉ Copied")
	})
	closeBtn.ConnectClicked(func() { win.Close() })

	addEscapeClose(win)

	win.Present()
}

// logText renders the layout state plus the current logbuf contents, with a
// placeholder when empty.
func logText(w *window) string {
	lines := make([]string, 0, 2+len(logbuf.Lines()))

	lines = append(lines, layoutState(w), "")
	lines = append(lines, logbuf.Lines()...)

	text := strings.Join(lines, "\n")

	if len(logbuf.Lines()) == 0 {
		text += "\nNo diagnostic events recorded yet."
	}

	return text
}

// layoutState summarizes every tab and pane widget state: mapped status and
// allocation, so a wedged (black, dead) pane is visible at a glance.
func layoutState(w *window) string {
	var b strings.Builder

	b.WriteString("== layout ==")

	for i, tab := range w.tabs {
		for _, id := range tab.area.tree.Leaves() {
			widget := tab.area.widgets[id]

			if widget == nil {
				fmt.Fprintf(&b, "\ntab %d pane %d: WIDGET MISSING", i+1, id)

				continue
			}

			bw := gtk.BaseWidget(widget)

			fmt.Fprintf(&b, "\ntab %d pane %d: mapped=%v alloc=%dx%d",
				i+1, id, bw.Mapped(), bw.Width(), bw.Height())
		}
	}

	return b.String()
}

// logDialogContent builds the scrollable log view plus footer buttons.
func logDialogContent(w *window) (*gtk.Box, *gtk.Button, *gtk.Button) {
	view := gtk.NewTextView()

	view.SetMonospace(true)
	view.SetEditable(false)
	view.SetCursorVisible(false)
	view.Buffer().SetText(logText(w))

	scroll := gtk.NewScrolledWindow()

	scroll.SetVExpand(true)
	scroll.SetChild(view)
	scroll.SetMarginTop(8)
	scroll.SetMarginBottom(8)
	scroll.SetMarginStart(8)
	scroll.SetMarginEnd(8)

	// scroll to the bottom, where the newest events are queued
	scroll.ConnectMap(func() {
		adj := scroll.VAdjustment()

		adj.SetValue(adj.Upper())
	})

	copyBtn := gtk.NewButtonWithLabel("Copy")

	closeBtn := gtk.NewButtonWithLabel("Close")

	closeBtn.AddCSSClass("suggested-action")

	footer := gtk.NewBox(gtk.OrientationHorizontal, 8)

	footer.SetHAlign(gtk.AlignEnd)
	footer.SetMarginTop(4)
	footer.SetMarginBottom(12)
	footer.SetMarginStart(12)
	footer.SetMarginEnd(12)
	footer.Append(copyBtn)
	footer.Append(closeBtn)

	content := gtk.NewBox(gtk.OrientationVertical, 0)

	content.Append(scroll)
	content.Append(footer)

	return content, copyBtn, closeBtn
}
