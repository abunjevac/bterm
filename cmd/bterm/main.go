package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/abunjevac/bterm/internal/config"
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
		},
		Action: run,
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

	ui.Run(ctx, bundle)

	return nil
}
