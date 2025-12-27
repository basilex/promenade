package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/shared/dto"
	"github.com/basilex/promenade/pkg/reference"
	"github.com/basilex/promenade/pkg/response"
)

// LanguageHandler handles language-related HTTP requests
type LanguageHandler struct {
	repo *reference.Repository
}

// NewLanguageHandler creates a new language handler
func NewLanguageHandler(repo *reference.Repository) *LanguageHandler {
	return &LanguageHandler{repo: repo}
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
func (h *LanguageHandler) ListLanguages(c *gin.Context) {
	ctx := c.Request.Context()

	languages, err := h.repo.ListLanguages(ctx)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "FETCH_ERROR", "Failed to fetch languages")
		return
	}

	response.Success(c, dto.ToLanguageResponses(languages))
}

// GetLanguageByCode godoc
// @Summary Get language by ISO code
// @Description Get language details by ISO 639-1 alpha-2 code
// @Tags Reference Data
// @Accept json
// @Produce json
// @Param code path string true "Language Code (ISO 639-1, e.g., en, uk, de)"
// @Success 200 {object} response.Response{data=dto.LanguageResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /languages/{code} [get]
func (h *LanguageHandler) GetLanguageByCode(c *gin.Context) {
	ctx := c.Request.Context()
	code := strings.ToLower(c.Param("code"))

	if len(code) != 2 {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_CODE", "Language code must be 2 characters (ISO 639-1)")
		return
	}

	language, err := h.repo.GetLanguageByCode(ctx, code)
	if err != nil {
		response.ErrorResponse(c, http.StatusNotFound, "LANGUAGE_NOT_FOUND", "Language not found")
		return
	}

	response.Success(c, dto.ToLanguageResponse(language))
}
