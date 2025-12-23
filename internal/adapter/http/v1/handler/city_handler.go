package handler

import (
	"net/http"
	"time"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

type CityHandler struct {
	cityUseCase usecase.CityUseCase
}

func NewCityHandler(cityUseCase usecase.CityUseCase) *CityHandler {
	return &CityHandler{
		cityUseCase: cityUseCase,
	}
}

// Create creates a new city
// @Summary Create city
// @Tags cities
// @Accept json
// @Produce json
// @Param request body dto.CreateCityRequest true "City data"
// @Success 201 {object} response.Response{data=dto.CityResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cities [post]
func (h *CityHandler) Create(c *gin.Context) {
	var req dto.CreateCityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	countryID, err := uuidv7.Parse(req.CountryID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	var regionID *uuidv7.UUID
	if req.RegionID != nil {
		parsedRegionID, err := uuidv7.Parse(*req.RegionID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid region ID", err)
			return
		}
		regionID = &parsedRegionID
	}

	now := time.Now()
	city := &entity.City{
		ID:                uuidv7.New(),
		CountryID:         countryID,
		RegionID:          regionID,
		Name:              req.Name,
		NameLocal:         req.NameLocal,
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		Population:        req.Population,
		IsCapital:         false,
		IsRegionalCapital: false,
		IsActive:          true,
		SortOrder:         0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if req.IsCapital != nil {
		city.IsCapital = *req.IsCapital
	}
	if req.IsRegionalCapital != nil {
		city.IsRegionalCapital = *req.IsRegionalCapital
	}
	if req.IsActive != nil {
		city.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		city.SortOrder = *req.SortOrder
	}

	if err := h.cityUseCase.Create(c.Request.Context(), city); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create city", err)
		return
	}

	resp := h.toCityResponse(city)
	response.Success(c, http.StatusCreated, resp)
}

// GetByID retrieves a city by ID
// @Summary Get city by ID
// @Tags cities
// @Produce json
// @Param id path string true "City ID" format(uuid)
// @Success 200 {object} response.Response{data=dto.CityResponse}
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "City not found"
// @Failure 500 {object} response.Response
// @Router /cities/{id} [get]
func (h *CityHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid city ID", err)
		return
	}

	city, err := h.cityUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "city not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get city", err)
		return
	}

	resp := h.toCityResponse(city)
	response.Success(c, http.StatusOK, resp)
}

// List retrieves all cities with pagination
// @Summary List cities
// @Tags cities
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.CityResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cities [get]
func (h *CityHandler) List(c *gin.Context) {
	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	cities, total, err := h.cityUseCase.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list cities", err)
		return
	}

	items := make([]dto.CityResponse, len(cities))
	for i, city := range cities {
		items[i] = *h.toCityResponse(&city)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// ListByCountry retrieves cities for a specific country
// @Summary List cities by country
// @Tags cities
// @Produce json
// @Param country_id path string true "Country ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.CityResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id}/cities [get]
func (h *CityHandler) ListByCountry(c *gin.Context) {
	countryID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	cities, total, err := h.cityUseCase.ListByCountry(c.Request.Context(), countryID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list cities", err)
		return
	}

	items := make([]dto.CityResponse, len(cities))
	for i, city := range cities {
		items[i] = *h.toCityResponse(&city)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// ListByRegion retrieves cities for a specific region
// @Summary List cities by region
// @Tags cities
// @Produce json
// @Param region_id path string true "Region ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.CityResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /regions/{id}/cities [get]
func (h *CityHandler) ListByRegion(c *gin.Context) {
	regionID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid region ID", err)
		return
	}

	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	cities, total, err := h.cityUseCase.ListByRegion(c.Request.Context(), regionID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list cities", err)
		return
	}

	items := make([]dto.CityResponse, len(cities))
	for i, city := range cities {
		items[i] = *h.toCityResponse(&city)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// ListCapitals retrieves all capital cities
// @Summary List capital cities
// @Tags cities
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.CityResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cities/capitals [get]
func (h *CityHandler) ListCapitals(c *gin.Context) {
	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	cities, total, err := h.cityUseCase.ListCapitals(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list capital cities", err)
		return
	}

	items := make([]dto.CityResponse, len(cities))
	for i, city := range cities {
		items[i] = *h.toCityResponse(&city)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// Search searches cities by name
// @Summary Search cities by name
// @Tags cities
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.CityResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /cities/search [get]
func (h *CityHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "search query is required", nil)
		return
	}

	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	cities, total, err := h.cityUseCase.SearchByName(c.Request.Context(), query, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to search cities", err)
		return
	}

	items := make([]dto.CityResponse, len(cities))
	for i, city := range cities {
		items[i] = *h.toCityResponse(&city)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// Update updates an existing city
// @Summary Update city
// @Tags cities
// @Accept json
// @Produce json
// @Param id path string true "City ID" format(uuid)
// @Param request body dto.UpdateCityRequest true "City data"
// @Success 200 {object} response.Response{data=dto.CityResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response "City not found"
// @Failure 500 {object} response.Response
// @Router /cities/{id} [put]
func (h *CityHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid city ID", err)
		return
	}

	var req dto.UpdateCityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Get existing city
	city, err := h.cityUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "city not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get city", err)
		return
	}

	// Parse region ID if provided
	if req.RegionID != nil {
		parsedRegionID, err := uuidv7.Parse(*req.RegionID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid region ID", err)
			return
		}
		city.RegionID = &parsedRegionID
	} else {
		city.RegionID = nil
	}

	// Update fields
	city.Name = req.Name
	city.NameLocal = req.NameLocal
	city.Latitude = req.Latitude
	city.Longitude = req.Longitude
	city.Population = req.Population
	city.UpdatedAt = time.Now()

	if req.IsCapital != nil {
		city.IsCapital = *req.IsCapital
	}
	if req.IsRegionalCapital != nil {
		city.IsRegionalCapital = *req.IsRegionalCapital
	}
	if req.IsActive != nil {
		city.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		city.SortOrder = *req.SortOrder
	}

	if err := h.cityUseCase.Update(c.Request.Context(), city); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update city", err)
		return
	}

	resp := h.toCityResponse(city)
	response.Success(c, http.StatusOK, resp)
}

// Delete deletes a city by ID
// @Summary Delete city
// @Tags cities
// @Produce json
// @Param id path string true "City ID" format(uuid)
// @Success 204
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "City not found"
// @Failure 500 {object} response.Response
// @Router /cities/{id} [delete]
func (h *CityHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid city ID", err)
		return
	}

	if err := h.cityUseCase.Delete(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "city not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete city", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CityHandler) toCityResponse(city *entity.City) *dto.CityResponse {
	return &dto.CityResponse{
		ID:                city.ID,
		RegionID:          city.RegionID,
		CountryID:         city.CountryID,
		Name:              city.Name,
		NameLocal:         city.NameLocal,
		Latitude:          city.Latitude,
		Longitude:         city.Longitude,
		Population:        city.Population,
		IsCapital:         city.IsCapital,
		IsRegionalCapital: city.IsRegionalCapital,
		IsActive:          city.IsActive,
		SortOrder:         city.SortOrder,
		CreatedAt:         city.CreatedAt,
		UpdatedAt:         city.UpdatedAt,
	}
}
