// Package check validates snapshots against declared requirements.
// Uses semver constraints for runtime versions and supports glob patterns for env filtering.
package check

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/gberghoff/envdiff/internal/config"
	"github.com/gberghoff/envdiff/internal/snapshot"
)

// Result represents a single check result
type Result struct {
	Category string `json:"category"` // "runtime" or "env"
	Name     string `json:"name"`
	Status   string `json:"status"`  // "pass", "fail", "warn"
	Message  string `json:"message"`
	Expected string `json:"expected,omitempty"`
	Actual   string `json:"actual,omitempty"`
	FixHint  string `json:"fix_hint,omitempty"`
}

// Report contains all check results
type Report struct {
	Results []Result `json:"results"`
	Passed  int      `json:"passed"`
	Failed  int      `json:"failed"`
	Warned  int      `json:"warned"`
}

// Check validates a snapshot against a configuration
func Check(snap *snapshot.Snapshot, cfg *config.Config) *Report {
	report := &Report{
		Results: []Result{},
	}

	// Check runtime versions
	for name, constraint := range cfg.Runtime {
		result := checkRuntime(snap, name, constraint, cfg.Fix[name])
		report.Results = append(report.Results, result)
		updateCounts(report, result.Status)
	}

	// Check required environment variables
	for _, name := range cfg.Env.Required {
		result := checkEnvRequired(snap, name, cfg.Fix[name])
		report.Results = append(report.Results, result)
		updateCounts(report, result.Status)
	}

	// Check expected environment variable values
	for name, expected := range cfg.Env.Expected {
		result := checkEnvExpected(snap, name, expected, cfg.Fix[name])
		report.Results = append(report.Results, result)
		updateCounts(report, result.Status)
	}

	return report
}

func checkRuntime(snap *snapshot.Snapshot, name, constraint string, fix config.FixConfig) Result {
	result := Result{
		Category: "runtime",
		Name:     name,
		Expected: constraint,
	}

	info, exists := snap.Runtime[name]
	if !exists || info == nil {
		result.Status = "fail"
		result.Message = "not installed"
		result.Actual = "(missing)"
		if fix.Missing != "" {
			result.FixHint = fix.Missing
		}
		return result
	}

	result.Actual = info.Version

	// Handle wildcard - any version is fine
	if constraint == "*" {
		result.Status = "pass"
		result.Message = "installed"
		return result
	}

	// Parse and check version constraint
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		result.Status = "warn"
		result.Message = fmt.Sprintf("invalid constraint: %s", constraint)
		return result
	}

	// Normalize version string (handle missing patch version)
	version := normalizeVersion(info.Version)
	v, err := semver.NewVersion(version)
	if err != nil {
		result.Status = "warn"
		result.Message = fmt.Sprintf("cannot parse version: %s", info.Version)
		return result
	}

	if c.Check(v) {
		result.Status = "pass"
		result.Message = fmt.Sprintf("satisfies %s", constraint)
	} else {
		result.Status = "fail"
		result.Message = fmt.Sprintf("does not satisfy %s", constraint)
		if fix.WrongVersion != "" {
			result.FixHint = fix.WrongVersion
		}
	}

	return result
}

func checkEnvRequired(snap *snapshot.Snapshot, name string, fix config.FixConfig) Result {
	result := Result{
		Category: "env",
		Name:     name,
		Expected: "(set)",
	}

	val, exists := snap.Env[name]
	if !exists || val == "" {
		result.Status = "fail"
		result.Message = "not set"
		result.Actual = "(missing)"
		if fix.Missing != "" {
			result.FixHint = fix.Missing
		}
	} else {
		result.Status = "pass"
		result.Message = "set"
		result.Actual = "(set)"
	}

	return result
}

func checkEnvExpected(snap *snapshot.Snapshot, name, expected string, fix config.FixConfig) Result {
	result := Result{
		Category: "env",
		Name:     name,
		Expected: expected,
	}

	val, exists := snap.Env[name]
	if !exists {
		result.Status = "fail"
		result.Message = "not set"
		result.Actual = "(missing)"
		if fix.Missing != "" {
			result.FixHint = fix.Missing
		}
		return result
	}

	result.Actual = val
	if val == expected {
		result.Status = "pass"
		result.Message = "matches"
	} else {
		result.Status = "fail"
		result.Message = fmt.Sprintf("expected %s", expected)
		if fix.WrongVersion != "" {
			result.FixHint = fix.WrongVersion
		}
	}

	return result
}

func updateCounts(report *Report, status string) {
	switch status {
	case "pass":
		report.Passed++
	case "fail":
		report.Failed++
	case "warn":
		report.Warned++
	}
}

func normalizeVersion(v string) string {
	// Remove leading 'v' if present
	v = strings.TrimPrefix(v, "v")

	// Add .0 suffix if only major.minor
	parts := strings.Split(v, ".")
	if len(parts) == 2 {
		v = v + ".0"
	} else if len(parts) == 1 {
		v = v + ".0.0"
	}

	return v
}

// ShouldIgnore checks if an environment variable should be ignored
func ShouldIgnore(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesPattern(name, pattern) {
			return true
		}
	}
	return false
}

func matchesPattern(name, pattern string) bool {
	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			return false // Invalid pattern syntax
		}
		return matched
	}
	return name == pattern
}
