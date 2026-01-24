package diff

import (
	"testing"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

func TestCompare_TwoIdenticalSnapshots(t *testing.T) {
	snap1 := &snapshot.Snapshot{
		System: snapshot.SystemInfo{
			OS:   "linux",
			Arch: "amd64",
		},
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go": {Version: "1.22.0", Path: "/usr/bin/go"},
		},
		Env: map[string]string{
			"NODE_ENV": "development",
		},
	}

	snap2 := &snapshot.Snapshot{
		System: snapshot.SystemInfo{
			OS:   "linux",
			Arch: "amd64",
		},
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go": {Version: "1.22.0", Path: "/usr/bin/go"},
		},
		Env: map[string]string{
			"NODE_ENV": "development",
		},
	}

	snapshots := map[string]*snapshot.Snapshot{
		"local": snap1,
		"ci":    snap2,
	}

	result := Compare(snapshots)

	if result.Summary.Different != 0 {
		t.Errorf("Different = %d, want 0 for identical snapshots", result.Summary.Different)
	}
	if result.Summary.Equal == 0 {
		t.Error("Equal should be > 0 for identical snapshots")
	}
}

func TestCompare_TwoDifferentSnapshots(t *testing.T) {
	snap1 := &snapshot.Snapshot{
		System: snapshot.SystemInfo{
			OS:   "linux",
			Arch: "amd64",
		},
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go": {Version: "1.22.0", Path: "/usr/bin/go"},
		},
		Env: map[string]string{
			"NODE_ENV": "development",
		},
	}

	snap2 := &snapshot.Snapshot{
		System: snapshot.SystemInfo{
			OS:   "darwin",
			Arch: "arm64",
		},
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go": {Version: "1.21.0", Path: "/opt/go/bin/go"},
		},
		Env: map[string]string{
			"NODE_ENV": "production",
		},
	}

	snapshots := map[string]*snapshot.Snapshot{
		"local": snap1,
		"ci":    snap2,
	}

	result := Compare(snapshots)

	if result.Summary.Different == 0 {
		t.Error("Different should be > 0 for different snapshots")
	}
	if result.Diffs["system"]["os"].Status != StatusDifferent {
		t.Error("OS field should be marked as different")
	}
	if result.Diffs["runtime"]["go"].Status != StatusDifferent {
		t.Error("go runtime should be marked as different")
	}
	if result.Diffs["env"]["NODE_ENV"].Status != StatusDifferent {
		t.Error("NODE_ENV should be marked as different")
	}
}

func TestCompare_MissingRuntime(t *testing.T) {
	snap1 := &snapshot.Snapshot{
		System:  snapshot.SystemInfo{},
		Runtime: map[string]*snapshot.RuntimeInfo{
			"go": {Version: "1.22.0", Path: "/usr/bin/go"},
		},
		Env: map[string]string{},
	}

	snap2 := &snapshot.Snapshot{
		System:  snapshot.SystemInfo{},
		Runtime: map[string]*snapshot.RuntimeInfo{},
		Env:     map[string]string{},
	}

	snapshots := map[string]*snapshot.Snapshot{
		"local": snap1,
		"ci":    snap2,
	}

	result := Compare(snapshots)

	goDiff := result.Diffs["runtime"]["go"]
	if goDiff.Status != StatusDifferent {
		t.Errorf("go runtime Status = %q, want %q", goDiff.Status, StatusDifferent)
	}
	if goDiff.NodeValues["ci"] != nil {
		t.Error("ci should have nil value for missing go runtime")
	}
}

func TestCompare_RedactedEnvVars(t *testing.T) {
	snap1 := &snapshot.Snapshot{
		System: snapshot.SystemInfo{},
		Runtime: map[string]*snapshot.RuntimeInfo{},
		Env: map[string]string{
			"API_KEY": "[REDACTED]",
		},
	}

	snap2 := &snapshot.Snapshot{
		System: snapshot.SystemInfo{},
		Runtime: map[string]*snapshot.RuntimeInfo{},
		Env: map[string]string{
			"API_KEY": "[REDACTED]",
		},
	}

	snapshots := map[string]*snapshot.Snapshot{
		"local": snap1,
		"ci":    snap2,
	}

	result := Compare(snapshots)

	apiKeyDiff := result.Diffs["env"]["API_KEY"]
	if apiKeyDiff.Status != StatusRedacted {
		t.Errorf("API_KEY Status = %q, want %q", apiKeyDiff.Status, StatusRedacted)
	}
	if result.Summary.Redacted != 1 {
		t.Errorf("Summary.Redacted = %d, want 1", result.Summary.Redacted)
	}
}

func TestCalculateMajority_ClearMajority(t *testing.T) {
	values := map[string]any{
		"node1": "1.22.0",
		"node2": "1.22.0",
		"node3": "1.21.0",
	}
	nodes := []string{"node1", "node2", "node3"}

	majority, outliers := calculateMajority(values, nodes)

	if majority != "1.22.0" {
		t.Errorf("majority = %v, want 1.22.0", majority)
	}
	if len(outliers) != 1 || outliers[0] != "node3" {
		t.Errorf("outliers = %v, want [node3]", outliers)
	}
}

func TestCalculateMajority_NoMajority(t *testing.T) {
	values := map[string]any{
		"node1": "1.22.0",
		"node2": "1.21.0",
		"node3": "1.20.0",
	}
	nodes := []string{"node1", "node2", "node3"}

	majority, outliers := calculateMajority(values, nodes)

	if majority != nil {
		t.Errorf("majority = %v, want nil when no clear majority", majority)
	}
	if outliers != nil {
		t.Errorf("outliers = %v, want nil when no clear majority", outliers)
	}
}

func TestCalculateMajority_TiedValues(t *testing.T) {
	values := map[string]any{
		"node1": "1.22.0",
		"node2": "1.22.0",
		"node3": "1.21.0",
		"node4": "1.21.0",
	}
	nodes := []string{"node1", "node2", "node3", "node4"}

	majority, _ := calculateMajority(values, nodes)

	if majority != nil {
		t.Errorf("majority = %v, want nil for tied values", majority)
	}
}
