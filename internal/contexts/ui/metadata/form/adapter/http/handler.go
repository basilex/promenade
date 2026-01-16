package http

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/ui/metadata/form"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// FormHandler handles UI form definition HTTP requests.
type FormHandler struct {
	usecase form.IUseCase
}

// NewFormHandler creates a new handler.
func NewFormHandler(usecase form.IUseCase) *FormHandler {
	return &FormHandler{usecase: usecase}
}

// Create creates a new form definition.
// @Summary Create form definition
// @Description Create a new UI form definition
// @Tags UIForm
// @Accept json
// @Produce json
// @Param request body CreateFormRequest true "Form creation request"
// @Success 201 {object} response.Response{data=FormResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /ui/forms [post]
func (h *FormHandler) Create(c *gin.Context) {
	var req CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	formEntity, err := form.NewFormDefinition(req.FormID, req.EntityType, req.Name, req.Layout, req.Fields)
	if err != nil {
		switch {
		case errors.Is(err, form.ErrFormIDRequired):
			response.BadRequest(c, "form_id is required")
		case errors.Is(err, form.ErrEntityTypeRequired):
			response.BadRequest(c, "entity_type is required")
		case errors.Is(err, form.ErrFormNameRequired):
			response.BadRequest(c, "name is required")
		case errors.Is(err, form.ErrLayoutRequired):
			response.BadRequest(c, "layout is required")
		case errors.Is(err, form.ErrFieldsRequired):
			response.BadRequest(c, "fields are required")
		default:
			response.InternalError(c, "Failed to create form definition")
		}
		return
	}

	formEntity.Description = req.Description
	if req.Validation != nil {
		formEntity.Validation = jsonstore.NewField(req.Validation)
	}
	if req.Events != nil {
		formEntity.Events = jsonstore.NewField(req.Events)
	}
	if req.Permissions != nil {
		formEntity.Permissions = jsonstore.NewField(req.Permissions)
	}
	if req.I18n != nil {
		formEntity.I18n = jsonstore.NewField(req.I18n)
	}
	if req.IsActive != nil {
		formEntity.IsActive = *req.IsActive
	}
	if req.TenantID != "" {
		tenantID, err := uuidv7.Parse(req.TenantID)
		if err != nil {
			response.BadRequest(c, "Invalid tenant_id format")
			return
		}
		formEntity.TenantID = &tenantID
	}
	if req.CreatedBy != "" {
		createdBy, err := uuidv7.Parse(req.CreatedBy)
		if err != nil {
			response.BadRequest(c, "Invalid created_by format")
			return
		}
		formEntity.CreatedBy = &createdBy
	}

	created, err := h.usecase.CreateForm(c.Request.Context(), formEntity)
	if err != nil {
		switch {
		case errors.Is(err, form.ErrFormIDExists):
			response.Conflict(c, "form_id already exists")
		case errors.Is(err, form.ErrFormIDRequired),
			errors.Is(err, form.ErrEntityTypeRequired),
			errors.Is(err, form.ErrFormNameRequired),
			errors.Is(err, form.ErrLayoutRequired),
			errors.Is(err, form.ErrFieldsRequired):
			response.BadRequest(c, err.Error())
		default:
			response.InternalError(c, "Failed to create form definition")
		}
		return
	}

	response.Created(c, ToFormResponse(created))
}

// GetByID retrieves a form definition by ID.
// @Summary Get form definition by ID
// @Description Get UI form definition by ID
// @Tags UIForm
// @Produce json
// @Param id path string true "Form ID"
// @Success 200 {object} response.Response{data=FormResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /ui/forms/{id} [get]
func (h *FormHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid form ID format")
		return
	}

	formEntity, err := h.usecase.GetForm(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, form.ErrFormNotFound) {
			response.NotFound(c, "Form definition not found")
			return
		}
		response.InternalError(c, "Failed to retrieve form definition")
		return
	}

	response.Success(c, ToFormResponse(formEntity))
}

// List lists form definitions.
// @Summary List form definitions
// @Description List form definitions by entity_type
// @Tags UIForm
// @Produce json
// @Param entity_type query string false "Entity type"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=[]FormResponse}
// @Failure 500 {object} response.Response
// @Router /ui/forms [get]
func (h *FormHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	entityType := c.Query("entity_type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	forms, total, err := h.usecase.ListForms(c.Request.Context(), entityType, pageSize, (page-1)*pageSize)
	if err != nil {
		response.InternalError(c, "Failed to list form definitions")
		return
	}

	resp := make([]FormResponse, 0, len(forms))
	for _, f := range forms {
		resp = append(resp, ToFormResponse(f))
	}

	response.SuccessWithPagination(c, resp, int64(total), page, pageSize)
}

// Update updates a form definition.
// @Summary Update form definition
// @Description Update an existing UI form definition
// @Tags UIForm
// @Accept json
// @Produce json
// @Param id path string true "Form ID"
// @Param request body UpdateFormRequest true "Form update request"
// @Success 200 {object} response.Response{data=FormResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /ui/forms/{id} [put]
func (h *FormHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid form ID format")
		return
	}

	var req UpdateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	update := &form.FormDefinition{
		Name:        req.Name,
		EntityType:  req.EntityType,
		Description: req.Description,
		Layout:      jsonstore.NewField(req.Layout),
		Fields:      jsonstore.NewField(req.Fields),
		IsActive:    req.IsActive,
	}

	if req.Validation != nil {
		update.Validation = jsonstore.NewField(req.Validation)
	} else {
		update.Validation = jsonstore.NewNullField[map[string]any]()
	}
	if req.Events != nil {
		update.Events = jsonstore.NewField(req.Events)
	} else {
		update.Events = jsonstore.NewNullField[map[string]any]()
	}
	if req.Permissions != nil {
		update.Permissions = jsonstore.NewField(req.Permissions)
	} else {
		update.Permissions = jsonstore.NewNullField[map[string]any]()
	}
	if req.I18n != nil {
		update.I18n = jsonstore.NewField(req.I18n)
	} else {
		update.I18n = jsonstore.NewNullField[map[string]map[string]string]()
	}
	if req.TenantID != "" {
		tenantID, err := uuidv7.Parse(req.TenantID)
		if err != nil {
			response.BadRequest(c, "Invalid tenant_id format")
			return
		}
		update.TenantID = &tenantID
	}

	updated, err := h.usecase.UpdateForm(c.Request.Context(), id, update)
	if err != nil {
		if errors.Is(err, form.ErrFormNotFound) {
			response.NotFound(c, "Form definition not found")
			return
		}
		if errors.Is(err, form.ErrEntityTypeRequired) ||
			errors.Is(err, form.ErrFormNameRequired) ||
			errors.Is(err, form.ErrLayoutRequired) ||
			errors.Is(err, form.ErrFieldsRequired) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "Failed to update form definition")
		return
	}

	response.Success(c, ToFormResponse(updated))
}

// Delete soft-deletes a form definition.
// @Summary Delete form definition
// @Description Delete a UI form definition
// @Tags UIForm
// @Produce json
// @Param id path string true "Form ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /ui/forms/{id} [delete]
func (h *FormHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid form ID format")
		return
	}

	if err := h.usecase.DeleteForm(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to delete form definition")
		return
	}

	response.SuccessWithMessage(c, "Form definition deleted")
}
