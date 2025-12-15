package main

import (
	"fmt"
	"os"

	"github.com/gberghoff/envdiff/internal/collector"
	"github.com/gberghoff/envdiff/internal/render"
	"github.com/gberghoff/envdiff/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	snapshotNoRedact bool
	snapshotOutput   string
	snapshotFormat   string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Capture an environment snapshot",
	Long: `Capture a snapshot of the current environment.

The snapshot includes:
  • System info (OS, arch, kernel, memory, CPU)
  • Runtime versions (go, node, python, docker, etc.)
  • Environment variables (secrets auto-redacted)
  • Network info (/etc/hosts, listening ports)

Examples:
  envdiff snapshot                    # Output JSON to stdout
  envdiff snapshot -o local.json      # Save to file
  envdiff snapshot --no-redact        # Include secret values
  envdiff snapshot --format cli       # Pretty terminal output`,
	RunE: runSnapshot,
}

func init() {
	snapshotCmd.Flags().BoolVar(&snapshotNoRedact, "no-redact", false, "Include actual secret values (use with caution)")
	snapshotCmd.Flags().StringVarP(&snapshotOutput, "output", "o", "", "Output file (default: stdout)")
	snapshotCmd.Flags().StringVar(&snapshotFormat, "format", "json", "Output format: json, cli, md")
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	// Create snapshot
	snap := snapshot.New()

	// Run collectors
	redact := !snapshotNoRedact
	if err := collector.CollectAll(snap, redact); err != nil {
		return fmt.Errorf("failed to collect environment: %w", err)
	}

	// Compute snapshot ID
	if err := snap.ComputeID(); err != nil {
		return fmt.Errorf("failed to compute snapshot ID: %w", err)
	}

	// Format output
	var output []byte
	var err error

	switch snapshotFormat {
	case "json":
		output, err = snap.ToJSON()
	case "cli":
		renderer := render.NewCLI()
		output = []byte(renderer.RenderSnapshot(snap))
	case "md", "markdown":
		renderer := render.NewMarkdown()
		output = []byte(renderer.RenderSnapshot(snap))
	default:
		return fmt.Errorf("unknown format: %s (use json, cli, or md)", snapshotFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Write output
	if snapshotOutput != "" {
		if err := os.WriteFile(snapshotOutput, output, 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Snapshot saved to %s\n", snapshotOutput)
	} else {
		fmt.Print(string(output))
	}

	return nil
}
