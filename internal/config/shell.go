package config

// InferShell picks the shell: explicit config value, else the provided env value
// (caller passes os.Getenv("SHELL")), else /bin/zsh. The env arg is a parameter to
// keep this function pure and testable.
func InferShell(configShell, envShell string) string {
	if configShell != "" {
		return configShell
	}

	if envShell != "" {
		return envShell
	}

	return "/bin/zsh"
}
