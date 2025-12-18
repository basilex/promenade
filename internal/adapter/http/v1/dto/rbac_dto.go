package dto

import (
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
)

// Role DTOs

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"required,min=5,max=500"`
}

type UpdateRoleRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"required,min=5,max=500"`
}

type AssignRoleRequest struct {
	UserID    string     `json:"user_id" binding:"required,uuid"`
	RoleID    string     `json:"role_id" binding:"required,uuid"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type RemoveRoleRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	RoleID string `json:"role_id" binding:"required,uuid"`
}

type SyncRolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required"`
}

type RoleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	DisplayName string               `json:"display_name"`
	Description *string              `json:"description,omitempty"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type UserRoleResponse struct {
	UserID     string       `json:"user_id"`
	Role       RoleResponse `json:"role"`
	AssignedAt time.Time    `json:"assigned_at"`
	AssignedBy *string      `json:"assigned_by,omitempty"`
	ExpiresAt  *time.Time   `json:"expires_at,omitempty"`
}

// Permission DTOs

type CreatePermissionRequest struct {
	Resource    string `json:"resource" binding:"required,min=2,max=50"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"required,min=5,max=500"`
}

type UpdatePermissionRequest struct {
	Description string `json:"description" binding:"required,min=5,max=500"`
}

type PermissionResponse struct {
	ID          string    `json:"id"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CheckPermissionRequest struct {
	Permission string `json:"permission" binding:"required"`
}

type CheckPermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required,min=1"`
}

type CheckPermissionResponse struct {
	HasPermission bool `json:"has_permission"`
}

// Mappers

func ToRoleResponse(role *entity.Role) RoleResponse {
	resp := RoleResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}

	if len(role.Permissions) > 0 {
		resp.Permissions = make([]PermissionResponse, len(role.Permissions))
		for i, perm := range role.Permissions {
			resp.Permissions[i] = ToPermissionResponse(perm)
		}
	}

	return resp
}

func ToRoleResponses(roles []*entity.Role) []RoleResponse {
	responses := make([]RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = ToRoleResponse(role)
	}
	return responses
}

func ToPermissionResponse(perm *entity.Permission) PermissionResponse {
	return PermissionResponse{
		ID:          perm.ID.String(),
		Resource:    perm.Resource,
		Action:      perm.Action,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
	}
}

func ToPermissionResponses(permissions []*entity.Permission) []PermissionResponse {
	responses := make([]PermissionResponse, len(permissions))
	for i, perm := range permissions {
		responses[i] = ToPermissionResponse(perm)
	}
	return responses
}

func ToUserRoleResponse(userRole *entity.UserRole, role *entity.Role) UserRoleResponse {
	resp := UserRoleResponse{
		UserID:     userRole.UserID.String(),
		Role:       ToRoleResponse(role),
		AssignedAt: userRole.AssignedAt,
		ExpiresAt:  userRole.ExpiresAt,
	}

	if userRole.AssignedBy != nil {
		assignedByStr := userRole.AssignedBy.String()
		resp.AssignedBy = &assignedByStr
	}

	return resp
}
