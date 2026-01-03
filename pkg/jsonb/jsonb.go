// Package jsonb provides PostgreSQL JSONB type wrappers for Go structs.
//
// ⚠️ DEPRECATED: Use pkg/jsonstore instead for cross-database compatibility.
//
// This package is PostgreSQL-specific (JSONB type). The new pkg/jsonstore package
// provides the same functionality but works with all databases (PostgreSQL, SQLite,
// MySQL, SQL Server) by storing JSON as TEXT.
//
// Migration guide:
//   - jsonb.Map → Use jsonstore.Field[map[string]any]
//   - jsonb.Array → Use jsonstore.Field[[]T]
//   - jsonb.JSON[T] → Use jsonstore.Field[T]
//
// This package remains for backward compatibility but will not receive new features.
// See pkg/jsonstore/README.md for migration examples.
//
// LEGACY USAGE:
//
//	type MyStruct struct {
//		Config jsonb.Map      `db:"config" json:"config"`           // JSONB object
//		Tags   jsonb.Array    `db:"tags" json:"tags"`               // JSONB array
//		Data   jsonb.JSON[T]  `db:"data" json:"data"`               // Generic typed JSONB
//	}
package jsonb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Map represents a JSONB object (map[string]any) in PostgreSQL.
// Use this for flexible key-value data like metadata, preferences, settings.
type Map map[string]any

// Value implements driver.Valuer for Map (write to DB)
func (m Map) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil // NULL in database
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner for Map (read from DB)
func (m *Map) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("jsonb.Map: failed to scan value, expected []byte")
	}

	var result map[string]any
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*m = Map(result)
	return nil
}

// Array represents a JSONB array ([]any) in PostgreSQL.
// Use this for lists of mixed types.
type Array []any

// Value implements driver.Valuer for Array (write to DB)
func (a Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil // NULL in database
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner for Array (read from DB)
func (a *Array) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("jsonb.Array: failed to scan value, expected []byte")
	}

	var result []any
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*a = Array(result)
	return nil
}

// JSON is a generic type wrapper for any Go struct/type stored as JSONB.
// Use this for strongly-typed JSONB columns (preferred over Map when structure is known).
//
// Example:
//
//	type WorkflowSchema struct {
//		States      []State      `json:"states"`
//		Transitions []Transition `json:"transitions"`
//	}
//
//	type WorkflowDefinition struct {
//		Definition jsonb.JSON[WorkflowSchema] `db:"definition" json:"definition"`
//	}
type JSON[T any] struct {
	Data  T
	Valid bool // false if NULL in database
}

// Value implements driver.Valuer for JSON (write to DB)
func (j JSON[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil // NULL in database
	}
	return json.Marshal(j.Data)
}

// Scan implements sql.Scanner for JSON (read from DB)
func (j *JSON[T]) Scan(value interface{}) error {
	if value == nil {
		j.Valid = false
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("jsonb.JSON: failed to scan value, expected []byte")
	}

	if err := json.Unmarshal(bytes, &j.Data); err != nil {
		return err
	}

	j.Valid = true
	return nil
}

// Set sets the value and marks it as valid
func (j *JSON[T]) Set(data T) {
	j.Data = data
	j.Valid = true
}

// Unset marks the value as NULL
func (j *JSON[T]) Unset() {
	var zero T
	j.Data = zero
	j.Valid = false
}

// MarshalJSON implements json.Marshaler for API responses
func (j JSON[T]) MarshalJSON() ([]byte, error) {
	if !j.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(j.Data)
}

// UnmarshalJSON implements json.Unmarshaler for API requests
func (j *JSON[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		j.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &j.Data); err != nil {
		return err
	}

	j.Valid = true
	return nil
}
