package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/liblaf/route-rules/internal/build"
	"github.com/liblaf/route-rules/internal/output"
)

func main() {
	log.SetFlags(0)
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("route-rules", flag.ContinueOnError)
	configPath := flags.String("config", "config/rulesets.json", "configuration file")
	outputDir := flags.String("output", "dist", "output directory")
	mihomoPath := flags.String("mihomo", "mihomo", "mihomo executable")
	generatedAtText := flags.String("generated-at", "", "RFC3339 build timestamp")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if *configPath == "" || *outputDir == "" || *mihomoPath == "" {
		return fmt.Errorf("config path, output directory, and Mihomo path are required")
	}

	generatedAt := time.Now().UTC().Truncate(time.Second)
	if *generatedAtText != "" {
		parsed, err := time.Parse(time.RFC3339, *generatedAtText)
		if err != nil {
			return fmt.Errorf("parse --generated-at: %w", err)
		}
		generatedAt = parsed.UTC()
	}

	cfg, err := build.LoadConfig(*configPath)
	if err != nil {
		return err
	}
	built, err := build.Run(ctx, cfg)
	if err != nil {
		return err
	}
	manifest, err := output.Write(ctx, *outputDir, cfg, built, *mihomoPath, generatedAt)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "generated %d artifacts in %s at %s\n", len(manifest.Artifacts), *outputDir, manifest.GeneratedAt.Format(time.RFC3339))
	return nil
}
