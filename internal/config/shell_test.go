package config_test

import (
	"testing"

	"github.com/abunjevac/bterm/internal/config"
	"github.com/stretchr/testify/require"
)

func TestInferShellPrefersConfig(t *testing.T) {
	require.Equal(t, "/bin/fish", config.InferShell("/bin/fish", "/bin/bash"))
}

func TestInferShellFallsBackToEnv(t *testing.T) {
	require.Equal(t, "/bin/bash", config.InferShell("", "/bin/bash"))
}

func TestInferShellDefaultsToZsh(t *testing.T) {
	require.Equal(t, "/bin/zsh", config.InferShell("", ""))
}
