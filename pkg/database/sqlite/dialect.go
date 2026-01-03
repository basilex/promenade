// Package sqlite provides SQLite dialect implementation
package sqlite

import (
	"fmt"

	"github.com/basilex/promenade/pkg/database"
)

// Dialect implements SQLite-specific SQL syntax
type Dialect struct {
	database.BaseDialect
}

// NewDialect creates a new SQLite dialect
func NewDialect() *Dialect {
	return &Dialect{
		BaseDialect: database.NewBaseDialect("sqlite"),
	}
}

// Placeholder returns SQLite parameter placeholder (?, ?, ?, ...)
func (d *Dialect) Placeholder(n int) string {
	return "?"
}

// SupportsReturning returns true (SQLite 3.35+ supports RETURNING clause)
func (d *Dialect) SupportsReturning() bool {
	return true // Requires SQLite 3.35+
}

// SupportsJSON returns false (SQLite stores JSON as TEXT, not native type)
func (d *Dialect) SupportsJSON() bool {
	return false
}

// SupportsJSONIndex returns false (SQLite doesn't support JSON-specific indexes)
func (d *Dialect) SupportsJSONIndex() bool {
	return false
}

// SupportsUUID returns false (SQLite doesn't have native UUID type, use TEXT)
func (d *Dialect) SupportsUUID() bool {
	return false
}

// QuoteIdentifier quotes identifiers with backticks
func (d *Dialect) QuoteIdentifier(name string) string {
	return "`" + name + "`"
}

// UUIDType returns TEXT for UUID storage (36 characters with hyphens)
func (d *Dialect) UUIDType() string {
	return "TEXT"
}

// JSONType returns TEXT for JSON storage
func (d *Dialect) JSONType() string {
	return "TEXT"
}

// TimestampType returns DATETIME for timestamp storage
func (d *Dialect) TimestampType() string {
	return "DATETIME"
}

// BoolType returns BOOLEAN (SQLite uses 0/1 integers internally)
func (d *Dialect) BoolType() string {
	return "BOOLEAN"
}

// JSONExtract returns SQL for extracting value from JSON using json_extract
// Example: SELECT json_extract(tags, '$.name') FROM customers
func (d *Dialect) JSONExtract(column, path string) string {
	return fmt.Sprintf("json_extract(%s, '$.%s')", d.QuoteIdentifier(column), path)
}

// JSONArrayContains returns SQL for checking if JSON array contains value
// Note: Less efficient than PostgreSQL's @> operator, filters in Go recommended
// Example: WHERE json_extract(tags, '$') LIKE '%"vip"%'
func (d *Dialect) JSONArrayContains(column, value string) string {
	return fmt.Sprintf("json_extract(%s, '$') LIKE '%%\"%s\"%%'", d.QuoteIdentifier(column), value)
}

// CreateUUIDCheckConstraint returns CHECK constraint for UUID format validation
func (d *Dialect) CreateUUIDCheckConstraint(column string) string {
	// UUID format: 8-4-4-4-12 (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
	return fmt.Sprintf(
		"CHECK(length(%s) = 36 AND %s GLOB '[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f]-[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]')",
		d.QuoteIdentifier(column),
		d.QuoteIdentifier(column),
	)
}

// CreateJSONCheckConstraint returns CHECK constraint for JSON format validation
func (d *Dialect) CreateJSONCheckConstraint(column string) string {
	return fmt.Sprintf(
		"CHECK(json_valid(%s))",
		d.QuoteIdentifier(column),
	)
}
