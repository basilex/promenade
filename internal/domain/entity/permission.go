package entity

import (
	"fmt"
	"strings"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Permission represents a specific action that can be performed on a resource
type Permission struct {
	ID          uuidv7.UUID `db:"id" json:"id"`
	Resource    string      `db:"resource" json:"resource"`       // e.g., "posts", "users", "comments"
	Action      string      `db:"action" json:"action"`           // e.g., "create", "read", "update", "delete"
	Description *string     `db:"description" json:"description"` // Human-readable description
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
}

// String returns permission in format "resource:action"
func (p *Permission) String() string {
	return fmt.Sprintf("%s:%s", p.Resource, p.Action)
}

// Matches checks if this permission matches the required permission
// Supports wildcards: "posts:*" matches "posts:read", "posts:write", etc.
// "*:read" matches "posts:read", "users:read", etc.
// "*:*" or "*" matches everything
func (p *Permission) Matches(required string) bool {
	parts := strings.Split(required, ":")
	if len(parts) != 2 {
		return false
	}

	reqResource, reqAction := parts[0], parts[1]

	// Check for full wildcard
	if p.Resource == "*" && (p.Action == "*" || p.Action == "") {
		return true
	}

	// Check resource match (with wildcard support)
	resourceMatch := p.Resource == "*" || p.Resource == reqResource

	// Check action match (with wildcard support)
	actionMatch := p.Action == "*" || p.Action == reqAction

	return resourceMatch && actionMatch
}

// NewPermission creates a new Permission from "resource:action" format
func NewPermission(permissionString string) (*Permission, error) {
	parts := strings.Split(permissionString, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid permission format: %s (expected resource:action)", permissionString)
	}

	resource := strings.TrimSpace(parts[0])
	action := strings.TrimSpace(parts[1])

	if resource == "" || action == "" {
		return nil, fmt.Errorf("resource and action cannot be empty")
	}

	return &Permission{
		ID:        uuidv7.New(),
		Resource:  resource,
		Action:    action,
		CreatedAt: time.Now(),
	}, nil
}

// MustNewPermission creates a new Permission or panics
func MustNewPermission(permissionString string) *Permission {
	p, err := NewPermission(permissionString)
	if err != nil {
		panic(err)
	}
	return p
}

// Validate validates permission fields
func (p *Permission) Validate() error {
	if p.Resource == "" {
		return fmt.Errorf("resource is required")
	}
	if p.Action == "" {
		return fmt.Errorf("action is required")
	}
	if len(p.Resource) > 50 {
		return fmt.Errorf("resource must be 50 characters or less")
	}
	if len(p.Action) > 50 {
		return fmt.Errorf("action must be 50 characters or less")
	}
	return nil
}

// Common permission constants for convenience
const (
	// Wildcard permissions
	PermissionAll = "*:*"

	// User permissions
	PermissionUsersCreate = "users:create"
	PermissionUsersRead   = "users:read"
	PermissionUsersUpdate = "users:update"
	PermissionUsersDelete = "users:delete"
	PermissionUsersList   = "users:list"
	PermissionUsersBan    = "users:ban"
	PermissionUsersAll    = "users:*"

	// Post permissions
	PermissionPostsCreate = "posts:create"
	PermissionPostsRead   = "posts:read"
	PermissionPostsUpdate = "posts:update"
	PermissionPostsDelete = "posts:delete"
	PermissionPostsList   = "posts:list"
	PermissionPostsAll    = "posts:*"

	// Comment permissions
	PermissionCommentsCreate = "comments:create"
	PermissionCommentsRead   = "comments:read"
	PermissionCommentsUpdate = "comments:update"
	PermissionCommentsDelete = "comments:delete"
	PermissionCommentsList   = "comments:list"
	PermissionCommentsAll    = "comments:*"

	// Profile permissions
	PermissionProfilesCreate = "profiles:create"
	PermissionProfilesRead   = "profiles:read"
	PermissionProfilesUpdate = "profiles:update"
	PermissionProfilesDelete = "profiles:delete"
	PermissionProfilesList   = "profiles:list"
	PermissionProfilesAll    = "profiles:*"

	// Role permissions (admin only)
	PermissionRolesCreate = "roles:create"
	PermissionRolesRead   = "roles:read"
	PermissionRolesUpdate = "roles:update"
	PermissionRolesDelete = "roles:delete"
	PermissionRolesList   = "roles:list"
	PermissionRolesAssign = "roles:assign"
	PermissionRolesAll    = "roles:*"
)
