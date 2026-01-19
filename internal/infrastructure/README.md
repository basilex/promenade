**Navigation**: [Home](../../README.md) > [Internal](../README.md) > Infrastructure

---

# Infrastructure - Cross-Cutting Concerns

**Shared infrastructure** components used across all bounded contexts.

---

## Directory Structure

```
internal/infrastructure/
 config/                 # Configuration management
    README.md           # Config documentation
    yaml_config.go      # YAML config loading
 database/               # Database infrastructure
     README.md          # Database documentation
     postgres.go        # PostgreSQL connection
     transaction.go     # Transaction management
 health/                 # Health monitoring
      README.md         # Health checks documentation
      health.go         # Health checker logic
      handler.go        # HTTP health endpoints
```

---

## Components

### Configuration

**YAML-based configuration** with environment variable overrides.

- **Files**: `config/app.{env}.yaml`
- **Environments**: dev, test, prod
- **Validation**: Startup validation of critical settings
- **Documentation**: [config/README.md](config/README.md)

### Database

**PostgreSQL connection and transaction management**.

- **Driver**: sqlx with connection pooling
- **Transactions**: Context-aware with auto-commit/rollback
- **Helper Methods**: Common repository operations via helper methods
- **Documentation**: [database/README.md](database/README.md)

### Health Checks

**Comprehensive dependency monitoring**.

- **Endpoints**: `/health`, `/health/db`, `/health/redis`, `/health/bus`
- **Status Levels**: healthy, degraded, unhealthy
- **Kubernetes**: Liveness/readiness probe support
- **Documentation**: [health/README.md](health/README.md)

---

## Usage

### Configuration

```go
import "github.com/basilex/promenade/internal/infrastructure/config"

cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}
```

### Database

```go
import "github.com/basilex/promenade/internal/infrastructure/database"

db, err := database.NewPostgresConnection(&cfg.Database.Postgres)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### Health Checks

```go
import "github.com/basilex/promenade/internal/infrastructure/health"

healthChecker := health.NewChecker(db, redis, eventBus, version)
healthHandler := health.NewHandler(healthChecker)
healthHandler.RegisterRoutes(router)
```

---

## Related Documentation

- [Main README](../../README.md) - Project overview
- [Documentation Index](../../docs/INDEX.md) - Complete documentation
- [Configuration Guide](config/README.md) - YAML config details
- [Database Guide](database/README.md) - PostgreSQL usage
- [Health Checks Guide](health/README.md) - Monitoring setup
- [Health Checks Implementation](../../docs/guides/health-checks.md) - Complete guide

---

**Purpose**: Shared infrastructure for all contexts  
**Status**: Production-ready  
**Maintainer**: Promenade Team
