package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCurrencyExtendedTest() (*gin.Engine, *mocks.MockCurrencyUseCase) {
	gin.SetMode(gin.TestMode)
	mockUC := new(mocks.MockCurrencyUseCase)
	handler := NewCurrencyHandler(mockUC)

	router := gin.New()
	router.POST("/currencies", handler.Create)
	router.GET("/currencies/:id", handler.GetByID)
	router.GET("/currencies/code/:code", handler.GetByCode)
	router.GET("/currencies", handler.List)
	router.PUT("/currencies/:id", handler.Update)
	router.DELETE("/currencies/:id", handler.Delete)
	router.GET("/currencies/:id/countries", handler.GetCountries)
	router.POST("/currencies/:id/countries", handler.AddCountry)
	router.DELETE("/currencies/:id/countries/:country_id", handler.RemoveCountry)

	return router, mockUC
}

// Create tests
func TestCurrencyHandler_Create_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	req := dto.CreateCurrencyRequest{
		Code:   "USD",
		Name:   "US Dollar",
		Symbol: "$",
	}

	mockUC.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Currency) bool {
		return c.Code == req.Code && c.Name == req.Name
	})).Return(nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCurrencyHandler_Create_Conflict(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	req := dto.CreateCurrencyRequest{
		Code:   "USD",
		Name:   "US Dollar",
		Symbol: "$",
	}

	mockUC.On("Create", mock.Anything, mock.Anything).
		Return(errors.New("duplicate key value"))

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/currencies", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// GetByID tests
func TestCurrencyHandler_GetByID_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	currency := &entity.Currency{
		ID:     currencyID,
		Code:   "USD",
		Name:   "US Dollar",
		Symbol: "$",
	}

	mockUC.On("GetByID", mock.Anything, currencyID, false).
		Return(currency, nil)

	r := httptest.NewRequest(http.MethodGet, "/currencies/"+currencyID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCurrencyHandler_GetByID_NotFound(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	mockUC.On("GetByID", mock.Anything, currencyID, false).
		Return(nil, entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodGet, "/currencies/"+currencyID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// GetByCode tests
func TestCurrencyHandler_GetByCode_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currency := &entity.Currency{
		ID:     uuidv7.New(),
		Code:   "USD",
		Name:   "US Dollar",
		Symbol: "$",
	}

	mockUC.On("GetByCode", mock.Anything, "USD", false).
		Return(currency, nil)

	r := httptest.NewRequest(http.MethodGet, "/currencies/code/USD", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCurrencyHandler_GetByCode_NotFound(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	mockUC.On("GetByCode", mock.Anything, "XXX", false).
		Return(nil, entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodGet, "/currencies/code/XXX", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// List tests
func TestCurrencyHandler_List_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencies := []entity.Currency{
		{
			ID:     uuidv7.New(),
			Code:   "USD",
			Name:   "US Dollar",
			Symbol: "$",
		},
	}

	mockUC.On("List", mock.Anything, 1, 20, false).
		Return(currencies, 1, nil)

	r := httptest.NewRequest(http.MethodGet, "/currencies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

// Update tests
func TestCurrencyHandler_Update_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	req := dto.UpdateCurrencyRequest{
		Name:   "United States Dollar",
		Code:   "USD",
		Symbol: "US$",
	}

	mockUC.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.Currency) bool {
		return c.ID == currencyID && c.Name == req.Name
	})).Return(nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/currencies/"+currencyID.String(), bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCurrencyHandler_Update_NotFound(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	req := dto.UpdateCurrencyRequest{
		Name:   "United States Dollar",
		Code:   "USD",
		Symbol: "US$",
	}

	mockUC.On("Update", mock.Anything, mock.Anything).
		Return(entity.ErrNotFound)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/currencies/"+currencyID.String(), bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// Delete tests
func TestCurrencyHandler_Delete_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	mockUC.On("Delete", mock.Anything, currencyID).Return(nil)

	r := httptest.NewRequest(http.MethodDelete, "/currencies/"+currencyID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCurrencyHandler_Delete_NotFound(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	mockUC.On("Delete", mock.Anything, currencyID).
		Return(entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodDelete, "/currencies/"+currencyID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// GetCountries tests
func TestCurrencyHandler_GetCountries_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	countries := []entity.Country{
		{
			ID:     uuidv7.New(),
			Code:   "+1",
			Name:   "United States",
			ISO2:   "US",
			ISO3:   "USA",
			Region: "north_america",
		},
	}

	mockUC.On("GetCountries", mock.Anything, currencyID).
		Return(countries, nil)

	r := httptest.NewRequest(http.MethodGet, "/currencies/"+currencyID.String()+"/countries", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCurrencyHandler_GetCountries_NotFound(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	mockUC.On("GetCountries", mock.Anything, currencyID).
		Return(nil, entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodGet, "/currencies/"+currencyID.String()+"/countries", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// AddCountry tests
func TestCurrencyHandler_AddCountry_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()

	req := dto.AddCountryToCurrencyRequest{
		CountryID: countryID.String(),
		IsPrimary: true,
	}

	mockUC.On("AddCountry", mock.Anything, currencyID, countryID, true).
		Return(nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/currencies/"+currencyID.String()+"/countries", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

// RemoveCountry tests
func TestCurrencyHandler_RemoveCountry_Success(t *testing.T) {
	router, mockUC := setupCurrencyExtendedTest()

	currencyID := uuidv7.New()
	countryID := uuidv7.New()

	mockUC.On("RemoveCountry", mock.Anything, currencyID, countryID).
		Return(nil)

	r := httptest.NewRequest(http.MethodDelete, "/currencies/"+currencyID.String()+"/countries/"+countryID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}
