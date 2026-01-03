package jsonstore

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constructors
func TestNewField(t *testing.T) {
	field := NewField([]string{"vip", "premium"})
	assert.False(t, field.IsNull())
	assert.Equal(t, []string{"vip", "premium"}, field.Get())
}

func TestNewNullField(t *testing.T) {
	field := NewNullField[[]string]()
	assert.True(t, field.IsNull())
	assert.Equal(t, []string(nil), field.Get())
}

// Test Get/Set
func TestField_GetSet(t *testing.T) {
	field := NewField([]string{"initial"})
	assert.Equal(t, []string{"initial"}, field.Get())

	field.Set([]string{"updated"})
	assert.Equal(t, []string{"updated"}, field.Get())
	assert.False(t, field.IsNull())
}

func TestField_SetClearsNull(t *testing.T) {
	field := NewNullField[[]string]()
	assert.True(t, field.IsNull())

	field.Set([]string{"value"})
	assert.False(t, field.IsNull())
}

// Test NULL handling
func TestField_SetNull(t *testing.T) {
	field := NewField([]string{"value"})
	field.SetNull()
	assert.True(t, field.IsNull())
	assert.Equal(t, []string(nil), field.Get())
}

// Test JSON marshaling
func TestField_MarshalJSON_Value(t *testing.T) {
	field := NewField([]string{"vip", "premium"})
	bytes, err := json.Marshal(field)
	require.NoError(t, err)
	assert.Equal(t, `["vip","premium"]`, string(bytes))
}

func TestField_MarshalJSON_Null(t *testing.T) {
	field := NewNullField[[]string]()
	bytes, err := json.Marshal(field)
	require.NoError(t, err)
	assert.Equal(t, "null", string(bytes))
}

// Test JSON unmarshaling
func TestField_UnmarshalJSON_Value(t *testing.T) {
	var field Field[[]string]
	err := json.Unmarshal([]byte(`["vip","premium"]`), &field)
	require.NoError(t, err)
	assert.False(t, field.IsNull())
	assert.Equal(t, []string{"vip", "premium"}, field.Get())
}

func TestField_UnmarshalJSON_Null(t *testing.T) {
	var field Field[[]string]
	err := json.Unmarshal([]byte("null"), &field)
	require.NoError(t, err)
	assert.True(t, field.IsNull())
}

// Test Scan (sql.Scanner)
func TestField_Scan_ByteSlice(t *testing.T) {
	var field Field[[]string]
	err := field.Scan([]byte(`["vip","premium"]`))
	require.NoError(t, err)
	assert.Equal(t, []string{"vip", "premium"}, field.Get())
}

func TestField_Scan_String(t *testing.T) {
	var field Field[[]string]
	err := field.Scan(`["vip","premium"]`)
	require.NoError(t, err)
	assert.Equal(t, []string{"vip", "premium"}, field.Get())
}

func TestField_Scan_Nil(t *testing.T) {
	var field Field[[]string]
	err := field.Scan(nil)
	require.NoError(t, err)
	assert.True(t, field.IsNull())
}

func TestField_Scan_InvalidType(t *testing.T) {
	var field Field[[]string]
	err := field.Scan(12345)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot scan type")
}

// Test Value (driver.Valuer)
func TestField_Value_NonNull(t *testing.T) {
	field := NewField([]string{"vip", "premium"})
	value, err := field.Value()
	require.NoError(t, err)
	assert.Equal(t, `["vip","premium"]`, value.(string))
}

func TestField_Value_Null(t *testing.T) {
	field := NewNullField[[]string]()
	value, err := field.Value()
	require.NoError(t, err)
	assert.Nil(t, value)
}

// Test utility methods
func TestField_String(t *testing.T) {
	field := NewField([]string{"vip"})
	assert.Equal(t, `["vip"]`, field.String())

	nullField := NewNullField[[]string]()
	assert.Equal(t, "NULL", nullField.String())
}

func TestField_Equal(t *testing.T) {
	field1 := NewField([]string{"vip"})
	field2 := NewField([]string{"vip"})
	assert.True(t, field1.Equal(field2))

	field3 := NewField([]string{"premium"})
	assert.False(t, field1.Equal(field3))
}

func TestField_Clone(t *testing.T) {
	original := NewField([]string{"vip"})
	cloned := original.Clone()
	assert.True(t, original.Equal(cloned))

	// Modify clone shouldn't affect original
	clonedSlice := cloned.Get()
	clonedSlice[0] = "modified"
	cloned.Set(clonedSlice)
	assert.Equal(t, "vip", original.Get()[0])
}

// Test with UUID arrays
func TestField_UUIDArray(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	field := NewField([]uuid.UUID{id1, id2})

	// Marshal to JSON
	bytes, err := json.Marshal(field)
	require.NoError(t, err)

	// Unmarshal back
	var retrieved Field[[]uuid.UUID]
	err = json.Unmarshal(bytes, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, 2, len(retrieved.Get()))
}

// Test round-trip through database
func TestField_RoundTrip(t *testing.T) {
	original := NewField([]string{"vip", "premium"})

	// Convert to driver.Value
	dbValue, err := original.Value()
	require.NoError(t, err)

	// Scan back
	var retrieved Field[[]string]
	err = retrieved.Scan(dbValue)
	require.NoError(t, err)

	assert.Equal(t, original.Get(), retrieved.Get())
}

// Additional edge case tests for full coverage

func TestField_Scan_EmptyString(t *testing.T) {
	var field Field[[]string]
	err := field.Scan("")
	require.NoError(t, err)
	assert.True(t, field.IsNull())
}

func TestField_Scan_InvalidJSON(t *testing.T) {
	var field Field[[]string]
	err := field.Scan([]byte(`{invalid`))
	require.Error(t, err)
}

func TestField_Value_EmptyArray(t *testing.T) {
	field := NewField([]string{})
	value, err := field.Value()
	require.NoError(t, err)
	assert.Equal(t, "[]", value.(string))
}

func TestField_Equal_BothNull(t *testing.T) {
	field1 := NewNullField[[]string]()
	field2 := NewNullField[[]string]()
	assert.True(t, field1.Equal(field2))
}

func TestField_Equal_OneNull(t *testing.T) {
	field1 := NewNullField[[]string]()
	field2 := NewField([]string{"value"})
	assert.False(t, field1.Equal(field2))
	assert.False(t, field2.Equal(field1))
}

func TestField_Clone_Null(t *testing.T) {
	original := NewNullField[[]string]()
	cloned := original.Clone()
	assert.True(t, original.IsNull())
	assert.True(t, cloned.IsNull())
}

func TestField_UnmarshalJSON_EmptyArray(t *testing.T) {
	var field Field[[]string]
	err := json.Unmarshal([]byte("[]"), &field)
	require.NoError(t, err)
	assert.False(t, field.IsNull())
	assert.Equal(t, []string{}, field.Get())
}

func TestField_MarshalJSON_Map(t *testing.T) {
	field := NewField(map[string]interface{}{
		"score": 95,
		"tier":  "gold",
	})
	bytes, err := json.Marshal(field)
	require.NoError(t, err)
	assert.Contains(t, string(bytes), "score")
	assert.Contains(t, string(bytes), "tier")
}

// Test marshaling errors
type unmarshalableType struct {
	Channel chan int // channels cannot be marshaled to JSON
}

func TestField_Value_MarshalError(t *testing.T) {
	field := NewField(unmarshalableType{Channel: make(chan int)})
	_, err := field.Value()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal JSON")
}

// Clone with unmarshalable type will return zero value (not NULL)
// because it ignores errors internally
func TestField_Clone_UnmarshalableType(t *testing.T) {
	field := NewField(unmarshalableType{Channel: make(chan int)})
	cloned := field.Clone()
	// Clone returns zero value on error (not NULL)
	assert.False(t, cloned.IsNull())
}
