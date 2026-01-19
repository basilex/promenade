package aggregate

import (
	"strings"

	roleerrors "github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/pkg/aggregate"
)

// Role represents a user role in the system
// This is an Aggregate Root in the Identity bounded context
type Role struct {
	aggregate.BaseAggregate

	// Identity
	Name        string // unique, lowercase identifier (e.g., "admin", "user", "manager")
	DisplayName string // Human-readable name for UI (e.g., "System Administrator")

	// Metadata
	Description string
	IsSystem    bool // System roles cannot be deleted
}

// NewRole creates a new role
func NewRole(name, displayName, description string) (*Role, error) {
	if err := validateRoleName(name); err != nil {
		return nil, err
	}

	if strings.TrimSpace(displayName) == "" {
		return nil, roleerrors.ErrRoleDisplayRequired
	}

	return &Role{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Name:          strings.ToLower(strings.TrimSpace(name)),
		DisplayName:   strings.TrimSpace(displayName),
		Description:   strings.TrimSpace(description),
		IsSystem:      false,
	}, nil
}

// NewSystemRole creates a system role (cannot be deleted)
func NewSystemRole(name, displayName, description string) (*Role, error) {
	role, err := NewRole(name, displayName, description)
	if err != nil {
		return nil, err
	}

	role.IsSystem = true
	return role, nil
}

// UpdateDescription updates the role description
func (r *Role) UpdateDescription(description string) {
	r.Description = strings.TrimSpace(description)
	r.Touch()
}

// UpdateDisplayName updates the role display name
func (r *Role) UpdateDisplayName(displayName string) error {
	trimmed := strings.TrimSpace(displayName)
	if trimmed == "" {
		return roleerrors.ErrRoleDisplayRequired
	}
	r.DisplayName = trimmed
	r.Touch()
	return nil
}

// CanDelete checks if the role can be deleted
func (r *Role) CanDelete() error {
	if r.IsSystem {
		return roleerrors.ErrCannotDeleteSystem
	}
	return nil
}

// Validate validates the role
func (r *Role) Validate() error {
	if err := validateRoleName(r.Name); err != nil {
		return err
	}

	if strings.TrimSpace(r.DisplayName) == "" {
		return roleerrors.ErrRoleDisplayRequired
	}

	if strings.TrimSpace(r.Description) == "" {
		return roleerrors.ErrRoleDescriptionRequired
	}

	return nil
}

// validateRoleName validates role name
func validateRoleName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return roleerrors.ErrRoleNameEmpty
	}

	if len(trimmed) < 2 {
		return roleerrors.ErrRoleNameTooShort
	}

	if len(trimmed) > 50 {
		return roleerrors.ErrRoleNameTooLong
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	for _, ch := range trimmed {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') &&
			(ch < '0' || ch > '9') && ch != '_' && ch != '-' {
			return roleerrors.ErrRoleNameInvalidChars
		}
	}

	return nil
}
