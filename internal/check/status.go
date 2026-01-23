package check

// CheckStatus represents the result status of a requirement check.
type CheckStatus string

// Check status constants for validation results.
const (
	StatusPass CheckStatus = "pass"
	StatusFail CheckStatus = "fail"
	StatusWarn CheckStatus = "warn"
)
