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

type LanguageHandler struct {
	languageUseCase usecase.ILanguageUseCase
}

func NewLanguageHandler(languageUseCase usecase.ILanguageUseCase) *LanguageHandler {
	return &LanguageHandler{
		languageUseCase: languageUseCase,
	}
}

// Create creates a new language
// @Summary Create language
// @Tags languages
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.CreateLanguageRequest true "Language data"
// @Success 201 {object} response.Response{data=dto.LanguageResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response "Language already exists"
// @Failure 500 {object} response.Response
// @Router /v1/languages [post]
func (h *LanguageHandler) Create(c *gin.Context) {
	var req dto.CreateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	language := entity.NewLanguage(
		req.Name,
		req.NativeName,
		req.Code,
		req.ISO639_2,
		req.IsRtl,
	)
	language.IsActive = req.IsActive
	language.SortOrder = req.SortOrder

	if err := h.languageUseCase.Create(c.Request.Context(), language); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create language", err)
		return
	}

	resp := h.toLanguageResponse(language)
	response.Success(c, http.StatusCreated, resp)
}

// GetByID retrieves a language by ID
// @Summary Get language by ID
// @Tags languages
// @Produce json
// @Param id path string true "Language ID" format(uuid)
// @Success 200 {object} response.Response{data=dto.LanguageResponse}
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Language not found"
// @Failure 500 {object} response.Response
// @Router /v1/languages/{id} [get]
func (h *LanguageHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid language ID", err)
		return
	}

	language, err := h.languageUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "language not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get language", err)
		return
	}

	resp := h.toLanguageResponse(language)
	response.Success(c, http.StatusOK, resp)
}

// GetByCode retrieves a language by ISO 639-1 code
// @Summary Get language by code
// @Tags languages
// @Produce json
// @Param code path string true "Language code (ISO 639-1)" example(en)
// @Success 200 {object} response.Response{data=dto.LanguageResponse}
// @Failure 404 {object} response.Response "Language not found"
// @Failure 500 {object} response.Response
// @Router /v1/languages/code/{code} [get]
func (h *LanguageHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")

	language, err := h.languageUseCase.GetByCode(c.Request.Context(), code)
	if err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "language not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get language", err)
		return
	}

	resp := h.toLanguageResponse(language)
	response.Success(c, http.StatusOK, resp)
}

// List retrieves all languages with pagination
// @Summary List languages
// @Tags languages
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=response.PaginatedResponse{items=[]dto.LanguageResponse}}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /v1/languages [get]
func (h *LanguageHandler) List(c *gin.Context) {
	params := pagination.Params{
		Page:     response.GetPageFromQuery(c),
		PageSize: response.GetPageSizeFromQuery(c),
	}

	languages, metadata, err := h.languageUseCase.List(c.Request.Context(), params)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list languages", err)
		return
	}

	languageResponses := make([]*dto.LanguageResponse, len(languages))
	for i, lang := range languages {
		languageResponses[i] = h.toLanguageResponse(lang)
	}

	c.JSON(http.StatusOK, response.PaginatedResponse{
		Items:      languageResponses,
		Page:       metadata.CurrentPage,
		PageSize:   metadata.Limit,
		Total:      metadata.Total,
		TotalPages: metadata.TotalPages,
		Timestamp:  time.Now().Unix(),
	})
}

// ListActive retrieves all active languages
// @Summary List active languages
// @Tags languages
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.LanguageResponse}
// @Failure 500 {object} response.Response
// @Router /v1/languages/active [get]
func (h *LanguageHandler) ListActive(c *gin.Context) {
	languages, err := h.languageUseCase.ListActive(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list active languages", err)
		return
	}

	languageResponses := make([]*dto.LanguageResponse, len(languages))
	for i, lang := range languages {
		languageResponses[i] = h.toLanguageResponse(lang)
	}

	response.Success(c, http.StatusOK, languageResponses)
}

// Update updates an existing language
// @Summary Update language
// @Tags languages
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Language ID" format(uuid)
// @Param request body dto.UpdateLanguageRequest true "Language data"
// @Success 200 {object} response.Response{data=dto.LanguageResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response "Language not found"
// @Failure 500 {object} response.Response
// @Router /v1/languages/{id} [put]
func (h *LanguageHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid language ID", err)
		return
	}

	var req dto.UpdateLanguageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	language := &entity.Language{
		ID:         id,
		Name:       req.Name,
		NativeName: req.NativeName,
		Code:       req.Code,
		ISO639_2:   req.ISO639_2,
		IsRtl:      req.IsRtl,
		IsActive:   req.IsActive,
		SortOrder:  req.SortOrder,
	}

	if err := h.languageUseCase.Update(c.Request.Context(), language); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "language not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update language", err)
		return
	}

	resp := h.toLanguageResponse(language)
	response.Success(c, http.StatusOK, resp)
}

// Delete deletes a language by ID
// @Summary Delete language
// @Tags languages
// @Security Bearer
// @Param id path string true "Language ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} response.Response "Invalid ID format"
// @Failure 404 {object} response.Response "Language not found"
// @Failure 500 {object} response.Response
// @Router /v1/languages/{id} [delete]
func (h *LanguageHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid language ID", err)
		return
	}

	if err := h.languageUseCase.Delete(c.Request.Context(), id); err != nil {
		if err == entity.ErrNotFound {
			response.Error(c, http.StatusNotFound, "language not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete language", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// toLanguageResponse converts entity.Language to dto.LanguageResponse
func (h *LanguageHandler) toLanguageResponse(language *entity.Language) *dto.LanguageResponse {
	return &dto.LanguageResponse{
		ID:         language.ID,
		Name:       language.Name,
		NativeName: language.NativeName,
		Code:       language.Code,
		ISO639_2:   language.ISO639_2,
		IsRtl:      language.IsRtl,
		IsActive:   language.IsActive,
		SortOrder:  language.SortOrder,
		CreatedAt:  language.CreatedAt,
		UpdatedAt:  language.UpdatedAt,
	}
}
