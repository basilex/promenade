package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type RoleHandler struct {
	roleUseCase usecase.RoleUseCase
}

func NewRoleHandler(roleUseCase usecase.RoleUseCase) *RoleHandler {
	return &RoleHandler{
		roleUseCase: roleUseCase,
	}
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with permissions
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateRoleRequest true "Create role request"
// @Success 201 {object} response.Response{data=dto.RoleResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	role, err := h.roleUseCase.CreateRole(c.Request.Context(), req.Name, req.DisplayName, req.Description)
	if err != nil {
		if errors.Is(err, usecase.ErrRoleAlreadyExists) {
			response.Error(c, http.StatusConflict, "role already exists", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create role", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToRoleResponse(role))
}

// GetRole godoc
// @Summary Get role by ID
// @Description Get detailed information about a specific role
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Success 200 {object} response.Response{data=dto.RoleResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	roleID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID", err)
		return
	}

	role, err := h.roleUseCase.GetRole(c.Request.Context(), roleID)
	if err != nil {
		if errors.Is(err, usecase.ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, "role not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get role", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToRoleResponse(role))
}

// ListRoles godoc
// @Summary List all roles
// @Description Get list of all roles with their permissions
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]dto.RoleResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.roleUseCase.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list roles", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToRoleResponses(roles))
}

// UpdateRole godoc
// @Summary Update role
// @Description Update role display name and description
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Param request body dto.UpdateRoleRequest true "Update role request"
// @Success 200 {object} response.Response{data=dto.RoleResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	roleID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID", err)
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	role, err := h.roleUseCase.UpdateRole(c.Request.Context(), roleID, req.DisplayName, req.Description)
	if err != nil {
		if errors.Is(err, usecase.ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, "role not found", err)
			return
		}
		if errors.Is(err, usecase.ErrCannotDeleteSystemRole) {
			response.Error(c, http.StatusForbidden, "cannot update system role", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update role", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToRoleResponse(role))
}

// DeleteRole godoc
// @Summary Delete role
// @Description Delete a role (system roles cannot be deleted)
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Success 204 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	roleID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID", err)
		return
	}

	if err := h.roleUseCase.DeleteRole(c.Request.Context(), roleID); err != nil {
		if errors.Is(err, usecase.ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, "role not found", err)
			return
		}
		if errors.Is(err, usecase.ErrCannotDeleteSystemRole) {
			response.Error(c, http.StatusForbidden, "cannot delete system role", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete role", err)
		return
	}

	response.Success(c, http.StatusNoContent, nil)
}

// SyncRolePermissions godoc
// @Summary Sync role permissions
// @Description Replace all role permissions with new set
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID" format(uuid)
// @Param request body dto.SyncRolePermissionsRequest true "Permission IDs"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles/{id}/permissions [put]
func (h *RoleHandler) SyncRolePermissions(c *gin.Context) {
	roleID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID", err)
		return
	}

	var req dto.SyncRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	// Parse permission IDs
	permissionIDs := make([]uuidv7.UUID, len(req.PermissionIDs))
	for i, idStr := range req.PermissionIDs {
		permID, err := uuidv7.Parse(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid permission ID", err)
			return
		}
		permissionIDs[i] = permID
	}

	if err := h.roleUseCase.SyncRolePermissions(c.Request.Context(), roleID, permissionIDs); err != nil {
		if errors.Is(err, usecase.ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, "role not found", err)
			return
		}
		if errors.Is(err, usecase.ErrPermissionNotFound) {
			response.Error(c, http.StatusNotFound, "permission not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to sync permissions", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "permissions synced successfully"})
}

// AssignRoleToUser godoc
// @Summary Assign role to user
// @Description Assign a role to a user with optional expiration
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AssignRoleRequest true "Assign role request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles/assign [post]
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	var req dto.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	userID, err := uuidv7.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	roleID, err := uuidv7.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID", err)
		return
	}

	// Get current user ID (who is assigning the role)
	assignedBy, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
		return
	}

	if err := h.roleUseCase.AssignRoleToUser(c.Request.Context(), userID, roleID, assignedBy, req.ExpiresAt); err != nil {
		if errors.Is(err, usecase.ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, "role not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to assign role", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "role assigned successfully"})
}

// RemoveRoleFromUser godoc
// @Summary Remove role from user
// @Description Remove a role assignment from a user
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.RemoveRoleRequest true "Remove role request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /roles/remove [post]
func (h *RoleHandler) RemoveRoleFromUser(c *gin.Context) {
	var req dto.RemoveRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	userID, err := uuidv7.Parse(req.UserID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	roleID, err := uuidv7.Parse(req.RoleID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role ID", err)
		return
	}

	if err := h.roleUseCase.RemoveRoleFromUser(c.Request.Context(), userID, roleID); err != nil {
		if errors.Is(err, usecase.ErrRoleNotFound) {
			response.Error(c, http.StatusNotFound, "role not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to remove role", err)
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "role removed successfully"})
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all active roles assigned to a user
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID" format(uuid)
// @Success 200 {object} response.Response{data=[]dto.RoleResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/roles [get]
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID, err := uuidv7.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	roles, err := h.roleUseCase.GetUserActiveRoles(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get user roles", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToRoleResponses(roles))
}

// GetUserPermissions godoc
// @Summary Get user permissions
// @Description Get all permissions assigned to a user through their roles
// @Tags roles
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID" format(uuid)
// @Success 200 {object} response.Response{data=[]dto.PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/permissions [get]
func (h *RoleHandler) GetUserPermissions(c *gin.Context) {
	userID, err := uuidv7.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	permissions, err := h.roleUseCase.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to get user permissions", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPermissionResponses(permissions))
}

// CheckPermission godoc
// @Summary Check user permission
// @Description Check if a user has a specific permission
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID" format(uuid)
// @Param request body dto.CheckPermissionRequest true "Check permission request"
// @Success 200 {object} response.Response{data=dto.CheckPermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/check-permission [post]
func (h *RoleHandler) CheckPermission(c *gin.Context) {
	userID, err := uuidv7.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user ID", err)
		return
	}

	var req dto.CheckPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	hasPermission, err := h.roleUseCase.HasPermission(c.Request.Context(), userID, req.Permission)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to check permission", err)
		return
	}

	response.Success(c, http.StatusOK, dto.CheckPermissionResponse{
		HasPermission: hasPermission,
	})
}
