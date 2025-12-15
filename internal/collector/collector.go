// Package collector gathers environment data from the local system.
// Each collector is independent and failures are isolated to allow partial snapshots.
package collector

import (
	"github.com/gberghoff/envdiff/internal/snapshot"
)

// Collector is the interface for all environment collectors
type Collector interface {
	Collect(s *snapshot.Snapshot) error
}

// CollectAll runs all collectors and populates the snapshot
func CollectAll(s *snapshot.Snapshot, redact bool) error {
	collectors := []Collector{
		&SystemCollector{},
		&RuntimeCollector{},
		&EnvCollector{Redact: redact},
		&NetworkCollector{},
	}

	for _, c := range collectors {
		if err := c.Collect(s); err != nil {
			// Log error but continue with other collectors
			// We want partial snapshots rather than failing completely
			continue
		}
	}

	return nil
}
