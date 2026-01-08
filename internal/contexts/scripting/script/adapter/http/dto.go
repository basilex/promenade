package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/scripting/script"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Request DTOs
// ============================================================================

// CreateScriptRequest represents request to create a script
type CreateScriptRequest struct {
	Name        string                 `json:"name" binding:"required,min=3,max=100"`
	Description string                 `json:"description" binding:"max=500"`
	Code        string                 `json:"code" binding:"required"`
	EntityType  string                 `json:"entity_type" binding:"required,oneof=customer order deal product"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateScriptRequest represents request to update a script
type UpdateScriptRequest struct {
	Code        *string                 `json:"code"`
	Description *string                 `json:"description" binding:"omitempty,max=500"`
	Metadata    *map[string]interface{} `json:"metadata"`
}

// ExecuteScriptRequest represents request to execute a script
type ExecuteScriptRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
}

// ValidateScriptRequest represents request to validate script syntax
type ValidateScriptRequest struct {
	Code string `json:"code" binding:"required"`
}

// ============================================================================
// Response DTOs
// ============================================================================

// ScriptResponse represents a script in API response
type ScriptResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Code        string                 `json:"code"`
	EntityType  string                 `json:"entity_type"`
	Status      string                 `json:"status"`
	Version     int                    `json:"version"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ArchivedAt  *time.Time             `json:"archived_at,omitempty"`
}

// ScriptListResponse represents list of scripts
type ScriptListResponse struct {
	Scripts []ScriptResponse `json:"scripts"`
	Total   int              `json:"total"`
}

// ============================================================================
// Execution DTOs
// ============================================================================

// ExecutionResponse represents script execution result
type ExecutionResponse struct {
	ExecutionID string      `json:"execution_id"`
	ScriptID    string      `json:"script_id"`
	ScriptName  string      `json:"script_name"`
	Status      string      `json:"status"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	DurationMs  int64       `json:"duration_ms"`
	ExecutedBy  string      `json:"executed_by"`
	ExecutedAt  time.Time   `json:"executed_at"`
}

// ExecutionHistoryResponse represents execution history
type ExecutionHistoryResponse struct {
	Executions []ExecutionResponse `json:"executions"`
	Total      int                 `json:"total"`
}

// ValidationResponse represents script validation result
type ValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// ============================================================================
// Converters (Entity → DTO)
// ============================================================================

// ToScriptResponse converts Script entity to DTO
func ToScriptResponse(s *script.Script) ScriptResponse {
	return ScriptResponse{
		ID:          s.GetID().String(),
		Name:        s.Name,
		Description: s.Description,
		Code:        s.Code,
		EntityType:  "", // TODO: Will be added to Script entity later
		Status:      string(s.Status),
		Version:     s.Version,
		Metadata:    s.Metadata,
		CreatedBy:   s.GetID().String(), // TODO: Will be added to Script entity later
		CreatedAt:   s.GetCreatedAt(),
		UpdatedAt:   s.GetUpdatedAt(),
		ArchivedAt:  nil, // TODO: Soft delete support later
	}
}

// ToScriptListResponse converts slice of Script entities to DTO
func ToScriptListResponse(scripts []*script.Script, total int) ScriptListResponse {
	responses := make([]ScriptResponse, 0, len(scripts))
	for _, s := range scripts {
		responses = append(responses, ToScriptResponse(s))
	}

	return ScriptListResponse{
		Scripts: responses,
		Total:   total,
	}
}

// ToExecutionResponse converts ScriptExecution entity to DTO
func ToExecutionResponse(exec *script.ScriptExecution) ExecutionResponse {
	response := ExecutionResponse{
		ExecutionID: exec.ID.String(),
		ScriptID:    exec.ScriptID.String(),
		ScriptName:  exec.ScriptName,
		Status:      "success", // TODO: Add Status field to ScriptExecution entity
		DurationMs:  int64(exec.DurationMs),
		ExecutedBy:  exec.ExecutedBy.String(),
		ExecutedAt:  exec.ExecutedAt,
	}

	if exec.Error == nil {
		response.Result = exec.OutputResult
	} else {
		response.Status = "error"
		response.Error = *exec.Error
	}

	return response
}

// ToExecutionHistoryResponse converts slice of ScriptExecution entities to DTO
func ToExecutionHistoryResponse(executions []*script.ScriptExecution, total int) ExecutionHistoryResponse {
	responses := make([]ExecutionResponse, 0, len(executions))
	for _, exec := range executions {
		responses = append(responses, ToExecutionResponse(exec))
	}

	return ExecutionHistoryResponse{
		Executions: responses,
		Total:      total,
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// ParseUUID parses string to UUID v7
func ParseUUID(s string) (uuidv7.UUID, error) {
	return uuidv7.Parse(s)
}
