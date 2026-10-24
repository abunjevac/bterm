package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	workDir = "/tmp"
	prog    = "bterm"
)

func TestSplitExec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantCLI []string
		wantCmd []string
	}{
		{"no command", []string{prog, "-w", workDir}, []string{prog, "-w", workDir}, nil},
		{
			"command with own flags",
			[]string{prog, "-w", workDir, "-e", "ssh", "-t", "-p", "2222", "host", "cd /srv && exec $SHELL -l"},
			[]string{prog, "-w", workDir},
			[]string{"ssh", "-t", "-p", "2222", "host", "cd /srv && exec $SHELL -l"},
		},
		{"long form", []string{prog, "--execute", "htop"}, []string{prog}, []string{"htop"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotCLI, gotCmd := splitExec(tc.args)

			require.Equal(t, tc.wantCLI, gotCLI)
			require.Equal(t, tc.wantCmd, gotCmd)
		})
	}
}
