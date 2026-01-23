package diff

// FieldStatus represents the comparison status of a field.
type FieldStatus string

// Field status constants for diff comparisons.
const (
	StatusEqual     FieldStatus = "equal"
	StatusDifferent FieldStatus = "different"
	StatusRedacted  FieldStatus = "redacted"
)
