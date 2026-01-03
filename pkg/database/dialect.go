// Package database provides database adapter abstraction for multi-database support.
// Supports PostgreSQL, SQLite, MySQL, and future databases.
package database

import "fmt"

// Dialect abstracts SQL syntax differences between databases.
// Each database has its own dialect implementation (Postgres, SQLite, MySQL, etc.)
type Dialect interface {
	// Name returns the database name (e.g., "postgres", "sqlite", "mysql")
	Name() string

	// Placeholder returns the parameter placeholder for position n (1-indexed)
	// Postgres: $1, $2, $3
	// SQLite/MySQL: ?, ?, ?
	// SQL Server: @p1, @p2, @p3
	Placeholder(n int) string

	// SupportsReturning returns true if database supports RETURNING clause
	// Postgres: true, SQLite 3.35+: true, MySQL < 8.0: false
	SupportsReturning() bool

	// SupportsJSON returns true if database has native JSON type
	// Postgres: true (JSONB), MySQL 5.7+: true (JSON), SQLite: false (use TEXT)
	SupportsJSON() bool

	// SupportsJSONIndex returns true if database can index JSON fields
	// Postgres: true (GIN index), MySQL: true (generated columns), SQLite: false
	SupportsJSONIndex() bool

	// SupportsUUID returns true if database has native UUID type
	// Postgres: true, SQLite: false (use TEXT), MySQL: false (use CHAR(36))
	SupportsUUID() bool

	// QuoteIdentifier quotes table/column names for safe SQL
	// Postgres: "table_name", MySQL: `table_name`, SQL Server: [table_name]
	QuoteIdentifier(name string) string

	// UUIDType returns the SQL type for storing UUIDs
	// Postgres: "UUID", SQLite: "TEXT", MySQL: "CHAR(36)"
	UUIDType() string

	// JSONType returns the SQL type for storing JSON
	// Postgres: "JSONB", MySQL: "JSON", SQLite: "TEXT"
	JSONType() string

	// TimestampType returns the SQL type for timestamps
	// Postgres: "TIMESTAMP", MySQL: "DATETIME", SQLite: "DATETIME"
	TimestampType() string

	// BoolType returns the SQL type for booleans
	// Postgres: "BOOLEAN", MySQL: "TINYINT(1)", SQLite: "BOOLEAN"
	BoolType() string
}

// BaseDialect provides common functionality for all dialects
type BaseDialect struct {
	name string
}

// NewBaseDialect creates a new base dialect
func NewBaseDialect(name string) BaseDialect {
	return BaseDialect{name: name}
}

// Name returns the database name
func (d BaseDialect) Name() string {
	return d.name
}

// ConvertPlaceholders converts $1, $2, $3 style placeholders to database-specific format
// Useful when writing queries in Postgres style and converting for other databases
func ConvertPlaceholders(query string, dialect Dialect) string {
	if dialect.Name() == "postgres" {
		return query // Already in correct format
	}

	// Simple implementation for ? placeholders (SQLite, MySQL)
	// For production, use a proper SQL parser
	result := ""
	paramNum := 1
	inString := false
	escaped := false

	for i := 0; i < len(query); i++ {
		char := query[i]

		// Handle string literals
		if char == '\'' && !escaped {
			inString = !inString
		}
		escaped = char == '\\' && !escaped

		// Replace $N with dialect-specific placeholder
		if !inString && char == '$' && i+1 < len(query) {
			// Find the number
			numStart := i + 1
			numEnd := numStart
			for numEnd < len(query) && query[numEnd] >= '0' && query[numEnd] <= '9' {
				numEnd++
			}

			if numEnd > numStart {
				// Found a number
				result += dialect.Placeholder(paramNum)
				paramNum++
				i = numEnd - 1
				continue
			}
		}

		result += string(char)
	}

	return result
}

// QuoteString safely quotes a string value for SQL (escapes single quotes)
func QuoteString(s string) string {
	return "'" + escapeString(s) + "'"
}

func escapeString(s string) string {
	result := ""
	for _, char := range s {
		if char == '\'' {
			result += "''"
		} else {
			result += string(char)
		}
	}
	return result
}

// ValidateIdentifier checks if a string is a valid SQL identifier
// Prevents SQL injection in table/column names
func ValidateIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	// Must start with letter or underscore
	first := rune(name[0])
	if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') && first != '_' {
		return fmt.Errorf("identifier must start with letter or underscore: %s", name)
	}

	// Must contain only letters, digits, underscores
	for _, char := range name {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '_' {
			return fmt.Errorf("identifier contains invalid character: %s", name)
		}
	}

	return nil
}
