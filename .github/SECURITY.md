# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in envdiff, please report it responsibly.

**Do not open a public GitHub issue for security vulnerabilities.**

Instead, please report security issues by emailing the maintainers directly or by using GitHub's private vulnerability reporting feature:

1. Go to the [Security tab](https://github.com/GBerghoff/envdiff/security) of this repository
2. Click "Report a vulnerability"
3. Provide a detailed description of the vulnerability

### What to include in your report

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any suggested fixes (optional)

### What to expect

- **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
- **Updates**: We will provide updates on our progress as we investigate
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days
- **Credit**: We will credit you in the release notes (unless you prefer to remain anonymous)

## Security Considerations

envdiff collects environment information which may include sensitive data. Please note:

- **Secret redaction is enabled by default** - Environment variables matching common secret patterns are automatically redacted
- Use `--no-redact` only when you are certain the output will not be shared or committed
- Snapshot files may contain system information (hostnames, IPs, tool versions) that could aid attackers
- Review snapshot output before sharing or committing to version control

## Scope

This security policy applies to:

- The envdiff CLI tool
- Official releases distributed via GitHub Releases
- The installation script (`install.sh`)

Third-party forks or modifications are not covered by this policy.
