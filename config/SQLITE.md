# SQLite Embedded Database Mode

**Perfect for**: Demos, laptops, small deployments, testing

---

## Quick Start

### 1. Use SQLite Configuration

```bash
# Use Makefile target (easiest)
make dev-sqlite

# Or manually with environment variables
DATABASE_DRIVER=sqlite ENVIRONMENT=development ./bin/promenade

# For testing
DATABASE_DRIVER=sqlite ENVIRONMENT=test ./bin/promenade

# For production (single instance only!)
DATABASE_DRIVER=sqlite ENVIRONMENT=production ./bin/promenade
```

### 2. Database File

SQLite creates database file at: `./data/promenade.db`

**Features**:
-  Single file database
-  No server process needed
-  Perfect for demos and development
-  Fast queries (in-memory possible)
-  Zero configuration

---

## Configuration Files

Promenade uses **driver-environment** format:

- `config/app.sqlite-dev.yaml` - Development with SQLite
- `config/app.sqlite-test.yaml` - Testing with SQLite (in-memory)
- `config/app.sqlite-prod.yaml` - Production with SQLite (single instance)

**Example**: `config/app.sqlite-dev.yaml`

```yaml
database:
  driver: "sqlite"  # Driver selection
  
  sqlite:
    path: "./data/promenade.db"  # Database file location
    mode: "rwc"                   # read-write-create
    cache: "shared"               # shared cache
    max_open_conns: 1             # SQLite best practice
    max_idle_conns: 1
```

### Path Options

- `./data/promenade.db` - File on disk (persistent)
- `:memory:` - In-memory database (fast, temporary)
- `file:promenade.db?mode=memory&cache=shared` - Named in-memory (shareable)

---

## Switching Databases

### From SQLite to PostgreSQL

1. **Change driver** in config:
   ```yaml
   database:
     driver: "postgres"  # Was: "sqlite"
   ```

2. **Restart application** - that's it!

All code is database-agnostic - no changes needed.

### From PostgreSQL to SQLite

Same process - just change driver to `"sqlite"`.

---

## Limitations

SQLite has some limitations compared to PostgreSQL:

1. **Concurrency**: Single writer at a time (read concurrency is good)
2. **JSONB**: Stored as TEXT, no native JSONB operators
3. **Full-text search**: Different syntax than PostgreSQL
4. **Performance**: Great for < 100K records, PostgreSQL better for millions

**Recommendation**: Use SQLite for demos/testing, PostgreSQL for production.

---

## Use Cases

###  Perfect For

- **Demos & Presentations**: No database server setup
- **Laptop Development**: Work offline, no Docker needed
- **Small Deployments**: < 10 users, simple workflows
- **Testing**: Fast in-memory tests
- **Embedded Systems**: Single executable + data file

###  Not Recommended For

- **High Concurrency**: > 100 simultaneous writes/sec
- **Large Datasets**: > 1M records
- **Distributed Systems**: Multiple servers accessing same DB
- **High Availability**: No replication built-in

---

## Migrations

Migrations work identically for SQLite and PostgreSQL:

```bash
# Migrations auto-run on startup
make dev

# Or manually
make migrate
```

**Database-agnostic SQL** is used throughout (see `pkg/database` dialect abstraction).

---

## Performance

**SQLite Performance** (on M1 MacBook):

- **Inserts**: ~10K/sec (single transaction)
- **Selects**: ~100K/sec (indexed)
- **Database Size**: 1GB = ~5M small records

**When to upgrade to PostgreSQL**:
- Need > 100 concurrent writes/sec
- Database > 100GB
- Multiple servers accessing same data

---

## Production Considerations

**Do NOT use SQLite in production** unless:

1. Single-server deployment
2. Low concurrent writes (< 10/sec)
3. Small dataset (< 1M records)
4. Easy to backup (single file)

**For production**, use PostgreSQL with connection pooling and replication.

---

## Backup & Restore

### Backup

```bash
# Copy database file
cp ./data/promenade.db ./backups/promenade-$(date +%Y%m%d).db

# Or use SQLite backup command
sqlite3 ./data/promenade.db ".backup './backups/promenade.db'"
```

### Restore

```bash
# Copy backup file back
cp ./backups/promenade-20260103.db ./data/promenade.db

# Restart application
./bin/promenade
```

---

## Troubleshooting

### "Database is locked"

**Cause**: Multiple writers trying to access SQLite simultaneously.

**Solution**:
1. Set `max_open_conns: 1` in config (default)
2. Enable WAL mode (automatic)
3. Or switch to PostgreSQL for high concurrency

### "No such file or directory"

**Cause**: Database file path doesn't exist.

**Solution**: Application creates `./data/` directory automatically.

### Migrations fail

**Cause**: SQL syntax might be PostgreSQL-specific.

**Solution**: Use `pkg/database` dialect abstraction for database-agnostic SQL.

---

## Related Documentation

- [Database Adapters Guide](../docs/guides/database-adapters.md)
- [Multi-Database Support](../docs/work-in-progress/DB_AGNOSTIC_REFACTORING.md)
- [Configuration Reference](../docs/reference/configuration.md)

---

**Status**:  Production-ready for embedded use cases  
**SQLite Version**: 3.45+ (via go-sqlite3 driver)  
**Last Updated**: January 3, 2026
