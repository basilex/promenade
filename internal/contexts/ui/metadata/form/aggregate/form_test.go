package aggregate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formerrors "github.com/basilex/promenade/internal/contexts/ui/metadata/form"
	"github.com/basilex/promenade/pkg/jsonstore"
)

func TestNewFormDefinition_Success(t *testing.T) {
	layout := map[string]any{"type": "grid", "columns": 2}
	fields := []map[string]any{
		{"name": "email", "type": "text", "required": true},
		{"name": "phone", "type": "text", "required": false},
	}

	form, err := NewFormDefinition("customer-form", "customer", "Customer Form", layout, fields)

	require.NoError(t, err)
	assert.NotNil(t, form)
	assert.Equal(t, "customer-form", form.FormID)
	assert.Equal(t, "customer", form.EntityType)
	assert.Equal(t, "Customer Form", form.Name)
	assert.True(t, form.IsActive)
	assert.NotNil(t, form.GetID())
}

func TestNewFormDefinition_EmptyFormID(t *testing.T) {
	layout := map[string]any{"type": "grid"}
	fields := []map[string]any{{"name": "email"}}

	form, err := NewFormDefinition("", "customer", "Customer Form", layout, fields)

	assert.Error(t, err)
	assert.Nil(t, form)
	assert.Equal(t, formerrors.ErrFormIDRequired, err)
}

func TestNewFormDefinition_EmptyEntityType(t *testing.T) {
	layout := map[string]any{"type": "grid"}
	fields := []map[string]any{{"name": "email"}}

	form, err := NewFormDefinition("customer-form", "", "Customer Form", layout, fields)

	assert.Error(t, err)
	assert.Nil(t, form)
	assert.Equal(t, formerrors.ErrEntityTypeRequired, err)
}

func TestNewFormDefinition_EmptyName(t *testing.T) {
	layout := map[string]any{"type": "grid"}
	fields := []map[string]any{{"name": "email"}}

	form, err := NewFormDefinition("customer-form", "customer", "", layout, fields)

	assert.Error(t, err)
	assert.Nil(t, form)
	assert.Equal(t, formerrors.ErrFormNameRequired, err)
}

func TestNewFormDefinition_NilLayout(t *testing.T) {
	fields := []map[string]any{{"name": "email"}}

	form, err := NewFormDefinition("customer-form", "customer", "Customer Form", nil, fields)

	assert.Error(t, err)
	assert.Nil(t, form)
	assert.Equal(t, formerrors.ErrLayoutRequired, err)
}

func TestNewFormDefinition_NilFields(t *testing.T) {
	layout := map[string]any{"type": "grid"}

	form, err := NewFormDefinition("customer-form", "customer", "Customer Form", layout, nil)

	assert.Error(t, err)
	assert.Nil(t, form)
	assert.Equal(t, formerrors.ErrFieldsRequired, err)
}

func TestFormDefinition_Validate_Success(t *testing.T) {
	layout := map[string]any{"type": "grid"}
	fields := []map[string]any{{"name": "email"}}

	form, err := NewFormDefinition("customer-form", "customer", "Customer Form", layout, fields)
	require.NoError(t, err)

	err = form.Validate()
	assert.NoError(t, err)
}

func TestFormDefinition_Validate_EmptyFormID(t *testing.T) {
	form := &FormDefinition{
		FormID:     "",
		EntityType: "customer",
		Name:       "Customer Form",
		Layout:     jsonstore.NewField(map[string]any{"type": "grid"}),
		Fields:     jsonstore.NewField([]map[string]any{{"name": "email"}}),
	}

	err := form.Validate()
	assert.Error(t, err)
	assert.Equal(t, formerrors.ErrFormIDRequired, err)
}

func TestFormDefinition_Validate_EmptyEntityType(t *testing.T) {
	form := &FormDefinition{
		FormID:     "customer-form",
		EntityType: "",
		Name:       "Customer Form",
		Layout:     jsonstore.NewField(map[string]any{"type": "grid"}),
		Fields:     jsonstore.NewField([]map[string]any{{"name": "email"}}),
	}

	err := form.Validate()
	assert.Error(t, err)
	assert.Equal(t, formerrors.ErrEntityTypeRequired, err)
}

func TestFormDefinition_Validate_EmptyName(t *testing.T) {
	form := &FormDefinition{
		FormID:     "customer-form",
		EntityType: "customer",
		Name:       "",
		Layout:     jsonstore.NewField(map[string]any{"type": "grid"}),
		Fields:     jsonstore.NewField([]map[string]any{{"name": "email"}}),
	}

	err := form.Validate()
	assert.Error(t, err)
	assert.Equal(t, formerrors.ErrFormNameRequired, err)
}
