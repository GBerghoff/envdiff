package check

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	keyStyle = lipgloss.NewStyle().
			Width(14)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))
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
	var icon, style string
	switch r.Status {
	case "pass":
		icon = checkStyle.Render("✓")
		style = "pass"
	case "fail":
		icon = crossStyle.Render("✗")
		style = "fail"
	case "warn":
		icon = warnStyle.Render("⚠")
		style = "warn"
	}

	line := fmt.Sprintf("  %s %s %s",
		icon,
		keyStyle.Render(r.Name),
		valueStyle.Render(r.Actual))

	if style == "pass" {
		line += dimStyle.Render(fmt.Sprintf(" (requires %s)", r.Expected))
	} else if style == "fail" {
		line += dimStyle.Render(fmt.Sprintf(" (requires %s)", r.Expected))
	}

	return line + "\n"
}

func collectFixHints(r *Report) []string {
	var hints []string
	for _, result := range r.Results {
		if result.Status == "fail" && result.FixHint != "" {
			hints = append(hints, fmt.Sprintf("%s: %s", result.Name, result.FixHint))
		}
	}
	return hints
}

// RenderJSON renders the check report as JSON
func (r *Report) RenderJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
