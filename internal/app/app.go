package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/liblaf/route-rules/internal/build"
	"github.com/liblaf/route-rules/internal/config"
	"github.com/liblaf/route-rules/internal/mihomo"
	"github.com/liblaf/route-rules/internal/output"
)

type Options struct {
	ConfigPath  string
	OutputDir   string
	MihomoPath  string
	GeneratedAt time.Time
}

func Run(ctx context.Context, options Options) error {
	if options.ConfigPath == "" || options.OutputDir == "" || options.MihomoPath == "" || options.GeneratedAt.IsZero() {
		return fmt.Errorf("config path, output directory, Mihomo path, and generation time are required")
	}
	cfg, err := config.Load(options.ConfigPath)
	if err != nil {
		return err
	}
	built, err := build.Run(ctx, cfg)
	if err != nil {
		return err
	}
	manifest, err := output.Write(ctx, options.OutputDir, cfg, built, mihomo.New(options.MihomoPath), options.GeneratedAt)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "generated %d artifacts in %s at %s\n", len(manifest.Artifacts), options.OutputDir, manifest.GeneratedAt.Format(time.RFC3339))
	return nil
}
