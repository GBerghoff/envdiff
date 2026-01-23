package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GBerghoff/envdiff/internal/diff"
	"github.com/GBerghoff/envdiff/internal/secrets"
	"github.com/GBerghoff/envdiff/internal/snapshot"
	"github.com/GBerghoff/envdiff/internal/ui"
)

// CLIRenderer renders output for the terminal
type CLIRenderer struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ui.ColorPrimary))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ui.ColorSecondary)).
			MarginTop(1)

	checkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorPass))

	crossStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorFail))

	redactedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorSecondary))

	keyStyle = lipgloss.NewStyle().
			Width(14)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorValue))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorSecondary))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorDivider))
)

// NewCLI creates a new CLI renderer
func NewCLI() *CLIRenderer {
	return &CLIRenderer{}
}

// RenderSnapshot renders a snapshot for terminal display
func (r *CLIRenderer) RenderSnapshot(s *snapshot.Snapshot) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("envdiff") + " — snapshot of " + s.Hostname + "\n")

	// System info
	b.WriteString(headerStyle.Render("SYSTEM") + "\n")
	b.WriteString(fmt.Sprintf("  %s %s\n", keyStyle.Render("os"), valueStyle.Render(s.System.OSVersion)))
	b.WriteString(fmt.Sprintf("  %s %s\n", keyStyle.Render("arch"), valueStyle.Render(s.System.Arch)))
	b.WriteString(fmt.Sprintf("  %s %s\n", keyStyle.Render("kernel"), valueStyle.Render(s.System.Kernel)))
	b.WriteString(fmt.Sprintf("  %s %d cores\n", keyStyle.Render("cpu"), s.System.CPUCores))
	b.WriteString(fmt.Sprintf("  %s %dGB\n", keyStyle.Render("memory"), s.System.MemoryGB))

	// Runtime info
	b.WriteString(headerStyle.Render("RUNTIME") + "\n")
	runtimes := sortedKeys(s.Runtime)
	for _, name := range runtimes {
		info := s.Runtime[name]
		if info != nil {
			b.WriteString(fmt.Sprintf("  %s %s %s\n",
				checkStyle.Render("✓"),
				keyStyle.Render(name),
				valueStyle.Render(info.Version)))
		}
	}

	// Environment summary
	b.WriteString(headerStyle.Render("ENVIRONMENT") + "\n")
	redactedCount := 0
	for _, v := range s.Env {
		if v == secrets.RedactedValue {
			redactedCount++
		}
	}
	b.WriteString(fmt.Sprintf("  %d variables (%d redacted)\n", len(s.Env), redactedCount))

	return b.String()
}

// RenderDiff renders a diff for terminal display
func (r *CLIRenderer) RenderDiff(d *diff.Diff) string {
	var b strings.Builder

	// Title
	nodeList := strings.Join(d.Nodes, " ↔ ")
	b.WriteString(titleStyle.Render("envdiff") + " — comparing " + nodeList + "\n")

	// Errors
	if len(d.Errors) > 0 {
		for node, err := range d.Errors {
			b.WriteString(fmt.Sprintf("  %s %s: %s\n", crossStyle.Render("⚠"), node, err))
		}
		b.WriteString("\n")
	}

	// Runtime diffs
	if len(d.Diffs["runtime"]) > 0 {
		b.WriteString(headerStyle.Render("RUNTIME") + "\n")
		runtimes := sortedMapKeys(d.Diffs["runtime"])
		for _, name := range runtimes {
			fieldDiff := d.Diffs["runtime"][name]
			b.WriteString(r.renderFieldDiff(name, fieldDiff, d.Nodes))
		}
	}

	// Environment diffs
	if len(d.Diffs["env"]) > 0 {
		b.WriteString(headerStyle.Render("ENVIRONMENT") + "\n")
		// Only show different or redacted fields
		envKeys := sortedMapKeys(d.Diffs["env"])
		shownCount := 0
		equalCount := 0
		for _, name := range envKeys {
			fieldDiff := d.Diffs["env"][name]
			if fieldDiff.Status != diff.StatusEqual {
				b.WriteString(r.renderFieldDiff(name, fieldDiff, d.Nodes))
				shownCount++
			} else {
				equalCount++
			}
		}
		if equalCount > 0 {
			b.WriteString(fmt.Sprintf("  %s %d variables match\n", checkStyle.Render("✓"), equalCount))
		}
	}

	// System diffs
	if len(d.Diffs["system"]) > 0 {
		b.WriteString(headerStyle.Render("SYSTEM") + "\n")
		systemKeys := sortedMapKeys(d.Diffs["system"])
		equalCount := 0
		for _, name := range systemKeys {
			fieldDiff := d.Diffs["system"][name]
			if fieldDiff.Status != diff.StatusEqual {
				b.WriteString(r.renderFieldDiff(name, fieldDiff, d.Nodes))
			} else {
				equalCount++
			}
		}
		if equalCount > 0 {
			b.WriteString(fmt.Sprintf("  %s %d fields match\n", checkStyle.Render("✓"), equalCount))
		}
	}

	// Summary
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render(strings.Repeat("─", 40)) + "\n")
	b.WriteString(fmt.Sprintf("%d different · %d equal · %d redacted\n",
		d.Summary.Different, d.Summary.Equal, d.Summary.Redacted))

	if d.Summary.Redacted > 0 {
		b.WriteString(dimStyle.Render("\nNote: Redacted values not compared. Use --no-redact to include.") + "\n")
	}

	return b.String()
}

func (r *CLIRenderer) renderFieldDiff(name string, fieldDiff *diff.FieldDiff, nodes []string) string {
	switch fieldDiff.Status {
	case diff.StatusEqual:
		val := formatValue(fieldDiff.NodeValues[nodes[0]])
		return fmt.Sprintf("  %s %s %s\n",
			checkStyle.Render("✓"),
			keyStyle.Render(name),
			valueStyle.Render(val))

	case diff.StatusRedacted:
		return fmt.Sprintf("  %s %s %s\n",
			redactedStyle.Render("⊘"),
			keyStyle.Render(name),
			redactedStyle.Render(secrets.RedactedValue))

	case diff.StatusDifferent:
		if len(nodes) == 2 {
			// Two-node diff: show "val1 → val2"
			val1 := formatValue(fieldDiff.NodeValues[nodes[0]])
			val2 := formatValue(fieldDiff.NodeValues[nodes[1]])
			return fmt.Sprintf("  %s %s %s → %s\n",
				crossStyle.Render("✗"),
				keyStyle.Render(name),
				valueStyle.Render(val1),
				valueStyle.Render(val2))
		} else {
			// Multi-node diff: show outliers
			var line strings.Builder
			line.WriteString(fmt.Sprintf("  %s %s ",
				crossStyle.Render("✗"),
				keyStyle.Render(name)))

			if fieldDiff.Majority != nil {
				line.WriteString(fmt.Sprintf("%s ", valueStyle.Render(formatValue(fieldDiff.Majority))))
				if len(fieldDiff.Outliers) > 0 {
					outlierValues := []string{}
					for _, node := range fieldDiff.Outliers {
						outlierValues = append(outlierValues,
							fmt.Sprintf("%s=%s", node, formatValue(fieldDiff.NodeValues[node])))
					}
					line.WriteString(dimStyle.Render("(outliers: " + strings.Join(outlierValues, ", ") + ")"))
				}
			} else {
				// No clear majority, list all values
				values := []string{}
				for _, node := range nodes {
					values = append(values, fmt.Sprintf("%s=%s", node, formatValue(fieldDiff.NodeValues[node])))
				}
				line.WriteString(dimStyle.Render(strings.Join(values, ", ")))
			}
			line.WriteString("\n")
			return line.String()
		}
	}
	return ""
}

func formatValue(v any) string {
	if v == nil {
		return "(missing)"
	}
	return fmt.Sprintf("%v", v)
}

func sortedKeys[V any](m map[string]*V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedMapKeys(m map[string]*diff.FieldDiff) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
