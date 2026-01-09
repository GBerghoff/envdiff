package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/GBerghoff/envdiff/internal/diff"
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// CLIRenderer renders output for the terminal
type CLIRenderer struct{}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	checkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	crossStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	redactedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	keyStyle = lipgloss.NewStyle().
			Width(14)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))
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
		if v == "[REDACTED]" {
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
			fd := d.Diffs["runtime"][name]
			b.WriteString(r.renderFieldDiff(name, fd, d.Nodes))
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
			fd := d.Diffs["env"][name]
			if fd.Status != "equal" {
				b.WriteString(r.renderFieldDiff(name, fd, d.Nodes))
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
			fd := d.Diffs["system"][name]
			if fd.Status != "equal" {
				b.WriteString(r.renderFieldDiff(name, fd, d.Nodes))
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

func (r *CLIRenderer) renderFieldDiff(name string, fd *diff.FieldDiff, nodes []string) string {
	switch fd.Status {
	case "equal":
		val := formatValue(fd.Values[nodes[0]])
		return fmt.Sprintf("  %s %s %s\n",
			checkStyle.Render("✓"),
			keyStyle.Render(name),
			valueStyle.Render(val))

	case "redacted":
		return fmt.Sprintf("  %s %s %s\n",
			redactedStyle.Render("⊘"),
			keyStyle.Render(name),
			redactedStyle.Render("[REDACTED]"))

	case "different":
		if len(nodes) == 2 {
			// Two-node diff: show "val1 → val2"
			val1 := formatValue(fd.Values[nodes[0]])
			val2 := formatValue(fd.Values[nodes[1]])
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

			if fd.Majority != nil {
				line.WriteString(fmt.Sprintf("%s ", valueStyle.Render(formatValue(fd.Majority))))
				if len(fd.Outliers) > 0 {
					outlierVals := []string{}
					for _, node := range fd.Outliers {
						outlierVals = append(outlierVals,
							fmt.Sprintf("%s=%s", node, formatValue(fd.Values[node])))
					}
					line.WriteString(dimStyle.Render("(outliers: " + strings.Join(outlierVals, ", ") + ")"))
				}
			} else {
				// No clear majority, list all values
				vals := []string{}
				for _, node := range nodes {
					vals = append(vals, fmt.Sprintf("%s=%s", node, formatValue(fd.Values[node])))
				}
				line.WriteString(dimStyle.Render(strings.Join(vals, ", ")))
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
