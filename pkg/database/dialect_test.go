package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock dialect for testing
type mockDialect struct {
	BaseDialect
	placeholderFormat string
}

func newMockDialect(name, format string) *mockDialect {
	return &mockDialect{
		BaseDialect:       NewBaseDialect(name),
		placeholderFormat: format,
	}
}

func (d *mockDialect) Placeholder(n int) string {
	if d.placeholderFormat == "?" {
		return "?"
	}
	return d.Name() + "_placeholder"
}

func (d *mockDialect) SupportsReturning() bool            { return true }
func (d *mockDialect) SupportsJSON() bool                 { return true }
func (d *mockDialect) SupportsJSONIndex() bool            { return true }
func (d *mockDialect) SupportsUUID() bool                 { return true }
func (d *mockDialect) QuoteIdentifier(name string) string { return `"` + name + `"` }
func (d *mockDialect) UUIDType() string                   { return "UUID" }
func (d *mockDialect) JSONType() string                   { return "JSON" }
func (d *mockDialect) TimestampType() string              { return "TIMESTAMP" }
func (d *mockDialect) BoolType() string                   { return "BOOLEAN" }

func TestBaseDialect_Name(t *testing.T) {
	dialect := NewBaseDialect("postgres")
	assert.Equal(t, "postgres", dialect.Name())

	dialect2 := NewBaseDialect("sqlite")
	assert.Equal(t, "sqlite", dialect2.Name())
}

func TestConvertPlaceholders_PostgresNoConversion(t *testing.T) {
	dialect := newMockDialect("postgres", "$")
	query := "SELECT * FROM users WHERE id = $1 AND email = $2"

	result := ConvertPlaceholders(query, dialect)

	// Postgres format should remain unchanged
	assert.Equal(t, query, result)
}

func TestConvertPlaceholders_SQLiteConversion(t *testing.T) {
	dialect := newMockDialect("sqlite", "?")
	query := "SELECT * FROM users WHERE id = $1 AND email = $2"

	result := ConvertPlaceholders(query, dialect)

	expected := "SELECT * FROM users WHERE id = ? AND email = ?"
	assert.Equal(t, expected, result)
}

func TestConvertPlaceholders_MultipleDigits(t *testing.T) {
	dialect := newMockDialect("sqlite", "?")
	query := "INSERT INTO users VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"

	result := ConvertPlaceholders(query, dialect)

	// Should have 11 question marks
	count := 0
	for _, char := range result {
		if char == '?' {
			count++
		}
	}
	assert.Equal(t, 11, count)
	assert.NotContains(t, result, "$")
}

func TestConvertPlaceholders_InStringLiterals(t *testing.T) {
	dialect := newMockDialect("sqlite", "?")
	query := "SELECT * FROM users WHERE email = '$1@example.com' AND id = $1"

	result := ConvertPlaceholders(query, dialect)

	// $1 inside string literal should NOT be converted
	assert.Contains(t, result, "'$1@example.com'")
	// But $1 outside should be converted to ?
	assert.Contains(t, result, "id = ?")
}

func TestConvertPlaceholders_EscapedQuotes(t *testing.T) {
	dialect := newMockDialect("sqlite", "?")
	query := `SELECT * FROM users WHERE name = 'O\'Brien' AND id = $1`

	result := ConvertPlaceholders(query, dialect)

	assert.Contains(t, result, `'O\'Brien'`)
	assert.Contains(t, result, "id = ?")
}

func TestConvertPlaceholders_NoDollarSign(t *testing.T) {
	dialect := newMockDialect("sqlite", "?")
	query := "SELECT * FROM users WHERE id = ? AND email = ?"

	result := ConvertPlaceholders(query, dialect)

	// Already in ? format, should remain unchanged
	assert.Equal(t, query, result)
}

func TestConvertPlaceholders_DollarNotFollowedByNumber(t *testing.T) {
	dialect := newMockDialect("sqlite", "?")
	query := "SELECT price FROM products WHERE currency = '$' AND id = $1"

	result := ConvertPlaceholders(query, dialect)

	// $ alone should not be converted, only $1
	assert.Contains(t, result, "currency = '$'")
	assert.Contains(t, result, "id = ?")
}

func TestQuoteString_Simple(t *testing.T) {
	result := QuoteString("hello")
	assert.Equal(t, "'hello'", result)
}

func TestQuoteString_WithSingleQuote(t *testing.T) {
	result := QuoteString("O'Brien")
	assert.Equal(t, "'O''Brien'", result) // Single quote escaped
}

func TestQuoteString_MultipleSingleQuotes(t *testing.T) {
	result := QuoteString("It's a test's value")
	assert.Equal(t, "'It''s a test''s value'", result)
}

func TestQuoteString_Empty(t *testing.T) {
	result := QuoteString("")
	assert.Equal(t, "''", result)
}

func TestEscapeString_NoQuotes(t *testing.T) {
	result := escapeString("hello world")
	assert.Equal(t, "hello world", result)
}

func TestEscapeString_WithQuotes(t *testing.T) {
	result := escapeString("it's working")
	assert.Equal(t, "it''s working", result)
}

func TestValidateIdentifier_Valid(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
	}{
		{"simple", "users"},
		{"with underscore", "user_id"},
		{"starts with underscore", "_private"},
		{"mixed case", "UserId"},
		{"with numbers", "user123"},
		{"all caps", "USER_ID"},
		{"long name", "very_long_table_name_with_many_words"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.identifier)
			assert.NoError(t, err, "Should be valid: %s", tt.identifier)
		})
	}
}

func TestValidateIdentifier_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		identifier  string
		errContains string
	}{
		{"empty", "", "cannot be empty"},
		{"starts with number", "123users", "must start with letter or underscore"},
		{"contains space", "user id", "invalid character"},
		{"contains dash", "user-id", "invalid character"},
		{"contains dot", "users.id", "invalid character"},
		{"contains special char", "user$id", "invalid character"},
		{"sql injection attempt", "users; DROP TABLE", "invalid character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.identifier)
			require.Error(t, err, "Should be invalid: %s", tt.identifier)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestValidateIdentifier_SQLInjectionPrevention(t *testing.T) {
	maliciousInputs := []string{
		"users; DROP TABLE users;--",
		"users' OR '1'='1",
		"users/*comment*/",
		"users--comment",
		"users\n",
		"users\t",
	}

	for _, input := range maliciousInputs {
		err := ValidateIdentifier(input)
		assert.Error(t, err, "Should reject malicious input: %s", input)
	}
}

// Benchmark tests
func BenchmarkConvertPlaceholders_ShortQuery(b *testing.B) {
	dialect := newMockDialect("sqlite", "?")
	query := "SELECT * FROM users WHERE id = $1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertPlaceholders(query, dialect)
	}
}

func BenchmarkConvertPlaceholders_LongQuery(b *testing.B) {
	dialect := newMockDialect("sqlite", "?")
	query := "INSERT INTO users (id, email, name, age, city, country, phone, address, zip, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertPlaceholders(query, dialect)
	}
}

func BenchmarkValidateIdentifier(b *testing.B) {
	identifier := "user_id_with_long_name"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateIdentifier(identifier) // Ignore error in benchmark
	}
}

func BenchmarkQuoteString(b *testing.B) {
	str := "It's a test string with quotes"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		QuoteString(str)
	}
}
