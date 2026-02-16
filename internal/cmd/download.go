package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/breca/vsixdler/internal/config"
	"github.com/breca/vsixdler/internal/marketplace"
)

func newDownloadCmd() *cobra.Command {
	var (
		cfgPath     string
		outDir      string
		concurrency int
		dryRun      bool
		verbose     bool
	)

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download VSIX extensions from the marketplace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDownload(cfgPath, outDir, concurrency, dryRun, verbose)
		},
	}

	cmd.Flags().StringVarP(&cfgPath, "config", "c", "extensions.yaml", "Config file path")
	cmd.Flags().StringVarP(&outDir, "output", "o", "./vsix", "Output directory")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "j", 4, "Max parallel downloads")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print plan without downloading")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging")

	return cmd
}

func runDownload(cfgPath, outDir string, concurrency int, dryRun, verbose bool) error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	log.Printf("loaded %d extension(s) from %s", len(cfg.Extensions), cfgPath)

	client := marketplace.NewClient(verbose)

	targets, err := marketplace.ResolveTargets(client, cfg.Extensions, verbose)
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		log.Println("no download targets resolved")
		return nil
	}

	fmt.Println()
	fmt.Printf("Download plan (%d file(s)):\n", len(targets))
	for _, t := range targets {
		fmt.Printf("  %s\n", t.Filename())
	}
	fmt.Println()

	if dryRun {
		log.Println("dry run — skipping downloads")
		return nil
	}

	return marketplace.DownloadAll(context.Background(), targets, outDir, concurrency, verbose)
}
