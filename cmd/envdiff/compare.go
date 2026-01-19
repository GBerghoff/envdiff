package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GBerghoff/envdiff/internal/diff"
	"github.com/GBerghoff/envdiff/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	compareOutput string
)

var compareCmd = &cobra.Command{
	Use:   "compare <snapshot1.json> <snapshot2.json> [snapshot3.json...]",
	Short: "Compare two or more snapshots",
	Long: `Compare environment snapshots and produce a diff.

Examples:
  envdiff compare local.json ci.json              # Compare two snapshots
  envdiff compare local.json ci.json staging.json # Compare multiple
  envdiff compare local.json ci.json -o diff.json # Save diff to file`,
	Args: cobra.MinimumNArgs(2),
	RunE: runCompare,
}

func init() {
	compareCmd.Flags().StringVarP(&compareOutput, "output", "o", "", "Output file (default: stdout)")
}

func runCompare(cmd *cobra.Command, args []string) error {
	snapshots := make(map[string]*snapshot.Snapshot)

	// Load all snapshots
	for _, path := range args {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		snap, err := snapshot.FromJSON(data)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Use filename (without extension) as node name
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		snapshots[name] = snap
	}

	// Compare
	d := diff.Compare(snapshots)

	// Output
	output, err := d.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to format diff: %w", err)
	}

	if compareOutput != "" {
		if err := os.WriteFile(compareOutput, output, 0600); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Diff saved to %s\n", compareOutput)
	} else {
		fmt.Print(string(output))
	}

	return nil
}
