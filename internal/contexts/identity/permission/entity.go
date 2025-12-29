package permission

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
	ErrPermissionNotFound      = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
)

// Permission represents a specific permission in the system
// This is an Aggregate Root in the Identity bounded context
type Permission struct {
	aggregate.BaseAggregate

	// Identity
	ID   uuidv7.UUID
	Name string // unique identifier (e.g., "users:read", "customers:write")

	// Resource and Action
	Resource string // e.g., "users", "customers", "orders"
	Action   string // e.g., "read", "write", "delete", "admin"

	// Metadata
	Description string

	// Lifecycle
	CreatedAt time.Time
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
		ID:            uuidv7.New(),
		Name:          name,
		Resource:      strings.ToLower(strings.TrimSpace(resource)),
		Action:        strings.ToLower(strings.TrimSpace(action)),
		Description:   strings.TrimSpace(description),
		CreatedAt:     time.Now(),
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
func validateAction(action string) error {
	trimmed := strings.TrimSpace(action)
	if trimmed == "" {
		return fmt.Errorf("action cannot be empty")
	}

	validActions := []string{"read", "write", "delete", "admin", "create", "update", "list"}
	for _, valid := range validActions {
		if trimmed == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid action: %s (allowed: %v)", trimmed, validActions)
}
