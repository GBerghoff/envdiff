package collector

import (
	"testing"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

func TestCollectAll_PopulatesSnapshot(t *testing.T) {
	snap := snapshot.New()

	err := CollectAll(snap, false)
	if err != nil {
		t.Fatalf("CollectAll() error = %v", err)
	}

	if snap.System.OS == "" {
		t.Error("System.OS should be populated")
	}
	if snap.System.Arch == "" {
		t.Error("System.Arch should be populated")
	}
	if len(snap.Env) == 0 {
		t.Error("Env should be populated with environment variables")
	}
}

func TestCollectAll_WithRedaction(t *testing.T) {
	snap := snapshot.New()

	err := CollectAll(snap, true)
	if err != nil {
		t.Fatalf("CollectAll() error = %v", err)
	}

	if len(snap.Env) == 0 {
		t.Error("Env should be populated even with redaction enabled")
	}
}

func TestSystemCollector_Collect(t *testing.T) {
	snap := snapshot.New()
	collector := &SystemCollector{}

	err := collector.Collect(snap)
	if err != nil {
		t.Fatalf("SystemCollector.Collect() error = %v", err)
	}

	if snap.System.OS == "" {
		t.Error("OS should be populated")
	}
	if snap.System.Arch == "" {
		t.Error("Arch should be populated")
	}
	if snap.System.CPUCores <= 0 {
		t.Error("CPUCores should be positive")
	}
	if snap.System.MemoryGB <= 0 {
		t.Error("MemoryGB should be positive")
	}
	if snap.Hostname == "" {
		t.Error("Hostname should be populated")
	}
}

func TestEnvCollector_Collect(t *testing.T) {
	snap := snapshot.New()
	collector := &EnvCollector{Redact: false}

	err := collector.Collect(snap)
	if err != nil {
		t.Fatalf("EnvCollector.Collect() error = %v", err)
	}

	if len(snap.Env) == 0 {
		t.Error("Env should be populated")
	}
	if _, hasPath := snap.Env["PATH"]; !hasPath {
		t.Error("PATH environment variable should be present")
	}
}

func TestEnvCollector_CollectWithRedaction(t *testing.T) {
	snap := snapshot.New()
	collector := &EnvCollector{Redact: true}

	err := collector.Collect(snap)
	if err != nil {
		t.Fatalf("EnvCollector.Collect() error = %v", err)
	}

	if len(snap.Env) == 0 {
		t.Error("Env should be populated even with redaction")
	}
}

func TestRuntimeCollector_Collect(t *testing.T) {
	snap := snapshot.New()
	collector := &RuntimeCollector{}

	err := collector.Collect(snap)
	if err != nil {
		t.Fatalf("RuntimeCollector.Collect() error = %v", err)
	}

	if snap.Runtime == nil {
		t.Error("Runtime map should be initialized")
	}
}

func TestNetworkCollector_Collect(t *testing.T) {
	snap := snapshot.New()
	collector := &NetworkCollector{}

	err := collector.Collect(snap)
	if err != nil {
		t.Fatalf("NetworkCollector.Collect() error = %v", err)
	}

	if snap.Network == nil {
		t.Error("Network should be populated")
	}
}
