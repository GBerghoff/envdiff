package main

import (
	"fmt"
	"os"

	"github.com/gberghoff/envdiff/internal/check"
	"github.com/gberghoff/envdiff/internal/collector"
	"github.com/gberghoff/envdiff/internal/config"
	"github.com/gberghoff/envdiff/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	checkFile   string
	checkJSON   bool
	checkQuiet  bool
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate local environment against requirements",
	Long: `Check that your local environment meets the requirements defined in envdiff.yaml.

Examples:
  envdiff check                    # Use ./envdiff.yaml
  envdiff check --file staging.yaml
  envdiff check --json             # Output as JSON
  envdiff check --quiet            # Exit code only (for git hooks)`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&checkFile, "file", "f", "envdiff.yaml", "Path to requirements file")
	checkCmd.Flags().BoolVar(&checkJSON, "json", false, "Output as JSON")
	checkCmd.Flags().BoolVarP(&checkQuiet, "quiet", "q", false, "Suppress output, exit code only")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(checkFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s found. Run 'envdiff init' to create one", checkFile)
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Take a snapshot (no redaction needed for check)
	snap := snapshot.New()
	if err := collector.CollectAll(snap, false); err != nil {
		return fmt.Errorf("failed to collect environment: %w", err)
	}

	// Run checks
	report := check.Check(snap, cfg)

	// Output
	if checkQuiet {
		// Silent mode - just return exit code
		if report.Failed > 0 {
			os.Exit(1)
		}
		return nil
	}

	if checkJSON {
		output, err := report.RenderJSON()
		if err != nil {
			return fmt.Errorf("failed to render JSON: %w", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Print(report.RenderCLI())
	}

	if report.Failed > 0 {
		os.Exit(1)
	}

	return nil
}
