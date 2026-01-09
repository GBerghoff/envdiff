package main

import (
	"fmt"
	"os"

	"github.com/GBerghoff/envdiff/internal/config"
	"github.com/spf13/cobra"
)

var (
	initForce bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create an envdiff.yaml template",
	Long: `Create a template envdiff.yaml file in the current directory.

The generated file includes common runtime requirements and
environment variable patterns. Customize it for your project.

Examples:
  envdiff init              # Create ./envdiff.yaml
  envdiff init --force      # Overwrite existing file`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Overwrite existing file")
}

func runInit(cmd *cobra.Command, args []string) error {
	filename := "envdiff.yaml"

	// Check if file exists
	if _, err := os.Stat(filename); err == nil && !initForce {
		return fmt.Errorf("%s already exists. Use --force to overwrite", filename)
	}

	// Write template
	template := config.Template()
	if err := os.WriteFile(filename, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	fmt.Printf("Created %s\n", filename)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit envdiff.yaml to match your project requirements")
	fmt.Println("  2. Run 'envdiff check' to validate your environment")
	fmt.Println("  3. Commit envdiff.yaml to your repository")

	return nil
}
