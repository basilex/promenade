package dto

import (
	"time"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
)

// CreateDefinitionRequest represents request to create a workflow definition
type CreateDefinitionRequest struct {
	Name        string                 `json:"name" binding:"required,min=3,max=200"`
	Description string                 `json:"description" binding:"omitempty,max=1000"`
	Version     string                 `json:"version" binding:"omitempty"` // Ignored - version is auto-incremented
	Schema      entity.WorkflowSchema  `json:"schema" binding:"required"`
	Tags        []string               `json:"tags" binding:"omitempty"`
	Metadata    map[string]interface{} `json:"metadata" binding:"omitempty"`
}

// UpdateDefinitionRequest represents request to update a workflow definition
type UpdateDefinitionRequest struct {
	Name        *string                 `json:"name" binding:"omitempty,min=3,max=200"`
	Description *string                 `json:"description" binding:"omitempty,max=1000"`
	Schema      *entity.WorkflowSchema  `json:"schema" binding:"omitempty"`
	Tags        []string                `json:"tags"`
	Metadata    map[string]interface{}  `json:"metadata"`
}

// WorkflowDefinitionResponse represents a workflow definition in API response
type WorkflowDefinitionResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	DisplayName string                `json:"display_name"`
	Description string                `json:"description"`
	Version     int                   `json:"version"` // int, not string
	Status      string                `json:"status"`
	Category    string                `json:"category"`
	Schema      entity.WorkflowSchema `json:"schema"`
	Tags        []string              `json:"tags"`
	CreatedBy   string                `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   *time.Time            `json:"deleted_at,omitempty"`
}

// ToWorkflowDefinitionResponse converts entity to response DTO
func ToWorkflowDefinitionResponse(def *entity.WorkflowDefinition) *WorkflowDefinitionResponse {
	return &WorkflowDefinitionResponse{
		ID:          def.ID.String(),
		Name:        def.Name,
		DisplayName: def.DisplayName,
		Description: def.Description,
		Version:     def.Version, // int type
		Status:      string(def.Status),
		Category:    def.Category,
		Schema:      def.Definition.Data, // Extract data from jsonb.JSON[WorkflowSchema]
		Tags:        def.Tags,
		CreatedBy:   def.CreatedBy.String(),
		CreatedAt:   def.CreatedAt,
		UpdatedAt:   def.UpdatedAt,
		DeletedAt:   def.DeletedAt,
	}
}

// ToWorkflowDefinitionListResponse converts entity list to response DTO list
func ToWorkflowDefinitionListResponse(defs []*entity.WorkflowDefinition) []*WorkflowDefinitionResponse {
	response := make([]*WorkflowDefinitionResponse, len(defs))
	for i, def := range defs {
		response[i] = ToWorkflowDefinitionResponse(def)
	}
	return response
}
