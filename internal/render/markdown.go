package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GBerghoff/envdiff/internal/diff"
	"github.com/GBerghoff/envdiff/internal/secrets"
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// MarkdownRenderer renders output as Markdown
type MarkdownRenderer struct{}

// NewMarkdown creates a new Markdown renderer
func NewMarkdown() *MarkdownRenderer {
	return &MarkdownRenderer{}
}

// RenderSnapshot renders a snapshot as Markdown
func (r *MarkdownRenderer) RenderSnapshot(s *snapshot.Snapshot) string {
	var b strings.Builder

	b.WriteString("# Environment Snapshot\n\n")
	fmt.Fprintf(&b, "**Host:** %s  \n", s.Hostname)
	fmt.Fprintf(&b, "**Timestamp:** %s  \n", s.Timestamp)
	fmt.Fprintf(&b, "**Collected via:** %s\n\n", s.CollectedVia)

	// System
	b.WriteString("## System\n\n")
	b.WriteString("| Field | Value |\n")
	b.WriteString("|-------|-------|\n")
	fmt.Fprintf(&b, "| OS | %s |\n", s.System.OSVersion)
	fmt.Fprintf(&b, "| Architecture | %s |\n", s.System.Arch)
	fmt.Fprintf(&b, "| Kernel | %s |\n", s.System.Kernel)
	fmt.Fprintf(&b, "| CPU Cores | %d |\n", s.System.CPUCores)
	fmt.Fprintf(&b, "| Memory | %d GB |\n", s.System.MemoryGB)
	b.WriteString("\n")

	// Runtime
	b.WriteString("## Runtime\n\n")
	b.WriteString("| Tool | Version | Path |\n")
	b.WriteString("|------|---------|------|\n")
	runtimes := sortedKeys(s.Runtime)
	for _, name := range runtimes {
		info := s.Runtime[name]
		if info != nil {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", name, info.Version, info.Path)
		}
	}
	b.WriteString("\n")

	// Environment (summary only)
	b.WriteString("## Environment\n\n")
	redactedCount := 0
	for _, v := range s.Env {
		if v == secrets.RedactedValue {
			redactedCount++
		}
	}
	fmt.Fprintf(&b, "%d variables (%d redacted)\n\n", len(s.Env), redactedCount)

	// Packages
	if s.Packages != nil && len(s.Packages.Items) > 0 {
		b.WriteString("## Packages (" + s.Packages.Manager + ")\n\n")
		b.WriteString("| Package | Version |\n")
		b.WriteString("|---------|---------|\n")
		pkgNames := make([]string, 0, len(s.Packages.Items))
		for name := range s.Packages.Items {
			pkgNames = append(pkgNames, name)
		}
		sort.Strings(pkgNames)

		for _, name := range pkgNames {
			version := s.Packages.Items[name]
			fmt.Fprintf(&b, "| %s | %s |\n", name, version)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderDiff renders a diff as Markdown
func (r *MarkdownRenderer) RenderDiff(d *diff.Diff) string {
	var b strings.Builder

	b.WriteString("# Environment Diff\n\n")
	fmt.Fprintf(&b, "**Generated:** %s  \n", d.GeneratedAt)
	fmt.Fprintf(&b, "**Nodes:** %d | **Different:** %d\n\n", len(d.Nodes), d.Summary.Different)

	// Summary table
	b.WriteString("## Summary\n\n")
	b.WriteString("| Node | Status | Issues |\n")
	b.WriteString("|------|--------|--------|\n")
	for _, node := range d.Nodes {
		if err, ok := d.Errors[node]; ok {
			fmt.Fprintf(&b, "| %s | ❌ error | %s |\n", node, err)
		} else {
			issues := r.getNodeIssues(d, node)
			if len(issues) == 0 {
				fmt.Fprintf(&b, "| %s | ✓ ok | — |\n", node)
			} else {
				fmt.Fprintf(&b, "| %s | ⚠ differs | %s |\n", node, strings.Join(issues, ", "))
			}
		}
	}
	b.WriteString("\n")

	// Runtime table
	if r.hasAnyDifferent(d.Diffs["runtime"]) {
		b.WriteString("## Runtime\n\n")
		b.WriteString(r.renderComparisonTable(d, "runtime"))
	}

	// Package table
	if r.hasAnyDifferent(d.Diffs["package"]) {
		b.WriteString("## Packages\n\n")
		b.WriteString(r.renderComparisonTable(d, "package"))
	}

	// Environment table (only different ones)
	if r.hasAnyDifferent(d.Diffs["env"]) {
		b.WriteString("## Environment\n\n")
		b.WriteString(r.renderComparisonTable(d, "env"))
	}

	// System table
	if r.hasAnyDifferent(d.Diffs["system"]) {
		b.WriteString("## System\n\n")
		b.WriteString(r.renderComparisonTable(d, "system"))
	}

	return b.String()
}

func (r *MarkdownRenderer) getNodeIssues(d *diff.Diff, node string) []string {
	var issues []string

	for section, fields := range d.Diffs {
		for name, fieldDiff := range fields {
			if fieldDiff.Status == diff.StatusDifferent {
				// Check if this node is an outlier
				for _, outlier := range fieldDiff.Outliers {
					if outlier == node {
						issues = append(issues, fmt.Sprintf("%s %s", section, name))
						break
					}
				}
				// For 2-node diffs, both are "different"
				if len(d.Nodes) == 2 && len(fieldDiff.Outliers) == 0 {
					// Only add once
					if node == d.Nodes[0] {
						issues = append(issues, name)
					}
				}
			}
		}
	}

	return issues
}

func (r *MarkdownRenderer) hasAnyDifferent(fields map[string]*diff.FieldDiff) bool {
	for _, fieldDiff := range fields {
		if fieldDiff.Status == diff.StatusDifferent || fieldDiff.Status == diff.StatusRedacted {
			return true
		}
	}
	return false
}

func (r *MarkdownRenderer) renderComparisonTable(d *diff.Diff, section string) string {
	var b strings.Builder

	fields := d.Diffs[section]
	if len(fields) == 0 {
		return ""
	}

	// Header row
	b.WriteString("| Field |")
	for _, node := range d.Nodes {
		fmt.Fprintf(&b, " %s |", node)
	}
	b.WriteString("\n")

	// Separator
	b.WriteString("|-------|")
	for range d.Nodes {
		b.WriteString("-------|")
	}
	b.WriteString("\n")

	// Data rows (only different/redacted fields)
	keys := sortedMapKeys(fields)
	for _, name := range keys {
		fieldDiff := fields[name]
		if fieldDiff.Status == diff.StatusEqual {
			continue
		}

		fmt.Fprintf(&b, "| %s |", name)
		for _, node := range d.Nodes {
			val := formatMarkdownValue(fieldDiff.NodeValues[node])
			// Bold outliers
			isOutlier := false
			for _, o := range fieldDiff.Outliers {
				if o == node {
					isOutlier = true
					break
				}
			}
			if isOutlier || (fieldDiff.Status == diff.StatusDifferent && len(d.Nodes) == 2) {
				val = fmt.Sprintf("**%s**", val)
			}
			fmt.Fprintf(&b, " %s |", val)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	return b.String()
}

func formatMarkdownValue(v any) string {
	if v == nil {
		return "—"
	}
	s := fmt.Sprintf("%v", v)
	if s == secrets.RedactedValue {
		return "🔒"
	}
	return s
}
