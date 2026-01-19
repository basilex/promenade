package aggregate

import (
	scripterrors "github.com/basilex/promenade/internal/contexts/scripting/script"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
)

// ScriptStatus represents the status of a script
type ScriptStatus string

const (
	ScriptStatusDraft    ScriptStatus = "draft"
	ScriptStatusActive   ScriptStatus = "active"
	ScriptStatusInactive ScriptStatus = "inactive"
	ScriptStatusArchived ScriptStatus = "archived"
)

// ScriptType represents the category of a script
type ScriptType string

const (
	ScriptTypeValidation   ScriptType = "validation"
	ScriptTypeWorkflow     ScriptType = "workflow"
	ScriptTypeReport       ScriptType = "report"
	ScriptTypePricing      ScriptType = "pricing"
	ScriptTypeNotification ScriptType = "notification"
	ScriptTypeAutomation   ScriptType = "automation"
	ScriptTypeCustom       ScriptType = "custom"
)

// Script is an aggregate root representing a LUA script
type Script struct {
	aggregate.BaseAggregate

	Name        string                             // Unique script name
	Description string                             // Script description
	Code        string                             // LUA code
	Version     int                                // Script version (increments on code changes)
	Status      ScriptStatus                       // Script status
	ScriptType  ScriptType                         // Script category (validation, workflow, etc.)
	EntityType  string                             // Target entity (customer, order, deal, etc.)
	Metadata    jsonstore.Field[map[string]string] // Additional metadata (database-agnostic)
}

// NewScript creates a new script with draft status
func NewScript(name, code string, scriptType ScriptType) (*Script, error) {
	if name == "" {
		return nil, scripterrors.ErrScriptNameEmpty
	}
	if code == "" {
		return nil, scripterrors.ErrScriptCodeEmpty
	}

	return &Script{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Name:          name,
		Code:          code,
		Version:       1,
		Status:        ScriptStatusDraft,
		ScriptType:    scriptType,
		Metadata:      jsonstore.NewField(make(map[string]string)),
	}, nil
}

// UpdateCode updates the script code and increments version
func (s *Script) UpdateCode(newCode string) error {
	if newCode == "" {
		return scripterrors.ErrScriptCodeEmpty
	}

	s.Code = newCode
	s.Version++
	s.Touch()
	return nil
}

// UpdateMetadata sets or updates a metadata field
func (s *Script) UpdateMetadata(key string, value string) {
	metadata := s.Metadata.Get()
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata[key] = value
	s.Metadata.Set(metadata)
	s.Touch()
}

// DeleteMetadata removes a metadata field
func (s *Script) DeleteMetadata(key string) {
	metadata := s.Metadata.Get()
	if metadata != nil {
		delete(metadata, key)
		s.Metadata.Set(metadata)
		s.Touch()
	}
}

// SetEntityType sets the target entity type
func (s *Script) SetEntityType(entityType string) {
	s.EntityType = entityType
	s.Touch()
}

// Activate changes script status to active
func (s *Script) Activate() error {
	if s.Status == ScriptStatusArchived {
		return scripterrors.ErrScriptCannotActivateArchived
	}

	s.Status = ScriptStatusActive
	s.Touch()
	return nil
}

// Deactivate changes script status to inactive
func (s *Script) Deactivate() error {
	if s.Status == ScriptStatusArchived {
		return scripterrors.ErrScriptCannotDeactivateArchived
	}

	s.Status = ScriptStatusInactive
	s.Touch()
	return nil
}

// Archive changes script status to archived (terminal state)
func (s *Script) Archive() {
	s.Status = ScriptStatusArchived
	s.Touch()
}

// IsActive returns true if script is active
func (s *Script) IsActive() bool {
	return s.Status == ScriptStatusActive
}

// IsDraft returns true if script is in draft status
func (s *Script) IsDraft() bool {
	return s.Status == ScriptStatusDraft
}

// IsArchived returns true if script is archived
func (s *Script) IsArchived() bool {
	return s.Status == ScriptStatusArchived
}

// GetMetadata returns a metadata value by key
func (s *Script) GetMetadata(key string) (string, bool) {
	metadata := s.Metadata.Get()
	if metadata == nil {
		return "", false
	}
	val, ok := metadata[key]
	return val, ok
}

// Validate checks if the script is valid
func (s *Script) Validate() error {
	if s.Name == "" {
		return scripterrors.ErrScriptNameEmpty
	}
	if s.Code == "" {
		return scripterrors.ErrScriptCodeEmpty
	}
	if s.Version < 1 {
		return scripterrors.ErrScriptVersionInvalid
	}

	// Validate status
	switch s.Status {
	case ScriptStatusDraft, ScriptStatusActive, ScriptStatusInactive, ScriptStatusArchived:
		// Valid statuses
	default:
		return scripterrors.ErrScriptStatusInvalid
	}

	return nil
}
