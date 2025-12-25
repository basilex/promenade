---
title: "Production-Ready REST API Framework"
description: "Built with Clean Architecture, modular plugin system, and namespace-based database migrations. Written in Go."
date: 2025-12-25
features:
  - icon: "🏗️"
    title: "Clean Architecture"
    description: "Strict layered architecture with Domain, Use Case, Adapter, and Infrastructure layers. Dependency rule enforced."

  - icon: "🧩"
    title: "Modular System"
    description: "Independent business modules with own entities, repositories, and use cases. Enable/disable modules dynamically."

  - icon: "🗄️"
    title: "PostgreSQL + UUID v7"
    description: "Time-ordered UUIDs for 2x faster inserts. Raw SQL with sqlx - no ORM overhead. BaseRepository pattern."

  - icon: "🔐"
    title: "Auth & RBAC"
    description: "JWT authentication, role-based access control with wildcard permissions. Session management included."

  - icon: "📡"
    title: "Event-Driven"
    description: "Dual event bus adapters (Memory/Redis). Async communication between modules. Production-ready."

  - icon: "🔄"
    title: "Namespace Migrations"
    description: "Each module has independent migration history. True module autonomy without schema conflicts."

  - icon: "✅"
    title: "400+ Tests"
    description: "Comprehensive test suite with unit, integration, and smoke tests. Manual mocks pattern for testability."

  - icon: "📚"
    title: "20+ Docs"
    description: "Extensive documentation in EN/UK/DE. Architecture guides, tutorials, and API references."

  - icon: "🚀"
    title: "Production-Ready"
    description: "Used in real-world applications. Commercial modules available (billing, audit, warehouse)."
---

## Quick Start

```bash
# Clone repository
git clone https://github.com/basilex/promenade.git
cd promenade

# Start PostgreSQL + Redis
make docker-up

# Run migrations
make migrate

# Start development server
make dev
```

Server starts on **http://localhost:8081**

## Core Features

### Clean Architecture

- **Domain Layer**: Business entities and interfaces
- **Use Case Layer**: Business logic implementation
- **Adapter Layer**: HTTP handlers, repositories
- **Infrastructure**: Database, event bus, configuration

### Module System

- **Core**: Auth, RBAC, event bus, reference data (always enabled)
- **Modules**: Posts, Profiles, Analytics (free), Billing (commercial)
- Auto-registration via `init()` functions
- Independent lifecycle management

### Database

- **PostgreSQL 16** with sqlx (raw SQL, no ORM)
- **UUID v7** primary keys (time-ordered for better performance)
- **Soft delete** pattern for user content
- **Namespace-based migrations** (core, posts, profiles, billing)

### Testing

- **275 core tests**: Domain entities + use cases
- **125+ module tests**: Posts, profiles, analytics, billing
- **Manual mocks**: Inline struct mocks for testability
- **~20 seconds** for full test suite

## Architecture Highlights

```
Core (Orchestrator)          Modules (Workers)
├── Auth & RBAC              ├── Posts Module
├── Event Bus                ├── Profiles Module
├── Database                 ├── Analytics Module (Free)
├── Logger                   └── Billing Module (Commercial)
├── Config
└── Reference Data
    ├── Countries (145)
    ├── Currencies (124)
    ├── Regions (30)
    ├── Cities (17)
    ├── Payment Methods (40+)
    ├── Timezones
    └── Languages
```

**Key Principle**: Core orchestrates, modules execute. Core knows WHEN to call modules, not HOW they work.

## What Makes Promenade Different?

✅ **True module independence** - no shared domain entities
✅ **Namespace-based migrations** - each module has own version history
✅ **UUID v7** - time-ordered for better database performance
✅ **Manual mocks** - testability without code generation
✅ **Event-driven** - async communication via Memory/Redis bus
✅ **Multilingual docs** - EN, UK, DE with more coming
✅ **Production-tested** - commercial modules battle-tested

## Learn More

- [Documentation](/docs/) - Complete guides and tutorials
- [Architecture](/docs/architecture/) - System design and principles
- [Modules](/docs/modules/) - Building custom modules
- [GitHub](https://github.com/basilex/promenade) - Source code and issues
