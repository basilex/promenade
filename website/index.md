---
layout: home

hero:
  name: Promenade Platform
  text: Modern Backend Platform
  tagline: Built with Domain-Driven Design, Event-Driven Architecture, and Clean Patterns
  image:
    src: /logo.svg
    alt: Promenade Platform
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/basilex/promenade

features:
  - icon: 🏗️
    title: Domain-Driven Design
    details: Pure DDD with Bounded Contexts, Aggregates, Value Objects, and Domain Events. Each context is autonomous with its own domain model and database schema.
    
  - icon: ⚡
    title: Event-Driven Architecture
    details: Central Event Bus with Memory and Redis adapters. 377K events/sec throughput, automatic retry, panic recovery, and graceful shutdown.
    
  - icon: 🔐
    title: JWT + RBAC
    details: Token-based authentication with Role-Based Access Control. 15-minute access tokens, 7-day refresh tokens, and Redis-backed token revocation.
    
  - icon: 📊
    title: Three-Tier Testing
    details: Professional test organization with 240+ tests. Unit tests (in-place), Smoke tests (mock-based), Integration tests (real DB). 90%+ coverage.
    
  - icon: 🎯
    title: Bounded Contexts
    details: Autonomous business domains - Identity, Customer Management, Order Management, Billing. Contexts communicate only via Event Bus.
    
  - icon: 🚀
    title: Production-Ready
    details: Rate limiting, health checks, graceful shutdown, structured logging, database migrations, Docker support, and comprehensive documentation.
---

## Quick Example

```go
// Publish domain event
event := bus.NewBaseEvent("user.registered", userID)
eventBus.Publish(ctx, bus.TopicUserRegistered, event)

// Subscribe to events
eventBus.Subscribe(bus.TopicUserRegistered, func(ctx context.Context, e bus.Event) error {
    // Handle user registration
    return notificationService.SendWelcomeEmail(ctx, e.AggregateID())
})
```

## Architecture Highlights

::: info Domain-Driven Design
**Aggregates** - Business entities with invariants  
**Value Objects** - Immutable domain concepts (Email, Phone, Money)  
**Domain Events** - Asynchronous communication between contexts  
**Sagas** - Distributed transaction orchestration  
:::

::: tip Event Bus
**Memory Adapter** - 377K events/sec, zero dependencies  
**Redis Adapter** - Distributed, persistent, scalable  
**Retry Policy** - Exponential backoff with panic recovery  
**67 Tests** - 100% passing, production-ready  
:::

::: warning JWT Authentication
**Access Tokens** - 15 minutes (API requests)  
**Refresh Tokens** - 7 days (token renewal)  
**Token Revocation** - Redis-backed blacklist  
**RBAC Support** - Role-based authorization middleware  
:::

## Technology Stack

- **Go 1.23+** - Fast, reliable, type-safe
- **PostgreSQL 16** - ACID transactions, JSONB, UUID v7
- **Redis** - Event Bus, token revocation, caching
- **Gin** - High-performance HTTP framework
- **sqlx** - Raw SQL with safety (no ORM overhead)
- **Testify** - Professional testing framework

## Community

- [GitHub Repository](https://github.com/basilex/promenade)
- [Documentation](https://github.com/basilex/promenade/tree/dev/docs)
- [Contributing Guide](https://github.com/basilex/promenade/blob/dev/CONTRIBUTING.md)
- [Issues](https://github.com/basilex/promenade/issues)

---

<div style="text-align: center; margin-top: 40px; color: #666;">
  <p><strong>Built with Domain-Driven Design and Go</strong></p>
  <p>Licensed under MIT | Copyright © 2024-2025 Promenade Platform</p>
</div>
