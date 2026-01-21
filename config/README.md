**Navigation**: [Home](../README.md) > Configuration Files

---

# Configuration Files

**Driver-environment YAML configuration** for Promenade Platform.

---

## File Naming Convention

**Format**: `app.{driver}-{env}.yaml`

```
config/
 # PostgreSQL configurations (production-ready)
 app.postgres-dev.yaml    # PostgreSQL + development
 app.postgres-test.yaml   # PostgreSQL + testing
 app.postgres-prod.yaml   # PostgreSQL + production

 # MS SQL Server configurations (planned)
 app.mssql-dev.yaml       # MS SQL Server + development
 app.mssql-test.yaml      # MS SQL Server + testing
 app.mssql-prod.yaml      # MS SQL Server + production
```

**Benefits**:

- Clear environment separation
- Database-specific configurations
- Simple configuration management
- Ready for multi-database support

---

## Usage

Configuration is loaded automatically based on the **DATABASE_DRIVER** and **ENVIRONMENT** variables:

```bash
# PostgreSQL + Development
DATABASE_DRIVER=postgres ENVIRONMENT=development ./bin/promenade

# MS SQL Server + Development (when ready)
DATABASE_DRIVER=mssql ENVIRONMENT=development ./bin/promenade

# Testing
DATABASE_DRIVER=postgres ENVIRONMENT=test ./bin/promenade
```

**Default**: `postgres` + `development`

**Makefile shortcuts**:

```bash
# PostgreSQL (production-ready)
make switch-postgres-dev
make dev

# MS SQL Server (planned)
make switch-mssql-dev
make docker-up-mssql
make dev
```

**Requirements**:

- PostgreSQL 14+ (production-ready)
- MS SQL Server 2022+ (planned)
- Docker Compose for local development
- See [docker/README.md](../docker/README.md) for setup

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
# app.postgres-prod.yaml
database:
  postgres:
    host: "${DB_HOST}" # Required
    password: "${DB_PASSWORD}" # Required
jwt:
  secret: "${JWT_SECRET}" # Required (min 32 chars)
```

**Set in production**:

```bash
export DB_HOST="production-db.example.com"
export DB_PASSWORD="secure_password"
export JWT_SECRET="very-long-secret-key-at-least-32-characters"

# Fiscal (Checkbox) printing
export CHECKBOX_API_KEY="checkbox-api-key"
export CHECKBOX_SANDBOX="true"         # optional
export CHECKBOX_TIMEOUT="30s"          # optional
export FISCAL_PDF_OUTPUT_DIR="tmp/fiscal/receipts" # optional
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
