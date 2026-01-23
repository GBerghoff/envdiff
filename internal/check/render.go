package check

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/GBerghoff/envdiff/internal/ui"
)

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

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorWarn))

	keyStyle = lipgloss.NewStyle().
			Width(14)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorValue))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorSecondary))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorDivider))
)

// RenderCLI renders the check report for terminal display
func (r *Report) RenderCLI() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("envdiff") + " — checking requirements\n")

	// Group by category
	runtimeResults := []Result{}
	envResults := []Result{}

	for _, result := range r.Results {
		switch result.Category {
		case "runtime":
			runtimeResults = append(runtimeResults, result)
		case "env":
			envResults = append(envResults, result)
		}
	}

	// Render runtime section
	if len(runtimeResults) > 0 {
		b.WriteString(headerStyle.Render("RUNTIME") + "\n")
		for _, result := range runtimeResults {
			b.WriteString(renderResult(result))
		}
	}

	// Render env section
	if len(envResults) > 0 {
		b.WriteString(headerStyle.Render("ENVIRONMENT") + "\n")
		for _, result := range envResults {
			b.WriteString(renderResult(result))
		}
	}

	// Summary
	b.WriteString("\n")
	b.WriteString(dividerStyle.Render(strings.Repeat("─", 40)) + "\n")
	b.WriteString(fmt.Sprintf("%d passed · %d failed", r.Passed, r.Failed))
	if r.Warned > 0 {
		b.WriteString(fmt.Sprintf(" · %d warnings", r.Warned))
	}
	b.WriteString("\n")

	// Fix hints
	hints := collectFixHints(r)
	if len(hints) > 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("To fix:") + "\n")
		for _, hint := range hints {
			b.WriteString(fmt.Sprintf("  • %s\n", hint))
		}
	}

	return b.String()
}

func renderResult(r Result) string {
	var icon string
	switch r.Status {
	case StatusPass:
		icon = checkStyle.Render("✓")
	case StatusFail:
		icon = crossStyle.Render("✗")
	case StatusWarn:
		icon = warnStyle.Render("⚠")
	}

	line := fmt.Sprintf("  %s %s %s",
		icon,
		keyStyle.Render(r.Name),
		valueStyle.Render(r.Actual))

	if r.Status == StatusPass || r.Status == StatusFail {
		line += dimStyle.Render(fmt.Sprintf(" (requires %s)", r.Expected))
	}

	return line + "\n"
}

func collectFixHints(r *Report) []string {
	var hints []string
	for _, result := range r.Results {
		if result.Status == StatusFail && result.FixHint != "" {
			hints = append(hints, fmt.Sprintf("%s: %s", result.Name, result.FixHint))
		}
	}
	return hints
}

// RenderJSON renders the check report as JSON
func (r *Report) RenderJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
