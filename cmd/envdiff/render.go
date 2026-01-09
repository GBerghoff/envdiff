package main

import (
	"fmt"
	"os"

	"github.com/GBerghoff/envdiff/internal/diff"
	"github.com/GBerghoff/envdiff/internal/render"
	"github.com/GBerghoff/envdiff/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	renderMarkdown bool
	renderOutput   string
)

var renderCmd = &cobra.Command{
	Use:   "render <file.json>",
	Short: "Render a snapshot or diff as CLI or Markdown output",
	Long: `Render a JSON snapshot or diff for human consumption.

Examples:
  envdiff render snapshot.json           # CLI output (default)
  envdiff render diff.json --md          # Markdown output
  envdiff render diff.json --md -o report.md`,
	Args: cobra.ExactArgs(1),
	RunE: runRender,
}

func init() {
	renderCmd.Flags().BoolVar(&renderMarkdown, "md", false, "Output as Markdown")
	renderCmd.Flags().StringVarP(&renderOutput, "output", "o", "", "Output file (default: stdout)")
}

func runRender(cmd *cobra.Command, args []string) error {
	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var output string

	// Try to detect if it's a diff or snapshot
	if isDiff(data) {
		d, err := diff.FromJSON(data)
		if err != nil {
			return fmt.Errorf("failed to parse diff: %w", err)
		}
		if renderMarkdown {
			renderer := render.NewMarkdown()
			output = renderer.RenderDiff(d)
		} else {
			renderer := render.NewCLI()
			output = renderer.RenderDiff(d)
		}
	} else {
		snap, err := snapshot.FromJSON(data)
		if err != nil {
			return fmt.Errorf("failed to parse snapshot: %w", err)
		}
		if renderMarkdown {
			renderer := render.NewMarkdown()
			output = renderer.RenderSnapshot(snap)
		} else {
			renderer := render.NewCLI()
			output = renderer.RenderSnapshot(snap)
		}
	}

	if renderOutput != "" {
		if err := os.WriteFile(renderOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Output saved to %s\n", renderOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

// isDiff checks if the JSON is a diff (has "diffs" key) or snapshot
func isDiff(data []byte) bool {
	// Simple heuristic: diffs have "diffs" and "nodes" keys
	d, err := diff.FromJSON(data)
	if err != nil {
		return false
	}
	return len(d.Nodes) > 0
}
