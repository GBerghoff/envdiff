package collector

import (
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// RuntimeDefinition defines how to detect a runtime/CLI tool
type RuntimeDefinition struct {
	Name       string
	Command    string
	Args       []string
	VersionRE  *regexp.Regexp
}

// Common runtimes and CLI tools to detect
var runtimes = []RuntimeDefinition{
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
	{
		Name:      "svn",
		Command:   "svn",
		Args:      []string{"--version", "--quiet"},
		VersionRE: regexp.MustCompile(`(\d+\.\d+\.?\d*)`),
	},
}

// RuntimeCollector detects installed runtimes and CLI tools
type RuntimeCollector struct{}

// Collect gathers runtime information using parallel detection
func (c *RuntimeCollector) Collect(snap *snapshot.Snapshot) error {
	var waitGroup sync.WaitGroup
	var mutex sync.Mutex

	for _, runtime := range runtimes {
		waitGroup.Add(1)
		go func(runtime RuntimeDefinition) {
			defer waitGroup.Done()
			if info := c.detectRuntime(runtime); info != nil {
				mutex.Lock()
				snap.Runtime[runtime.Name] = info
				mutex.Unlock()
			}
		}(runtime)
	}
	waitGroup.Wait()
	return nil
}

func (c *RuntimeCollector) detectRuntime(runtime RuntimeDefinition) *snapshot.RuntimeInfo {
	path, err := exec.LookPath(runtime.Command)
	if err != nil {
		return nil // Not installed
	}

	cmd := exec.Command(runtime.Command, runtime.Args...)
	// Capture both stdout and stderr (java outputs to stderr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Command exists but failed to get version
		return &snapshot.RuntimeInfo{
			Version: "unknown",
			Path:    path,
		}
	}

	version := c.extractVersion(string(out), runtime.VersionRE)
	if version == "" {
		version = "unknown"
	}

	return &snapshot.RuntimeInfo{
		Version: version,
		Path:    path,
	}
}

func (c *RuntimeCollector) extractVersion(output string, pattern *regexp.Regexp) string {
	matches := pattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
