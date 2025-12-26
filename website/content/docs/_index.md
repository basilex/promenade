---
title: "Documentation"
description: "Complete technical documentation for Promenade"
---

## Technical Documentation

Comprehensive guides, diagrams, and references for Promenade architecture and development.

---

## Architecture

<div class="docs-grid">

### [Architecture Overview](/promenade/docs/architecture)

Complete system architecture with diagrams, layers, and communication patterns.

- Clean Architecture layers
- Module system design
- Event-driven communication
- Deployment architecture

### [Database Schema](/promenade/docs/database-schema)

Detailed database schema with ER diagrams and relationships.

- Core tables (users, sessions, RBAC)
- Reference data (countries, currencies, regions, cities)
- Module tables (posts, profiles, analytics, billing)
- Indexes and performance

### [Module Development](/promenade/docs/module-development)

Step-by-step guide to creating new modules.

- Module structure and lifecycle
- Communication patterns
- Best practices
- Examples and troubleshooting

</div>

---

## Quick Links

### For Developers

- [Getting Started Guide](/promenade/docs/getting-started)
- [Module Development](/promenade/docs/module-development)
- [Database Schema](/promenade/docs/database-schema)
- [API Documentation](/api/v1/docs/swagger/index.html)

### For Architects

- [Architecture Overview](/promenade/docs/architecture)
- [Database Schema](/promenade/docs/database-schema)
- [Module Development](/promenade/docs/module-development)
- [Clean Architecture](/promenade/features/clean-architecture)

### API & Integration

- [Authentication & RBAC](/promenade/features/authentication)
- [Event Bus System](/promenade/features/event-bus)
- [REST API v1](/api/v1/docs/swagger/index.html)
- [REST API v2](/api/v2/docs/swagger/index.html)

---

## Features Deep Dive

<div class="features-grid">

<div class="feature-card">

### [Clean Architecture](/promenade/features/clean-architecture)

Strict layered architecture with clear separation of concerns.

</div>

<div class="feature-card">

### [Modular System](/promenade/features/module-system)

Independent business modules with plugin architecture.

</div>

<div class="feature-card">

### [Authentication & RBAC](/promenade/features/authentication)

JWT authentication with role-based access control.

</div>

<div class="feature-card">

### [Database & Migrations](/promenade/features/database)

PostgreSQL with UUID v7 and namespace-based migrations.

</div>

<div class="feature-card">

### [Testing Infrastructure](/promenade/features/testing)

400+ tests with manual mocks pattern.

</div>

<div class="feature-card">

### [Event-Driven Architecture](/promenade/features/event-bus)

Dual event bus adapters for async communication.

</div>

</div>

---

## External Resources

- [GitHub Repository](https://github.com/basilex/promenade)
- [Issue Tracker](https://github.com/basilex/promenade/issues)
- [Contributing Guide](https://github.com/basilex/promenade/blob/dev/CONTRIBUTING.md)
- [MIT License](https://github.com/basilex/promenade/blob/dev/LICENSE)

---

<style>
.docs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 2rem;
  margin: 2rem 0;
}

.docs-grid > div {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.docs-grid > div:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
}

.docs-grid h3 {
  margin-top: 0;
  margin-bottom: 0.5rem;
}

.docs-grid h3 a {
  color: var(--color-primary);
  text-decoration: none;
}

.docs-grid p {
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary);
}

.docs-grid ul {
  margin: 0;
  padding-left: 1.25rem;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}
</style>
