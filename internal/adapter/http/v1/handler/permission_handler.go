package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type PermissionHandler struct {
	permissionUseCase usecase.IPermissionUseCase
}

func NewPermissionHandler(permissionUseCase usecase.IPermissionUseCase) *PermissionHandler {
	return &PermissionHandler{
		permissionUseCase: permissionUseCase,
	}
}

// CreatePermission godoc
// @Summary Create a new permission
// @Description Create a new permission with resource and action
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreatePermissionRequest true "Create permission request"
// @Success 201 {object} response.Response{data=dto.PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req dto.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	permission, err := h.permissionUseCase.CreatePermission(c.Request.Context(), req.Resource, req.Action, req.Description)
	if err != nil {
		if errors.Is(err, usecase.ErrPermissionAlreadyExists) {
			response.Error(c, http.StatusConflict, "permission already exists", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to create permission", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToPermissionResponse(permission))
}

// GetPermission godoc
// @Summary Get permission by ID
// @Description Get detailed information about a specific permission
// @Tags permissions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Permission ID" format(uuid)
// @Success 200 {object} response.Response{data=dto.PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	permissionID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID", err)
		return
	}

	permission, err := h.permissionUseCase.GetPermission(c.Request.Context(), permissionID)
	if err != nil {
		if errors.Is(err, usecase.ErrPermissionNotFound) {
			response.Error(c, http.StatusNotFound, "permission not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to get permission", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPermissionResponse(permission))
}

// ListPermissions godoc
// @Summary List all permissions
// @Description Get list of all available permissions
// @Tags permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]dto.PermissionResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	permissions, err := h.permissionUseCase.ListPermissions(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list permissions", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPermissionResponses(permissions))
}

// UpdatePermission godoc
// @Summary Update permission
// @Description Update permission description
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Permission ID" format(uuid)
// @Param request body dto.UpdatePermissionRequest true "Update permission request"
// @Success 200 {object} response.Response{data=dto.PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	permissionID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID", err)
		return
	}

	var req dto.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body", err)
		return
	}

	permission, err := h.permissionUseCase.UpdatePermission(c.Request.Context(), permissionID, req.Description)
	if err != nil {
		if errors.Is(err, usecase.ErrPermissionNotFound) {
			response.Error(c, http.StatusNotFound, "permission not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to update permission", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPermissionResponse(permission))
}

// DeletePermission godoc
// @Summary Delete permission
// @Description Delete a permission
// @Tags permissions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Permission ID" format(uuid)
// @Success 204 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	permissionID, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid permission ID", err)
		return
	}

	if err := h.permissionUseCase.DeletePermission(c.Request.Context(), permissionID); err != nil {
		if errors.Is(err, usecase.ErrPermissionNotFound) {
			response.Error(c, http.StatusNotFound, "permission not found", err)
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to delete permission", err)
		return
	}

	response.Success(c, http.StatusNoContent, nil)
}

// FindPermissionsByResource godoc
// @Summary Find permissions by resource
// @Description Get all permissions for a specific resource
// @Tags permissions
// @Produce json
// @Security BearerAuth
// @Param resource query string true "Resource name"
// @Success 200 {object} response.Response{data=[]dto.PermissionResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /permissions/search [get]
func (h *PermissionHandler) FindPermissionsByResource(c *gin.Context) {
	resource := c.Query("resource")
	if resource == "" {
		response.Error(c, http.StatusBadRequest, "resource parameter is required", nil)
		return
	}

	permissions, err := h.permissionUseCase.FindByResource(c.Request.Context(), resource)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to find permissions", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToPermissionResponses(permissions))
}
