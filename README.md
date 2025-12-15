# envdiff

> "It works on my machine" → "Show me exactly why."

Compare any two environments and get a precise, categorized diff of everything that could possibly matter.

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

```
envdiff — snapshot of macbook-pro

SYSTEM
  os             macOS 14.2.1
  arch           arm64
  kernel         23.2.0
  cpu            10 cores
  memory         32GB

RUNTIME
  ✓ go             1.22.0
  ✓ node           20.11.0
  ✓ python         3.12.1
  ✓ docker         24.0.7

ENVIRONMENT
  47 variables (5 redacted)
```

### Diff (CLI)

```
envdiff — comparing local ↔ ci

RUNTIME
  ✗ node      20.11.0 → 18.19.0
  ✗ python    3.12.1 → (missing)
  ✓ go        1.22.0

ENVIRONMENT
  ✗ NODE_ENV       development → production
  ⊘ DATABASE_URL   [REDACTED]
  ✓ 42 variables match

────────────────────────────────────────
2 different · 42 equal · 1 redacted
```

### Check

```
envdiff — checking requirements

RUNTIME
  ✓ go             1.22.0     (requires >= 1.21.0)
  ✓ node           20.11.0    (requires >= 18.0.0)
  ✗ python         (missing)  (requires >= 3.10.0)

ENVIRONMENT
  ✓ DATABASE_URL   (set)
  ✗ AWS_PROFILE    (missing)

────────────────────────────────────────
3 passed · 2 failed

To fix:
  • python: brew install python@3.10
  • AWS_PROFILE: Set required environment variable
```

## Use Cases

1. **"Why does CI fail when it passes locally?"** — Snapshot both, diff them
2. **New hire onboarding** — `envdiff check` beats a 47-step setup doc
3. **Incident debugging** — "What changed since it last worked?"
