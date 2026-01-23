// Package diff compares snapshots and identifies environment discrepancies.
// Supports both pairwise comparison and N-node majority/outlier detection.
package diff

import (
	"github.com/GBerghoff/envdiff/internal/secrets"
	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// Compare compares multiple snapshots and produces a Diff
func Compare(snapshots map[string]*snapshot.Snapshot) *Diff {
	result := New()

	// Build node list
	for name := range snapshots {
		result.Nodes = append(result.Nodes, name)
		result.Snapshots[name] = snapshots[name]
	}

	result.Summary.TotalNodes = len(result.Nodes)
	result.Summary.SuccessfulNodes = len(result.Nodes)

	// Initialize diff sections
	result.Diffs["system"] = make(map[string]*FieldDiff)
	result.Diffs["runtime"] = make(map[string]*FieldDiff)
	result.Diffs["env"] = make(map[string]*FieldDiff)

	// Compare system fields
	compareSystemFields(result, snapshots)

	// Compare runtime versions
	compareRuntimeFields(result, snapshots)

	// Compare environment variables
	compareEnvFields(result, snapshots)

	return result
}

func compareSystemFields(result *Diff, snapshots map[string]*snapshot.Snapshot) {
	fields := []struct {
		name   string
		getter func(*snapshot.Snapshot) any
	}{
		{"os", func(snap *snapshot.Snapshot) any { return snap.System.OS }},
		{"os_version", func(snap *snapshot.Snapshot) any { return snap.System.OSVersion }},
		{"arch", func(snap *snapshot.Snapshot) any { return snap.System.Arch }},
		{"kernel", func(snap *snapshot.Snapshot) any { return snap.System.Kernel }},
		{"cpu_cores", func(snap *snapshot.Snapshot) any { return snap.System.CPUCores }},
		{"memory_gb", func(snap *snapshot.Snapshot) any { return snap.System.MemoryGB }},
	}

	for _, f := range fields {
		values := make(map[string]any)
		for name, snap := range snapshots {
			values[name] = f.getter(snap)
		}
		result.Diffs["system"][f.name] = createFieldDiff(values, result.Nodes)
		updateSummary(result, result.Diffs["system"][f.name])
	}
}

func compareRuntimeFields(result *Diff, snapshots map[string]*snapshot.Snapshot) {
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
		result.Diffs["runtime"][rt] = createFieldDiff(values, result.Nodes)
		updateSummary(result, result.Diffs["runtime"][rt])
	}
}

func compareEnvFields(result *Diff, snapshots map[string]*snapshot.Snapshot) {
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

		fieldDiff := createFieldDiff(values, result.Nodes)

		// If any value is redacted, mark the whole field as redacted
		if anyRedacted {
			fieldDiff.Status = StatusRedacted
			fieldDiff.Majority = nil
			fieldDiff.Outliers = nil
		}

		result.Diffs["env"][envKey] = fieldDiff
		updateSummary(result, fieldDiff)
	}
}

func createFieldDiff(values map[string]any, nodes []string) *FieldDiff {
	fieldDiff := &FieldDiff{
		NodeValues: values,
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
		fieldDiff.Status = StatusEqual
		return fieldDiff
	}

	fieldDiff.Status = StatusDifferent

	// For N>2 nodes, calculate majority and outliers
	if len(nodes) > 2 {
		fieldDiff.Majority, fieldDiff.Outliers = calculateMajority(values, nodes)
	}

	return fieldDiff
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

func updateSummary(result *Diff, fieldDiff *FieldDiff) {
	result.Summary.TotalFields++
	switch fieldDiff.Status {
	case StatusEqual:
		result.Summary.Equal++
	case StatusDifferent:
		result.Summary.Different++
	case StatusRedacted:
		result.Summary.Redacted++
	}
}
