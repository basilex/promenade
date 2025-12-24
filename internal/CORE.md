# Promenade Core

**Core** is the **orchestrator layer** of Promenade. It provides infrastructure services, manages module lifecycle, and coordinates system-wide operations **without knowing implementation details** of business modules.

---

## Core Responsibilities

### 1. **Orchestration** (not Implementation)

Core **manages** but does not **implement** business logic:

```go
//  Core does this (orchestration)
for _, module := range enabledModules {
    if err := module.Initialize(ctx, coreServices); err != nil {
        return err
    }
}

//  Core does NOT do this (implementation)
posts := &PostRepository{}
posts.CreatePost(...)  // This belongs in modules/posts!
```

**Key Principle**: Core knows **WHEN** to call modules, not **HOW** they work.

---

### 2. **Infrastructure Services**

Core provides shared services that modules can use:

#### **Authentication & Authorization**

- User registration, login, JWT tokens
- RBAC system (roles, permissions, wildcards)
- Session management
- Password reset & email verification

**Location**: `internal/domain/entity/` (User, Session), `internal/usecase/` (IAuthUseCase, RBACUseCase)

#### **Database & Transactions**

- PostgreSQL connection management
- Transaction support (`TransactionManager`)
- Base repository pattern for common operations
- UUID v7 generation

**Location**: `internal/infrastructure/database/`

#### **Event IBus**

- Dual-adapter event bus (Memory/Redis)
- Pub/Sub for asynchronous operations
- Health checks and graceful fallback

**Location**: `pkg/bus/`, configured in core, used by modules

#### **Logging**

- Structured logging with `slog`
- Context-aware loggers (request_id, user_id)
- JSON/text format support

**Location**: `pkg/logger/`

#### **Configuration**

- YAML-based configuration
- Environment-specific configs (dev, test, prod)
- IModule configuration loading

**Location**: `internal/infrastructure/config/`

#### **Scheduler**

- Cron-based job scheduler
- Used for automated purge system
- Graceful shutdown support

**Location**: `internal/infrastructure/scheduler/`

---

### 3. **IModule Management**

Core provides the **IModule Registry** system:

```go
// IModule interface (defined by core)
type IModule interface {
    Name() string
    Initialize(ctx context.Context, core Core) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}

// Core interface (provided to modules)
type Core interface {
    DB() *sqlx.DB
    Logger() *logger.Logger
    EventBus() bus.IBus
    Config() *config.Config
    // ... other services
}
```

**Lifecycle Flow:**

1. **Registration**: Modules register via `init()` → `module.DefaultRegistry.Register()`
2. **Initialization**: Core calls `Initialize()` with services
3. **Startup**: Core calls `Start()` for background workers
4. **Shutdown**: Core calls `Stop()` for graceful cleanup

**Location**: `pkg/module/`, `cmd/api/main.go`

---

### 4. **Migration Management**

Core runs migrations with namespace awareness:

```go
migrationManager := migration.NewManager(db, "migrations")

// Always migrate core first
migrationManager.MigrateNamespace(ctx, "core")

// Then migrate enabled modules
for _, moduleName := range enabledModules {
    migrationManager.MigrateNamespace(ctx, moduleName)
}
```

**Core migrations** (`migrations/core/`):

- Schema dependencies (UUID v7 function, extensions)
- Authentication & authorization tables
- RBAC tables (roles, permissions, user_roles)
- Reference data (timezones, languages, countries, currencies, regions, cities, payment methods)

**IModule migrations** (`migrations/{module}/`):

- IModule-specific tables (e.g., `user_posts`, `post_comments`)
- Independent versioning per namespace

**Location**: `pkg/migration/`, `cmd/migrate/`, `migrations/core/`

---

### 5. **Reference Data** (Справочники)

Core provides **shared reference tables** used by multiple modules:

#### **Countries** (`core_countries` table)

- 145 countries worldwide
- ISO 3166-1 codes (alpha-2, alpha-3), phone codes, regions
- Used by: profiles (user location), warehouse (shipping), regions/cities
- **Endpoint**: GET /api/v1/countries

#### **Currencies** (`core_currencies` table)

- 124 currencies with ISO 4217 codes, symbols
- 146 country-currency relationships (including Eurozone, dual-currency countries)
- Used by: warehouse (pricing), payments, international transactions
- **Endpoint**: GET /api/v1/currencies

#### **Regions** (`core_regions` table)

- 30 administrative regions (states, oblasts, provinces, Länder)
- USA states, Russia oblasts, Ukraine oblasts, UK countries, German Länder, etc.
- Foreign key: region → country
- **Endpoints**: GET /api/v1/regions, GET /api/v1/countries/:id/regions

#### **Cities** (`core_cities` table)

- 17 major cities with coordinates, population, capital flags
- Foreign keys: city → region (nullable), city → country
- Special queries: capitals, search by name
- **Endpoints**: GET /api/v1/cities, GET /api/v1/cities/capitals, GET /api/v1/cities/search

#### **Payment Methods** (`core_payment_methods` table)

- 40+ payment methods: cards, digital wallets, crypto, BNPL
- Used by: warehouse (checkout), payments, e-commerce
- **Endpoint**: GET /api/v1/payment-methods

#### **Timezones** (`core_timezones` table)

- IANA timezone names, UTC offsets
- Used by: profiles (user timezone), scheduling, events
- **Endpoint**: GET /api/v1/timezones

#### **Languages** (`core_languages` table)

- ISO 639 codes, native names, RTL flag
- Used by: profiles (user language), i18n, localization
- **Endpoint**: GET /api/v1/languages

**Migrations**:

- `migrations/core/000004_core_ref_timezones.up.sql`
- `migrations/core/000005_core_ref_languages.up.sql`
- `migrations/core/000006_core_ref_countries_currencies.up.sql` (145 countries, 124 currencies)
- `migrations/core/000007_core_ref_regions_cities.up.sql` (30 regions, 17 cities)
- `migrations/core/000008_core_ref_payment_methods.up.sql` (40+ methods)

**Location**: `internal/domain/entity/` (entity definitions), `internal/usecase/` (business logic)

---

### 6. **Purge Orchestration**

Core **schedules** purge jobs but modules **define** what to purge:

```go
// Core: Scheduler runs periodically
scheduler.Schedule("0 2 * * *", func() {
    // Core gets policies from registry
    policies := purge.DefaultPolicyRegistry.GetAllPolicies()

    // Core calls handlers for each policy
    for _, policy := range policies {
        handler := purge.DefaultRegistry.GetHandler(policy.EntityName)
        handler.Purge(ctx, policy)
    }
})

// IModule: Registers policy
purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
})
```

**Core never knows** what `user_posts` is or how to purge it - it just calls the registered handler.

**Location**: `pkg/purge/`, `internal/infrastructure/scheduler/`

---

## 🚫 Core Does NOT Do

### Business Logic

Core does not implement domain-specific business rules:

```go
//  WRONG - Core should not know about posts
func (c *Core) CreatePost(title, content string) error {
    // This belongs in modules/posts!
}

//  CORRECT - Core provides service
func (c *Core) DB() *sqlx.DB {
    return c.db
}
```

### IModule-Specific Entities

Core does not define entities like `Post`, `Comment`, `Product`:

```go
//  WRONG - in internal/domain/entity/
type Post struct {
    ID      string
    Title   string
    Content string
}

//  CORRECT - in internal/modules/posts/domain/entity/
package entity

type Post struct { ... }
```

### Import IModule Code

Core **never** imports from `internal/modules/*`:

```go
//  WRONG
import "github.com/basilex/promenade/internal/modules/posts"

//  CORRECT
import "github.com/basilex/promenade/pkg/module"  // Interface only
```

---

## Core Structure

```
internal/
├── domain/                    # Core domain (auth, RBAC entities)
│   ├── entity/
│   │   ├── user.go            # User entity
│   │   ├── session.go         # Session entity
│   │   ├── role.go            # RBAC role
│   │   ├── permission.go      # RBAC permission
│   │   └── reference/         # Reference data entities
│   │       ├── country.go
│   │       ├── currency.go
│   │       ├── timezone.go
│   │       └── language.go
│   ├── event/                 # Domain events
│   │   ├── user_events.go
│   │   └── auth_events.go
│   └── repository/            # Repository interfaces
│       ├── user_repository.go
│       ├── session_repository.go
│       └── rbac_repository.go
│
├── usecase/                   # Core use cases
│   ├── auth_usecase.go        # Authentication logic
│   ├── rbac_usecase.go        # Authorization logic
│   └── user_usecase.go        # User management
│
├── adapter/                   # Adapters
│   ├── http/                  # HTTP handlers
│   │   ├── v1/
│   │   │   ├── handler/       # Core API handlers
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── user_handler.go
│   │   │   │   └── health_handler.go
│   │   │   └── router/        # Route registration
│   │   │       ├── router.go  # Main router (orchestrates)
│   │   │       ├── init_auth.go
│   │   │       └── init_purge.go
│   │   └── v2/                # API v2
│   └── repository/postgres/   # PostgreSQL implementations
│       ├── user_repository.go
│       ├── session_repository.go
│       └── rbac_repository.go
│
└── infrastructure/            # Infrastructure services
    ├── config/                # Configuration loading
    ├── database/              # DB connection & transactions
    ├── scheduler/             # Cron scheduler
    └── notification/          # Email service
```

---

## 🔄 Request Flow Example

### Authentication Flow

1. **HTTP Request** → `internal/adapter/http/v1/handler/auth_handler.go`
2. **Handler** → Binds request, calls use case
3. **Use Case** → `internal/usecase/auth_usecase.go` (business logic)
4. **Repository** → `internal/adapter/repository/postgres/user_repository.go`
5. **Database** → PostgreSQL via sqlx
6. **Event** → Publishes `UserRegisteredEvent` to event bus
7. **Response** → Returns JWT tokens

**Core's role**: Provides DB connection, event bus, JWT manager - but does NOT know about business entities like `Post` or `Product`.

---

## Core Checklist

When working on core, ensure:

- [ ] No imports from `internal/modules/*`
- [ ] No hardcoded business logic (e.g., post validation rules)
- [ ] No direct entity creation (use interfaces/registries)
- [ ] Provides services via `module.Core` interface
- [ ] Uses registries for extensibility (purge, modules)
- [ ] Configuration is environment-aware
- [ ] Logging uses structured context
- [ ] Database operations use transactions when needed
- [ ] Events are published for async operations

---

## 🔗 Related Documentation

- **[README.md](../README.md)** - Architecture overview
- **[docs/ARCHITECTURE_OVERVIEW.md](../docs/ARCHITECTURE_OVERVIEW.md)** - Visual diagrams
- **[docs/ARCHITECTURE_QUICKREF.md](../docs/ARCHITECTURE_QUICKREF.md)** - Quick reference
- **[internal/modules/README.md](modules/README.md)** - IModule system
- **[docs/MODULE_INDEPENDENCE.md](../docs/MODULE_INDEPENDENCE.md)** - IModule autonomy rules
- **[docs/PURGE_ARCHITECTURE.md](../docs/PURGE_ARCHITECTURE.md)** - Purge orchestration
- **[migrations/README.md](../migrations/README.md)** - Migration system

---

**Core = Orchestrator. Modules = Workers.**
