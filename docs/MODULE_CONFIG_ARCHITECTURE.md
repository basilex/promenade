# Module Configuration Architecture

## Overview

Modules in Promenade are **fully autonomous vertical slices** that manage their own configuration. Each module loads its own config from its own directory tree, ensuring complete independence from the core system.

## Architecture Principles

### 1. Module Autonomy

- Each module is responsible for loading its own configuration
- Modules store configs in their own directory: `internal/modules/{module_name}/config/`
- Core does not load or manage module configs
- Core only provides infrastructure (DB, EventBus, JWT) and environment context

### 2. Environment Support

- Each module has three environment-specific config files:
  - `config.dev.yaml` - Development settings
  - `config.test.yaml` - Test settings
  - `config.prod.yaml` - Production settings
- Module SDK automatically loads the correct config based on `ENVIRONMENT` variable

### 3. Configuration Loading Flow

```
Application Start
    ↓
Core loads app.{env}.yaml
    ↓
Core initializes infrastructure (DB, EventBus, etc.)
    ↓
Module Registry discovers modules
    ↓
For each enabled module:
    Module.Initialize(ctx, core) is called
        ↓
    Module loads internal/modules/{name}/config/config.{env}.yaml
        ↓
    Module validates settings (license, features, etc.)
        ↓
    Module initializes repositories, use cases, handlers
        ↓
    Module.RegisterRoutes() registers HTTP endpoints
```

## Directory Structure

```
config/
├── app.dev.yaml          # Core config - development
├── app.test.yaml         # Core config - test
├── app.prod.yaml         # Core config - production
└── modules.yaml          # Module registry (enabled/disabled)

internal/modules/
├── posts/
│   ├── config/
│   │   ├── config.dev.yaml   # Posts module dev config
│   │   ├── config.test.yaml  # Posts module test config
│   │   └── config.prod.yaml  # Posts module prod config
│   ├── entity/
│   ├── repository/
│   ├── usecase/
│   ├── handler/
│   └── module.go            # Loads own config in Initialize()
│
├── warehouse/
│   ├── config/
│   │   ├── config.dev.yaml   # Warehouse module dev config
│   │   ├── config.test.yaml  # Warehouse module test config
│   │   └── config.prod.yaml  # Warehouse module prod config
│   └── module.go            # Loads own config in Initialize()
│
└── profiles/
    ├── config/
    │   ├── config.dev.yaml   # Profiles module dev config
    │   ├── config.test.yaml  # Profiles module test config
    │   └── config.prod.yaml  # Profiles module prod config
    └── ...
```

## Configuration Scopes

### Core Config (`config/app.{env}.yaml`)

**Managed by**: `internal/infrastructure/config/yaml_config.go`

Contains:

- Application metadata (name, version, environment)
- Server settings (host, port, timeouts)
- Database connection (host, port, credentials)
- JWT settings (secret, expiration)
- Logging configuration
- CORS settings
- Event bus adapter (memory/redis)
- Rate limiting
- Email service

**Never contains**: Module-specific business logic settings

### Module Config (`internal/modules/{name}/config/config.{env}.yaml`)

**Managed by**: Each module using `pkg/module/config`

Contains:

- Module metadata (name, version, enabled flag)
- Module-specific settings
- Purge policies (retention days, batch sizes)
- Permissions definitions
- Feature flags
- License keys (for commercial modules)

**Never contains**: Core infrastructure settings

## Implementation Example

### Posts Module Config Loading

```go
// internal/modules/posts/module.go
package posts

import (
    "context"
    "github.com/basilex/promenade/pkg/module"
    moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

type PostsModule struct {
    *module.BaseModule
    config *moduleconfig.Config
    // ... other fields
}

func (m *PostsModule) Initialize(ctx context.Context, core *module.Core) error {
    // Call base implementation
    if err := m.BaseModule.Initialize(ctx, core); err != nil {
        return err
    }

    // Load module's own config from its own directory
    config, err := moduleconfig.Load("internal/modules/posts/config", core.Config.AppName)
    if err != nil {
        return fmt.Errorf("failed to load posts config: %w", err)
    }
    m.config = config

    // Use config settings
    retentionDays := m.config.GetRetentionDays("posts", 30)
    batchSize := m.config.GetBatchSize("posts", 100)

    // ... initialize repositories, use cases, handlers

    return nil
}
```

### Module Config File Structure

```yaml
# internal/modules/posts/config/config.dev.yaml
module:
  enabled: true
  name: "posts"
  version: "1.0.0"

settings:
  max_content_length: 10000
  allow_markdown: true

purge:
  enabled: true
  schedule: "0 2 * * *"
  settings:
    posts:
      retention_days: 30
      batch_size: 100
    comments:
      retention_days: 60
      batch_size: 50

permissions:
  - resource: "posts"
    actions: ["create", "read", "update", "delete"]
  - resource: "comments"
    actions: ["create", "read", "update", "delete"]

features:
  enable_likes: true
  enable_sharing: true
```

## Module Config SDK

### Loading Config

```go
import moduleconfig "github.com/basilex/promenade/pkg/module/config"

// Load config for current environment
config, err := moduleconfig.Load("internal/modules/{name}/config", environment)
```

### Accessing Settings

```go
// Get purge retention days with default fallback
retentionDays := config.GetRetentionDays("entity_name", 30)

// Get purge batch size with default fallback
batchSize := config.GetBatchSize("entity_name", 100)

// Check if feature is enabled
enabled := config.GetFeature("enable_likes", true)

// Get nested setting
value := config.GetNestedSetting("settings", "max_content_length")
```

## Migration from Centralized Configs

### Old Architecture (Deprecated)

```
config/modules/
├── posts.dev.yaml         Centralized
├── posts.test.yaml        Centralized
├── posts.prod.yaml        Centralized
└── ...

Core loads all module configs  Tight coupling
Core passes configs to modules  Dependency
```

### New Architecture (Current)

```
internal/modules/posts/config/
├── config.dev.yaml        Module-owned
├── config.test.yaml       Module-owned
└── config.prod.yaml       Module-owned

Module loads own config  Autonomous
Module manages own settings  Independent
```

## Benefits

### 1. True Module Independence

- Modules can be developed, tested, and deployed independently
- No core changes needed when adding/modifying module configs
- Modules are truly autonomous plugins

### 2. Better Encapsulation

- Config lives with the code it configures
- Clear ownership and responsibility
- Easier to understand module capabilities

### 3. Simplified Core

- Core only manages infrastructure
- No business logic in core configs
- Cleaner separation of concerns

### 4. Easier Testing

- Each module can have different test configs
- No global config pollution
- Module tests are isolated

### 5. Flexible Deployment

- Enable/disable modules without core changes
- Different environments can have different module configs
- Commercial modules can have license validation

## Commercial Modules

Commercial modules (e.g., warehouse) require license keys:

```yaml
# internal/modules/warehouse/config/config.prod.yaml
module:
  enabled: true
  name: "warehouse"
  version: "1.2.0"
  license_key: "your-commercial-license-key-here"

settings:
  max_items: 100000
  enable_barcode_scanner: true
```

The module validates the license during `Initialize()`:

```go
func (m *WarehouseModule) verifyLicense() error {
    licenseKey := m.config.GetNestedSetting("module", "license_key").(string)
    if licenseKey == "" {
        return fmt.Errorf("warehouse module requires license key")
    }
    // Validate license...
    return nil
}
```

## Best Practices

### DO 

- Store module configs in `internal/modules/{name}/config/`
- Load config in module's `Initialize()` method
- Use environment-specific config files
- Validate critical settings during initialization
- Use `pkg/module/config` helper methods
- Provide sensible defaults for optional settings

### DON'T 

- Put module configs in `config/modules/` (deprecated)
- Load module configs in core
- Put module settings in core config
- Hardcode environment-specific values
- Skip config validation
- Access other modules' configs

## Debugging

### Check which config is loaded:

```go
logger.FromContext(ctx).Info("Module config loaded",
    "module", m.GetMetadata().Name,
    "config_path", "internal/modules/posts/config",
    "environment", os.Getenv("ENVIRONMENT"),
)
```

### Verify environment:

```bash
echo $ENVIRONMENT  # Should be: development, test, or production
```

### Config loading failures:

- Check file exists: `internal/modules/{name}/config/config.{env}.yaml`
- Verify YAML syntax
- Check file permissions
- Ensure environment is set correctly

## Related Documentation

- [Module Development Guide](MODULE_DEVELOPMENT.md)
- [Module Independence](MODULE_INDEPENDENCE.md)
- [Testing Infrastructure](TESTING_INFRASTRUCTURE.md)
- [Config Migration](CONFIG_MIGRATION.md)
