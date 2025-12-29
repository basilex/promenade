# Command Line Tools

**Application entry points** for server and utilities.

---

## Directory Structure

```
cmd/
 api/              # HTTP API server
    main.go       # Server entry point
 migrate/          # Migration CLI tool
     main.go      # Migration runner
```

---

## Commands

### API Server

**HTTP server** with all contexts and routes.

```bash
# Build
make build

# Run
./bin/promenade

# Or directly
go run cmd/api/main.go
```

**Entry point**: [api/main.go](api/main.go)

**Initialization order**:
1. Load configuration
2. Initialize logger
3. Connect to database
4. Run migrations (auto)
5. Initialize Event Bus
6. Initialize JWT manager
7. Register context routers
8. Start HTTP server

---

### Migration Tool

**Database migration** management CLI.

```bash
# Run all migrations
make migrate

# Run specific namespace
make migrate-identity

# Create new migration
make migrate-new CONTEXT=identity NAME=add_field
```

**Entry point**: [migrate/main.go](migrate/main.go)

**See**: [migrations/README.md](../migrations/README.md) for migration details

---

## Related Documentation

- [Main README](../README.md) - Project overview
- [Documentation Index](../docs/INDEX.md) - Complete documentation
- [Configuration](../config/README.md) - Config files
- [Migrations](../migrations/README.md) - Database migrations

---

**Status**: Production-ready  
**Go Version**: 1.23+  
**Maintainer**: Promenade Team
