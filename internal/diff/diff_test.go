package diff

import (
	"encoding/json"
	"testing"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

func TestNew(t *testing.T) {
	result := New()

	if result.SchemaVersion != snapshot.SchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", result.SchemaVersion, snapshot.SchemaVersion)
	}
	if result.GeneratedAt == "" {
		t.Error("GeneratedAt should be set")
	}
	if result.Nodes == nil {
		t.Error("Nodes should be initialized")
	}
	if result.Errors == nil {
		t.Error("Errors map should be initialized")
	}
	if result.Diffs == nil {
		t.Error("Diffs map should be initialized")
	}
	if result.Snapshots == nil {
		t.Error("Snapshots map should be initialized")
	}
}

func TestToJSON(t *testing.T) {
	result := New()
	result.Nodes = []string{"node1", "node2"}
	result.Summary.TotalNodes = 2

	data, err := result.ToJSON()
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
}

func TestFromJSON(t *testing.T) {
	original := New()
	original.Nodes = []string{"local", "ci"}
	original.Summary.TotalNodes = 2
	original.Summary.Different = 3

	data, _ := original.ToJSON()

	restored, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if len(restored.Nodes) != 2 {
		t.Errorf("len(Nodes) = %d, want 2", len(restored.Nodes))
	}
	if restored.Summary.Different != 3 {
		t.Errorf("Summary.Different = %d, want 3", restored.Summary.Different)
	}
}

func TestFromJSON_InvalidData(t *testing.T) {
	_, err := FromJSON([]byte("not valid json"))
	if err == nil {
		t.Error("FromJSON() should return error for invalid JSON")
	}
}
