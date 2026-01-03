# JWT Package

**JSON Web Token (JWT) authentication** for Promenade Platform - token generation, validation, and RBAC middleware.

---

## Features

- **Token Generation**: Access + Refresh token pairs
- **Token Validation**: Parse and validate JWT tokens with custom claims
- **RBAC Support**: Role-based access control via claims
- **Gin Middleware**: Authentication and authorization middleware
- **Context Helpers**: Extract claims and user ID from Gin context
- **Configurable**: Secret key, TTL, issuer customization
- **Well-Tested**: 18 tests (11 core + 7 middleware), 87%+ coverage

---

## Quick Start

### 1. Configuration

Add JWT section to your config file (e.g., `config/app.postgres-dev.yaml`):

```yaml
jwt:
  secret: "your-secret-key-at-least-32-characters"
  access_token_duration: 15m
  refresh_token_duration: 168h  # 7 days
  issuer: "promenade-platform"
```

**⚠️ Security**: Use strong secrets in production (at least 32 characters, random)

### 2. Initialize Manager

```go
import "github.com/basilex/promenade/pkg/jwt"

// Load config
cfg, _ := config.Load()

// Create JWT manager
jwtManager := jwt.NewManager(jwt.Config{
    SecretKey:            cfg.JWT.Secret,
    AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
    RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
    Issuer:               cfg.JWT.Issuer,
})
```

### 3. Generate Tokens

```go
// Generate token pair for user
userID := uuidv7.New()
email := "user@example.com"
roles := []string{"user", "customer"}

tokenPair, err := jwtManager.GenerateTokenPair(userID, email, roles)
if err != nil {
    return err
}

// TokenPair contains:
// - AccessToken: short-lived token for API access
// - RefreshToken: long-lived token for refreshing access token
// - ExpiresAt: access token expiration time
// - TokenType: "Bearer"
```

### 4. Validate Tokens

```go
// Validate access token
claims, err := jwtManager.ValidateAccessToken(tokenString)
if err != nil {
    return err  // Invalid or expired
}

// Extract user info
userID := claims.UserID   // UUID as string
email := claims.Email     // User email
roles := claims.Roles     // User roles
```

### 5. Protect Routes with Middleware

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/basilex/promenade/pkg/jwt"
)

router := gin.New()

// Public routes
router.POST("/auth/login", loginHandler)
router.POST("/auth/register", registerHandler)

// Protected routes (require authentication)
protected := router.Group("/api")
protected.Use(jwt.AuthMiddleware(jwtManager))
{
    protected.GET("/profile", getProfileHandler)
    protected.PUT("/profile", updateProfileHandler)
}

// Admin-only routes (require "admin" role)
admin := router.Group("/api/admin")
admin.Use(jwt.AuthMiddleware(jwtManager))
admin.Use(jwt.RequireRole("admin"))
{
    admin.GET("/users", listUsersHandler)
    admin.DELETE("/users/:id", deleteUserHandler)
}
```

---

## API Reference

### Types

#### Config

```go
type Config struct {
    SecretKey            string        // Secret key for signing tokens (required)
    AccessTokenDuration  time.Duration // Access token TTL (default: 15m)
    RefreshTokenDuration time.Duration // Refresh token TTL (default: 7 days)
    Issuer               string        // Token issuer (default: "promenade-platform")
}
```

#### Claims

```go
type Claims struct {
    UserID string   // User UUID as string
    Email  string   // User email
    Roles  []string // User roles for RBAC
    jwt.RegisteredClaims
}
```

#### TokenPair

```go
type TokenPair struct {
    AccessToken  string    // JWT access token
    RefreshToken string    // JWT refresh token
    ExpiresAt    time.Time // Access token expiration
    TokenType    string    // Always "Bearer"
}
```

### Manager Methods

#### NewManager

```go
func NewManager(config Config) *Manager
```

Creates a new JWT manager with configuration. Applies defaults for missing values.

#### GenerateTokenPair

```go
func (m *Manager) GenerateTokenPair(userID uuidv7.UUID, email string, roles []string) (*TokenPair, error)
```

Generates access and refresh token pair for user.

**Parameters**:
- `userID`: User UUID v7
- `email`: User email address
- `roles`: User roles (e.g., `[]string{"user", "admin"}`)

**Returns**: TokenPair with both tokens, expiration time, and token type.

#### ValidateAccessToken

```go
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error)
```

Validates access token and extracts claims.

**Returns**: Claims if valid, error if expired or invalid.

#### ValidateRefreshToken

```go
func (m *Manager) ValidateRefreshToken(tokenString string) (*Claims, error)
```

Validates refresh token and extracts claims.

**Returns**: Claims if valid, error if expired or invalid.

#### RefreshAccessToken

```go
func (m *Manager) RefreshAccessToken(refreshToken string) (*TokenPair, error)
```

Generates new token pair using valid refresh token.

**Returns**: New TokenPair with refreshed tokens.

### Claims Methods

#### ExtractUserID

```go
func (c *Claims) ExtractUserID() (uuidv7.UUID, error)
```

Parses user ID from string to UUID v7.

#### HasRole

```go
func (c *Claims) HasRole(role string) bool
```

Checks if user has specific role.

#### HasAnyRole

```go
func (c *Claims) HasAnyRole(roles ...string) bool
```

Checks if user has any of specified roles (OR logic).

#### HasAllRoles

```go
func (c *Claims) HasAllRoles(roles ...string) bool
```

Checks if user has all specified roles (AND logic).

---

## Middleware

### AuthMiddleware

```go
func AuthMiddleware(manager *Manager) gin.HandlerFunc
```

Validates JWT from `Authorization: Bearer <token>` header. Stores claims in Gin context.

**Usage**:

```go
router.Use(jwt.AuthMiddleware(jwtManager))
```

**Responses**:
- 401 Unauthorized: Missing, invalid format, or invalid/expired token
- Stores claims in context with key `jwt_claims`
- Stores user ID in context with key `user_id`

### RequireRole

```go
func RequireRole(role string) gin.HandlerFunc
```

Requires user to have specific role. Use after `AuthMiddleware`.

**Usage**:

```go
router.Use(jwt.AuthMiddleware(jwtManager))
router.Use(jwt.RequireRole("admin"))
```

**Responses**:
- 403 Forbidden: User lacks required role

### RequireAnyRole

```go
func RequireAnyRole(roles ...string) gin.HandlerFunc
```

Requires user to have at least one of specified roles (OR logic).

**Usage**:

```go
router.Use(jwt.AuthMiddleware(jwtManager))
router.Use(jwt.RequireAnyRole("admin", "moderator"))
```

### RequireAllRoles

```go
func RequireAllRoles(roles ...string) gin.HandlerFunc
```

Requires user to have all specified roles (AND logic).

**Usage**:

```go
router.Use(jwt.AuthMiddleware(jwtManager))
router.Use(jwt.RequireAllRoles("admin", "superadmin"))
```

---

## Context Helpers

### GetClaims

```go
func GetClaims(c *gin.Context) *Claims
```

Extracts JWT claims from Gin context. Returns `nil` if not found.

**Usage**:

```go
func handler(c *gin.Context) {
    claims := jwt.GetClaims(c)
    if claims == nil {
        // Not authenticated
        return
    }
    // Use claims.UserID, claims.Email, claims.Roles
}
```

### GetUserID

```go
func GetUserID(c *gin.Context) string
```

Extracts user ID string from Gin context. Returns `""` if not found.

---

## Examples

### Login Endpoint

```go
func (h *UserHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }

    // Validate credentials
    user, err := h.usecase.Authenticate(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        response.Error(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
        return
    }

    // Generate JWT tokens
    roles := []string{"user"}  // Get from user entity
    tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID, user.Email, roles)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", err.Error())
        return
    }

    response.Success(c, gin.H{
        "user": user,
        "tokens": tokenPair,
    })
}
```

### Refresh Token Endpoint

```go
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req struct {
        RefreshToken string `json:"refresh_token" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }

    // Generate new token pair
    tokenPair, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
    if err != nil {
        response.Error(c, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error())
        return
    }

    response.Success(c, tokenPair)
}
```

### Protected Endpoint with Role Check

```go
func (h *UserHandler) SuspendUser(c *gin.Context) {
    // Extract current user from JWT
    claims := jwt.GetClaims(c)
    if claims == nil {
        response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "No authentication claims")
        return
    }
    currentUserID, _ := claims.ExtractUserID()

    // Check admin role (alternative to RequireRole middleware)
    if !claims.HasRole("admin") {
        response.Error(c, http.StatusForbidden, "FORBIDDEN", "Admin role required")
        return
    }

    // Get target user ID from URL
    targetUserID, _ := uuidv7.Parse(c.Param("id"))

    // Suspend user
    err := h.usecase.SuspendUser(c.Request.Context(), currentUserID, targetUserID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "SUSPEND_FAILED", err.Error())
        return
    }

    response.Success(c, gin.H{"message": "User suspended"})
}
```

### Role-Based Route Groups

```go
func RegisterRoutes(router *gin.Engine, jwtManager *jwt.Manager, handlers *Handlers) {
    api := router.Group("/api/v1")

    // Public routes
    auth := api.Group("/auth")
    {
        auth.POST("/register", handlers.Register)
        auth.POST("/login", handlers.Login)
        auth.POST("/refresh", handlers.RefreshToken)
    }

    // Authenticated routes
    user := api.Group("/users")
    user.Use(jwt.AuthMiddleware(jwtManager))
    {
        user.GET("/me", handlers.GetMyProfile)
        user.PUT("/me", handlers.UpdateMyProfile)
    }

    // Moderator routes (admin OR moderator)
    moderation := api.Group("/moderation")
    moderation.Use(jwt.AuthMiddleware(jwtManager))
    moderation.Use(jwt.RequireAnyRole("admin", "moderator"))
    {
        moderation.GET("/reports", handlers.ListReports)
        moderation.POST("/reports/:id/resolve", handlers.ResolveReport)
    }

    // Admin-only routes
    admin := api.Group("/admin")
    admin.Use(jwt.AuthMiddleware(jwtManager))
    admin.Use(jwt.RequireRole("admin"))
    {
        admin.GET("/users", handlers.ListAllUsers)
        admin.POST("/users/:id/suspend", handlers.SuspendUser)
        admin.POST("/users/:id/ban", handlers.BanUser)
        admin.DELETE("/users/:id", handlers.DeleteUser)
    }

    // Super admin routes (both roles required)
    superadmin := api.Group("/superadmin")
    superadmin.Use(jwt.AuthMiddleware(jwtManager))
    superadmin.Use(jwt.RequireAllRoles("admin", "superadmin"))
    {
        superadmin.GET("/system/config", handlers.GetSystemConfig)
        superadmin.PUT("/system/config", handlers.UpdateSystemConfig)
    }
}
```

---

## Security Best Practices

### Secret Key Management

**❌ DON'T**:
```yaml
# config/app.postgres-prod.yaml
jwt:
  secret: "weak-secret"  # Too short, predictable
```

**✅ DO**:
```yaml
# config/app.postgres-prod.yaml
jwt:
  secret: "${JWT_SECRET}"  # Load from environment variable
```

```bash
# Use strong random secret (at least 32 characters)
export JWT_SECRET=$(openssl rand -base64 32)
```

### Token Expiration

**Short-lived access tokens** (15 minutes):
- Reduces risk of token theft
- Requires frequent refresh

**Long-lived refresh tokens** (7 days):
- Better UX (less frequent re-login)
- Store securely (HTTP-only cookies)

### Token Storage

**Client-side**:
- Access token: Memory or sessionStorage (not localStorage)
- Refresh token: HTTP-only secure cookie

**Server-side**:
- No token storage needed (stateless)
- Optional: Store refresh tokens in Redis for revocation

### Token Revocation

**Basic** (current):
- Wait for token expiration (max 15 minutes)

**Advanced** (recommended for production):
- Store refresh tokens in Redis with TTL
- Delete from Redis on logout
- Check Redis on token refresh

---

## Testing

### Run Tests

```bash
# All JWT tests
go test ./pkg/jwt -v

# With coverage
go test ./pkg/jwt -cover

# Generate coverage report
go test ./pkg/jwt -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Statistics

| Test File            | Tests | Coverage |
|---------------------|-------|----------|
| jwt_test.go         | 11    | 85%      |
| middleware_test.go  | 7     | 90%      |
| **Total**           | **18**| **87%**  |

**Test Coverage**:
- Token generation: ✅
- Token validation: ✅
- Token refresh: ✅
- Claims extraction: ✅
- Role checking: ✅
- Middleware authentication: ✅
- Middleware authorization: ✅
- Context helpers (GetUserID, GetClaims): ✅
- Panic recovery: ✅

---

## Architecture

### Token Flow

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ 1. POST /auth/login (email, password)
       ▼
┌─────────────────────────────────────┐
│  Server: Login Handler              │
│  - Validate credentials             │
│  - Generate token pair              │
└──────┬──────────────────────────────┘
       │ 2. Return {access_token, refresh_token}
       ▼
┌─────────────┐
│   Client    │ Store tokens
│  (Memory)   │
└──────┬──────┘
       │ 3. GET /api/users/me
       │    Authorization: Bearer <access_token>
       ▼
┌─────────────────────────────────────┐
│  Server: AuthMiddleware             │
│  - Extract token from header        │
│  - Validate token                   │
│  - Store claims in context          │
└──────┬──────────────────────────────┘
       │ 4. Handler accesses claims
       ▼
┌─────────────────────────────────────┐
│  Protected Handler                  │
│  claims := jwt.GetClaims(c)         │
│  userID := claims.UserID            │
└─────────────────────────────────────┘
```

### Token Refresh Flow

```
┌─────────────┐
│   Client    │ Access token expired
└──────┬──────┘
       │ 1. POST /auth/refresh
       │    {refresh_token: "..."}
       ▼
┌─────────────────────────────────────┐
│  Server: Refresh Handler            │
│  - Validate refresh token           │
│  - Generate new token pair          │
└──────┬──────────────────────────────┘
       │ 2. Return new {access_token, refresh_token}
       ▼
┌─────────────┐
│   Client    │ Update stored tokens
│  (Memory)   │
└─────────────┘
```

---

## Integration

### Wire Up in Main Application

**cmd/api/main.go**:

```go
// Initialize JWT manager
jwtManager := jwt.NewManager(jwt.Config{
    SecretKey:            cfg.JWT.Secret,
    AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
    RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
    Issuer:               cfg.JWT.Issuer,
})

// Pass to context routers
identityRouter := identity.NewRouter(db, jwtManager)
identityRouter.RegisterRoutes(api)
```

**internal/contexts/identity/router.go**:

```go
type Router struct {
    userHandler *userHTTP.UserHandler
    jwtManager  *jwt.Manager
}

func NewRouter(db *sqlx.DB, jwtManager *jwt.Manager) *Router {
    userRepo := userRepo.NewUserRepository(db)
    userUseCase := user.NewUseCase(userRepo)
    userHandler := userHTTP.NewUserHandler(userUseCase, jwtManager)

    return &Router{
        userHandler: userHandler,
        jwtManager:  jwtManager,
    }
}

func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
    identity := api.Group("/identity")
    {
        // Public routes
        identity.POST("/register", r.userHandler.Register)
        identity.POST("/login", r.userHandler.Login)

        // Protected routes
        users := identity.Group("/users")
        users.Use(jwt.AuthMiddleware(r.jwtManager))
        {
            users.GET("/me", r.userHandler.GetMyProfile)
            users.PUT("/me", r.userHandler.UpdateMyProfile)
        }
    }
}
```

---

## Roadmap

### Current Features

- ✅ Token generation (access + refresh)
- ✅ Token validation
- ✅ Custom claims (user_id, email, roles)
- ✅ RBAC helpers (HasRole, HasAnyRole, HasAllRoles)
- ✅ Gin middleware (auth + RBAC)
- ✅ Context helpers

### Planned Features

- [ ] Token blacklist/revocation (Redis)
- [ ] Token refresh rotation
- [ ] Token introspection endpoint
- [ ] JWT key rotation
- [ ] Multi-tenant support
- [ ] OAuth2 integration
- [ ] Session management

---

## Related Documentation

- [Main README](../../README.md)
- [Documentation Index](../../docs/INDEX.md)
- [Identity Context](../../internal/contexts/identity/README.md)
- [Testing Guide](../../test/README.md)

---

**Last Updated**: 208 tests, 87
**Status**: Production-ready  
**Test Coverage**: 15 tests, 83% coverage  
**Maintainer**: Promenade Team
