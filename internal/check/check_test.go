package check

import (
	"testing"

	"github.com/GBerghoff/envdiff/internal/config"
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

func TestCheck_RuntimeConstraints(t *testing.T) {
	snap := &snapshot.Snapshot{
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go":   {Version: "1.22.0", Path: "/usr/local/go/bin/go"},
			"node": {Version: "18.17.0", Path: "/usr/bin/node"},
		},
	}

	cfg := &config.Config{
		Runtime: map[string]string{
			"go":     ">= 1.21.0",
			"node":   ">= 20.0.0", // will fail
			"python": "*",        // missing, will fail
		},
		Fix: map[string]config.FixConfig{},
	}

	report := Check(snap, cfg)

	if report.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", report.Passed)
	}
	if report.Failed != 2 {
		t.Errorf("expected 2 failed, got %d", report.Failed)
	}
}

func TestCheck_EnvRequired(t *testing.T) {
	snap := &snapshot.Snapshot{
		Env: map[string]string{
			"HOME":     "/home/user",
			"NODE_ENV": "development",
		},
	}

	cfg := &config.Config{
		Runtime: map[string]string{},
		Env: config.EnvConfig{
			Required: []string{"HOME", "MISSING_VAR"},
		},
		Fix: map[string]config.FixConfig{},
	}

	report := Check(snap, cfg)

	if report.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", report.Passed)
	}
	if report.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", report.Failed)
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.22", "1.22.0"},
		{"v1.22.0", "1.22.0"},
		{"1", "1.0.0"},
		{"1.2.3", "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeVersion(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	patterns := []string{"HOME", "LC_*", "*_SESSION*"}

	tests := []struct {
		name     string
		expected bool
	}{
		{"HOME", true},
		{"LC_ALL", true},
		{"LC_CTYPE", true},
		{"SSH_SESSION_ID", true},
		{"PATH", false},
		{"NODE_ENV", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldIgnore(tt.name, patterns)
			if got != tt.expected {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}
