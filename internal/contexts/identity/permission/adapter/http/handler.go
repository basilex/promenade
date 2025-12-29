package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// PermissionHandler handles HTTP requests for permission operations
type PermissionHandler struct {
	usecase permission.IUseCase
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(usecase permission.IUseCase) *PermissionHandler {
	return &PermissionHandler{
		usecase: usecase,
	}
}

// Create handles POST /permissions
// @Summary Create a new permission
// @Description Create a new permission with resource and action
// @Tags permissions
// @Accept json
// @Produce json
// @Param request body CreatePermissionRequest true "Permission creation request"
// @Success 201 {object} PermissionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /permissions [post]
func (h *PermissionHandler) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	created, err := h.usecase.CreatePermission(
		c.Request.Context(),
		req.Resource,
		req.Action,
		req.Description,
	)
	if err != nil {
		if errors.Is(err, permission.ErrPermissionNameExists) {
			response.Conflict(c, "permission already exists")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, ToPermissionResponse(created))
}

// GetByID handles GET /permissions/:id
// @Summary Get permission by ID
// @Description Retrieve a permission by its UUID
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission UUID"
// @Success 200 {object} PermissionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /permissions/{id} [get]
func (h *PermissionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	permID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid permission ID")
		return
	}

	found, err := h.usecase.GetPermission(c.Request.Context(), permID)
	if err != nil {
		if errors.Is(err, permission.ErrPermissionNotFound) {
			response.NotFound(c, "permission not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToPermissionResponse(found))
}

// GetByName handles GET /permissions/name/:name
// @Summary Get permission by name
// @Description Retrieve a permission by its name (resource:action)
// @Tags permissions
// @Accept json
// @Produce json
// @Param name path string true "Permission name (resource:action)"
// @Success 200 {object} PermissionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /permissions/name/{name} [get]
func (h *PermissionHandler) GetByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.BadRequest(c, "permission name is required")
		return
	}

	found, err := h.usecase.GetPermissionByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, permission.ErrPermissionNotFound) {
			response.NotFound(c, "permission not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToPermissionResponse(found))
}

// Update handles PUT /permissions/:id
// @Summary Update permission
// @Description Update an existing permission's description
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission UUID"
// @Param request body UpdatePermissionRequest true "Permission update request"
// @Success 200 {object} PermissionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /permissions/{id} [put]
func (h *PermissionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	permID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid permission ID")
		return
	}

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.usecase.UpdatePermission(
		c.Request.Context(),
		permID,
		req.Description,
	)
	if err != nil {
		if errors.Is(err, permission.ErrPermissionNotFound) {
			response.NotFound(c, "permission not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToPermissionResponse(updated))
}

// Delete handles DELETE /permissions/:id
// @Summary Delete permission
// @Description Soft delete a permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission UUID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	permID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid permission ID")
		return
	}

	err = h.usecase.DeletePermission(c.Request.Context(), permID)
	if err != nil {
		if errors.Is(err, permission.ErrPermissionNotFound) {
			response.NotFound(c, "permission not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /permissions
// @Summary List permissions
// @Description Retrieve a paginated list of permissions
// @Tags permissions
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} PermissionListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /permissions [get]
func (h *PermissionHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	perms, total, err := h.usecase.ListPermissions(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToPermissionListResponse(perms, total, limit, offset))
}

// GetRolePermissions handles GET /roles/:roleId/permissions
// @Summary Get role permissions
// @Description Retrieve all permissions assigned to a role
// @Tags permissions
// @Accept json
// @Produce json
// @Param roleId path string true "Role UUID"
// @Success 200 {array} PermissionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /roles/{roleId}/permissions [get]
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	roleID, err := uuidv7.Parse(roleIDStr)
	if err != nil {
		response.BadRequest(c, "invalid role ID")
		return
	}

	perms, err := h.usecase.GetRolePermissions(c.Request.Context(), roleID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	result := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		result[i] = ToPermissionResponse(p)
	}

	response.Success(c, result)
}
