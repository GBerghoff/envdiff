# Troubleshooting Guide

## Table of Contents

- [Installation Issues](#installation-issues)
- [Snapshot Issues](#snapshot-issues)
- [Compare Issues](#compare-issues)
- [Check Issues](#check-issues)
- [Output Issues](#output-issues)
- [CI/CD Issues](#cicd-issues)
- [Getting Help](#getting-help)

---

## Installation Issues

### Binary not found after install

**Symptom:** After installation, running `envdiff` returns "command not found" or similar.

**Solution:** The envdiff binary is likely not in your system PATH. Add the Go bin directory to your PATH:

```bash
# For bash (add to ~/.bashrc or ~/.bash_profile)
export PATH="$PATH:$(go env GOPATH)/bin"

# For zsh (add to ~/.zshrc)
export PATH="$PATH:$(go env GOPATH)/bin"

# For fish (add to ~/.config/fish/config.fish)
set -gx PATH $PATH (go env GOPATH)/bin
```

After updating your shell configuration, reload it:

```bash
source ~/.bashrc  # or ~/.zshrc, etc.
```

Alternatively, verify the binary location and add it explicitly:

```bash
which envdiff
# or
ls $(go env GOPATH)/bin/envdiff
```

### Permission denied on install script

**Symptom:** Running the install script fails with "Permission denied".

**Solution:** Make the install script executable or run it with bash directly:

```bash
# Option 1: Make executable
chmod +x install.sh
./install.sh

# Option 2: Run with bash
bash install.sh

# Option 3: If installing to a system directory, use sudo
sudo bash install.sh
```

### Go install fails

**Symptom:** `go install` command fails with errors.

**Solution:** Check the following:

1. **Verify Go is installed and in PATH:**
   ```bash
   go version
   ```

2. **Ensure Go version is compatible (1.21+):**
   ```bash
   go version
   # Should show go1.21 or higher
   ```

3. **Check GOPATH and GOBIN are set correctly:**
   ```bash
   go env GOPATH
   go env GOBIN
   ```

4. **Try clearing the module cache:**
   ```bash
   go clean -modcache
   go install github.com/GBerghoff/envdiff@latest
   ```

---

## Snapshot Issues

### Missing runtime versions (tool not in PATH)

**Symptom:** Snapshot shows empty or missing version information for runtimes like Node.js, Python, Go, etc.

**Solution:** The runtime executables must be in your PATH for envdiff to detect them:

1. **Verify the tool is accessible:**
   ```bash
   which node
   which python
   which go
   ```

2. **If using version managers (nvm, pyenv, etc.), ensure they're initialized:**
   ```bash
   # For nvm
   source ~/.nvm/nvm.sh

   # For pyenv
   eval "$(pyenv init -)"
   ```

3. **Run envdiff in the same shell context as your development environment.**

### Permission denied reading /etc/hosts

**Symptom:** Error message about permission denied when accessing `/etc/hosts` or other system files.

**Solution:**

1. **Run with appropriate permissions:**
   ```bash
   # If you need to read protected system files
   sudo envdiff snapshot
   ```

2. **Or exclude system file collection if not needed** (check available flags):
   ```bash
   envdiff snapshot --help
   ```

### Secrets showing as [REDACTED] when you need values

**Symptom:** Environment variables or sensitive data appear as `[REDACTED]` in the snapshot output.

**Solution:** By default, envdiff redacts potentially sensitive values for security. If you need to see the actual values (e.g., for debugging in a secure environment):

```bash
# Use the --no-redact flag to show actual values
envdiff snapshot --no-redact

# Warning: Be careful not to commit or share unredacted snapshots
# as they may contain sensitive information like API keys and passwords
```

**Security Note:** Never share unredacted snapshots publicly or commit them to version control.

---

## Compare Issues

### "No differences found" when expecting differences

**Symptom:** Running `envdiff compare` reports no differences, but you know there should be some.

**Solution:**

1. **Verify you're comparing the correct snapshots:**
   ```bash
   # Check the snapshot files contain expected data
   cat snapshot1.json | head -50
   cat snapshot2.json | head -50
   ```

2. **Ensure snapshots were taken at different times/states:**
   ```bash
   # Check timestamps in snapshots
   grep -i "timestamp\|created" snapshot1.json snapshot2.json
   ```

3. **Check if differences are in sections you're filtering out:**
   ```bash
   # Compare with verbose output if available
   envdiff compare snapshot1.json snapshot2.json --verbose
   ```

4. **Verify the snapshots aren't identical files:**
   ```bash
   diff snapshot1.json snapshot2.json
   ```

### Schema version mismatch between snapshots

**Symptom:** Error about incompatible or mismatched schema versions when comparing snapshots.

**Solution:**

1. **Check the schema versions in both snapshots:**
   ```bash
   grep -i "schema\|version" snapshot1.json | head -5
   grep -i "schema\|version" snapshot2.json | head -5
   ```

2. **Regenerate snapshots with the same version of envdiff:**
   ```bash
   # Update envdiff to latest
   go install github.com/GBerghoff/envdiff@latest

   # Regenerate both snapshots
   envdiff snapshot -o snapshot1.json
   # (change environment)
   envdiff snapshot -o snapshot2.json
   ```

3. **If comparing historical snapshots, check if a migration tool is available:**
   ```bash
   envdiff --help
   ```

---

## Check Issues

### envdiff.yaml not found

**Symptom:** Running `envdiff check` fails because it cannot find the configuration file.

**Solution:**

1. **Ensure envdiff.yaml exists in your project root:**
   ```bash
   ls -la envdiff.yaml
   ```

2. **Create a basic configuration file if it doesn't exist:**
   ```yaml
   # envdiff.yaml
   version: "1"
   checks:
     runtimes:
       node: ">=18.0.0"
       go: ">=1.21.0"
     env:
       - NODE_ENV
       - PATH
   ```

3. **Specify the config file path explicitly:**
   ```bash
   envdiff check --config /path/to/your/envdiff.yaml
   ```

4. **Check for typos in the filename** (must be `envdiff.yaml` or `envdiff.yml`).

### Semver constraint syntax errors

**Symptom:** Errors parsing version constraints in envdiff.yaml.

**Solution:** Use valid semver constraint syntax:

```yaml
# Valid constraint examples
runtimes:
  node: ">=18.0.0"           # Greater than or equal
  node: "^18.0.0"            # Compatible with 18.x.x
  node: "~18.1.0"            # Approximately 18.1.x
  node: ">=18.0.0 <20.0.0"   # Range
  node: "18.0.0 || 20.0.0"   # Either version
  go: ">=1.21"               # Partial versions OK
  python: ">=3.9.0"

# Invalid examples (avoid these)
runtimes:
  node: "18"                 # Missing constraint operator
  node: ">= 18.0.0"          # Space after operator (may cause issues)
  node: "v18.0.0"            # Don't include 'v' prefix
```

### Environment variable not detected

**Symptom:** `envdiff check` reports a required environment variable as missing, but it's set.

**Solution:**

1. **Verify the variable is actually set:**
   ```bash
   echo $VARIABLE_NAME
   printenv VARIABLE_NAME
   ```

2. **Check for typos in envdiff.yaml:**
   ```yaml
   env:
     - MY_VARIABLE    # Case-sensitive!
     - my_variable    # This is different from above
   ```

3. **Ensure the variable is exported (not just set):**
   ```bash
   # This won't be detected:
   MY_VAR=value

   # This will be detected:
   export MY_VAR=value
   ```

4. **If running in a subshell or script, ensure variables are passed through:**
   ```bash
   export MY_VAR=value
   envdiff check
   ```

---

## Output Issues

### Colors not showing (terminal compatibility)

**Symptom:** Output appears without colors or with strange escape characters.

**Solution:**

1. **Check if your terminal supports colors:**
   ```bash
   echo $TERM
   # Should show something like xterm-256color, screen-256color, etc.
   ```

2. **Force color output or disable it:**
   ```bash
   # Force colors (if supported)
   envdiff snapshot --color=always

   # Disable colors
   envdiff snapshot --color=never
   # or
   envdiff snapshot --no-color
   ```

3. **If piping output, colors are typically disabled automatically. Use `--color=always` to force them:**
   ```bash
   envdiff compare a.json b.json --color=always | less -R
   ```

4. **For Windows terminals, use Windows Terminal or enable ANSI escape sequences in your terminal emulator.**

### Markdown output formatting

**Symptom:** Markdown output doesn't render correctly or has formatting issues.

**Solution:**

1. **Use the markdown output format flag:**
   ```bash
   envdiff compare a.json b.json --format markdown > report.md
   ```

2. **For proper rendering, view in a Markdown viewer:**
   ```bash
   # View in terminal with glow (if installed)
   envdiff compare a.json b.json --format markdown | glow -

   # Or save and open in your editor/viewer
   envdiff compare a.json b.json --format markdown > report.md
   ```

3. **If copying to documentation, ensure your Markdown renderer supports the table syntax used.**

---

## CI/CD Issues

### Exit codes explanation

envdiff uses standard exit codes for CI/CD integration:

| Exit Code | Meaning |
|-----------|---------|
| `0` | Success - all checks passed, or comparison completed |
| `1` | Failure - checks failed, differences found (for compare), or errors occurred |

**Usage in CI/CD pipelines:**

```bash
# GitHub Actions example
- name: Check environment
  run: |
    envdiff check
    # Pipeline fails automatically if exit code is non-zero

# Or capture the exit code
- name: Check environment
  run: |
    if envdiff check; then
      echo "All checks passed"
    else
      echo "Environment checks failed"
      exit 1
    fi
```

```bash
# Generic CI script
#!/bin/bash
set -e  # Exit on any error

envdiff check
echo "Environment validation passed!"
```

### JSON output for automation

**For parsing in CI/CD pipelines, use JSON output:**

```bash
# Output as JSON
envdiff snapshot --format json -o snapshot.json
envdiff compare a.json b.json --format json
envdiff check --format json

# Parse with jq
envdiff check --format json | jq '.passed'
envdiff compare a.json b.json --format json | jq '.differences'

# Example: Check specific values
if [ "$(envdiff check --format json | jq -r '.passed')" = "true" ]; then
  echo "All checks passed"
else
  echo "Checks failed"
  envdiff check --format json | jq '.failures'
  exit 1
fi
```

**GitHub Actions example with JSON parsing:**

```yaml
- name: Validate environment
  run: |
    RESULT=$(envdiff check --format json)
    if echo "$RESULT" | jq -e '.passed' > /dev/null; then
      echo "Environment validation successful"
    else
      echo "Environment validation failed:"
      echo "$RESULT" | jq '.failures'
      exit 1
    fi
```

---

## Getting Help

If you're still experiencing issues after trying the solutions above:

1. **Check existing issues:** Search the [GitHub Issues](https://github.com/GBerghoff/envdiff/issues) to see if your problem has already been reported or resolved.

2. **Open a new issue:** If your problem is new, please [open an issue](https://github.com/GBerghoff/envdiff/issues/new) with:
   - Your envdiff version (`envdiff --version`)
   - Your operating system and version
   - The command you ran
   - The complete error message
   - Steps to reproduce the issue

3. **Include relevant context:**
   - Sanitized configuration files (remove sensitive data)
   - Relevant environment details
   - Any workarounds you've tried

4. **Check the documentation:** Review the [README](README.md) for usage examples and configuration options.
