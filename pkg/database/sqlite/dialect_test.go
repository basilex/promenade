package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDialect(t *testing.T) {
	dialect := NewDialect()
	assert.NotNil(t, dialect)
	assert.Equal(t, "sqlite", dialect.Name())
}

func TestDialect_Placeholder(t *testing.T) {
	dialect := NewDialect()
	// SQLite uses ? for all placeholders
	for _, n := range []int{1, 2, 10, 100} {
		result := dialect.Placeholder(n)
		assert.Equal(t, "?", result)
	}
}

func TestDialect_Supports(t *testing.T) {
	dialect := NewDialect()
	assert.True(t, dialect.SupportsReturning())
	assert.False(t, dialect.SupportsJSON())
	assert.False(t, dialect.SupportsJSONIndex())
	assert.False(t, dialect.SupportsUUID())
}

func TestDialect_Types(t *testing.T) {
	dialect := NewDialect()
	assert.Equal(t, "TEXT", dialect.UUIDType())
	assert.Equal(t, "TEXT", dialect.JSONType())
	assert.Equal(t, "DATETIME", dialect.TimestampType())
	assert.Equal(t, "BOOLEAN", dialect.BoolType())
}

func TestDialect_QuoteIdentifier(t *testing.T) {
	dialect := NewDialect()
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "`users`"},
		{"user_id", "`user_id`"},
		{"_private", "`_private`"},
	}
	for _, tt := range tests {
		result := dialect.QuoteIdentifier(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestDialect_JSONExtract(t *testing.T) {
	dialect := NewDialect()
	result := dialect.JSONExtract("metadata", "name")
	expected := "json_extract(`metadata`, '$.name')"
	assert.Equal(t, expected, result)
}

func TestDialect_JSONArrayContains(t *testing.T) {
	dialect := NewDialect()
	result := dialect.JSONArrayContains("tags", "vip")
	expected := "json_extract(`tags`, '$') LIKE '%\"vip\"%'"
	assert.Equal(t, expected, result)
}

func TestDialect_CreateUUIDCheckConstraint(t *testing.T) {
	dialect := NewDialect()
	result := dialect.CreateUUIDCheckConstraint("id")
	assert.Contains(t, result, "CHECK(length(`id`) = 36")
	assert.Contains(t, result, "GLOB")
}

func TestDialect_CreateJSONCheckConstraint(t *testing.T) {
	dialect := NewDialect()
	result := dialect.CreateJSONCheckConstraint("metadata")
	expected := "CHECK(json_valid(`metadata`))"
	assert.Equal(t, expected, result)
}
