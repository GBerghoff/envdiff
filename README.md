![envdiff](assets/banner.png)

# envdiff

[![CI](https://github.com/GBerghoff/envdiff/actions/workflows/ci.yml/badge.svg)](https://github.com/GBerghoff/envdiff/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GBerghoff/envdiff)](https://goreportcard.com/report/github.com/GBerghoff/envdiff)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/GBerghoff/envdiff)](https://github.com/GBerghoff/envdiff/releases/latest)
[![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos%20%7C%20windows-lightgrey)](https://github.com/GBerghoff/envdiff/releases)

> "It works on my machine" → "Show me exactly why."

Compare any two environments and get a precise, categorized diff of everything that could possibly matter.

## Why envdiff?

Unlike running `diff <(env)` or manual comparison:

- **Structured comparison** - Categorizes by system, runtime, env vars, network
- **Auto-redaction** - Secrets detected and hidden automatically
- **Declarative validation** - Define requirements in YAML, validate anywhere
- **Multiple formats** - JSON, CLI, Markdown outputs
- **CI/CD ready** - Exit codes and JSON for automation

## Install

```bash
# Quick install (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/GBerghoff/envdiff/main/install.sh | sh

# Homebrew (coming soon)
# brew install GBerghoff/tap/envdiff

# Go install
go install github.com/GBerghoff/envdiff/cmd/envdiff@latest

# Or download binary directly from releases
# https://github.com/GBerghoff/envdiff/releases
```

## Quick Start

```bash
# Take a snapshot of your environment
envdiff snapshot -o local.json

# Check your environment against requirements
envdiff init                    # Create envdiff.yaml template
envdiff check                   # Validate against requirements

# Compare two snapshots
envdiff compare local.json ci.json
envdiff compare local.json ci.json | envdiff render -
```

## Commands

### `envdiff snapshot`

Capture your environment to JSON.

```bash
envdiff snapshot                    # JSON to stdout
envdiff snapshot -o local.json      # Save to file
envdiff snapshot --format cli       # Pretty terminal output
envdiff snapshot --format md        # Markdown output
envdiff snapshot --no-redact        # Include secret values
```

**What's captured:**
- System info (OS, arch, kernel, memory, CPU)
- Runtime versions (go, node, python, docker, kubectl, etc.)
- Environment variables (secrets auto-redacted)
- Network info (/etc/hosts, listening ports)

### `envdiff compare`

Compare two or more snapshots.

```bash
envdiff compare local.json ci.json              # Two snapshots
envdiff compare local.json ci.json staging.json # Multiple snapshots
envdiff compare local.json ci.json -o diff.json # Save diff
```

### `envdiff render`

Render JSON snapshots or diffs for humans.

```bash
envdiff render snapshot.json        # CLI output
envdiff render diff.json --md       # Markdown output
```

### `envdiff check`

Validate your environment against `envdiff.yaml`.

```bash
envdiff check                       # Use ./envdiff.yaml
envdiff check --file staging.yaml   # Custom config
envdiff check --json                # JSON output
envdiff check --quiet               # Exit code only (for CI/hooks)
```

### `envdiff init`

Create an `envdiff.yaml` template.

```bash
envdiff init                        # Create ./envdiff.yaml
envdiff init --force                # Overwrite existing
```

## Configuration

`envdiff.yaml` defines your environment requirements:

```yaml
runtime:
  go: ">= 1.21.0"
  node: ">= 18.0.0"
  python: ">= 3.10.0"
  docker: "*"              # any version

env:
  required:
    - DATABASE_URL
    - AWS_PROFILE

  expected:
    NODE_ENV: development

  ignore:
    - TERM
    - SHELL
    - "*_SESSION*"

fix:
  node:
    missing: "brew install node@20"
    wrong_version: "nvm use 20"
```

## Example Output

### Snapshot (CLI)

![envdiff snapshot](assets/snapshot.gif)

### Diff (CLI)

![envdiff diff](assets/diff.gif)

### Check

![envdiff check](assets/check.gif)
