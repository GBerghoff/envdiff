// Package render formats snapshots and diffs for human consumption.
// Implementations handle terminal output and markdown export.
package render

import (
	"github.com/gberghoff/envdiff/internal/diff"
	"github.com/gberghoff/envdiff/internal/snapshot"
)

// Renderer is the interface for output renderers
type Renderer interface {
	RenderSnapshot(s *snapshot.Snapshot) string
	RenderDiff(d *diff.Diff) string
}
