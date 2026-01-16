package form_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	formHTTP "github.com/basilex/promenade/internal/contexts/ui/metadata/form/adapter/http"
	"github.com/basilex/promenade/internal/contexts/ui/metadata/form"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

type MockFormUseCase struct {
	CreateFormFunc       func(ctx context.Context, form *form.FormDefinition) (*form.FormDefinition, error)
	GetFormFunc          func(ctx context.Context, id uuidv7.UUID) (*form.FormDefinition, error)
	GetFormByFormIDFunc  func(ctx context.Context, formID string) (*form.FormDefinition, error)
	UpdateFormFunc       func(ctx context.Context, id uuidv7.UUID, update *form.FormDefinition) (*form.FormDefinition, error)
	DeleteFormFunc       func(ctx context.Context, id uuidv7.UUID) error
	ListFormsFunc        func(ctx context.Context, entityType string, limit, offset int) ([]*form.FormDefinition, int, error)
	ListAllFormsFunc     func(ctx context.Context, limit, offset int) ([]*form.FormDefinition, int, error)
}

func (m *MockFormUseCase) CreateForm(ctx context.Context, formEntity *form.FormDefinition) (*form.FormDefinition, error) {
	if m.CreateFormFunc != nil {
		return m.CreateFormFunc(ctx, formEntity)
	}
	return nil, errors.New("CreateFormFunc not implemented")
}

func (m *MockFormUseCase) GetForm(ctx context.Context, id uuidv7.UUID) (*form.FormDefinition, error) {
	if m.GetFormFunc != nil {
		return m.GetFormFunc(ctx, id)
	}
	return nil, errors.New("GetFormFunc not implemented")
}

func (m *MockFormUseCase) GetFormByFormID(ctx context.Context, formID string) (*form.FormDefinition, error) {
	if m.GetFormByFormIDFunc != nil {
		return m.GetFormByFormIDFunc(ctx, formID)
	}
	return nil, errors.New("GetFormByFormIDFunc not implemented")
}

func (m *MockFormUseCase) UpdateForm(ctx context.Context, id uuidv7.UUID, update *form.FormDefinition) (*form.FormDefinition, error) {
	if m.UpdateFormFunc != nil {
		return m.UpdateFormFunc(ctx, id, update)
	}
	return nil, errors.New("UpdateFormFunc not implemented")
}

func (m *MockFormUseCase) DeleteForm(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFormFunc != nil {
		return m.DeleteFormFunc(ctx, id)
	}
	return errors.New("DeleteFormFunc not implemented")
}

func (m *MockFormUseCase) ListForms(ctx context.Context, entityType string, limit, offset int) ([]*form.FormDefinition, int, error) {
	if m.ListFormsFunc != nil {
		return m.ListFormsFunc(ctx, entityType, limit, offset)
	}
	return nil, 0, errors.New("ListFormsFunc not implemented")
}

func (m *MockFormUseCase) ListAllForms(ctx context.Context, limit, offset int) ([]*form.FormDefinition, int, error) {
	if m.ListAllFormsFunc != nil {
		return m.ListAllFormsFunc(ctx, limit, offset)
	}
	return nil, 0, errors.New("ListAllFormsFunc not implemented")
}

func TestFormHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{
		CreateFormFunc: func(ctx context.Context, formEntity *form.FormDefinition) (*form.FormDefinition, error) {
			return formEntity, nil
		},
	}

	handler := formHTTP.NewFormHandler(mockUC)
	router.POST("/ui/forms", handler.Create)

	body := map[string]any{
		"form_id":     "customer_form",
		"entity_type": "customer",
		"name":        "Customer Form",
		"layout": map[string]any{
			"type": "single_column",
		},
		"fields": []map[string]any{
			{"name": "email", "type": "text"},
		},
	}

	w := smoke.MakeRequest(t, router, "POST", "/ui/forms", body)

	smoke.AssertSuccessResponse(t, w, 201)
}

func TestFormHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{}

	handler := formHTTP.NewFormHandler(mockUC)
	router.POST("/ui/forms", handler.Create)

	w := smoke.MakeRequest(t, router, "POST", "/ui/forms", map[string]any{})

	smoke.AssertErrorResponse(t, w, 400, "BAD_REQUEST")
}

func TestFormHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{
		GetFormFunc: func(ctx context.Context, id uuidv7.UUID) (*form.FormDefinition, error) {
			return fakeForm(t), nil
		},
	}

	handler := formHTTP.NewFormHandler(mockUC)
	router.GET("/ui/forms/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/ui/forms/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, w, 200)
}

func TestFormHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{
		GetFormFunc: func(ctx context.Context, id uuidv7.UUID) (*form.FormDefinition, error) {
			return nil, form.ErrFormNotFound
		},
	}

	handler := formHTTP.NewFormHandler(mockUC)
	router.GET("/ui/forms/:id", handler.GetByID)

	w := smoke.MakeRequest(t, router, "GET", "/ui/forms/"+smoke.FakeUUID(), nil)

	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

func TestFormHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{
		ListFormsFunc: func(ctx context.Context, entityType string, limit, offset int) ([]*form.FormDefinition, int, error) {
			return []*form.FormDefinition{fakeForm(t)}, 1, nil
		},
	}

	handler := formHTTP.NewFormHandler(mockUC)
	router.GET("/ui/forms", handler.List)

	w := smoke.MakeRequest(t, router, "GET", "/ui/forms?page=1&page_size=20&entity_type=customer", nil)

	smoke.AssertSuccessResponse(t, w, 200)
}

func TestFormHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{
		UpdateFormFunc: func(ctx context.Context, id uuidv7.UUID, update *form.FormDefinition) (*form.FormDefinition, error) {
			return fakeForm(t), nil
		},
	}

	handler := formHTTP.NewFormHandler(mockUC)
	router.PUT("/ui/forms/:id", handler.Update)

	body := map[string]any{
		"entity_type": "customer",
		"name":        "Customer Form",
		"layout": map[string]any{
			"type": "single_column",
		},
		"fields": []map[string]any{
			{"name": "email", "type": "text"},
		},
		"is_active": true,
	}

	w := smoke.MakeRequest(t, router, "PUT", "/ui/forms/"+smoke.FakeUUID(), body)

	smoke.AssertSuccessResponse(t, w, 200)
}

func TestFormHandler_Update_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{
		UpdateFormFunc: func(ctx context.Context, id uuidv7.UUID, update *form.FormDefinition) (*form.FormDefinition, error) {
			return nil, form.ErrFormNotFound
		},
	}

	handler := formHTTP.NewFormHandler(mockUC)
	router.PUT("/ui/forms/:id", handler.Update)

	body := map[string]any{
		"entity_type": "customer",
		"name":        "Customer Form",
		"layout": map[string]any{
			"type": "single_column",
		},
		"fields": []map[string]any{
			{"name": "email", "type": "text"},
		},
		"is_active": true,
	}

	w := smoke.MakeRequest(t, router, "PUT", "/ui/forms/"+smoke.FakeUUID(), body)

	smoke.AssertErrorResponse(t, w, 404, "NOT_FOUND")
}

func TestFormHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockFormUseCase{
		DeleteFormFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := formHTTP.NewFormHandler(mockUC)
	router.DELETE("/ui/forms/:id", handler.Delete)

	w := smoke.MakeRequest(t, router, "DELETE", "/ui/forms/"+smoke.FakeUUID(), nil)

	smoke.AssertSuccessResponse(t, w, 200)
}

func fakeForm(t *testing.T) *form.FormDefinition {
	t.Helper()

	entity, err := form.NewFormDefinition(
		"customer_form",
		"customer",
		"Customer Form",
		map[string]any{"type": "single_column"},
		[]map[string]any{{"name": "email", "type": "text"}},
	)
	assert.NoError(t, err)
	return entity
}