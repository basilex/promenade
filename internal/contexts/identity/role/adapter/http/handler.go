package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// RoleHandler handles HTTP requests for role operations
type RoleHandler struct {
	usecase role.IUseCase
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(usecase role.IUseCase) *RoleHandler {
	return &RoleHandler{
		usecase: usecase,
	}
}

// Create handles POST /roles
// @Summary Create a new role
// @Description Create a new role with name and display name
// @Tags roles
// @Accept json
// @Produce json
// @Param request body CreateRoleRequest true "Role creation request"
// @Success 201 {object} RoleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	created, err := h.usecase.CreateRole(
		c.Request.Context(),
		req.Name,
		req.DisplayName,
		req.Description,
	)
	if err != nil {
		if errors.Is(err, role.ErrRoleNameExists) {
			response.Conflict(c, "role name already exists")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, ToRoleResponse(created))
}

// GetByID handles GET /roles/:id
// @Summary Get role by ID
// @Description Retrieve a role by its UUID
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role UUID"
// @Success 200 {object} RoleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /roles/{id} [get]
func (h *RoleHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid role ID")
		return
	}

	found, err := h.usecase.GetRole(c.Request.Context(), roleID)
	if err != nil {
		if errors.Is(err, role.ErrRoleNotFound) {
			response.NotFound(c, "role not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToRoleResponse(found))
}

// GetByName handles GET /roles/name/:name
// @Summary Get role by name
// @Description Retrieve a role by its name
// @Tags roles
// @Accept json
// @Produce json
// @Param name path string true "Role name"
// @Success 200 {object} RoleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /roles/name/{name} [get]
func (h *RoleHandler) GetByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.BadRequest(c, "role name is required")
		return
	}

	found, err := h.usecase.GetRoleByName(c.Request.Context(), name)
	if err != nil {
		if errors.Is(err, role.ErrRoleNotFound) {
			response.NotFound(c, "role not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToRoleResponse(found))
}

// Update handles PUT /roles/:id
// @Summary Update role
// @Description Update an existing role's display name and description
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role UUID"
// @Param request body UpdateRoleRequest true "Role update request"
// @Success 200 {object} RoleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid role ID")
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.usecase.UpdateRole(
		c.Request.Context(),
		roleID,
		req.DisplayName,
		req.Description,
	)
	if err != nil {
		if errors.Is(err, role.ErrRoleNotFound) {
			response.NotFound(c, "role not found")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToRoleResponse(updated))
}

// Delete handles DELETE /roles/:id
// @Summary Delete role
// @Description Soft delete a role (system roles are protected)
// @Tags roles
// @Accept json
// @Produce json
// @Param id path string true "Role UUID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	roleID, err := uuidv7.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "invalid role ID")
		return
	}

	err = h.usecase.DeleteRole(c.Request.Context(), roleID)
	if err != nil {
		if errors.Is(err, role.ErrRoleNotFound) {
			response.NotFound(c, "role not found")
			return
		}
		if errors.Is(err, role.ErrCannotDeleteSystem) {
			response.Forbidden(c, "cannot delete system role")
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// List handles GET /roles
// @Summary List roles
// @Description Retrieve a paginated list of roles
// @Tags roles
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} RoleListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	roles, total, err := h.usecase.ListRoles(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToRoleListResponse(roles, total, limit, offset))
}

// GetUserRoles handles GET /users/:userId/roles
// @Summary Get user roles
// @Description Retrieve all roles assigned to a user
// @Tags roles
// @Accept json
// @Produce json
// @Param userId path string true "User UUID"
// @Success 200 {array} RoleResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{userId}/roles [get]
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuidv7.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	roles, err := h.usecase.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	result := make([]RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = ToRoleResponse(r)
	}

	response.Success(c, result)
}
