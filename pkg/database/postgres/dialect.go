// Package postgres provides PostgreSQL dialect implementation
package postgres

import (
	"fmt"

	"github.com/basilex/promenade/pkg/database"
)

// Dialect implements PostgreSQL-specific SQL syntax
type Dialect struct {
	database.BaseDialect
}

// NewDialect creates a new PostgreSQL dialect
func NewDialect() *Dialect {
	return &Dialect{
		BaseDialect: database.NewBaseDialect("postgres"),
	}
}

// Placeholder returns PostgreSQL parameter placeholder ($1, $2, $3, ...)
func (d *Dialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

// SupportsReturning returns true (PostgreSQL supports RETURNING clause)
func (d *Dialect) SupportsReturning() bool {
	return true
}

// SupportsJSON returns true (PostgreSQL has native JSONB type)
func (d *Dialect) SupportsJSON() bool {
	return true
}

// SupportsJSONIndex returns true (PostgreSQL supports GIN indexes on JSONB)
func (d *Dialect) SupportsJSONIndex() bool {
	return true
}

// SupportsUUID returns true (PostgreSQL has native UUID type)
func (d *Dialect) SupportsUUID() bool {
	return true
}

// QuoteIdentifier quotes identifiers with double quotes
func (d *Dialect) QuoteIdentifier(name string) string {
	return `"` + name + `"`
}

// UUIDType returns PostgreSQL UUID type
func (d *Dialect) UUIDType() string {
	return "UUID"
}

// JSONType returns PostgreSQL JSONB type (preferred over JSON)
func (d *Dialect) JSONType() string {
	return "JSONB"
}

// TimestampType returns PostgreSQL TIMESTAMP type
func (d *Dialect) TimestampType() string {
	return "TIMESTAMP"
}

// BoolType returns PostgreSQL BOOLEAN type
func (d *Dialect) BoolType() string {
	return "BOOLEAN"
}

// CreateJSONIndex generates GIN index SQL for JSONB column
func (d *Dialect) CreateJSONIndex(table, column, indexName string) string {
	return fmt.Sprintf(
		"CREATE INDEX %s ON %s USING gin(%s)",
		d.QuoteIdentifier(indexName),
		d.QuoteIdentifier(table),
		d.QuoteIdentifier(column),
	)
}

// JSONContains returns SQL for checking if JSONB contains value
// Example: WHERE tags @> '["vip"]'::jsonb
func (d *Dialect) JSONContains(column, value string) string {
	return fmt.Sprintf("%s @> '%s'::jsonb", d.QuoteIdentifier(column), value)
}

// JSONExtract returns SQL for extracting value from JSONB
// Example: SELECT tags->>'name' FROM customers
func (d *Dialect) JSONExtract(column, path string) string {
	return fmt.Sprintf("%s->>'%s'", d.QuoteIdentifier(column), path)
}
