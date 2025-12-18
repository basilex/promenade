# Authorization Middleware Guide

Comprehensive guide to using RBAC (Role-Based Access Control) authorization middleware in Promenade.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Permission System](#permission-system)
- [Role System](#role-system)
- [Middleware Methods](#middleware-methods)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Error Handling](#error-handling)
- [Testing Authorization](#testing-authorization)

## Overview

The Authorization Middleware provides **flexible, fine-grained access control** for API endpoints using a permission-based RBAC system. It supports:

- [+] **Permission-based checks** - Granular control with `resource:action` format
- [+] **Role-based checks** - Quick checks for user roles (superadmin, admin, etc.)
- [+] **Wildcard permissions** - `*:*`, `posts:*`, `*:read` patterns
- [+] **Composite checks** - RequireAny, RequireAll for complex logic
- [+] **Clean separation** - Works independently from authentication middleware

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Request                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              RequireAuth Middleware                         │
│          (Validates JWT, sets user_id)                      │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│         RequirePermission Middleware                        │
│    1. Get user_id from context                              │
│    2. Query user's roles                                    │
│    3. Check role permissions (with wildcard support)        │
│    4. Allow/Deny request                                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                   Handler Function                          │
└─────────────────────────────────────────────────────────────┘
```

## Permission System

### Permission Format

Permissions follow the `resource:action` pattern:

```
resource:action
   │       │
   │       └─ Action: create, read, update, delete, manage, *
   └───────── Resource: posts, users, comments, roles, *
```

### Examples

| Permission          | Description                            |
| ------------------- | -------------------------------------- |
| `posts:create`      | Can create posts                       |
| `posts:read`        | Can read posts                         |
| `posts:*`           | Can perform any action on posts        |
| `*:read`            | Can read any resource                  |
| `*:*`               | Can perform any action on any resource |
| `users:ban`         | Can ban users (custom action)          |
| `comments:moderate` | Can moderate comments                  |

### Wildcard Permissions

Wildcards provide powerful permission inheritance:

```go
// User has permission "posts:*"
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "posts:update")  // [+] TRUE
HasPermission(userID, "posts:delete")  // [+] TRUE
HasPermission(userID, "users:create")  // [X] FALSE

// User has permission "*:read"
HasPermission(userID, "posts:read")    // [+] TRUE
HasPermission(userID, "users:read")    // [+] TRUE
HasPermission(userID, "posts:create")  // [X] FALSE

// User has permission "*:*" (superadmin)
HasPermission(userID, "posts:create")  // [+] TRUE
HasPermission(userID, "users:delete")  // [+] TRUE
HasPermission(userID, "anything:anything") // [+] TRUE
```

## Role System

### System Roles

5 predefined system roles with different permission levels:

| Role         | Display Name        | Permissions            | Use Case                      |
| ------------ | ------------------- | ---------------------- | ----------------------------- |
| `superadmin` | Super Administrator | `*:*` (all)            | Full system access            |
| `admin`      | Administrator       | Most resources         | System administration         |
| `moderator`  | Moderator           | Content moderation     | Content review & moderation   |
| `user`       | User                | Own content management | Regular users                 |
| `guest`      | Guest               | Read-only access       | Unauthenticated/limited users |

### Role Permissions Breakdown

**Superadmin** (`*:*`):

- Full access to everything
- Cannot be deleted (system role)

**Admin**:

```
users:create, users:read, users:update, users:delete, users:ban, users:suspend
roles:read, roles:assign
permissions:read
posts:*, comments:*, profiles:*
```

**Moderator**:

```
posts:read, posts:update, posts:delete
comments:read, comments:update, comments:delete, comments:moderate
users:read, users:suspend
```

**User**:

```
posts:create, posts:read, posts:update (own), posts:delete (own)
comments:create, comments:read, comments:update (own), comments:delete (own)
profiles:read, profiles:update (own)
```

**Guest**:

```
posts:read, comments:read, profiles:read
```

## Middleware Methods

### RequirePermission

Check if user has a **single specific permission**.

```go
func (m *AuthorizationMiddleware) RequirePermission(permission string) gin.HandlerFunc
```

**Usage:**

```go
router.POST("/posts",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

**Returns:**

- `200 OK` - User has permission, proceeds to handler
- `401 Unauthorized` - User not authenticated
- `403 Forbidden` - User lacks permission
- `500 Internal Server Error` - Database error checking permissions

---

### RequireAnyPermission

Check if user has **at least one** of the specified permissions (OR logic).

```go
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc
```

**Usage:**

```go
// Allow if user can read OR moderate comments
router.GET("/comments/flagged",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission("comments:read", "comments:moderate"),
    handler.GetFlaggedComments,
)
```

**Use Cases:**

- Alternative permissions (admin OR moderator)
- Feature access with multiple entry points
- Gradual permission upgrades

---

### RequireAllPermissions

Check if user has **all** specified permissions (AND logic).

```go
func (m *AuthorizationMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc
```

**Usage:**

```go
// Require both publish AND schedule permissions
router.POST("/posts/schedule",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions("posts:create", "posts:schedule"),
    handler.SchedulePost,
)
```

**Use Cases:**

- Compound operations requiring multiple permissions
- Sensitive operations needing layered checks
- Feature combinations

---

### RequireRole

Check if user has a **specific role** by name.

```go
func (m *AuthorizationMiddleware) RequireRole(roleName string) gin.HandlerFunc
```

**Usage:**

```go
// Only administrators can access
router.GET("/admin/dashboard",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.GetAdminDashboard,
)
```

**Note:** Prefer `RequirePermission` over `RequireRole` for better flexibility.

---

### RequireAnyRole

Check if user has **at least one** of the specified roles (OR logic).

```go
func (m *AuthorizationMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc
```

**Usage:**

```go
// Allow admins OR moderators
router.GET("/moderation/queue",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyRole("admin", "moderator"),
    handler.GetModerationQueue,
)
```

**Use Cases:**

- Administrative areas with multiple role levels
- Feature access for similar roles
- Legacy systems transitioning from role-based to permission-based

## Usage Examples

### Example 1: Basic CRUD Protection

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")

    // Public read access (no auth required)
    posts.GET("", handler.ListPosts)
    posts.GET("/:id", handler.GetPost)

    // Authenticated operations
    posts.Use(r.authMiddleware.RequireAuth())
    {
        // Specific permissions for each operation
        posts.POST("",
            r.authzMiddleware.RequirePermission("posts:create"),
            handler.CreatePost,
        )

        posts.PUT("/:id",
            r.authzMiddleware.RequirePermission("posts:update"),
            handler.UpdatePost,
        )

        posts.DELETE("/:id",
            r.authzMiddleware.RequirePermission("posts:delete"),
            handler.DeletePost,
        )
    }
}
```

### Example 2: Admin-Only Endpoints

```go
func (r *UserRouter) Setup(api *gin.RouterGroup) {
    users := api.Group("/users")
    users.Use(r.authMiddleware.RequireAuth())

    // Regular user operations
    users.GET("/me", handler.GetMe)
    users.PUT("/me", handler.UpdateProfile)

    // Admin-only operations
    admin := users.Group("")
    admin.Use(r.authzMiddleware.RequirePermission("users:manage"))
    {
        admin.GET("", handler.ListAllUsers)
        admin.POST("/:id/ban", handler.BanUser)
        admin.POST("/:id/suspend", handler.SuspendUser)
    }
}
```

### Example 3: Flexible Access with Multiple Permissions

```go
func (r *CommentRouter) Setup(api *gin.RouterGroup) {
    comments := api.Group("/comments")

    // View comments - any of these permissions works
    comments.GET("/:id",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAnyPermission(
            "comments:read",
            "comments:moderate",
            "*:read",
        ),
        handler.GetComment,
    )

    // Moderate comments - need both read and moderate
    comments.POST("/:id/moderate",
        r.authMiddleware.RequireAuth(),
        r.authzMiddleware.RequireAllPermissions(
            "comments:read",
            "comments:moderate",
        ),
        handler.ModerateComment,
    )
}
```

### Example 4: Role-Based Dashboard Access

```go
func (r *DashboardRouter) Setup(api *gin.RouterGroup) {
    dashboards := api.Group("/dashboard")
    dashboards.Use(r.authMiddleware.RequireAuth())

    // User dashboard - any authenticated user
    dashboards.GET("/user", handler.GetUserDashboard)

    // Moderator dashboard - moderators and admins
    dashboards.GET("/moderator",
        r.authzMiddleware.RequireAnyRole("moderator", "admin", "superadmin"),
        handler.GetModeratorDashboard,
    )

    // Admin dashboard - admins only
    dashboards.GET("/admin",
        r.authzMiddleware.RequireRole("admin"),
        handler.GetAdminDashboard,
    )
}
```

### Example 5: Complex Business Logic

```go
func (r *PostRouter) Setup(api *gin.RouterGroup) {
    posts := api.Group("/posts")
    posts.Use(r.authMiddleware.RequireAuth())

    // Publishing requires both create and publish permissions
    posts.POST("/:id/publish",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
        ),
        handler.PublishPost,
    )

    // Scheduling requires create, publish, AND schedule
    posts.POST("/:id/schedule",
        r.authzMiddleware.RequireAllPermissions(
            "posts:create",
            "posts:publish",
            "posts:schedule",
        ),
        handler.SchedulePost,
    )

    // Featuring requires moderator OR admin role + feature permission
    posts.POST("/:id/feature",
        r.authzMiddleware.RequireAnyRole("admin", "moderator"),
        r.authzMiddleware.RequirePermission("posts:feature"),
        handler.FeaturePost,
    )
}
```

### Example 6: Migration Setup

Initialize authorization in your router setup:

```go
// cmd/api/main.go or router initialization
func setupRouters(
    authMiddleware *middleware.AuthMiddleware,
    authzMiddleware *middleware.AuthorizationMiddleware,
) *gin.Engine {
    r := gin.New()

    // Public routes
    api := r.Group("/api/v1")

    // Auth routes (no authorization needed)
    authRouter := router.NewAuthRouter(authHandler, authMiddleware)
    authRouter.Setup(api)

    // Protected routes with authorization
    postRouter := router.NewPostRouter(postHandler, authMiddleware, authzMiddleware)
    postRouter.Setup(api)

    userRouter := router.NewUserRouter(userHandler, authMiddleware, authzMiddleware)
    userRouter.Setup(api)

    return r
}
```

## Best Practices

### 1. Always Use RequireAuth First

Authorization middleware requires authentication context:

```go
// [+] CORRECT - Auth before authorization
router.POST("/posts",
    authMiddleware.RequireAuth(),           // First: authenticate
    authzMiddleware.RequirePermission(...), // Then: authorize
    handler.CreatePost,
)

// [X] WRONG - Authorization without authentication
router.POST("/posts",
    authzMiddleware.RequirePermission(...), // Will fail - no user_id
    handler.CreatePost,
)
```

### 2. Prefer Permissions Over Roles

Permissions provide better flexibility and maintainability:

```go
// [+] BETTER - Permission-based (flexible)
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequirePermission("posts:delete"),
    handler.DeletePost,
)

// [!] ACCEPTABLE but less flexible - Role-based
router.DELETE("/posts/:id",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireRole("admin"),
    handler.DeletePost,
)
```

**Why?**

- Adding new roles doesn't require code changes
- Permissions can be reassigned without touching code
- More granular control

### 3. Use Descriptive Permission Names

```go
// [+] GOOD - Clear intent
"posts:create"
"posts:publish"
"posts:feature"
"posts:schedule"

// [X] BAD - Vague
"posts:manage"  // What does "manage" mean?
"posts:admin"   // Too generic
```

### 4. Leverage Wildcards for Admin Roles

```go
// In your migration/seeding
INSERT INTO permissions (resource, action) VALUES
    ('*', '*'),           -- Superadmin: everything
    ('posts', '*'),       -- Content admin: all post operations
    ('*', 'read');        -- Viewer: read everything
```

### 5. Group Related Permissions

```go
// Group by feature area
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// All post write operations require posts:* or posts:write
write := posts.Group("")
write.Use(authzMiddleware.RequirePermission("posts:write"))
{
    write.POST("", handler.CreatePost)
    write.PUT("/:id", handler.UpdatePost)
    write.DELETE("/:id", handler.DeletePost)
}

// Public read operations
posts.GET("", handler.ListPosts)
posts.GET("/:id", handler.GetPost)
```

### 6. Handle Owner-Only Resources in Handler

Don't use authorization middleware for owner checks:

```go
// [+] CORRECT - Check ownership in handler
func (h *PostHandler) UpdatePost(c *gin.Context) {
    userID := middleware.GetUserIDOrPanic(c)
    postID := c.Param("id")

    post, err := h.postUC.GetByID(c.Request.Context(), postID)
    if err != nil {
        response.Error(c, http.StatusNotFound, "post not found", err)
        return
    }

    // Check ownership OR admin permission
    if post.UserID != userID {
        hasAdmin, _ := h.roleUC.HasPermission(c.Request.Context(), userID, "posts:*")
        if !hasAdmin {
            response.Error(c, http.StatusForbidden, "can only update own posts", nil)
            return
        }
    }

    // Proceed with update...
}

// [X] WRONG - Trying to check ownership in middleware
// Middleware doesn't have access to resource details
```

### 7. Use RequireAny for Fallback Permissions

```go
// Allow operation if user has specific permission OR is admin
router.POST("/posts/:id/feature",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAnyPermission(
        "posts:feature",  // Specific permission
        "posts:*",        // Full post access
        "*:*",            // Superadmin
    ),
    handler.FeaturePost,
)
```

## Common Patterns

### Pattern 1: Admin Override

Allow admins to bypass ownership checks:

```go
// Any admin permission overrides ownership
authzMiddleware.RequireAnyPermission(
    "posts:update",  // Regular user permission
    "posts:*",       // Post admin
    "*:*",           // Superadmin
)
```

### Pattern 2: Graduated Permissions

Different permission levels for same resource:

```go
// Level 1: Basic read
authzMiddleware.RequirePermission("posts:read")

// Level 2: Read + write
authzMiddleware.RequireAllPermissions("posts:read", "posts:write")

// Level 3: Full access
authzMiddleware.RequirePermission("posts:*")
```

### Pattern 3: Cross-Resource Permissions

Operations affecting multiple resources:

```go
// Publishing a post might require both post and media permissions
router.POST("/posts/:id/publish",
    authMiddleware.RequireAuth(),
    authzMiddleware.RequireAllPermissions(
        "posts:publish",
        "media:attach",  // If post includes images
    ),
    handler.PublishPost,
)
```

### Pattern 4: Conditional Authorization

Different permissions for different endpoints:

```go
posts := api.Group("/posts")
posts.Use(authMiddleware.RequireAuth())

// Draft posts - create permission only
posts.POST("/drafts",
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreateDraft,
)

// Published posts - create + publish permissions
posts.POST("/publish",
    authzMiddleware.RequireAllPermissions("posts:create", "posts:publish"),
    handler.CreateAndPublish,
)
```

## Error Handling

### HTTP Status Codes

| Status | Meaning               | Cause                                   |
| ------ | --------------------- | --------------------------------------- |
| 401    | Unauthorized          | User not authenticated (no JWT)         |
| 403    | Forbidden             | User authenticated but lacks permission |
| 500    | Internal Server Error | Database error checking permissions     |

### Error Response Format

```json
{
  "error": "insufficient permissions",
  "message": "You don't have permission to perform this action"
}
```

### Client-Side Handling

```typescript
// TypeScript/JavaScript example
try {
  const response = await fetch("/api/v1/posts", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(postData),
  });

  if (response.status === 401) {
    // Redirect to login
    window.location.href = "/login";
  } else if (response.status === 403) {
    // Show "insufficient permissions" message
    showError("You do not have permission to create posts");
  } else if (response.ok) {
    // Success
    const post = await response.json();
  }
} catch (error) {
  console.error("Request failed:", error);
}
```

## Testing Authorization

### Unit Testing Middleware

```go
func TestAuthorizationMiddleware_RequirePermission(t *testing.T) {
    // Setup
    mockRoleUC := mocks.NewMockRoleUseCase(t)
    authzMiddleware := middleware.NewAuthorizationMiddleware(mockRoleUC)

    t.Run("allows user with permission", func(t *testing.T) {
        // Create test context with user_id
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        // Mock permission check - returns true
        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(true, nil)

        // Create handler chain
        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        // Execute
        handler(c)

        // Assert
        assert.Equal(t, 200, w.Code)
    })

    t.Run("denies user without permission", func(t *testing.T) {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Set("user_id", testUserID)

        mockRoleUC.EXPECT().
            HasPermission(mock.Anything, testUserID, "posts:create").
            Return(false, nil)

        handler := authzMiddleware.RequirePermission("posts:create")(func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })

        handler(c)

        assert.Equal(t, 403, w.Code)
    })
}
```

### Integration Testing

```go
func TestPostEndpoints_Authorization(t *testing.T) {
    testDB := helpers.SetupTestDB(t)
    defer testDB.Close()
    defer testDB.CleanupTables(t)

    // Create test users with different roles
    adminUser := helpers.UserFixture(t, testDB.DB)
    regularUser := helpers.UserFixture(t, testDB.DB)

    // Assign roles
    assignRole(t, testDB, adminUser.ID, "admin")
    assignRole(t, testDB, regularUser.ID, "user")

    // Generate tokens
    adminToken := generateToken(t, adminUser.ID)
    userToken := generateToken(t, regularUser.ID)

    t.Run("admin can delete any post", func(t *testing.T) {
        post := createTestPost(t, testDB, regularUser.ID)

        req := httptest.NewRequest("DELETE", "/api/v1/posts/"+post.ID, nil)
        req.Header.Set("Authorization", "Bearer "+adminToken)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, 200, w.Code)
    })

    t.Run("user cannot delete others' posts", func(t *testing.T) {
        post := createTestPost(t, testDB, adminUser.ID)

        req := httptest.NewRequest("DELETE", "/api/v1/posts/"+post.ID, nil)
        req.Header.Set("Authorization", "Bearer "+userToken)

        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)

        assert.Equal(t, 403, w.Code)
    })
}
```

### Manual Testing with curl

```bash
# 1. Login and get token
TOKEN=$(curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}' \
  | jq -r '.data.access_token')

# 2. Test protected endpoint
curl -X POST http://localhost:8081/api/v1/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Post","content":"Content"}'

# Expected responses:
# 200 OK - Success
# 401 Unauthorized - Invalid/missing token
# 403 Forbidden - Insufficient permissions
```

## Troubleshooting

### Issue: 401 Unauthorized on Protected Endpoint

**Symptoms:**

```json
{
  "error": "user not authenticated"
}
```

**Causes:**

1. Missing `RequireAuth()` middleware before `RequirePermission()`
2. Invalid JWT token
3. Expired token

**Solution:**

```go
// Ensure auth middleware is applied first
router.POST("/posts",
    authMiddleware.RequireAuth(),  // ← Must be before authorization
    authzMiddleware.RequirePermission("posts:create"),
    handler.CreatePost,
)
```

### Issue: 403 Forbidden for Superadmin

**Symptoms:**
Superadmin user gets 403 on endpoints they should access.

**Causes:**

1. Wildcard permission (`*:*`) not properly checked
2. Role not assigned to user
3. Permission not assigned to role

**Solution:**

```sql
-- Verify superadmin has wildcard permission
SELECT r.name, p.resource, p.action
FROM roles r
JOIN role_permissions rp ON rp.role_id = r.id
JOIN permissions p ON p.id = rp.permission_id
WHERE r.name = 'superadmin';

-- Should return: name='superadmin', resource='*', action='*'

-- Verify user has superadmin role
SELECT u.email, r.name
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE u.id = '<user_uuid>';
```

### Issue: Database Performance with Permission Checks

**Symptoms:**
Slow response times on protected endpoints.

**Solution:**
Implement caching in RoleUseCase:

```go
// Use Redis/in-memory cache for permission checks
func (uc *RoleUseCase) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("user:%s:permission:%s", userID, permission)
    if cached, found := uc.cache.Get(cacheKey); found {
        return cached.(bool), nil
    }

    // Query database
    hasPermission, err := uc.repo.HasPermission(ctx, userID, permission)
    if err != nil {
        return false, err
    }

    // Cache for 5 minutes
    uc.cache.Set(cacheKey, hasPermission, 5*time.Minute)

    return hasPermission, nil
}
```

## Summary

The Authorization Middleware provides powerful, flexible access control for your API:

[+] **Permission-based** - Fine-grained control with `resource:action` format  
[+] **Wildcard support** - Powerful inheritance with `*` patterns  
[+] **Composite checks** - AND/OR logic for complex requirements  
[+] **Role shortcuts** - Quick role-based checks when needed  
[+] **Clean architecture** - Separates authorization from authentication  
[+] **Production-ready** - Battle-tested error handling and performance

**Quick Reference:**

```go
// Single permission check
RequirePermission("posts:create")

// Any of multiple permissions (OR)
RequireAnyPermission("posts:read", "posts:*", "*:*")

// All of multiple permissions (AND)
RequireAllPermissions("posts:create", "posts:publish")

// Role-based check
RequireRole("admin")

// Any of multiple roles (OR)
RequireAnyRole("admin", "moderator")
```

For more information, see:

- [RBAC Implementation](RBAC_IMPLEMENTATION.md)
- [Testing Guide](TESTING_GUIDE.md)
- [API Documentation](../README.md)
