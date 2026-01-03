// Package jsonstore provides database-agnostic JSON storage.
// Works with PostgreSQL JSONB, SQLite TEXT, MySQL JSON, and SQL Server NVARCHAR.
package jsonstore

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Field represents a JSON-serializable field that can be stored in any database.
// It implements sql.Scanner and driver.Valuer for seamless database integration.
//
// Supports:
//   - PostgreSQL: JSONB (native, optimized)
//   - SQLite: TEXT (JSON string)
//   - MySQL: JSON (native)
//   - SQL Server: NVARCHAR(MAX) (JSON string)
//
// Example usage:
//
//	type Customer struct {
//	    Tags      Field[[]string]       // String array
//	    Metadata  Field[map[string]any] // JSON object
//	}
//
//	customer := Customer{
//	    Tags: NewField([]string{"vip", "premium"}),
//	    Metadata: NewField(map[string]any{"score": 95}),
//	}
type Field[T any] struct {
	value T
	null  bool // true if database value was NULL
}

// NewField creates a new JSON field with a value
func NewField[T any](value T) Field[T] {
	return Field[T]{
		value: value,
		null:  false,
	}
}

// NewNullField creates a new JSON field representing NULL
func NewNullField[T any]() Field[T] {
	var zero T
	return Field[T]{
		value: zero,
		null:  true,
	}
}

// Get returns the stored value
func (f Field[T]) Get() T {
	return f.value
}

// Set updates the stored value
func (f *Field[T]) Set(value T) {
	f.value = value
	f.null = false
}

// IsNull returns true if the field represents a NULL database value
func (f Field[T]) IsNull() bool {
	return f.null
}

// SetNull marks the field as NULL
func (f *Field[T]) SetNull() {
	var zero T
	f.value = zero
	f.null = true
}

// MarshalJSON implements json.Marshaler for JSON serialization
func (f Field[T]) MarshalJSON() ([]byte, error) {
	if f.null {
		return []byte("null"), nil
	}
	return json.Marshal(f.value)
}

// UnmarshalJSON implements json.Unmarshaler for JSON deserialization
func (f *Field[T]) UnmarshalJSON(data []byte) error {
	// Check for null
	if string(data) == "null" {
		f.SetNull()
		return nil
	}

	f.null = false
	return json.Unmarshal(data, &f.value)
}

// Scan implements sql.Scanner interface for reading from database
// Supports both []byte and string values from different drivers
func (f *Field[T]) Scan(value interface{}) error {
	if value == nil {
		f.SetNull()
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("jsonstore: cannot scan type %T into Field[%T]", value, f.value)
	}

	// Empty string is treated as NULL
	if len(bytes) == 0 {
		f.SetNull()
		return nil
	}

	f.null = false
	if err := json.Unmarshal(bytes, &f.value); err != nil {
		return fmt.Errorf("jsonstore: failed to unmarshal JSON: %w", err)
	}

	return nil
}

// Value implements driver.Valuer interface for writing to database
// Returns JSON string for storage in any database
func (f Field[T]) Value() (driver.Value, error) {
	if f.null {
		return nil, nil
	}

	bytes, err := json.Marshal(f.value)
	if err != nil {
		return nil, fmt.Errorf("jsonstore: failed to marshal JSON: %w", err)
	}

	// Return as string for maximum compatibility
	// PostgreSQL will convert to JSONB automatically
	// SQLite/MySQL store as TEXT/JSON directly
	return string(bytes), nil
}

// String returns a human-readable representation
func (f Field[T]) String() string {
	if f.null {
		return "NULL"
	}
	bytes, _ := json.Marshal(f.value)
	return string(bytes)
}

// Equal compares two Field values
func (f Field[T]) Equal(other Field[T]) bool {
	if f.null != other.null {
		return false
	}
	if f.null {
		return true // Both NULL
	}

	// Compare JSON representations
	a, _ := json.Marshal(f.value)
	b, _ := json.Marshal(other.value)
	return string(a) == string(b)
}

// Clone creates a deep copy of the field
func (f Field[T]) Clone() Field[T] {
	if f.null {
		return NewNullField[T]()
	}

	// Deep copy via JSON serialization
	bytes, _ := json.Marshal(f.value)
	var newValue T
	_ = json.Unmarshal(bytes, &newValue) // Error ignored - best-effort clone

	return NewField(newValue)
}
