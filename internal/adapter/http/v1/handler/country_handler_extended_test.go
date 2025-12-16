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
	"github.com/basilex/promenade/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCountryExtendedTest() (*gin.Engine, *mocks.MockCountryUseCase) {
	gin.SetMode(gin.TestMode)
	mockUC := new(mocks.MockCountryUseCase)
	handler := NewCountryHandler(mockUC)

	router := gin.New()
	router.POST("/countries", handler.Create)
	router.GET("/countries/:id", handler.GetByID)
	router.GET("/countries/code/:code", handler.GetByCode)
	router.GET("/countries", handler.List)
	router.PUT("/countries/:id", handler.Update)
	router.DELETE("/countries/:id", handler.Delete)
	router.GET("/countries/:id/currencies", handler.GetCurrencies)
	router.POST("/countries/:id/currencies", handler.AddCurrency)
	router.DELETE("/countries/:id/currencies/:currency_id", handler.RemoveCurrency)

	return router, mockUC
}

// Create tests
func TestCountryHandler_Create_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	req := dto.CreateCountryRequest{
		Code:   "+1",
		Name:   "United States",
		ISO2:   "US",
		ISO3:   "USA",
		Region: "north_america",
	}

	mockUC.On("Create", mock.Anything, mock.MatchedBy(func(c *entity.Country) bool {
		return c.Code == req.Code && c.Name == req.Name
	})).Return(nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/countries", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCountryHandler_Create_Conflict(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	req := dto.CreateCountryRequest{
		Code:   "+1",
		Name:   "United States",
		ISO2:   "US",
		ISO3:   "USA",
		Region: "north_america",
	}

	mockUC.On("Create", mock.Anything, mock.Anything).
		Return(errors.New("duplicate key value"))

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/countries", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// GetByID tests
func TestCountryHandler_GetByID_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	country := &entity.Country{
		ID:     countryID,
		Code:   "+1",
		Name:   "United States",
		ISO2:   "US",
		ISO3:   "USA",
		Region: "north_america",
	}

	mockUC.On("GetByID", mock.Anything, countryID, false).
		Return(country, nil)

	r := httptest.NewRequest(http.MethodGet, "/countries/"+countryID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCountryHandler_GetByID_NotFound(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	mockUC.On("GetByID", mock.Anything, countryID, false).
		Return(nil, entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodGet, "/countries/"+countryID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// GetByCode tests
func TestCountryHandler_GetByCode_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	country := &entity.Country{
		ID:     uuid.New(),
		Code:   "+1",
		Name:   "United States",
		ISO2:   "US",
		ISO3:   "USA",
		Region: "north_america",
	}

	mockUC.On("GetByCode", mock.Anything, "+1", false).
		Return(country, nil)

	r := httptest.NewRequest(http.MethodGet, "/countries/code/+1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCountryHandler_GetByCode_NotFound(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	mockUC.On("GetByCode", mock.Anything, "+999", false).
		Return(nil, entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodGet, "/countries/code/+999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// List tests
func TestCountryHandler_List_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countries := []entity.Country{
		{
			ID:     uuid.New(),
			Code:   "+1",
			Name:   "United States",
			ISO2:   "US",
			ISO3:   "USA",
			Region: "north_america",
		},
	}

	mockUC.On("List", mock.Anything, 1, 20, false).
		Return(countries, 1, nil)

	r := httptest.NewRequest(http.MethodGet, "/countries", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCountryHandler_ListByRegion_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countries := []entity.Country{
		{
			ID:     uuid.New(),
			Code:   "+1",
			Name:   "United States",
			ISO2:   "US",
			ISO3:   "USA",
			Region: "north_america",
		},
	}

	mockUC.On("ListByRegion", mock.Anything, "north_america", 1, 20, false).
		Return(countries, 1, nil)

	r := httptest.NewRequest(http.MethodGet, "/countries?region=north_america", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

// Update tests
func TestCountryHandler_Update_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	req := dto.UpdateCountryRequest{
		Name:   "United States of America",
		Code:   "+1",
		ISO2:   "US",
		ISO3:   "USA",
		Region: "north_america",
	}

	mockUC.On("Update", mock.Anything, mock.MatchedBy(func(c *entity.Country) bool {
		return c.ID == countryID && c.Name == req.Name
	})).Return(nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/countries/"+countryID.String(), bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCountryHandler_Update_NotFound(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	req := dto.UpdateCountryRequest{
		Name:   "United States of America",
		Code:   "+1",
		ISO2:   "US",
		ISO3:   "USA",
		Region: "north_america",
	}

	mockUC.On("Update", mock.Anything, mock.Anything).
		Return(entity.ErrNotFound)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPut, "/countries/"+countryID.String(), bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// Delete tests
func TestCountryHandler_Delete_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	mockUC.On("Delete", mock.Anything, countryID).Return(nil)

	r := httptest.NewRequest(http.MethodDelete, "/countries/"+countryID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCountryHandler_Delete_NotFound(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	mockUC.On("Delete", mock.Anything, countryID).
		Return(entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodDelete, "/countries/"+countryID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockUC.AssertExpectations(t)
}

// GetCurrencies tests
func TestCountryHandler_GetCurrencies_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	currencies := []entity.Currency{
		{
			ID:     uuid.New(),
			Code:   "USD",
			Name:   "US Dollar",
			Symbol: "$",
		},
	}

	mockUC.On("GetCurrencies", mock.Anything, countryID).
		Return(currencies, nil)

	r := httptest.NewRequest(http.MethodGet, "/countries/"+countryID.String()+"/currencies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

func TestCountryHandler_GetCurrencies_NotFound(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	mockUC.On("GetCurrencies", mock.Anything, countryID).
		Return(nil, entity.ErrNotFound)

	r := httptest.NewRequest(http.MethodGet, "/countries/"+countryID.String()+"/currencies", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUC.AssertExpectations(t)
}

// AddCurrency tests
func TestCountryHandler_AddCurrency_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	currencyID := uuid.New()

	req := dto.AddCurrencyToCountryRequest{
		CurrencyID: currencyID.String(),
		IsPrimary:  true,
	}

	mockUC.On("AddCurrency", mock.Anything, countryID, currencyID, true).
		Return(nil)

	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/countries/"+countryID.String()+"/currencies", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}

// RemoveCurrency tests
func TestCountryHandler_RemoveCurrency_Success(t *testing.T) {
	router, mockUC := setupCountryExtendedTest()

	countryID := uuid.New()
	currencyID := uuid.New()

	mockUC.On("RemoveCurrency", mock.Anything, countryID, currencyID).
		Return(nil)

	r := httptest.NewRequest(http.MethodDelete, "/countries/"+countryID.String()+"/currencies/"+currencyID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockUC.AssertExpectations(t)
}
