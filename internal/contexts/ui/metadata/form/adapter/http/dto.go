package http

import "github.com/basilex/promenade/internal/contexts/ui/metadata/form"

// CreateFormRequest defines payload for creating a form.
type CreateFormRequest struct {
	FormID      string                       `json:"form_id" binding:"required"`
	EntityType  string                       `json:"entity_type" binding:"required"`
	Name        string                       `json:"name" binding:"required"`
	Description string                       `json:"description"`
	Layout      map[string]any               `json:"layout" binding:"required"`
	Fields      []map[string]any             `json:"fields" binding:"required"`
	Validation  map[string]any               `json:"validation"`
	Events      map[string]any               `json:"events"`
	Permissions map[string]any               `json:"permissions"`
	I18n        map[string]map[string]string `json:"i18n"`
	IsActive    *bool                        `json:"is_active"`
	TenantID    string                       `json:"tenant_id"`
	CreatedBy   string                       `json:"created_by"`
}

// UpdateFormRequest defines payload for updating a form.
type UpdateFormRequest struct {
	EntityType  string                       `json:"entity_type" binding:"required"`
	Name        string                       `json:"name" binding:"required"`
	Description string                       `json:"description"`
	Layout      map[string]any               `json:"layout" binding:"required"`
	Fields      []map[string]any             `json:"fields" binding:"required"`
	Validation  map[string]any               `json:"validation"`
	Events      map[string]any               `json:"events"`
	Permissions map[string]any               `json:"permissions"`
	I18n        map[string]map[string]string `json:"i18n"`
	IsActive    bool                         `json:"is_active" binding:"required"`
	TenantID    string                       `json:"tenant_id"`
}

// FormResponse defines response payload for a form.
type FormResponse struct {
	ID          string                       `json:"id"`
	FormID      string                       `json:"form_id"`
	EntityType  string                       `json:"entity_type"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Layout      map[string]any               `json:"layout"`
	Fields      []map[string]any             `json:"fields"`
	Validation  map[string]any               `json:"validation"`
	Events      map[string]any               `json:"events"`
	Permissions map[string]any               `json:"permissions"`
	I18n        map[string]map[string]string `json:"i18n"`
	Version     int                          `json:"version"`
	IsActive    bool                         `json:"is_active"`
	TenantID    *string                      `json:"tenant_id,omitempty"`
	CreatedBy   *string                      `json:"created_by,omitempty"`
	CreatedAt   string                       `json:"created_at"`
	UpdatedAt   string                       `json:"updated_at"`
}

// ToFormResponse converts entity to response.
func ToFormResponse(entity *form.FormDefinition) FormResponse {
	var tenantID *string
	if entity.TenantID != nil {
		val := entity.TenantID.String()
		tenantID = &val
	}

	var createdBy *string
	if entity.CreatedBy != nil {
		val := entity.CreatedBy.String()
		createdBy = &val
	}

	return FormResponse{
		ID:          entity.GetID().String(),
		FormID:      entity.FormID,
		EntityType:  entity.EntityType,
		Name:        entity.Name,
		Description: entity.Description,
		Layout:      entity.Layout.Get(),
		Fields:      entity.Fields.Get(),
		Validation:  entity.Validation.Get(),
		Events:      entity.Events.Get(),
		Permissions: entity.Permissions.Get(),
		I18n:        entity.I18n.Get(),
		Version:     entity.Version,
		IsActive:    entity.IsActive,
		TenantID:    tenantID,
		CreatedBy:   createdBy,
		CreatedAt:   entity.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   entity.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
