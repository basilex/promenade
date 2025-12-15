package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/adapter/http/v1/dto"
	"github.com/basilex/promenade/internal/usecase"
)

type RoleHandler struct {
	roleUseCase *usecase.RoleUseCase
}

func NewRoleHandler(roleUseCase *usecase.RoleUseCase) *RoleHandler {
	return &RoleHandler{roleUseCase: roleUseCase}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	role, err := h.roleUseCase.CreateRole(c.Request.Context(), req.Name)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create role", err)
		return
	}

	response.Success(c, http.StatusCreated, dto.ToRoleResponse(role))
}

func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role id", err)
		return
	}

	role, err := h.roleUseCase.GetRole(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "role not found", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToRoleResponse(role))
}

func (h *RoleHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	offset := (page - 1) * limit

	roles, total, err := h.roleUseCase.ListRoles(c.Request.Context(), limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list roles", err)
		return
	}

	roleResponses := make([]*dto.RoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = dto.ToRoleResponse(role)
	}

	resp := &dto.ListRolesResponse{
		Roles: roleResponses,
		Total: total,
		Page:  page,
	}

	response.Success(c, http.StatusOK, resp)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role id", err)
		return
	}

	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	role, err := h.roleUseCase.GetRole(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "role not found", err)
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Active != nil {
		role.Active = *req.Active
	}

	if err := h.roleUseCase.UpdateRole(c.Request.Context(), role); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update role", err)
		return
	}

	response.Success(c, http.StatusOK, dto.ToRoleResponse(role))
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid role id", err)
		return
	}

	if err := h.roleUseCase.DeleteRole(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to delete role", err)
		return
	}

	c.Status(http.StatusNoContent)
}
