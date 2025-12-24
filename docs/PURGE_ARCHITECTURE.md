# Purge System Architecture

## Overview

The purge system is designed to automatically clean up soft-deleted records based on retention policies. It follows the **orchestrator pattern** where:

- **Core** manages infrastructure (scheduler, use case, event bus)
- **Modules** own the business logic (handlers, retention policies)

## Architecture Principles

### 1. IModule Independence

Each module is responsible for:

- Registering purge handlers for their entities
- Defining retention policies for their entities
- Implementing the actual purge logic

Core **never** knows about specific entity types or retention policies.

### 2. Registry Pattern

Two global registries enable module independence:

#### Handler Registry (`purge.DefaultRegistry`)

```go
// IModule registers handler during initialization
handler := NewPostPurgeHandler(db)
purge.DefaultRegistry.Register(handler)
```

#### Policy Registry (`purge.DefaultPolicyRegistry`)

```go
// IModule registers retention policy
policy := purge.RetentionPolicy{
    EntityName:    "user_posts",
    RetentionDays: 90,
    Enabled:       true,
}
purge.DefaultPolicyRegistry.RegisterPolicy(policy)
```

### 3. Core Orchestration

Core's responsibilities are limited to:

1. **Infrastructure**: Load purge config from YAML (enabled, schedule, dry_run, batch_size)
2. **Orchestration**: Collect policies from registry, create use case, start scheduler
3. **Execution**: Trigger purge operations on schedule
4. **Events**: Publish purge events (success/failure)

Core **does NOT**:

- Know about specific entities
- Define retention policies
- Implement purge logic

## Configuration Structure

### Core Config (`config/app.*.yaml`)

```yaml
# Core purge infrastructure only
purge:
  enabled: true # Master switch
  schedule: "0 2 * * *" # Cron schedule (2 AM daily)
  dry_run: false # Preview mode
  batch_size: 1000 # Records per batch
```

### IModule Config (`internal/modules/{name}/config/config.*.yaml`)

```yaml
# IModule's retention policies
purge:
  user_posts:
    retention_days: 90 # Keep for 90 days
    enabled: true

  post_comments:
    retention_days: 30 # Keep for 30 days
    enabled: true
```

## Implementation Flow

### 1. IModule Initialization

```go
func (m *PostsModule) Initialize(core *Core) error {
    // Load module config
    cfg := moduleconfig.Load("internal/modules/posts/config", environment)

    // Get retention settings
    postsRetentionDays := cfg.GetRetentionDays("purge.user_posts.retention_days")
    commentsRetentionDays := cfg.GetRetentionDays("purge.post_comments.retention_days")

    // Register purge handler
    postHandler := NewPostPurgeHandler(db)
    purge.DefaultRegistry.Register(postHandler)

    // Register retention policy
    policy := purge.RetentionPolicy{
        EntityName:    "user_posts",
        RetentionDays: postsRetentionDays,
        Enabled:       true,
    }
    purge.DefaultPolicyRegistry.RegisterPolicy(policy)

    return nil
}
```

### 2. Core Initialization

```go
func InitPurgeModule(purgeConfig config.PurgeConfig, eventBus bus.IBus) {
    // Get all registered handlers
    handlerRegistry := purge.DefaultRegistry

    // Get all registered policies
    policyRegistry := purge.DefaultPolicyRegistry
    policies := policyRegistry.GetAllPolicies()

    // Convert to domain entities
    domainPolicies := convertToEntityPolicies(policies, purgeConfig.Enabled)

    // Create use case
    useCase := usecase.NewPurgeUseCase(
        handlerRegistry,
        domainPolicies,
        purgeConfig.BatchSize,
        eventBus,
    )

    // Create and start scheduler
    scheduler := scheduler.NewScheduler(useCase, purgeConfig.Schedule, ...)
    scheduler.Start(ctx)
}
```

### 3. Purge Execution

```go
// Scheduler triggers purge on cron schedule
func (s *Scheduler) runPurge(ctx context.Context) {
    // Use case iterates over policies
    for _, policy := range policies {
        // Get handler from registry
        handler, ok := registry.Get(policy.EntityName)
        if !ok {
            continue // Skip if no handler
        }

        // Execute purge
        cutoffDate := time.Now().AddDate(0, 0, -policy.RetentionDays)
        recordsPurged, err := handler.Purge(ctx, cutoffDate, batchSize, dryRun)

        // Publish events
        if err != nil {
            eventBus.Publish(ctx, bus.TopicPurgeFailed, ...)
        } else {
            eventBus.Publish(ctx, bus.TopicPurgeCompleted, ...)
        }
    }
}
```

## Creating a New Purge Handler

### 1. Implement Handler Interface

```go
// internal/modules/mymodule/adapter/purge/handler.go
package purge

import (
    "context"
    "time"
    "github.com/jmoiron/sqlx"
)

type MyEntityPurgeHandler struct {
    db *sqlx.DB
}

func NewMyEntityPurgeHandler(db *sqlx.DB) *MyEntityPurgeHandler {
    return &MyEntityPurgeHandler{db: db}
}

func (h *MyEntityPurgeHandler) EntityName() string {
    return "my_entities"
}

func (h *MyEntityPurgeHandler) Purge(
    ctx context.Context,
    cutoffDate time.Time,
    batchSize int,
    dryRun bool,
) (int64, error) {
    query := `
        SELECT id FROM my_entities
        WHERE deleted_at IS NOT NULL
        AND deleted_at < $1
        LIMIT $2
    `

    var ids []string
    if err := h.db.SelectContext(ctx, &ids, query, cutoffDate, batchSize); err != nil {
        return 0, err
    }

    if dryRun {
        return int64(len(ids)), nil // Preview only
    }

    deleteQuery := `DELETE FROM my_entities WHERE id = ANY($1)`
    result, err := h.db.ExecContext(ctx, deleteQuery, pq.Array(ids))
    if err != nil {
        return 0, err
    }

    return result.RowsAffected()
}
```

### 2. Register in IModule

```go
// internal/modules/mymodule/module.go
func (m *MyModule) Initialize(core *Core) error {
    // Load config
    cfg := moduleconfig.Load("internal/modules/mymodule/config", env)
    retentionDays := cfg.GetRetentionDays("purge.my_entities.retention_days")

    // Register handler
    handler := purge.NewMyEntityPurgeHandler(m.db)
    if err := purge.DefaultRegistry.Register(handler); err != nil {
        return err
    }

    // Register policy
    policy := purge.RetentionPolicy{
        EntityName:    "my_entities",
        RetentionDays: retentionDays,
        Enabled:       true,
    }
    if err := purge.DefaultPolicyRegistry.RegisterPolicy(policy); err != nil {
        return err
    }

    slog.Info("Registered purge for my_entities", "retention_days", retentionDays)
    return nil
}
```

### 3. Add IModule Config

```yaml
# internal/modules/mymodule/config/config.dev.yaml
module:
  name: "mymodule"
  enabled: true

purge:
  my_entities:
    retention_days: 60
    enabled: true
```

## Benefits

###  IModule Independence

- Modules own their purge logic completely
- No core dependencies on module entities
- Easy to add/remove modules

###  Configuration Clarity

- Core: Infrastructure settings
- Modules: Business policies
- Clear separation of concerns

###  Testability

- Handlers can be unit tested independently
- Policies can be changed without code changes
- Registry pattern enables easy mocking

###  Maintainability

- Changes to retention policies don't require core changes
- New entities automatically picked up via registry
- Centralized orchestration logic

## Monitoring

### Events Published

- `purge.entity.completed` - Entity purge succeeded
- `purge.entity.failed` - Entity purge failed
- `purge.all.completed` - Full purge cycle completed

### Logs

```
level=info msg="Registered purge for user_posts" retention_days=90
level=info msg="Starting purge operation" entity=user_posts cutoff_date=2024-09-22
level=info msg="Purge operation completed" entity=user_posts records_purged=150 duration=2.3s
```

### Admin Endpoints

- `GET /api/v1/admin/purge/policies` - List all retention policies
- `POST /api/v1/admin/purge/preview/:entity` - Preview purge for entity
- `POST /api/v1/admin/purge/execute/:entity` - Manually trigger purge

## Migration from Old System

**Before** (Core knew about entities):

```go
//  Core had entity-specific config
type PurgeConfig struct {
    RetentionDaysUserPosts    int
    RetentionDaysPostComments int
}
```

**After** (Core only has infrastructure):

```go
//  Core has only infrastructure settings
type PurgeConfig struct {
    Enabled   bool
    Schedule  string
    DryRun    bool
    BatchSize int
}
```

Modules now register their own policies via `purge.DefaultPolicyRegistry`.

## See Also

- [IModule Independence](MODULE_INDEPENDENCE.md)
- [IModule Development](MODULE_DEVELOPMENT.md)
- [Testing Guide](TESTING_GUIDE.md)
