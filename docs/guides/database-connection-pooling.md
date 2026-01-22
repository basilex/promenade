# Database Connection Pooling Guide

**Status**:  Implemented  
**Last Updated**: 2026-01-22  
**Related**: [Database Query Patterns](./database-query-patterns.md)

---

## Overview

Connection pooling is critical for PostgreSQL performance. Go's `database/sql` provides built-in pooling with configurable parameters. This guide explains optimal settings for Promenade.

---

## Configuration

### Location

**Config Files**: `config/app.{driver}-{env}.yaml`  
**Implementation**: [`internal/infrastructure/database/postgres.go`](../../internal/infrastructure/database/postgres.go#L33-L35)

### Current Settings

#### Development (`app.postgres-dev.yaml`)

```yaml
database:
  postgres:
    host: "localhost"
    port: 5432
    user: "system"
    password: "passw0rd"
    database: "promenade_dev"
    sslmode: "disable"
    # Connection pool settings (defaults used)
    max_open_conns: 25 # Max concurrent connections
    max_idle_conns: 10 # Idle connections kept alive
    conn_max_lifetime: 30m # Recycle connections after 30 min
```

#### Production (`app.postgres-prod.yaml`)

```yaml
database:
  postgres:
    host: "${DB_HOST}"
    port: 5432
    user: "${DB_USER}"
    password: "${DB_PASSWORD}"
    database: "${DB_NAME}"
    sslmode: "require"
    # Production pool settings (tuned for load)
    max_open_conns: 50 # Higher for production traffic
    max_idle_conns: 25 # 50% of max_open_conns
    conn_max_lifetime: 30m # Match PostgreSQL's idle_in_transaction_session_timeout
```

### Implementation

```go
// internal/infrastructure/database/postgres.go
func NewPostgresConnection(cfg *config.PostgresSection) (*sqlx.DB, error) {
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, err
    }

    // Apply connection pool settings
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

    return db, nil
}
```

---

## Parameter Guide

### 1. MaxOpenConns (Maximum Open Connections)

**What it does**: Limits total concurrent connections to PostgreSQL

**Default**: Unlimited (not recommended)  
**Promenade Dev**: 25  
**Promenade Prod**: 50

#### How to Calculate

```
MaxOpenConns = (CPU cores × 2) + disk spindles
```

**Examples**:

- 4 cores + 1 SSD = `(4 × 2) + 1 = 9` → round up to **10-15**
- 8 cores + 1 SSD = `(8 × 2) + 1 = 17` → round up to **20-25**
- 16 cores + 2 SSDs = `(16 × 2) + 2 = 34` → round up to **40-50**

**Why this formula?**:

- PostgreSQL is CPU-bound for most queries
- Each core can handle ~2 concurrent queries efficiently
- Disk spindles add parallelism for I/O-bound queries
- SSD = 1 effective spindle (high throughput)

#### Warning Signs

**Too Low** (e.g., 5):

-  High connection wait times
-  Errors: `"too many clients already"`
-  Increased request latency

**Too High** (e.g., 200):

-  PostgreSQL context switching overhead
-  Memory exhaustion (each connection ~10MB)
-  Degraded query performance

**Monitoring**:

```sql
-- Check active connections
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';

-- Check waiting connections
SELECT count(*) FROM pg_stat_activity WHERE wait_event IS NOT NULL;
```

---

### 2. MaxIdleConns (Idle Connection Pool)

**What it does**: Number of idle connections kept alive and ready

**Default**: 2 (too low for production)  
**Promenade Dev**: 10  
**Promenade Prod**: 25

#### How to Calculate

```
MaxIdleConns = MaxOpenConns × 0.5
```

**Examples**:

- MaxOpenConns=25 → MaxIdleConns=**12-13**
- MaxOpenConns=50 → MaxIdleConns=**25**
- MaxOpenConns=100 → MaxIdleConns=**50**

**Why 50%?**:

- Keeps connections warm for burst traffic
- Reduces connection establishment overhead (~10-50ms)
- Balances resource usage vs. performance
- PostgreSQL idle connections use minimal memory

#### Trade-offs

**Higher MaxIdleConns**:

-  Faster request handling (no connection setup)
-  Better handling of traffic spikes
-  More PostgreSQL memory usage (~10MB per conn)

**Lower MaxIdleConns**:

-  Lower memory footprint
-  Slower response times (connection setup delay)
-  Poor performance under load

**Monitoring**:

```sql
-- Check idle connections
SELECT count(*) FROM pg_stat_activity WHERE state = 'idle';

-- Check connection age
SELECT pid, usename, state, now() - state_change AS idle_time
FROM pg_stat_activity
WHERE state = 'idle'
ORDER BY idle_time DESC;
```

---

### 3. ConnMaxLifetime (Connection Recycling)

**What it does**: Maximum age of a connection before it's closed and recreated

**Default**: 0 (connections never recycled - dangerous!)  
**Promenade Dev**: 30m (30 minutes)  
**Promenade Prod**: 30m

#### How to Calculate

```
ConnMaxLifetime = PostgreSQL's idle_in_transaction_session_timeout
```

**Typical values**:

- Development: **30m** (match PostgreSQL default)
- Production: **30m-1h** (balance between recycling and stability)

**Why recycle connections?**:

1. **PostgreSQL idle session timeout**: Prevents connections from exceeding server-side limits
2. **Memory leak mitigation**: Forces Go to release connection resources
3. **Load balancer compatibility**: Helps with connection draining during deployments
4. **Network failure recovery**: Closes stale connections after network issues

#### PostgreSQL Configuration

Check your PostgreSQL settings:

```sql
SHOW idle_in_transaction_session_timeout;
-- Default: 0 (disabled)
-- Recommended: 30min

SHOW statement_timeout;
-- Default: 0 (disabled)
-- Recommended: 30s for queries
```

**Match Go's ConnMaxLifetime to PostgreSQL's timeout**:

- PostgreSQL: `idle_in_transaction_session_timeout = 30min`
- Go: `conn_max_lifetime: 30m`

#### Warning Signs

**Too Short** (e.g., 1m):

-  Excessive connection churn (new connections every minute)
-  Increased latency (connection setup overhead)
-  PostgreSQL logs flooded with connect/disconnect events

**Too Long** (e.g., 0 - never recycle):

-  Stale connections after network failures
-  Memory leaks accumulate over time
-  Connections may exceed PostgreSQL's `idle_in_transaction_session_timeout`

**Monitoring**:

```sql
-- Check long-running idle connections
SELECT pid, usename, state, now() - backend_start AS age
FROM pg_stat_activity
WHERE state = 'idle'
  AND backend_start < now() - INTERVAL '1 hour'
ORDER BY age DESC;
```

---

## Production Tuning Guide

### Step 1: Baseline Metrics

Collect current metrics:

```bash
# Application metrics
- Active connections: db.Stats().OpenConnections
- Idle connections: db.Stats().Idle
- Wait count: db.Stats().WaitCount
- Wait duration: db.Stats().WaitDuration

# PostgreSQL metrics
psql -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';"
psql -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'idle';"
```

### Step 2: Load Testing

Use [k6](https://k6.io/) or [vegeta](https://github.com/tsenart/vegeta):

```bash
# Example: 100 req/sec for 5 minutes
vegeta attack -rate=100 -duration=5m -targets=targets.txt | vegeta report
```

**Observe**:

- p50, p95, p99 latency
- Connection wait times
- PostgreSQL CPU/memory usage
- Error rates

### Step 3: Adjust Settings

#### Scenario A: High Wait Times

**Symptom**: `db.Stats().WaitCount` increasing, p95 latency high

**Solution**: Increase `MaxOpenConns`

```yaml
# Before
max_open_conns: 25

# After
max_open_conns: 50
```

#### Scenario B: PostgreSQL CPU 100%

**Symptom**: PostgreSQL CPU maxed out, queries slow

**Solution**: Decrease `MaxOpenConns` (too many concurrent queries)

```yaml
# Before
max_open_conns: 100

# After
max_open_conns: 50
```

#### Scenario C: Slow Connection Establishment

**Symptom**: First request after idle period is slow (~50ms)

**Solution**: Increase `MaxIdleConns`

```yaml
# Before
max_idle_conns: 5

# After
max_idle_conns: 25
```

### Step 4: Monitor in Production

**Prometheus metrics** (implement via `expvar` or `prometheus/client_golang`):

```go
// Example metrics to expose
dbOpenConns.Set(float64(db.Stats().OpenConnections))
dbIdleConns.Set(float64(db.Stats().Idle))
dbWaitCount.Add(float64(db.Stats().WaitCount))
dbWaitDuration.Add(db.Stats().WaitDuration.Seconds())
```

**Grafana dashboard**:

- Graph: `db_open_conns` vs `max_open_conns` (should stay below limit)
- Graph: `db_wait_count` (should be near zero)
- Graph: `db_wait_duration_seconds` (should be < 10ms p99)

---

## Environment-Specific Recommendations

### Development

**Goal**: Fast feedback, minimal resource usage

```yaml
max_open_conns: 10 # Small for local dev
max_idle_conns: 5 # Half of max_open
conn_max_lifetime: 30m # Match PostgreSQL default
```

### Testing (CI/CD)

**Goal**: Isolated tests, no connection conflicts

```yaml
max_open_conns: 5 # Minimal for single-threaded tests
max_idle_conns: 2 # Low idle pool
conn_max_lifetime: 10m # Short for fast test runs
```

### Staging

**Goal**: Match production settings

```yaml
max_open_conns: 50 # Same as production
max_idle_conns: 25 # Same as production
conn_max_lifetime: 30m # Same as production
```

### Production (Low Traffic)

**Goal**: Handle 100-500 req/sec

```yaml
max_open_conns: 25 # Sufficient for low traffic
max_idle_conns: 12 # 50% of max_open
conn_max_lifetime: 30m # Standard
```

### Production (High Traffic)

**Goal**: Handle 1000+ req/sec

```yaml
max_open_conns: 100 # Higher for sustained load
max_idle_conns: 50 # 50% of max_open
conn_max_lifetime: 1h # Longer to reduce churn
```

### Horizontal Scaling

**For massive traffic** (10,000+ req/sec):

- Run multiple app instances (e.g., 10 pods in Kubernetes)
- Each instance: `max_open_conns: 25`
- Total connections: `10 × 25 = 250`
- PostgreSQL: Ensure `max_connections > 250` (e.g., 300)

**PostgreSQL Configuration**:

```sql
-- postgresql.conf
max_connections = 300
shared_buffers = 8GB  -- 25% of total RAM
```

---

## Common Mistakes

###  Mistake 1: Using Default Settings

**Problem**:

```yaml
# No pool settings = unlimited connections
database:
  postgres:
    host: "localhost"
    # Missing: max_open_conns, max_idle_conns, conn_max_lifetime
```

**Impact**: Application can open 1000s of connections, crash PostgreSQL

**Fix**: Always set explicit limits

---

###  Mistake 2: MaxIdleConns > MaxOpenConns

**Problem**:

```yaml
max_open_conns: 10
max_idle_conns: 20 #  Impossible!
```

**Impact**: Go automatically caps `MaxIdleConns` at `MaxOpenConns`

**Fix**: `MaxIdleConns <= MaxOpenConns` (typically 50%)

---

###  Mistake 3: Setting ConnMaxLifetime = 0

**Problem**:

```yaml
conn_max_lifetime: 0 #  Never recycle connections
```

**Impact**: Stale connections accumulate, memory leaks, exceed PostgreSQL timeouts

**Fix**: Always set `conn_max_lifetime: 30m` minimum

---

###  Mistake 4: Per-Request Connections

**Problem**:

```go
//  BAD: Creating new connection per request
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    db, _ := sql.Open("postgres", dsn)  // New connection!
    defer db.Close()
    // ... query ...
}
```

**Impact**:

- Massive overhead (~50ms per connection)
- Exceeds `max_connections` limit
- Connection establishment failures

**Fix**: Use singleton DB instance with connection pool

```go
//  GOOD: Use shared connection pool
var db *sqlx.DB  // Initialized once at startup

func HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Reuses connection from pool
    rows, _ := db.QueryContext(r.Context(), query)
    // ...
}
```

---

## Troubleshooting

### Error: "too many clients already"

**Cause**: Application exceeded PostgreSQL's `max_connections`

**Check**:

```sql
SHOW max_connections;  -- e.g., 100

SELECT count(*) FROM pg_stat_activity;  -- e.g., 105 (over limit!)
```

**Solutions**:

1. **Increase PostgreSQL max_connections**:

   ```sql
   -- postgresql.conf
   max_connections = 200
   ```

   Then restart PostgreSQL.

2. **Decrease application MaxOpenConns**:

   ```yaml
   max_open_conns: 50 # Down from 100
   ```

3. **Use connection pooler** (PgBouncer):
   - Application → PgBouncer (1000 conns) → PostgreSQL (100 conns)
   - PgBouncer multiplexes connections

---

### High Latency on First Request

**Cause**: Connection establishment overhead after idle period

**Check**:

```go
stats := db.Stats()
fmt.Printf("Idle: %d, Open: %d\n", stats.Idle, stats.OpenConnections)
// Output: Idle: 0, Open: 0 (pool was empty!)
```

**Solution**: Increase `MaxIdleConns`

```yaml
max_idle_conns: 25 # Keep connections warm
```

---

### Connection Leaks

**Symptom**: `db.Stats().OpenConnections` keeps growing

**Cause**: Not closing `rows` or `tx` properly

**Check**:

```go
//  BAD: Forgot to close rows
rows, _ := db.Query("SELECT * FROM users")
// Missing: defer rows.Close()

//  BAD: Tx not committed/rolled back
tx, _ := db.Begin()
// Missing: defer tx.Rollback() or tx.Commit()
```

**Solution**: Always defer cleanup

```go
//  GOOD
rows, _ := db.Query("SELECT * FROM users")
defer func() { _ = rows.Close() }()

tx, _ := db.Begin()
defer tx.Rollback()  // No-op if Commit() called
```

See [Database Query Patterns](./database-query-patterns.md) for more.

---

## References

### Go Documentation

- [database/sql.DB.SetMaxOpenConns](https://pkg.go.dev/database/sql#DB.SetMaxOpenConns)
- [database/sql.DB.SetMaxIdleConns](https://pkg.go.dev/database/sql#DB.SetMaxIdleConns)
- [database/sql.DB.SetConnMaxLifetime](https://pkg.go.dev/database/sql#DB.SetConnMaxLifetime)
- [database/sql.DBStats](https://pkg.go.dev/database/sql#DBStats)

### PostgreSQL Documentation

- [Connection Management](https://www.postgresql.org/docs/current/runtime-config-connection.html)
- [max_connections](https://www.postgresql.org/docs/current/runtime-config-connection.html#GUC-MAX-CONNECTIONS)
- [idle_in_transaction_session_timeout](https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-IDLE-IN-TRANSACTION-SESSION-TIMEOUT)

### External Resources

- [Configuring sql.DB for Better Performance](https://www.alexedwards.net/blog/configuring-sqldb)
- [PostgreSQL Connection Pooling: Part 1 – Pros & Cons](https://www.percona.com/blog/2018/06/27/postgresql-connection-pooling-part-1-pros-cons/)

### Related Documentation

- [Database Query Patterns](./database-query-patterns.md)
- [Configuration README](../../config/README.md)
- [PostgreSQL Implementation](../../internal/infrastructure/database/postgres.go)

---

## Change Log

| Date       | Change                          | Author         |
| ---------- | ------------------------------- | -------------- |
| 2026-01-22 | Initial documentation created   | GitHub Copilot |
| TBD        | Add Prometheus metrics examples | -              |
| TBD        | Add PgBouncer integration guide | -              |

---

## Quick Reference Card

```yaml
# Development
max_open_conns: 10
max_idle_conns: 5
conn_max_lifetime: 30m

# Production (Low Traffic)
max_open_conns: 25
max_idle_conns: 12
conn_max_lifetime: 30m

# Production (High Traffic)
max_open_conns: 100
max_idle_conns: 50
conn_max_lifetime: 1h

# Formula
MaxOpenConns = (CPU cores × 2) + disk spindles
MaxIdleConns = MaxOpenConns × 0.5
ConnMaxLifetime = Match PostgreSQL idle_in_transaction_session_timeout
```
