# Promenade

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Production-ready REST API** built with **Clean Architecture**, **modular plugin system**, and **namespace-based database migrations**. Features autonomous business modules, PostgreSQL with UUID v7, event-driven architecture, and comprehensive testing infrastructure.

---

## Architecture Overview

Promenade follows a **strict layered architecture** where the **Core orchestrates** and **Modules execute** business logic:

**Core Layer** - Orchestrator + Infrastructure + Shared Services

- Authentication & Authorization (RBAC)
- Event Bus (Memory/Redis)
- Database & Transactions
- Logging & Configuration
- Module Registry & Lifecycle
- Reference Data (countries, currencies, regions, cities, timezones, languages, payment methods)

**Module Layer** - Independent Vertical Slices (Domain Areas)

| Module        | Entities                         | Description                   | Status   |
| ------------- | -------------------------------- | ----------------------------- | -------- |
| Posts         | posts, comments, likes           | User-generated content        | Free     |
| Profiles      | contacts, profiles               | User profiles                 | Free     |
| **Analytics** | **metrics, reports, dashboards** | **Analytics and reporting**   | **Free** |
| Warehouse     | products, inventory              | Inventory management (future) | Planned  |

Each module is self-contained with:

- Own domain entities & business logic
- Own database migrations (namespace-based)
- Own repositories & use cases
- Own HTTP handlers & routes
- Own configuration & lifecycle
- Optional: Purge policies, permissions, events

### Core Principles

1. **Core as Orchestrator**

   - Core **manages** module lifecycle (init, start, stop)
   - Core **provides** shared services (auth, events, DB, logging)
   - Core **knows WHEN** to call modules, but not **HOW** they work
   - Core **never** imports module-specific code

2. **Modules as Workers**

   - Modules **implement** domain-specific business logic
   - Modules **register** themselves via `init()` functions
   - Modules are **independent** - can be enabled/disabled without affecting others
   - Modules **never** import other modules' code (only `pkg/*`)
   - Each module = complete vertical slice (entities → handlers)

3. **Clean Architecture Layers**

   ```
   Domain (entities, interfaces) → Use Case (business logic)
      ↓                                      ↓
   Adapter (repos, handlers) → Infrastructure (DB, HTTP, events)
   ```

   - **Dependency Rule**: Inner layers never depend on outer layers
   - Use cases depend only on domain interfaces, never on concrete implementations

4. **Namespace-Based Migrations**
   - Each namespace (core, posts, profiles) has independent version history
   - Migrations live in `migrations/{namespace}/NNNNNN_description.{up|down}.sql`
   - Core migrations run first, then enabled modules
   - True module autonomy - enable/disable without schema conflicts

---

## Quick Start

### Prerequisites

- **Go 1.25+**
- **Docker & Docker Compose** (for PostgreSQL, Redis)
- **Make** (for automation)

### 1. Clone & Setup

```bash
git clone https://github.com/basilex/promenade.git
cd promenade

# Install development dependencies
make install

# Start PostgreSQL + Redis via Docker
make docker-up
```

### 2. Run Migrations

Migrations run **automatically** on application startup, but you can also run them manually:

```bash
# Check migration status for all namespaces
make migrate-status

# Run all migrations (core + enabled modules)
make migrate

# Run specific namespace
make migrate-core                    # Core only
make migrate-module MODULE=posts     # Specific module
```

See [migrations/README.md](migrations/README.md) for detailed migration guide.

### 3. Start Application

```bash
# Development mode (hot reload, debug logging)
make dev

# Or build and run binary
make build
./bin/promenade
```

Server starts on **http://localhost:8081**

---

## Documentation Structure

### Core Documentation

| Document                                                           | Description                                                     |
| ------------------------------------------------------------------ | --------------------------------------------------------------- |
| **[docs/ARCHITECTURE_OVERVIEW.md](docs/ARCHITECTURE_OVERVIEW.md)** | Visual architecture diagrams, layer responsibilities, lifecycle |
| **[docs/ARCHITECTURE_QUICKREF.md](docs/ARCHITECTURE_QUICKREF.md)** | Quick reference, decision trees, common mistakes                |
| **[docs/ARCHITECTURE_AUDIT.md](docs/ARCHITECTURE_AUDIT.md)**       | Architecture compliance audit, verification checklist           |
| **[internal/CORE.md](internal/CORE.md)**                           | Core responsibilities and boundaries                            |

### Module System

| Document                                                                     | Description                                     |
| ---------------------------------------------------------------------------- | ----------------------------------------------- |
| **[internal/modules/README.md](internal/modules/README.md)**                 | Module system overview, structure, registration |
| **[docs/MODULE_DEVELOPMENT.md](docs/MODULE_DEVELOPMENT.md)**                 | Creating new modules, best practices            |
| **[docs/MODULE_INDEPENDENCE.md](docs/MODULE_INDEPENDENCE.md)**               | Module autonomy rules, dependency management    |
| **[docs/MODULE_CONFIG_ARCHITECTURE.md](docs/MODULE_CONFIG_ARCHITECTURE.md)** | Module configuration system                     |

### Infrastructure & Systems

| Document                                                             | Description                                  |
| -------------------------------------------------------------------- | -------------------------------------------- |
| **[migrations/README.md](migrations/README.md)**                     | Namespace-based migration system, CLI usage  |
| **[docs/MIGRATION_ARCHITECTURE.md](docs/MIGRATION_ARCHITECTURE.md)** | Migration system design and implementation   |
| **[docs/PURGE_ARCHITECTURE.md](docs/PURGE_ARCHITECTURE.md)**         | Automated data purge system (registry-based) |
| **[pkg/bus/README.md](pkg/bus/README.md)**                           | Event bus (Memory/Redis adapters)            |
| **[docs/REDIS_BUS_TESTING.md](docs/REDIS_BUS_TESTING.md)**           | Testing Redis event bus                      |

### Development Guides

| Document                                                             | Description                                |
| -------------------------------------------------------------------- | ------------------------------------------ |
| **[docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)**   | Makefile system (dev, test, prod commands) |
| **[test/README.md](test/README.md)**                                 | Testing infrastructure (200+ tests)        |
| **[docs/TESTING_GUIDE.md](docs/TESTING_GUIDE.md)**                   | Testing best practices, patterns           |
| **[docs/TESTING_INFRASTRUCTURE.md](docs/TESTING_INFRASTRUCTURE.md)** | Test infrastructure setup                  |

### Technical References

| Document                                           | Description                                      |
| -------------------------------------------------- | ------------------------------------------------ |
| **[docs/UUID_V7_GUIDE.md](docs/UUID_V7_GUIDE.md)** | UUID v7 implementation and benefits              |
| **[docs/SOFT_DELETE.md](docs/SOFT_DELETE.md)**     | Soft delete pattern for user content             |
| **[docs/AUTHORIZATION.md](docs/AUTHORIZATION.md)** | RBAC system (4 roles, wildcard permissions)      |
| **[docs/LOGGING.md](docs/LOGGING.md)**             | Structured logging with slog                     |
| **[docs/VALIDATION.md](docs/VALIDATION.md)**       | Request validation patterns                      |
| **[docs/CREDENTIALS.md](docs/CREDENTIALS.md)**     | Default test users and credentials               |
| **[docs/INDEX.md](docs/INDEX.md)**                 | Complete documentation index with learning paths |

---

## Module System

### Available Modules

#### **Posts Module** (`internal/modules/posts`)

User-generated content management:

- **Entities**: Posts, Comments, Likes
- **Features**: Create/edit posts, comment threads, like system
- **Migrations**: 3 migrations (namespace: `posts`)
- **Config**: `config/modules.yaml` → `posts`

**Full documentation**: [internal/modules/posts/README.md](internal/modules/posts/README.md)

#### **Profiles Module** (`internal/modules/profiles`)

User profile and contact management:

- **Entities**: UserProfiles, UserContacts
- **Features**: Profile management, contact information
- **Migrations**: 2 migrations (namespace: `profiles`)
- **Config**: `config/modules.yaml` → `profiles`

**Full documentation**: [internal/modules/profiles/README.md](internal/modules/profiles/README.md)

#### **Analytics Module** (`internal/modules/analytics`) - Free

Analytics, metrics, and reporting:

- **Status**: Free - Available for all users
- **Entities**: Metrics, Reports, Dashboards
- **Features**: Metrics collection, custom reports, visual dashboards
- **Migrations**: 1 migration (namespace: `analytics`)
- **Use Case**: Business intelligence, performance monitoring, data insights

**Full documentation**: [internal/modules/analytics/README.md](internal/modules/analytics/README.md)

#### **Warehouse Module** (`internal/modules/warehouse`) - Future Module

Inventory and product management (planned):

- **Status**: Planned - Structure exists as placeholder, not yet implemented
- **Use Case**: E-commerce, inventory systems, retail

**Planned documentation**: [internal/modules/warehouse/README.md](internal/modules/warehouse/README.md)

### Module Structure

Each module follows a consistent structure:

```
internal/modules/{module}/
├── module.go           # Module registration & lifecycle
├── domain/
│   └── entity/         # Domain entities
├── repository/         # Data access interfaces & implementations
├── usecase/            # Business logic
├── adapter/
│   └── handler/        # HTTP handlers & DTOs
└── README.md           # Module-specific documentation
```

### Enabling/Disabling Modules

Edit `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts # User-generated content
    - profiles # User profiles + contacts
    - analytics # Business analytics (requires license)
    # - warehouse  # Future: Inventory management
```

Modules load automatically on application startup.

---

## Database Migrations

### Namespace-Based System

Each namespace maintains **independent version history**:

```
migrations/
├── core/               # Core infrastructure (always runs first)
│   ├── 000001_core_init_uuid_v7.up.sql
│   ├── 000002_core_auth_full.up.sql
│   ├── 000003_core_rbac_full.up.sql
│   ├── 000004_core_ref_timezones.up.sql
│   ├── 000005_core_ref_languages.up.sql
│   ├── 000006_core_ref_countries_currencies.up.sql    # 145 countries, 124 currencies
│   ├── 000007_core_ref_regions_cities.up.sql          # 30 regions, 17 cities
│   └── 000008_core_ref_payment_methods.up.sql         # 40+ payment methods
├── posts/              # Posts module migrations
│   ├── 000001_posts_posts.up.sql
│   ├── 000002_posts_comments.up.sql
│   └── 000003_posts_comment_likes.up.sql
├── profiles/           # Profiles module migrations
│   ├── 000001_profiles_contacts.up.sql
│   └── 000002_profiles_profiles.up.sql
└── analytics/          # Analytics module migrations (commercial)
    └── 000001_analytics_tables.up.sql
```

### Migration Commands

```bash
# Status for all namespaces
make migrate-status

# Run all (core + enabled modules)
make migrate

# Run specific namespace
make migrate-core
make migrate-module MODULE=posts

# Rollback
make migrate-rollback MODULE=posts STEPS=1

# Create new migration
make migrate-create MODULE=posts NAME=add_post_views
make migrate-create-core NAME=add_audit_log
```

**Auto-migrations**: Migrations run automatically on app startup (core first, then enabled modules).

**Full guide**: [migrations/README.md](migrations/README.md)

---

## Authentication & Authorization

### Default Test Users

| Email                           | Password   | Role      | Permissions             |
| ------------------------------- | ---------- | --------- | ----------------------- |
| `system@promenade.com`          | `passw0rd` | Admin     | Full access (`*`)       |
| `admin@promenade.com`           | `passw0rd` | Admin     | User/content management |
| `moderator@promenade.com`       | `passw0rd` | Moderator | Content moderation      |
| `alexander.vasilenko@gmail.com` | `03041965` | User      | Basic operations        |

**Change passwords before production deployment!**

### RBAC System

- **4 System Roles**: Admin, Moderator, User, Guest
- **Wildcard Permissions**: `posts:*` (all post actions), `*` (full access)
- **Resource-Action Format**: `posts:create`, `users:delete`, `comments:moderate`

**Full RBAC guide**: [docs/AUTHORIZATION.md](docs/AUTHORIZATION.md)

---

## Testing

**388 tests** across all layers (100% passing, ~41 seconds):

```bash
# Run all tests (unit + integration + smoke)
make test               # All tests (~41s)

# Run by type
make test-unit          # Unit tests only (183 tests, ~5s)
make test-integration   # Integration tests (91 tests, ~36s)
make test-smoke         # Smoke tests (114 tests, ~4s)

# Coverage report
make test-coverage
```

### Test Structure

- **Core Tests**: Domain entities (Country, Currency, Language, Timezone, Permission, Role, User, Session, Purge policies)
- **Core Use Cases**: Auth, RBAC, Reference data CRUD, Purge operations
- **Module Tests**: Posts (Comment, Post, PostStatus), Profiles (UserContact, UserProfile, ContactType, Gender), Analytics (Metrics, Reports, License validation)
- **Integration Tests**: Repository operations with real PostgreSQL on port 5433
- **Test Helpers**: `test/helpers/` and `test/integration/` for fixtures, database setup, transaction management

**Testing guides**:

- [test/README.md](test/README.md) - Test infrastructure
- [docs/TESTING_GUIDE.md](docs/TESTING_GUIDE.md) - Best practices

---

## Event Bus

**Dual-adapter event bus** for asynchronous operations:

### Memory Adapter

- In-memory Pub/Sub (goroutines + channels)
- **Use case**: Development, testing, single-instance deployments
- **Pros**: Zero dependencies, fast, simple
- **Config**: `BUS_ADAPTER=memory` (default)

### Redis Adapter

- Distributed Pub/Sub via Redis
- **Use case**: Production multi-instance deployments
- **Pros**: Persistent, scalable, fault-tolerant
- **Config**: `BUS_ADAPTER=redis` + Redis connection settings
- **Fallback**: Auto-falls back to memory if Redis unavailable

### Usage Example

```go
// Publish event
event := &UserRegisteredEvent{
    BaseEvent: bus.BaseEvent{ID: uuid.New().String()},
    UserID:    user.ID,
    Email:     user.Email,
}
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// Subscribe to events
eventBus.Subscribe(ctx, bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    evt := e.(*UserRegisteredEvent)
    // Send welcome email
    return emailService.SendWelcome(ctx, evt.Email)
})
```

**Full guide**: [pkg/bus/README.md](pkg/bus/README.md)

---

## Makefile Commands

### Development

```bash
make dev                # Start dev server (hot reload)
make build              # Build production binary
make run                # Run built binary
make lint               # Run linter (golangci-lint)
make fmt                # Format code
make config-show        # Show YAML configuration (use ENV=dev|test|prod)
```

### Testing

```bash
make test                      # All tests (core + modules)
make test-core                 # Core tests only (domain + usecase)
make test-modules              # All module tests
make test-module-posts         # Posts module tests
make test-module-profiles      # Profiles module tests
make test-coverage             # Generate HTML coverage report
```

### Database

```bash
make migrate                   # Run all migrations (core + enabled modules)
make migrate-status            # Show migration status
make migrate-core              # Migrate core only
make migrate-module MODULE=posts          # Migrate specific module
make migrate-rollback MODULE=posts STEPS=1  # Rollback
make migrate-create MODULE=posts NAME=xxx  # Create module migration
make migrate-create-core NAME=xxx          # Create core migration
```

### Docker

```bash
make docker-up          # Start PostgreSQL + Redis
make docker-down        # Stop services
make docker-clean       # Remove containers + volumes
make docker-logs        # View logs
```

### Swagger

```bash
make swagger-all        # Generate API docs (v1 + v2)
make swagger-v1         # Generate v1 docs only
make swagger-v2         # Generate v2 docs only
```

**Full Makefile guide**: [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)

---

## Docker

### Development Setup

```bash
# Start services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down

# Clean slate (remove volumes)
make docker-clean
```

### Services

- **PostgreSQL 16**: Port 5432, user `system`, database `promenade_dev`
- **Redis 7**: Port 6379 (for distributed event bus)

**Docker guide**: [docker/README.md](docker/README.md)

---

## Key Technical Features

### UUID v7 Primary Keys

Time-ordered UUIDs for **2x faster inserts** than UUID v4 and better B-tree performance.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    ...
);
```

[docs/UUID_V7_GUIDE.md](docs/UUID_V7_GUIDE.md)

### Soft Delete Pattern

User-generated content (posts, comments) uses `deleted_at` timestamp for safe deletion.

```go
// Always filter soft-deleted records
WHERE deleted_at IS NULL
```

[docs/SOFT_DELETE.md](docs/SOFT_DELETE.md)

### Automated Purge System

Registry-based system where modules register their retention policies:

```go
purge.DefaultPolicyRegistry.RegisterPolicy(purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
})
```

Core scheduler runs purge jobs via cron. Core doesn't know about specific entities.

[docs/PURGE_ARCHITECTURE.md](docs/PURGE_ARCHITECTURE.md)

### Structured Logging

Context-aware logging with `slog`:

```go
log := logger.FromContext(ctx)  // Includes request_id, user_id
log.Info("User registered", "email", user.Email)
```

[docs/LOGGING.md](docs/LOGGING.md)

---

## API Documentation

### Swagger UI

- **v1 API**: http://localhost:8081/api/v1/docs/swagger/index.html
- **v2 API**: http://localhost:8081/api/v2/docs/swagger/index.html

### Health Check

```bash
curl http://localhost:8081/api/v1/health
```

Response:

```json
{
  "status": "success",
  "data": {
    "status": "healthy",
    "database": "connected",
    "timestamp": "2025-12-22T16:40:00Z"
  }
}
```

### API Versioning

- **v1**: Current stable API (`internal/adapter/http/v1`)
- **v2**: Next-generation API (`internal/adapter/http/v2`)

Both versions have:

- Isolated handlers and DTOs
- Separate Swagger documentation
- Independent route registration

---

## Project Structure

```
promenade/
├── cmd/
│   ├── api/                    # Main application entry point
│   └── migrate/                # Migration CLI tool
├── internal/
│   ├── domain/                 # Core domain (entities, interfaces)
│   │   ├── entity/             # Domain entities (User, Session)
│   │   ├── event/              # Domain events (UserRegistered, etc.)
│   │   └── repository/         # Repository interfaces
│   ├── usecase/                # Core use cases (auth, RBAC)
│   ├── adapter/                # Adapters (HTTP, repositories)
│   │   ├── http/
│   │   │   ├── v1/             # API v1 (handlers, DTOs, routes)
│   │   │   └── v2/             # API v2
│   │   └── repository/postgres/ # PostgreSQL implementations
│   ├── infrastructure/         # Infrastructure (DB, config, scheduler)
│   │   ├── database/
│   │   ├── config/
│   │   ├── notification/
│   │   └── scheduler/
│   └── modules/                # Business modules (plugins)
│       ├── posts/              # Posts + comments + likes
│       ├── profiles/           # User profiles + contacts
│       ├── analytics/          # Analytics + reports (Commercial, active)
│       └── warehouse/          # Inventory management (future)
├── pkg/                        # Shared packages (reusable)
│   ├── bus/                    # Event bus (memory/redis)
│   ├── jwt/                    # JWT manager
│   ├── logger/                 # Structured logger
│   ├── migration/              # Migration manager
│   ├── module/                 # Module registry
│   ├── purge/                  # Purge registry
│   ├── response/               # HTTP response helpers
│   ├── uuidv7/                 # UUID v7 generator
│   └── validator/              # Request validation
├── migrations/                 # Namespace-based migrations
│   ├── core/                   # Core migrations (auth, RBAC, ref data)
│   ├── posts/                  # Posts module migrations
│   └── profiles/               # Profiles module migrations
├── test/                       # Test infrastructure
│   ├── helpers/                # Test helpers (fixtures, DB setup)
│   ├── integration/            # Integration tests
│   ├── smoke/                  # Smoke tests
│   └── mocks/                  # Mock implementations
├── config/                     # Configuration files
│   ├── app.dev.yaml            # Dev environment config
│   ├── app.test.yaml           # Test environment config
│   ├── app.prod.yaml           # Production config
│   └── modules.yaml            # Module enable/disable + settings
├── docs/                       # Documentation
├── scripts/                    # Helper scripts
├── templates/                  # Email templates
└── docker/                     # Docker configs
```

---

## Configuration

### Environment-Specific Configs

Config priority (YAML first, `.env` fallback):

1. `config/app.{dev|test|prod}.yaml` - Core infrastructure settings
2. `config/modules.yaml` - Module enable/disable + module-specific settings
3. `.env.{env}.local` / `.env.{env}` / `.env` - Legacy support

### Example: `config/app.dev.yaml`

```yaml
app:
  name: "Promenade API"
  environment: "development"
  version: "1.0.0"

server:
  host: "0.0.0.0"
  port: 8081

database:
  host: "localhost"
  port: 5432
  user: "system"
  password: "passw0rd"
  database: "promenade_dev"

jwt:
  secret: "your-super-secret-jwt-key-change-in-production"
  access_token_duration: 15m
  refresh_token_duration: 168h

bus:
  adapter: "memory" # or "redis"
  worker_pool_size: 4

purge:
  enabled: true
  schedule: "0 2 * * *" # Daily at 2 AM
  batch_size: 1000
```

### Example: `config/modules.yaml`

```yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics # Commercial module (requires license)
    # - warehouse  # Future: Inventory management

  config:
    posts:
      version: "1.0.0"
      settings:
        max_post_length: 10000
        max_comment_depth: 10

    analytics:
      version: "1.0.0"
      license_key: "" # Set via ANALYTICS_LICENSE_KEY env var
      settings:
        metrics_retention_days: 90
```

**Config guide**: [docs/MODULE_CONFIG_ARCHITECTURE.md](docs/MODULE_CONFIG_ARCHITECTURE.md)

---

## Creating New Modules

### Step 1: Create Module Structure

```bash
mkdir -p internal/modules/mymodule/{domain/entity,repository,usecase,adapter/handler}
```

### Step 2: Implement Module Interface

```go
// internal/modules/mymodule/module.go
package mymodule

import "github.com/basilex/promenade/pkg/module"

type Module struct{}

func (m *Module) Name() string { return "mymodule" }

func (m *Module) Initialize(ctx context.Context, core module.Core) error {
    // Register routes, permissions, purge handlers
    return nil
}

func (m *Module) Start(ctx context.Context) error {
    // Start background workers
    return nil
}

func (m *Module) Stop(ctx context.Context) error {
    // Graceful shutdown
    return nil
}

func init() {
    module.DefaultRegistry.Register(&Module{})
}
```

### Step 3: Create Migrations

```bash
make migrate-create MODULE=mymodule NAME=create_tables
```

### Step 4: Enable Module

Add to `config/modules.yaml`:

```yaml
modules:
  enabled:
    - mymodule
```

**Full guide**: [docs/MODULE_DEVELOPMENT.md](docs/MODULE_DEVELOPMENT.md)

---

## Learning Paths

### For New Developers

1. **Start**: [docs/ARCHITECTURE_QUICKREF.md](docs/ARCHITECTURE_QUICKREF.md) - 15-minute overview
2. **Core Concepts**: [internal/CORE.md](internal/CORE.md) - Core responsibilities
3. **Module System**: [internal/modules/README.md](internal/modules/README.md)
4. **Hands-on**: Create a simple module following [docs/MODULE_DEVELOPMENT.md](docs/MODULE_DEVELOPMENT.md)

### For DevOps/Deployment

1. **Makefile**: [docs/MAKEFILE_ARCHITECTURE.md](docs/MAKEFILE_ARCHITECTURE.md)
2. **Migrations**: [migrations/README.md](migrations/README.md)
3. **Docker**: [docker/README.md](docker/README.md)
4. **Configuration**: [docs/MODULE_CONFIG_ARCHITECTURE.md](docs/MODULE_CONFIG_ARCHITECTURE.md)

### For Architects

1. **Architecture Overview**: [docs/ARCHITECTURE_OVERVIEW.md](docs/ARCHITECTURE_OVERVIEW.md)
2. **Audit & Verification**: [docs/ARCHITECTURE_AUDIT.md](docs/ARCHITECTURE_AUDIT.md)
3. **Module Independence**: [docs/MODULE_INDEPENDENCE.md](docs/MODULE_INDEPENDENCE.md)
4. **Migration System**: [docs/MIGRATION_ARCHITECTURE.md](docs/MIGRATION_ARCHITECTURE.md)

**Complete index**: [docs/INDEX.md](docs/INDEX.md)

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow architecture principles (see [docs/ARCHITECTURE_QUICKREF.md](docs/ARCHITECTURE_QUICKREF.md))
4. Write tests (maintain 100% pass rate)
5. Commit changes (`git commit -m 'Add amazing feature'`)
6. Push to branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Support

- **Documentation**: [docs/INDEX.md](docs/INDEX.md)
- **Issues**: [GitHub Issues](https://github.com/basilex/promenade/issues)
- **Email**: alexander.vasilenko@gmail.com

---

**Built with Clean Architecture and Go**
