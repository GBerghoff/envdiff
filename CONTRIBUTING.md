# Contributing to envdiff

#### **Did you find a bug?**

* Check if it was already reported under [Issues](https://github.com/GBerghoff/envdiff/issues).
* If not, [open a new issue](https://github.com/GBerghoff/envdiff/issues/new) with a clear description and steps to reproduce.

#### **Did you write a patch that fixes a bug?**

* Open a pull request with a clear description of the problem and solution.
* Make sure tests pass: `go test -race ./...`

#### **Do you want to add a feature?**

* Open an issue first to discuss the idea before writing code.
* Good first contributions: new collectors, output formats, or platform-specific improvements.

## Development

```bash
git clone https://github.com/GBerghoff/envdiff.git
cd envdiff
go build ./cmd/envdiff/
go test -race ./...
```

Lint with [golangci-lint](https://golangci-lint.run/) before submitting:

```bash
golangci-lint run ./...
```

## Commits

Use conventional commits: `feat:`, `fix:`, `docs:`, `ci:`, `refactor:`, `test:`

```
feat: add Docker collector
fix: handle missing PATH variable
```

Thanks for contributing!
