// Package collector gathers environment data from the local system.
// Each collector is independent and failures are isolated to allow partial snapshots.
package collector

import (
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// Collector is the interface for all environment collectors
type Collector interface {
	Collect(snap *snapshot.Snapshot) error
}

// CollectAll runs all collectors and populates the snapshot
func CollectAll(snap *snapshot.Snapshot, redact bool) error {
	collectors := []Collector{
		&SystemCollector{},
		&RuntimeCollector{},
		&EnvCollector{Redact: redact},
		&NetworkCollector{},
	}

	for _, c := range collectors {
		if err := c.Collect(snap); err != nil {
			// Log error but continue with other collectors
			// We want partial snapshots rather than failing completely
			continue
		}
	}

	return nil
}
