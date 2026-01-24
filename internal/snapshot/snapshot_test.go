package snapshot

import (
	"encoding/json"
	"testing"
)

func TestNew(t *testing.T) {
	snap := New()

	if snap.SchemaVersion != SchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", snap.SchemaVersion, SchemaVersion)
	}
	if snap.CollectedVia != "local" {
		t.Errorf("CollectedVia = %q, want %q", snap.CollectedVia, "local")
	}
	if snap.Timestamp == "" {
		t.Error("Timestamp should be set")
	}
	if snap.Runtime == nil {
		t.Error("Runtime map should be initialized")
	}
	if snap.Env == nil {
		t.Error("Env map should be initialized")
	}
}

func TestComputeID(t *testing.T) {
	snap := New()
	snap.Hostname = "test-host"
	snap.System.OS = "linux"

	err := snap.ComputeID()
	if err != nil {
		t.Fatalf("ComputeID() error = %v", err)
	}

	if snap.SnapshotID == "" {
		t.Error("SnapshotID should be set after ComputeID")
	}
	if len(snap.SnapshotID) != 8 {
		t.Errorf("SnapshotID length = %d, want 8", len(snap.SnapshotID))
	}
}

func TestComputeID_Deterministic(t *testing.T) {
	snap1 := &Snapshot{
		SchemaVersion: SchemaVersion,
		Hostname:      "test-host",
		Timestamp:     "2024-01-01T00:00:00Z",
		CollectedVia:  "local",
		Runtime:       make(map[string]*RuntimeInfo),
		Env:           make(map[string]string),
	}

	snap2 := &Snapshot{
		SchemaVersion: SchemaVersion,
		Hostname:      "test-host",
		Timestamp:     "2024-01-01T00:00:00Z",
		CollectedVia:  "local",
		Runtime:       make(map[string]*RuntimeInfo),
		Env:           make(map[string]string),
	}

	if err := snap1.ComputeID(); err != nil {
		t.Fatalf("ComputeID() for snap1 error = %v", err)
	}
	if err := snap2.ComputeID(); err != nil {
		t.Fatalf("ComputeID() for snap2 error = %v", err)
	}

	if snap1.SnapshotID != snap2.SnapshotID {
		t.Errorf("Same content should produce same ID: %q != %q", snap1.SnapshotID, snap2.SnapshotID)
	}
}

func TestToJSON(t *testing.T) {
	snap := New()
	snap.Hostname = "test-host"
	snap.System.OS = "linux"
	snap.Runtime["go"] = &RuntimeInfo{Version: "1.22.0", Path: "/usr/bin/go"}

	data, err := snap.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON() should return non-empty data")
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if parsed["hostname"] != "test-host" {
		t.Errorf("hostname = %v, want %q", parsed["hostname"], "test-host")
	}
}

func TestFromJSON(t *testing.T) {
	original := New()
	original.Hostname = "test-host"
	original.System.OS = "linux"
	original.System.Arch = "amd64"
	original.Runtime["go"] = &RuntimeInfo{Version: "1.22.0", Path: "/usr/bin/go"}
	original.Env["NODE_ENV"] = "development"

	data, _ := original.ToJSON()

	restored, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if restored.Hostname != original.Hostname {
		t.Errorf("Hostname = %q, want %q", restored.Hostname, original.Hostname)
	}
	if restored.System.OS != original.System.OS {
		t.Errorf("System.OS = %q, want %q", restored.System.OS, original.System.OS)
	}
	if restored.Runtime["go"] == nil {
		t.Error("Runtime[go] should exist")
	}
	if restored.Env["NODE_ENV"] != "development" {
		t.Errorf("Env[NODE_ENV] = %q, want %q", restored.Env["NODE_ENV"], "development")
	}
}

func TestFromJSON_InvalidData(t *testing.T) {
	_, err := FromJSON([]byte("invalid json"))
	if err == nil {
		t.Error("FromJSON() should return error for invalid JSON")
	}
}
