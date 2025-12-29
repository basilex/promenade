package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/permission"
)

// CreatePermissionRequest represents a permission creation request
type CreatePermissionRequest struct {
	Resource    string `json:"resource" binding:"required,min=2,max=50"`
	Action      string `json:"action" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=500"`
}

// UpdatePermissionRequest represents a permission update request
type UpdatePermissionRequest struct {
	Description string `json:"description" binding:"max=500"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          string    `json:"id"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PermissionListResponse represents a paginated list of permissions
type PermissionListResponse struct {
	Permissions []PermissionResponse `json:"permissions"`
	Total       int                  `json:"total"`
	Limit       int                  `json:"limit"`
	Offset      int                  `json:"offset"`
}

// ToPermissionResponse converts a Permission entity to PermissionResponse DTO
func ToPermissionResponse(p *permission.Permission) PermissionResponse {
	return PermissionResponse{
		ID:          p.ID.String(),
		Resource:    p.Resource,
		Action:      p.Action,
		Name:        p.Name,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ToPermissionListResponse converts a slice of permissions to response DTOs
func ToPermissionListResponse(perms []*permission.Permission, total, limit, offset int) PermissionListResponse {
	result := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		result[i] = ToPermissionResponse(p)
	}
	return PermissionListResponse{
		Permissions: result,
		Total:       total,
		Limit:       limit,
		Offset:      offset,
	}
}
