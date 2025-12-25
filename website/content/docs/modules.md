---
title: "Modules"
description: "Complete guide to Promenade's modular architecture and available modules"
aliases:
  - /docs/modules/
---

## Module System

Promenade's revolutionary modular architecture where business logic lives in independent, licensable modules.

---

## Core Concept

**Core = Orchestrator** (Infrastructure, no business logic)

- Authentication & Authorization (JWT, RBAC)
- Event Bus (Memory/Redis adapters)
- Database Management (Connection pool, transactions)
- Reference Data (Countries, currencies, regions, cities, payment methods)
- Logging, Configuration, Scheduler

**Modules = Workers** (Business domains)

- Complete vertical slices (entity → usecase → adapter)
- Independent lifecycle (enable/disable in config)
- Own migrations (namespace-based)
- Own permissions (RBAC integration)
- Event-driven communication

---

## Available Modules

### Free Modules

<div class="docs-grid">

<div class="feature-card">

#### Posts Module

User-generated content management.

**Features:**

- Create/edit posts with markdown
- Nested comments (threading)
- Like system
- Soft delete with purge policies
- Slug generation

**Entities:** `UserPost`, `Comment`, `CommentLike`

[View Source](https://github.com/basilex/promenade/tree/dev/internal/modules/posts)

</div>

<div class="feature-card">

#### Profiles Module

User profile and contact management.

**Features:**

- User profiles with bio, avatar, location
- Multiple contact methods (email, phone, social)
- Privacy settings (public/private/friends)
- Contact verification

**Entities:** `UserProfile`, `UserContact`

[View Source](https://github.com/basilex/promenade/tree/dev/internal/modules/profiles)

</div>

<div class="feature-card">

#### Analytics Module

Metrics collection and reporting.

**Features:**

- Custom metrics recording
- Time-series aggregation
- Dashboard data
- 90-day retention

**Entities:** `Metric`, `MetricAggregate`

[View Source](https://github.com/basilex/promenade/tree/dev/internal/modules/analytics)

</div>

</div>

---

### Commercial Modules

<div class="docs-grid">

<div class="feature-card">

#### Notifications Module 🔔

Multi-channel notification system.

**Features:**

- Multi-channel delivery (Email, SMS, Push, In-App)
- User preferences (per-channel, per-type)
- Quiet hours with timezone support
- Status tracking (sent → delivered → opened → clicked)
- System notifications bypass quiet hours
- Flexible JSONB data storage

**Entities:** `Notification`, `UserPreference`

**Coverage:** 48 tests (20 entity + 15 usecase + 13 integration)

[View Source](https://github.com/basilex/promenade/tree/dev/internal/modules/notifications)

</div>

<div class="feature-card">

#### Billing Module 💰

Production-ready subscription management.

**Features:**

- Flexible billing plans (monthly/quarterly/annual)
- Trial periods
- Subscription lifecycle (active/paused/cancelled)
- Invoice generation
- Payment tracking
- Multiple payment methods

**Entities:** `Plan`, `Subscription`, `Invoice`, `Payment`

**Coverage:** 375 tests, 100% entity + usecase coverage

[View Source](https://github.com/basilex/promenade/tree/dev/internal/modules/billing)

</div>

<div class="feature-card">

#### Audit Module 🔒

Immutable audit logs for compliance.

**Features:**

- Immutable audit trail
- HMAC-SHA256 signatures
- SOC 2 / GDPR compliance
- Tamper detection
- Long-term retention

**Use Cases:** Financial apps, Healthcare, Regulated industries

</div>

<div class="feature-card">

#### Warehouse Module 📦

Inventory and product management (planned).

**Features:**

- Product catalog
- Stock tracking
- Inventory adjustments
- Multi-location support

**Status:** Planned for Q1 2026

</div>

</div>

---

## Module Development

### Create Your Own Module

Follow our [Module Development Guide](/docs/module-development) to build custom modules.

**Quick Structure:**

```
internal/modules/mymodule/
├── module.go              # IModule interface implementation
├── register.go            # Auto-registration via init()
├── config/                # Environment-specific configs
├── entity/                # Domain entities
├── usecase/               # Business logic
└── adapter/
    ├── http/              # HTTP handlers & DTOs
    └── repository/        # PostgreSQL implementations
```

### Module Features

✅ **Auto-registration** - `init()` function registers module  
✅ **Own configuration** - YAML configs per environment  
✅ **Own migrations** - Namespace-based, independent history  
✅ **Own permissions** - RBAC integration  
✅ **Event communication** - Publish/subscribe to domain events  
✅ **Background workers** - Cron jobs, queue processors  
✅ **Health checks** - Graceful startup/shutdown

---

## Module Independence

**Modules NEVER import:**

- `internal/domain` (core domain)
- `internal/usecase` (core use cases)
- `internal/adapter` (core adapters)
- Other modules

**Modules CAN use:**

- `pkg/*` (shared packages)
- Core infrastructure via `module.Core`
- Events for module-to-module communication

This enforces true architectural autonomy.

---

## Enabling Modules

### Configuration

Edit `config/modules.yaml`:

```yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics
    - notifications # Requires license
    - billing # Requires license
    - audit # Requires license
    # - warehouse  # Not yet available
```

### Import in main.go

```go
// cmd/api/main.go
import (
    _ "github.com/basilex/promenade/internal/modules/posts"
    _ "github.com/basilex/promenade/internal/modules/profiles"
    _ "github.com/basilex/promenade/internal/modules/analytics"
    _ "github.com/basilex/promenade/internal/modules/notifications"
    _ "github.com/basilex/promenade/internal/modules/billing"
)
```

### Run Migrations

```bash
make migrate                     # All enabled modules
make migrate-module MODULE=posts # Specific module
```

---

## Commercial Licensing

Commercial modules (billing, audit, warehouse) require license keys.

**Generate License:**

```bash
./scripts/generate-license.sh billing PRO 365
```

**Set Environment Variable:**

```bash
export BILLING_LICENSE_KEY="PROMENADE-BILLING-PRO-20261225-xxxxx"
```

**Or configure in YAML:**

```yaml
# config/modules.yaml
modules:
  config:
    billing:
      license_key: "PROMENADE-BILLING-PRO-20261225-xxxxx"
```

Contact: alexander.vasilenko@gmail.com for enterprise licensing.

---

## Learn More

- [Module Development Guide](/docs/module-development) - Complete tutorial
- [Architecture Overview](/docs/architecture) - System design
- [Database Schema](/docs/database-schema) - Module tables
- [Example Modules](https://github.com/basilex/promenade/tree/dev/internal/modules) - Source code

---

## Support

🐛 [Report Issues](https://github.com/basilex/promenade/issues)  
💬 [Discussions](https://github.com/basilex/promenade/discussions)  
📧 alexander.vasilenko@gmail.com
