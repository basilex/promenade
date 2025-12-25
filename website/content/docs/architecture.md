---
title: "Architecture"
description: "Comprehensive architecture documentation with diagrams"
weight: 10
---

## Architecture Overview

Promenade follows **Clean Architecture** principles with a **modular plugin system**.

### System Layers

```mermaid
graph TB
    A[Application Entry Point] --> B[Core Infrastructure]
    B --> C[Core Domain]
    B --> D[Module System]
    C --> E[Core Use Cases]
    D --> F[Module Domains]
    F --> G[Module Use Cases]
    E --> H[Core Adapters]
    G --> I[Module Adapters]
    H --> J[Database/HTTP/Events]
    I --> J
```

### Core vs Modules

**Core = Infrastructure + Security + Reference Data**

- Always enabled
- Authentication (JWT)
- RBAC (roles, permissions)
- Reference data (countries, currencies, regions, cities, payment methods, timezones, languages)
- Event bus, Database, Logger, Scheduler

**Modules = Business Logic**

- Optional, can be enabled/disabled
- Complete vertical slices (entity → usecase → adapter)
- Own migrations, configs, permissions
- Licensed separately (free or commercial)

---

## Clean Architecture Layers

### Dependency Flow

```mermaid
graph LR
    A[Domain Layer] --> B[Use Case Layer]
    B --> C[Adapter Layer]
    C --> D[Infrastructure Layer]

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#a78bfa
    style D fill:#c084fc
```

**Key Rule:** Inner layers never depend on outer layers!

### Layer Responsibilities

**1. Domain Layer** (`internal/domain/entity/`)

- Pure business entities
- Validation methods
- No dependencies on frameworks
- Example: `User`, `Post`, `Comment`

**2. Use Case Layer** (`internal/usecase/`)

- Business logic only
- Depends on domain interfaces
- Example: `RegisterUser`, `CreatePost`

**3. Adapter Layer** (`internal/adapter/`)

- HTTP handlers
- Repository implementations
- Framework integration
- Example: `UserHandler`, `PostgresUserRepo`

**4. Infrastructure Layer** (`internal/infrastructure/`)

- Database connections
- Event bus
- External services
- Example: `PostgreSQL`, `Redis`, `SMTP`

---

## Module Architecture

### Module Structure

Each module is a complete vertical slice:

```
internal/modules/posts/
├── module.go              # Module interface implementation
├── register.go            # Auto-registration
├── config/                # Own configs per environment
│   ├── config.dev.yaml
│   ├── config.test.yaml
│   └── config.prod.yaml
├── entity/                # Domain entities
│   ├── post.go
│   └── comment.go
├── usecase/               # Business logic
│   ├── post_usecase.go
│   └── comment_usecase.go
└── adapter/
    ├── http/              # Handlers, DTOs, routes
    │   ├── handler/
    │   └── dto/
    └── repository/        # Data access
        └── postgres/
```

### Module Lifecycle

```mermaid
sequenceDiagram
    participant A as Application
    participant R as Registry
    participant M as Module
    participant C as Core

    A->>R: Register modules (init)
    R->>M: Initialize(core)
    M->>C: Use DB, EventBus, etc.
    M->>R: RegisterRoutes()
    M->>R: RegisterPermissions()
    A->>M: Start()
    Note over M: Background workers
    A->>M: Stop() on shutdown
```

### Module Independence

**Modules NEVER import:**

- `internal/domain` (core domain)
- `internal/usecase` (core use cases)
- `internal/adapter` (core adapters)
- Other modules

**Modules CAN use:**

- `pkg/*` (shared packages)
- Core infrastructure via `module.Core`
- Events for inter-module communication

---

## Communication Patterns

### Event-Driven Communication

```mermaid
graph LR
    A[Posts Module] -->|user.registered| B[Event Bus]
    B -->|Subscribe| C[Email Module]
    B -->|Subscribe| D[Analytics Module]
    B -->|Subscribe| E[Audit Module]

    style A fill:#38bdf8
    style B fill:#fbbf24
    style C fill:#818cf8
    style D fill:#a78bfa
    style E fill:#c084fc
```

**Benefits:**

- Loose coupling between modules
- Async processing
- Easy to add new subscribers
- No direct module dependencies

### Example Event Flow

```go
// Posts module publishes
event := &PostCreatedEvent{
    PostID: "123",
    UserID: "456",
    Title:  "My Post",
}
eventBus.Publish(ctx, "post.created", event)

// Email module subscribes
eventBus.Subscribe("post.created", func(e Event) {
    SendNotification(e.PostID)
})
```

---

## Database Architecture

### UUID v7 Primary Keys

Time-ordered UUIDs for better performance:

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    email TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**Benefits:**

- ⚡ 2x faster inserts (B-tree locality)
- 📉 Reduced index fragmentation
- 🔍 Sortable by creation time
- 🆔 Globally unique

### Namespace-Based Migrations

```
migrations/
├── core/                    # Core always runs first
│   ├── 000001_init.sql
│   ├── 000002_auth.sql
│   └── 000003_rbac.sql
├── posts/                   # Posts module
│   ├── 000001_posts.sql
│   └── 000002_comments.sql
└── billing/                 # Billing module
    └── 000001_tables.sql
```

**Each namespace has independent version history!**

### Schema Overview

```mermaid
erDiagram
    USERS ||--o{ SESSIONS : has
    USERS ||--o{ USER_ROLES : has
    ROLES ||--o{ USER_ROLES : assigned
    ROLES ||--o{ ROLE_PERMISSIONS : has
    PERMISSIONS ||--o{ ROLE_PERMISSIONS : granted

    USERS ||--o{ USER_POSTS : creates
    USER_POSTS ||--o{ POST_COMMENTS : has
    USERS ||--o{ POST_COMMENTS : writes

    USERS ||--o{ USER_PROFILES : has
    USER_PROFILES ||--o{ USER_CONTACTS : has
```

---

## Purge System Architecture

### Registry-Based Design

```mermaid
graph TB
    A[Core Scheduler] -->|Runs cron| B[Purge Use Case]
    B -->|Gets policies| C[Policy Registry]
    B -->|Gets handlers| D[Handler Registry]

    E[Posts Module] -->|Register| C
    E -->|Register| D

    F[Profiles Module] -->|Register| C
    F -->|Register| D

    G[Billing Module] -->|Register| C
    G -->|Register| D

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#fbbf24
    style D fill:#fbbf24
```

**Key Points:**

- Core knows WHEN to purge (schedule, batch size)
- Modules define WHAT to purge (entities, retention days)
- True module independence via registries

---

## Testing Architecture

### Test Pyramid

```mermaid
graph TB
    A[Smoke Tests<br/>5-10 tests<br/>E2E critical flows]
    B[Integration Tests<br/>50+ tests<br/>Real DB]
    C[Unit Tests<br/>400+ tests<br/>Manual mocks]

    A --> B
    B --> C

    style A fill:#38bdf8
    style B fill:#818cf8
    style C fill:#a78bfa
```

**Coverage:**

- Core: 275 tests
- Posts: 33 tests (83.3%)
- Profiles: 21 tests (80.4%)
- Billing: 375 tests (100%)
- Analytics: 11 tests

**Total: 400+ tests running in ~20 seconds**

---

## Deployment Architecture

### Development

```mermaid
graph LR
    A[Developer] -->|make dev| B[Docker Compose]
    B --> C[PostgreSQL:5432]
    B --> D[Redis:6379]
    E[Go App:8081] --> C
    E --> D
```

### Production

```mermaid
graph TB
    A[GitHub Actions] -->|Deploy| B[Docker Image]
    B --> C[Kubernetes Pod]
    C --> D[PostgreSQL RDS]
    C --> E[Redis ElastiCache]
    C --> F[S3 Storage]

    G[Load Balancer] --> C

    style A fill:#38bdf8
    style C fill:#818cf8
    style D fill:#fbbf24
    style E fill:#fbbf24
    style F fill:#fbbf24
```

---

## Next Steps

- [Module Development Guide](/docs/MODULE_DEVELOPMENT)
- [Testing Strategy](/docs/TESTING_GUIDE)
- [Database Patterns](/features/database)
- [API Documentation](/api/v1/docs/swagger/index.html)
