package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gberghoff/envdiff/internal/diff"
	"github.com/gberghoff/envdiff/internal/snapshot"
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
		if err := os.WriteFile(compareOutput, output, 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Diff saved to %s\n", compareOutput)
	} else {
		fmt.Print(string(output))
	}

	return nil
}

// Helper to load snapshots from a file that might contain an array
func loadSnapshots(path string) (map[string]*snapshot.Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try to parse as array first
	var snapArray []*snapshot.Snapshot
	if err := json.Unmarshal(data, &snapArray); err == nil {
		result := make(map[string]*snapshot.Snapshot)
		for _, s := range snapArray {
			result[s.Hostname] = s
		}
		return result, nil
	}

	// Try as single snapshot
	var snap snapshot.Snapshot
	if err := json.Unmarshal(data, &snap); err == nil {
		return map[string]*snapshot.Snapshot{
			snap.Hostname: &snap,
		}, nil
	}

	return nil, fmt.Errorf("could not parse snapshot file")
}
