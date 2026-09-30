package legacy_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/abunjevac/bterm/internal/legacy"
)

const (
	telnetProgram = "telnet"
	legacyProgram = "foo"
	otherProgram  = "zsh"
)

func TestSequence(t *testing.T) {
	t.Parallel()

	programs := []string{telnetProgram, legacyProgram}

	tests := []struct {
		name    string
		key     legacy.Key
		program string
		want    []byte
	}{
		{"backspace in legacy program", legacy.Backspace, telnetProgram, []byte{0x08}},
		{"delete in legacy program", legacy.Delete, legacyProgram, []byte{0x7f}},
		{"home in legacy program", legacy.Home, telnetProgram, []byte("\x1b[1~")},
		{"end in legacy program", legacy.End, legacyProgram, []byte("\x1b[4~")},
		{"home in other program", legacy.Home, otherProgram, nil},
		{"end in other program", legacy.End, otherProgram, nil},
		{"backspace in other program", legacy.Backspace, otherProgram, nil},
		{"delete in other program", legacy.Delete, otherProgram, nil},
		{"partial name does not match", legacy.Backspace, "telnetd", nil},
		{"unknown program", legacy.Backspace, "", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, legacy.Sequence(tc.key, tc.program, programs))
		})
	}
}

func TestSequenceEmptyList(t *testing.T) {
	t.Parallel()

	require.Nil(t, legacy.Sequence(legacy.Backspace, telnetProgram, nil))
}

func TestProgramName(t *testing.T) {
	t.Parallel()

	want, err := os.ReadFile("/proc/self/comm")
	require.NoError(t, err)

	got, err := legacy.ProgramName(os.Getpid())
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(string(want)), got)
}

func TestProgramNameMissingProcess(t *testing.T) {
	t.Parallel()

	_, err := legacy.ProgramName(-1)
	require.Error(t, err)
}
