// Package config handles parsing and generation of envdiff.yaml files.
// The config schema defines runtime constraints, required env vars, and remediation hints.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the envdiff.yaml configuration file
type Config struct {
	Runtime  map[string]string `yaml:"runtime"`
	Env      EnvConfig         `yaml:"env"`
	Packages []string          `yaml:"packages,omitempty"`
	Fix      map[string]FixConfig `yaml:"fix,omitempty"`
}

// EnvConfig holds environment variable requirements
type EnvConfig struct {
	Required []string          `yaml:"required,omitempty"`
	Expected map[string]string `yaml:"expected,omitempty"`
	Ignore   []string          `yaml:"ignore,omitempty"`
}

// FixConfig holds remediation hints
type FixConfig struct {
	Missing      string `yaml:"missing,omitempty"`
	WrongVersion string `yaml:"wrong_version,omitempty"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		Runtime: map[string]string{
			"go":     ">= 1.21.0",
			"node":   ">= 18.0.0",
			"python": ">= 3.10.0",
			"docker": "*",
		},
		Env: EnvConfig{
			Required: []string{},
			Expected: map[string]string{},
			Ignore: []string{
				"TERM",
				"SHELL",
				"HOME",
				"USER",
				"LOGNAME",
				"PATH",
				"PWD",
				"OLDPWD",
				"LANG",
				"LC_*",
				"_",
				"SHLVL",
				"*_SESSION*",
				"*_TOKEN*",
			},
		},
		Packages: []string{},
		Fix:      map[string]FixConfig{},
	}
}

// ToYAML serializes the config to YAML
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// Template returns a commented YAML template
func Template() string {
	return `# envdiff.yaml - Environment requirements
# Run 'envdiff check' to validate your environment against these requirements

# Runtime version constraints (semver)
# Supported operators: =, !=, >, <, >=, <=, ~> (pessimistic), ^ (caret), *
runtime:
  go: ">= 1.21.0"
  node: ">= 18.0.0"
  python: ">= 3.10.0"
  docker: "*"  # any version, just needs to exist
  # kubectl: "~> 1.28.0"  # >= 1.28.0 and < 1.29.0

# Environment variable requirements
env:
  # Variables that must be set (value doesn't matter)
  required:
    # - DATABASE_URL
    # - AWS_PROFILE

  # Variables that must have specific values
  expected:
    # NODE_ENV: development
    # LOG_LEVEL: debug

  # Variables to ignore when comparing (glob patterns supported)
  ignore:
    - TERM
    - SHELL
    - HOME
    - USER
    - PATH
    - PWD
    - LANG
    - "LC_*"
    - "*_SESSION*"
    - "*_TOKEN*"

# Packages to verify (opt-in, not checked by default)
# packages:
#   - nginx
#   - redis-server

# Remediation hints (shown when check fails)
# fix:
#   node:
#     missing: "brew install node@20"
#     wrong_version: "nvm use 20"
#   DATABASE_URL:
#     missing: "Copy from 1Password vault 'Dev Secrets'"
`
}
