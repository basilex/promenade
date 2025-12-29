# Role-Based Access Control (RBAC) Implementation

**Complete guide to RBAC architecture and implementation in Promenade Platform**

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Database Schema](#database-schema)
- [Domain Model](#domain-model)
- [API Reference](#api-reference)
- [JWT Integration](#jwt-integration)
- [Middleware](#middleware)
- [Use Cases](#use-cases)
- [Testing](#testing)
- [Best Practices](#best-practices)

---

## Overview

Promenade implements **Role-Based Access Control (RBAC)** as a core security mechanism for authorization. RBAC provides fine-grained access control through a combination of **Roles** and **Permissions**.

### Key Concepts

**Role** - Named collection of permissions assigned to users (e.g., "admin", "manager", "user")

**Permission** - Specific action on a resource (e.g., "users:create", "customers:delete")

**User-Role Assignment** - Many-to-many relationship between users and roles

**Role-Permission Assignment** - Many-to-many relationship between roles and permissions

### Benefits

- **Separation of Concerns**: Authorization logic decoupled from business logic
- **Flexibility**: Easy to add/modify roles and permissions without code changes
- **Scalability**: Supports complex permission hierarchies
- **Audit Trail**: Track who has what permissions and when assigned
- **JWT Integration**: Roles embedded in JWT tokens for stateless authorization

---

## Architecture

### Bounded Context

RBAC is implemented within the **Identity Context** as two aggregates:

```
internal/contexts/identity/
 role/
    entity.go                    # Role aggregate root
    repository.go                # Repository interface
    usecase.go                   # Business logic
    usecase_test.go             # Use case tests
    adapter/
        http/handler/
            role_handler.go      # HTTP handlers
            dto/
                role_dto.go      # Data transfer objects
        repository/postgres/
            role_repository.go   # PostgreSQL implementation
 permission/
     entity.go                    # Permission aggregate root
     repository.go                # Repository interface
     usecase.go                   # Business logic
     usecase_test.go             # Use case tests
     adapter/
         http/handler/
             permission_handler.go  # HTTP handlers
             dto/
                 permission_dto.go  # Data transfer objects
         repository/postgres/
             permission_repository.go  # PostgreSQL implementation
```

### Aggregate Roots

**Role Aggregate:**
- ID (UUID v7)
- Name (unique, e.g., "admin")
- Description
- IsSystem flag (prevents deletion of system roles)
- Permissions collection (via junction table)
- Soft delete support

**Permission Aggregate:**
- ID (UUID v7)
- Name (computed: "resource:action")
- Resource (e.g., "users", "customers")
- Action (e.g., "create", "read", "update", "delete")
- Description
- No soft delete (permissions are permanent)

---

## Database Schema

### Tables

```sql
-- Permissions table
CREATE TABLE identity_permissions (
    id          UUID PRIMARY KEY DEFAULT uuid_v7(),
    name        VARCHAR(101) GENERATED ALWAYS AS (resource || ':' || action) STORED NOT NULL,
    resource    VARCHAR(50) NOT NULL,
    action      VARCHAR(50) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at  TIMESTAMP WITH TIME ZONE,
    UNIQUE(resource, action)
);

-- Roles table
CREATE TABLE identity_roles (
    id          UUID PRIMARY KEY DEFAULT uuid_v7(),
    name        VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_system   BOOLEAN DEFAULT false,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at  TIMESTAMP WITH TIME ZONE
);

-- User-Role junction table
CREATE TABLE identity_user_roles (
    user_id     UUID NOT NULL REFERENCES identity_users(id) ON DELETE CASCADE,
    role_id     UUID NOT NULL REFERENCES identity_roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    assigned_by UUID REFERENCES identity_users(id),
    PRIMARY KEY (user_id, role_id)
);

-- Role-Permission junction table
CREATE TABLE identity_role_permissions (
    role_id       UUID NOT NULL REFERENCES identity_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES identity_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);
```

### Key Features

**GENERATED ALWAYS Column:**
- `identity_permissions.name` is computed as `resource || ':' || action`
- Automatically maintained by PostgreSQL
- Ensures consistency between resource/action and name

**Soft Delete:**
- Roles support soft delete (`deleted_at` column)
- Permissions do NOT support soft delete (permanent)

**Cascade Deletion:**
- Deleting user removes their role assignments
- Deleting role removes role-permission assignments

### Indexes

```sql
-- Permissions indexes
CREATE INDEX idx_identity_permissions_resource ON identity_permissions(resource);
CREATE INDEX idx_identity_permissions_action ON identity_permissions(action);
CREATE INDEX idx_identity_permissions_deleted_at ON identity_permissions(deleted_at) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_identity_permissions_name ON identity_permissions(name) WHERE deleted_at IS NULL;

-- Roles indexes
CREATE INDEX idx_identity_roles_name ON identity_roles(name);
CREATE INDEX idx_identity_roles_deleted_at ON identity_roles(deleted_at) WHERE deleted_at IS NULL;

-- User-Role indexes
CREATE INDEX idx_identity_user_roles_user_id ON identity_user_roles(user_id);
CREATE INDEX idx_identity_user_roles_role_id ON identity_user_roles(role_id);

-- Role-Permission indexes
CREATE INDEX idx_identity_role_permissions_role_id ON identity_role_permissions(role_id);
CREATE INDEX idx_identity_role_permissions_permission_id ON identity_role_permissions(permission_id);
```

---

## Domain Model

### Permission Entity

```go
package permission

import (
    "time"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type Permission struct {
    ID          uuidv7.UUID
    Name        string    // Computed: "resource:action"
    Resource    string    // e.g., "users", "customers"
    Action      string    // e.g., "create", "read", "update", "delete"
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Factory method
func NewPermission(resource, action, description string) (*Permission, error) {
    if resource == "" || action == "" {
        return nil, fmt.Errorf("resource and action are required")
    }
    
    return &Permission{
        ID:          uuidv7.New(),
        Resource:    resource,
        Action:      action,
        Description: description,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }, nil
}

// Name is computed by database GENERATED ALWAYS column
func (p *Permission) ComputedName() string {
    return p.Resource + ":" + p.Action
}
```

### Role Entity

```go
package role

import (
    "time"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type Role struct {
    ID          uuidv7.UUID
    Name        string
    Description string
    IsSystem    bool
    Permissions []uuidv7.UUID  // Permission IDs
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   *time.Time
}

// Factory method
func NewRole(name, description string) (*Role, error) {
    if name == "" {
        return nil, fmt.Errorf("role name is required")
    }
    
    return &Role{
        ID:          uuidv7.New(),
        Name:        name,
        Description: description,
        IsSystem:    false,
        Permissions: []uuidv7.UUID{},
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }, nil
}

func (r *Role) AssignPermission(permissionID uuidv7.UUID) {
    r.Permissions = append(r.Permissions, permissionID)
}

func (r *Role) CanDelete() bool {
    return !r.IsSystem && r.DeletedAt == nil
}
```

---

## API Reference

### Permission Endpoints

**Base URL:** `/api/v1/identity/permissions`

All endpoints require JWT authentication.

#### Create Permission

```http
POST /api/v1/identity/permissions
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "resource": "customers",
  "action": "export",
  "description": "Export customer data to CSV"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "019b6956-cd52-7874-8354-e67ed4745a36",
    "name": "customers:export",
    "resource": "customers",
    "action": "export",
    "description": "Export customer data to CSV",
    "created_at": "2025-12-29T10:00:00Z"
  }
}
```

#### List Permissions

```http
GET /api/v1/identity/permissions
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "id": "...",
      "name": "users:create",
      "resource": "users",
      "action": "create",
      "description": "Create new users"
    },
    ...
  ]
}
```

#### Get Permission by ID

```http
GET /api/v1/identity/permissions/:id
Authorization: Bearer <access_token>
```

#### Get Permission by Name

```http
GET /api/v1/identity/permissions/by-name/:name
Authorization: Bearer <access_token>

# Example: GET /api/v1/identity/permissions/by-name/users:create
```

#### Update Permission

```http
PUT /api/v1/identity/permissions/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "description": "Updated description"
}
```

Note: `resource` and `action` cannot be modified (name is computed).

#### Delete Permission

```http
DELETE /api/v1/identity/permissions/:id
Authorization: Bearer <access_token>
```

#### Get Role Permissions

```http
GET /api/v1/identity/permissions/role/:roleId
Authorization: Bearer <access_token>
```

Returns all permissions assigned to a specific role.

---

### Role Endpoints

**Base URL:** `/api/v1/identity/roles`

All endpoints require JWT authentication.

#### Create Role

```http
POST /api/v1/identity/roles
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "developer",
  "description": "Development team role"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "019b6958-a1f3-7e94-8d2a-5c3e8b9f4a7d",
    "name": "developer",
    "description": "Development team role",
    "is_system": false,
    "created_at": "2025-12-29T10:05:00Z"
  }
}
```

#### List Roles

```http
GET /api/v1/identity/roles
Authorization: Bearer <access_token>
```

#### Get Role by ID

```http
GET /api/v1/identity/roles/:id
Authorization: Bearer <access_token>
```

#### Get Role by Name

```http
GET /api/v1/identity/roles/by-name/:name
Authorization: Bearer <access_token>

# Example: GET /api/v1/identity/roles/by-name/admin
```

#### Update Role

```http
PUT /api/v1/identity/roles/:id
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "description": "Updated description"
}
```

Note: `name` cannot be modified for system roles.

#### Delete Role

```http
DELETE /api/v1/identity/roles/:id
Authorization: Bearer <access_token>
```

Note: System roles (`is_system = true`) cannot be deleted.

#### Get User Roles

```http
GET /api/v1/identity/roles/user/:userId
Authorization: Bearer <access_token>
```

Returns all roles assigned to a specific user.

---

## Seed Data

### System Roles

5 system roles are created during migration:

| Role        | Description                         | Permissions Count |
|-------------|-------------------------------------|-------------------|
| superadmin  | Full system access                  | 30 (all)          |
| admin       | Administrative access               | 25                |
| manager     | Management functions                | 15                |
| user        | Standard user access                | 8                 |
| guest       | Read-only access                    | 5                 |

### Default Permissions

29+ permissions are seeded across resources:

**Wildcard Permissions:**
- `*:*` - Full system access

**User Management:**
- `users:create`, `users:read`, `users:update`, `users:delete`
- `users:suspend`, `users:ban`, `users:activate`

**Role Management:**
- `roles:create`, `roles:read`, `roles:update`, `roles:delete`

**Permission Management:**
- `permissions:create`, `permissions:read`, `permissions:update`, `permissions:delete`

**Contact Management:**
- `contacts:create`, `contacts:read`, `contacts:update`, `contacts:delete`

**Profile Management:**
- `profiles:create`, `profiles:read`, `profiles:update`, `profiles:delete`

**Customer Management:**
- `customers:create`, `customers:read`, `customers:update`, `customers:delete`
- `customers:export`

---

## JWT Integration

### Token Claims

JWT tokens include user roles:

```go
type Claims struct {
    UserID string   `json:"user_id"`
    Email  string   `json:"email"`
    Roles  []string `json:"roles"`  // ["admin", "manager"]
    jwt.RegisteredClaims
}
```

### Login Flow

```go
func (h *UserHandler) Login(c *gin.Context) {
    // 1. Authenticate user
    user, err := h.usecase.Authenticate(ctx, email, password)
    
    // 2. Load user roles from database
    roles, err := h.roleRepo.GetUserRoles(ctx, user.ID)
    roleNames := extractRoleNames(roles)
    
    // 3. Generate JWT with roles
    tokenPair, err := h.jwtManager.GenerateTokenPair(
        user.ID, 
        user.Email, 
        roleNames,  // ["admin"]
    )
    
    // 4. Return tokens
    response.Success(c, tokenPair)
}
```

### Token Validation

```go
// Middleware extracts roles from JWT
func AuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c.GetHeader("Authorization"))
        
        claims, err := jwtManager.ValidateAccessToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, ...)
            return
        }
        
        // Store claims in context
        c.Set("jwt_claims", claims)
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

---

## Middleware

### Authentication Middleware

```go
router.Use(jwt.AuthMiddleware(jwtManager))
```

Validates JWT token and stores claims in context.

### Authorization Middleware

#### RequireRole

Requires user to have specific role:

```go
admin := router.Group("/admin")
admin.Use(jwt.AuthMiddleware(jwtManager))
admin.Use(jwt.RequireRole("admin"))
{
    admin.GET("/users", handler.ListUsers)
}
```

#### RequireAnyRole

Requires user to have at least one role (OR logic):

```go
moderation := router.Group("/moderation")
moderation.Use(jwt.AuthMiddleware(jwtManager))
moderation.Use(jwt.RequireAnyRole("admin", "moderator"))
{
    moderation.GET("/reports", handler.ListReports)
}
```

#### RequireAllRoles

Requires user to have all roles (AND logic):

```go
superadmin := router.Group("/superadmin")
superadmin.Use(jwt.AuthMiddleware(jwtManager))
superadmin.Use(jwt.RequireAllRoles("admin", "superadmin"))
{
    superadmin.GET("/system/config", handler.GetSystemConfig)
}
```

### Usage in Handlers

```go
func (h *Handler) DeleteUser(c *gin.Context) {
    // Get current user from JWT
    claims := jwt.MustGetClaims(c)
    currentUserID, _ := claims.ExtractUserID()
    
    // Check if user has admin role
    if !claims.HasRole("admin") {
        response.Error(c, 403, "FORBIDDEN", "Admin role required")
        return
    }
    
    // Business logic
    targetUserID := c.Param("id")
    err := h.usecase.DeleteUser(ctx, currentUserID, targetUserID)
    ...
}
```

---

## Use Cases

### Permission Use Cases

```go
type IUseCase interface {
    CreatePermission(ctx context.Context, resource, action, description string) (*Permission, error)
    GetPermission(ctx context.Context, id uuidv7.UUID) (*Permission, error)
    GetPermissionByName(ctx context.Context, name string) (*Permission, error)
    ListPermissions(ctx context.Context) ([]*Permission, error)
    UpdatePermission(ctx context.Context, id uuidv7.UUID, description string) error
    DeletePermission(ctx context.Context, id uuidv7.UUID) error
    GetRolePermissions(ctx context.Context, roleID uuidv7.UUID) ([]*Permission, error)
}
```

### Role Use Cases

```go
type IUseCase interface {
    CreateRole(ctx context.Context, name, description string) (*Role, error)
    GetRole(ctx context.Context, id uuidv7.UUID) (*Role, error)
    GetRoleByName(ctx context.Context, name string) (*Role, error)
    ListRoles(ctx context.Context) ([]*Role, error)
    UpdateRole(ctx context.Context, id uuidv7.UUID, description string) error
    DeleteRole(ctx context.Context, id uuidv7.UUID) error
    GetUserRoles(ctx context.Context, userID uuidv7.UUID) ([]*Role, error)
    AssignRoleToUser(ctx context.Context, userID, roleID uuidv7.UUID, assignedBy uuidv7.UUID) error
    RemoveRoleFromUser(ctx context.Context, userID, roleID uuidv7.UUID) error
    AssignPermissionToRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error
    RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuidv7.UUID) error
}
```

---

## Testing

### Unit Tests

**Location:** In-place with entity/usecase files

```go
// internal/contexts/identity/permission/entity_test.go
func TestPermission_NewPermission(t *testing.T) {
    perm, err := NewPermission("users", "create", "Create users")
    
    assert.NoError(t, err)
    assert.Equal(t, "users", perm.Resource)
    assert.Equal(t, "create", perm.Action)
    assert.Equal(t, "users:create", perm.ComputedName())
}

// internal/contexts/identity/role/usecase_test.go
func TestUseCase_CreateRole(t *testing.T) {
    mockRepo := NewMockRoleRepository()
    uc := NewUseCase(mockRepo)
    
    role, err := uc.CreateRole(ctx, "developer", "Dev team")
    
    assert.NoError(t, err)
    assert.Equal(t, "developer", role.Name)
}
```

### Smoke Tests

**Location:** `test/smoke/contexts/identity/permission/`, `test/smoke/contexts/identity/role/`

Mock-based handler tests without database:

```go
func TestPermissionHandler_Create(t *testing.T) {
    gin.SetMode(gin.TestMode)
    mockUC := new(MockPermissionUseCase)
    handler := NewPermissionHandler(mockUC)
    
    // Setup mock expectations
    mockPerm := &permission.Permission{
        ID: uuidv7.New(),
        Resource: "users",
        Action: "create",
    }
    mockUC.On("CreatePermission", mock.Anything, "users", "create", "desc").
        Return(mockPerm, nil)
    
    // Make request
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/", body)
    
    handler.Create(c)
    
    assert.Equal(t, 201, w.Code)
    mockUC.AssertExpectations(t)
}
```

### Integration Tests

**Location:** `test/integration/contexts/identity/permission/`, `test/integration/contexts/identity/role/`

Full E2E tests with real database:

```go
func TestPermissionRepository_Create(t *testing.T) {
    db := integration.SetupTestDBWithCleanTables(t)
    repo := postgres.NewPermissionRepository(db.DB)
    ctx := context.Background()
    
    perm, _ := permission.NewPermission("users", "create", "desc")
    err := repo.Create(ctx, perm)
    
    require.NoError(t, err)
    
    // Verify name column is auto-generated
    retrieved, err := repo.GetByName(ctx, "users:create")
    require.NoError(t, err)
    assert.Equal(t, "users:create", retrieved.Name)
}
```

### Test Coverage

- Permission aggregate: 45+ tests
- Role aggregate: 50+ tests
- HTTP handlers: 14+ smoke tests
- Repository: 20+ integration tests

---

## Best Practices

### Permission Naming

**DO:**
- Use lowercase resource names: `users`, `customers`, `orders`
- Use CRUD actions: `create`, `read`, `update`, `delete`
- Use specific actions: `export`, `import`, `approve`
- Pattern: `resource:action` (e.g., `users:create`)

**DON'T:**
- Use camelCase: `Users:Create` ❌
- Use spaces: `users create` ❌
- Use unclear actions: `users:modify` ❌

### Role Design

**System Roles:**
- Mark as `is_system = true`
- Cannot be deleted
- Updated via migrations only

**Custom Roles:**
- Business-specific (e.g., "sales_rep", "accountant")
- Can be created/modified via API
- Deleted when no longer needed

### Security

**JWT Tokens:**
- Keep access token TTL short (15 minutes)
- Refresh tokens for long sessions
- Include minimal roles in JWT (avoid bloat)

**Middleware Order:**
```go
router.Use(jwt.AuthMiddleware(jwtManager))  // First: authenticate
router.Use(jwt.RequireRole("admin"))        // Then: authorize
```

**Database:**
- Soft delete roles (preserve audit trail)
- Hard delete permissions (never removed in production)
- Use indexes on junction tables for performance

### Performance

**Caching:**
- Cache role-permission mappings (Redis)
- TTL: 5-10 minutes
- Invalidate on role/permission changes

**Queries:**
- Use JOINs to fetch roles + permissions in one query
- Avoid N+1 queries when listing users with roles

---

## Migration

**File:** `migrations/identity/000003_authorization.up.sql`

Run migrations:

```bash
make migrate-identity
```

Or manually:

```bash
go run cmd/migrate/main.go --cmd=up --namespace=identity
```

**Important:** Migration includes seed data for system roles and permissions. Do not modify after initial deployment.

---

## Future Enhancements

**Planned Features:**

- [ ] Role hierarchy (parent roles)
- [ ] Time-based role assignments (temporary access)
- [ ] Permission groups (collections of related permissions)
- [ ] Audit log for role/permission changes
- [ ] UI for role management
- [ ] Export/import role configurations
- [ ] Permission templates for common scenarios

---

## Related Documentation

- [JWT Package](../pkg/jwt/README.md) - JWT authentication
- [Identity Context](../internal/contexts/identity/README.md) - Identity bounded context
- [Testing Patterns](TESTING_PATTERNS.md) - Testing guide
- [Main README](../README.md) - Project overview

---

**Last Updated:** December 29, 2025  
**Status:** Production-ready  
**Maintainer:** Promenade Team
