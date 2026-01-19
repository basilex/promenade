package language_test

import (
	"context"
	"testing"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	languageHTTP "github.com/basilex/promenade/internal/contexts/shared/language/adapter/http"
	languageAggregate "github.com/basilex/promenade/internal/contexts/shared/language/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockLanguageUseCase is a mock implementation of language.IUseCase for testing
type MockLanguageUseCase struct {
	GetByIDFunc   func(ctx context.Context, id uuidv7.UUID) (*languageAggregate.Language, error)
	GetByCodeFunc func(ctx context.Context, code string) (*languageAggregate.Language, error)
	ListFunc      func(ctx context.Context) ([]*languageAggregate.Language, error)
	CreateFunc    func(ctx context.Context, language *languageAggregate.Language) error
	UpdateFunc    func(ctx context.Context, language *languageAggregate.Language) error
	DeleteFunc    func(ctx context.Context, id uuidv7.UUID) error
}

func (m *MockLanguageUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*languageAggregate.Language, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockLanguageUseCase) GetByCode(ctx context.Context, code string) (*languageAggregate.Language, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	return nil, nil
}

func (m *MockLanguageUseCase) List(ctx context.Context) ([]*languageAggregate.Language, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockLanguageUseCase) Create(ctx context.Context, l *languageAggregate.Language) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, l)
	}
	return nil
}

func (m *MockLanguageUseCase) Update(ctx context.Context, l *languageAggregate.Language) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, l)
	}
	return nil
}

func (m *MockLanguageUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// fakeLanguage creates a fake language for testing
func fakeLanguage() *languageAggregate.Language {
	l, _ := languageAggregate.NewLanguage("en", "English", "English")
	return l
}

func TestLanguageHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		CreateFunc: func(ctx context.Context, l *languageAggregate.Language) error {
			return nil
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.POST("/languages", handler.CreateLanguage)

	body := map[string]any{
		"code":        "en",
		"name":        "English",
		"native_name": "English",
	}

	w := smoke.MakeRequest(t, router, "POST", "/languages", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestLanguageHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{}
	handler := languageHTTP.NewHandler(mockUC)
	router.POST("/languages", handler.CreateLanguage)

	body := map[string]any{
		"code": "en",
		// Missing required fields
	}

	w := smoke.MakeRequest(t, router, "POST", "/languages", body)
	smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
}

func TestLanguageHandler_GetByCode_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		GetByCodeFunc: func(ctx context.Context, code string) (*languageAggregate.Language, error) {
			return fakeLanguage(), nil
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.GET("/languages/:code", handler.GetLanguageByCode)

	w := smoke.MakeRequest(t, router, "GET", "/languages/en", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestLanguageHandler_GetByCode_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		GetByCodeFunc: func(ctx context.Context, code string) (*languageAggregate.Language, error) {
			return nil, language.ErrLanguageNotFound
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.GET("/languages/:code", handler.GetLanguageByCode)

	w := smoke.MakeRequest(t, router, "GET", "/languages/xx", nil)
	smoke.AssertErrorResponse(t, w, 404, "LANGUAGE_NOT_FOUND")
}

func TestLanguageHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		ListFunc: func(ctx context.Context) ([]*languageAggregate.Language, error) {
			return []*languageAggregate.Language{fakeLanguage()}, nil
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.GET("/languages", handler.ListLanguages)

	w := smoke.MakeRequest(t, router, "GET", "/languages", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestLanguageHandler_List_EmptyResult(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		ListFunc: func(ctx context.Context) ([]*languageAggregate.Language, error) {
			return []*languageAggregate.Language{}, nil
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.GET("/languages", handler.ListLanguages)

	w := smoke.MakeRequest(t, router, "GET", "/languages", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestLanguageHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*languageAggregate.Language, error) {
			return fakeLanguage(), nil
		},
		UpdateFunc: func(ctx context.Context, l *languageAggregate.Language) error {
			return nil
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.PUT("/languages/:id", handler.UpdateLanguage)

	body := map[string]any{
		"name":        "English (US)",
		"native_name": "English",
	}

	w := smoke.MakeRequest(t, router, "PUT", "/languages/"+smoke.FakeUUID(), body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestLanguageHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.DELETE("/languages/:id", handler.DeleteLanguage)

	w := smoke.MakeRequest(t, router, "DELETE", "/languages/"+smoke.FakeUUID(), nil)

	// Delete returns 204 No Content
	if w.Code != 204 {
		t.Errorf("expected status code 204, got %d", w.Code)
	}
}

func TestLanguageHandler_Delete_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLanguageUseCase{
		DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return language.ErrLanguageNotFound
		},
	}

	handler := languageHTTP.NewHandler(mockUC)
	router.DELETE("/languages/:id", handler.DeleteLanguage)

	w := smoke.MakeRequest(t, router, "DELETE", "/languages/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, 500, "DELETE_ERROR")
}
