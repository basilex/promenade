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

type RegionHandler struct {
	regionUseCase usecase.RegionUseCase
}

func NewRegionHandler(regionUseCase usecase.RegionUseCase) *RegionHandler {
	return &RegionHandler{
		regionUseCase: regionUseCase,
	}
}

// Create creates a new region
// @Summary Create region
// @Tags regions
// @Accept json
// @Produce json
// @Param request body dto.CreateRegionRequest true "Region data"
// @Success 201 {object} response.Response{data=dto.RegionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /regions [post]
func (h *RegionHandler) Create(c *gin.Context) {
	var req dto.CreateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	countryID, err := uuidv7.Parse(req.CountryID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	now := time.Now()
	region := &entity.Region{
		ID:         uuidv7.New(),
		CountryID:  countryID,
		Name:       req.Name,
		Code:       req.Code,
		RegionType: req.RegionType,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		IsActive:   true,
		SortOrder:  0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if req.IsActive != nil {
		region.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		region.SortOrder = *req.SortOrder
	}

	if err := h.regionUseCase.Create(c.Request.Context(), region); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create region", err)
		return
	}

	resp := h.toRegionResponse(region)
	response.Success(c, http.StatusCreated, resp)
}

// GetByID retrieves a region by ID
// @Summary Get region by ID
// @Tags regions
// @Produce json
// @Param id path string true "Region ID" format(uuid)
// @Success 200 {object} response.Response{data=dto.RegionResponse}
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Region not found"
// @Failure 500 {object} response.Response
// @Router /regions/{id} [get]
func (h *RegionHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid region ID", err)
		return
	}

	region, err := h.regionUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "region not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get region", err)
		return
	}

	resp := h.toRegionResponse(region)
	response.Success(c, http.StatusOK, resp)
}

// List retrieves all regions with pagination
// @Summary List regions
// @Tags regions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.RegionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /regions [get]
func (h *RegionHandler) List(c *gin.Context) {
	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	regions, total, err := h.regionUseCase.List(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list regions", err)
		return
	}

	items := make([]dto.RegionResponse, len(regions))
	for i, region := range regions {
		items[i] = *h.toRegionResponse(&region)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// ListByCountry retrieves regions for a specific country
// @Summary List regions by country
// @Tags regions
// @Produce json
// @Param country_id path string true "Country ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.RegionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /countries/{id}/regions [get]
func (h *RegionHandler) ListByCountry(c *gin.Context) {
	countryID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid country ID", err)
		return
	}

	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	regions, total, err := h.regionUseCase.ListByCountry(c.Request.Context(), countryID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list regions", err)
		return
	}

	items := make([]dto.RegionResponse, len(regions))
	for i, region := range regions {
		items[i] = *h.toRegionResponse(&region)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// ListActive retrieves only active regions
// @Summary List active regions
// @Tags regions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.PaginatedResponse{data=[]dto.RegionResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /regions/active [get]
func (h *RegionHandler) ListActive(c *gin.Context) {
	page := response.GetPageFromQuery(c)
	pageSize := response.GetPageSizeFromQuery(c)

	regions, total, err := h.regionUseCase.ListActive(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list active regions", err)
		return
	}

	items := make([]dto.RegionResponse, len(regions))
	for i, region := range regions {
		items[i] = *h.toRegionResponse(&region)
	}

	resp := response.NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(http.StatusOK, resp)
}

// Update updates an existing region
// @Summary Update region
// @Tags regions
// @Accept json
// @Produce json
// @Param id path string true "Region ID" format(uuid)
// @Param request body dto.UpdateRegionRequest true "Region data"
// @Success 200 {object} response.Response{data=dto.RegionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response "Region not found"
// @Failure 500 {object} response.Response
// @Router /regions/{id} [put]
func (h *RegionHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid region ID", err)
		return
	}

	var req dto.UpdateRegionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// Get existing region
	region, err := h.regionUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "region not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get region", err)
		return
	}

	// Update fields
	region.Name = req.Name
	region.Code = req.Code
	region.RegionType = req.RegionType
	region.Latitude = req.Latitude
	region.Longitude = req.Longitude
	region.UpdatedAt = time.Now()

	if req.IsActive != nil {
		region.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		region.SortOrder = *req.SortOrder
	}

	if err := h.regionUseCase.Update(c.Request.Context(), region); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update region", err)
		return
	}

	resp := h.toRegionResponse(region)
	response.Success(c, http.StatusOK, resp)
}

// Delete deletes a region by ID
// @Summary Delete region
// @Tags regions
// @Produce json
// @Param id path string true "Region ID" format(uuid)
// @Success 204
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Region not found"
// @Failure 500 {object} response.Response
// @Router /regions/{id} [delete]
func (h *RegionHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid region ID", err)
		return
	}

	if err := h.regionUseCase.Delete(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "region not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete region", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *RegionHandler) toRegionResponse(region *entity.Region) *dto.RegionResponse {
	return &dto.RegionResponse{
		ID:         region.ID,
		CountryID:  region.CountryID,
		Name:       region.Name,
		Code:       region.Code,
		RegionType: region.RegionType,
		Latitude:   region.Latitude,
		Longitude:  region.Longitude,
		IsActive:   region.IsActive,
		SortOrder:  region.SortOrder,
		CreatedAt:  region.CreatedAt,
		UpdatedAt:  region.UpdatedAt,
	}
}
