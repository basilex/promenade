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

### [Quick Start](/docs/getting-started)

Get up and running in 5 minutes. Install dependencies, start services, and verify installation.

### [Architecture Overview](/docs/architecture)

Understand the core principles: Clean Architecture, Core vs Modules, and system design.

### [Database Setup](/docs/database-schema)

Learn about PostgreSQL configuration, UUID v7, migrations, and schema design.

</div>

---

## Development Guides

<div class="docs-grid">

### [Module Development](/docs/module-development)

Complete guide to creating custom business modules with examples and best practices.

### [Testing Strategy](/docs/getting-started#testing)

How to write unit, integration, and smoke tests. Learn the manual mocks pattern.

### [API Integration](/api/v1/docs/swagger/index.html)

Explore REST API endpoints, authentication, and integration patterns via Swagger UI.

</div>

---

## Architecture Guides

<div class="docs-grid">

### [Clean Architecture](/features/clean-architecture)

Deep dive into Domain, Use Case, Adapter, and Infrastructure layers.

### [Module System](/features/module-system)

How the module system works: registration, lifecycle, and communication patterns.

### [Event-Driven Architecture](/features/event-bus)

Using the event bus for async communication between modules (Memory/Redis adapters).

</div>

---

## Database Guides

<div class="docs-grid">

### [Database Schema](/docs/database-schema)

Complete schema documentation with ER diagrams for all tables.

### [UUID v7 Guide](/features/database#uuid-v7)

Why UUID v7 is 2x faster than UUID v4 and how to use it.

### [Soft Delete Pattern](/features/database#soft-delete)

Implementing soft delete for user content with automated purge policies.

</div>

---

## Advanced Topics

<div class="docs-grid">

### [RBAC & Permissions](/features/authentication)

Role-based access control with wildcard permissions and 4 system roles.

### [Reference Data](/docs/database-schema#reference-data)

145 countries, 124 currencies, 30 regions, 17 cities, 40+ payment methods.

### [Namespace Migrations](/docs/database-schema#migrations)

Independent migration history for each module without conflicts.

</div>

---

## External Resources

- [GitHub Repository](https://github.com/basilex/promenade) - Source code and issues
- [Example Modules](https://github.com/basilex/promenade/tree/dev/internal/modules) - Posts, Profiles, Analytics, Billing
- [API Documentation](/api/v1/docs/swagger/index.html) - Interactive Swagger UI

---

## Need Help?

- Check [Documentation Index](/docs/) for all available guides
- Visit [GitHub Discussions](https://github.com/basilex/promenade/discussions) for questions
- Report issues on [GitHub Issues](https://github.com/basilex/promenade/issues)
