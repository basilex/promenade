# Promenade AI Instructions

Essential guide for AI agents working in Promenade. For detailed documentation, see [README.md](../README.md) and [docs/](../docs/).

## Core Architecture Principles

### 1. Domain-Driven Design (DDD) with Bounded Contexts

**Architecture**: Event-Driven DDD with Clean Architecture layers

- **Bounded Contexts** - Autonomous business domains (Identity, Customer Management, Order Management, Billing)
- **Aggregates** - Business entities with invariants and transactional boundaries
- **Value Objects** - Immutable domain concepts (Email, Phone, Money, Address)
- **Domain Events** - Asynchronous communication via Event Bus (Memory/Redis adapters)
- **No ORM** - Raw SQL with sqlx + BaseRepository pattern
- **UUID v7 Only** - Use `pkg/uuidv7.New()` for all IDs (time-ordered, 2x faster inserts)

### 2. Bounded Contexts Structure

**Directory**: `internal/contexts/{context-name}/`

Each context is autonomous with:

- Own domain model (aggregates, entities, value objects)
- Own database schema (migrations in `migrations/{context-name}/`)
- Own HTTP routes (registered via Router)
- No cross-context imports (communicate via Event Bus)

**Available Contexts**:

- **Shared** (`internal/contexts/shared/`) - Reference data: Country, Currency, Language, Timezone (read-only)
- **Identity** (`internal/contexts/identity/`) - User, Contact, Profile aggregates (authentication, contacts, profiles)
- **Customer Management** (`internal/contexts/customer-mgmt/`) - Customer aggregate (lifecycle, B2B, segmentation) | Company, Deal, Interaction planned
- **Order Management** (planned) - Order, OrderItem, Fulfillment
- **Billing** (planned) - Invoice, Payment, Subscription
- **Warehouse** (planned) - Inventory management

**Context isolation**: Contexts communicate ONLY via Event Bus (no direct dependencies)

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

**Repository with BaseRepository**:

```go
// Repository interface in aggregate package
type IRepository interface {
    Create(ctx context.Context, contact *Contact) error
    GetByID(ctx context.Context, id uuidv7.UUID) (*Contact, error)
    Update(ctx context.Context, contact *Contact) error
    Delete(ctx context.Context, id uuidv7.UUID) error
}

// Implementation embeds BaseRepository
type contactRepository struct {
    *BaseRepository
}

func NewContactRepository(db *sqlx.DB) IRepository {
    return &contactRepository{
        BaseRepository: NewBaseRepository(db),
    }
}
```

**Use Case Pattern**:

```go
// Interface defines business operations
type IUseCase interface {
    CreateEmailContact(ctx context.Context, userID uuidv7.UUID, email, label string, isPrimary bool) (*Contact, error)
    GetContact(ctx context.Context, contactID uuidv7.UUID) (*Contact, error)
    VerifyContact(ctx context.Context, contactID uuidv7.UUID) error
}

// Implementation (lowercase struct)
type useCase struct {
    repo IRepository
}

// Constructor: simple NewUseCase (NOT New{Entity}UseCase)
func NewUseCase(repo IRepository) IUseCase {
    return &useCase{repo: repo}
}
```

**Domain Events** (via Event Bus):

```go
// Publish event after aggregate state change
event := bus.NewEvent("contact.verified", map[string]interface{}{
    "contact_id": contact.ID.String(),
    "user_id":    contact.UserID.String(),
    "type":       contact.Type,
})
eventBus.Publish(ctx, event)
```

**Context Propagation**: Always pass `ctx` - carries transaction, logger, request ID, user info

## Essential Workflows

### Make Commands (use `make help` for full list)

**Development**:

- `make dev` - Start Postgres, run migrations, start app
- `make build` - Build binary
- `make lint` / `make fmt` - Code quality checks

**Testing** (three-tier strategy):

- `make test` - All tests with race detector (~40s)
- `make test-unit` - Unit tests only (fast, ~5s)
- `make test-smoke` - Smoke tests (mock-based, ~0.3s)
- `make test-integration` - Integration tests with real DB (~5s, auto-starts test DB)
- `make test-coverage` - HTML coverage report
- Test DB: `make test-db-start` / `make test-db-stop` (auto-managed by test-integration)

**Migrations** (namespace-based per context):

- `make migrate` - Run all migrations (core → shared → identity → customer-mgmt)
- `make migrate-core` - Core migrations (UUID v7 extensions)
- `make migrate-identity` - Identity context migrations
- `make migrate-customer-mgmt` - Customer Management context migrations
- `make migrate-new CONTEXT=identity NAME=xxx` - Create new migration

**Docker**:

- `make docker-up` - Start PostgreSQL (localhost:5432)
- `make docker-down` - Stop PostgreSQL
- `make docker-ps` - Show running containers

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

## Data Patterns

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

## Event Bus & Configuration

**Event Bus** - Dual adapters (factory with graceful fallback):

- **Memory** (`pkg/bus/memory`): Dev/test, in-process, fast
- **Redis** (`pkg/bus/redis`): Prod, distributed, persistent
- Config: `bus.adapter: memory|redis` in `config/app.{env}.yaml`

**Config Loading**:

- Core: `config/app.{env}.yaml` (ENVIRONMENT=dev/test/prod)
- Overrides: Sensitive values via env vars (DB_PASSWORD, JWT_SECRET, REDIS_ADDR)
- Access: `logger.FromContext(ctx)`, `database.GetTx(ctx)`

## Testing Strategy

**Test Organization**: Tests live alongside code (`*_test.go` in same directory) with additional smoke/integration tests in mirror path structure

**Three-Tier Test Strategy**:

1. **Unit Tests** (in-place): Fast feedback, test individual components
   - Location: Same directory as production code (`entity_test.go`, `usecase_test.go`)
   - Run: `make test-unit` (~5s)
2. **Smoke Tests** (`test/smoke/contexts/`): Mock-based handler validation, no DB
   - Location: Mirror path structure (e.g., `test/smoke/contexts/shared/country/handler_test.go`)
   - Run: `make test-smoke` (~0.3s)
3. **Integration Tests** (`test/integration/contexts/`): Full E2E with real database
   - Location: Mirror path structure (e.g., `test/integration/contexts/identity/contact/repository_test.go`)
   - Run: `make test-integration` (~5s, auto-starts test DB)

**Running Tests**:

```bash
make test                      # All tests with race detector (~40s)
make test-unit                 # Unit tests only (~5s)
make test-smoke                # Smoke tests (~0.3s)
make test-integration          # Integration tests with real DB (~5s)
make test-coverage             # HTML coverage report

# Context-specific tests
go test ./internal/contexts/identity/... -v
go test ./test/smoke/contexts/shared/... -v
go test ./test/integration/contexts/identity/... -v

# Package tests
go test ./pkg/bus/... -v
go test ./pkg/uuidv7/... -v
```

**Test Statistics**: 150+ tests across 32 packages, 90%+ average coverage

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

## Event Bus & Configuration

**Event Bus** - Dual adapters (factory with graceful fallback):

- **Memory** (`pkg/bus/memory`): Dev/test, in-process, fast
- **Redis** (`pkg/bus/redis`): Prod, distributed, persistent
- Config: `bus.adapter: memory|redis` in `config/app.{env}.yaml`

**Config Loading**:

- Core: `config/app.{env}.yaml` (ENVIRONMENT=dev/test/prod)
- Overrides: Sensitive values via env vars (DB_PASSWORD, JWT_SECRET, REDIS_ADDR)
- Access: `logger.FromContext(ctx)`, `database.GetTx(ctx)`

## Critical Gotchas

**Top Mistakes**:

1. **UUID v4 vs v7**: NEVER `uuid.New()` (v4). Always `pkg/uuidv7.New()` (time-ordered)
2. **Soft Delete**: Always `WHERE deleted_at IS NULL` in SELECT queries
3. **Context Isolation**: Contexts communicate ONLY via Event Bus (no direct imports between contexts)
4. **Context Chain**: Always pass `ctx`. `getExecutor(ctx)` needs it for tx/db selection
5. **Logger**: `logger.FromContext(ctx)` not global logger (preserves request context)
6. **Migration Namespaces**: Migrations run in order: core → shared → identity → customer-mgmt

**Debugging Quick Reference**:

- **DB Connection**: `make docker-ps` → check Postgres on 5432, test DB on 5433
- **Test Failures**: Start with `make test-unit` (~5s), then `make test-integration` (~5s)
- **Migration Issues**: Check `schema_migrations` table, verify namespace (`core/`, `shared/`, `identity/`)
- **Context Won't Load**: Verify router imported and registered in `cmd/api/main.go`
- **Event Bus Issues**: Check adapter config (`memory` for dev, `redis` for prod) in `config/app.{env}.yaml`

## Code Review Checklist

**Before submitting PR, verify every file**:

### Naming Consistency

- [ ] All interfaces: `I{Entity}UseCase`, `I{Entity}Repository` (full words, no abbreviations)
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

**Red Flags** (автоматично reject):

- Interface name `UserPostUC` (use `IUserPostUseCase`)
- Public struct `type UserPostUseCase struct` (must be lowercase `useCase`)
- Constructor `NewCustomerUseCase()` or `NewContactUseCase()` (use simple `NewUseCase()`)
- Repository method `FindByID` (use `GetByID`)
- `uuid.New()` instead of `uuidv7.New()`
- Handler calls repository directly (must go through usecase)
- Business logic in handler or repository (must be in usecase)
- Missing `ctx context.Context` as first parameter
- Directory named `entities/` or `handlers/` (must be singular)

## Key Files

| File                                                           | Purpose                           |
| -------------------------------------------------------------- | --------------------------------- |
| [cmd/api/main.go](../cmd/api/main.go)                          | Entry point, context registration |
| [Makefile](../Makefile) + [Makefile.\*.mk](../Makefile.dev.mk) | All workflows (modular)           |
| [pkg/uuidv7/uuidv7.go](../pkg/uuidv7/uuidv7.go)                | Time-ordered UUIDs                |
| [pkg/bus/README.md](../pkg/bus/README.md)                      | Event Bus documentation           |
| [internal/infrastructure/database/transaction.go][tx]          | Transaction management            |
| [internal/contexts/shared/router.go][shared]                   | Shared context router             |
| [internal/contexts/identity/router.go][identity]               | Identity context router           |
| [docs/](../docs/)                                              | Architecture guides (20+ docs)    |
| [test/README.md](../test/README.md)                            | Testing guide                     |

[tx]: ../internal/infrastructure/database/transaction.go
[shared]: ../internal/contexts/shared/router.go
[identity]: ../internal/contexts/identity/router.go

---

**For comprehensive documentation**: [README.md](../README.md) | [docs/INDEX.md](../docs/INDEX.md) | [docs/CLEAN_ARCHITECTURE_SUMMARY.md](../docs/CLEAN_ARCHITECTURE_SUMMARY.md)
