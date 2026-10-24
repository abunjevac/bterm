package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/abunjevac/bterm/internal/config"
	"github.com/abunjevac/bterm/internal/keymap"
	"github.com/abunjevac/bterm/internal/ui"
	"github.com/abunjevac/bterm/internal/version"
)

func main() {
	args, command := splitExec(os.Args)

	cmd := &cli.Command{
		Name:    "bterm",
		Version: version.Version,
		Usage:   "an opinionated GTK4 + VTE terminal emulator",
		Description: "Use -e COMMAND [ARGS...] as the last option to run a command instead of the shell;\n" +
			"everything after -e belongs to the command.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "path to config dir (default: ~/.config/bterm)",
			},
			&cli.StringFlag{
				Name:    "working-directory",
				Aliases: []string{"w"},
				Usage:   "path to working dir (default: current dir)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(ctx, cmd, command)
		},
		Commands: []*cli.Command{
			{
				Name:  "keymap",
				Usage: "print or restore the default keymap",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "path to config dir (default: ~/.config/bterm)",
					},
					&cli.BoolFlag{
						Name:    "write",
						Aliases: []string{"w"},
						Usage:   "overwrite keymap.toml in the config dir with the default keymap",
					},
				},
				Action: keymapCommand,
			},
		},
	}

	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "bterm: %v\n", err)

		os.Exit(1)
	}
}

func run(ctx context.Context, cmd *cli.Command, command []string) error {
	cfgDir, err := config.ResolveDir(cmd.String("config"))
	if err != nil {
		return fmt.Errorf("resolve config dir: %w", err)
	}

	bundle, err := config.Load(cfgDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return cli.Exit("", ui.Run(ctx, bundle, cmd.String("working-directory"), command))
}

func keymapCommand(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool("write") {
		cfgDir, err := config.ResolveDir(cmd.String("config"))
		if err != nil {
			return fmt.Errorf("resolve config dir: %w", err)
		}

		dst := filepath.Join(cfgDir, "keymap.toml")

		content := keymap.KeymapHeader + keymap.DefaultKeymapTOML

		if err := os.WriteFile(dst, []byte(content), 0o644); err != nil { //nolint:gosec
			return fmt.Errorf("write keymap: %w", err)
		}

		fmt.Fprintf(os.Stdout, "keymap written to %s\n", dst)

		return nil
	}

	fmt.Fprint(os.Stdout, keymap.DefaultKeymapTOML)

	return nil
}

// splitExec separates "-e COMMAND [ARGS...]" from the bterm arguments. Like xterm's -e
// it must come last: everything after it is passed to the command untouched, so the
// command's own flags never reach the bterm flag parser.
func splitExec(args []string) ([]string, []string) {
	for i, a := range args {
		if a == "-e" || a == "--execute" {
			return args[:i], args[i+1:]
		}
	}

	return args, nil
}
