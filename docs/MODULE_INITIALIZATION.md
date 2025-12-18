# Module Initialization Pattern

## Problem

As the application grows to 10-15+ modules, the `main.go` file becomes unwieldy with hundreds of lines of initialization code:

```go
// [X] BAD: Unscalable approach
userRepo := postgres.NewUserRepository(db)
sessionRepo := postgres.NewSessionRepository(db)
roleRepo := postgres.NewRoleRepository(db)
permissionRepo := postgres.NewPermissionRepository(db)
profileRepo := postgres.NewProfileRepository(db)
notificationRepo := postgres.NewNotificationRepository(db)
// ... 50 more repositories

authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, jwtManager)
rbacUseCase := usecase.NewRBACUseCase(roleRepo, permissionRepo, cache)
profileUseCase := usecase.NewProfileUseCase(profileRepo, userRepo, storage)
notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, queue)
// ... 50 more use cases

authHandler := handler.NewAuthHandler(authUseCase)
rbacHandler := handler.NewRBACHandler(rbacUseCase)
profileHandler := handler.NewProfileHandler(profileUseCase)
notificationHandler := handler.NewNotificationHandler(notificationUseCase)
// ... 50 more handlers

authRouter := router.NewAuthRouter(authHandler, authMiddleware)
rbacRouter := router.NewRBACRouter(rbacHandler, authMiddleware, rbacMiddleware)
profileRouter := router.NewProfileRouter(profileHandler, authMiddleware)
notificationRouter := router.NewNotificationRouter(notificationHandler, authMiddleware)
// ... 50 more routers
```

**Problems:**

- 300+ lines of boilerplate in main.go
- Hard to see what modules exist
- Difficult to track dependencies
- Each new module requires touching main.go in multiple places
- No encapsulation of module-specific logic

## Solution: Module Initializers

Each module provides a single initialization function that encapsulates its entire dependency chain.

### Pattern Structure

```
internal/adapter/http/v1/router/
├── router.go           # Main V1Router aggregator
├── auth_router.go      # AuthRouter routes
├── init_auth.go        # Auth module initializer ← NEW
├── rbac_router.go      # RBACRouter routes (future)
└── init_rbac.go        # RBAC module initializer (future)
```

### Implementation

#### 1. Module Initializer (`init_auth.go`)

```go
package router

import (
    "github.com/jmoiron/sqlx"
    "github.com/basilex/promenade/internal/adapter/http/shared/middleware"
    "github.com/basilex/promenade/internal/adapter/http/v1/handler"
    "github.com/basilex/promenade/internal/adapter/repository/postgres"
    "github.com/basilex/promenade/internal/usecase"
    jwtpkg "github.com/basilex/promenade/pkg/jwt"
)

// InitAuthModule initializes the complete auth module
// Encapsulates: repo → usecase → handler → router
func InitAuthModule(
    db *sqlx.DB,
    jwtManager *jwtpkg.JWTManager,
    authMiddleware *middleware.AuthMiddleware,
) *AuthRouter {
    // Repository layer
    userRepo := postgres.NewUserRepository(db)
    sessionRepo := postgres.NewSessionRepository(db)

    // Use case layer
    authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, jwtManager)

    // Handler layer
    authHandler := handler.NewAuthHandler(authUseCase)

    // Router layer
    return NewAuthRouter(authHandler, authMiddleware)
}
```

**Key principles:**

- Single function per module
- Takes only shared dependencies (db, middleware, config)
- Returns ready-to-use router
- Encapsulates internal wiring
- Clear dependency flow

#### 2. Clean main.go

```go
func main() {
    // ... config, db, jwt setup ...

    // Shared infrastructure
    authMiddleware := middleware.NewAuthMiddleware(jwtManager)
    cache := redis.NewCache(cfg.Redis)
    queue := rabbitmq.NewQueue(cfg.RabbitMQ)

    // Initialize modules (one line per module!)
    authRouter := router.InitAuthModule(db, jwtManager, authMiddleware)
    rbacRouter := router.InitRBACModule(db, authMiddleware, cache)
    profileRouter := router.InitProfileModule(db, authMiddleware, cache)
    notificationRouter := router.InitNotificationModule(db, authMiddleware, queue)
    analyticsRouter := router.InitAnalyticsModule(db, authMiddleware, cache)

    // Aggregate into main router
    v1Router := router.NewV1Router(
        authRouter,
        rbacRouter,
        profileRouter,
        notificationRouter,
        analyticsRouter,
    )

    // ... server setup ...
}
```

**Benefits:**

- [+] main.go stays ~100 lines regardless of module count
- [+] One line per module initialization
- [+] Clear module list
- [+] Module changes don't require main.go changes
- [+] Easy to see shared dependencies

## Example: Adding RBAC Module

### Step 1: Create `init_rbac.go`

```go
package router

import (
    "github.com/jmoiron/sqlx"
    "github.com/basilex/promenade/internal/adapter/http/shared/middleware"
    "github.com/basilex/promenade/internal/adapter/http/v1/handler"
    "github.com/basilex/promenade/internal/adapter/repository/postgres"
    "github.com/basilex/promenade/internal/usecase"
    "github.com/basilex/promenade/pkg/cache"
)

// InitRBACModule initializes the complete RBAC module
func InitRBACModule(
    db *sqlx.DB,
    authMiddleware *middleware.AuthMiddleware,
    cache cache.Cache,
) *RBACRouter {
    // Repositories
    roleRepo := postgres.NewRoleRepository(db)
    permissionRepo := postgres.NewPermissionRepository(db)
    userRoleRepo := postgres.NewUserRoleRepository(db)

    // Use cases
    rbacUseCase := usecase.NewRBACUseCase(
        roleRepo,
        permissionRepo,
        userRoleRepo,
        cache,
    )

    // Handlers
    rbacHandler := handler.NewRBACHandler(rbacUseCase)

    // Router with RBAC middleware
    rbacMiddleware := middleware.NewRBACMiddleware(rbacUseCase, cache)
    return NewRBACRouter(rbacHandler, authMiddleware, rbacMiddleware)
}
```

### Step 2: Update `main.go` (one line!)

```go
// Initialize modules
authRouter := router.InitAuthModule(db, jwtManager, authMiddleware)
rbacRouter := router.InitRBACModule(db, authMiddleware, cache)  // ← NEW

// Aggregate
v1Router := router.NewV1Router(authRouter, rbacRouter)  // ← add parameter
```

### Step 3: Update `router.go`

```go
type V1Router struct {
    authRouter *AuthRouter
    rbacRouter *RBACRouter  // ← NEW
}

func NewV1Router(
    authRouter *AuthRouter,
    rbacRouter *RBACRouter,  // ← NEW
) *V1Router {
    return &V1Router{
        authRouter: authRouter,
        rbacRouter: rbacRouter,
    }
}

func (r *V1Router) Setup(rg *gin.RouterGroup) {
    r.authRouter.Setup(rg)
    r.rbacRouter.Setup(rg)  // ← NEW
}
```

## Advanced: Shared Dependencies Pattern

For complex modules that share many dependencies:

### Option 1: Dependency Bundle

```go
type SharedDependencies struct {
    DB             *sqlx.DB
    JWTManager     *jwt.JWTManager
    AuthMiddleware *middleware.AuthMiddleware
    Cache          cache.Cache
    Queue          queue.Queue
    Logger         *logger.Logger
}

func InitAuthModule(deps *SharedDependencies) *AuthRouter {
    // Use deps.DB, deps.JWTManager, etc.
}
```

### Option 2: Builder Pattern

```go
type ModuleBuilder struct {
    db             *sqlx.DB
    jwtManager     *jwt.JWTManager
    authMiddleware *middleware.AuthMiddleware
}

func NewModuleBuilder(db *sqlx.DB, jwt *jwt.JWTManager) *ModuleBuilder {
    return &ModuleBuilder{
        db:         db,
        jwtManager: jwt,
    }
}

func (b *ModuleBuilder) InitAuthModule() *AuthRouter {
    // Has access to all common deps
}

func (b *ModuleBuilder) InitRBACModule(cache cache.Cache) *RBACRouter {
    // Can add module-specific deps
}
```

## Testing

Module initializers are easily testable:

```go
func TestInitAuthModule(t *testing.T) {
    db := testdb.NewMockDB()
    jwtManager := jwt.NewJWTManager("secret", 1*time.Hour, 24*time.Hour)
    authMiddleware := middleware.NewAuthMiddleware(jwtManager)

    router := InitAuthModule(db, jwtManager, authMiddleware)

    assert.NotNil(t, router)
    // Test that routes are registered correctly
}
```

## Benefits Summary

| Aspect              | Before                | After                   |
| ------------------- | --------------------- | ----------------------- |
| main.go size        | 300+ lines            | ~100 lines              |
| Add new module      | Touch 4+ places       | One function + one line |
| Module overview     | Scattered             | Clear in one place      |
| Dependency tracking | Manual                | Encapsulated            |
| Testing             | Complex mocking       | Test initializers       |
| Onboarding          | "Where's auth setup?" | `init_auth.go`          |

## Migration Path

1. [+] **Done**: Created `init_auth.go` for auth module
2. **Next**: When adding RBAC, create `init_rbac.go`
3. **Future**: Extract existing modules (if any) to initializers
4. **Final**: All modules follow consistent pattern

## File Checklist

- [+] `internal/adapter/http/v1/router/init_auth.go` - Auth initializer
- [+] `cmd/api/main.go` - Cleaned up to use initializer
- -> `docs/MODULE_INITIALIZATION.md` - This documentation

## Real-world Example

Imagine 15 modules in production:

```go
func main() {
    // ... infrastructure setup ...

    // Initialize all modules (15 lines instead of 400+)
    authRouter := router.InitAuthModule(db, jwtManager, authMiddleware)
    rbacRouter := router.InitRBACModule(db, authMiddleware, cache)
    profileRouter := router.InitProfileModule(db, authMiddleware, storage)
    notificationRouter := router.InitNotificationModule(db, authMiddleware, queue)
    messagingRouter := router.InitMessagingModule(db, authMiddleware, queue)
    analyticsRouter := router.InitAnalyticsModule(db, authMiddleware, clickhouse)
    billingRouter := router.InitBillingModule(db, authMiddleware, stripe)
    subscriptionRouter := router.InitSubscriptionModule(db, authMiddleware, paypal)
    webhookRouter := router.InitWebhookModule(db, authMiddleware, validator)
    searchRouter := router.InitSearchModule(db, authMiddleware, elasticsearch)
    exportRouter := router.InitExportModule(db, authMiddleware, s3)
    importRouter := router.InitImportModule(db, authMiddleware, s3)
    reportRouter := router.InitReportModule(db, authMiddleware, pdf)
    auditRouter := router.InitAuditModule(db, authMiddleware, logger)
    adminRouter := router.InitAdminModule(db, authMiddleware, rbacMiddleware)

    // Aggregate (15 parameters, but still readable)
    v1Router := router.NewV1Router(
        authRouter,
        rbacRouter,
        profileRouter,
        notificationRouter,
        messagingRouter,
        analyticsRouter,
        billingRouter,
        subscriptionRouter,
        webhookRouter,
        searchRouter,
        exportRouter,
        importRouter,
        reportRouter,
        auditRouter,
        adminRouter,
    )

    // ... server start ...
}
```

**Still maintainable!** *
