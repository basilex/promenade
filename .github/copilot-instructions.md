# Promenade AI Instructions

Essential guide for AI agents working in Promenade. For detailed documentation, see [README.md](../README.md) and [docs/](../docs/).

---

## Quick Start for AI Agents

**First Time Here?** Start with these commands to orient yourself:

```bash
# 1. Check current workspace configuration
make workspace                  # Shows DATABASE_DRIVER + ENVIRONMENT

# 2. Switch to your preferred setup (pick one)
make switch-postgres-dev        # PostgreSQL + development (most common)
make switch-sqlite-dev          # SQLite + development (no Docker needed)
make switch-sqlite-test         # SQLite + testing (fast tests)

# 3. Start development
make dev                        # Starts server + migrations
# OR for a clean start:
make dev-fresh                  # Clean DB + migrations + server

# 4. Run tests (four-tier strategy)
make test-unit                  # Fast unit tests (~5s, no DB)
make test-smoke                 # HTTP handler tests (~0.6s, no DB)
make test-integration           # Full E2E with DB (~14s)
make test                       # All tests with race detector (~60s)

# 5. Before any commit
make pre-push                   # Lint + test + build (catches CI failures early)
```

**Key Architecture Concepts**:
- **Bounded Contexts** = Autonomous business domains (Identity, Customer, Order, Billing, Warehouse)
- **Event Bus** = Contexts communicate ONLY via events (no direct imports)
- **errors.go** = All domain errors centralized (NEVER use fmt.Errorf in usecase.go)
- **BaseAggregate** = All entities embed it (provides ID, Version, timestamps)
- **UUID v7** = Always use `uuidv7.New()` (time-ordered, 2x faster inserts)

**Critical Files to Read**:
- `.promenade.workspace` - Current database driver + environment
- `cmd/api/bootstrap.go` - Dependency injection order (Logger → DB → Redis → Event Bus → Routers)
- `docs/guides/unified-error-handling-standard.md` - Error handling patterns (1287 lines, master reference)
- `test/smoke/README.md` - HTTP handler testing guide (182 tests, 100% pass rate)

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
- **Order Management** (`internal/contexts/order-mgmt/`) - Order aggregate (Production) | OrderLine entity | Contract, FulfillmentSaga (COMPLETE) 
- **Billing** (`internal/contexts/billing/`) - Invoice, Payment, Subscription (all Production)
- **Warehouse** (`internal/contexts/warehouse/`) - Inventory, StockMovement, Product, Location (ALL Production, 100% complete)
- **Scripting** (`internal/contexts/scripting/`) - Script aggregate (HTTP Layer Complete) | LUA execution engine | 10 REST endpoints operational

**Context isolation**: Contexts communicate ONLY via Event Bus (no direct dependencies)

**Latest Progress** (January 12, 2026):
-  Phase 1 COMPLETE: Handler Security Audit (36 handlers, 417 fixes, 100% production-ready)
-  Phase 2 🎉 MAJOR MILESTONE: Domain Errors Refactoring (11/18 sessions complete - 61.1%)
  - Session 1 COMPLETE: warehouse/location (14 domain constants, 70 fixes, GOLD STANDARD)
  - Session 2 COMPLETE: warehouse/inventory (22 domain constants, 17 fmt.Errorf eliminated)
  - Session 3 COMPLETE: warehouse/stockmovement (13 domain constants, 28 fmt.Errorf eliminated)
  - Session 4 COMPLETE: warehouse/product (27 domain constants, 60 total replacements)
  - Session 5 COMPLETE: identity/role (5 domain constants, 8 fmt.Errorf eliminated, 49 tests)
  - Session 6 COMPLETE: identity/permission (6 domain constants, 7 fmt.Errorf eliminated, 52 tests)
  - Session 7 COMPLETE: identity/profile (16 domain constants, 19 fmt.Errorf eliminated, 36 tests)
  - Session 8 COMPLETE: identity/contact (11 domain constants, 17 fmt.Errorf handled, 32 tests)
  - Session 9 COMPLETE: order-mgmt/contract (8 domain constants, production-ready)
  - Session 10 COMPLETE: Entity Tests Refactoring (14 patterns across 3 files, 4 new error constants)
  - Session 11 COMPLETE: UseCase Tests Refactoring (22+ patterns across 7 files, errors.Is() migration)
  - Warehouse Context: 100% complete (all 4 aggregates production-ready)
  - Identity Context: 100% complete (all 4 aggregates production-ready)
  - Entity Tests: 100% type-safe (contact, customer, interaction - all using errors.Is())
  - UseCase Tests: ~88% type-safe (inventory 10, product 9, deal 2, adaptive strategy for wrapped errors)
  - Remaining: 7 sessions (Customer-Mgmt aggregates, Billing aggregates, Integration Tests, Documentation)
-  Phase 3 IN PROGRESS: LUA Scripting + UI Metadata Foundation (Week 1 Day 4 Complete - HTTP Layer Operational)
-  Order Management: Entity tests in progress (entity_test.go active development)
-  Fulfillment Saga COMPLETE: Distributed transaction orchestration (orchestrator.go, 104 lines, 57 tests)
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
- Phase 3 LUA Scripting Engine: HTTP Layer Complete (Week 1 Day 4 Complete)
  - pkg/scripting/engine.go: LUA VM wrapper with context support (210 lines)
  - pkg/scripting/sandbox.go: Security restrictions (90 lines, memory 50MB, timeout 5s, no filesystem/network)
  - pkg/scripting/stdlib.go: Standard Library integrated with real UseCases (185 lines, Customer/Order/Deal/Query/Date APIs)
  - pkg/scripting/engine_test.go: 21 tests + 3 benchmarks (all passing)
  - pkg/scripting/README.md: Complete documentation (600+ lines)
  - internal/contexts/scripting/: Full context with Script aggregate, HTTP layer, Router
  - 10 REST Endpoints: /api/v1/scripts/* (Create, List, GetByID, Update, Delete, Execute, Validate, etc.)
  - 12 Smoke Tests: Handler validation (100% pass rate)
  - Standard Library: Real UseCase integration (CustomerUseCase, OrderUseCase, DealUseCase, *sqlx.DB)
  - Dependencies: gopher-lua (LUA interpreter), gopher-luar (Go-LUA bridge) - installed
  - Status: HTTP layer operational, Standard Library integrated, ready for Week 2 (Script storage)
- Error Handling Documentation COMPLETE:
  - docs/guides/unified-error-handling-standard.md (1287 lines, master reference)
  - docs/reference/DOMAIN_ERRORS_REFACTORING_PLAN.md (23,193 lines, 18-session plan)
  - docs/reference/domain-errors-audit.md (12,504 lines, initial audit)
  - Session summaries: [docs/refactoring/sessions/](docs/refactoring/sessions/) (11 compact summaries)
- Test Infrastructure: All systems validated (2465+ tests: 2232+ unit, 182+ smoke, 76+ integration)
- Order Entity Tests: Active development in progress (entity_test.go)

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
type Order struct {
    aggregate.BaseAggregate  // Provides: ID, Version, CreatedAt, UpdatedAt, DeletedAt
    OrderNumber string       // ORD-2026-001
    CustomerID  uuidv7.UUID
    Lines       []OrderLine
    Total       valueobject.Money
    Status      OrderStatus
    // ... other fields
}

//  NEVER DO THIS - Field duplication
type WrongOrder struct {
    aggregate.BaseAggregate
    ID        uuidv7.UUID  //  DUPLICATE - already in BaseAggregate
    CreatedAt time.Time    //  DUPLICATE - already in BaseAggregate
    UpdatedAt time.Time    //  DUPLICATE - already in BaseAggregate
}

// Factory method - BaseAggregate auto-initializes ID, CreatedAt, UpdatedAt
func NewOrder(customerID uuidv7.UUID, currency string) (*Order, error) {
    if customerID == uuidv7.Nil {
        return nil, ErrCustomerIDRequired
    }
    return &Order{
        BaseAggregate: aggregate.NewBaseAggregate(),  // Sets ID, timestamps
        OrderNumber:   generateOrderNumber(time.Now()),
        CustomerID:    customerID,
        Lines:         []OrderLine{},
        Total:         valueobject.Money{Amount: 0, Currency: currency},
        Currency:      currency,
        Status:        OrderStatusPending,
        OrderDate:     time.Now(),
    }, nil
}

//  Use Touch() for timestamp updates
func (o *Order) Confirm() error {
    if o.Status != OrderStatusPending {
        return ErrOrderAlreadyConfirmed
    }
    o.Status = OrderStatusConfirmed
    now := time.Now()
    o.ConfirmedAt = &now
    o.Touch()  // Updates UpdatedAt via BaseAggregate
    return nil
}

//  Use GetID() for ID access
func (r *orderRepository) Update(ctx context.Context, order *Order) error {
    query := `UPDATE orders SET ... WHERE id = $1`
    return r.Exec(ctx, query, order.GetID())  // Not order.ID
}
```

**Repository with BaseRepository**:

```go
// Repository interface in aggregate package (usually IRepository)
type IRepository interface {
    Create(ctx context.Context, contact *Contact) errorand `ICustomerUseCase` (entity-specific names for backward compatibility
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
    CreateEmailContact(ctx context.Context, u11, 2026):
All contexts now consistently use lowercase `type useCase struct` pattern:
- Identity context: ✅ Refactored (User, Profile, Role, Permission, Contact)
- Customer Management: ✅ Uses `ICustomerUseCase` + lowercase `useCase` (legacy compatibility)
- Order Management: ✅ Consistent lowercase pattern
- Billing, Warehouse: ✅ Already consistent

**Rule**: Always use lowercase `type useCase struct` for new code and when refactoring existing code.
**Exception**: Customer Management retains `ICustomerUseCase` interface name for backward compatibility
type useCase struct {
    repo IRepository
}

// Constructor: simple NewUseCase (NOT New{Entity}UseCase)
func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}
```

**EXCEPTION**: Customer Management uses `ICustomerUseCase` (entity-specific name)

**Naming Consistency** (standardized January 9, 2026):
All contexts now consistently use lowercase `type useCase struct` pattern:
- Identity context: ✅ Refactored (User, Profile, Role, Permission, Contact)
- Customer Management, Order Management, Billing, Warehouse: ✅ Already consistent

**Rule**: Always use lowercase `type useCase struct` for new code and when refactoring existing code.

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

---

### Fulfillment Saga

**Distributed transaction orchestration** for order fulfillment process - coordinates Payment, Inventory, and Shipping operations with automatic compensation on failures.

**Concept**: The Fulfillment Saga implements the **Saga pattern** to manage complex, multi-step order fulfillment without requiring distributed ACID transactions. It ensures data consistency across multiple bounded contexts (Payment, Inventory, Shipping) through choreographed state transitions and compensating actions.

**Design**:
- **State Machine**: 7 states (pending → payment_processing → inventory_processing → shipping_processing → completed)
- **Compensation Logic**: Automatic rollback on failures (compensating → compensated → cancelled)
- **Optimistic Locking**: Version-based concurrency control prevents lost updates
- **JSONB Storage**: Flexible arrays for completed steps and reserved items
- **UTC Timestamps**: Consistent timezone handling across all operations
- **Idempotent Steps**: Safe to retry operations without side effects

**Key Features**:
- Orchestrator pattern coordinates all fulfillment steps
- PostgreSQL persistence with full ACID guarantees
- Concurrent saga execution with conflict detection
- Monitoring endpoint for in-progress sagas
- 100% test coverage (57 tests: 42 unit + 15 integration)

**State Flow**:
```
pending → payment_processing → inventory_processing → shipping_processing → completed
              ↓ (on failure)
         compensating → compensated → cancelled
```

**Implementation Status**: Production-ready. All 57 tests passing (100%). Repository with optimistic locking. Orchestrator with compensation logic.

**See**: [Fulfillment Saga README](internal/contexts/order-mgmt/fulfillment/README.md) for complete implementation guide

---

## Warehouse Context

**Current Status** (January 7, 2026):  **100% COMPLETE** - ALL 4 Aggregates PRODUCTION

**Inventory Aggregate** (`internal/contexts/warehouse/inventory/`) -  **PRODUCTION**:

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

**StockMovement Aggregate** (`internal/contexts/warehouse/stockmovement/`) -  **PRODUCTION**:

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

**Product Aggregate** (`internal/contexts/warehouse/product/`) -  **PRODUCTION**:

- **Catalog Management**: SKU, name, description, category, brand
- **Stock Tracking**: Track across all warehouses
- **Multi-location**: Support for distributed inventory
- **16 API Endpoints**: Complete CRUD + business operations
- **139 Tests Passing**: 25 entity + 83 usecase + 10 smoke + 21 integration

**Location Aggregate** (`internal/contexts/warehouse/location/`) -  **PRODUCTION**:

- **Warehouse Locations**: Name, code, type (warehouse/retail/dropship/virtual)
- **Address Management**: Full address with country integration
- **Capacity Tracking**: Track location capacity and utilization
- **Hierarchical Structure**: Support for zones and bins
- **14 API Endpoints**: Complete CRUD + location operations
- **74 Tests Passing**: 48 entity + 17 integration + 9 smoke

**Warehouse Integration** ( COMPLETE - January 7, 2026):

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

**Status**:  Week 1 COMPLETE - HTTP Layer Operational (January 9, 2026) | Week 2-3 IN PROGRESS

**Strategic Vision**: Transform Promenade into low-code enterprise platform where business users can customize logic and forms without Go recompilation.

### Architecture Overview

```

                    Business User Layer                        

  Visual LUA Builder    Form Designer    Template Library   

         LUA Scripting Engine            UI Metadata System  

                    Core DDD Contexts                          
  Identity  Customer  Order  Billing  Warehouse  ...     

              Event Bus  PostgreSQL  Redis                   

```

### LUA Scripting Engine

**Location**: `pkg/scripting/`

**Purpose**: Embedded LUA engine for dynamic business logic without Go recompilation

**Components**:
- `engine.go` (300+ lines) - LUA VM wrapper with context support
- `sandbox.go` (150+ lines) - Security restrictions (memory 50MB, timeout 5s, no filesystem/network)
- `stdlib.go` (200+ lines) - Standard library with Promenade APIs

**Standard Library API** (Real UseCase Integration):
```lua
-- Customer operations (calls CustomerUseCase.GetCustomer/UpgradeCustomerTier)
local tier = Customer.GetTier("customer-uuid")
Customer.SetTier("customer-uuid", "pro")
local status = Customer.GetStatus("customer-uuid")
-- Returns: "free", "basic", "pro", "enterprise" (tier)
-- Returns: "lead", "prospect", "customer", "churned" (status)

-- Order operations (calls OrderUseCase.GetOrder)
local status = Order.GetStatus("order-uuid")
local totalCents = Order.GetTotal("order-uuid")
-- Returns: "pending", "confirmed", "processing", "fulfilled", "cancelled"
-- Returns: int64 (e.g., 125000 for $1,250.00)

-- Deal operations (calls DealUseCase.MarkDealAsWon/MarkDealAsLost)
Deal.Approve("deal-uuid")
Deal.Reject("deal-uuid", "Budget constraints")
local stage = Deal.GetStage("deal-uuid")
-- Returns: "lead", "qualified", "proposal", "negotiation", "closed_won", "closed_lost"

-- Notifications (planned)
Notify.SendEmail("user@example.com", "welcome_email")
Notify.SendSMS("+1234567890", "Order shipped")

-- Safe queries (SELECT only, direct SQL execution via *sqlx.DB)
local results = Query.Execute("SELECT name, email FROM customers WHERE tier = 'premium' LIMIT 10")
-- Security: Only SELECT queries, keyword blacklist (DELETE, UPDATE, INSERT, etc.)

-- Date utilities
local now = Date.Now()           -- Returns: "2026-01-08T12:00:00Z"
local formatted = Date.Format(now, "2006-01-02")  -- Returns: "2026-01-08"
local month = Date.GetMonth()    -- Returns: 1 (January)
```

**Backend Flow Examples**:
- `Customer.GetTier(id)` → StandardLibrary.Customer.GetTier(L) → customerUC.GetCustomer(ctx, uuid) → returns cust.Tier
- `Order.GetTotal(id)` → StandardLibrary.Order.GetTotal(L) → orderUC.GetOrder(ctx, uuid) → returns ord.Total.Amount (Money value object)
- `Deal.Approve(id)` → StandardLibrary.Deal.Approve(L) → dealUC.MarkDealAsWon(ctx, uuid, "Approved via LUA") → executes real business logic
- `Query.Execute(sql)` → StandardLibrary.Query.Execute(L) → db.QueryContext(ctx, sql) → converts rows to LUA table

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

**Testing**: 33 tests covering (21 unit + 12 smoke, 100% pass rate):
- Simple execution (2 + 2 = 4)
- Parameters passing (x + y)
- Function execution (greet("Alice"))
- Timeout protection (infinite loop detection)
- Sandbox validation (dangerous functions blocked)
- Standard library (Customer, Order, Deal, Query, Date APIs with real UseCase integration)
- HTTP handlers (Create, List, GetByID, Update, Delete, Execute, Validate)
- Response format validation (`{"status":"success","data":{...}}`)

**HTTP Layer** (10 REST Endpoints at `/api/v1/scripts/*`):
- POST /api/v1/scripts - Create script
- GET /api/v1/scripts - List scripts
- GET /api/v1/scripts/:id - Get script by ID
- PUT /api/v1/scripts/:id - Update script
- DELETE /api/v1/scripts/:id - Delete script
- POST /api/v1/scripts/:id/execute - Execute script
- POST /api/v1/scripts/validate - Validate script syntax
- GET /api/v1/scripts/:id/versions - List script versions (planned)
- POST /api/v1/scripts/:id/versions/:version/restore - Restore version (planned)

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
- Week 1: LUA Engine + HTTP Layer ( COMPLETE - January 8, 2026)
  - pkg/scripting/: Engine, Sandbox, Standard Library with real UseCase integration
  - internal/contexts/scripting/: Script aggregate, HTTP handlers, Router
  - 10 REST endpoints operational, 33 tests passing (21 unit + 12 smoke)
- Week 2: Script Storage + FormDefinition aggregate (IN PROGRESS)
  - Script versioning and database persistence
  - Migration for scripting_scripts and scripting_script_versions tables
  - FormDefinition aggregate initialization
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

### Error Handling (Three-Layer Architecture)

**GOLD STANDARD** (Warehouse context - Production ready, see `docs/guides/unified-error-handling-standard.md`)

Promenade uses a **three-layer error architecture** ensuring security, consistency, and maintainability:

#### Layer 1: errors.go (Domain Constants)

**Location**: `internal/contexts/{context}/{aggregate}/errors.go`

All domain errors defined as constants in three categories:

```go
package location

import "errors"

// Repository Errors - Data access failures
var (
    // ErrLocationNotFound is returned when location doesn't exist
    ErrLocationNotFound = errors.New("location not found")
)

// Business Logic Errors - Domain rule violations
var (
    // ErrLocationCodeExists is returned when code already exists
    ErrLocationCodeExists = errors.New("location code already exists")
    
    // ErrLocationAlreadyDeleted is returned when operating on deleted location
    ErrLocationAlreadyDeleted = errors.New("location already deleted")
    
    // ErrLocationHasChildren is returned when deleting location with children
    ErrLocationHasChildren = errors.New("cannot delete location with children")
)

// Technical Operation Errors - Operation wrappers
var (
    // ErrLocationCreateFailed is returned when creation fails
    ErrLocationCreateFailed = errors.New("failed to create location")
    
    // ErrLocationUpdateFailed is returned when update fails
    ErrLocationUpdateFailed = errors.New("failed to update location")
)
```

#### Layer 2: usecase.go (Zero Inline Errors)

**CRITICAL RULES**:
- ❌ **NEVER** use `fmt.Errorf()` - eliminated completely
- ❌ **NEVER** use inline `errors.New()` - all errors must be constants
- ✅ **ALWAYS** return domain constants from errors.go

```go
// ✅ CORRECT - Domain constants only
func (uc *useCase) CreateLocation(ctx context.Context, code, name string, ...) (*Location, error) {
    // Check for duplicate code
    existing, _ := uc.repo.GetLocationByCode(ctx, code)
    if existing != nil {
        return nil, ErrLocationCodeExists  // Domain constant
    }
    
    // Validate parent if provided
    if parentID != nil {
        parent, err := uc.repo.GetLocation(ctx, *parentID)
        if err != nil {
            return nil, ErrParentLocationNotFound  // Domain constant
        }
        if parent.DeletedAt != nil {
            return nil, ErrParentLocationDeleted  // Domain constant
        }
    }
    
    // Create location
    loc := location.NewLocation(code, name, locationType)
    if err := uc.repo.Create(ctx, loc); err != nil {
        return nil, ErrLocationCreateFailed  // Technical wrapper
    }
    
    return loc, nil
}

// ❌ WRONG - Inline errors (Phase 2 eliminates these)
func (uc *useCase) CreateLocationWrong(ctx context.Context, code string) (*Location, error) {
    if code == "" {
        return nil, fmt.Errorf("code is required")  // ❌ NEVER DO THIS
    }
    return nil, errors.New("failed to create")  // ❌ NEVER DO THIS
}
```

#### Layer 3: handler.go (Security-Aware Mapping)

**Handler pattern** (Phase 1: Security + Phase 2: Domain errors):

```go
func (h *LocationHandler) Create(c *gin.Context) {
    var req CreateLocationRequest
    
    // 1. Validation errors - EXPOSE details (user input issues)
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err.Error())  // Safe: validation feedback
        return
    }
    
    // 2. Call use case
    loc, err := h.usecase.CreateLocation(c.Request.Context(), req.Code, req.Name, ...)
    
    // 3. Domain errors - MAP to user-friendly messages with errors.Is()
    if err != nil {
        if errors.Is(err, location.ErrLocationCodeExists) {
            response.BadRequest(c, "Location code already exists")  // User-friendly
            return
        }
        if errors.Is(err, location.ErrParentLocationNotFound) {
            response.NotFound(c, "Parent location not found")  // Clear message
            return
        }
        if errors.Is(err, location.ErrParentLocationDeleted) {
            response.BadRequest(c, "Parent location is deleted")  // Actionable
            return
        }
        
        // 4. System errors - HIDE details (security)
        response.InternalError(c, "Failed to create location")  // Generic fallback
        return
    }
    
    response.Created(c, toLocationResponse(loc))
}

// ❌ WRONG - Pre-refactoring patterns (DO NOT USE)
func (h *LocationHandler) CreateWrong(c *gin.Context) {
    loc, err := h.usecase.CreateLocation(...)
    if err != nil {
        // ❌ Information leakage (Phase 1 security issue)
        response.InternalError(c, err.Error())  // Exposes: "sql: no rows in result set"
        
        // ❌ String comparison anti-pattern (Phase 2 maintainability issue)
        if err.Error() == "code already exists" {  // Fragile
            response.BadRequest(c, "Code exists")
        }
    }
}
```

#### Testing Pattern (errors.Is assertions)

```go
func TestUseCase_CreateLocation_CodeExists(t *testing.T) {
    // ... setup ...
    
    loc, err := uc.CreateLocation(ctx, "EXISTING-CODE", "Test", ...)
    
    // ✅ CORRECT - Type-safe error checking
    assert.Error(t, err)
    assert.True(t, errors.Is(err, location.ErrLocationCodeExists))
    assert.Nil(t, loc)
    
    // ❌ WRONG - String comparison (fragile)
    assert.Equal(t, "location code already exists", err.Error())  // Don't do this
}
```

**See**: `docs/guides/unified-error-handling-standard.md` (1287 lines, master reference)

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

1. **Three-Layer Error Architecture** (CRITICAL - Phase 1 & 2 Standard):
   - ❌ **NEVER** use `fmt.Errorf()` in usecase.go - all errors must be domain constants in errors.go
   - ❌ **NEVER** use string comparison in handlers - use `errors.Is(err, ErrDomainConstant)`
   - ❌ **NEVER** expose system errors to users - map to user-friendly messages in handlers
   - ✅ **Layer 1**: Define all errors as constants in errors.go (3 categories: Repository, Business Logic, Technical)
   - ✅ **Layer 2**: Return domain constants from usecase.go (zero inline errors)
   - ✅ **Layer 3**: Map with errors.Is() in handlers, expose validation/domain, hide system errors
   - 📖 **Master Reference**: `docs/guides/unified-error-handling-standard.md` (1287 lines)
   - 🏆 **Gold Standard**: `internal/contexts/warehouse/location/` (14 constants, zero fmt.Errorf)

2. **BaseAggregate Field Duplication** (FIXED Jan 2026): NEVER duplicate ID, CreatedAt, UpdatedAt in entities - already in BaseAggregate
   -  `type Entity struct { aggregate.BaseAggregate; ID uuid.UUID }` - WRONG
   -  `type Entity struct { aggregate.BaseAggregate }` - CORRECT
   - Always use `entity.Touch()` instead of `entity.UpdatedAt = time.Now()`
   - Always use `entity.GetID()` instead of `entity.ID` in repositories

3. **UUID v4 vs v7**: NEVER `uuid.New()` (v4). Always `pkg/uuidv7.New()` (time-ordered)

4. **Soft Delete**: Always `WHERE deleted_at IS NULL` in SELECT queries

5. **Context Isolation**: Contexts communicate ONLY via Event Bus (no direct imports between contexts)

6. **Context Chain**: Always pass `ctx`. `getExecutor(ctx)` needs it for tx/db selection

7. **Logger**: `logger.FromContext(ctx)` not global logger (preserves request context)

8. **Migration Namespaces**: Migrations run in order: core → shared → identity → customer-mgmt → order-mgmt

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
- [ ] Errors defined in errors.go, returned from usecase, checked in handler with `errors.Is()`

### Error Handling

- [ ] All domain errors defined as constants in `errors.go` (3 categories)
- [ ] ZERO `fmt.Errorf()` in `usecase.go` - use domain constants only
- [ ] ZERO `errors.New()` inline - all errors must be constants
- [ ] Handler uses `errors.Is()` for error checking (NOT string comparison)
- [ ] Validation errors EXPOSE details (safe user feedback)
- [ ] System errors HIDE details (generic messages only)

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
- `fmt.Errorf()` or inline `errors.New()` in usecase.go (use domain constants from errors.go)
- String comparison for errors in handlers (use `errors.Is()`)
- Exposing system errors to users (leak internal implementation details)
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
