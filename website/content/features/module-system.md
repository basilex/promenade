---
title: "Modular System"
description: "Independent business modules with plugin architecture"
weight: 2
---

## Modular Plugin System

Promenade's **module system** enables building applications as collections of independent, reusable modules.

### What is a Module?

Each module is a **complete vertical slice**:

- Own domain entities
- Own repositories
- Own use cases
- Own HTTP handlers
- Own database migrations
- Own configuration

### Available Modules

**Free Modules:**

- **Posts** - User-generated content (posts, comments, likes)
- **Profiles** - User profiles and contacts
- **Analytics** - Metrics, reports, dashboards

**Commercial Modules:**

- **Billing** - Subscription management, invoicing, payments
- **Audit** - Immutable audit logs with cryptographic signatures
- **Warehouse** - Inventory management (coming soon)

### Module Independence

```go
// Each module implements this interface
type IModule interface {
    Initialize(ctx, core) error
    RegisterRoutes(router)
    RegisterMigrations() []Migration
    RegisterPermissions() []Permission
    Start(ctx) error
    Stop(ctx) error
}
```

### Enable/Disable Modules

```yaml
# config/modules.yaml
modules:
  enabled:
    - posts
    - profiles
    - analytics
    # - warehouse  # Disable by commenting out
```

### Benefits

✅ **True independence** - Modules don't import each other  
✅ **Dynamic loading** - Enable/disable without code changes  
✅ **Own migrations** - Each module has independent schema history  
✅ **Licensable** - Sell commercial modules separately

[Module Development Guide →](/promenade/docs/MODULE_DEVELOPMENT)
