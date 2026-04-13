package collector

import (
	"regexp"
	"testing"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

func TestRegisterRuntime(t *testing.T) {
	def := RuntimeDefinition{
		Name:      "test-runtime",
		Command:   "echo",
		Args:      []string{"v1.2.3"},
		VersionRE: regexp.MustCompile(`v(\d+\.\d+\.\d+)`),
	}

	RegisterRuntime(def)

	if _, ok := Registry["test-runtime"]; !ok {
		t.Error("test-runtime should be in Registry")
	}

	snap := snapshot.New()
	collector := &RuntimeCollector{
		Definitions: []RuntimeDefinition{def},
	}

	err := collector.Collect(snap)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	info, ok := snap.Runtime["test-runtime"]
	if !ok {
		t.Fatal("test-runtime should be in snapshot")
	}

	if info.Version != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", info.Version)
	}
}
