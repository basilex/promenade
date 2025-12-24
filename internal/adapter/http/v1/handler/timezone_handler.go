package handler

import (
	"net/http"
	"time"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/gin-gonic/gin"
)

type TimezoneHandler struct {
	timezoneUseCase usecase.ITimezoneUseCase
}

func NewTimezoneHandler(timezoneUseCase usecase.ITimezoneUseCase) *TimezoneHandler {
	return &TimezoneHandler{
		timezoneUseCase: timezoneUseCase,
	}
}

// Create creates a new timezone
// @Summary Create timezone
// @Tags timezones
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateTimezoneRequest true "Timezone data"
// @Success 201 {object} response.Response{data=dto.TimezoneResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response "Timezone already exists"
// @Failure 500 {object} response.Response
// @Router /v1/timezones [post]
func (h *TimezoneHandler) Create(c *gin.Context) {
	var req dto.CreateTimezoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	timezone := entity.NewTimezone(
		req.Name,
		req.Abbreviation,
		req.UtcOffset,
		req.UtcDstOffset,
		req.Description,
	)
	timezone.IsActive = req.IsActive

	if err := h.timezoneUseCase.Create(c.Request.Context(), timezone); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create timezone", err)
		return
	}

	resp := h.toTimezoneResponse(timezone)
	response.Success(c, http.StatusCreated, resp)
}

// GetByID retrieves a timezone by ID
// @Summary Get timezone by ID
// @Tags timezones
// @Produce json
// @Param id path string true "Timezone ID" format(uuid)
// @Success 200 {object} response.Response{data=dto.TimezoneResponse}
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Timezone not found"
// @Failure 500 {object} response.Response
// @Router /v1/timezones/{id} [get]
func (h *TimezoneHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid timezone ID", err)
		return
	}

	timezone, err := h.timezoneUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "timezone not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get timezone", err)
		return
	}

	resp := h.toTimezoneResponse(timezone)
	response.Success(c, http.StatusOK, resp)
}

// GetByName retrieves a timezone by name
// @Summary Get timezone by name
// @Tags timezones
// @Produce json
// @Param name path string true "Timezone name" example(Europe/Moscow)
// @Success 200 {object} response.Response{data=dto.TimezoneResponse}
// @Failure 404 {object} response.Response "Timezone not found"
// @Failure 500 {object} response.Response
// @Router /v1/timezones/name/{name} [get]
func (h *TimezoneHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	timezone, err := h.timezoneUseCase.GetByName(c.Request.Context(), name)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "timezone not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get timezone", err)
		return
	}

	resp := h.toTimezoneResponse(timezone)
	response.Success(c, http.StatusOK, resp)
}

// List retrieves all timezones with pagination
// @Summary List timezones
// @Tags timezones
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]dto.TimezoneResponse}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/timezones [get]
func (h *TimezoneHandler) List(c *gin.Context) {
	params := pagination.Params{
		Page:     response.GetPageFromQuery(c),
		PageSize: response.GetPageSizeFromQuery(c),
	}

	timezones, metadata, err := h.timezoneUseCase.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list timezones", err)
		return
	}

	timezoneResponses := make([]*dto.TimezoneResponse, len(timezones))
	for i, tz := range timezones {
		timezoneResponses[i] = h.toTimezoneResponse(tz)
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Items:      timezoneResponses,
		Page:       metadata.CurrentPage,
		PageSize:   metadata.Limit,
		Total:      metadata.Total,
		TotalPages: metadata.TotalPages,
		Timestamp:  time.Now().Unix(),
	})
}

// ListActive retrieves all active timezones
// @Summary List active timezones
// @Tags timezones
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.TimezoneResponse}
// @Failure 500 {object} response.Response
// @Router /v1/timezones/active [get]
func (h *TimezoneHandler) ListActive(c *gin.Context) {
	timezones, err := h.timezoneUseCase.ListActive(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list active timezones", err)
		return
	}

	timezoneResponses := make([]*dto.TimezoneResponse, len(timezones))
	for i, tz := range timezones {
		timezoneResponses[i] = h.toTimezoneResponse(tz)
	}

	response.Success(c, http.StatusOK, timezoneResponses)
}

// Update updates an existing timezone
// @Summary Update timezone
// @Tags timezones
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Timezone ID" format(uuid)
// @Param request body dto.UpdateTimezoneRequest true "Timezone data"
// @Success 200 {object} response.Response{data=dto.TimezoneResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response "Timezone not found"
// @Failure 500 {object} response.Response
// @Router /v1/timezones/{id} [put]
func (h *TimezoneHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid timezone ID", err)
		return
	}

	var req dto.UpdateTimezoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	timezone := &entity.Timezone{
		ID:           id,
		Name:         req.Name,
		Abbreviation: req.Abbreviation,
		UtcOffset:    req.UtcOffset,
		UtcDstOffset: req.UtcDstOffset,
		Description:  req.Description,
		IsActive:     req.IsActive,
	}

	if err := h.timezoneUseCase.Update(c.Request.Context(), timezone); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "timezone not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update timezone", err)
		return
	}

	resp := h.toTimezoneResponse(timezone)
	response.Success(c, http.StatusOK, resp)
}

// Delete deletes a timezone by ID
// @Summary Delete timezone
// @Tags timezones
// @Security Bearer
// @Param id path string true "Timezone ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Timezone not found"
// @Failure 500 {object} response.Response
// @Router /v1/timezones/{id} [delete]
func (h *TimezoneHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid timezone ID", err)
		return
	}

	if err := h.timezoneUseCase.Delete(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "timezone not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete timezone", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// toTimezoneResponse converts entity.Timezone to dto.TimezoneResponse
func (h *TimezoneHandler) toTimezoneResponse(timezone *entity.Timezone) *dto.TimezoneResponse {
	return &dto.TimezoneResponse{
		ID:           timezone.ID,
		Name:         timezone.Name,
		Abbreviation: timezone.Abbreviation,
		UtcOffset:    timezone.UtcOffset,
		UtcDstOffset: timezone.UtcDstOffset,
		Description:  timezone.Description,
		IsActive:     timezone.IsActive,
		CreatedAt:    timezone.CreatedAt,
		UpdatedAt:    timezone.UpdatedAt,
	}
}
