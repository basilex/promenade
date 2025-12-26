---
title: "Guides"
description: "Complete collection of guides and tutorials for Promenade"
aliases:
  - /docs/guides/
---

## Guides & Tutorials

Comprehensive guides for learning and working with Promenade.

---

## Getting Started

<div class="docs-grid">

### [Quick Start](/promenade/docs/getting-started)

Get up and running in 5 minutes. Install dependencies, start services, and verify installation.

### [Architecture Overview](/promenade/docs/architecture)

Understand the core principles: Clean Architecture, Core vs Modules, and system design.

### [Database Setup](/promenade/docs/database-schema)

Learn about PostgreSQL configuration, UUID v7, migrations, and schema design.

</div>

---

## Development Guides

<div class="docs-grid">

### [Module Development](/promenade/docs/module-development)

Complete guide to creating custom business modules with examples and best practices.

### [Testing Strategy](/promenade/docs/getting-started#testing)

How to write unit, integration, and smoke tests. Learn the manual mocks pattern.

### [API Integration](/api/v1/docs/swagger/index.html)

Explore REST API endpoints, authentication, and integration patterns via Swagger UI.

</div>

---

## Architecture Guides

<div class="docs-grid">

### [Clean Architecture](/promenade/features/clean-architecture)

Deep dive into Domain, Use Case, Adapter, and Infrastructure layers.

### [Module System](/promenade/features/module-system)

How the module system works: registration, lifecycle, and communication patterns.

### [Event-Driven Architecture](/promenade/features/event-bus)

Using the event bus for async communication between modules (Memory/Redis adapters).

</div>

---

## Database Guides

<div class="docs-grid">

### [Database Schema](/promenade/docs/database-schema)

Complete schema documentation with ER diagrams for all tables.

### [UUID v7 Guide](/promenade/features/database#uuid-v7)

Why UUID v7 is 2x faster than UUID v4 and how to use it.

### [Soft Delete Pattern](/promenade/features/database#soft-delete)

Implementing soft delete for user content with automated purge policies.

</div>

---

## Advanced Topics

<div class="docs-grid">

### [RBAC & Permissions](/promenade/features/authentication)

Role-based access control with wildcard permissions and 4 system roles.

### [Reference Data](/promenade/docs/database-schema#reference-data)

145 countries, 124 currencies, 30 regions, 17 cities, 40+ payment methods.

### [Namespace Migrations](/promenade/docs/database-schema#migrations)

Independent migration history for each module without conflicts.

</div>

---

## External Resources

- [GitHub Repository](https://github.com/basilex/promenade) - Source code and issues
- [Example Modules](https://github.com/basilex/promenade/tree/dev/internal/modules) - Posts, Profiles, Analytics, Billing
- [API Documentation](/api/v1/docs/swagger/index.html) - Interactive Swagger UI

---

## Need Help?

- Check [Documentation Index](/promenade/docs/) for all available guides
- Visit [GitHub Discussions](https://github.com/basilex/promenade/discussions) for questions
- Report issues on [GitHub Issues](https://github.com/basilex/promenade/issues)
