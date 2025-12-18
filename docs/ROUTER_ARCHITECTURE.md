# Router Architecture

## Modular Router Design

The V1 API router uses a modular architecture to maintain scalability and maintainability as the API grows to 100+ endpoints.

## Structure

```
internal/adapter/http/v1/router/
├── router.go          # Main V1Router - aggregates all module routers
├── auth_router.go     # AuthRouter - handles /auth/* endpoints
├── rbac_router.go     # RBACRouter - handles /rbac/* endpoints (future)
└── profile_router.go  # ProfileRouter - handles /profiles/* endpoints (future)
```

## Design Principles

### 1. Module Isolation

Each module (auth, rbac, profiles, etc.) has its own router file that encapsulates:

- All endpoints for that module
- Module-specific middleware
- Public vs protected route separation
- Clear grouping by functionality

### 2. Main Router as Aggregator

The main `V1Router` in `router.go`:

- Simply aggregates module routers
- No route registration logic
- Easy to add/remove modules
- Clear overview of all modules

### 3. Shared Dependencies

Module routers receive their dependencies through constructor:

```go
func NewAuthRouter(
    authHandler *handler.AuthHandler,
    authMiddleware *middleware.AuthMiddleware,
) *AuthRouter
```

If modules need shared data/middleware, pass it through constructor.

## Example: Auth Router

### File: `auth_router.go`

```go
type AuthRouter struct {
    authHandler    *handler.AuthHandler
    authMiddleware *middleware.AuthMiddleware
}

func (r *AuthRouter) Setup(rg *gin.RouterGroup) {
    auth := rg.Group("/auth")

    // Public routes
    r.setupPublicRoutes(auth)

    // Protected routes
    protected := auth.Group("")
    protected.Use(r.authMiddleware.RequireAuth())
    r.setupProtectedRoutes(protected)
}
```

**Benefits:**

- All `/auth/*` logic in one place
- Internal organization with private methods
- Easy to understand auth flow
- Testable in isolation

## Adding New Module

To add a new module (e.g., RBAC):

### 1. Create `rbac_router.go`

```go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/internal/adapter/http/v1/handler"
    "github.com/basilex/promenade/internal/adapter/http/shared/middleware"
)

type RBACRouter struct {
    rbacHandler    *handler.RBACHandler
    authMiddleware *middleware.AuthMiddleware
}

func NewRBACRouter(
    rbacHandler *handler.RBACHandler,
    authMiddleware *middleware.AuthMiddleware,
) *RBACRouter {
    return &RBACRouter{
        rbacHandler:    rbacHandler,
        authMiddleware: authMiddleware,
    }
}

func (r *RBACRouter) Setup(rg *gin.RouterGroup) {
    rbac := rg.Group("/rbac")
    rbac.Use(r.authMiddleware.RequireAuth())

    // Roles
    roles := rbac.Group("/roles")
    r.setupRoleRoutes(roles)

    // Permissions
    permissions := rbac.Group("/permissions")
    r.setupPermissionRoutes(permissions)
}

func (r *RBACRouter) setupRoleRoutes(rg *gin.RouterGroup) {
    rg.GET("", r.rbacHandler.ListRoles)
    rg.POST("", r.rbacHandler.CreateRole)
    rg.GET("/:id", r.rbacHandler.GetRole)
    rg.PUT("/:id", r.rbacHandler.UpdateRole)
    rg.DELETE("/:id", r.rbacHandler.DeleteRole)
}

func (r *RBACRouter) setupPermissionRoutes(rg *gin.RouterGroup) {
    rg.GET("", r.rbacHandler.ListPermissions)
    rg.POST("/:role_id/permissions", r.rbacHandler.AssignPermissions)
}
```

### 2. Update `router.go`

```go
type V1Router struct {
    authRouter *AuthRouter
    rbacRouter *RBACRouter  // Add new module
}

func NewV1Router(
    authRouter *AuthRouter,
    rbacRouter *RBACRouter,  // Add parameter
) *V1Router {
    return &V1Router{
        authRouter: authRouter,
        rbacRouter: rbacRouter,
    }
}

func (r *V1Router) Setup(rg *gin.RouterGroup) {
    r.authRouter.Setup(rg)
    r.rbacRouter.Setup(rg)  // Register module
}
```

### 3. Update `cmd/api/main.go`

```go
// V1 Module Routers
authRouter := router.NewAuthRouter(authHandler, authMiddleware)
rbacRouter := router.NewRBACRouter(rbacHandler, authMiddleware)

// V1 Main Router
v1Router := router.NewV1Router(authRouter, rbacRouter)
v1Router.Setup(api)
```

## Sharing Data Between Modules

If modules need shared resources:

### Option 1: Pass Through Constructor

```go
type ProfileRouter struct {
    profileHandler *handler.ProfileHandler
    authMiddleware *middleware.AuthMiddleware
    cache          *cache.Cache  // Shared cache
}
```

### Option 2: Shared Context

```go
type RouterContext struct {
    AuthMiddleware *middleware.AuthMiddleware
    Cache          *cache.Cache
    RateLimiter    *ratelimit.Limiter
}

func NewAuthRouter(
    handler *handler.AuthHandler,
    ctx *RouterContext,
) *AuthRouter
```

## Route Organization Example

With 150 endpoints organized by modules:

```
/api/v1/
├── /auth/*                    (10 endpoints)
│   ├── POST /register
│   ├── POST /login
│   ├── POST /logout
│   └── ...
├── /rbac/*                    (20 endpoints)
│   ├── /roles/*              (10 endpoints)
│   └── /permissions/*        (10 endpoints)
├── /profiles/*                (25 endpoints)
│   ├── GET /profiles
│   ├── GET /profiles/:id
│   └── ...
├── /notifications/*           (15 endpoints)
├── /settings/*                (20 endpoints)
└── /analytics/*               (60 endpoints)
    ├── /analytics/users/*
    ├── /analytics/revenue/*
    └── ...
```

Each module in its own file, main router stays ~30 lines.

## Testing

Module routers are easily testable in isolation:

```go
func TestAuthRouter_Setup(t *testing.T) {
    router := gin.New()
    authRouter := NewAuthRouter(mockHandler, mockMiddleware)

    authRouter.Setup(router.Group("/api"))

    // Test route registration
    routes := router.Routes()
    assert.Contains(t, routes, "POST /api/auth/login")
}
```

## Benefits Summary

[+] **Scalability**: Easy to add 100+ endpoints without chaos
[+] **Maintainability**: Each module is self-contained
[+] **Readability**: Clear separation of concerns
[+] **Testability**: Test modules in isolation
[+] **Reusability**: Share middleware/context across modules
[+] **Team-friendly**: Multiple devs can work on different modules
[+] **Discoverability**: Easy to find where specific endpoints are defined

## Migration Path

Current: All routes in `router.go`
→ Step 1: Extract auth to `auth_router.go` [+] (Done)
→ Step 2: Create `rbac_router.go` when RBAC is implemented
→ Step 3: Continue pattern for new features
