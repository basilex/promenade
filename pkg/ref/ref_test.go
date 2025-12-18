package ref_test

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/ref"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"non-empty string", "hello", "hello"},
		{"empty string", "", ""},
		{"with spaces", "  spaces  ", "  spaces  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.String(tt.input)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestStringValue(t *testing.T) {
	tests := []struct {
		name  string
		input *string
		want  string
	}{
		{"non-nil pointer", ref.String("hello"), "hello"},
		{"nil pointer", nil, ""},
		{"empty string pointer", ref.String(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.StringValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringOr(t *testing.T) {
	tests := []struct {
		name         string
		input        *string
		defaultValue string
		want         string
	}{
		{"non-nil pointer", ref.String("hello"), "default", "hello"},
		{"nil pointer", nil, "default", "default"},
		{"empty string pointer", ref.String(""), "default", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.StringOr(tt.input, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTime(t *testing.T) {
	now := time.Now()
	got := ref.Time(now)
	assert.NotNil(t, got)
	assert.Equal(t, now, *got)
}

func TestTimeValue(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input *time.Time
		want  time.Time
	}{
		{"non-nil pointer", ref.Time(now), now},
		{"nil pointer", nil, time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.TimeValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTimeOr(t *testing.T) {
	now := time.Now()
	defaultTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		input        *time.Time
		defaultValue time.Time
		want         time.Time
	}{
		{"non-nil pointer", ref.Time(now), defaultTime, now},
		{"nil pointer", nil, defaultTime, defaultTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.TimeOr(tt.input, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUUID(t *testing.T) {
	uuid := uuidv7.New()
	got := ref.UUID(uuid)
	assert.NotNil(t, got)
	assert.Equal(t, uuid, *got)
}

func TestUUIDValue(t *testing.T) {
	uuid := uuidv7.New()

	tests := []struct {
		name  string
		input *uuidv7.UUID
		want  uuidv7.UUID
	}{
		{"non-nil pointer", ref.UUID(uuid), uuid},
		{"nil pointer", nil, uuidv7.UUID{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.UUIDValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  int
	}{
		{"positive", 42, 42},
		{"negative", -10, -10},
		{"zero", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.Int(tt.input)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestIntValue(t *testing.T) {
	tests := []struct {
		name  string
		input *int
		want  int
	}{
		{"non-nil pointer", ref.Int(42), 42},
		{"nil pointer", nil, 0},
		{"zero value pointer", ref.Int(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.IntValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIntOr(t *testing.T) {
	tests := []struct {
		name         string
		input        *int
		defaultValue int
		want         int
	}{
		{"non-nil pointer", ref.Int(42), 100, 42},
		{"nil pointer", nil, 100, 100},
		{"zero value pointer", ref.Int(0), 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.IntOr(tt.input, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		want  bool
	}{
		{"true", true, true},
		{"false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.Bool(tt.input)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, *got)
		})
	}
}

func TestBoolValue(t *testing.T) {
	tests := []struct {
		name  string
		input *bool
		want  bool
	}{
		{"non-nil true", ref.Bool(true), true},
		{"non-nil false", ref.Bool(false), false},
		{"nil pointer", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.BoolValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBoolOr(t *testing.T) {
	tests := []struct {
		name         string
		input        *bool
		defaultValue bool
		want         bool
	}{
		{"non-nil true", ref.Bool(true), false, true},
		{"non-nil false", ref.Bool(false), true, false},
		{"nil pointer default true", nil, true, true},
		{"nil pointer default false", nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.BoolOr(tt.input, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsNil(t *testing.T) {
	tests := []struct {
		name  string
		input *string
		want  bool
	}{
		{"nil pointer", nil, true},
		{"non-nil pointer", ref.String("hello"), false},
		{"empty string pointer", ref.String(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.IsNil(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSet(t *testing.T) {
	tests := []struct {
		name  string
		input *string
		want  bool
	}{
		{"nil pointer", nil, false},
		{"empty string pointer", ref.String(""), false},
		{"non-empty string pointer", ref.String("hello"), true},
		{"whitespace string pointer", ref.String("  "), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ref.IsSet(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
