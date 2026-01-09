// Package diff compares snapshots and identifies environment discrepancies.
// Supports both pairwise comparison and N-node majority/outlier detection.
package diff

import (
	"github.com/GBerghoff/envdiff/internal/secrets"
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// Compare compares multiple snapshots and produces a Diff
func Compare(snapshots map[string]*snapshot.Snapshot) *Diff {
	d := New()

	// Build node list
	for name := range snapshots {
		d.Nodes = append(d.Nodes, name)
		d.Snapshots[name] = snapshots[name]
	}

	d.Summary.TotalNodes = len(d.Nodes)
	d.Summary.SuccessfulNodes = len(d.Nodes)

	// Initialize diff sections
	d.Diffs["system"] = make(map[string]*FieldDiff)
	d.Diffs["runtime"] = make(map[string]*FieldDiff)
	d.Diffs["env"] = make(map[string]*FieldDiff)

	// Compare system fields
	compareSystemFields(d, snapshots)

	// Compare runtime versions
	compareRuntimeFields(d, snapshots)

	// Compare environment variables
	compareEnvFields(d, snapshots)

	return d
}

func compareSystemFields(d *Diff, snapshots map[string]*snapshot.Snapshot) {
	fields := []struct {
		name   string
		getter func(*snapshot.Snapshot) any
	}{
		{"os", func(s *snapshot.Snapshot) any { return s.System.OS }},
		{"os_version", func(s *snapshot.Snapshot) any { return s.System.OSVersion }},
		{"arch", func(s *snapshot.Snapshot) any { return s.System.Arch }},
		{"kernel", func(s *snapshot.Snapshot) any { return s.System.Kernel }},
		{"cpu_cores", func(s *snapshot.Snapshot) any { return s.System.CPUCores }},
		{"memory_gb", func(s *snapshot.Snapshot) any { return s.System.MemoryGB }},
	}

	for _, f := range fields {
		values := make(map[string]any)
		for name, snap := range snapshots {
			values[name] = f.getter(snap)
		}
		d.Diffs["system"][f.name] = createFieldDiff(values, d.Nodes)
		updateSummary(d, d.Diffs["system"][f.name])
	}
}

func compareRuntimeFields(d *Diff, snapshots map[string]*snapshot.Snapshot) {
	// Gather all runtime keys across all snapshots
	allRuntimes := make(map[string]bool)
	for _, snap := range snapshots {
		for rt := range snap.Runtime {
			allRuntimes[rt] = true
		}
	}

	for rt := range allRuntimes {
		values := make(map[string]any)
		for name, snap := range snapshots {
			if info, ok := snap.Runtime[rt]; ok && info != nil {
				values[name] = info.Version
			} else {
				values[name] = nil
			}
		}
		d.Diffs["runtime"][rt] = createFieldDiff(values, d.Nodes)
		updateSummary(d, d.Diffs["runtime"][rt])
	}
}

func compareEnvFields(d *Diff, snapshots map[string]*snapshot.Snapshot) {
	// Gather all env keys across all snapshots
	allEnvs := make(map[string]bool)
	for _, snap := range snapshots {
		for k := range snap.Env {
			allEnvs[k] = true
		}
	}

	for envKey := range allEnvs {
		values := make(map[string]any)
		anyRedacted := false

		for name, snap := range snapshots {
			if val, ok := snap.Env[envKey]; ok {
				values[name] = val
				if secrets.IsRedacted(val) {
					anyRedacted = true
				}
			} else {
				values[name] = nil
			}
		}

		fd := createFieldDiff(values, d.Nodes)

		// If any value is redacted, mark the whole field as redacted
		if anyRedacted {
			fd.Status = "redacted"
			fd.Majority = nil
			fd.Outliers = nil
		}

		d.Diffs["env"][envKey] = fd
		updateSummary(d, fd)
	}
}

func createFieldDiff(values map[string]any, nodes []string) *FieldDiff {
	fd := &FieldDiff{
		Values: values,
	}

	// Check if all values are equal
	var firstVal any
	allEqual := true
	first := true

	for _, val := range values {
		if first {
			firstVal = val
			first = false
		} else if val != firstVal {
			allEqual = false
			break
		}
	}

	if allEqual {
		fd.Status = "equal"
		return fd
	}

	fd.Status = "different"

	// For N>2 nodes, calculate majority and outliers
	if len(nodes) > 2 {
		fd.Majority, fd.Outliers = calculateMajority(values, nodes)
	}

	return fd
}

func calculateMajority(values map[string]any, nodes []string) (any, []string) {
	// Count occurrences of each value
	counts := make(map[any]int)
	nodesByValue := make(map[any][]string)

	for node, val := range values {
		counts[val]++
		nodesByValue[val] = append(nodesByValue[val], node)
	}

	// Find the value with the highest count
	var maxVal any
	maxCount := 0
	for val, count := range counts {
		if count > maxCount {
			maxCount = count
			maxVal = val
		}
	}

	// Only consider it a majority if it appears more than once
	// and more than other values
	if maxCount <= 1 || maxCount <= len(nodes)/2 {
		return nil, nil
	}

	// Find outliers (nodes that don't have the majority value)
	var outliers []string
	for node, val := range values {
		if val != maxVal {
			outliers = append(outliers, node)
		}
	}

	return maxVal, outliers
}

func updateSummary(d *Diff, fd *FieldDiff) {
	d.Summary.TotalFields++
	switch fd.Status {
	case "equal":
		d.Summary.Equal++
	case "different":
		d.Summary.Different++
	case "redacted":
		d.Summary.Redacted++
	}
}
