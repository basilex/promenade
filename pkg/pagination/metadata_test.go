package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	m := NewMetadata(100, 10, 0)
	assert.Equal(t, 100, m.Total)
	assert.Equal(t, 10, m.TotalPages)
	assert.True(t, m.HasNext)
	assert.False(t, m.HasPrev)
}

func TestPage2(t *testing.T) {
	m := NewMetadata(100, 10, 10)
	assert.Equal(t, 2, m.CurrentPage)
	assert.True(t, m.HasPrev)
}

func TestLastPage(t *testing.T) {
	m := NewMetadata(100, 10, 90)
	assert.False(t, m.HasNext)
}

func TestZeroLimit(t *testing.T) {
	m := NewMetadata(100, 0, 0)
	assert.Equal(t, 10, m.Limit)
}

func TestOffset(t *testing.T) {
	p := Params{Page: 1, PageSize: 10}
	assert.Equal(t, 0, p.GetOffset())
	p = Params{Page: 2, PageSize: 10}
	assert.Equal(t, 10, p.GetOffset())
}

func TestLimit(t *testing.T) {
	p := Params{PageSize: 10}
	assert.Equal(t, 10, p.GetLimit())
	p = Params{}
	assert.Equal(t, 20, p.GetLimit())
	p = Params{Limit: 15}
	assert.Equal(t, 15, p.GetLimit())
	p = Params{PageSize: 10, Limit: 15}
	assert.Equal(t, 10, p.GetLimit())
}

func TestEdgeCases(t *testing.T) {
	m := NewMetadata(0, 10, 0)
	assert.Equal(t, 1, m.TotalPages)
	p := Params{Page: 0}
	assert.Equal(t, 0, p.GetOffset())
	p = Params{PageSize: 0}
	assert.Equal(t, 0, p.GetOffset())
}
