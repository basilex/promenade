package ref

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTo_Int(t *testing.T) {
	v := 42
	p := To(v)

	assert.NotNil(t, p)
	assert.Equal(t, 42, *p)
}

func TestTo_String(t *testing.T) {
	v := "hello"
	p := To(v)

	assert.NotNil(t, p)
	assert.Equal(t, "hello", *p)
}

func TestTo_Float64(t *testing.T) {
	v := 3.14
	p := To(v)

	assert.NotNil(t, p)
	assert.Equal(t, 3.14, *p)
}

func TestTo_Bool(t *testing.T) {
	v := true
	p := To(v)

	assert.NotNil(t, p)
	assert.True(t, *p)
}

func TestTo_Struct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	v := TestStruct{Name: "John", Age: 30}
	p := To(v)

	assert.NotNil(t, p)
	assert.Equal(t, "John", p.Name)
	assert.Equal(t, 30, p.Age)
}

func TestTo_ZeroValues(t *testing.T) {
	// Test that zero values work correctly
	intPtr := To(0)
	assert.NotNil(t, intPtr)
	assert.Equal(t, 0, *intPtr)

	strPtr := To("")
	assert.NotNil(t, strPtr)
	assert.Equal(t, "", *strPtr)

	boolPtr := To(false)
	assert.NotNil(t, boolPtr)
	assert.False(t, *boolPtr)
}

func TestInt64(t *testing.T) {
	v := int64(331002651)
	p := Int64(v)

	assert.NotNil(t, p)
	assert.Equal(t, int64(331002651), *p)
}

func TestInt(t *testing.T) {
	v := 42
	p := Int(v)

	assert.NotNil(t, p)
	assert.Equal(t, 42, *p)
}

func TestFloat64(t *testing.T) {
	v := 38.8951
	p := Float64(v)

	assert.NotNil(t, p)
	assert.Equal(t, 38.8951, *p)
}
