package render

import (
	"strings"
	"testing"

	"github.com/GBerghoff/envdiff/internal/diff"
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

func TestCLIRenderer_RenderSnapshot(t *testing.T) {
	snap := &snapshot.Snapshot{
		Hostname: "test-host",
		System: snapshot.SystemInfo{
			OS:        "linux",
			OSVersion: "Ubuntu 22.04",
			Arch:      "amd64",
			Kernel:    "5.15.0",
			CPUCores:  8,
			MemoryGB:  16,
		},
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go":   {Version: "1.22.0", Path: "/usr/bin/go"},
			"node": {Version: "20.0.0", Path: "/usr/bin/node"},
		},
		Env: map[string]string{
			"NODE_ENV": "development",
			"API_KEY":  "[REDACTED]",
		},
	}

	renderer := NewCLI()
	output := renderer.RenderSnapshot(snap)

	if !strings.Contains(output, "test-host") {
		t.Error("Output should contain hostname")
	}
	if !strings.Contains(output, "SYSTEM") {
		t.Error("Output should contain SYSTEM section")
	}
	if !strings.Contains(output, "RUNTIME") {
		t.Error("Output should contain RUNTIME section")
	}
	if !strings.Contains(output, "ENVIRONMENT") {
		t.Error("Output should contain ENVIRONMENT section")
	}
	if !strings.Contains(output, "1.22.0") {
		t.Error("Output should contain go version")
	}
}

func TestCLIRenderer_RenderDiff(t *testing.T) {
	diffResult := &diff.Diff{
		Nodes: []string{"local", "ci"},
		Summary: diff.Summary{
			TotalNodes: 2,
			Different:  2,
			Equal:      4,
			Redacted:   1,
		},
		Diffs: map[string]map[string]*diff.FieldDiff{
			"runtime": {
				"go": {
					Status: diff.StatusDifferent,
					NodeValues: map[string]any{
						"local": "1.22.0",
						"ci":    "1.21.0",
					},
				},
			},
			"env": {
				"NODE_ENV": {
					Status: diff.StatusDifferent,
					NodeValues: map[string]any{
						"local": "development",
						"ci":    "production",
					},
				},
				"API_KEY": {
					Status: diff.StatusRedacted,
					NodeValues: map[string]any{
						"local": "[REDACTED]",
						"ci":    "[REDACTED]",
					},
				},
			},
			"system": {},
		},
		Errors:    map[string]string{},
		Snapshots: map[string]*snapshot.Snapshot{},
	}

	renderer := NewCLI()
	output := renderer.RenderDiff(diffResult)

	if !strings.Contains(output, "local") {
		t.Error("Output should contain node name 'local'")
	}
	if !strings.Contains(output, "ci") {
		t.Error("Output should contain node name 'ci'")
	}
	if !strings.Contains(output, "RUNTIME") {
		t.Error("Output should contain RUNTIME section")
	}
	if !strings.Contains(output, "2 different") {
		t.Error("Output should contain difference count")
	}
}

func TestMarkdownRenderer_RenderSnapshot(t *testing.T) {
	snap := &snapshot.Snapshot{
		Hostname:     "test-host",
		Timestamp:    "2024-01-01T00:00:00Z",
		CollectedVia: "local",
		System: snapshot.SystemInfo{
			OSVersion: "Ubuntu 22.04",
			Arch:      "amd64",
			Kernel:    "5.15.0",
			CPUCores:  8,
			MemoryGB:  16,
		},
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go": {Version: "1.22.0", Path: "/usr/bin/go"},
		},
		Env: map[string]string{
			"NODE_ENV": "development",
		},
	}

	renderer := NewMarkdown()
	output := renderer.RenderSnapshot(snap)

	if !strings.Contains(output, "# Environment Snapshot") {
		t.Error("Output should contain markdown title")
	}
	if !strings.Contains(output, "## System") {
		t.Error("Output should contain System section")
	}
	if !strings.Contains(output, "## Runtime") {
		t.Error("Output should contain Runtime section")
	}
	if !strings.Contains(output, "|") {
		t.Error("Output should contain markdown table")
	}
	if !strings.Contains(output, "test-host") {
		t.Error("Output should contain hostname")
	}
}

func TestMarkdownRenderer_RenderDiff(t *testing.T) {
	diffResult := &diff.Diff{
		GeneratedAt: "2024-01-01T00:00:00Z",
		Nodes:       []string{"local", "ci"},
		Summary: diff.Summary{
			TotalNodes: 2,
			Different:  1,
			Equal:      5,
		},
		Diffs: map[string]map[string]*diff.FieldDiff{
			"runtime": {
				"go": {
					Status: diff.StatusDifferent,
					NodeValues: map[string]any{
						"local": "1.22.0",
						"ci":    "1.21.0",
					},
				},
			},
			"env":    {},
			"system": {},
		},
		Errors:    map[string]string{},
		Snapshots: map[string]*snapshot.Snapshot{},
	}

	renderer := NewMarkdown()
	output := renderer.RenderDiff(diffResult)

	if !strings.Contains(output, "# Environment Diff") {
		t.Error("Output should contain markdown title")
	}
	if !strings.Contains(output, "## Summary") {
		t.Error("Output should contain Summary section")
	}
	if !strings.Contains(output, "## Runtime") {
		t.Error("Output should contain Runtime section when there are differences")
	}
	if !strings.Contains(output, "local") {
		t.Error("Output should contain node names")
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{nil, "(missing)"},
		{"hello", "hello"},
		{123, "123"},
		{"", ""},
	}

	for _, test := range tests {
		result := formatValue(test.input)
		if result != test.expected {
			t.Errorf("formatValue(%v) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestFormatMarkdownValue(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{nil, "—"},
		{"hello", "hello"},
		{"[REDACTED]", "🔒"},
	}

	for _, test := range tests {
		result := formatMarkdownValue(test.input)
		if result != test.expected {
			t.Errorf("formatMarkdownValue(%v) = %q, want %q", test.input, result, test.expected)
		}
	}
}
