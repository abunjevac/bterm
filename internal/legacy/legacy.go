package legacy

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Key is a key whose encoding differs in legacy mode.
type Key int

const (
	Backspace Key = iota
	Delete
	Home
	End
)

// Sequence returns the bytes a legacy program expects for key, or nil when
// program is not in programs. Legacy programs expect ^H for Backspace and DEL
// for Delete, and vt220-style ESC[1~ and ESC[4~ for Home and End, instead of
// what xterm-compatible terminals send.
func Sequence(key Key, program string, programs []string) []byte {
	if !slices.Contains(programs, program) {
		return nil
	}

	switch key {
	case Backspace:
		return []byte{0x08}
	case Delete:
		return []byte{0x7f}
	case Home:
		return []byte("\x1b[1~")
	case End:
		return []byte("\x1b[4~")
	default:
		return nil
	}
}

// ProgramName returns the kernel command name (comm) of process pid. The
// kernel truncates it to 15 characters.
func ProgramName(pid int) (string, error) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm")
	if err != nil {
		return "", fmt.Errorf("read process name: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}
