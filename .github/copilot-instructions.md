# Promenade AI Instructions

Essential guide for AI agents working in Promenade. For detailed documentation, see [README.md](../README.md) and [docs/](../docs/).

## Core Architecture Principles

### 1. Clean Architecture with Module Independence

**Four Layers**: Domain → Use Case → Adapter → Infrastructure

- **Strict Dependency Rule**: Inner layers never depend on outer layers
- **No ORM**: Raw SQL with sqlx + BaseRepository pattern
- **UUID v7 Only**: Use `pkg/uuidv7.New()` for all IDs, never `uuid.New()` (v4) or auto-increment

### 2. Core vs Modules Pattern

**Core** (`internal/domain`, `internal/usecase`) - Always enabled:

- Auth, RBAC, event bus, database, logging
- Reference data: countries, currencies, regions, cities, payment methods, timezones, languages
- Registries: modules, purge, permissions

**Modules** (`internal/modules/*`) - Optional vertical slices:

- **CRITICAL**: Modules MUST NOT import `internal/domain|usecase|adapter`. Only `pkg/*` allowed.
- Self-contained: entity → repo → usecase → handler → routes
- Auto-register via `init()` in `register.go`, import in [cmd/api/main.go](../cmd/api/main.go)
- Examples:
  - `posts` (posts+comments+likes) - Free
  - `profiles` (user profiles+contacts) - Free
  - `analytics` (metrics+reports+dashboards) - Free
  - `notifications` (preferences+history+multi-channel) - Commercial (requires license)
  - `audit` (immutable audit logs+signatures) - Commercial (requires license)
  - `billing` (subscriptions+invoices+payments) - Commercial (requires license), production-ready
  - `workflows` (state machine+BPMN-inspired workflow engine) - Commercial (requires license), production-ready
  - `warehouse` (inventory management) - Commercial (requires license), structure exists but not implemented

**Core orchestrates, modules execute**. Core knows WHEN to call modules, not HOW they work.

### 3. Key Patterns

**BaseRepository**: Core and each module have own BaseRepository (duplication maintains independence)

```go
//Every repo embeds BaseRepository
type UserRepository struct {
    *BaseRepository
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
    var user entity.User
    err := r.Get(ctx, &user, `SELECT * FROM users WHERE id = $1`, id)
    return &user, err
}
```

**Transactions**: Use `TransactionManager.WithTransaction(ctx, func(ctx) error)`. Repos auto-select tx/db via `getExecutor(ctx)`.

**Soft Delete**: Tables with `deleted_at` MUST filter `WHERE deleted_at IS NULL` in all SELECT queries.

**Event Bus**: Dual adapters (Memory for dev, Redis for prod). Config: `bus.adapter: memory|redis` in `config/app.{env}.yaml`.

**Context Propagation**: Always pass `ctx` - carries transaction, logger, request ID, user info. Use `logger.FromContext(ctx)`, never global logger.

## Essential Workflows

### Make Commands (use `make help` for full list)

**Development**:

- `make dev` - Start Postgres, run migrations, start app
- `make build` - Build binary (auto-runs swagger generation)
- `make lint` / `make fmt` - Code quality checks

**Testing** (tests live alongside code in `*_test.go`):

- `make test` - All tests (unit + integration + smoke)
- `make test-unit` - Unit only (~5s)
- `make test-integration` - Integration with real DB port 5433 (~36s)
- `make test-coverage` - HTML coverage report
- Test DB: `make test-db-start` / `make test-db-stop`

**Migrations** (namespace-based: `migrations/{namespace}/NNNNNN_*.sql`):

- `make migrate-up` / `make migrate-down`
- `make migrate-create MODULE=posts NAME=xxx` - New module migration
- `make migrate-create-core NAME=xxx` - New core migration

**API**:

- `make swagger-all` - Generate v1 + v2 API docs (run after handler/DTO changes)

## API & HTTP Patterns

**Versioning**: v1 and v2 APIs isolated (handlers, DTOs, routers, Swagger docs)

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
func (h *PostHandler) Create(c *gin.Context) {
    var req CreatePostDTO
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    post, err := h.usecase.CreatePost(c.Request.Context(), req.Title, req.Content)
    if errors.Is(err, domain.ErrPostNotFound) {
        response.Error(c, http.StatusNotFound, "POST_NOT_FOUND", err.Error())
        return
    }
    response.Success(c, post)
}
```

**Middleware**: Recovery → Request ID → Logger → CORS. Per-route: `RequireAuth()`, `RequirePermission("posts:create")`

## Data Patterns

**Repository pattern** (embed BaseRepository):

```go
type PostRepository struct {
    *BaseRepository  //Provides Get, Select, Exec, NamedExec, getExecutor
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*entity.Post, error) {
    var post entity.Post
    query := `SELECT * FROM user_posts WHERE id = $1 AND deleted_at IS NULL`
    return &post, r.Get(ctx, &post, query, id)
}
```

**Transactions** (context-aware):

```go
err := tm.WithTransaction(ctx, func(ctx context.Context) error {
    //All repo calls use same tx via getExecutor(ctx)
    if err := userRepo.Create(ctx, user); err != nil {
        return err  //Auto-rollback
    }
    return postRepo.Create(ctx, post)  //Auto-commit if no error
})
```

**Critical Rules**:

- UUID v7: `pkg/uuidv7.New()` ONLY (never `uuid.New()`)
- Soft delete: Always add `WHERE deleted_at IS NULL`
- Context: Always pass `ctx` for tx/logger propagation
- Logger: `logger.FromContext(ctx)` not global logger

## Adding a New Module

1. **Structure** in `internal/modules/mymodule/`:

   ```
   ├── module.go              # Implement pkg/module.Module interface
   ├── register.go            # init() auto-registration
   ├── config/                # Own YAML configs
   ├── domain/entity/         # Business entities
   ├── domain/repository/     # Repo interfaces
   ├── usecase/               # Business logic
   └── adapter/
       ├── http/handler/      # HTTP handlers
       └── repository/postgres/ # DB implementation + BaseRepository
   ```

2. **module.go** - Implement `pkg/module.Module` interface:

   ```go
   package mymodule

   import "github.com/basilex/promenade/pkg/module"

   type MyModule struct {
       core *module.Core
       // ... handlers, usecases, config
   }

   func New() module.Module {
       return &MyModule{}
   }

   // Implement required methods:
   // - Metadata() module.Metadata
   // - Dependencies() []string
   // - Initialize(ctx, core) error
   // - RegisterRoutes(router)
   // - RegisterMigrations() []module.Migration
   // - RegisterEventHandlers(bus) error
   // - RegisterPermissions() []module.Permission
   // - Start(ctx) error
   // - Stop(ctx) error
   // - HealthCheck(ctx) error
   ```

3. **register.go** template:

   ```go
   package mymodule

   import (
       "log/slog"
       "github.com/basilex/promenade/pkg/module"
   )

   func init() {
       if err := module.DefaultRegistry.Register(New()); err != nil {
           slog.Error("Failed to register mymodule", "error", err)
       }
   }
   ```

4. **Import** in `cmd/api/main.go`: `_ "github.com/basilex/promenade/internal/modules/mymodule"`
5. **Enable** in `config/modules.yaml`:
   ```yaml
   modules:
     enabled:
       - posts
       - profiles
       - analytics # Free
       - audit # Commercial (requires license)
       - billing # Commercial (requires license)
       - workflows # Commercial (requires license)
       - mymodule # Add here
   ```
6. **Migrations**: `make migrate-create MODULE=mymodule NAME=init`

**CRITICAL**: Never import `internal/domain`, `internal/usecase`, or `internal/adapter` in modules. Only `pkg/*` allowed.

## Event Bus & Configuration

**Event Bus** - Dual adapters (factory with graceful fallback):

- **Memory** (`pkg/bus/memory`): Dev/test, in-process, fast
- **Redis** (`pkg/bus/redis`): Prod, distributed, persistent
- Config: `bus.adapter: memory|redis` in `config/app.{env}.yaml`

**Config Loading**:

- Core: `config/app.{env}.yaml` (ENVIRONMENT=dev/test/prod)
- Modules: `internal/modules/{name}/config/config.{env}.yaml`
- Overrides: Sensitive values via env vars (DB_PASSWORD, JWT_SECRET)
- Access: `logger.FromContext(ctx)`, `database.GetTx(ctx)`

## Testing Strategy

**Test Organization**: Tests live alongside code (`*_test.go` in same directory)

**Test Types**:

- **Core Entity Tests** (39 tests) - Domain entities: Country, Currency, Language, Timezone, Permission, Role, User, Session, Purge
- **Core UseCase Tests** (236 tests) - Business logic: Auth (Register, Login, Logout, RefreshToken, GetMe, Sessions), RBAC (CountryUseCase, CurrencyUseCase, LanguageUseCase, PermissionUseCase, PurgeUseCase, RoleUseCase, TimezoneUseCase)
- **Module Entity Tests**:
  - Posts module (33 tests, 83.3% coverage) - PostStatus, UserPost lifecycle, validation, slug generation
  - Profiles module (21 tests, 80.4% coverage) - UserContact, UserProfile, privacy, validation
  - Analytics module (11 tests) - Metric validation, MetricAggregate, usecase operations
  - Notifications module (48 tests) - 35 unit (20 entity + 15 usecase) + 13 integration tests - Multi-channel delivery, quiet hours, preferences
  - Audit module (tests planned) - AuditEvent entity, signature verification, immutable logging
  - Billing module (375 tests, 100% coverage) - Plan, Subscription, Invoice, Payment entities + all use cases
  - Workflows module (183 tests, 100% passing) - 55 entity + 59 usecase + 51 repository + 18 handler - State machine, graph validation, workflow lifecycle
- **Utilities Tests** (51 tests, 89.5% avg) - response (100%), validator (80%), logger (83.8%), pagination (94.1%)

**Test Execution** (~25 seconds total):

```bash
make test                      # All tests (680+ tests)
make test-core                 # Core tests (275 tests: 39 entity + 236 usecase)
make test-modules              # All module tests
make test-module-posts         # Posts module (33 tests)
make test-module-profiles      # Profiles module (21 tests)
make test-module-analytics     # Analytics module (11 tests)
make test-module-notifications # Notifications module (48 tests: 35 unit + 13 integration)
make test-module-billing       # Billing module (375 tests)
make test-module-workflows     # Workflows module (183 tests: 55 entity + 59 usecase + 51 repo + 18 handler)
make test-coverage             # HTML coverage report
```

**Example**:

```go
func TestPostUsecase_Create(t *testing.T) {
    mockRepo := mocks.NewMockPostRepository(t)
    usecase := NewPostUsecase(mockRepo, nil)

    mockRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil)

    post, err := usecase.CreatePost(context.Background(), "Title", "Content")
    assert.NoError(t, err)
    assert.NotEmpty(t, post.ID)
}
```

**Helpers**: [test/integration/](../test/integration/) for DB setup, fixtures

## Codebase Conventions & Patterns

### Naming Conventions

**DTOs** (Data Transfer Objects):

- Request DTOs: `Create{Entity}Request`, `Update{Entity}Request`
- Response DTOs: `{Entity}Response`
- Nested DTOs: `FeaturedImageDTO`, `MetadataDTO`
- Use pointers for optional fields in Update DTOs: `Title *string`

**Entities**:

- Business entities in `domain/entity/` (e.g., `UserPost`, `Comment`, `UserProfile`)
- Use `uuidv7.UUID` for all IDs
- Include struct tags: `db` for sqlx, `json` for API, `validate` for rules
- Example: `Title string \`db:"title" json:"title" validate:"required,min=3,max=255"\``

**Errors**:

- Define domain errors as variables in usecase layer: `var ErrProfileNotFound = errors.New("profile not found")`
- Common patterns: `Err{Entity}NotFound`, `Err{Entity}AlreadyExists`, `ErrUnauthorized{Action}`, `ErrInvalid{Field}`
- Never use `errors.New()` in handlers - check with `errors.Is(err, usecase.ErrXxx)`
- Wrap errors with context: `fmt.Errorf("failed to create user: %w", err)`

**Repositories**:

- Interface in `domain/repository/`, implementation in `adapter/repository/postgres/`
- Method names: `GetByID`, `GetByUserID`, `GetBySlug`, `Create`, `Update`, `Delete`, `List{Entity}s`
- Always return `(*Entity, error)` for single, `([]*Entity, error)` for multiple
- Queries use positional parameters: `$1`, `$2`, etc.

**Use Cases**:

- Interface first, then implementation (lowercase struct)
- Method signatures: `(ctx context.Context, params...) (result, error)`
- Always check entity ownership: `if post.UserID != requestUserID { return ErrUnauthorized }`
- Business logic here, not in handlers or entities

### Validation Patterns

**Entity Validation** (in domain layer):

```go
func (p *UserPost) Validate() error {
    if err := p.ValidateTitle(); err != nil {
        return err
    }
    if err := p.ValidateSlug(); err != nil {
        return err
    }
    return nil
}

func (p *UserPost) ValidateTitle() error {
    title := strings.TrimSpace(p.Title)
    if title == "" {
        return fmt.Errorf("title is required")
    }
    if utf8.RuneCountInString(title) < 3 {
        return fmt.Errorf("title must be at least 3 characters")
    }
    return nil
}
```

**Request Validation** (in handler layer):

- Use Gin binding tags: `binding:"required,min=3,max=200"`
- Validate in handler: `if err := c.ShouldBindJSON(&req); err != nil`
- Common validators: `required`, `email`, `min`, `max`, `oneof`, `omitempty`, `dive` (for arrays)

**Business Rules Validation** (in usecase layer):

- Check uniqueness constraints before creating
- Validate ownership before updating/deleting
- Apply business logic (e.g., can't publish archived post)

### Error Handling Flow

**Handler → UseCase → Repository**:

```go
//Handler: Check error type and return appropriate HTTP code
post, err := h.postUC.GetPost(ctx, postID)
if err != nil {
    if err == entity.ErrNotFound {
        response.Error(c, http.StatusNotFound, "post not found", err)
        return
    }
    response.Error(c, http.StatusInternalServerError, "failed to get post", err)
    return
}

//UseCase: Define domain errors and add context
if existingPost != nil {
    return nil, ErrSlugAlreadyExists
}
if err := uc.postRepo.Create(ctx, post); err != nil {
    return nil, fmt.Errorf("failed to create post: %w", err)
}

//Repository: Return raw errors from database
if err := r.Get(ctx, &post, query, id); err != nil {
    if err == sql.ErrNoRows {
        return nil, entity.ErrNotFound
    }
    return nil, err
}
```

### Entity State Methods

**Status Management** (entities control their own state):

```go
//Prefer methods over direct field assignment
func (p *UserPost) Publish() {
    p.Status = PostStatusPublished
    now := time.Now()
    p.PublishedAt = &now
    p.ScheduledAt = nil
}

func (p *UserPost) Archive() {
    p.Status = PostStatusArchived
}

func (u *User) Suspend(reason string, until *time.Time) {
    u.Status = UserStatusSuspended
    u.SuspendedReason = &reason
    u.SuspendedUntil = until
}
```

### DTO Conversion Pattern

**Always in adapter layer** (`adapter/http/dto/`):

```go
//Entity → Response DTO
func ToPostResponse(post *entity.UserPost) PostResponse {
    return PostResponse{
        ID:       post.ID.String(),
        UserID:   post.UserID.String(),
        Title:    post.Title,
        //... map all fields
    }
}

//Request DTO → Entity (in usecase, not DTO layer)
post, err := entity.NewUserPost(userID, req.Title, slug, req.Content)
```

### Database Field Naming

- Snake_case: `user_id`, `created_at`, `is_public`, `featured_image`
- JSONB columns: `featured_image`, `tags`, `categories`, `meta_keywords`
- Soft delete: `deleted_at TIMESTAMP NULL`
- Timestamps: Always `created_at` and `updated_at`, use `DEFAULT CURRENT_TIMESTAMP`

### Response Helpers

**Standard responses** (`pkg/response`):

```go
//Success with data
response.Success(c, http.StatusOK, data)
response.Success(c, http.StatusCreated, data)

//Errors with message
response.Error(c, http.StatusBadRequest, "invalid request", err)
response.Error(c, http.StatusNotFound, "post not found", err)
response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
```

**Pagination helpers**:

```go
page := response.GetPageFromQuery(c)        //Default: 1
pageSize := response.GetPageSizeFromQuery(c) //Default: 20, max: 100
```

## Commercial Modules (Optional)

**License System** (signature-based HMAC-SHA256):

- Format: `PROMENADE-{MODULE}-{TIER}-{EXPIRY}-{SIGNATURE}`
- Tiers: BASIC, PRO, ENTERPRISE (different features/retention)
- Generation: `./scripts/generate-license.sh <module> <tier> <days>`
- Config: Set `{MODULE}_LICENSE_KEY` env var or `license_required: false` for dev
- Examples:
  - `analytics` module (metrics, reports, dashboards) - **FREE** (no license required)
  - `notifications` module (preferences, history, multi-channel delivery, rate limiting) - **COMMERCIAL** (requires license)
  - `audit` module (immutable audit logs, cryptographic signatures, tamper detection) - **COMMERCIAL** (requires license)
  - `billing` module (subscriptions, invoices, payments) - **COMMERCIAL** (requires license)
  - `workflows` module (state machine, BPMN-inspired workflow engine) - **COMMERCIAL** (requires license)
  - `warehouse` module (inventory management) - **COMMERCIAL** (requires license, not yet implemented)

## Code Consistency Standards

### Professional Naming Rules (Єдині Стандарти)

**CRITICAL**: Every name in this project follows strict patterns. No exceptions, no variations.

**Interface Names** (завжди з великої літери):

```go
//CORRECT - PascalCase with "I" prefix and "UseCase" suffix
type IUserPostUseCase interface { }
type IAuthUseCase interface { }
type IRoleUseCase interface { }

//CORRECT - PascalCase with "I" prefix and "Repository" suffix (never "Repo")
type IUserPostRepository interface { }
type IUserRepository interface { }
type ICountryRepository interface { }

//WRONG - inconsistent casing or abbreviations
type userPostUsecase interface { }  //lowercase
type UserPostUC interface { }       //abbreviated (missing I prefix)
type UserRepo interface { }         //"Repo" instead of "Repository" (missing I prefix)
```

**Implementation Names** (lowercase private):

```go
//CORRECT - lowercase struct, PascalCase constructor returning I-prefixed interface
type userPostUseCase struct { }
func NewUserPostUseCase(...) IUserPostUseCase { return &userPostUseCase{} }

type authUseCase struct { }
func NewAuthUseCase(...) IAuthUseCase { return &authUseCase{} }

//WRONG - inconsistent patterns
type UserPostUseCase struct { }  //Public struct (should be lowercase)
type userpostUseCase struct { }  //missing camelCase
type IUserPostUseCase struct { } //Interface name used for struct
```

**Handler Names** (always "{Entity}Handler"):

```go
//CORRECT - Entity name + "Handler", fields use I-prefixed interfaces
type UserPostHandler struct { postUC usecase.IUserPostUseCase }
type UserProfileHandler struct { profileUC usecase.IUserProfileUseCase }
type AuthHandler struct { authUC usecase.IAuthUseCase }

//WRONG - inconsistent suffixes
type PostsHandler struct { }    //plural
type UserPostHdl struct { }     //abbreviated
```

**Method Names** (consistent prefixes):

```go
// Repository methods - always start with Get/Create/Update/Delete/List
func (r *UserRepository) GetByID(ctx, id) (*User, error)
func (r *UserRepository) GetByEmail(ctx, email) (*User, error)
func (r *UserRepository) Create(ctx, user) error
func (r *UserRepository) Update(ctx, user) error
func (r *UserRepository) Delete(ctx, id) error
func (r *UserRepository) ListUsers(ctx, limit, offset) ([]*User, error)

// UseCase methods - business operations
func (uc *authUseCase) Login(ctx, email, password) (string, error)
func (uc *authUseCase) Register(ctx, email, name) (*User, error)
func (uc *userPostUseCase) PublishPost(ctx, userID, postID) error

//WRONG - inconsistent prefixes
func (r *UserRepository) FindByID(ctx, id) (*User, error)  //Use "GetByID"
func (r *UserRepository) FetchUsers(ctx) ([]*User, error)  //Use "ListUsers"
func (r *UserRepository) Remove(ctx, id) error             //Use "Delete"
```

### File & Directory Naming

**Directory structure** (завжди однакова):

```
internal/modules/{module}/
├── module.go              # MUST be named "module.go"
├── register.go            # MUST be named "register.go"
├── config/
│   ├── config.dev.yaml   # MUST be "config.{env}.yaml"
│   └── config.test.yaml
├── domain/
│   ├── entity/           # MUST be "entity" (not "entities")
│   │   ├── user_post.go  # snake_case file names
│   │   └── comment.go
│   └── repository/       # MUST be "repository" (not "repositories")
│       ├── user_post_repository.go
│       └── comment_repository.go
├── usecase/              # MUST be "usecase" (not "usecases")
│   ├── post_usecase.go
│   └── comment_usecase.go
└── adapter/
    ├── http/
    │   ├── handler/      # MUST be "handler" (not "handlers")
    │   │   ├── post_handler.go
    │   │   └── comment_handler.go
    │   └── dto/          # MUST be "dto" (not "dtos")
    │       ├── post_dto.go
    │       └── comment_dto.go
    └── repository/postgres/
        ├── base_repository.go
        ├── user_post_repository.go
        └── comment_repository.go
```

**File naming**:

- Entities: `{entity_name}.go` → `user_post.go`, `comment.go`, `user_profile.go`
- Repositories: `{entity_name}_repository.go` → `user_post_repository.go`
- Use cases: `{entity_name}_usecase.go` → `post_usecase.go`, `auth_usecase.go`
- Handlers: `{entity_name}_handler.go` → `post_handler.go`
- DTOs: `{entity_name}_dto.go` → `post_dto.go`
- Tests: `{file_name}_test.go` → `post_usecase_test.go`

### Variable & Parameter Naming

**Context** (завжди перший параметр):

```go
//CORRECT - ctx is always first parameter
func (uc *postUseCase) CreatePost(ctx context.Context, userID uuid.UUID, title string) (*Post, error)
func (r *PostRepository) GetByID(ctx context.Context, id uuid.UUID) (*Post, error)
func (h *PostHandler) Create(c *gin.Context)

//WRONG - ctx not first
func CreatePost(userID uuid.UUID, ctx context.Context, title string) (*Post, error)
```

**Standard abbreviations** (завжди однакові):

```go
uc  - usecase      //e.g., postUC := NewPostUseCase(repo)
ctx - context      //e.g., ctx context.Context
req - request      //e.g., var req CreatePostRequest
err - error        //e.g., if err != nil { }
tx  - transaction  //e.g., tx, err := db.BeginTx(ctx)
db  - database     //e.g., db *sqlx.DB
c   - gin.Context  //e.g., func (h *Handler) Get(c *gin.Context)
```

**Receiver names** (consistent):

```go
//CORRECT - consistent abbreviations
func (r *UserRepository) GetByID(ctx, id) (*User, error)    //r = repository
func (uc *authUseCase) Login(ctx, email) (string, error)   //uc = usecase
func (h *PostHandler) Create(c *gin.Context)               //h = handler
func (u *User) Validate() error                            //u = entity instance
func (p *UserPost) Publish()                               //p = entity instance

//WRONG - inconsistent receivers
func (repo *UserRepository) GetByID(...)    //Use "r"
func (usecase *authUseCase) Login(...)      //Use "uc"
func (handler *PostHandler) Create(...)     //Use "h"
```

### Error Messages (Однакові формулювання)

**Error variable names**:

```go
//CORRECT - Err{Entity}{Condition}
var (
    ErrUserNotFound              = errors.New("user not found")
    ErrPostNotFound              = errors.New("post not found")
    ErrProfileAlreadyExists      = errors.New("profile already exists")
    ErrSlugAlreadyExists         = errors.New("slug already exists")
    ErrUnauthorizedAccess        = errors.New("unauthorized access")
    ErrInvalidInput              = errors.New("invalid input")
)

//WRONG - inconsistent naming
var (
    UserNotFoundError = errors.New(...)  //suffix instead of prefix
    ErrNoUser = errors.New(...)          //too abbreviated
    NotFound = errors.New(...)           //missing entity
)
```

**Error wrapping format**:

```go
//CORRECT - descriptive context + %w
return nil, fmt.Errorf("failed to create user: %w", err)
return nil, fmt.Errorf("failed to get post by ID: %w", err)
return nil, fmt.Errorf("failed to update profile: %w", err)

//WRONG - inconsistent format
return nil, fmt.Errorf("error: %w", err)        //not descriptive
return nil, fmt.Errorf("cannot create: %v", err) //%v instead of %w
return nil, errors.Wrap(err, "failed")          //use fmt.Errorf
```

### Import Order (Завжди однаковий)

```go
import (
    //1. Standard library (alphabetical)
    "context"
    "errors"
    "fmt"
    "time"

    //2. External packages (alphabetical)
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    //3. Internal packages (alphabetical by path depth)
    "github.com/basilex/promenade/internal/domain/entity"
    "github.com/basilex/promenade/internal/domain/repository"
    "github.com/basilex/promenade/pkg/response"
    "github.com/basilex/promenade/pkg/uuidv7"
)
```

### Onboarding Checklist (Для нових розробників)

**Before writing code, verify**:

- [ ] All interfaces start with `I` and end with `UseCase` or `Repository` (never abbreviated)
- [ ] All private structs are lowercase: `userPostUseCase`, `authUseCase`
- [ ] All constructors return I-prefixed interface: `func New...() IInterfaceName`
- [ ] All repository methods start with: `Get`, `Create`, `Update`, `Delete`, `List`
- [ ] All handlers have `Handler` suffix: `UserPostHandler`, `AuthHandler`
- [ ] All errors start with `Err`: `ErrUserNotFound`, `ErrSlugExists`
- [ ] Context is always first parameter: `(ctx context.Context, ...)`
- [ ] Files follow snake_case: `user_post.go`, `post_repository.go`
- [ ] Directories are singular: `entity/`, `repository/`, `handler/`, `usecase/`
- [ ] Module structure matches template exactly (no creative variations)

**Quick self-check questions**:

1. Would a new developer understand this without asking?
2. Does this name match existing patterns exactly?
3. Is this the same as how it's done in other modules?
4. Can I find 3 similar examples in the codebase?

**If unsure**: Search codebase for similar code and copy the pattern exactly.

## Critical Gotchas

**Top Mistakes**:

1. **UUID v4 vs v7**: NEVER `uuid.New()` (v4). Always `pkg/uuidv7.New()` (time-ordered)
2. **Soft Delete**: Always `WHERE deleted_at IS NULL` in SELECT queries
3. **Module Dependencies**: No `internal/domain|usecase|adapter` imports in modules. Only `pkg/*`
4. **Context Chain**: Always pass `ctx`. `getExecutor(ctx)` needs it for tx/db selection
5. **Logger**: `logger.FromContext(ctx)` not global logger (preserves request context)
6. **Migration Namespaces**: Core migrations run first. Wrong namespace breaks history

**Debugging Quick Reference**:

- **DB Connection**: `make docker-ps` → check Postgres on 5432, test DB on 5433
- **Test Failures**: Start with `make test-unit` (~5s), then `make test-integration` (~36s)
- **Migration Issues**: Check `schema_migrations` table, verify namespace (`core/`, `posts/`, etc.)
- **Module Won't Load**: Verify import in `cmd/api/main.go` + enabled in `config/modules.yaml`
- **API 404s**: Run `make swagger-all` after handler changes
- **License Errors**: Check `{MODULE}_LICENSE_KEY` env var or set `license_required: false`

## Code Review Checklist

**Before submitting PR, verify every file**:

### Naming Consistency

- [ ] All interfaces: `I{Entity}UseCase`, `I{Entity}Repository` (full words, no abbreviations)
- [ ] All structs: lowercase `{entity}UseCase`, `{entity}Repository`
- [ ] All constructors: `New{Entity}UseCase() I{Entity}UseCase`
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
- Public struct `type UserPostUseCase struct` (must be lowercase)
- Repository method `FindByID` (use `GetByID`)
- `uuid.New()` instead of `uuidv7.New()`
- Handler calls repository directly (must go through usecase)
- Business logic in handler or repository (must be in usecase)
- Missing `ctx context.Context` as first parameter
- Directory named `entities/` or `handlers/` (must be singular)

## Key Files

| File                                                                | Purpose                        |
| ------------------------------------------------------------------- | ------------------------------ |
| [cmd/api/main.go](../cmd/api/main.go)                               | Entry point, module loading    |
| [Makefile](../Makefile) + [Makefile.\*.mk](../Makefile.dev.mk)      | All workflows (modular)        |
| [pkg/module/module.go](../pkg/module/module.go)                     | Module interface & registry    |
| [pkg/uuidv7/uuidv7.go](../pkg/uuidv7/uuidv7.go)                     | Time-ordered UUIDs             |
| [internal/infrastructure/database/transaction.go][tx]               | Transaction management         |
| [internal/adapter/repository/postgres/base_repository.go][baserepo] | Core BaseRepository            |
| [scripts/generate-license.sh](../scripts/generate-license.sh)       | Commercial license generation  |
| [docs/](../docs/)                                                   | Architecture guides (20+ docs) |
| [test/integration/](../test/integration/)                           | Test utilities & fixtures      |

[tx]: ../internal/infrastructure/database/transaction.go
[baserepo]: ../internal/adapter/repository/postgres/base_repository.go

---

**For comprehensive documentation**: [README.md](../README.md) | [docs/INDEX.md](../docs/INDEX.md) | [docs/ARCHITECTURE_QUICKREF.md](../docs/ARCHITECTURE_QUICKREF.md)
