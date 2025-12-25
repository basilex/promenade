# Promenade Architecture Overview

This document provides a high-level overview of the Promenade architecture, organized around **Clean Architecture principles** and a **plugin-based module system**.

---

## Architecture Layers

### 1. Application Layer

**Location:** `cmd/api/main.go`

**Responsibilities:**

- Initialize infrastructure (DB, EventBus, Config, Logger)
- Load core configuration
- Initialize module system (discovery, registration, lifecycle)
- Start HTTP server

---

### 2. Core Infrastructure

**Location:** `internal/infrastructure/`

**Components:**

| Component     | Purpose                          |
| ------------- | -------------------------------- |
| Database      | PostgreSQL connection management |
| Event IBus     | Memory/Redis Pub/Sub adapters    |
| Scheduler     | Cron-based job scheduling        |
| Config        | YAML configuration loader        |
| Logger        | Structured logging with slog     |
| Email Service | Asynchronous email sending       |

**Provides:** Shared services to all modules (DB, EventBus, Config, JWT, Logger)

---

### 3. Core Domain

**Location:** `internal/domain/entity/`

**Security & Authentication** (always enabled):

- User (authentication only: email, password, roles)
- Session (JWT tokens)
- Role (RBAC roles)
- Permission (resource:action format)

**Reference Data** (stable, shared):

- Country (195+ ISO codes, regions)
- Currency (170+ ISO 4217 codes)
- Timezone (500+ IANA timezones)
- Language (180+ ISO 639 codes)

**Purpose:** Core entities required by ALL modules. No business logic.

---

### 4. Core Use Cases

**Location:** `internal/usecase/`

**Available Use Cases:**

| Use Case           | Purpose                          |
| ------------------ | -------------------------------- |
| Auth UseCase       | Register, login, logout          |
| Role UseCase       | CRUD roles, role assignment      |
| Permission UseCase | CRUD permissions, access control |
| Country UseCase    | List countries, get by code      |
| Currency UseCase   | List currencies, get by code     |
| Purge UseCase      | Orchestrate purge jobs           |

**Note:** Core use cases do NOT contain business logic - only infrastructure and security.

---

### 5. Core API Routes

**Location:** `internal/adapter/http/v1/`

**Endpoints:**

| Route                  | Purpose                  |
| ---------------------- | ------------------------ |
| /api/v1/health         | Health checks            |
| /api/v1/auth/\*        | Login, register, logout  |
| /api/v1/countries/\*   | Reference data           |
| /api/v1/currencies/\*  | Reference data           |
| /api/v1/roles/\*       | RBAC management          |
| /api/v1/permissions/\* | RBAC management          |
| /api/v1/admin/\*       | Purge, system management |

---

## Module System

### Module Registry ### IModule Registry & Manager Manager

**Location:** `pkg/module/`

**Purpose:** Orchestrates module lifecycle

**Registry Features:**

- `Register(module)` - Auto-register via `init()`
- `GetEnabled(config)` - Filter enabled modules from config
- `InitializeAll()` - Initialize in dependency order
- `StartAll()` - Start background workers
- `StopAll()` - Graceful shutdown

**Provides to Modules:**

- Shared DB connection
- Shared EventBus
- Shared JWT manager
- Shared Config loader

---

### Module: Posts

**Location:** `internal/modules/posts/`

**Status:** Enabled (Free)

**Entities:** Post, Comment, Like

**Features:**

- Create, update, delete posts
- Threaded comments (configurable max depth)
- Like posts and comments
- Soft delete with configurable retention

**Configuration:**

- max_content_length: 10000
- comments.max_depth: 10
- purge.user_posts.retention_days: 90
- purge.post_comments.retention_days: 30

**Routes:** /api/v1/posts/_, /api/v1/comments/_, /api/v1/likes/\*

---

### Module: Profiles

**Location:** `internal/modules/profiles/`

**Status:** Enabled (Free)

**Entities:** Profile, Contact

**Features:**

- User profile management
- Contact information (email, phone, etc.)
- Contact verification
- Primary contact designation

**Configuration:**

- profiles.max_per_user: 1
- contacts.max_per_user: 5
- contacts.verification_required: true

**Routes:** /api/v1/profiles/_, /api/v1/contacts/_

---

### Module: Warehouse

**Location:** `internal/modules/warehouse/`

**Status:** Disabled (Commercial - requires license)

**Features** (when licensed):

- Inventory management
- Stock tracking
- Barcode scanning
- Warehouse locations

**Routes:** /api/v1/warehouse/\* (when enabled)

---

## Inter-Module Communication

### Event IBus

**Location:** `pkg/bus/`

**Adapters:**

| Adapter | Use Case                  | Features                |
| ------- | ------------------------- | ----------------------- |
| Memory  | Dev/Test/Single-instance  | Fast, zero dependencies |
| Redis   | Production/Multi-instance | Distributed, persistent |

**Event Flow:**

1. Module A publishes event to EventBus
2. EventBus fans out to all subscribers
3. Modules B, C, D process event asynchronously

**Examples:**

- user.registered → send welcome email (async)
- post.created → update user stats (async)
- purge.completed → log to audit (async)

---

## Configuration Architecture

### Core Configuration

```
config/
├── app.dev.yaml     - Core infrastructure (dev)
├── app.test.yaml    - Core infrastructure (test)
├── app.prod.yaml    - Core infrastructure (prod)
└── modules.yaml     - Which modules to load
```

### Module Configuration

```
internal/modules/posts/config/
├── config.dev.yaml  - Posts module settings (dev)
├── config.test.yaml - Posts module settings (test)
└── config.prod.yaml - Posts module settings (prod)

internal/modules/profiles/config/
├── config.dev.yaml  - Profiles module settings (dev)
├── config.test.yaml - Profiles module settings (test)
└── config.prod.yaml - Profiles module settings (prod)
```

**Environment Overrides:** `.env.example` (optional)

---

## Module Lifecycle

### 1. Auto-Registration (via init())

Module registers itself on package import:

```go
package posts

func init() {
    module.DefaultRegistry.Register(New())
}
```

### 2. Discovery & Filtering

- Read `config/modules.yaml`
- Filter enabled modules
- Resolve dependencies (topological sort)

### 3. Initialization (in dependency order)

For each module:

- Load module config from `config/config.*.yaml`
- Setup repositories
- Setup use cases
- Setup handlers
- Register purge handlers
- Register retention policies
- Register permissions

### 4. Route Registration

For each module:

- Mount module routes to router
- Apply middleware (auth, RBAC, etc.)

### 5. Event Subscription

For each module:

- Subscribe to relevant events
- Setup async event handlers

### 6. Start (background workers)

For each module:

- Start cron jobs
- Start background workers

### 7. Runtime

- Modules process HTTP requests
- Publish/subscribe to events
- Execute scheduled tasks

### 8. Shutdown (on SIGTERM/SIGINT)

For each module (reverse order):

- Stop workers gracefully
- Close connections
- Cleanup resources

---

## Key Design Principles

### 1. CORE = INFRASTRUCTURE + REFERENCE DATA

- Core provides services (DB, EventBus, Config, JWT)
- Core contains stable reference data (countries, currencies)
- Core manages security (auth, RBAC)
- Core does NOT contain business logic

### 2. MODULES = BUSINESS LOGIC

- Modules are self-contained vertical slices
- Modules own their entities, use cases, adapters
- Modules register handlers, policies, permissions
- Modules can be enabled/disabled via config
- Modules do NOT import from `internal/domain` or `internal/usecase`
- Modules do NOT depend on each other directly (use events)

### 3. PLUGIN ARCHITECTURE

- Dynamic loading via module registry
- Dependency resolution (topological sort)
- Lifecycle management (init → start → stop)
- Auto-registration via `init()`

### 4. CONFIGURATION AUTONOMY

- Core: `config/app.*.yaml` (infrastructure only)
- Modules: `internal/modules/{name}/config/config.*.yaml`
- Each module loads its own config
- Environment-specific configs (dev, test, prod)

### 5. LOOSE COUPLING

- Event-driven communication (pub/sub)
- Registry pattern (handlers, policies, permissions)
- Interface-based dependencies
- No direct module-to-module imports

### 6. LICENSING SUPPORT

- Per-module license keys
- License validation in `Initialize()`
- Graceful degradation if license invalid

---

## Benefits

### Modularity

- Add new modules without touching core
- Remove modules without breaking others
- Test modules independently

### Scalability

- Commercial modules (warehouse, fleet, finance)
- License-based feature enablement
- Easy to add new verticals

### Maintainability

- Clear boundaries (core vs modules)
- Single responsibility (each module owns its domain)
- Configuration clarity (no monolithic config)

### Testability

- Unit test modules in isolation
- Integration test with real/mock core
- Smoke test critical flows

### Deployability

- Enable only needed modules per deployment
- A/B testing of new modules
- Gradual rollout of features

---

## Current Status

### Core

- Infrastructure (DB, EventBus, Scheduler, Config, Logger)
- Security (Auth, RBAC, JWT, Sessions)
- Reference Data (Countries, Currencies, Timezones, Languages)
- Module Management (Registry, Lifecycle, Config Loader)
- Purge Orchestration (Registry-based, no entity knowledge)

### Modules

- **Posts** (posts + comments + likes) - Fully independent
- **Profiles** (profiles + contacts) - Fully independent
- **Warehouse** (inventory management) - Commercial, disabled

### Architecture Compliance

- Core contains ONLY infrastructure + reference data
- Modules are FULLY independent (no core imports)
- Configuration is AUTONOMOUS (each module owns config)
- Licensing is SUPPORTED (ready for commercial modules)

### Recent Improvements

- Purge system refactored (core = orchestrator, modules = workers)
- Profiles/Contacts migrated to module (removed from core)
- 15,000+ lines of code removed from core
- Complete module independence achieved

---

## Related Documentation

- [ARCHITECTURE_AUDIT.md](ARCHITECTURE_AUDIT.md) - Architecture compliance audit
- [ARCHITECTURE_QUICKREF.md](ARCHITECTURE_QUICKREF.md) - Quick reference guide
- [MODULE_DEVELOPMENT.md](MODULE_DEVELOPMENT.md) - Creating new modules
- [MODULE_INDEPENDENCE.md](MODULE_INDEPENDENCE.md) - Module independence principles
- [../internal/CORE.md](../internal/CORE.md) - Core components details
