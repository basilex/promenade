package dto

import (
	"time"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
)

// StartInstanceRequest represents a request to start a workflow instance
type StartInstanceRequest struct {
	DefinitionID      string                 `json:"definition_id" binding:"required"`
	ExternalReference *string                `json:"external_reference"`
	Priority          int                    `json:"priority" binding:"min=0,max=100"`
	Context           map[string]interface{} `json:"context"`
	AssignedTo        *string                `json:"assigned_to"`
}

// UpdateInstanceRequest represents a request to update a workflow instance
type UpdateInstanceRequest struct {
	Priority   *int    `json:"priority" binding:"omitempty,min=0,max=100"`
	AssignedTo *string `json:"assigned_to"`
}

// WorkflowInstanceResponse represents a workflow instance in responses
type WorkflowInstanceResponse struct {
	ID                string                 `json:"id"`
	DefinitionID      string                 `json:"definition_id"`
	DefinitionVersion int                    `json:"definition_version"`
	ExternalReference *string                `json:"external_reference,omitempty"`
	Status            string                 `json:"status"`
	CurrentState      string                 `json:"current_state"`
	Priority          int                    `json:"priority"`
	Context           map[string]interface{} `json:"context,omitempty"`
	AssignedTo        *string                `json:"assigned_to,omitempty"`
	StartedBy         string                 `json:"started_by"`
	StartedAt         *time.Time             `json:"started_at,omitempty"` // Pointer in entity
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage      *string                `json:"error_message,omitempty"`
	RetryCount        int                    `json:"retry_count"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// ToWorkflowInstanceResponse converts entity to DTO
func ToWorkflowInstanceResponse(instance *entity.WorkflowInstance) *WorkflowInstanceResponse {
	// jsonb.Map is already map[string]interface{}, no unmarshaling needed
	contextMap := instance.Context

	resp := &WorkflowInstanceResponse{
		ID:                instance.ID.String(),
		DefinitionID:      instance.DefinitionID.String(),
		DefinitionVersion: instance.DefinitionVersion,
		ExternalReference: instance.ExternalReference,
		Status:            string(instance.Status),
		CurrentState:      instance.CurrentState,
		Priority:          instance.Priority,
		Context:           contextMap,
		StartedBy:         instance.StartedBy.String(),
		StartedAt:         instance.StartedAt, // Already pointer
		CompletedAt:       instance.CompletedAt,
		ErrorMessage:      instance.ErrorMessage,
		RetryCount:        instance.RetryCount,
		CreatedAt:         instance.CreatedAt,
		UpdatedAt:         instance.UpdatedAt,
	}

	if instance.AssignedTo != nil {
		assignedToStr := instance.AssignedTo.String()
		resp.AssignedTo = &assignedToStr
	}

	return resp
}

// ToWorkflowInstanceListResponse converts slice of entities to DTOs
func ToWorkflowInstanceListResponse(instances []*entity.WorkflowInstance) []*WorkflowInstanceResponse {
	result := make([]*WorkflowInstanceResponse, len(instances))
	for i, instance := range instances {
		result[i] = ToWorkflowInstanceResponse(instance)
	}
	return result
}
