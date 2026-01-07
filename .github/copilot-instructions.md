# Promenade AI Instructions

Essential guide for AI agents working in Promenade. For detailed documentation, see [README.md](../README.md) and [docs/](../docs/).

---

## Core Architecture Principles

### 1. Domain-Driven Design (DDD) with Bounded Contexts

**Architecture**: Event-Driven DDD with Clean Architecture layers

- **Bounded Contexts** - Autonomous business domains (Identity, Customer Management, Order Management, Billing)
- **Aggregates** - Business entities with invariants and transactional boundaries
- **Value Objects** - Immutable domain concepts (Email, Phone, Money, Address)
- **Domain Events** - Asynchronous communication via Event Bus (Memory/Redis adapters)
- **No ORM** - Raw SQL with sqlx + BaseRepository pattern
- **Multi-Database** - PostgreSQL (production), SQLite (dev/demo), MySQL (planned) via dialect abstraction
- **UUID v7 Only** - Use `pkg/uuidv7.New()` for all IDs (time-ordered, 2x faster inserts)
- **JWT Authentication** - Token-based auth with RBAC (15min access, 7 days refresh)
- **Go 1.24+** - Modern Go with range-over-func and improved type inference

### 2. Bounded Contexts Structure

**Directory**: `internal/contexts/{context-name}/`

Each context is autonomous with:

- Own domain model (aggregates, entities, value objects)
- Own database schema (migrations in `migrations/{context-name}/`)
- Own HTTP routes (registered via Router)
- No cross-context imports (communicate via Event Bus)

**Available Contexts**:

- **Shared** (`internal/contexts/shared/`) - Reference data: Country, Currency, Language, Timezone (read-only) - Production
- **Identity** (`internal/contexts/identity/`) - User, Contact, Profile, Role, Permission aggregates - Production
  - User: Registration, authentication, password management
  - Contact: Email, phone, address management
  - Profile: Personal info, bio, avatar, localization
  - Role & Permission: RBAC implementation - Production
- **Customer Management** (`internal/contexts/customer-mgmt/`) - Customer | Company | Deal | Interaction | Analytics (all Production)
- **Order Management** (`internal/contexts/order-mgmt/`) - Order aggregate (Production) | OrderLine entity | Contract, Fulfillment planned
- **Billing** (`internal/contexts/billing/`) - Invoice, Payment, Subscription (all Production)
- **Warehouse** (`internal/contexts/warehouse/`) - Inventory, StockMovement, Product, Location (ALL Production, 100% complete)

**Context isolation**: Contexts communicate ONLY via Event Bus (no direct dependencies)

**Latest Progress** (January 7, 2026):
- ✅ Phase 1 COMPLETE: API Documentation & Developer Portal (5 days, 2x faster than planned)
- ✅ Phase 2 COMPLETE: Warehouse Context (100% complete - ALL 4 aggregates PRODUCTION)
- 🚧 Phase 3 IN PROGRESS: LUA Scripting + UI Metadata Foundation (Week 1 started)
- Business Documentation COMPLETE: Multi-language business overviews for executives
  - docs/business/BUSINESS_OVERVIEW.md (English, 500+ lines, 15 sections)
  - docs/business/BUSINESS_OVERVIEW_UK.md (Ukrainian, complete translation)
  - docs/business/BUSINESS_OVERVIEW_DE.md (German, complete translation)
  - docs/business/BUSINESS_OVERVIEW_FR/ES/PT/JP/ZH.md (Additional translations)
  - Positioned first in docs/INDEX.md and prominently in README.md
- Warehouse Context - ALL 4 Aggregates PRODUCTION:
  - Inventory: 141 tests (97 unit + 23 integration + 21 smoke), 14 API endpoints
  - StockMovement: 45 tests (11 entity + 10 usecase + 9 smoke + 15 integration), audit trail
  - Product: 139 tests (25 entity + 83 usecase + 10 smoke + 21 integration), 16 API endpoints
  - Location: 74 tests (48 entity + 17 integration + 9 smoke), 14 API endpoints
- Warehouse Integration COMPLETE: Automated order-inventory synchronization via Event Bus
  - ReservationService: Business logic for stock operations (230 lines)
  - OrderEventHandler: Event handlers for order lifecycle (230 lines, 3 handlers)
  - Event Flow: order.confirmed → Reserve Stock | order.cancelled → Release Stock | order.fulfilled → Commit Stock
  - Bootstrap Integration: Initialized in cmd/api/bootstrap.go with automatic event handler registration
  - 34 Integration Tests: E2E testing with Order Management context
- Phase 3 LUA Scripting Engine: OPERATIONAL (Week 1 Day 1 Complete - After File Recovery)
  - pkg/scripting/engine.go: LUA VM wrapper with context support (210 lines, recreated)
  - pkg/scripting/sandbox.go: Security restrictions (90 lines, memory 50MB, timeout 5s, no filesystem/network)
  - pkg/scripting/stdlib.go: Standard library stubs (185 lines, Customer, Order, Deal, Notify, Query, Date APIs)
  - pkg/scripting/engine_test.go: 18 tests + 3 benchmarks (185 lines, all passing)
  - pkg/scripting/README.md: Complete documentation (600+ lines)
  - Dependencies: gopher-lua (LUA interpreter), gopher-luar (Go-LUA bridge) - installed
  - File Corruption Incident: All 4 files corrupted by formatter, successfully recovered in ~30 minutes
  - Status: All tests passing, build successful, ready for Standard Library implementation
- Test Infrastructure: All systems validated (2453+ tests: 2220+ unit, 170+ smoke, 76+ integration)

### 3. Aggregate Structure Pattern

**Directory structure** per aggregate in a context:

```
internal/contexts/{context}/{aggregate}/
 entity.go                     # Domain entity (aggregate root)
 entity_test.go                # Entity tests
 repository.go                 # Repository interface
 usecase.go                    # Business logic (use cases)
 usecase_test.go               # Use case tests
 adapter/
     http/handler/
        {aggregate}_handler.go    # HTTP handlers
        dto/
            {aggregate}_dto.go    # Data transfer objects
     repository/postgres/
         base_repository.go        # BaseRepository (per context)
         {aggregate}_repository.go # PostgreSQL implementation
```

**Example - Identity Contact Aggregate**:

```
internal/contexts/identity/contact/
 entity.go                     # Contact aggregate (Email, Phone, Address)
 entity_test.go
 repository.go                 # IRepository interface
 usecase.go                    # IUseCase interface + implementation
 usecase_test.go
 adapter/
     http/handler/
        contact_handler.go
        dto/
            contact_dto.go
     repository/postgres/
         base_repository.go
         contact_repository.go
```

### 4. Key Patterns

**BaseAggregate Pattern** (CRITICAL - Updated January 2026):

```go
// ALL entities MUST embed BaseAggregate and NEVER duplicate its fields
type Contact struct {
    aggregate.BaseAggregate  // Provides: ID, Version, CreatedAt, UpdatedAt, DeletedAt
    UserID uuidv7.UUID
    Type   ContactType
    // ... other fields
}

//  NEVER DO THIS - Field duplication
type WrongContact struct {
    aggregate.BaseAggregate
    ID        uuidv7.UUID  //  DUPLICATE - already in BaseAggregate
    CreatedAt time.Time    //  DUPLICATE - already in BaseAggregate
    UpdatedAt time.Time    //  DUPLICATE - already in BaseAggregate
}

// Factory method - BaseAggregate auto-initializes ID, CreatedAt, UpdatedAt
func NewContact(userID uuidv7.UUID) *Contact {
    return &Contact{
        BaseAggregate: aggregate.NewBaseAggregate(),  // Sets ID, timestamps
        UserID:        userID,
    }
}

//  Use Touch() for timestamp updates
func (c *Contact) Verify() {
    c.IsVerified = true
    c.Touch()  // Updates UpdatedAt via BaseAggregate
}

//  Use GetID() for ID access
func (r *contactRepository) Update(ctx context.Context, contact *Contact) error {
    query := `UPDATE contacts SET ... WHERE id = $1`
    return r.Exec(ctx, query, contact.GetID())  // Not contact.ID
}
```

**Repository with BaseRepository**:

```go
// Repository interface in aggregate package (usually IRepository)
type IRepository interface {
    Create(ctx context.Context, contact *Contact) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error)
    Update(ctx context.Context, contact *Contact) error
    Delete(ctx context.Context, id uuidv7.UUID) error
}

// Implementation embeds BaseRepository (lowercase private struct)
type contactRepository struct {
    *BaseRepository
}

func NewContactRepository(db *sqlx.DB) IRepository {
    return &contactRepository{
        BaseRepository: NewBaseRepository(db),
    }
}
```

**EXCEPTION**: Customer Management uses `ICustomerRepository` (entity-specific name)

**Use Case Pattern**:

```go
// Interface defines business operations (usually IUseCase)
type IUseCase interface {
    CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*Contact, error)
    GetContact(ctx context.Context, contactID uuidv7.UUID) (*Contact, error)
    VerifyContact(ctx context.Context, contactID uuidv7.UUID) error
}

// Implementation (lowercase struct, always "useCase")
type useCase struct {
    repo IRepository
}

// Constructor: simple NewUseCase (NOT New{Entity}UseCase)
func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}
```

**EXCEPTION**: Customer Management uses `ICustomerUseCase` (entity-specific name)

**Domain Events** (via Event Bus):

```go
// Publish event after aggregate state change
event := bus.NewBaseEvent("contact.verified", contact.ID)
if err := eventBus.Publish(ctx, bus.TopicContactVerified, event); err != nil {
    // Log but don't fail - events are fire-and-forget
    logger.FromContext(ctx).Error("Failed to publish event", slog.Any("error", err))
}
```

**Context Propagation**: Always pass `ctx` - carries transaction, logger, request ID, user info

**Bootstrap Pattern** (see `cmd/api/bootstrap.go`):
1. Logger initialization → Database connection → Migrations
2. Redis → Cache → JWT + Token Revoker
3. Event Bus → Health Checker
4. Context routers → Server start

### Multi-Database Support

**Database-agnostic architecture** via dialect abstraction (`pkg/database/dialect.go`):

- **PostgreSQL**: Production (default), native JSONB, UUID, full features
- **SQLite**: Development/demos, embedded, zero config, TEXT for JSONB/UUID
- **MySQL**: Planned, JSON type, CHAR(36) for UUID

**Switching databases**:

```bash
# PostgreSQL (production)
DATABASE_DRIVER=postgres ENVIRONMENT=production ./bin/promenade

# SQLite (development, no Docker needed)
DATABASE_DRIVER=sqlite ENVIRONMENT=development ./bin/promenade
# Or: make dev-sqlite

# Configuration files (driver-environment format)
config/app.postgres-dev.yaml   # PostgreSQL + development
config/app.sqlite-dev.yaml     # SQLite + development
config/app.sqlite-test.yaml    # SQLite + testing (in-memory)
```

**Key patterns**:
- Write queries with `?` placeholders, auto-convert to `$1` (Postgres) or `?` (SQLite)
- Store JSON as TEXT (cross-database), type-safe via `jsonstore.Field[T]`
- UUID generation in Go code (`uuidv7.New()`), not database defaults
- Timestamp management via `.Touch()` method (no database triggers)

## Essential Workflows

### Workspace Management (CRITICAL - Read First!)

**Core Concept**: `.promenade.workspace` file defines `DATABASE_DRIVER` and `ENVIRONMENT` for all commands.

**Configure once, work anywhere**:

```bash
# Switch to PostgreSQL + development
make switch-postgres-dev    # Creates .promenade.workspace with postgres+dev
make dev                    # Runs development server (validates workspace)

# Switch to SQLite + testing
make switch-sqlite-test     # Updates .promenade.workspace to sqlite+test
make test-all               # Runs all tests (validates workspace)

# Check current configuration
make workspace              # Shows DATABASE_DRIVER and ENVIRONMENT
```

**9 Switchers** (driver-environment combos):
- `make switch-postgres-dev` → PostgreSQL + development (default)
- `make switch-postgres-test` → PostgreSQL + test
- `make switch-postgres-prod` → PostgreSQL + production
- `make switch-sqlite-dev` → SQLite + development (no Docker needed)
- `make switch-sqlite-test` → SQLite + test
- `make switch-sqlite-prod` → SQLite + production
- `make switch-mysql-dev` → MySQL + development (planned)
- `make switch-mysql-test` → MySQL + test (planned)
- `make switch-mysql-prod` → MySQL + production (planned)

**Environment-Aware Runners** (validate workspace before execution):
- `make dev` - Development server (requires ENVIRONMENT=development)
- `make test-all` - All tests (warns if not ENVIRONMENT=test)
- `make prod` - Production runner (requires ENVIRONMENT=production)

**Workspace Commands** (database/environment agnostic):
- `make workspace` - Show current DATABASE_DRIVER + ENVIRONMENT
- `make build` - Build binary (no validation needed)
- `make fmt` - Format code (no validation needed)
- `make lint` - Run linters (no validation needed)

**See**: [Workspace Management Guide](docs/guides/workspace-management.md) for complete architecture

### Make Commands (Modular System)

**Makefile Structure**:
- `Makefile` - Main (workspace switchers, validation, help)
- `Makefile.dev.mk` - Development workflow (dev, docker, migrations)
- `Makefile.test.mk` - Testing infrastructure (test-all, benchmarks)
- `Makefile.prod.mk` - Production deployment (prod, swagger)

**Development** (from Makefile.dev.mk):

- `make dev` - Run development server (validates workspace, requires ENVIRONMENT=development)
- `make dev-fresh` - Fresh start with clean database
- `make build` - Build binary
- `make lint` / `make fmt` - Code quality checks

**Testing** (from Makefile.test.mk, four-tier strategy):

- `make test-all` - All tests runner (warns if not ENVIRONMENT=test)
- `make test` - All tests with race detector (~60s, 2341+ tests)
- `make test-unit` - Unit tests only (fast, ~5s, no workspace needed)
- `make test-smoke` - Smoke tests for handlers (HTTP validation, no DB, ~0.6s, 182 tests)
- `make test-integration` - Integration tests with real DB (~14s, validates workspace)
- `make test-benchmark` - Benchmark tests (performance measurement, validates workspace)
- `make test-coverage` - HTML coverage report
- Test DB: Auto-starts on port 5433 with `promenade_test` database

**Smoke Testing**: 182 tests across 18 handlers, 100% pass rate - see [test/smoke/README.md](test/smoke/README.md)

**Smoke Testing Pattern**:

**Purpose**: Validate HTTP handlers with minimal effort (80/20 rule)  
**Location**: `test/smoke/contexts/{context}/{aggregate}/handler_test.go`  
**Coverage**: 6-12 tests per handler (6 minimum, 9 standard, 12 for complex)  
**Run**: `make test-smoke` (~0.6s, no database needed)

**Key Characteristics**:
- Mock UseCase with function fields (no external dependencies)
- Test HTTP status codes (200, 201, 404, 400, 500)
- Validate response format (`{"status":"success","data":{...}}`)
- No database, no complex business logic
- Fast execution (< 100ms per test)

**Example**:
```go
// test/smoke/contexts/order-mgmt/order/handler_test.go
package order_test

import (
    "testing"
    "github.com/basilex/promenade/test/smoke"
)

// MockOrderUseCase with function fields
type MockOrderUseCase struct {
    GetByIDFunc func(ctx context.Context, id uuid.UUID) (*order.Order, error)
}

func TestOrderHandler_GetByID_Success(t *testing.T) {
    router := smoke.SetupRouter()
    
    // Mock UseCase returns fake order
    mockUC := &MockOrderUseCase{
        GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*order.Order, error) {
            return fakeOrder(), nil
        },
    }
    
    handler := orderHTTP.NewOrderHandler(mockUC)
    router.GET("/orders/:id", handler.GetByID)
    
    // Test request
    resp := smoke.MakeRequest(t, router, "GET", "/orders/"+smoke.FakeUUID(), nil)
    smoke.AssertSuccessResponse(t, resp, 200)
}

func TestOrderHandler_GetByID_NotFound(t *testing.T) {
    router := smoke.SetupRouter()
    
    mockUC := &MockOrderUseCase{
        GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*order.Order, error) {
            return nil, order.ErrOrderNotFound
        },
    }
    
    handler := orderHTTP.NewOrderHandler(mockUC)
    router.GET("/orders/:id", handler.GetByID)
    
    resp := smoke.MakeRequest(t, router, "GET", "/orders/"+smoke.FakeUUID(), nil)
    smoke.AssertErrorResponse(t, resp, 404, "ORDER_NOT_FOUND")
}
```

**Helper Utilities** (`test/smoke/testutils.go`):
- `SetupRouter()` - Gin test router
- `MakeRequest(t, router, method, path, body)` - HTTP request helper
- `AssertSuccessResponse(t, resp, expectedCode)` - Validate success response
- `AssertErrorResponse(t, resp, expectedCode, expectedErrorCode)` - Validate error response
- `FakeUUID()` - Generate test UUID v7

**When to Write Smoke Tests**:
-  For all HTTP handlers (6-12 tests per handler)
-  When adding new endpoints
-  Before integration tests (faster feedback)
-  Not for complex business logic (use integration tests)
-  Not for repository methods (use integration tests with real DB)

**See**: [test/smoke/README.md](../test/smoke/README.md) for complete smoke testing guide

**Migrations** (database-agnostic via workspace):

- `make migrate` - Run all migrations (auto-detects DATABASE_DRIVER from workspace)
- `make migrate-core` - Core migrations only
- `make migrate-shared` - Shared context migrations
- `make migrate-identity` - Identity context migrations
- `make migrate-customer-mgmt` - Customer management migrations
- `make migrate-order-mgmt` - Order management migrations
- `make migrate-billing` - Billing migrations
- `make migrate-status` - Show migration status
- `make migrate-new CONTEXT=identity NAME=xxx` - Create new migration
- Migration order: core → shared → identity → customer-mgmt → order-mgmt → billing

**Docker** (database-aware via workspace):

- `make docker-up` - Start database containers (auto-detects DATABASE_DRIVER, skips for SQLite)
- `make docker-down` - Stop containers
- `make docker-ps` - Show running containers
- `make docker-clean` - Remove all containers and volumes

**Local CI Validation** (run before every push):

- `make pre-push` - Run all CI checks locally (lint + test + build)
- `make ci-lint` - Run golangci-lint (same as GitHub Actions)
- `make ci-test` - Run all tests with race detector
- `make ci-build` - Test compilation
- Saves 4+ minutes per failed push by catching issues early

## API & HTTP Patterns

**Context Routers**: Each context has its own router that registers routes

```go
// Identity context router (multi-aggregate)
type Router struct {
    contactHandler *contactHTTP.ContactHandler
    profileHandler *profileHTTP.ProfileHandler
    userHandler    *userHTTP.UserHandler
}

func NewRouter(db *sqlx.DB) *Router {
    // Initialize Contact aggregate
    contactRepository := contactRepo.NewContactRepository(db)
    contactUseCase := contact.NewUseCase(contactRepository)  // Simple NewUseCase
    contactHandler := contactHTTP.NewContactHandler(contactUseCase)

    // Initialize Profile aggregate
    profileRepository := profileRepo.NewProfileRepository(db)
    profileUseCase := profile.NewUseCase(profileRepository)
    profileHandler := profileHTTP.NewProfileHandler(profileUseCase)

    return &Router{
        contactHandler: contactHandler,
        profileHandler: profileHandler,
    }
}

func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
    identity := api.Group("/identity")
    {
        contacts := identity.Group("/contacts")
        {
            contacts.POST("", r.contactHandler.Create)
            contacts.GET("/:id", r.contactHandler.GetByID)
        }
        profiles := identity.Group("/profiles")
        {
            profiles.POST("", r.profileHandler.Create)
            profiles.GET("/:id", r.profileHandler.GetByID)
        }
    }
}
```

**Response format** (`pkg/response`):

```go
//Success
{"status":"success","data":{...}}
//Error
{"status":"error","error":{"code":"VALIDATION_ERROR","message":"..."}}
//Paginated
{"status":"success","data":[...],"pagination":{"total":100,"page":1,"page_size":20}}
```

**Handler pattern**:

```go
func (h *ContactHandler) Create(c *gin.Context) {
    var req CreateContactDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    contact, err := h.usecase.CreateEmailContact(c.Request.Context(), req.UserID, req.Email, req.Label, req.IsPrimary)
    if errors.Is(err, ErrContactNotFound) {
        response.Error(c, http.StatusNotFound, "CONTACT_NOT_FOUND", err.Error())
        return
    }
    response.Success(c, contact)
}
```

## JWT Authentication & Authorization

**JWT Manager**: Initialized in `cmd/api/main.go` and passed to Identity context router

```go
// Initialize JWT Manager in main.go
jwtManager := jwt.NewManager(jwt.Config{
    SecretKey:            cfg.JWT.Secret,
    AccessTokenDuration:  cfg.JWT.AccessTokenDuration,  // 15 minutes
    RefreshTokenDuration: cfg.JWT.RefreshTokenDuration, // 7 days
    Issuer:               cfg.JWT.Issuer,
})

// Pass to Identity router
identityRouter := identity.NewRouter(db, jwtManager, tokenRevoker)
```

**Rate Limiting**: IP-based protection for authentication endpoints

```go
// Rate limiters in router.go
loginLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/5), 1)     // 5 per minute
registerLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/3), 1)  // 3 per minute

// Apply to public auth routes
users.POST("/register", registerLimiter.Limit(), handler.Register)
users.POST("/login", loginLimiter.Limit(), handler.Login)
```

**Token Revocation**: Redis-based token blacklist for logout

```go
// Initialize token revoker (cmd/api/main.go)
tokenRevoker := jwt.NewTokenRevoker(redisClient)

// Apply middleware with revocation check
protected.Use(jwt.AuthMiddleware(jwtManager, tokenRevoker))

// Revoke token on logout
if err := h.tokenRevoker.RevokeToken(ctx, tokenString, ttl); err != nil {
    return err
}
```

**Protecting Routes** (see `internal/contexts/identity/router.go`):

```go
// Protected routes require JWT middleware
contacts := identity.Group("/contacts")
contacts.Use(jwt.AuthMiddleware(r.jwtManager))  // Apply middleware
{
    contacts.POST("", r.contactHandler.Create)
    contacts.GET("", r.contactHandler.List)
}

// Public routes (no middleware)
public := identity.Group("/public")
{
    public.GET("/profiles", r.profileHandler.ListPublic)
}
```

**Login Flow** (User Handler):

```go
// 1. Authenticate user
user, err := h.useCase.Authenticate(ctx, req.Email, req.Password)

// 2. Generate JWT tokens
accessToken, err := h.jwtManager.GenerateToken(user.ID.String(), user.Roles())
refreshToken, err := h.jwtManager.GenerateRefreshToken(user.ID.String())

// 3. Return tokens in response
response.Success(c, LoginResponse{
    AccessToken:  accessToken,
    RefreshToken: refreshToken,
    User:         toUserDTO(user),
})
```

**Token Refresh** (see `pkg/jwt/README.md`):

```go
// Client sends refresh token
refreshToken := req.RefreshToken

// Validate and generate new access token
claims, err := h.jwtManager.ValidateToken(refreshToken)
newAccessToken, err := h.jwtManager.GenerateToken(claims.UserID, claims.Roles)
```

**Role-Based Access Control (RBAC)**:

```go
// Roles stored in JWT claims
claims.Roles = []string{"admin", "manager", "user"}

// Require specific roles (middleware)
admin := api.Group("/admin")
admin.Use(jwt.RequireRoles(jwtManager, "admin"))  // Only admins
{
    admin.POST("/users", handler.CreateUser)
}

// Check roles in handler
userID := jwt.GetUserID(c)  // Extract from context
roles := jwt.GetRoles(c)    // Get user roles
if !contains(roles, "admin") {
    response.Error(c, http.StatusForbidden, "FORBIDDEN", "Admin access required")
    return
}
```

**Critical JWT Rules**:

- **Never hardcode secrets**: Use env vars in production (`JWT_SECRET`)
- **15 min access tokens**: Short TTL for security, use refresh tokens for long sessions
- **Load roles on login**: User aggregate must fetch roles from DB (`user.Roles()`)
- **Validate tokens**: Always use `jwtManager.ValidateToken()`, never decode manually
- **Context propagation**: Extract `userID` from JWT claims via `jwt.GetUserID(c)`
- **Token revocation**: Check blacklist on protected routes (graceful degradation if Redis unavailable)

## Data Patterns

**Invalidation**: Pattern-based deletion on updates

```go
// Invalidate all country-related cache entries
r.cache.DeletePattern(ctx, "countries:*")
```

**Repository pattern** (embed BaseRepository):

```go
type contactRepository struct {
    *BaseRepository  //Provides Get, Select, Exec, NamedExec, getExecutor
}

func (r *contactRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error) {
    var row contactRow
    query := `SELECT * FROM identity_contacts WHERE id = $1 AND deleted_at IS NULL`
    if err := r.Get(ctx, &row, query, id); err != nil {
        return nil, err
    }
    return row.toEntity()
}
```

**Transactions** (context-aware via TransactionManager):

```go
err := tm.WithTransaction(ctx, func(ctx context.Context) error {
    //All repo calls use same tx via getExecutor(ctx)
    if err := contactRepo.Create(ctx, contact); err != nil {
        return err  //Auto-rollback
    }
    return userRepo.Update(ctx, user)  //Auto-commit if no error
})
```

**Critical Rules**:

- UUID v7: `pkg/uuidv7.New()` ONLY (never `uuid.New()`)
- Soft delete: Always add `WHERE deleted_at IS NULL`
- Context: Always pass `ctx` for tx/logger propagation
- Logger: `logger.FromContext(ctx)` not global logger

**JSON Storage** (cross-database pattern via `pkg/jsonstore`):

```go
// Store arrays/objects as JSON TEXT (works on Postgres JSONB and SQLite TEXT)
type Customer struct {
    Tags jsonstore.Field[[]string]  // Type-safe JSON field
}

// In entity
customer.Tags.Set([]string{"vip", "enterprise"})
tags := customer.Tags.Get()  // []string

// In repository (database-agnostic)
query := `INSERT INTO customers (id, tags) VALUES ($1, $2)`
tagsJSON := customer.Tags.MarshalJSON()  // Handles nil safely
```

**Key benefits**:
- Type-safe: Generic `Field[T]` ensures compile-time safety
- Cross-database: Works on Postgres JSONB and SQLite TEXT
- Nil-safe: Empty arrays serialize as `[]` not `null`
- Zero config: No database-specific code needed

## Customer Management & Deal Pipeline

**Customer Context** (`internal/contexts/customer-mgmt/customer/`):

- **Customer Lifecycle**: Lead → Prospect → Customer → Churned (state machine)
- **Customer Tiers**: free, basic, pro, enterprise
- **Segmentation**: JSONB tags for flexible metadata
- **Sales Rep Assignment**: Track ownership and source
- **14 API Endpoints**: Complete CRUD + business operations

**Company Context** (`internal/contexts/customer-mgmt/company/`):

- **B2B Support**: Legal entities for business customers
- **Tax & Legal**: Tax ID, legal name, registration number
- **Parent-Subsidiary**: Support for corporate hierarchies
- **Industry & Size**: Classification and employee count
- **14 API Endpoints**: Complete CRUD + hierarchical operations

**Deal Context** (`internal/contexts/customer-mgmt/deal/`):

- **Deal Stages**: lead → qualified → proposal → negotiation → closed_won/closed_lost
- **Probability Tracking**: Auto-calculated per stage (10% → 100%)
- **Money Value Object**: Type-safe handling with currency support
- **Pipeline Statistics**: Filter by stage, customer, or sales rep
- **12 API Endpoints**: Complete CRUD + stage transitions

**Interaction Context** (`internal/contexts/customer-mgmt/interaction/`):

- **Interaction Types**: Call, Email, Meeting, Note, SMS, Chat
- **Directions**: Inbound (customer-initiated), Outbound (company-initiated)
- **Outcomes**: Successful, Failed, No Answer, Voicemail, Busy, Scheduled, Not Interested
- **JSONB Attendees**: Multi-participant tracking with flexible arrays
- **Follow-up Management**: Flag and schedule follow-up actions
- **Duration Tracking**: Automatic calculation for ended interactions
- **14 API Endpoints**: Complete CRUD + business operations

**Analytics Context** (`internal/contexts/customer-mgmt/analytics/`):

- **CQRS Pattern**: Separate read models for reporting (9 query methods)
- **Direct SQL**: No repository abstraction for optimized analytics
- **8 GET Endpoints**: Customer overview, lifecycle, segmentation, deal pipeline, conversions, sales rep performance, revenue time series, interaction insights
- **Denormalized Queries**: LEFT JOIN across aggregates for performance
- **Real-time Metrics**: Dashboard-ready analytics

**Business Rules**:
- Customer state transitions enforce lifecycle rules
- Deal stage transitions are validated (can't skip stages)
- Automatic probability updates on stage changes
- Win/loss tracking with actual close dates
- Interaction duration auto-calculated when ended
- Follow-up tracking with due dates and completion status

**See**: [Customer Management Guide](docs/concepts/customer-management.md) | [Company Management Guide](docs/concepts/company-management.md) | [Deal Management Guide](docs/concepts/deal-management.md) | [Interaction Management Guide](docs/concepts/interaction-management.md) | [Analytics README](internal/contexts/customer-mgmt/analytics/README.md)

## Order Management

**Order Context** (`internal/contexts/order-mgmt/order/`):

- **Order Creation**: Generate orders with auto-numbered format (ORD-YYYY-NNNNNN)
- **Line Items**: Add/remove products with automatic total calculation
- **State Machine**: pending → confirmed → processing → fulfilled (or cancelled)
- **Money Handling**: Type-safe cents-based precision
- **14 API Endpoints**: Complete CRUD + state transitions

**Order Entity Methods**:

```go
func (o *Order) AddLine(productID uuidv7.UUID, quantity int, unitPrice Money) error
func (o *Order) RemoveLine(lineID uuidv7.UUID) error
func (o *Order) UpdateLine(lineID uuidv7.UUID, quantity int) error
func (o *Order) Confirm() error            // pending → confirmed
func (o *Order) StartProcessing() error    // confirmed → processing
func (o *Order) MarkFulfilled() error      // processing → fulfilled
func (o *Order) Cancel(reason string) error
```

**Business Rules**:
- Order must have at least one line item to confirm
- Cannot modify confirmed orders (must cancel and recreate)
- Terminal states (fulfilled, cancelled) are immutable
- Total automatically recalculated on line item changes

## Warehouse Context

**Current Status** (January 7, 2026): ✅ **100% COMPLETE** - ALL 4 Aggregates PRODUCTION

**Inventory Aggregate** (`internal/contexts/warehouse/inventory/`) - ✅ **PRODUCTION**:

- **Stock Tracking**: QuantityOnHand, QuantityReserved, QuantityAvailable, QuantityCommitted
- **Reorder Management**: ReorderPoint, MinStock, MaxStock thresholds
- **Cost Tracking**: UnitCost, TotalCost with weighted average
- **14 API Endpoints**: CRUD + stock operations (receive, reserve, release, commit)
- **141 Tests Passing**: 97 unit + 23 integration + 21 smoke

**Key Operations**:
```go
func (i *Inventory) ReceiveStock(quantity int, unitCost float64) error
func (i *Inventory) ReserveStock(quantity int, orderID uuidv7.UUID) error
func (i *Inventory) ReleaseReservation(quantity int, orderID uuidv7.UUID) error
func (i *Inventory) CommitStock(quantity int, orderID uuidv7.UUID) error
```

**Business Rules**:
- QuantityAvailable = QuantityOnHand - QuantityReserved
- Cannot reserve more than available stock
- Weighted average cost calculation on receipt
- Auto-update TotalCost on stock movements

**StockMovement Aggregate** (`internal/contexts/warehouse/stockmovement/`) - ✅ **PRODUCTION**:

- **Movement Types**: Receipt, Reservation, ReservationRelease, Commit, Adjustment, Transfer, Damage, Return
- **Audit Trail**: Immutable append-only log for compliance
- **Cost Tracking**: UnitCostCents, TotalCostCents, CurrencyCode
- **Reference Tracking**: Links to Orders, POs, Adjustments
- **Location Tracking**: FromWarehouse/ToWarehouse for transfers
- **45 Tests Passing**: 11 entity + 10 usecase + 9 smoke + 15 integration

**Entity Structure**:
```go
type StockMovement struct {
    aggregate.BaseAggregate
    InventoryID        uuidv7.UUID
    Type               MovementType
    Quantity           int  // Positive or negative
    QuantityBeforeMove int  // Snapshot before
    QuantityAfterMove  int  // Snapshot after
    FromWarehouseID    *uuidv7.UUID
    ToWarehouseID      *uuidv7.UUID
    ReferenceType      string  // "order", "po", "adjustment"
    ReferenceID        *uuidv7.UUID
    Reason             string  // Required for adjustments
    CreatedBy          uuidv7.UUID
    MovementDate       time.Time
}
```

**Key Business Rules**:
- Movements are immutable (append-only)
- Adjustments require reason
- Transfers require both from and to warehouses
- Quantity cannot be zero
- GetImpact() returns +/- for stock calculations

**Implementation Details**:
- Repository: 11 methods (Create, GetByID, GetByInventoryID, GetByType, GetByReference, GetByDateRange, GetSummaryByInventory, GetRecentMovements, CountByType, List, Count)
- UseCase: 9 methods covering all business operations
- PostgreSQL placeholders ($1, $2, $3) used throughout
- FK constraints properly handled with createTestInventory helpers
- Date range queries use real dates (not time.Time{})
- All SQL queries use proper placeholder conversion (? → $N)

**Product Aggregate** (`internal/contexts/warehouse/product/`) - ✅ **PRODUCTION**:

- **Catalog Management**: SKU, name, description, category, brand
- **Stock Tracking**: Track across all warehouses
- **Multi-location**: Support for distributed inventory
- **16 API Endpoints**: Complete CRUD + business operations
- **139 Tests Passing**: 25 entity + 83 usecase + 10 smoke + 21 integration

**Location Aggregate** (`internal/contexts/warehouse/location/`) - ✅ **PRODUCTION**:

- **Warehouse Locations**: Name, code, type (warehouse/retail/dropship/virtual)
- **Address Management**: Full address with country integration
- **Capacity Tracking**: Track location capacity and utilization
- **Hierarchical Structure**: Support for zones and bins
- **14 API Endpoints**: Complete CRUD + location operations
- **74 Tests Passing**: 48 entity + 17 integration + 9 smoke

**Warehouse Integration** (✅ COMPLETE - January 7, 2026):

Automated order-inventory synchronization via Event Bus provides seamless stock management:

**Architecture**:
- **ReservationService** (`internal/contexts/warehouse/integration/reservation_service.go`): Core business logic
  - `ReserveForOrder(orderID, items, reservedBy)` - Reserve stock when order confirmed
  - `ReleaseForOrder(orderID, items, releasedBy)` - Release reservation when order cancelled
  - `CommitForOrder(orderID, items, committedBy)` - Commit stock when order fulfilled
  
- **OrderEventHandler** (`internal/contexts/warehouse/integration/order_event_handler.go`): Event listeners
  - `HandleOrderConfirmed()` - Responds to `order.confirmed` event
  - `HandleOrderCancelled()` - Responds to `order.cancelled` event
  - `HandleOrderFulfilled()` - Responds to `order.fulfilled` event

**Event Flow**:
```
Order.Confirm() → order.confirmed event → ReservationService.ReserveForOrder()
Order.Cancel()  → order.cancelled event → ReservationService.ReleaseForOrder()
Order.Fulfill() → order.fulfilled event → ReservationService.CommitForOrder()
```

**Bootstrap Integration** (`cmd/api/bootstrap.go`):
```go
// Initialize Warehouse Integration (ReservationService + OrderEventHandler)
func initWarehouseIntegration(db *sqlx.DB, eventBus bus.IBus) (*integration.OrderEventHandler, error) {
    invRepo := inventoryRepo.NewInventoryRepository(db)
    inventoryUC := inventory.NewUseCase(invRepo)
    reservationService := integration.NewReservationService(inventoryUC)
    orderEventHandler := integration.NewOrderEventHandler(reservationService)
    
    // Register event handlers automatically
    if err := orderEventHandler.RegisterHandlers(eventBus); err != nil {
        return nil, err
    }
    return orderEventHandler, nil
}
```

**Testing**: 34 integration tests validate E2E workflows with Order Management context

**See**: [Warehouse README](internal/contexts/warehouse/README.md) for complete documentation

## Phase 3: LUA Scripting + UI Metadata Foundation

**Status**: 🚧 IN PROGRESS (Week 1 started January 7, 2026)

**Strategic Vision**: Transform Promenade into low-code enterprise platform where business users can customize logic and forms without Go recompilation.

### Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                    Business User Layer                        │
├──────────────────────────────────────────────────────────────┤
│  Visual LUA Builder  │  Form Designer  │  Template Library   │
├──────────────────────────────────────────────────────────────┤
│         LUA Scripting Engine        │    UI Metadata System  │
├──────────────────────────────────────────────────────────────┤
│                    Core DDD Contexts                          │
│  Identity │ Customer │ Order │ Billing │ Warehouse │ ...     │
├──────────────────────────────────────────────────────────────┤
│              Event Bus │ PostgreSQL │ Redis                   │
└──────────────────────────────────────────────────────────────┘
```

### LUA Scripting Engine

**Location**: `pkg/scripting/`

**Purpose**: Embedded LUA engine for dynamic business logic without Go recompilation

**Components**:
- `engine.go` (300+ lines) - LUA VM wrapper with context support
- `sandbox.go` (150+ lines) - Security restrictions (memory 50MB, timeout 5s, no filesystem/network)
- `stdlib.go` (200+ lines) - Standard library with Promenade APIs

**Standard Library API**:
```lua
-- Customer operations
local tier = Customer.GetTier("customer-123")
Customer.SetTier("customer-123", "premium")

-- Order operations
local status = Order.GetStatus("order-456")
Order.SetStatus("order-456", "confirmed")

-- Deal operations
Deal.Approve("deal-789")
Deal.Reject("deal-789", "Budget constraints")

-- Notifications
Notify.SendEmail("user@example.com", "welcome_email")
Notify.SendSMS("+1234567890", "Order shipped")

-- Safe queries (SELECT only)
local results = Query.Execute("SELECT * FROM customers WHERE tier = 'premium' LIMIT 10")

-- Date utilities
local now = Date.Now()
local month = Date.GetMonth()
```

**Use Cases**:
1. **Custom Validation Rules**: Dynamic validation without Go code
2. **Dynamic Pricing**: Tier-based pricing, seasonal discounts
3. **Workflow Automation**: Auto-escalation, notifications
4. **Business Rules Engine**: Complex approval logic
5. **Custom Reports**: Generate reports via LUA scripts

**Security Model**:
- Memory limit: 50MB (configurable)
- CPU timeout: 5s (configurable)
- No filesystem access (sandbox)
- No network access (sandbox)
- Database: Read-only SELECT queries via Query module
- Dangerous functions removed: dofile, loadfile, require, setfenv

**Performance Targets**:
- Simple scripts: < 100ms
- Complex scripts: < 500ms
- Memory: < 10MB per script
- Benchmarks included in engine_test.go

**3-Level Approach** (Progressive Enhancement):
1. **Visual Builder** (Week 3-4): Drag-and-drop workflow designer for non-developers
2. **Template System** (Week 2): Pre-built templates with parameter customization
3. **Full LUA Code** (Week 1): Direct LUA scripting for advanced users

**Testing**: 20+ tests covering:
- Simple execution (2 + 2 = 4)
- Parameters passing (x + y)
- Function execution (greet("Alice"))
- Timeout protection (infinite loop detection)
- Sandbox validation (dangerous functions blocked)
- Standard library (Customer, Order, Deal APIs)

**Dependencies**:
- `github.com/yuin/gopher-lua` - LUA interpreter in Go
- `github.com/layeh/gopher-luar` - Go-LUA value conversion

**See**: 
- [pkg/scripting/README.md](../pkg/scripting/README.md) - Complete scripting guide
- [docs/roadmap/PHASE3_LUA_UI_FOUNDATION.md](../docs/roadmap/PHASE3_LUA_UI_FOUNDATION.md) - 3-week implementation plan

### UI Metadata System (Planned Week 2-3)

**Location**: `internal/contexts/ui/metadata/form/`

**Purpose**: Oracle Forms-style metadata-driven UI system

**Concept**: Store form definitions as JSONB in PostgreSQL, generate frontend forms dynamically

**FormDefinition Aggregate**:
```go
type FormDefinition struct {
    aggregate.BaseAggregate
    Name        string              // "customer_form", "order_form"
    Context     string              // "customer-mgmt", "order-mgmt"
    Entity      string              // "Customer", "Order"
    Version     int                 // Versioning support
    IsActive    bool
    
    // JSONB stored fields
    Fields      jsonstore.Field[[]FormField]     // Field definitions
    Layout      jsonstore.Field[FormLayout]      // Layout config
    Validation  jsonstore.Field[[]ValidationRule] // Validation rules
    Events      jsonstore.Field[[]EventHandler]  // LUA event handlers
}

type FormField struct {
    Name       string  // "email", "amount"
    Label      string  // "Email Address", "Order Amount"
    Type       string  // "text", "number", "select", "date"
    Required   bool
    DefaultVal interface{}
    Visible    bool
    Readonly   bool
}
```

**Storage Strategy**:
- JSONB columns in PostgreSQL (flexible schema)
- Versioning via form_versions table
- Multi-tenant support (tenant_id column)
- Audit trail (created_by, updated_by, updated_at)

**LUA Integration**:
- Event handlers: onLoad, onChange, onSubmit, onValidate
- Custom validation rules in LUA
- Dynamic field visibility based on business rules
- Inter-field dependencies

**Example Form Definition**:
```json
{
  "name": "customer_form",
  "context": "customer-mgmt",
  "entity": "Customer",
  "fields": [
    {"name": "email", "label": "Email", "type": "text", "required": true},
    {"name": "tier", "label": "Tier", "type": "select", "options": ["free", "pro", "premium"]}
  ],
  "events": {
    "onSubmit": "function(form) Customer.SetTier(form.id, form.tier) end"
  }
}
```

**API Endpoints** (Planned):
- POST /api/v1/ui/forms - Create form definition
- GET /api/v1/ui/forms/:name - Get form by name
- PUT /api/v1/ui/forms/:name - Update form
- GET /api/v1/ui/forms/:name/versions - List versions
- POST /api/v1/ui/forms/:name/versions/:version/restore - Restore version

**Frontend Integration**:
- React component library generates forms from metadata
- No hard-coded forms in frontend
- Forms update without frontend redeployment
- Multi-language support via field labels

**Implementation Timeline**:
- Week 1: LUA Engine (STARTED)
- Week 2: Script Context + FormDefinition aggregate
- Week 3: UI Metadata API + Examples + Documentation

**See**: [docs/roadmap/PHASE3_LUA_UI_FOUNDATION.md](../docs/roadmap/PHASE3_LUA_UI_FOUNDATION.md)

## Adding a New Aggregate

1. **Structure** in `internal/contexts/{context}/{aggregate}/`:

   ```
    entity.go              # Domain entity (aggregate root)
    entity_test.go         # Entity tests
    repository.go          # IRepository interface
    usecase.go             # IUseCase interface + implementation
    usecase_test.go        # Use case tests
    adapter/
        http/handler/
           {aggregate}_handler.go
           dto/
               {aggregate}_dto.go
        repository/postgres/
            base_repository.go  # If first aggregate in context
            {aggregate}_repository.go
   ```

2. **entity.go** - Define aggregate root with factory methods:

   ```go
   package contact

   import "github.com/basilex/promenade/pkg/uuidv7"

   type Contact struct {
       ID         uuidv7.UUID
       UserID     uuidv7.UUID
       Type       ContactType
       Email      *valueobject.Email   // Value object
       IsPrimary  bool
       IsVerified bool
   }

   func NewEmailContact(userID uuidv7.UUID, email, label string) (*Contact, error) {
       emailVO, err := valueobject.NewEmail(email)
       if err != nil {
           return nil, err
       }
       return &Contact{
           ID:     uuidv7.New(),
           UserID: userID,
           Email:  &emailVO,
       }, nil
   }
   ```

3. **repository.go** - Define repository interface:

   ```go
   type IRepository interface {
       Create(ctx context.Context, contact *Contact) error
       GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error)
       Update(ctx context.Context, contact *Contact) error
       Delete(ctx context.Context, id uuidv7.UUID) error
   }
   ```

4. **Register in context router** (`internal/contexts/{context}/router.go`):

   ```go
   func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
       context := api.Group("/context")
       {
           aggregates := context.Group("/aggregates")
           {
               aggregates.POST("", r.aggregateHandler.Create)
               aggregates.GET("/:id", r.aggregateHandler.GetByID)
           }
       }
   }
   ```

5. **Migrations**: `make migrate-new CONTEXT={context} NAME=add_{aggregate}_table`

## CQRS for Analytics (Customer Management)

**When to use CQRS**: For read-heavy analytical queries that span multiple aggregates.

**Pattern**: Analytics module (`internal/contexts/customer-mgmt/analytics/`)

```go
// NO repository interface - direct SQL for optimal reads
type Analytics struct {
    db *sqlx.DB
}

// Denormalized query with LEFT JOIN across aggregates
func (a *Analytics) GetCustomerOverview(ctx context.Context) (*CustomerOverview, error) {
    query := `
        SELECT 
            COUNT(DISTINCT c.id) as total_customers,
            COUNT(DISTINCT d.id) as total_deals,
            COUNT(DISTINCT i.id) as total_interactions
        FROM customer_customers c
        LEFT JOIN customer_deals d ON d.customer_id = c.id
        LEFT JOIN customer_interactions i ON i.customer_id = c.id
        WHERE c.deleted_at IS NULL
    `
    var overview CustomerOverview
    err := a.db.GetContext(ctx, &overview, query)
    return &overview, err
}
```

**Key differences from regular aggregates**:
- NO repository pattern (direct DB access)
- NO business logic (read-only reporting)
- Denormalized queries (performance > normalization)
- Separate read models from write models

**See**: [Analytics README](internal/contexts/customer-mgmt/analytics/README.md) for complete CQRS implementation

## Event Bus & Configuration

**Event Bus** - Dual adapters (factory with graceful fallback):

- **Memory** (`pkg/bus/memory`): Dev/test, in-process, fast
- **Redis** (`pkg/bus/redis`): Prod, distributed, persistent
- Config: `bus.adapter: memory|redis` in `config/app.{env}.yaml`

**Cross-Context Integration Pattern** (see Warehouse-Order integration):

Contexts communicate ONLY via Event Bus - no direct imports. Example flow:

```go
// 1. Order context publishes domain event
func (o *Order) Confirm() error {
    o.Status = OrderStatusConfirmed
    event := bus.NewBaseEvent("order.confirmed", o.ID)
    eventBus.Publish(ctx, bus.TopicOrderConfirmed, event)
    return nil
}

// 2. Warehouse context subscribes to event
func (h *OrderEventHandler) RegisterHandlers(eventBus bus.IBus) error {
    return eventBus.Subscribe(bus.TopicOrderConfirmed, h.HandleOrderConfirmed)
}

// 3. Handler processes event
func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, e bus.Event) error {
    var event OrderConfirmedEvent
    json.Unmarshal(e.Metadata()["payload"].([]byte), &event)
    return h.reservationService.ReserveForOrder(ctx, event.OrderID, event.Items, event.ConfirmedBy)
}

// 4. Bootstrap registers handlers automatically (cmd/api/bootstrap.go)
orderEventHandler, _ := initWarehouseIntegration(db, eventBus)
```

**Key Principles**:
- Events are fire-and-forget (log errors, don't fail operations)
- Handlers registered at bootstrap (not runtime)
- Use topic constants (`bus.TopicOrderConfirmed`) not strings
- Event payloads in metadata as JSON
- Idempotent handlers (safe to retry)

**Config Loading**:

- Core: `config/app.{env}.yaml` (ENVIRONMENT=dev/test/prod)
- Overrides: Sensitive values via env vars (DB_PASSWORD, JWT_SECRET, REDIS_ADDR)
- Access: `logger.FromContext(ctx)`, `database.GetTx(ctx)`

## Health Checks & Monitoring

**Health Checker**: Monitors all dependencies with graceful degradation

```go
// Initialize in cmd/api/main.go
healthChecker := health.NewChecker(db, redisClient, eventBus, cfg.App.Version)
healthHandler := health.NewHandler(healthChecker)
healthHandler.RegisterRoutes(r)
```

**Available Endpoints**:
- `GET /health` - Overall system health (DB + Redis + Event Bus)
- `GET /health/db` - PostgreSQL health check
- `GET /health/redis` - Redis health check (optional)
- `GET /health/bus` - Event Bus health check

**Status Levels**: `healthy` (200), `degraded` (200), `unhealthy` (503)

## Caching Layer

**Cache Client**: Redis-based caching with graceful fallback

```go
// Initialize in cmd/api/main.go
cacheClient, err := cache.NewCache(cacheConfig, cacheRedisClient)

// Pass to context routers that need caching
sharedRouter := shared.NewRouter(db, cacheClient)  // Reference data with cache
```

**Cache Usage in Repository**:

```go
// Try cache first
cached, err := r.cache.Get(ctx, cacheKey, &countries)
if err == nil && cached {
    return countries, nil  // Cache hit
}

// Cache miss - fetch from DB
countries, err := r.fetchFromDB(ctx)
if err != nil {
    return nil, err
}

// Store in cache
r.cache.Set(ctx, cacheKey, countries, ttl)
return countries, nil
```

**TTL Strategy**:
- Reference data: 1h-24h (Country, Currency, Language, Timezone)
- User data: 10-30m (Profile, Customer)
- Session data: 30m-1h (temporary state)

**Invalidation**: Pattern-based deletion on updates

```go
// Invalidate all country-related cache entries
r.cache.DeletePattern(ctx, "countries:*")
```

## Testing Strategy

**Test Organization**: Tests live alongside code (`*_test.go` in same directory) with additional integration/benchmark/smoke tests in mirror path structure

**Four-Tier Test Strategy**:

1. **Unit Tests** (in-place): Fast feedback, test individual components
   - Location: Same directory as production code (`entity_test.go`, `usecase_test.go`)
   - Run: `make test-unit` (~5s)
2. **Smoke Tests** (`test/smoke/contexts/`): HTTP handler validation (80/20 rule)
   - Location: Mirror path structure (e.g., `test/smoke/contexts/order-mgmt/order/handler_test.go`)
   - Run: `make test-smoke` (~2s, no DB needed)
   - Purpose: Verify HTTP status codes, response format, routing, error mapping
   - Coverage: 2 tests per handler (success + not found)
3. **Integration Tests** (`test/integration/contexts/`): Full E2E with real database
   - Location: Mirror path structure (e.g., `test/integration/contexts/identity/contact/repository_test.go`)
   - Run: `make test-integration` (~14s, auto-starts test DB)
4. **Benchmark Tests** (`test/benchmark/contexts/`): Performance measurement with real database
   - Location: Mirror path structure (e.g., `test/benchmark/contexts/identity/user/repository_bench_test.go`)
   - Run: `make test-benchmark` (~5s per benchmark, auto-starts test DB)
   - Purpose: Validate optimizations (e.g., N+1 query fixes), measure query performance

**Running Tests**:

```bash
make test                      # All tests with race detector (~40s)
make test-unit                 # Unit tests only (~5s)
make test-smoke                # Smoke tests for handlers (~2s, no DB)
make test-integration          # Integration tests with real DB (~14s)
make test-benchmark            # Benchmark tests (5s per benchmark)
make test-coverage             # HTML coverage report

# Context-specific tests
go test ./internal/contexts/identity/... -v
go test ./test/smoke/contexts/order-mgmt/order -v
go test ./test/integration/contexts/identity/... -v
go test -bench=. ./test/benchmark/contexts/identity/user -v

# Package tests
go test ./pkg/bus/... -v
go test ./pkg/uuidv7/... -v
go test ./pkg/jwt/... -v
```

**Test Statistics**: 2200+ tests across 88+ packages, 90%+ average coverage

**DTO Testing Guidelines**:

** WRITE DTO tests for**:
- Complex transformation logic (value objects: Email, Phone, Address, Money)
- JSONB fields with custom marshaling
- Conditional logic for nullable/optional fields
- Nested structures with multiple levels

** SKIP DTO tests for**:
- Simple field-to-field mappings
- Straightforward struct copying
- Already covered by handler integration tests
- Read-only responses without transformation

**Examples with DTO tests**: Contact (value objects), User (JSONB roles), Profile (many nullables), Customer (JSONB tags)  
**Examples without DTO tests**: Order, Deal, Interaction, Analytics, Company (simple structures, covered by handlers)

**Smoke Testing Pattern**:

**Purpose**: Validate HTTP handlers with minimal effort (80/20 rule)  
**Location**: `test/smoke/contexts/{context}/{aggregate}/handler_test.go`  
**Coverage**: 6-12 tests per handler (6 minimum, 9 standard, 12 for complex)  
**Run**: `make test-smoke` (~0.6s, no database needed)
**Run**: `make test-smoke` (~2s, no database needed)

**Key Characteristics**:
- Mock UseCase with function fields (no external dependencies)
- Test HTTP status codes (200, 201, 404, 400, 500)
- Validate response format (`{"status":"success","data":{...}}`)
- No database, no complex business logic
- Fast execution (< 100ms per test)

**Example**:
```go
// test/smoke/contexts/order-mgmt/order/handler_test.go
package order_test

import (
    "testing"
    "github.com/basilex/promenade/test/smoke"
)

// MockOrderUseCase with function fields
type MockOrderUseCase struct {
    GetByIDFunc func(ctx context.Context, id uuid.UUID) (*order.Order, error)
}

func TestOrderHandler_GetByID_Success(t *testing.T) {
    router := smoke.SetupRouter()
    
    // Mock UseCase returns fake order
    mockUC := &MockOrderUseCase{
        GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*order.Order, error) {
            return fakeOrder(), nil
        },
    }
    
    handler := orderHTTP.NewOrderHandler(mockUC)
    router.GET("/orders/:id", handler.GetByID)
    
    // Test request
    resp := smoke.MakeRequest(t, router, "GET", "/orders/"+smoke.FakeUUID(), nil)
    smoke.AssertSuccessResponse(t, resp, 200)
}

func TestOrderHandler_GetByID_NotFound(t *testing.T) {
    router := smoke.SetupRouter()
    
    mockUC := &MockOrderUseCase{
        GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*order.Order, error) {
            return nil, order.ErrOrderNotFound
        },
    }
    
    handler := orderHTTP.NewOrderHandler(mockUC)
    router.GET("/orders/:id", handler.GetByID)
    
    resp := smoke.MakeRequest(t, router, "GET", "/orders/"+smoke.FakeUUID(), nil)
    smoke.AssertErrorResponse(t, resp, 404, "ORDER_NOT_FOUND")
}
```

**Helper Utilities** (`test/smoke/testutils.go`):
- `SetupRouter()` - Gin test router
- `MakeRequest(t, router, method, path, body)` - HTTP request helper
- `AssertSuccessResponse(t, resp, expectedCode)` - Validate success response
- `AssertErrorResponse(t, resp, expectedCode, expectedErrorCode)` - Validate error response
- `FakeUUID()` - Generate test UUID v7

**When to Write Smoke Tests**:
-  For all HTTP handlers (6-12 tests per handler)
-  When adding new endpoints
-  Before integration tests (faster feedback)
-  Not for complex business logic (use integration tests)
-  Not for repository methods (use integration tests with real DB)

**See**: [test/smoke/README.md](../test/smoke/README.md) for complete smoke testing guide

**Example Entity Test**:

```go
func TestContact_NewEmailContact(t *testing.T) {
    userID := uuidv7.New()
    contact, err := NewEmailContact(userID, "test@example.com", "Work")

    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, contact.ID)
    assert.Equal(t, userID, contact.UserID)
    assert.Equal(t, ContactTypeEmail, contact.Type)
    assert.NotNil(t, contact.Email)
}
```

**Example UseCase Test** (with mock):

```go
func TestUseCase_CreateEmailContact(t *testing.T) {
    mockRepo := NewMockRepository()
    usecase := NewUseCase(mockRepo)

    contact, err := usecase.CreateEmailContact(ctx, userID, "test@example.com", "Work", false)

    assert.NoError(t, err)
    assert.NotNil(t, contact)
    assert.True(t, mockRepo.CreateCalled)
}
```

**Integration Test Pattern** (real DB):

```go
func TestUserRepository_Create(t *testing.T) {
    db := integration.SetupTestDBWithCleanTables(t)  // Auto-starts test DB, runs migrations
    repo := postgres.NewUserRepository(db.DB)
    ctx := context.Background()

    u, _ := user.NewUser("test@example.com", "password123")
    err := repo.Create(ctx, u)
    
    require.NoError(t, err)
    
    // Verify persistence
    retrieved, err := repo.GetByID(ctx, u.ID)
    require.NoError(t, err)
    assert.Equal(t, u.ID, retrieved.ID)
}
```

**Integration Test Common Pitfalls** (from Dec 2025 fixes):

1. **Unique Constraint Violations**: Add UUID suffix to test emails in loops
   ```go
   //  WRONG - duplicate emails in loop
   for i := 0; i < 3; i++ {
       email := fmt.Sprintf("test_%d@example.com", i)
   }
   
   //  CORRECT - unique emails with UUID suffix
   for i := 0; i < 3; i++ {
       email := fmt.Sprintf("test_%s_%d@example.com", uuidv7.New().String()[:8], i)
   }
   ```

2. **Transaction Rollback Tests**: Don't use `t.FailNow()` in transaction tests
   ```go
   //  WRONG - t.FailNow() prevents rollback verification
   tx, _ := db.BeginTx(ctx, nil)
   if err != nil {
       t.FailNow()  // Code after this never runs!
   }
   
   //  CORRECT - manual transaction + separate verification
   tx, _ := db.BeginTx(ctx, nil)
   ctx = database.SetTxToContext(ctx, tx)
   // ... test code ...
   tx.Rollback()
   // Verify with new transaction
   ```

3. **Pagination Parameters**: pageSize before page number
   ```go
   //  WRONG - page=10, pageSize=0 → LIMIT 0
   users, _ := repo.ListUsers(ctx, 10, 0)
   
   //  CORRECT - page=1, pageSize=10
   users, _ := repo.ListUsers(ctx, 1, 10)
   ```

**Test Helpers**: [test/integration/](../test/integration/) for DB setup and utilities

## Codebase Conventions & Patterns

### Naming Conventions

**Interfaces**: PascalCase with `I` prefix

```go
type IRepository interface { }      // Repository interface
type IUseCase interface { }         // Use case interface
```

**Implementations**: lowercase private structs

```go
type contactRepository struct { *BaseRepository }
func NewContactRepository(db *sqlx.DB) IRepository {
    return &contactRepository{BaseRepository: NewBaseRepository(db)}
}

type useCase struct { repo IRepository }
func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}
```

**Handlers**: PascalCase with `Handler` suffix

```go
type ContactHandler struct { usecase IUseCase }
func NewContactHandler(uc IUseCase) *ContactHandler {
    return &ContactHandler{usecase: uc}
}
```

**Repository Methods**: Standard CRUD prefixes

```go
GetByID(ctx, id)              // Retrieve single entity
GetByUserID(ctx, userID)      // Query by field
Create(ctx, entity)           // Insert new
Update(ctx, entity)           // Modify existing
Delete(ctx, id)               // Remove (soft delete)
List{Entity}s(ctx, filters)   // Query multiple
```

**Errors**: Prefix with `Err{Entity}{Condition}`

```go
var (
    ErrContactNotFound      = errors.New("contact not found")
    ErrEmailAlreadyExists   = errors.New("email already exists")
    ErrUnauthorizedAccess   = errors.New("unauthorized access")
)

// Wrap errors with context
return nil, fmt.Errorf("failed to create contact: %w", err)
```

**Files & Directories**:

- Aggregates: `entity.go`, `repository.go`, `usecase.go`
- Tests: `entity_test.go`, `usecase_test.go`
- Handlers: `{aggregate}_handler.go`
- DTOs: `{aggregate}_dto.go`
- Directory names: singular (`entity/`, `handler/`, `repository/`)

### Validation Patterns

**Entity Validation** (business rules in domain):

```go
func (c *Contact) Validate() error {
    if c.Email != nil && c.Email.Value() == "" {
        return fmt.Errorf("email cannot be empty")
    }
    // ... other validations
    return nil
}
```

**Request Validation** (Gin binding tags):

```go
type CreateContactDTO struct {
    Email   string `json:"email" binding:"required,email"`
    Label   string `json:"label" binding:"required,min=1,max=50"`
}
```

**Business Logic Validation** (in use case):

```go
// Check uniqueness before creating
exists, err := uc.repo.ExistsPrimaryForUserAndType(ctx, userID, ContactTypeEmail)
if exists {
    return nil, fmt.Errorf("primary email already exists")
}
```

### Error Handling Flow

**Handler checks domain errors and returns HTTP codes**:

```go
contact, err := h.usecase.GetContact(ctx, contactID)
if errors.Is(err, ErrContactNotFound) {
    response.Error(c, http.StatusNotFound, "CONTACT_NOT_FOUND", err.Error())
    return
}
if err != nil {
    response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
    return
}
response.Success(c, contact)
```

**UseCase wraps errors with context**:

```go
if err := uc.repo.Create(ctx, contact); err != nil {
    return nil, fmt.Errorf("failed to create contact: %w", err)
}
```

### Entity State Methods

**Encapsulate state changes in methods**:

```go
func (c *Contact) SetAsPrimary() {
    c.IsPrimary = true
}

func (c *Contact) Verify() {
    c.IsVerified = true
}

func (c *Contact) UpdateVisibility(isPublic bool) {
    c.IsPublic = isPublic
}
```

### DTO Conversion

**Handler converts between DTOs and entities**:

```go
// Request → UseCase parameters
contact, err := h.usecase.CreateEmailContact(
    c.Request.Context(),
    req.UserID,
    req.Email,
    req.Label,
    req.IsPrimary,
)

// Entity → Response DTO
response.Success(c, ToContactDTO(contact))
```

### Database Conventions

- Snake_case: `user_id`, `created_at`, `is_primary`, `contact_type`
- Soft delete: `deleted_at TIMESTAMP NULL` (always filter in queries)
- Timestamps: `created_at`, `updated_at` with `DEFAULT CURRENT_TIMESTAMP`
- Foreign keys: `{table}_id` (e.g., `user_id`, `contact_id`)

### Response Helpers

**Use pkg/response for consistent API responses**:

```go
response.Success(c, data)                    // 200 OK with data
response.Error(c, code, "ERROR_CODE", msg)   // Error with code and message
```

## Critical Gotchas

**Top Mistakes**:

1. **BaseAggregate Field Duplication** (FIXED Jan 2026): NEVER duplicate ID, CreatedAt, UpdatedAt in entities - already in BaseAggregate
   -  `type Entity struct { aggregate.BaseAggregate; ID uuid.UUID }` - WRONG
   -  `type Entity struct { aggregate.BaseAggregate }` - CORRECT
   - Always use `entity.Touch()` instead of `entity.UpdatedAt = time.Now()`
   - Always use `entity.GetID()` instead of `entity.ID` in repositories
2. **UUID v4 vs v7**: NEVER `uuid.New()` (v4). Always `pkg/uuidv7.New()` (time-ordered)
3. **Soft Delete**: Always `WHERE deleted_at IS NULL` in SELECT queries
4. **Context Isolation**: Contexts communicate ONLY via Event Bus (no direct imports between contexts)
5. **Context Chain**: Always pass `ctx`. `getExecutor(ctx)` needs it for tx/db selection
6. **Logger**: `logger.FromContext(ctx)` not global logger (preserves request context)
7. **Migration Namespaces**: Migrations run in order: core → shared → identity → customer-mgmt → order-mgmt

**Debugging Quick Reference**:

- **Workspace Not Configured**: Run `make workspace` to check state, then `make switch-postgres-dev` or `make switch-sqlite-dev`
- **Wrong Environment**: `make dev` requires ENVIRONMENT=development, `make test-all` warns if not test
- **DB Connection**: `make docker-ps` → check Postgres on 5432, test DB on 5433 (SQLite: no Docker needed)
- **Test Failures**: Start with `make test-unit` (~5s, no workspace), then `make test-integration` (~14s, needs workspace)
- **Migration Issues**: Check `schema_migrations` table, verify namespace (`core/`, `shared/`, `identity/`, `customer-mgmt/`, `order-mgmt/`)
- **Context Won't Load**: Verify router imported and registered in `cmd/api/bootstrap.go`
- **Event Bus Issues**: Check adapter config (`memory` for dev, `redis` for prod) in `config/app.{env}.yaml`
- **Integration Test DB**: Auto-started on port 5433, uses `promenade_test` database (Docker container for Postgres)

## Code Review Checklist

**Before submitting PR, verify every file**:

### Naming Consistency

- [ ] All interfaces: `I{Entity}UseCase`, `I{Entity}Repository` (full words, no abbreviations)
  - **Exception**: Most contexts use `IUseCase` and `IRepository` (generic names)
  - **Exception**: Customer Management uses `ICustomerUseCase`, `ICustomerRepository` (entity-specific)
- [ ] All structs: lowercase `useCase`, `{entity}Repository` (useCase is generic, repos are entity-specific)
- [ ] All constructors: `NewUseCase() IUseCase` (NEVER `New{Entity}UseCase`) and `New{Entity}Repository()`
- [ ] All handlers: `{Entity}Handler struct { {entity}UC usecase.I{Entity}UseCase }`
- [ ] All errors: `Err{Entity}{Condition}` (e.g., `ErrUserNotFound`)
- [ ] All files: `{entity_name}_{type}.go` (snake_case)

### Method Signatures

- [ ] Context first: `func Method(ctx context.Context, ...)`
- [ ] Repository methods: `GetByID`, `GetByXxx`, `Create`, `Update`, `Delete`, `ListXxx`
- [ ] Error returns: Always `(*Entity, error)` or `([]*Entity, error)` or `error`
- [ ] Receiver names: `r` (repo), `uc` (usecase), `h` (handler), first letter of entity

### Code Structure

- [ ] Handler calls usecase, NOT repository directly
- [ ] UseCase has business logic, handler has HTTP logic
- [ ] Repository has ONLY data access, no business logic
- [ ] Entity has validation methods, not validation in usecase
- [ ] DTOs convert in adapter layer, never in usecase
- [ ] Errors defined in usecase, checked in handler with `errors.Is()`

### SQL & Data

- [ ] All IDs use `uuidv7.New()`, never `uuid.New()`
- [ ] Soft delete queries include `WHERE deleted_at IS NULL`
- [ ] Positional parameters: `$1`, `$2` (never named in raw SQL)
- [ ] Repository methods use `getExecutor(ctx)` for tx support
- [ ] All timestamps: `created_at`, `updated_at`, `deleted_at` (snake_case)

### Documentation

- [ ] Swagger comments on all public handlers
- [ ] Error variable has comment: `//ErrUserNotFound is returned when...`
- [ ] Complex logic has brief inline comment
- [ ] README.md updated if module added/changed

### Testing

- [ ] Unit tests for usecase business logic
- [ ] Integration tests for repository methods
- [ ] Test file: `{filename}_test.go` in same directory
- [ ] Manual mocks created in `mocks_test.go` if needed (inline structs)
- [ ] DTO tests ONLY for complex cases (value objects, JSONB, conditional logic)
- [ ] Skip DTO tests for simple field-to-field mappings (covered by handler tests)

**Red Flags** (automatic rejection):

- Interface name `UserPostUC` (use `IUserPostUseCase`)
- Public struct `type UserPostUseCase struct` (must be lowercase `useCase`)
- Constructor `NewCustomerUseCase()` or `NewContactUseCase()` (use simple `NewUseCase()`)
- Repository method `FindByID` (use `GetByID`)
- `uuid.New()` instead of `uuidv7.New()`
- Handler calls repository directly (must go through usecase)
- Business logic in handler or repository (must be in usecase)
- Missing `ctx context.Context` as first parameter
## Key Files

| File                                                           | Purpose                           |
| -------------------------------------------------------------- | --------------------------------- |
| [cmd/api/main.go](../cmd/api/main.go)                          | Entry point, context registration |
| [cmd/api/bootstrap.go](../cmd/api/bootstrap.go)                | Dependency injection & initialization |
| [Makefile](../Makefile) + [Makefile.\*.mk](../Makefile.dev.mk) | All workflows (modular system)           |
| [.promenade.workspace](../.promenade.workspace.example)        | Workspace state (DATABASE_DRIVER + ENVIRONMENT) |
| [pkg/uuidv7/uuidv7.go](../pkg/uuidv7/uuidv7.go)                | Time-ordered UUIDs                |
| [pkg/jwt/README.md](../pkg/jwt/README.md)                      | JWT authentication docs           |
| [pkg/bus/README.md](../pkg/bus/README.md)                      | Event Bus documentation           |
| [internal/infrastructure/database/transaction.go][tx]          | Transaction management            |
| [internal/contexts/shared/router.go][shared]                   | Shared context router             |
| [internal/contexts/identity/router.go][identity]               | Identity context router           |
| [docs/work-in-progress/](../docs/work-in-progress/)            | Current work items & TODOs        |
| [docs/](../docs/)                                              | Architecture guides (30+ docs)    |
| [test/README.md](../test/README.md)                            | Testing guide                     |

[tx]: ../internal/infrastructure/database/transaction.go
[shared]: ../internal/contexts/shared/router.go
[identity]: ../internal/contexts/identity/router.go

---

**For comprehensive documentation**: [README.md](../README.md) | [docs/INDEX.md](../docs/INDEX.md) | [docs/concepts/clean-architecture.md](../docs/concepts/clean-architecture.md)
