package jsonb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    Map
		wantNull bool
		wantJSON string
	}{
		{
			name:     "nil map returns NULL",
			input:    nil,
			wantNull: true,
		},
		{
			name:     "empty map",
			input:    Map{},
			wantJSON: `{}`,
		},
		{
			name: "map with data",
			input: Map{
				"name": "John",
				"age":  30,
			},
			wantJSON: `{"age":30,"name":"John"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.input.Value()
			require.NoError(t, err)

			if tt.wantNull {
				assert.Nil(t, val)
			} else {
				bytes, ok := val.([]byte)
				require.True(t, ok)
				assert.JSONEq(t, tt.wantJSON, string(bytes))
			}
		})
	}
}

func TestMap_Scan(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantMap   Map
		wantError bool
	}{
		{
			name:    "nil value",
			input:   nil,
			wantMap: nil,
		},
		{
			name:  "valid JSON",
			input: []byte(`{"name":"John","age":30}`),
			wantMap: Map{
				"name": "John",
				"age":  float64(30),
			},
		},
		{
			name:      "invalid type",
			input:     "not a byte slice",
			wantError: true,
		},
		{
			name:      "invalid JSON",
			input:     []byte(`{invalid`),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Map
			err := m.Scan(tt.input)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantMap, m)
			}
		})
	}
}

func TestArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    Array
		wantNull bool
		wantJSON string
	}{
		{
			name:     "nil array returns NULL",
			input:    nil,
			wantNull: true,
		},
		{
			name:     "empty array",
			input:    Array{},
			wantJSON: `[]`,
		},
		{
			name:     "array with data",
			input:    Array{"one", 2, true},
			wantJSON: `["one",2,true]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.input.Value()
			require.NoError(t, err)

			if tt.wantNull {
				assert.Nil(t, val)
			} else {
				bytes, ok := val.([]byte)
				require.True(t, ok)
				assert.JSONEq(t, tt.wantJSON, string(bytes))
			}
		})
	}
}

func TestArray_Scan(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantArray Array
		wantError bool
	}{
		{
			name:      "nil value",
			input:     nil,
			wantArray: nil,
		},
		{
			name:      "valid JSON array",
			input:     []byte(`["one",2,true]`),
			wantArray: Array{"one", float64(2), true},
		},
		{
			name:      "invalid type",
			input:     "not a byte slice",
			wantError: true,
		},
		{
			name:      "invalid JSON",
			input:     []byte(`[invalid`),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Array
			err := a.Scan(tt.input)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantArray, a)
			}
		})
	}
}

func TestJSON_Value(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name     string
		input    JSON[TestStruct]
		wantNull bool
		wantJSON string
	}{
		{
			name:     "invalid (NULL)",
			input:    JSON[TestStruct]{Valid: false},
			wantNull: true,
		},
		{
			name: "valid struct",
			input: JSON[TestStruct]{
				Data:  TestStruct{Name: "John", Age: 30},
				Valid: true,
			},
			wantJSON: `{"name":"John","age":30}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.input.Value()
			require.NoError(t, err)

			if tt.wantNull {
				assert.Nil(t, val)
			} else {
				bytes, ok := val.([]byte)
				require.True(t, ok)
				assert.JSONEq(t, tt.wantJSON, string(bytes))
			}
		})
	}
}

func TestJSON_Scan(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name      string
		input     interface{}
		wantData  TestStruct
		wantValid bool
		wantError bool
	}{
		{
			name:      "nil value",
			input:     nil,
			wantValid: false,
		},
		{
			name:      "valid JSON",
			input:     []byte(`{"name":"John","age":30}`),
			wantData:  TestStruct{Name: "John", Age: 30},
			wantValid: true,
		},
		{
			name:      "invalid type",
			input:     "not a byte slice",
			wantError: true,
		},
		{
			name:      "invalid JSON",
			input:     []byte(`{invalid`),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSON[TestStruct]
			err := j.Scan(tt.input)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValid, j.Valid)
				if tt.wantValid {
					assert.Equal(t, tt.wantData, j.Data)
				}
			}
		})
	}
}

func TestJSON_SetUnset(t *testing.T) {
	type TestStruct struct {
		Value string `json:"value"`
	}

	j := JSON[TestStruct]{}
	assert.False(t, j.Valid, "should be invalid initially")

	j.Set(TestStruct{Value: "test"})
	assert.True(t, j.Valid, "should be valid after Set")
	assert.Equal(t, "test", j.Data.Value)

	j.Unset()
	assert.False(t, j.Valid, "should be invalid after Unset")
	assert.Equal(t, "", j.Data.Value, "should be zero value")
}

func TestJSON_MarshalJSON(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name     string
		input    JSON[TestStruct]
		wantJSON string
	}{
		{
			name:     "invalid returns null",
			input:    JSON[TestStruct]{Valid: false},
			wantJSON: `null`,
		},
		{
			name: "valid returns data",
			input: JSON[TestStruct]{
				Data:  TestStruct{Name: "John"},
				Valid: true,
			},
			wantJSON: `{"name":"John"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes, err := json.Marshal(tt.input)
			require.NoError(t, err)
			assert.JSONEq(t, tt.wantJSON, string(bytes))
		})
	}
}

func TestJSON_UnmarshalJSON(t *testing.T) {
	type TestStruct struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name      string
		input     string
		wantData  TestStruct
		wantValid bool
		wantError bool
	}{
		{
			name:      "null input",
			input:     `null`,
			wantValid: false,
		},
		{
			name:      "valid JSON",
			input:     `{"name":"John"}`,
			wantData:  TestStruct{Name: "John"},
			wantValid: true,
		},
		{
			name:      "invalid JSON",
			input:     `{invalid`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var j JSON[TestStruct]
			err := json.Unmarshal([]byte(tt.input), &j)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValid, j.Valid)
				if tt.wantValid {
					assert.Equal(t, tt.wantData, j.Data)
				}
			}
		})
	}
}
