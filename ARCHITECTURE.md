# Architecture

## Overview

envdiff captures environment snapshots and intelligently diffs them, designed to answer questions like:
- "Why does CI fail when it passes locally?"
- "What changed since it last worked?"
- "Does the new hire's environment match the team's?"

The tool follows a modular architecture with clear separation of concerns:
1. **Data Collection** - Gathers environment information from the local system
2. **Configuration** - Defines environment requirements via YAML
3. **Comparison** - Diffs snapshots and identifies discrepancies
4. **Validation** - Checks environments against declared constraints
5. **Rendering** - Formats output for humans (CLI or Markdown)

## Directory Structure

```
envdiff/
├── cmd/envdiff/           # CLI layer (Cobra commands)
│   ├── main.go            # Entry point, root command setup
│   ├── snapshot.go        # 'envdiff snapshot' command
│   ├── compare.go         # 'envdiff compare' command
│   ├── check.go           # 'envdiff check' command
│   ├── render.go          # 'envdiff render' command
│   └── init.go            # 'envdiff init' command
│
├── internal/
│   ├── collector/         # Data gathering from the system
│   │   ├── collector.go   # Collector interface and orchestration
│   │   ├── system.go      # OS, architecture, hardware info
│   │   ├── runtime.go     # Language/tool versions (Go, Node, Python, etc.)
│   │   ├── env.go         # Environment variables
│   │   └── network.go     # Network configuration, listening ports
│   │
│   ├── config/            # YAML configuration handling
│   │   └── config.go      # Config struct, parsing, templates
│   │
│   ├── snapshot/          # Core data model
│   │   └── snapshot.go    # Snapshot struct, JSON serialization
│   │
│   ├── diff/              # Comparison engine
│   │   ├── diff.go        # Diff struct, field diff types
│   │   └── compare.go     # Comparison logic, majority/outlier detection
│   │
│   ├── check/             # Environment validation
│   │   ├── check.go       # Validation logic, semver constraints
│   │   └── render.go      # Check result formatting
│   │
│   ├── render/            # Output formatting
│   │   ├── render.go      # Renderer interface
│   │   ├── cli.go         # Terminal output with lipgloss styling
│   │   └── markdown.go    # Markdown table output
│   │
│   └── secrets/           # Secret detection and redaction
│       └── detect.go      # Pattern-based secret identification
│
└── envdiff.yaml           # Example configuration file
```

## Core Concepts

### Snapshot

A `Snapshot` is an immutable, content-addressable capture of an environment at a point in time. Snapshots include:

| Section | Contents |
|---------|----------|
| System | OS, version, architecture, kernel, CPU cores, memory |
| Runtime | Installed tools with versions and paths (Go, Node, Python, Docker, etc.) |
| Env | Environment variables (with optional redaction) |
| Network | Hosts file entries, listening ports |

Snapshots are serialized to JSON and identified by a content hash (first 8 hex chars of SHA-256).

### Configuration

The `envdiff.yaml` configuration file declares environment requirements:

```yaml
runtime:
  go: ">= 1.21.0"
  node: ">= 18.0.0"
  docker: "*"          # any version

env:
  required:
    - DATABASE_URL
  expected:
    NODE_ENV: development
  ignore:
    - "LC_*"
    - "*_TOKEN*"

fix:
  node:
    missing: "brew install node@20"
    wrong_version: "nvm use 20"
```

### Diff

A `Diff` compares two or more snapshots and categorizes each field as:
- **equal** - Same value across all nodes
- **different** - Values differ between nodes
- **redacted** - Contains secrets, not compared

For N>2 node comparisons, the diff engine identifies majority values and outliers.

### Check

The `Check` operation validates a local snapshot against configuration constraints. It produces a report with:
- **pass** - Requirement satisfied
- **fail** - Requirement not met
- **warn** - Unable to verify (e.g., invalid constraint syntax)

## Data Flow

### Snapshot Creation

```
┌─────────────────┐     ┌───────────────────┐     ┌──────────────┐
│  CLI Command    │────▶│  CollectAll()     │────▶│  Snapshot    │
│  snapshot       │     │                   │     │  (JSON)      │
└─────────────────┘     └───────────────────┘     └──────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │  Collectors:        │
                    │  - SystemCollector  │
                    │  - RuntimeCollector │
                    │  - EnvCollector     │
                    │  - NetworkCollector │
                    └─────────────────────┘
```

### Environment Comparison

```
┌──────────────┐     ┌───────────────────┐     ┌──────────────┐
│  Snapshots   │────▶│  diff.Compare()   │────▶│  Diff        │
│  (2+ nodes)  │     │                   │     │  result      │
└──────────────┘     └───────────────────┘     └──────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │  Per-field analysis │
                    │  - System fields    │
                    │  - Runtime versions │
                    │  - Env variables    │
                    └─────────────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │  Renderer           │
                    │  - CLI (terminal)   │
                    │  - Markdown         │
                    └─────────────────────┘
```

### Environment Check

```
┌──────────────┐     ┌───────────────────┐     ┌──────────────┐
│  Snapshot +  │────▶│  check.Check()    │────▶│  Report      │
│  Config      │     │                   │     │  (pass/fail) │
└──────────────┘     └───────────────────┘     └──────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │  Validations:       │
                    │  - Semver checks    │
                    │  - Required env     │
                    │  - Expected values  │
                    └─────────────────────┘
```

## Design Patterns

### Collector Pattern

The `Collector` interface defines a single method for gathering environment data:

```go
type Collector interface {
    Collect(s *snapshot.Snapshot) error
}
```

Each collector is independent and failures are isolated, allowing partial snapshots. Implementations:

| Collector | Responsibility |
|-----------|---------------|
| `SystemCollector` | OS, architecture, hardware information |
| `RuntimeCollector` | Installed tools and their versions |
| `EnvCollector` | Environment variables (with redaction support) |
| `NetworkCollector` | Network configuration and listening ports |

The `CollectAll()` function orchestrates all collectors:

```go
func CollectAll(s *snapshot.Snapshot, redact bool) error {
    collectors := []Collector{
        &SystemCollector{},
        &RuntimeCollector{},
        &EnvCollector{Redact: redact},
        &NetworkCollector{},
    }
    for _, c := range collectors {
        if err := c.Collect(s); err != nil {
            continue // Partial snapshot over total failure
        }
    }
    return nil
}
```

### Renderer Pattern

The `Renderer` interface provides polymorphic output formatting:

```go
type Renderer interface {
    RenderSnapshot(s *snapshot.Snapshot) string
    RenderDiff(d *diff.Diff) string
}
```

Implementations:

| Renderer | Output |
|----------|--------|
| `CLIRenderer` | Styled terminal output using lipgloss |
| `MarkdownRenderer` | GitHub-flavored Markdown tables |

### Strategy Pattern in Comparison

The `Compare()` function uses different strategies based on the number of nodes:

- **2-node comparison**: Simple pairwise diff showing `value1 -> value2`
- **N-node comparison**: Majority/outlier detection identifying consensus and deviations

```go
if len(nodes) == 2 {
    // Show "val1 -> val2"
} else {
    // Calculate majority and identify outliers
    fd.Majority, fd.Outliers = calculateMajority(values, nodes)
}
```

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework for commands, flags, and help |
| `gopkg.in/yaml.v3` | YAML configuration parsing |
| `github.com/Masterminds/semver/v3` | Semantic version constraint validation |
| `github.com/charmbracelet/lipgloss` | Terminal styling and formatting |

## Extension Points

### Adding a New Collector

1. Create a new file in `internal/collector/` (e.g., `docker.go`)
2. Implement the `Collector` interface:

```go
type DockerCollector struct{}

func (c *DockerCollector) Collect(s *snapshot.Snapshot) error {
    // Gather Docker-specific data
    // Populate relevant fields in s
    return nil
}
```

3. Register the collector in `CollectAll()`:

```go
collectors := []Collector{
    &SystemCollector{},
    &RuntimeCollector{},
    &EnvCollector{Redact: redact},
    &NetworkCollector{},
    &DockerCollector{},  // New collector
}
```

### Adding a New Output Format

1. Create a new file in `internal/render/` (e.g., `json.go`)
2. Implement the `Renderer` interface:

```go
type JSONRenderer struct{}

func (r *JSONRenderer) RenderSnapshot(s *snapshot.Snapshot) string {
    data, _ := json.MarshalIndent(s, "", "  ")
    return string(data)
}

func (r *JSONRenderer) RenderDiff(d *diff.Diff) string {
    data, _ := json.MarshalIndent(d, "", "  ")
    return string(data)
}
```

3. Add the format option to the CLI commands in `cmd/envdiff/`

### Adding New Check Types

1. Extend the `Config` struct in `internal/config/config.go`:

```go
type Config struct {
    Runtime   map[string]string    `yaml:"runtime"`
    Env       EnvConfig            `yaml:"env"`
    Packages  []string             `yaml:"packages,omitempty"`
    Services  []string             `yaml:"services,omitempty"`  // New
    Fix       map[string]FixConfig `yaml:"fix,omitempty"`
}
```

2. Add the validation logic in `internal/check/check.go`:

```go
func Check(snap *snapshot.Snapshot, cfg *config.Config) *Report {
    // ... existing checks ...

    // Check required services
    for _, service := range cfg.Services {
        result := checkService(snap, service, cfg.Fix[service])
        report.Results = append(report.Results, result)
        updateCounts(report, result.Status)
    }

    return report
}
```

### Adding Secret Detection Patterns

Add new regex patterns to `secretPatterns` in `internal/secrets/detect.go`:

```go
var secretPatterns = []*regexp.Regexp{
    // ... existing patterns ...
    regexp.MustCompile(`(?i)github[_-]?token`),
    regexp.MustCompile(`(?i)stripe[_-]?key`),
}
```
