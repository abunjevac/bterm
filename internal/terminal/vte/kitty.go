package vte

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/abunjevac/bterm/internal/terminal/kitty"
)

const maxPendingKitty = 256

// kittyEnabled controls whether the kitty keyboard protocol is active.
// Set to false to completely disable the feature: negotiation sequences
// pass through to VTE (which ignores them) and the key encoder never fires.
// This is a const so the compiler eliminates the dead branch at build time.
const kittyEnabled = true

// kittyParser filters kitty keyboard protocol negotiation sequences from
// the data stream. These CSI sequences would be misinterpreted by VTE,
// so they must be stripped before forwarding output to the terminal widget.
//
// Only sequences whose CSI final byte is 'u' are treated as kitty keyboard
// protocol sequences. Other CSI sequences that share the same prefix bytes
// (=, ?, >, <) — such as DEC private modes (CSI ? ... h/l), SGR mouse
// events (CSI < ... M), and xterm modifier reports (CSI > ...) — are passed
// through untouched.
type kittyParser struct {
	state   kitty.State
	pending []byte
}

// kittyResult holds the output of filtering kitty negotiation sequences.
type kittyResult struct {
	// out is the data to forward to the terminal widget.
	out []byte
	// response is a reply to send back to the shell, or nil.
	response []byte
}

// Filter removes kitty keyboard protocol negotiation sequences from data.
// It updates the internal state and prepares responses for query sequences.
// When kittyEnabled is false, data passes through unchanged.
func (p *kittyParser) Filter(data []byte) kittyResult {
	if !kittyEnabled {
		return kittyResult{out: data}
	}

	data = p.prependPending(data)

	out := make([]byte, 0, len(data))
	response := []byte{}

	for len(data) > 0 {
		idx := bytes.Index(data, []byte{0x1b, '['})

		if idx < 0 {
			out = append(out, data...)

			break
		}

		// Not enough data to read the prefix byte.
		if idx+2 >= len(data) {
			p.pending = append(p.pending, data...)

			break
		}

		// Scan forward from after ESC [ to find the CSI final byte.
		// CSI final bytes are in the range 0x40-0x7e.
		finalOff := findCSIFinal(data, idx+2)

		if finalOff < 0 {
			if len(data[idx:]) > maxPendingKitty {
				// Too long — not a valid CSI, pass ESC [ through.
				out = append(out, data[idx], data[idx+1])
				data = data[idx+2:]

				continue
			}

			out = append(out, data[:idx]...)
			p.pending = append(p.pending, data[idx:]...)

			break
		}

		// Check if this is a kitty keyboard protocol sequence.
		// Kitty sequences have a private-marker prefix (=, ?, >, <) and
		// final byte 'u'. Everything else passes through to VTE.
		prefix := data[idx+2]

		if !isKittyCSIPrefix(prefix) || data[finalOff] != 'u' {
			out = append(out, data[:finalOff+1]...)
			data = data[finalOff+1:]

			continue
		}

		out = append(out, data[:idx]...)
		content := data[idx+3 : finalOff]
		response = append(response, p.handleKittyCSI(prefix, content)...)

		data = data[finalOff+1:]
	}

	return kittyResult{out: out, response: response}
}

// findCSIFinal returns the offset of the CSI final byte (0x40-0x7e)
// starting from offset start, or -1 if not found.
func findCSIFinal(data []byte, start int) int {
	for i := start; i < len(data); i++ {
		if data[i] >= 0x40 {
			return i
		}
	}

	return -1
}

// Flags returns the current kitty keyboard protocol flags.
func (p *kittyParser) Flags() int { return p.state.Flags() }

// Disambiguate reports whether the disambiguate flag is active.
func (p *kittyParser) Disambiguate() bool {
	if !kittyEnabled {
		return false
	}

	return p.state.Disambiguate()
}

// handleKittyCSI processes a kitty keyboard protocol negotiation sequence
// and returns an optional response string.
func (p *kittyParser) handleKittyCSI(prefix byte, content []byte) string {
	s := string(content)

	switch prefix {
	case '=':
		// CSI = flags ; mode u — set flags.
		return p.handleSetFlags(s)
	case '?':
		// CSI ? u with empty content is a query. CSI ? <flags> u is a
		// response (e.g. echoed by the shell) — strip it without
		// generating a new response to avoid an echo loop.
		if s == "" {
			return p.state.QueryResponse()
		}

		return ""
	case '>':
		// CSI > flags u — push flags.
		return p.handlePushFlags(s)
	case '<':
		// CSI < count u — pop flags.
		return p.handlePopFlags(s)
	default:
		return ""
	}
}

func (p *kittyParser) handleSetFlags(s string) string {
	parts := strings.SplitN(s, ";", 2)

	flags, err := strconv.Atoi(parts[0])
	if err != nil {
		return ""
	}

	mode := 1 // default: replace.

	if len(parts) == 2 {
		if m, err := strconv.Atoi(parts[1]); err == nil {
			mode = m
		}
	}

	p.state.Set(flags, mode)

	return ""
}

func (p *kittyParser) handlePushFlags(s string) string {
	flags, err := strconv.Atoi(s)
	if err != nil {
		flags = 0
	}

	p.state.Push(flags)

	return ""
}

func (p *kittyParser) handlePopFlags(s string) string {
	count, err := strconv.Atoi(s)
	if err != nil {
		count = 1
	}

	p.state.Pop(count)

	return ""
}

func (p *kittyParser) prependPending(data []byte) []byte {
	if len(p.pending) == 0 {
		return data
	}

	combined := make([]byte, 0, len(p.pending)+len(data))
	combined = append(combined, p.pending...)
	combined = append(combined, data...)

	p.pending = nil

	return combined
}

// isKittyCSIPrefix reports whether byte is a kitty keyboard protocol
// CSI prefix: =, ?, >, or <.
func isKittyCSIPrefix(b byte) bool {
	return b == '=' || b == '?' || b == '>' || b == '<'
}
