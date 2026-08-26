package vte

import (
	"bytes"
	"encoding/base64"

	"github.com/abunjevac/bterm/internal/terminal"
)

const maxPendingOSC = 64 * 1024

type oscParser struct {
	pending []byte
	out     []byte
}

// oscResult holds the output of filtering a chunk of terminal data.
type oscResult struct {
	out           []byte                  // data to forward to the terminal widget
	notes         []terminal.Notification // terminal notifications extracted from OSC sequences
	clipboardText string                  // decoded text from an OSC 52 clipboard-write, or ""
}

func (p *oscParser) Filter(data []byte) oscResult {
	data = prependPending(&p.pending, data)

	p.out = p.out[:0]
	notes := make([]terminal.Notification, 0)
	clipboardText := ""

	for len(data) > 0 {
		idx := bytes.Index(data, []byte{0x1b, ']'})

		if idx < 0 {
			p.out = append(p.out, data...)

			break
		}

		p.out = append(p.out, data[:idx]...)

		data = data[idx:]

		end, termLen := oscTerminator(data[2:])

		if end < 0 {
			if len(data) > maxPendingOSC {
				p.out = append(p.out, data[0])

				data = data[1:]

				continue
			}

			p.pending = append(p.pending, data...)

			break
		}

		content := data[2 : 2+end]
		seqEnd := 2 + end + termLen

		keep, note, clipText := classifyOSC(content)

		if note != nil {
			notes = append(notes, *note)
		}

		if clipText != "" {
			clipboardText = clipText
		}

		if keep {
			p.out = append(p.out, data[:seqEnd]...)
		}

		data = data[seqEnd:]
	}

	return oscResult{out: p.out, notes: notes, clipboardText: clipboardText}
}

// prependPending merges any buffered partial data with new data, clearing the buffer.
func prependPending(pending *[]byte, data []byte) []byte {
	if len(*pending) == 0 {
		return data
	}

	combined := make([]byte, 0, len(*pending)+len(data))
	combined = append(combined, *pending...)
	combined = append(combined, data...)

	*pending = nil

	return combined
}

// classifyOSC inspects an OSC sequence's content and returns:
//   - keep: whether the sequence should be forwarded to the terminal widget.
//   - note: a terminal notification, or nil if the sequence is not a notification.
//   - clipText: decoded clipboard text from an OSC 52 write, or "" if not a clipboard write.
func classifyOSC(content []byte) (bool, *terminal.Notification, string) {
	if n, ok := parseNotificationOSC(content); ok {
		return false, &n, ""
	}

	if text, ok := parseClipboardOSC(content); ok {
		// VTE does not implement OSC 52 — strip it and handle the clipboard ourselves
		return false, nil, text
	}

	return true, nil, ""
}

func oscTerminator(data []byte) (int, int) {
	bel := bytes.IndexByte(data, 0x07)
	st := bytes.Index(data, []byte{0x1b, '\\'})

	switch {
	case bel < 0 && st < 0:
		return -1, 0
	case bel < 0:
		return st, 2
	case st < 0:
		return bel, 1
	case bel < st:
		return bel, 1
	default:
		return st, 2
	}
}

func parseNotificationOSC(content []byte) (terminal.Notification, bool) {
	parts := bytes.SplitN(content, []byte(";"), 4)

	if len(parts) >= 3 && string(parts[0]) == "777" && string(parts[1]) == "notify" {
		note := terminal.Notification{Title: string(parts[2])}

		if len(parts) == 4 {
			note.Message = string(parts[3])
		}

		return note, true
	}

	parts = bytes.SplitN(content, []byte(";"), 2)

	// OSC 9;<message> is a simple notification. Exclude OSC 9;4;... which is
	// the progress reporting protocol (iTerm2/ConEmu), not a notification.
	// A bare "9;4" (no third part) is a notification with body "4".
	if len(parts) >= 2 && string(parts[0]) == "9" && !bytes.HasPrefix(parts[1], []byte("4;")) {
		return terminal.Notification{Message: string(parts[1])}, true
	}

	return terminal.Notification{}, false
}

// parseClipboardOSC decodes an OSC 52 clipboard-write sequence.
// Format: 52;c;<base64-data> or 52;<selector>;<base64-data>. A query (data
// is "?") or clear (data is empty) returns false.
func parseClipboardOSC(content []byte) (string, bool) {
	parts := bytes.SplitN(content, []byte(";"), 3)

	if len(parts) < 3 || string(parts[0]) != "52" {
		return "", false
	}

	data := parts[2]
	// "?" is a clipboard query, empty is a clear — neither is a copy
	if len(data) == 0 || string(data) == "?" {
		return "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", false
	}

	return string(decoded), true
}
