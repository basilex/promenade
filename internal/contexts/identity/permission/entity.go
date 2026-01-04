package permission

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/aggregate"
)

// Permission represents a specific permission in the system
// This is an Aggregate Root in the Identity bounded context
type Permission struct {
	aggregate.BaseAggregate

	// Identity
	Name string // unique identifier (e.g., "users:read", "customers:write")

	// Resource and Action
	Resource string // e.g., "users", "customers", "orders"
	Action   string // e.g., "read", "write", "delete", "admin"

	// Metadata
	Description string
}

// NewPermission creates a new permission
func NewPermission(resource, action, description string) (*Permission, error) {
	// Normalize to lowercase
	resource = strings.ToLower(strings.TrimSpace(resource))
	action = strings.ToLower(strings.TrimSpace(action))

	if err := validateResource(resource); err != nil {
		return nil, err
	}

	if err := validateAction(action); err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s:%s", resource, action)

	return &Permission{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Name:          name,
		Resource:      strings.ToLower(strings.TrimSpace(resource)),
		Action:        strings.ToLower(strings.TrimSpace(action)),
		Description:   strings.TrimSpace(description),
	}, nil
}

// Validate validates the permission
func (p *Permission) Validate() error {
	if err := validateResource(p.Resource); err != nil {
		return err
	}

	if err := validateAction(p.Action); err != nil {
		return err
	}

	if p.Description == "" {
		return fmt.Errorf("description is required")
	}

	return nil
}

// UpdateDescription updates the permission description
func (p *Permission) UpdateDescription(description string) {
	p.Description = strings.TrimSpace(description)
	p.Touch()
}

// validateResource validates resource name
func validateResource(resource string) error {
	trimmed := strings.TrimSpace(resource)
	if trimmed == "" {
		return fmt.Errorf("resource cannot be empty")
	}

	if len(trimmed) < 2 {
		return fmt.Errorf("resource must be at least 2 characters")
	}

	if len(trimmed) > 50 {
		return fmt.Errorf("resource must not exceed 50 characters")
	}

	return nil
}

// validateAction validates action name
// Allows wildcard (*) and common CRUD operations
// Extensible: any lowercase alphanumeric string is valid
func validateAction(action string) error {
	trimmed := strings.TrimSpace(action)
	if trimmed == "" {
		return fmt.Errorf("action cannot be empty")
	}

	if len(trimmed) > 50 {
		return fmt.Errorf("action must not exceed 50 characters")
	}

	// Allow wildcard
	if trimmed == "*" {
		return nil
	}

	// Allow any lowercase alphanumeric with underscores and hyphens
	// This makes it extensible for future actions
	for _, ch := range trimmed {
		if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') && ch != '_' && ch != '-' {
			return fmt.Errorf("action must contain only lowercase letters, numbers, underscores and hyphens")
		}
	}

	return nil
}
