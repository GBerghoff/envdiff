package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Runtime == nil {
		t.Error("Runtime map should be initialized")
	}
	if _, hasGo := cfg.Runtime["go"]; !hasGo {
		t.Error("Default config should include go runtime constraint")
	}
	if len(cfg.Env.Ignore) == 0 {
		t.Error("Default config should have ignore patterns")
	}
	if cfg.Fix == nil {
		t.Error("Fix map should be initialized")
	}
}

func TestDefaultConfig_RuntimeConstraints(t *testing.T) {
	cfg := DefaultConfig()

	expectedRuntimes := []string{"go", "node", "python", "docker"}
	for _, runtime := range expectedRuntimes {
		if _, exists := cfg.Runtime[runtime]; !exists {
			t.Errorf("Default config should include %s runtime", runtime)
		}
	}
}

func TestToYAML(t *testing.T) {
	cfg := &Config{
		Runtime: map[string]string{
			"go": ">= 1.21.0",
		},
		Env: EnvConfig{
			Required: []string{"DATABASE_URL"},
		},
	}

	data, err := cfg.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("ToYAML() should return non-empty data")
	}

	yamlStr := string(data)
	if !strings.Contains(yamlStr, "go:") {
		t.Error("YAML should contain go runtime")
	}
	if !strings.Contains(yamlStr, "DATABASE_URL") {
		t.Error("YAML should contain required env var")
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "envdiff.yaml")

	yamlContent := `
runtime:
  go: ">= 1.21.0"
  node: ">= 18.0.0"
env:
  required:
    - DATABASE_URL
  expected:
    NODE_ENV: production
  ignore:
    - HOME
    - PATH
fix:
  go:
    missing: "brew install go"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Runtime["go"] != ">= 1.21.0" {
		t.Errorf("Runtime[go] = %q, want %q", cfg.Runtime["go"], ">= 1.21.0")
	}
	if len(cfg.Env.Required) != 1 || cfg.Env.Required[0] != "DATABASE_URL" {
		t.Errorf("Env.Required = %v, want [DATABASE_URL]", cfg.Env.Required)
	}
	if cfg.Env.Expected["NODE_ENV"] != "production" {
		t.Errorf("Env.Expected[NODE_ENV] = %q, want %q", cfg.Env.Expected["NODE_ENV"], "production")
	}
	if cfg.Fix["go"].Missing != "brew install go" {
		t.Errorf("Fix[go].Missing = %q, want %q", cfg.Fix["go"].Missing, "brew install go")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() should return error for missing file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	if err := os.WriteFile(configPath, []byte("invalid: [yaml: content"), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestTemplate(t *testing.T) {
	template := Template()

	if len(template) == 0 {
		t.Error("Template() should return non-empty string")
	}
	if !strings.Contains(template, "runtime:") {
		t.Error("Template should contain runtime section")
	}
	if !strings.Contains(template, "env:") {
		t.Error("Template should contain env section")
	}
}

