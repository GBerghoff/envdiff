package diff

import (
	"encoding/json"
	"time"

	"github.com/GBerghoff/envdiff/internal/snapshot"
)

// FieldDiff represents the diff status for a single field
type FieldDiff struct {
	Status     FieldStatus    `json:"status"`
	NodeValues map[string]any `json:"values"` // values keyed by node name
	Majority   any            `json:"majority,omitempty"`
	Outliers   []string       `json:"outliers,omitempty"`
}

// Summary contains diff statistics
type Summary struct {
	TotalNodes      int `json:"total_nodes"`
	SuccessfulNodes int `json:"successful_nodes"`
	FailedNodes     int `json:"failed_nodes"`
	TotalFields     int `json:"total_fields"`
	Equal           int `json:"equal"`
	Different       int `json:"different"`
	Redacted        int `json:"redacted"`
}

// Diff represents a comparison between multiple snapshots
type Diff struct {
	SchemaVersion string                        `json:"schema_version"`
	GeneratedAt   string                        `json:"generated_at"`
	Nodes         []string                      `json:"nodes"`
	Errors        map[string]string             `json:"errors"`
	Summary       Summary                       `json:"summary"`
	Diffs         map[string]map[string]*FieldDiff `json:"diffs"`
	Snapshots     map[string]*snapshot.Snapshot `json:"snapshots"`
}

// New creates a new Diff
func New() *Diff {
	return &Diff{
		SchemaVersion: snapshot.SchemaVersion,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		Nodes:         []string{},
		Errors:        make(map[string]string),
		Diffs:         make(map[string]map[string]*FieldDiff),
		Snapshots:     make(map[string]*snapshot.Snapshot),
	}
}

// ToJSON serializes the diff to JSON
func (d *Diff) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// FromJSON deserializes a diff from JSON
func FromJSON(data []byte) (*Diff, error) {
	var d Diff
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &d, nil
}
