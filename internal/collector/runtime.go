package collector

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// runtimeDef defines how to detect a runtime/CLI tool
type runtimeDef struct {
	Name       string
	Command    string
	Args       []string
	VersionRE  *regexp.Regexp
}

// Common runtimes and CLI tools to detect
var runtimes = []runtimeDef{
	{
		Name:      "go",
		Command:   "go",
		Args:      []string{"version"},
		VersionRE: regexp.MustCompile(`go(\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "node",
		Command:   "node",
		Args:      []string{"-v"},
		VersionRE: regexp.MustCompile(`v?(\d+\.\d+\.\d+)`),
	},
	{
		Name:      "python",
		Command:   "python3",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`Python (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "ruby",
		Command:   "ruby",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`ruby (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "rust",
		Command:   "rustc",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`rustc (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "java",
		Command:   "java",
		Args:      []string{"-version"},
		VersionRE: regexp.MustCompile(`version "(\d+\.?\d*\.?\d*)"`),
	},
	{
		Name:      "docker",
		Command:   "docker",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`Docker version (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "kubectl",
		Command:   "kubectl",
		Args:      []string{"version", "--client", "--short"},
		VersionRE: regexp.MustCompile(`v?(\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "aws",
		Command:   "aws",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`aws-cli/(\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "gcloud",
		Command:   "gcloud",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`Google Cloud SDK (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "terraform",
		Command:   "terraform",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`Terraform v(\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "git",
		Command:   "git",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`git version (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "npm",
		Command:   "npm",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`(\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "yarn",
		Command:   "yarn",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`(\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "pnpm",
		Command:   "pnpm",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`(\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "pip",
		Command:   "pip3",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`pip (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "cargo",
		Command:   "cargo",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`cargo (\d+\.\d+\.?\d*)`),
	},
	{
		Name:      "make",
		Command:   "make",
		Args:      []string{"--version"},
		VersionRE: regexp.MustCompile(`GNU Make (\d+\.\d+\.?\d*)`),
	},
}

// RuntimeCollector detects installed runtimes and CLI tools
type RuntimeCollector struct{}

// Collect gathers runtime information
func (c *RuntimeCollector) Collect(s *snapshot.Snapshot) error {
	for _, rt := range runtimes {
		info := c.detectRuntime(rt)
		if info != nil {
			s.Runtime[rt.Name] = info
		}
	}
	return nil
}

func (c *RuntimeCollector) detectRuntime(rt runtimeDef) *snapshot.RuntimeInfo {
	path, err := exec.LookPath(rt.Command)
	if err != nil {
		return nil // Not installed
	}

	cmd := exec.Command(rt.Command, rt.Args...)
	// Capture both stdout and stderr (java outputs to stderr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Command exists but failed to get version
		return &snapshot.RuntimeInfo{
			Version: "unknown",
			Path:    path,
		}
	}

	version := c.extractVersion(string(out), rt.VersionRE)
	if version == "" {
		version = "unknown"
	}

	return &snapshot.RuntimeInfo{
		Version: version,
		Path:    path,
	}
}

func (c *RuntimeCollector) extractVersion(output string, re *regexp.Regexp) string {
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
