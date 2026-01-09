package collector

import (
	"os"
	"strings"

	"github.com/GBerghoff/envdiff/internal/secrets"
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// EnvCollector gathers environment variables
type EnvCollector struct {
	Redact bool
}

// Collect gathers environment variables
func (c *EnvCollector) Collect(s *snapshot.Snapshot) error {
	env := make(map[string]string)

	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	if c.Redact {
		s.Env = secrets.RedactEnv(env)
	} else {
		s.Env = env
	}

	return nil
}
