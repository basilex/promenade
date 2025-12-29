# Role-Based Access Control (RBAC)

**Enterprise-grade authorization system for Promenade Platform**

## Overview

Promenade implements **Role-Based Access Control (RBAC)** for fine-grained authorization through **Roles** and **Permissions**.

### Key Concepts

**Role** - Named collection of permissions (e.g., "admin", "manager", "user")

**Permission** - Specific action on resource (e.g., "users:create", "customers:delete")

**User-Role Assignment** - Many-to-many relationship

**Role-Permission Assignment** - Many-to-many relationship

### Benefits

- **Separation of Concerns** - Authorization decoupled from business logic
- **Flexibility** - Add/modify roles without code changes
- **Scalability** - Supports complex permission hierarchies
- **Audit Trail** - Track permission assignments
- **JWT Integration** - Roles embedded in tokens for stateless auth

## Architecture

### Bounded Context

RBAC is part of the **Identity Context** with two aggregates:

**Role Aggregate:**
- ID (UUID v7)
- Name (unique, e.g., "admin")
- Description
- IsSystem flag (protects system roles)
- Permissions collection
- Soft delete support

**Permission Aggregate:**
- ID (UUID v7)
- Name (computed: "resource:action")
- Resource (e.g., "users", "customers")
- Action (e.g., "create", "read", "update", "delete")
- Description

## Database Schema

### Tables

**identity_roles:**
```sql
CREATE TABLE identity_roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

**identity_permissions:**
```sql
CREATE TABLE identity_permissions (
    id UUID PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    name VARCHAR(100) GENERATED ALWAYS AS (resource || ':' || action) STORED,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(resource, action)
);
```

**Junction Tables:**
```sql
CREATE TABLE identity_user_roles (
    user_id UUID REFERENCES identity_users(id),
    role_id UUID REFERENCES identity_roles(id),
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE identity_role_permissions (
    role_id UUID REFERENCES identity_roles(id),
    permission_id UUID REFERENCES identity_permissions(id),
    PRIMARY KEY (role_id, permission_id)
);
```

### Seed Data

**5 System Roles:**
- `superadmin` - Full system access
- `admin` - Admin operations
- `manager` - Team management
- `user` - Standard user access
- `guest` - Read-only access

**29+ Permissions:**
- Wildcard: `*:*` (all resources, all actions)
- Users: `users:create`, `users:read`, `users:update`, `users:delete`
- Roles: `roles:create`, `roles:read`, `roles:update`, `roles:delete`
- Customers: `customers:create`, `customers:read`, `customers:update`, `customers:delete`
- And more...

## API Endpoints

### Role Management

```http
POST   /api/v1/identity/roles              # Create role
GET    /api/v1/identity/roles              # List roles
GET    /api/v1/identity/roles/:id          # Get by ID
GET    /api/v1/identity/roles/name/:name   # Get by name
PUT    /api/v1/identity/roles/:id          # Update role
DELETE /api/v1/identity/roles/:id          # Delete role (soft)
GET    /api/v1/identity/users/:id/roles    # Get user's roles
```

### Permission Management

```http
POST   /api/v1/identity/permissions                    # Create permission
GET    /api/v1/identity/permissions                    # List all
GET    /api/v1/identity/permissions/:id                # Get by ID
GET    /api/v1/identity/permissions/name/:name        # Get by name
PUT    /api/v1/identity/permissions/:id                # Update
DELETE /api/v1/identity/permissions/:id                # Delete
GET    /api/v1/identity/roles/:id/permissions          # Get role permissions
```

## JWT Integration

### Token Structure

Roles embedded in JWT claims for stateless authorization:

```json
{
  "user_id": "01JGABC...",
  "email": "admin@example.com",
  "roles": ["admin", "manager"],
  "exp": 1735488246,
  "iat": 1735487346
}
```

### Login Flow

```go
// 1. Authenticate user
user, err := h.useCase.Authenticate(ctx, req.Email, req.Password)

// 2. Load user roles from database
roles := user.Roles() // ["admin", "manager"]

// 3. Generate JWT with roles
accessToken, err := h.jwtManager.GenerateToken(user.ID.String(), user.Email.Value(), roles)

// 4. Return token
response.Success(c, LoginResponse{
    AccessToken:  accessToken,
    RefreshToken: refreshToken,
    User:         toUserDTO(user),
})
```

## Middleware

### Authentication Middleware

Validates JWT and extracts claims:

```go
router.Use(jwt.AuthMiddleware(jwtManager))
```

### Authorization Middleware

**RequireRole** - Single role required:
```go
admin := router.Group("/admin")
admin.Use(jwt.RequireRole("admin"))
{
    admin.GET("/users", handler.ListUsers)
}
```

**RequireAnyRole** - At least one role (OR logic):
```go
moderation := router.Group("/moderation")
moderation.Use(jwt.RequireAnyRole("admin", "moderator"))
{
    moderation.GET("/reports", handler.ListReports)
}
```

**RequireAllRoles** - All roles required (AND logic):
```go
superadmin := router.Group("/superadmin")
superadmin.Use(jwt.RequireAllRoles("admin", "superadmin"))
{
    superadmin.GET("/system/config", handler.GetSystemConfig)
}
```

## Usage Examples

### Check Roles in Handler

```go
func (h *UserHandler) SuspendUser(c *gin.Context) {
    // Extract current user from JWT
    claims := jwt.MustGetClaims(c)
    
    // Check admin role
    if !claims.HasRole("admin") {
        response.Error(c, http.StatusForbidden, "FORBIDDEN", "Admin role required")
        return
    }
    
    // Business logic...
}
```

### Multiple Role Checks

```go
// Check any role (OR)
if claims.HasAnyRole("admin", "moderator") {
    // Allow access
}

// Check all roles (AND)
if claims.HasAllRoles("admin", "superadmin") {
    // Allow access
}
```

## Testing

**Test Coverage:**
- Role entity: 15+ tests
- Permission entity: 12+ tests
- Role use case: 20+ tests
- Permission use case: 18+ tests
- HTTP handlers: 14 smoke tests
- Repository: 25+ integration tests

**Example Test:**

```go
func TestRoleUseCase_CreateRole(t *testing.T) {
    mockRepo := new(MockRoleRepository)
    uc := NewUseCase(mockRepo)
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    
    role, err := uc.CreateRole(ctx, "manager", "Manager role")
    
    assert.NoError(t, err)
    assert.NotNil(t, role)
    assert.Equal(t, "manager", role.Name)
}
```

## Best Practices

### Security

1. **Protect system roles** - `is_system=true` prevents deletion
2. **Validate permissions** - Check resource:action format
3. **Audit changes** - Log all role/permission assignments
4. **Least privilege** - Grant minimum required permissions
5. **Regular review** - Audit user roles periodically

### Performance

1. **Cache roles** - Store in Redis with TTL
2. **JWT roles** - Avoid DB lookups on every request
3. **Batch operations** - Assign multiple roles at once
4. **Index properly** - Unique indexes on name fields

### Development

1. **Seed data first** - Run migrations before app start
2. **Test with real DB** - Integration tests for RBAC
3. **Document permissions** - Clear descriptions
4. **Use constants** - Define permission strings as constants

## Next Steps

- [JWT Package](/packages/jwt) - Token generation and validation
- [Identity Context](/contexts/identity) - User management
- [GitHub RBAC Docs](https://github.com/basilex/promenade/blob/dev/docs/RBAC.md) - Complete 891-line guide
