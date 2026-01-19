package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/language/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/language/dto"
	"github.com/basilex/promenade/internal/contexts/shared/language/usecase"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Handler handles language-related HTTP requests
type Handler struct {
	u usecase.ILanguageUseCase
}

// NewHandler creates a new language handler
func NewHandler(u usecase.ILanguageUseCase) *Handler {
	return &Handler{u: u}
}

// ListLanguages godoc
// @Summary List all languages
// @Description Get list of all active languages with ISO codes
// @Tags Reference Data
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.LanguageResponse}
// @Failure 500 {object} response.Response
// @Router /languages [get]
func (h *Handler) ListLanguages(c *gin.Context) {
	ctx := c.Request.Context()

	languages, err := h.u.List(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch languages")
		return
	}

	response.Success(c, dto.ToLanguageResponses(languages))
}

// GetLanguageByCode godoc
// @Summary Get language by ISO code
// @Description Get language details by ISO 639-1 code
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param code path string true "Language Code (ISO 639-1, e.g., en, uk, de)"
// @Success 200 {object} response.Response{data=dto.LanguageResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /languages/{code} [get]
func (h *Handler) GetLanguageByCode(c *gin.Context) {
	ctx := c.Request.Context()
	code := strings.ToLower(c.Param("code"))

	if len(code) != 2 {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CODE", "Language code must be 2 characters (ISO 639-1)")
		return
	}

	language, err := h.u.GetByCode(ctx, code)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "LANGUAGE_NOT_FOUND", "Language not found")
		return
	}

	response.Success(c, dto.ToLanguageResponse(language))
}

// CreateLanguage creates a new language
func (h *Handler) CreateLanguage(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateLanguageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	language, err := aggregate.NewLanguage(req.Code, req.Name, req.NativeName)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_DATA", err.Error())
		return
	}

	language.Code3 = req.Code3

	if err := h.u.Create(ctx, language); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_ERROR", "Failed to create language")
		return
	}

	response.Created(c, dto.ToLanguageResponse(language))
}

// UpdateLanguage updates an existing language
func (h *Handler) UpdateLanguage(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var req dto.UpdateLanguageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	uuidID, err := uuidv7.Parse(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	existing, err := h.u.GetByID(ctx, uuidID)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "LANGUAGE_NOT_FOUND", "Language not found")
		return
	}

	existing.Code3 = req.Code3
	existing.Name = req.Name
	existing.NativeName = req.NativeName
	existing.IsActive = req.IsActive

	if err := h.u.Update(ctx, existing); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_ERROR", "Failed to update language")
		return
	}

	response.Success(c, dto.ToLanguageResponse(existing))
}

// DeleteLanguage deletes a language
func (h *Handler) DeleteLanguage(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	uuidID, err := uuidv7.Parse(id)
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid UUID format")
		return
	}

	if err := h.u.Delete(ctx, uuidID); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_ERROR", "Failed to delete language")
		return
	}

	c.Status(http.StatusNoContent)
}
