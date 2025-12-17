package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CurrencyHandler{}
	router.POST("/currencies", handler.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/currencies", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "error")
}

func TestCurrencyHandler_GetByID_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CurrencyHandler{}
	router.GET("/currencies/:id", handler.GetByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/currencies/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "error")
}

func TestCurrencyHandler_Update_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CurrencyHandler{}
	router.PUT("/currencies/:id", handler.Update)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/currencies/invalid-uuid", bytes.NewBufferString("{}"))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCurrencyHandler_Delete_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CurrencyHandler{}
	router.DELETE("/currencies/:id", handler.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/currencies/invalid-uuid", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCurrencyHandler_GetCountries_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CurrencyHandler{}
	router.GET("/currencies/:id/countries", handler.GetCountries)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/currencies/invalid-uuid/countries", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCurrencyHandler_AddCountry_InvalidCurrencyUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CurrencyHandler{}
	router.POST("/currencies/:id/countries", handler.AddCountry)

	reqBody := map[string]string{
		"country_id": "some-valid-looking-uuid-but-currency-id-is-invalid",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/currencies/invalid-uuid/countries", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCurrencyHandler_RemoveCountry_InvalidCurrencyUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	handler := &CurrencyHandler{}
	router.DELETE("/currencies/:id/countries/:country_id", handler.RemoveCountry)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/currencies/invalid-uuid/countries/some-country-id", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
