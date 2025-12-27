package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MockUseCase struct {
	mock.Mock
}

func (m *MockUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*language.Language, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*language.Language), args.Error(1)
}

func (m *MockUseCase) GetByCode(ctx context.Context, code string) (*language.Language, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*language.Language), args.Error(1)
}

func (m *MockUseCase) List(ctx context.Context) ([]*language.Language, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*language.Language), args.Error(1)
}

func (m *MockUseCase) Create(ctx context.Context, l *language.Language) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockUseCase) Update(ctx context.Context, l *language.Language) error {
	args := m.Called(ctx, l)
	return args.Error(0)
}

func (m *MockUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestHandler_ListLanguages(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/languages", handler.ListLanguages)

	t.Run("success", func(t *testing.T) {
		languages := []*language.Language{
			{ID: uuidv7.New(), Code: "en", Code3: "eng", Name: "English", NativeName: "English", IsActive: true},
			{ID: uuidv7.New(), Code: "uk", Code3: "ukr", Name: "Ukrainian", NativeName: "Українська", IsActive: true},
		}
		mockUC.On("List", mock.Anything).Return(languages, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/languages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		
		mockUC.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockUC.On("List", mock.Anything).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/languages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_GetLanguageByCode(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.GET("/languages/code/:code", handler.GetLanguageByCode)

	t.Run("success", func(t *testing.T) {
		l := &language.Language{
			ID:         uuidv7.New(),
			Code:       "en",
			Code3:      "eng",
			Name:       "English",
			NativeName: "English",
			IsActive:   true,
		}
		mockUC.On("GetByCode", mock.Anything, "en").Return(l, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/languages/code/en", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		
		mockUC.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockUC.On("GetByCode", mock.Anything, "xx").Return(nil, language.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/languages/code/xx", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_CreateLanguage(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.POST("/languages", handler.CreateLanguage)

	t.Run("success", func(t *testing.T) {
		reqBody := CreateLanguageRequest{
			Code:       "en",
			Code3:      "eng",
			Name:       "English",
			NativeName: "English",
		}
		mockUC.On("Create", mock.Anything, mock.AnythingOfType("*language.Language")).Return(nil).Once()

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/languages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		reqBody := CreateLanguageRequest{
			Code: "", // Empty code
			Name: "English",
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/languages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_UpdateLanguage(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.PUT("/languages/:id", handler.UpdateLanguage)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		existing := &language.Language{
			ID:         id,
			Code:       "en",
			Name:       "English",
			NativeName: "English",
		}
		
		mockUC.On("GetByID", mock.Anything, id).Return(existing, nil).Once()
		mockUC.On("Update", mock.Anything, mock.AnythingOfType("*language.Language")).Return(nil).Once()

		reqBody := UpdateLanguageRequest{
			Name:       "English (US)",
			NativeName: "English (US)",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/languages/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("GetByID", mock.Anything, id).Return(nil, language.ErrNotFound).Once()

		reqBody := UpdateLanguageRequest{
			Name: "Test",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/languages/"+id.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}

func TestHandler_DeleteLanguage(t *testing.T) {
	mockUC := new(MockUseCase)
	handler := NewHandler(mockUC)
	router := setupTestRouter()
	router.DELETE("/languages/:id", handler.DeleteLanguage)

	t.Run("success", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/languages/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/languages/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		id := uuidv7.New()
		mockUC.On("Delete", mock.Anything, id).Return(language.ErrNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/languages/"+id.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUC.AssertExpectations(t)
	})
}
