package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/role"
)

// CreateRoleRequest represents a role creation request
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateRoleRequest represents a role update request
type UpdateRoleRequest struct {
	DisplayName string `json:"display_name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleListResponse represents a paginated list of roles
type RoleListResponse struct {
	Roles  []RoleResponse `json:"roles"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// ToRoleResponse converts a Role entity to RoleResponse DTO
func ToRoleResponse(r *role.Role) RoleResponse {
	return RoleResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// ToRoleListResponse converts a slice of roles to response DTOs
func ToRoleListResponse(roles []*role.Role, total, limit, offset int) RoleListResponse {
	result := make([]RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = ToRoleResponse(r)
	}
	return RoleListResponse{
		Roles:  result,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}