package secrets

import "testing"

func TestIsSecret(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		expected bool
	}{
		{"password detected", "DB_PASSWORD", true},
		{"token detected", "GITHUB_TOKEN", true},
		{"api key detected", "API_KEY", true},
		{"secret detected", "AWS_SECRET_ACCESS_KEY", true},
		{"auth detected", "AUTH_HEADER", true},
		{"jwt detected", "JWT_SECRET", true},
		{"safe variable", "NODE_ENV", false},
		{"safe variable", "HOME", false},
		{"safe variable", "PATH", false},
		{"case insensitive", "password", true},
		{"case insensitive", "Password", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSecret(tt.envName)
			if got != tt.expected {
				t.Errorf("IsSecret(%q) = %v, want %v", tt.envName, got, tt.expected)
			}
		})
	}
}

func TestRedactEnv(t *testing.T) {
	input := map[string]string{
		"HOME":        "/home/user",
		"DB_PASSWORD": "secret123",
		"API_KEY":     "key456",
		"NODE_ENV":    "development",
	}

	result := RedactEnv(input)

	if result["HOME"] != "/home/user" {
		t.Errorf("HOME should not be redacted")
	}
	if result["NODE_ENV"] != "development" {
		t.Errorf("NODE_ENV should not be redacted")
	}
	if result["DB_PASSWORD"] != RedactedValue {
		t.Errorf("DB_PASSWORD should be redacted, got %q", result["DB_PASSWORD"])
	}
	if result["API_KEY"] != RedactedValue {
		t.Errorf("API_KEY should be redacted, got %q", result["API_KEY"])
	}
}

func TestIsRedacted(t *testing.T) {
	if !IsRedacted(RedactedValue) {
		t.Error("RedactedValue should be detected as redacted")
	}
	if !IsRedacted("  [REDACTED]  ") {
		t.Error("RedactedValue with whitespace should be detected")
	}
	if IsRedacted("actual-value") {
		t.Error("actual-value should not be detected as redacted")
	}
}
