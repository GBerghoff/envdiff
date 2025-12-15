// Package secrets identifies and redacts sensitive environment variables.
// Detection is pattern-based and intentionally conservative to avoid false negatives.
package secrets

import (
	"regexp"
	"strings"
)

// secretPatterns contains regex patterns that identify secret environment variables
var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)password`),
	regexp.MustCompile(`(?i)secret`),
	regexp.MustCompile(`(?i)token`),
	regexp.MustCompile(`(?i)api[_-]?key`),
	regexp.MustCompile(`(?i)auth`),
	regexp.MustCompile(`(?i)credential`),
	regexp.MustCompile(`(?i)private[_-]?key`),
	regexp.MustCompile(`(?i)access[_-]?key`),
	regexp.MustCompile(`(?i)ssh[_-]?key`),
	regexp.MustCompile(`(?i)database[_-]?url`),
	regexp.MustCompile(`(?i)connection[_-]?string`),
	regexp.MustCompile(`(?i)^db[_-]`),
	regexp.MustCompile(`(?i)_dsn$`),
	regexp.MustCompile(`(?i)bearer`),
	regexp.MustCompile(`(?i)jwt`),
	regexp.MustCompile(`(?i)session`),
	regexp.MustCompile(`(?i)cookie`),
	regexp.MustCompile(`(?i)encryption`),
	regexp.MustCompile(`(?i)cert`),
}

// RedactedValue is the placeholder for redacted secrets
const RedactedValue = "[REDACTED]"

// IsSecret checks if an environment variable name looks like a secret
func IsSecret(name string) bool {
	for _, pattern := range secretPatterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

// RedactEnv takes a map of environment variables and redacts secret values
func RedactEnv(env map[string]string) map[string]string {
	result := make(map[string]string, len(env))
	for k, v := range env {
		if IsSecret(k) {
			result[k] = RedactedValue
		} else {
			result[k] = v
		}
	}
	return result
}

// IsRedacted checks if a value is the redacted placeholder
func IsRedacted(value string) bool {
	return strings.TrimSpace(value) == RedactedValue
}
