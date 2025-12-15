package models

import (
	"time"
)

// Resource represents a managed resource in the system
type Resource struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateResourceRequest represents the request body for creating a resource
type CreateResourceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
}

// UpdateResourceRequest represents the request body for updating a resource
type UpdateResourceRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Status      string `json:"status,omitempty"`
}
