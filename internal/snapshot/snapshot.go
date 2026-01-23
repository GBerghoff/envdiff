// Package snapshot defines the core data model for environment captures.
// Snapshots are content-addressable via their ID for reliable comparison.
package snapshot

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// SchemaVersion is the current schema version for snapshots and diffs.
const SchemaVersion = "1"

// RuntimeInfo holds version and path for a single runtime/CLI tool
type RuntimeInfo struct {
	Version string `json:"version"`
	Path    string `json:"path"`
}

// SystemInfo contains OS and hardware information
type SystemInfo struct {
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	Arch      string `json:"arch"`
	Kernel    string `json:"kernel"`
	CPUCores  int    `json:"cpu_cores"`
	MemoryGB  int    `json:"memory_gb"`
	Hostname  string `json:"hostname"`
}

// PackageInfo contains package manager and installed packages
type PackageInfo struct {
	Manager string            `json:"manager"`
	Items   map[string]string `json:"items"`
}

// NetworkInfo contains network-related information
type NetworkInfo struct {
	Hosts          map[string]string `json:"hosts"`
	ListeningPorts []int             `json:"listening_ports"`
}

// Snapshot represents a complete environment snapshot
type Snapshot struct {
	SchemaVersion string                   `json:"schema_version"`
	SnapshotID    string                   `json:"snapshot_id"`
	Timestamp     string                   `json:"timestamp"`
	Hostname      string                   `json:"hostname"`
	CollectedVia  string                   `json:"collected_via"`
	System        SystemInfo               `json:"system"`
	Runtime       map[string]*RuntimeInfo  `json:"runtime"`
	Env           map[string]string        `json:"env"`
	Packages      *PackageInfo             `json:"packages,omitempty"`
	Network       *NetworkInfo             `json:"network,omitempty"`
}

// New creates a new Snapshot with default values
func New() *Snapshot {
	return &Snapshot{
		SchemaVersion: SchemaVersion,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		CollectedVia:  "local",
		Runtime:       make(map[string]*RuntimeInfo),
		Env:           make(map[string]string),
	}
}

// ComputeID generates the snapshot_id from content hash
func (s *Snapshot) ComputeID() error {
	// Temporarily clear the ID to get consistent hash
	s.SnapshotID = ""
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to compute snapshot ID: %w", err)
	}
	hash := sha256.Sum256(data)
	s.SnapshotID = fmt.Sprintf("%x", hash[:4]) // First 8 hex chars
	return nil
}

// ToJSON serializes the snapshot to JSON
func (s *Snapshot) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON deserializes a snapshot from JSON
func FromJSON(data []byte) (*Snapshot, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
