package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formerrors "github.com/basilex/promenade/internal/contexts/ui/metadata/form"
	"github.com/basilex/promenade/internal/contexts/ui/metadata/form/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockRepository implements IFormRepository for testing
type MockRepository struct {
	CreateFunc      func(ctx context.Context, form *aggregate.FormDefinition) error
	GetByIDFunc     func(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error)
	GetByFormIDFunc func(ctx context.Context, formID string) (*aggregate.FormDefinition, error)
	UpdateFunc      func(ctx context.Context, form *aggregate.FormDefinition) error
	DeleteFunc      func(ctx context.Context, id uuidv7.UUID) error
	ListFunc        func(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error)
	ListAllFunc     func(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error)
}

func (m *MockRepository) Create(ctx context.Context, form *aggregate.FormDefinition) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, form)
	}
	return nil
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, formerrors.ErrFormNotFound
}

func (m *MockRepository) GetByFormID(ctx context.Context, formID string) (*aggregate.FormDefinition, error) {
	if m.GetByFormIDFunc != nil {
		return m.GetByFormIDFunc(ctx, formID)
	}
	return nil, formerrors.ErrFormNotFound
}

func (m *MockRepository) Update(ctx context.Context, form *aggregate.FormDefinition) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, form)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRepository) List(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, entityType, limit, offset)
	}
	return []*aggregate.FormDefinition{}, 0, nil
}

func (m *MockRepository) ListAll(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx, limit, offset)
	}
	return []*aggregate.FormDefinition{}, 0, nil
}

func TestNewFormUseCase(t *testing.T) {
	uc := NewFormUseCase(&MockRepository{})
	assert.NotNil(t, uc)
}

func TestCreateForm_Success(t *testing.T) {
	repo := &MockRepository{
		GetByFormIDFunc: func(ctx context.Context, formID string) (*aggregate.FormDefinition, error) {
			return nil, formerrors.ErrFormNotFound
		},
		CreateFunc: func(ctx context.Context, form *aggregate.FormDefinition) error {
			return nil
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	form, err := aggregate.NewFormDefinition(
		"customer-form",
		"customer",
		"Customer Form",
		map[string]any{"type": "grid"},
		[]map[string]any{{"name": "email", "type": "text"}},
	)
	require.NoError(t, err)

	result, err := uc.CreateForm(ctx, form)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "customer-form", result.FormID)
}

func TestCreateForm_NilForm(t *testing.T) {
	repo := &MockRepository{}
	uc := NewFormUseCase(repo)
	ctx := context.Background()

	result, err := uc.CreateForm(ctx, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, formerrors.ErrFormCreateFailed, err)
}

func TestCreateForm_DuplicateFormID(t *testing.T) {
	existingForm, _ := aggregate.NewFormDefinition(
		"customer-form",
		"customer",
		"Existing Form",
		map[string]any{"type": "grid"},
		[]map[string]any{{"name": "email"}},
	)

	repo := &MockRepository{
		GetByFormIDFunc: func(ctx context.Context, formID string) (*aggregate.FormDefinition, error) {
			if formID == "customer-form" {
				return existingForm, nil
			}
			return nil, formerrors.ErrFormNotFound
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	form, err := aggregate.NewFormDefinition(
		"customer-form",
		"customer",
		"New Form",
		map[string]any{"type": "grid"},
		[]map[string]any{{"name": "name"}},
	)
	require.NoError(t, err)

	result, err := uc.CreateForm(ctx, form)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, formerrors.ErrFormIDExists, err)
}

func TestGetForm_Success(t *testing.T) {
	formID := uuidv7.New()
	expectedForm := &aggregate.FormDefinition{
		FormID:     "customer-form",
		EntityType: "customer",
		Name:       "Customer Form",
		Layout:     jsonstore.NewField(map[string]any{"type": "grid"}),
		Fields:     jsonstore.NewField([]map[string]any{{"name": "email"}}),
	}
	expectedForm.ID = formID

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error) {
			if id == formID {
				return expectedForm, nil
			}
			return nil, formerrors.ErrFormNotFound
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	result, err := uc.GetForm(ctx, formID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, formID, result.GetID())
}

func TestGetForm_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error) {
			return nil, formerrors.ErrFormNotFound
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	result, err := uc.GetForm(ctx, uuidv7.New())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, formerrors.ErrFormNotFound, err)
}

func TestGetFormByFormID_Success(t *testing.T) {
	expectedForm := &aggregate.FormDefinition{
		FormID:     "customer-form",
		EntityType: "customer",
		Name:       "Customer Form",
		Layout:     jsonstore.NewField(map[string]any{"type": "grid"}),
		Fields:     jsonstore.NewField([]map[string]any{{"name": "email"}}),
	}

	repo := &MockRepository{
		GetByFormIDFunc: func(ctx context.Context, formID string) (*aggregate.FormDefinition, error) {
			if formID == "customer-form" {
				return expectedForm, nil
			}
			return nil, formerrors.ErrFormNotFound
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	result, err := uc.GetFormByFormID(ctx, "customer-form")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "customer-form", result.FormID)
}

func TestUpdateForm_Success(t *testing.T) {
	formID := uuidv7.New()
	existingForm := &aggregate.FormDefinition{
		FormID:     "customer-form",
		EntityType: "customer",
		Name:       "Old Name",
		Layout:     jsonstore.NewField(map[string]any{"type": "grid"}),
		Fields:     jsonstore.NewField([]map[string]any{{"name": "email"}}),
	}
	existingForm.ID = formID

	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error) {
			if id == formID {
				return existingForm, nil
			}
			return nil, formerrors.ErrFormNotFound
		},
		UpdateFunc: func(ctx context.Context, form *aggregate.FormDefinition) error {
			return nil
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	update := &aggregate.FormDefinition{
		EntityType: "customer",
		Name:       "New Name",
		Layout:     jsonstore.NewField(map[string]any{"type": "flex"}),
		Fields:     jsonstore.NewField([]map[string]any{{"name": "name"}}),
	}

	result, err := uc.UpdateForm(ctx, formID, update)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Name", result.Name)
}

func TestUpdateForm_NotFound(t *testing.T) {
	repo := &MockRepository{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error) {
			return nil, formerrors.ErrFormNotFound
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	update := &aggregate.FormDefinition{
		Name: "New Name",
	}

	result, err := uc.UpdateForm(ctx, uuidv7.New(), update)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, formerrors.ErrFormNotFound, err)
}

func TestDeleteForm_Success(t *testing.T) {
	formID := uuidv7.New()

	repo := &MockRepository{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			if id == formID {
				return nil
			}
			return formerrors.ErrFormNotFound
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	err := uc.DeleteForm(ctx, formID)
	assert.NoError(t, err)
}

func TestListForms_Success(t *testing.T) {
	forms := []*aggregate.FormDefinition{
		{
			FormID:     "customer-form",
			EntityType: "customer",
			Name:       "Customer Form",
			Layout:     jsonstore.NewField(map[string]any{"type": "grid"}),
			Fields:     jsonstore.NewField([]map[string]any{{"name": "email"}}),
		},
	}

	repo := &MockRepository{
		ListFunc: func(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
			if entityType == "customer" {
				return forms, 1, nil
			}
			return []*aggregate.FormDefinition{}, 0, nil
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	results, total, err := uc.ListForms(ctx, "customer", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, results, 1)
	assert.Equal(t, "customer-form", results[0].FormID)
}

func TestListAllForms_Success(t *testing.T) {
	forms := []*aggregate.FormDefinition{
		{
			FormID:     "customer-form",
			EntityType: "customer",
			Name:       "Customer Form",
		},
		{
			FormID:     "order-form",
			EntityType: "order",
			Name:       "Order Form",
		},
	}

	repo := &MockRepository{
		ListAllFunc: func(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
			return forms, len(forms), nil
		},
	}

	uc := NewFormUseCase(repo)
	ctx := context.Background()

	results, total, err := uc.ListAllForms(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, results, 2)
}
