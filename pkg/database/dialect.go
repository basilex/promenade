// Package database provides database adapter abstraction.
// Optimized for PostgreSQL with dialect interface for potential future extensions.
package database

import "fmt"

// Dialect abstracts SQL syntax differences for PostgreSQL.
// Interface design allows for potential future database support.
type Dialect interface {
	// Name returns the database name (currently only "postgres" supported)
	Name() string

	// Placeholder returns the parameter placeholder for position n (1-indexed)
	// Postgres: $1, $2, $3
	// SQL Server: @p1, @p2, @p3
	Placeholder(n int) string

	// SupportsReturning returns true if database supports RETURNING clause
	// Postgres: true
	SupportsReturning() bool

	// SupportsJSON returns true if database has native JSON type
	// Postgres: true (JSONB)
	SupportsJSON() bool

	// SupportsJSONIndex returns true if database can index JSON fields
	// Postgres: true (GIN index)
	SupportsJSONIndex() bool

	// SupportsUUID returns true if database has native UUID type
	// Postgres: true
	SupportsUUID() bool

	// QuoteIdentifier quotes table/column names for safe SQL
	// Postgres: "table_name"
	QuoteIdentifier(name string) string

	// UUIDType returns the SQL type for storing UUIDs
	// Postgres: "UUID"
	UUIDType() string

	// JSONType returns the SQL type for storing JSON
	// Postgres: "JSONB"
	JSONType() string

	// TimestampType returns the SQL type for timestamps
	// Postgres: "TIMESTAMP"
	TimestampType() string

	// BoolType returns the SQL type for booleans
	// Postgres: "BOOLEAN"
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
// Currently optimized for PostgreSQL (no conversion needed)
func ConvertPlaceholders(query string, dialect Dialect) string {
	if dialect.Name() == "postgres" {
		return query // Already in correct format
	}

	// Implementation for alternative placeholder styles
	// Reserved for potential future database support
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
