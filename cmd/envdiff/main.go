package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "envdiff",
	Short: "Compare environments and surface the differences that matter",
	Long: `envdiff captures environment snapshots and intelligently diffs them,
surfacing the 2-3 differences that actually matter—not drowning you in noise.

Primary use cases:
  • "Why does CI fail when it passes locally?"
  • "New hire onboarding" — envdiff check beats a 47-step setup doc
  • "What changed since it last worked?" — incident debugging`,
	Version: version,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(renderCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(initCmd)
}
