package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/abunjevac/bterm/internal/config"
	"github.com/abunjevac/bterm/internal/terminal"
	"github.com/diamondburned/gotk4/pkg/core/glib"
)

const (
	btermAppName      = "bterm"
	btermDesktopEntry = "io.github.abunjevac.bterm"
)

type terminalNotification struct {
	Title   string
	Message string
}

func (w *window) installTerminalNotifications(t terminal.Terminal) {
	cfg := w.bundle.Config

	if cfg == nil || cfg.TerminalNotificationMethod != config.TerminalNotificationDBus {
		return
	}

	t.OnNotification(func(title, message string) {
		w.handleTerminalNotification(terminalNotification{Title: title, Message: message})
	})
}

// installClipboardDetection wires the OSC 52 clipboard-copy callback.
// VTE does not implement OSC 52, so bterm decodes the base64 payload and
// writes directly to the GDK clipboard, then shows a toast.
func (w *window) installClipboardDetection(t terminal.Terminal) {
	t.OnClipboardCopy(func(text string) {
		glib.IdleAdd(func() bool {
			if clip := w.clipboard(); clip != nil {
				clip.SetText(text)
			}

			w.toast.show("⧉ Copied")

			return false
		})
	})
}

func (w *window) handleTerminalNotification(n terminalNotification) {
	cfg := w.bundle.Config

	if cfg == nil || cfg.TerminalNotificationMethod != config.TerminalNotificationDBus {
		return
	}

	title := cleanNotificationArg(n.Title)
	message := cleanNotificationArg(n.Message)

	if title == "" {
		title = cfg.Title
	}

	if message == "" {
		message = title
		title = cfg.Title
	}

	argv := notificationDBusArgs(title, message)

	go func() {
		output, err := exec.CommandContext(context.Background(), "gdbus", argv...).CombinedOutput()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "bterm: terminal notification: %v: %s\n", err, strings.TrimSpace(string(output)))
		}
	}()
}

func notificationDBusArgs(title, message string) []string {
	return []string{
		"call",
		"--session",
		"--dest", "org.freedesktop.Notifications",
		"--object-path", "/org/freedesktop/Notifications",
		"--method", "org.freedesktop.Notifications.Notify",
		gvariantString(btermAppName),
		"0",
		gvariantString(btermDesktopEntry),
		gvariantString(title),
		gvariantString(message),
		"[]",
		"{'desktop-entry': <" + gvariantString(btermDesktopEntry) + ">}",
		"int32 -1",
	}
}

func gvariantString(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `'`, `\'`)

	return `'` + replacer.Replace(value) + `'`
}

func cleanNotificationArg(value string) string {
	return strings.TrimSpace(strings.ReplaceAll(value, "\x00", ""))
}
