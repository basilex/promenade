package validator

import (
	"testing"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestReg(t *testing.T) {
	v := validator.New()
	err := RegisterCustomValidators(v)
	assert.NoError(t, err)
}

func TestUpper(t *testing.T) {
	v := validator.New()
	RegisterCustomValidators(v)
	type S struct { F string `validate:"uppercase"` }
	assert.NoError(t, v.Struct(S{F: "HELLO"}))
	assert.Error(t, v.Struct(S{F: "hello"}))
}

func TestAlpha(t *testing.T) {
	v := validator.New()
	RegisterCustomValidators(v)
	type S struct { F string `validate:"alpha"` }
	assert.NoError(t, v.Struct(S{F: "hello"}))
	assert.Error(t, v.Struct(S{F: "hello123"}))
}

func TestAlphaNum(t *testing.T) {
	v := validator.New()
	RegisterCustomValidators(v)
	type S struct { F string `validate:"alphanum"` }
	assert.NoError(t, v.Struct(S{F: "hello123"}))
	assert.Error(t, v.Struct(S{F: "hello 123"}))
}

func TestNoSpecial(t *testing.T) {
	v := validator.New()
	RegisterCustomValidators(v)
	type S struct { F string `validate:"no_special"` }
	assert.NoError(t, v.Struct(S{F: "hello-world"}))
	assert.Error(t, v.Struct(S{F: "hello.world"}))
}
