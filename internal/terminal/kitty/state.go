package kitty

import "strconv"

// maxStackDepth limits the push/pop stack to prevent denial-of-service.
const maxStackDepth = 64

// State tracks the kitty keyboard protocol flags for a single terminal.
// Each terminal maintains its own state, negotiated at runtime by the
// application via CSI sequences.
type State struct {
	flags int
	stack []int
}

// Flags returns the current progressive enhancement flags.
func (s *State) Flags() int { return s.flags }

// Disambiguate reports whether the disambiguate flag is active.
func (s *State) Disambiguate() bool { return s.flags&FlagDisambiguate != 0 }

// Set applies new flags according to the mode parameter:
//   - mode 1: set all set bits, clear all unset bits (replace)
//   - mode 2: set all set bits, leave unset bits unchanged (union)
//   - mode 3: clear all set bits, leave unset bits unchanged (subtract)
func (s *State) Set(flags, mode int) {
	switch mode {
	case 2:
		s.flags |= flags
	case 3:
		s.flags &^= flags
	default:
		s.flags = flags
	}
}

// Push saves the current flags onto the stack and applies newFlags.
// If newFlags is omitted (zero), the flags are cleared.
func (s *State) Push(newFlags int) {
	if len(s.stack) >= maxStackDepth {
		// Evict oldest entry to prevent unbounded growth.
		s.stack = s.stack[1:]
	}

	s.stack = append(s.stack, s.flags)
	s.flags = newFlags
}

// Pop removes count entries from the stack. The flags become the value
// at the popped level. If count exceeds the stack size, the stack is
// emptied and flags are reset to zero.
func (s *State) Pop(count int) {
	if count <= 0 {
		count = 1
	}

	if count > len(s.stack) {
		s.flags = 0
		s.stack = s.stack[:0]

		return
	}

	idx := len(s.stack) - count
	s.flags = s.stack[idx]
	s.stack = s.stack[:idx]
}

// QueryResponse returns the CSI ? flags u response for a capability query.
func (s *State) QueryResponse() string {
	return "\x1b[?" + strconv.Itoa(s.flags) + "u"
}
