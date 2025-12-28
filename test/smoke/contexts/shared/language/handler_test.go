package language_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	languageHTTP "github.com/basilex/promenade/internal/contexts/shared/language/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockLanguageUseCase - minimal mock for smoke tests
type MockLanguageUseCase struct {
	mock.Mock
}

func (m *MockLanguageUseCase) List(ctx context.Context) ([]*language.Language, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*language.Language), args.Error(1)
}

func (m *MockLanguageUseCase) GetByCode(ctx context.Context, code string) (*language.Language, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*language.Language), args.Error(1)
}

func (m *MockLanguageUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*language.Language, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*language.Language), args.Error(1)
}

func (m *MockLanguageUseCase) Create(ctx context.Context, l *language.Language) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockLanguageUseCase) Update(ctx context.Context, l *language.Language) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockLanguageUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupLanguageRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestLanguageHandler_Smoke - smoke tests for Language handler
func TestLanguageHandler_Smoke(t *testing.T) {
	mockUC := new(MockLanguageUseCase)
	handler := languageHTTP.NewHandler(mockUC)
	router := setupLanguageRouter()

	// Register routes
	router.GET("/languages", handler.ListLanguages)
	router.GET("/languages/code/:code", handler.GetLanguageByCode)
	router.POST("/languages", handler.CreateLanguage)
	router.PUT("/languages/:id", handler.UpdateLanguage)
	router.DELETE("/languages/:id", handler.DeleteLanguage)

	t.Run("List returns 200", func(t *testing.T) {
		languages := []*language.Language{
			{ID: uuidv7.New(), Code: "en", Name: "English", NativeName: "English", IsActive: true},
			{ID: uuidv7.New(), Code: "uk", Name: "Ukrainian", NativeName: "Українська", IsActive: true},
		}
		mockUC.On("List", mock.Anything).Return(languages, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/languages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "List should return 200")
	})

	t.Run("GetByCode returns 200", func(t *testing.T) {
		l := &language.Language{
			ID:         uuidv7.New(),
			Code:       "uk",
			Name:       "Ukrainian",
			NativeName: "Українська",
			IsActive:   true,
		}
		mockUC.On("GetByCode", mock.Anything, "uk").Return(l, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/languages/code/uk", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "GetByCode should return 200")
	})

	t.Run("Create returns 201", func(t *testing.T) {
		reqBody := languageHTTP.CreateLanguageRequest{
			Code:       "pl",
			Name:       "Polish",
			NativeName: "Polski",
		}
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*language.Language")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/languages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Create should return 201")
	})

	t.Run("Update returns 200", func(t *testing.T) {
		id := uuidv7.New()
		existing := &language.Language{
			ID:       id,
			Code:     "pl",
			Name:     "Polish",
			IsActive: true,
		}
		reqBody := languageHTTP.UpdateLanguageRequest{
			Name:       "Polish Updated",
			NativeName: "Polski",
			IsActive:   true,
		}
		mockUC.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, mock.AnythingOfType("*language.Language")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/languages/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Update should return 200")
	})

	t.Run("Delete returns 204", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/languages/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code, "Delete should return 204")
	})
}
