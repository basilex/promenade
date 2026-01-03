package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDialect(t *testing.T) {
	dialect := NewDialect()

	assert.NotNil(t, dialect)
	assert.Equal(t, "postgres", dialect.Name())
}

func TestDialect_Placeholder(t *testing.T) {
	dialect := NewDialect()

	tests := []struct {
		n        int
		expected string
	}{
		{1, "$1"},
		{2, "$2"},
		{10, "$10"},
		{100, "$100"},
		{999, "$999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := dialect.Placeholder(tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDialect_SupportsReturning(t *testing.T) {
	dialect := NewDialect()
	assert.True(t, dialect.SupportsReturning(), "PostgreSQL supports RETURNING clause")
}

func TestDialect_SupportsJSON(t *testing.T) {
	dialect := NewDialect()
	assert.True(t, dialect.SupportsJSON(), "PostgreSQL has native JSONB type")
}

func TestDialect_SupportsJSONIndex(t *testing.T) {
	dialect := NewDialect()
	assert.True(t, dialect.SupportsJSONIndex(), "PostgreSQL supports GIN indexes")
}

func TestDialect_SupportsUUID(t *testing.T) {
	dialect := NewDialect()
	assert.True(t, dialect.SupportsUUID(), "PostgreSQL has native UUID type")
}

func TestDialect_QuoteIdentifier(t *testing.T) {
	dialect := NewDialect()

	tests := []struct {
		input    string
		expected string
	}{
		{"users", `"users"`},
		{"user_id", `"user_id"`},
		{"UserID", `"UserID"`},
		{"_private", `"_private"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := dialect.QuoteIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDialect_UUIDType(t *testing.T) {
	dialect := NewDialect()
	assert.Equal(t, "UUID", dialect.UUIDType())
}

func TestDialect_JSONType(t *testing.T) {
	dialect := NewDialect()
	assert.Equal(t, "JSONB", dialect.JSONType(), "Should prefer JSONB over JSON")
}

func TestDialect_TimestampType(t *testing.T) {
	dialect := NewDialect()
	assert.Equal(t, "TIMESTAMP", dialect.TimestampType())
}

func TestDialect_BoolType(t *testing.T) {
	dialect := NewDialect()
	assert.Equal(t, "BOOLEAN", dialect.BoolType())
}

func TestDialect_CreateJSONIndex(t *testing.T) {
	dialect := NewDialect()

	result := dialect.CreateJSONIndex("customers", "tags", "idx_customer_tags")

	expected := `CREATE INDEX "idx_customer_tags" ON "customers" USING gin("tags")`
	assert.Equal(t, expected, result)
}

func TestDialect_CreateJSONIndex_WithSpecialChars(t *testing.T) {
	dialect := NewDialect()

	result := dialect.CreateJSONIndex("customer_data", "metadata_field", "idx_meta")

	assert.Contains(t, result, `"customer_data"`)
	assert.Contains(t, result, `"metadata_field"`)
	assert.Contains(t, result, `"idx_meta"`)
	assert.Contains(t, result, "USING gin")
}

func TestDialect_JSONContains(t *testing.T) {
	dialect := NewDialect()

	tests := []struct {
		name     string
		column   string
		value    string
		expected string
	}{
		{
			"simple array",
			"tags",
			`["vip"]`,
			`"tags" @> '["vip"]'::jsonb`,
		},
		{
			"multiple values",
			"tags",
			`["vip","premium"]`,
			`"tags" @> '["vip","premium"]'::jsonb`,
		},
		{
			"object",
			"metadata",
			`{"status":"active"}`,
			`"metadata" @> '{"status":"active"}'::jsonb`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dialect.JSONContains(tt.column, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDialect_JSONExtract(t *testing.T) {
	dialect := NewDialect()

	tests := []struct {
		name     string
		column   string
		path     string
		expected string
	}{
		{
			"simple field",
			"metadata",
			"name",
			`"metadata"->>'name'`,
		},
		{
			"numeric field",
			"settings",
			"score",
			`"settings"->>'score'`,
		},
		{
			"nested path",
			"data",
			"user.email",
			`"data"->>'user.email'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dialect.JSONExtract(tt.column, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Integration test with all features
func TestDialect_FullQuery(t *testing.T) {
	dialect := NewDialect()

	// Build a complex query using all dialect features
	query := "SELECT " +
		dialect.JSONExtract("metadata", "name") + " AS name, " +
		dialect.QuoteIdentifier("id") + ", " +
		dialect.QuoteIdentifier("created_at") +
		" FROM " + dialect.QuoteIdentifier("customers") +
		" WHERE " + dialect.JSONContains("tags", `["vip"]`) +
		" AND " + dialect.QuoteIdentifier("id") + " = " + dialect.Placeholder(1)

	// Verify query structure
	assert.Contains(t, query, `"metadata"->>'name'`)
	assert.Contains(t, query, `"customers"`)
	assert.Contains(t, query, `"tags" @> '["vip"]'::jsonb`)
	assert.Contains(t, query, "$1")
}

// Benchmark tests
func BenchmarkDialect_Placeholder(b *testing.B) {
	dialect := NewDialect()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialect.Placeholder(5)
	}
}

func BenchmarkDialect_QuoteIdentifier(b *testing.B) {
	dialect := NewDialect()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialect.QuoteIdentifier("user_id")
	}
}

func BenchmarkDialect_JSONContains(b *testing.B) {
	dialect := NewDialect()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialect.JSONContains("tags", `["vip"]`)
	}
}

func BenchmarkDialect_CreateJSONIndex(b *testing.B) {
	dialect := NewDialect()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialect.CreateJSONIndex("customers", "tags", "idx_tags")
	}
}
