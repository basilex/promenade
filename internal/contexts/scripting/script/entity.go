package script

import (
	"fmt"

	"github.com/basilex/promenade/pkg/aggregate"
)

// ScriptStatus represents the status of a script
type ScriptStatus string

const (
	ScriptStatusDraft    ScriptStatus = "draft"
	ScriptStatusActive   ScriptStatus = "active"
	ScriptStatusInactive ScriptStatus = "inactive"
	ScriptStatusArchived ScriptStatus = "archived"
)

// Script is an aggregate root representing a LUA script
type Script struct {
	aggregate.BaseAggregate

	Name        string                 // Unique script name
	Description string                 // Script description
	Code        string                 // LUA code
	Version     int                    // Script version (increments on code changes)
	Status      ScriptStatus           // Script status
	Metadata    map[string]interface{} // Additional metadata (stored as JSONB)
}

// NewScript creates a new script with draft status
func NewScript(name, code string) (*Script, error) {
	if name == "" {
		return nil, fmt.Errorf("script name cannot be empty")
	}
	if code == "" {
		return nil, fmt.Errorf("script code cannot be empty")
	}

	return &Script{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Name:          name,
		Code:          code,
		Version:       1,
		Status:        ScriptStatusDraft,
		Metadata:      make(map[string]interface{}),
	}, nil
}

// UpdateCode updates the script code and increments version
func (s *Script) UpdateCode(newCode string) error {
	if newCode == "" {
		return fmt.Errorf("script code cannot be empty")
	}

	s.Code = newCode
	s.Version++
	s.Touch()
	return nil
}

// UpdateMetadata sets or updates a metadata field
func (s *Script) UpdateMetadata(key string, value interface{}) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	s.Metadata[key] = value
	s.Touch()
}

// DeleteMetadata removes a metadata field
func (s *Script) DeleteMetadata(key string) {
	if s.Metadata != nil {
		delete(s.Metadata, key)
		s.Touch()
	}
}

// Activate changes script status to active
func (s *Script) Activate() error {
	if s.Status == ScriptStatusArchived {
		return fmt.Errorf("cannot activate archived script")
	}

	s.Status = ScriptStatusActive
	s.Touch()
	return nil
}

// Deactivate changes script status to inactive
func (s *Script) Deactivate() error {
	if s.Status == ScriptStatusArchived {
		return fmt.Errorf("cannot deactivate archived script")
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
func (s *Script) GetMetadata(key string) (interface{}, bool) {
	if s.Metadata == nil {
		return nil, false
	}
	val, ok := s.Metadata[key]
	return val, ok
}

// Validate checks if the script is valid
func (s *Script) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("script name cannot be empty")
	}
	if s.Code == "" {
		return fmt.Errorf("script code cannot be empty")
	}
	if s.Version < 1 {
		return fmt.Errorf("script version must be at least 1")
	}

	// Validate status
	switch s.Status {
	case ScriptStatusDraft, ScriptStatusActive, ScriptStatusInactive, ScriptStatusArchived:
		// Valid statuses
	default:
		return fmt.Errorf("invalid script status: %s", s.Status)
	}

	return nil
}
