package role

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Common errors
var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleAlreadyExists  = errors.New("role already exists")
	ErrCannotDeleteSystem = errors.New("cannot delete system role")
)

// Role represents a user role in the system
// This is an Aggregate Root in the Identity bounded context
type Role struct {
	aggregate.BaseAggregate

	// Identity
	ID   uuidv7.UUID
	Name string // unique, lowercase (e.g., "admin", "user", "manager")

	// Metadata
	Description string
	IsSystem    bool // System roles cannot be deleted

	// Lifecycle
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewRole creates a new role
func NewRole(name, description string) (*Role, error) {
	if err := validateRoleName(name); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Role{
		BaseAggregate: aggregate.NewBaseAggregate(),
		ID:            uuidv7.New(),
		Name:          strings.ToLower(strings.TrimSpace(name)),
		Description:   strings.TrimSpace(description),
		IsSystem:      false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// NewSystemRole creates a system role (cannot be deleted)
func NewSystemRole(name, description string) (*Role, error) {
	role, err := NewRole(name, description)
	if err != nil {
		return nil, err
	}

	role.IsSystem = true
	return role, nil
}

// UpdateDescription updates the role description
func (r *Role) UpdateDescription(description string) {
	r.Description = strings.TrimSpace(description)
	r.UpdatedAt = time.Now()
}

// CanDelete checks if the role can be deleted
func (r *Role) CanDelete() error {
	if r.IsSystem {
		return ErrCannotDeleteSystem
	}
	return nil
}

// Validate validates the role
func (r *Role) Validate() error {
	if err := validateRoleName(r.Name); err != nil {
		return err
	}

	if r.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

// validateRoleName validates role name
func validateRoleName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("role name cannot be empty")
	}

	if len(trimmed) < 2 {
		return fmt.Errorf("role name must be at least 2 characters")
	}

	if len(trimmed) > 50 {
		return fmt.Errorf("role name must not exceed 50 characters")
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	for _, ch := range trimmed {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return fmt.Errorf("role name contains invalid characters")
		}
	}

	return nil
}
