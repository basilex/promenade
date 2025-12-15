package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/basilex/promenade/internal/domain/entity"
)

type RoleResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name      string    `json:"name" example:"Sample Role"`
	Active    bool      `json:"active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

type CreateRoleRequest struct {
	Name string `json:"name" binding:"required,min=2,max=255" example:"Sample Role"`
}

type UpdateRoleRequest struct {
	Name   string `json:"name" binding:"omitempty,min=2,max=255" example:"Updated Role"`
	Active *bool  `json:"active" example:"true"`
}

type ListRolesResponse struct {
	Roles []*RoleResponse `json:"roles"`
	Total int             `json:"total" example:"100"`
	Page  int             `json:"page" example:"1"`
}

func ToRoleResponse(role *entity.Role) *RoleResponse {
	return &RoleResponse{
		ID:        role.ID,
		Name:      role.Name,
		Active:    role.Active,
		CreatedAt: role.CreatedAt,
	}
}
