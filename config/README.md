**Navigation**: [Home](../README.md) > Configuration Files

---

# Configuration Files

**Driver-environment YAML configuration** for Promenade Platform.

---

## File Naming Convention

**Format**: `app.{driver}-{env}.yaml`

```
config/
 # PostgreSQL configurations
 app.postgres-dev.yaml    # PostgreSQL + development
 app.postgres-test.yaml   # PostgreSQL + testing
 app.postgres-prod.yaml   # PostgreSQL + production
 
 # SQLite configurations
 app.sqlite-dev.yaml      # SQLite + development (embedded)
 app.sqlite-test.yaml     # SQLite + testing (in-memory)
 app.sqlite-prod.yaml     # SQLite + production (single instance)
 
 # Future: MySQL, SQL Server, etc.
 # app.mysql-dev.yaml
 # app.sqlserver-prod.yaml
```

**Benefits**:
- ✅ Clear driver selection at file level
- ✅ Easy to find all configs for specific database
- ✅ Scalable for multiple database engines
- ✅ Environment isolation per driver

---

## Usage

Configuration is loaded automatically based on **two** environment variables:

```bash
# PostgreSQL development
DATABASE_DRIVER=postgres ENVIRONMENT=development ./bin/promenade

# SQLite development
DATABASE_DRIVER=sqlite ENVIRONMENT=development ./bin/promenade

# PostgreSQL production
DATABASE_DRIVER=postgres ENVIRONMENT=production ./bin/promenade
```

**Defaults**:
- `DATABASE_DRIVER`: `postgres`
- `ENVIRONMENT`: `development`

**Makefile shortcuts**:
```bash
make dev-postgres  # PostgreSQL + development
make dev-sqlite    # SQLite + development
make dev-mysql     # MySQL + development (coming soon)
```

---

## Structure

Each config file contains:

- **App**: Name, version, environment
- **Server**: Host, port, timeouts
- **Database**: PostgreSQL + Redis settings
- **Bus**: Event Bus configuration (Memory/Redis)
- **JWT**: Authentication settings
- **Logging**: Level, format, output

---

## Environment Variables

Production config uses environment variable overrides:

```yaml
# app.postgres-prod.yaml or app.sqlite-prod.yaml
database:
  postgres:
    host: "${DB_HOST}"           # Required
    password: "${DB_PASSWORD}"   # Required
jwt:
  secret: "${JWT_SECRET}"        # Required (min 32 chars)
```

**Set in production**:

```bash
export DB_HOST="production-db.example.com"
export DB_PASSWORD="secure_password"
export JWT_SECRET="very-long-secret-key-at-least-32-characters"
```

---

## Related Documentation

- [Main README](../README.md) - Project overview
- [Configuration Package](../internal/infrastructure/config/README.md) - Implementation details
- [Documentation Index](../docs/INDEX.md) - Complete documentation

---

**Status**: Production-ready  
**Validation**: Startup validation with security checks  
**Maintainer**: Promenade Team
