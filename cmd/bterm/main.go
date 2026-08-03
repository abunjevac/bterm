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
	cmd := &cli.Command{
		Name:    "bterm",
		Version: version.Version,
		Usage:   "an opinionated GTK4 + VTE terminal emulator",
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
		Action: run,
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

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "bterm: %v\n", err)

		os.Exit(1)
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	cfgDir, err := config.ResolveDir(cmd.String("config"))
	if err != nil {
		return fmt.Errorf("resolve config dir: %w", err)
	}

	bundle, err := config.Load(cfgDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ui.Run(ctx, bundle, cmd.String("working-directory"))

	return nil
}

func keymapCommand(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool("write") {
		cfgDir, err := config.ResolveDir(cmd.String("config"))
		if err != nil {
			return fmt.Errorf("resolve config dir: %w", err)
		}

		dst := filepath.Join(cfgDir, "keymap.toml")

		content := `# bterm keymap — action = "binding" (or = ["b1","b2"])` + "\n" + keymap.DefaultKeymapTOML

		if err := os.WriteFile(dst, []byte(content), 0o644); err != nil { //nolint:gosec
			return fmt.Errorf("write keymap: %w", err)
		}

		fmt.Fprintf(os.Stdout, "keymap written to %s\n", dst)

		return nil
	}

	fmt.Fprint(os.Stdout, keymap.DefaultKeymapTOML)

	return nil
}
