**Navigation**: [Home](../README.md) > Configuration Files

---

# Configuration Files

**Environment-specific YAML configuration** for Promenade Platform.

---

## Files

```
config/
 app.dev.yaml       # Development (localhost, Memory bus)
 app.test.yaml      # Testing (test DB on port 5433)
 app.prod.yaml      # Production (env vars, Redis bus)
```

---

## Usage

Configuration is loaded automatically based on `ENVIRONMENT` variable:

```bash
ENVIRONMENT=dev ./promenade        # Uses app.dev.yaml
ENVIRONMENT=test ./promenade       # Uses app.test.yaml
ENVIRONMENT=production ./promenade # Uses app.prod.yaml
```

**Default**: `dev`

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
# app.prod.yaml
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
