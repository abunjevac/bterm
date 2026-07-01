package vte

import "bytes"

type terminalNotification struct {
	Title   string
	Message string
}

type oscParser struct {
	pending []byte
}

const maxPendingOSC = 64 * 1024

func (p *oscParser) Filter(data []byte) ([]byte, []terminalNotification) {
	if len(p.pending) > 0 {
		combined := make([]byte, 0, len(p.pending)+len(data))

		combined = append(combined, p.pending...)
		combined = append(combined, data...)

		data = combined

		p.pending = nil
	}

	out := make([]byte, 0, len(data))
	notes := make([]terminalNotification, 0)

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

		if note, ok := parseNotificationOSC(content); ok {
			notes = append(notes, note)
		} else {
			out = append(out, data[:seqEnd]...)
		}

		data = data[seqEnd:]
	}

	return out, notes
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
