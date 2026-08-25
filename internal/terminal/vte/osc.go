package vte

import (
	"bytes"
	"encoding/base64"
)

type terminalNotification struct {
	Title   string
	Message string
}

type oscParser struct {
	pending []byte
}

const maxPendingOSC = 64 * 1024

// oscResult holds the output of filtering a chunk of terminal data.
type oscResult struct {
	// out is the data to forward to the terminal widget.
	out []byte
	// notes are terminal notifications extracted from OSC sequences.
	notes []terminalNotification
	// clipboardText holds decoded text from an OSC 52 clipboard-write, or "".
	clipboardText string
}

func (p *oscParser) Filter(data []byte) oscResult {
	data = p.prependPending(data)

	out := make([]byte, 0, len(data))
	notes := make([]terminalNotification, 0)
	clipboardText := ""

	for len(data) > 0 {
		idx := bytes.Index(data, []byte{0x1b, ']'})

		if idx < 0 {
			out = append(out, data...)

			break
		}

		out = append(out, data[:idx]...)
		data = data[idx:]

		end, termLen := oscTerminator(data[2:])

		if end < 0 {
			if len(data) > maxPendingOSC {
				out = append(out, data[0])
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
			out = append(out, data[:seqEnd]...)
		}

		data = data[seqEnd:]
	}

	return oscResult{out: out, notes: notes, clipboardText: clipboardText}
}

// prependPending merges any buffered partial sequence with new data.
func (p *oscParser) prependPending(data []byte) []byte {
	if len(p.pending) == 0 {
		return data
	}

	combined := make([]byte, 0, len(p.pending)+len(data))
	combined = append(combined, p.pending...)
	combined = append(combined, data...)

	p.pending = nil

	return combined
}

// classifyOSC inspects an OSC sequence's content and returns:
//   - keep: whether the sequence should be forwarded to the terminal widget.
//   - note: a terminal notification, or nil if the sequence is not a notification.
//   - clipText: decoded clipboard text from an OSC 52 write, or "" if not a clipboard write.
func classifyOSC(content []byte) (bool, *terminalNotification, string) {
	if n, ok := parseNotificationOSC(content); ok {
		return false, &n, ""
	}

	if text, ok := parseClipboardOSC(content); ok {
		// VTE does not implement OSC 52 — strip it and handle the clipboard ourselves.
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

func parseNotificationOSC(content []byte) (terminalNotification, bool) {
	parts := bytes.SplitN(content, []byte(";"), 4)

	if len(parts) >= 3 && string(parts[0]) == "777" && string(parts[1]) == "notify" {
		note := terminalNotification{Title: string(parts[2])}

		if len(parts) == 4 {
			note.Message = string(parts[3])
		}

		return note, true
	}

	parts = bytes.SplitN(content, []byte(";"), 3)

	if len(parts) >= 2 && string(parts[0]) == "9" && string(parts[1]) != "4" {
		return terminalNotification{Message: string(content[2:])}, true
	}

	return terminalNotification{}, false
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
	// "?" is a clipboard query, empty is a clear — neither is a copy.
	if len(data) == 0 || string(data) == "?" {
		return "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", false
	}

	return string(decoded), true
}
