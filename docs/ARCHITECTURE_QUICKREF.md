# Promenade Architecture - Quick Reference

## Core vs Modules: Simple Rule

**CORE** = Infrastructure + Reference Data + Auth

- Always enabled, provides services

**MODULES** = Business Logic

- Optional, licensable, independent

---

## What Goes in Core?

### BELONGS in Core:

1. **Infrastructure Services**

   - Database connection management
   - Event bus (Memory/Redis adapters)
   - Scheduler (cron jobs)
   - Configuration loader
   - Logger
   - Email service
   - JWT manager

2. **Security Foundation**

   - User authentication (login, register, password)
   - RBAC (roles, permissions, access control)
   - Sessions (JWT tokens)

3. **Reference Data**

   - Countries (145 countries, ISO 3166-1 codes, regions)
   - Currencies (124 currencies, ISO 4217, symbols)
   - Regions (30 admin regions: states, oblasts, provinces, Länder)
   - Cities (17 major cities with coordinates, population, capitals)
   - Payment Methods (40+ methods: cards, wallets, crypto, BNPL)
   - Timezones (IANA timezone database)
   - Languages (ISO 639 codes)
   - _Stable, rarely-changing data shared across modules_

4. **Management Interfaces**
   - IModule registry
   - Purge handler registry
   - Purge policy registry
   - Event bus interface

### DOES NOT Belong in Core:

- Business entities (Post, Comment, Profile, etc.)
- Business use cases
- Business HTTP handlers
- Business routes
- Business configuration
- Entity-specific logic

**Rule of thumb:** If it's a business concept that could be sold separately, it's a MODULE.

---

## What Goes in Modules?

### IModule Structure

```
internal/modules/mymodule/
├── module.go              # IModule implementation
├── register.go            # Auto-registration via init()
├── config/                # Own YAML configs per environment
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── entity/                # Domain entities
├── usecase/               # Business logic
└── adapter/
    ├── http/              # Handlers, DTOs, routes
    ├── repository/        # Postgres implementations
    └── purge/             # Purge handlers (if needed)
```

### IModule Checklist

- [ ] Has own entity/usecase/adapter structure
- [ ] Loads own config from `config/config.*.yaml`
- [ ] Registers routes in `RegisterRoutes()`
- [ ] Registers permissions in `RegisterPermissions()`
- [ ] Registers purge handlers (if soft-delete entities)
- [ ] No imports from `internal/domain` or `internal/usecase`
- [ ] Uses only `pkg/*` packages

---

## IModule Development Workflow

### 1. Create IModule

```bash
mkdir -p internal/modules/mymodule/{config,entity,usecase,adapter/http/handler}
```

### 2. Implement IModule Interface

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type MyModule struct {
    *module.BaseModule
    db *sqlx.DB
    // ... other fields
}

func New() module.IModule {
    return &MyModule{
        BaseModule: module.NewBaseModule(module.Metadata{
            Name:        "mymodule",
            DisplayName: "My IModule",
            Version:     "1.0.0",
            Description: "Does something useful",
        }),
    }
}

func (m *MyModule) Initialize(ctx context.Context, core *module.Core) error {
    // 1. Load module config
    cfg := moduleconfig.Load("internal/modules/mymodule/config", os.Getenv("ENVIRONMENT"))

    // 2. Setup repositories, use cases, handlers
    m.db = core.DB

    // 3. Register purge handlers (if needed)
    // 4. Register retention policies (if needed)

    return nil
}

func (m *MyModule) RegisterRoutes(router *gin.RouterGroup) {
    group := router.Group("/mymodule")
    {
        group.GET("", m.handler.List)
        group.POST("", m.handler.Create)
    }
}

func (m *MyModule) RegisterPermissions() []module.Permission {
    return []module.Permission{
        {Resource: "mymodule", Action: "read", Description: "View items"},
        {Resource: "mymodule", Action: "create", Description: "Create items"},
    }
}

// ... implement other interface methods
```

### 3. Auto-Register

```go
// internal/modules/mymodule/register.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 4. Add Configuration

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true
  version: "1.0.0"

mymodule:
  max_items: 100
  allow_public: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

### 5. Enable in Main Config

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - mymodule # Add here
```

### 6. Import in main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/mymodule"  // Add here
)
```

---

## Configuration Rules

### Core Configuration

**File:** `config/app.{dev|test|prod}.yaml`

**Contains ONLY:**

- Infrastructure settings (DB, server, JWT, logging)
- Event bus configuration
- Purge infrastructure (enabled, schedule, batch_size)
- CORS settings
- Email service settings

**Does NOT contain:**

- Entity-specific retention days → Modules
- IModule-specific settings → Modules
- Business logic configuration → Modules

### Module Configuration

**File:** `internal/modules/{name}/config/config.{dev|test|prod}.yaml`

**Contains:**

- IModule metadata (name, version)
- IModule-specific settings
- Purge retention policies (if applicable)
- Feature flags (if applicable)

**Loaded by:** Each module via `pkg/module/config.Load()`

---

## Purge System Rules

### OLD WAY (Core knows entities)

```go
//  WRONG - Core has entity-specific config
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

### NEW WAY (Core orchestrates only)

**Core Config:**

```yaml
purge:
  enabled: true
  schedule: "0 2 * * *"
  batch_size: 1000
```

**IModule Config:**

```yaml
purge:
  user_posts:
    retention_days: 90
    enabled: true
```

**IModule Registration:**

```go
// Module registers handler
handler := purge.NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)

// Module registers policy
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

**Core Orchestration:**

```go
// Core gets ALL policies from registry
policies := purge.DefaultPolicyRegistry.GetAllPolicies()

// Core creates use case
useCase := usecase.NewPurgeUseCase(
    purge.DefaultRegistry,  // handlers
    policies,               // from modules
    batchSize,
    eventBus,
)

// Core starts scheduler
scheduler.Start(ctx)
```

**Result:** Core knows nothing about `user_posts` or retention days!

---

## Communication Patterns

### Event-Driven (Preferred)

```go
// IModule A publishes
event := &PostCreatedEvent{PostID: id}
eventBus.Publish(ctx, "post.created", event)

// IModule B subscribes
eventBus.Subscribe("post.created", func(e bus.Event) {
    // Handle async
})
```

**Benefits:** Loose coupling, async processing

### Direct Registry (Use Sparingly)

```go
// Get another module
postsModule := core.Registry.Get("posts")

// Type assert and call
if api, ok := postsModule.(PostsAPI); ok {
    stats := api.GetStats(userID)
}
```

**Use only when:** Sync response needed, can't use events

---

## Common Mistakes to Avoid

### Importing Core Packages in Modules

```go
//  WRONG
import "github.com/basilex/promenade/internal/domain/entity"
import "github.com/basilex/promenade/internal/usecase"
```

**Fix:** Define entities in module's own `entity/` package.

### Hardcoding Business Values in Code

```go
//  WRONG
const maxCommentLength = 2000
```

**Fix:** Load from module config.

### Putting Business Logic in Core

```go
//  WRONG - PostUseCase in internal/usecase/
```

**Fix:** Move to module's `usecase/` package.

### Core Knowing About IModule Entities

```go
//  WRONG - Core has retention days for posts
type PurgeConfig struct {
    RetentionDaysUserPosts int
}
```

**Fix:** Module registers retention policy via registry.

---

## Testing Strategy

### Core Tests

- Unit test infrastructure services
- Integration test auth/RBAC
- Test reference data repositories

### Module Tests

- Unit test business logic (use cases)
- Integration test repositories
- Test handlers with mock use cases

### Smoke Tests

- End-to-end critical flows
- Test inter-module communication via events
- Verify module routes work

---

## Quick Commands

```bash
# Build
make build

# Run dev
make dev

# Run tests
make test

# Run specific module tests
go test ./internal/modules/posts/...

# Run smoke tests
make test-smoke

# Generate module boilerplate
make generate ENTITY=MyEntity

# Create migration
make migrate-create NAME=add_my_table
```

---

## Decision Tree: Core or IModule?

```
Is it infrastructure (DB, logger, event bus)?
└─> YES → CORE

Is it security (auth, RBAC)?
└─> YES → CORE

Is it reference data (countries, currencies, regions, cities, payment methods)?
└─> YES → CORE

Is it stable and used by multiple modules?
└─> YES → Consider CORE (or shared pkg)

Is it business logic?
└─> YES → MODULE

Can it be sold separately?
└─> YES → MODULE

Is it entity-specific?
└─> YES → MODULE

When in doubt?
└─> MODULE (easier to move to core later than vice versa)
```

---

## Resources

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Detailed architecture review
- [ARCHITECTURE_OVERVIEW.md](ARCHITECTURE_OVERVIEW.md) - Visual architecture
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - IModule development guide
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Independence principles
- [PURGE_ARCHITECTURE.md](PURGE_ARCHITECTURE.md) - Purge system details
- [internal/CORE.md](../internal/CORE.md) - Core components documentation

---

**Remember:**

- Core = Infrastructure + Reference Data + Auth
- Modules = Business Logic (independent, licensable)
- Use registries for loose coupling
- Events for async communication
- Configuration autonomy for each module
