# Troubleshooting Guide

**Complete troubleshooting reference** for common issues in Promenade Platform.

---

## Overview

This guide covers:

1. **Server & Application Issues** - Startup, configuration, crashes
2. **Authentication & Authorization** - Token issues, permissions, rate limits
3. **Database Problems** - Connection, migrations, queries
4. **API Errors** - HTTP status codes, validation, timeouts
5. **Performance Issues** - Slow queries, memory leaks
6. **Docker & Infrastructure** - Container issues, networking
7. **Testing Problems** - Test failures, flaky tests
8. **Development Workflow** - Common development mistakes

---

## 1. Server & Application Issues

### Problem: Server Won't Start

**Symptoms**:
```bash
$ make dev
Error: failed to start server
```

**Common Causes**:

#### Cause 1: Port Already in Use

**Check**:
```bash
lsof -i :8081
```

**Solution**:
```bash
# Kill process using port
kill -9 <PID>

# Or change port in config
# config/app.postgres-dev.yaml
server:
  port: 8082  # Use different port
```

---

#### Cause 2: Database Not Running

**Check**:
```bash
make docker-ps
# Should show promenade_postgres running
```

**Solution**:
```bash
# Start database
make docker-up

# Wait for PostgreSQL
sleep 3

# Try again
make dev
```

---

#### Cause 3: Invalid Configuration

**Check logs**:
```bash
# Look for validation errors
./bin/promenade

# Output might show:
# ERROR: Invalid configuration: JWT secret too short
```

**Solution**:
```bash
# Check config file
cat config/app.postgres-dev.yaml

# Ensure all required fields present:
# - jwt.secret (min 32 chars)
# - database.host
# - database.password
```

---

### Problem: Application Crashes on Startup

**Symptoms**:
```
panic: runtime error: invalid memory address
```

**Common Causes**:

#### Cause 1: Missing Environment Variables

**Check**:
```bash
echo $DATABASE_DRIVER
echo $ENVIRONMENT
```

**Solution**:
```bash
# Configure workspace
make switch-postgres-dev

# Or set manually
export DATABASE_DRIVER=postgres
export ENVIRONMENT=development
```

---

#### Cause 2: Database Connection Failed

**Error**:
```
FATAL: Failed to connect to PostgreSQL: connection refused
```

**Check**:
```bash
# Test PostgreSQL connection
docker exec -it promenade_postgres psql -U system -d promenade_dev -c "SELECT 1;"
```

**Solution**:
```bash
# Restart database
make docker-down
make docker-up

# Check database logs
docker logs promenade_postgres
```

---

### Problem: Migrations Fail

**Symptoms**:
```
ERROR: Failed to run migrations namespace=core error=pq: relation already exists
```

**Diagnosis**:
```bash
# Check migration status
make migrate-status

# Look for applied migrations
psql -h localhost -U system -d promenade_dev -c "SELECT * FROM schema_migrations ORDER BY id;"
```

**Solutions**:

#### Option 1: Reset Database (Development Only)
```bash
make db-fresh
# Drops database, recreates, runs all migrations
```

#### Option 2: Manual Migration Fix
```bash
# Rollback specific migration
psql -h localhost -U system -d promenade_dev

# Delete migration entry
DELETE FROM schema_migrations WHERE namespace='core' AND version='000002';

# Re-run migration
make migrate-core
```

---

## 2. Authentication & Authorization Issues

### Problem: 401 Unauthorized - Missing Token

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "missing authorization header"
  }
}
```

**Cause**: Authorization header not included in request

**Solution**:
```bash
# Include Bearer token
curl -X GET http://localhost:8081/api/v1/customers \
  -H "Authorization: Bearer $TOKEN"

# Check token is set
echo $TOKEN
```

---

### Problem: 401 Unauthorized - Invalid Token

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid token signature"
  }
}
```

**Common Causes**:

#### Cause 1: Token Corrupted

**Check**:
```bash
# View token (should have 3 parts separated by dots)
echo $TOKEN
# Should look like: eyJhbGc...abc.eyJ1c2V...xyz.Gz5o8K...123
```

**Solution**: Login again to get fresh token

---

#### Cause 2: Wrong JWT Secret

**Check**:
```bash
# Ensure JWT_SECRET matches between token generation and validation
grep "secret:" config/app.postgres-dev.yaml
```

**Solution**: Use same secret everywhere, don't change mid-session

---

### Problem: 401 Unauthorized - Expired Token

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "token is expired"
  }
}
```

**Cause**: Access token expired (15 minute TTL)

**Solutions**:

#### Option 1: Refresh Token
```bash
curl -X POST http://localhost:8081/api/v1/identity/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"'"$REFRESH_TOKEN"'"}'
```

#### Option 2: Login Again
```bash
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Pass123"}'
```

#### Option 3: Auto-Refresh Script
```bash
# Save to refresh-token.sh
#!/bin/bash
TOKEN_FILE="$HOME/.promenade-token"
REFRESH_TOKEN_FILE="$HOME/.promenade-refresh-token"

refresh_token() {
  REFRESH_TOKEN=$(cat $REFRESH_TOKEN_FILE)
  RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/identity/auth/refresh \
    -H "Content-Type: application/json" \
    -d '{"refresh_token":"'"$REFRESH_TOKEN"'"}')
  
  NEW_TOKEN=$(echo $RESPONSE | jq -r '.data.access_token')
  NEW_REFRESH=$(echo $RESPONSE | jq -r '.data.refresh_token')
  
  echo $NEW_TOKEN > $TOKEN_FILE
  echo $NEW_REFRESH > $REFRESH_TOKEN_FILE
  
  export TOKEN=$NEW_TOKEN
  export REFRESH_TOKEN=$NEW_REFRESH
}

# Run every 14 minutes (before 15 min expiry)
while true; do
  refresh_token
  sleep 840  # 14 minutes
done
```

---

### Problem: 403 Forbidden - Insufficient Permissions

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "FORBIDDEN",
    "message": "admin role required"
  }
}
```

**Diagnosis**:
```bash
# Check current user roles
curl -X GET http://localhost:8081/api/v1/identity/users/me \
  -H "Authorization: Bearer $TOKEN"

# Response shows roles
{
  "data": {
    "id": "...",
    "email": "user@example.com",
    "roles": ["user"]  # Missing "admin"
  }
}
```

**Solutions**:

#### Option 1: Request Admin Access
Contact system administrator to assign admin role

#### Option 2: Login as Admin
```bash
# Use admin credentials
curl -X POST http://localhost:8081/api/v1/identity/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"AdminPass123"}'
```

---

### Problem: 429 Too Many Requests

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "too many login attempts, retry after 60 seconds"
  }
}
```

**Cause**: Rate limit exceeded

**Rate Limits**:
- Login: 5 requests/minute per IP
- Register: 3 requests/minute per IP

**Solution**:
```bash
# Wait for rate limit window
sleep 60

# Retry request
curl -X POST http://localhost:8081/api/v1/identity/users/login ...
```

**Prevention**:
```bash
# Add delay between requests in scripts
for i in {1..10}; do
  curl -X POST http://localhost:8081/api/v1/identity/users/login ...
  sleep 15  # 15 seconds between attempts
done
```

---

## 3. Database Problems

### Problem: Database Connection Refused

**Error**:
```
FATAL: Failed to connect to PostgreSQL: connection refused
```

**Check**:
```bash
# Is database running?
make docker-ps

# Test connection
psql -h localhost -U system -d promenade_dev -c "SELECT 1;"
```

**Solutions**:

#### Solution 1: Start Database
```bash
make docker-up
sleep 3  # Wait for startup
```

#### Solution 2: Check Port
```bash
# Verify PostgreSQL on port 5432
netstat -an | grep 5432

# Or check Docker port mapping
docker port promenade_postgres
```

#### Solution 3: Check Credentials
```bash
# Verify config
grep -A5 "postgres:" config/app.postgres-dev.yaml

# Should match:
# host: localhost
# port: 5432
# user: system
# password: passw0rd
# database: promenade_dev
```

---

### Problem: Migration Already Applied

**Error**:
```
ERROR: pq: duplicate key value violates unique constraint "schema_migrations_pkey"
```

**Diagnosis**:
```bash
# Check applied migrations
psql -h localhost -U system -d promenade_dev -c \
  "SELECT namespace, version, applied_at FROM schema_migrations ORDER BY applied_at;"
```

**Solution**:
```bash
# Skip duplicate (migration system handles this)
# Or reset database for clean slate
make db-fresh
```

---

### Problem: Slow Queries

**Symptoms**:
- API responses > 1 second
- Database CPU usage high

**Diagnosis**:
```sql
-- Check slow queries (PostgreSQL)
SELECT 
  pid,
  now() - query_start AS duration,
  query 
FROM pg_stat_activity 
WHERE state = 'active' 
  AND now() - query_start > interval '1 second'
ORDER BY duration DESC;
```

**Common Causes**:

#### Cause 1: Missing Index

**Check**:
```sql
-- List tables without indexes
SELECT tablename 
FROM pg_tables 
WHERE schemaname = 'public' 
  AND tablename NOT IN (
    SELECT DISTINCT tablename 
    FROM pg_indexes 
    WHERE schemaname = 'public'
  );
```

**Solution**: Add index
```sql
-- Example: Index on customer email
CREATE INDEX idx_customers_email ON customer_customers(email);

-- Index on foreign key
CREATE INDEX idx_orders_customer_id ON order_orders(customer_id);
```

---

#### Cause 2: N+1 Query Problem

**Symptoms**: Multiple queries in loop

**Example Bad Code**:
```go
// BAD: N+1 queries
customers, _ := repo.ListCustomers(ctx, 1, 100)
for _, customer := range customers {
  company, _ := repo.GetCompany(ctx, customer.CompanyID)  // N queries!
  // ...
}
```

**Solution**: LEFT JOIN
```go
// GOOD: Single query with JOIN
query := `
  SELECT 
    c.*,
    comp.name AS company_name
  FROM customer_customers c
  LEFT JOIN customer_companies comp ON comp.id = c.company_id
  WHERE c.deleted_at IS NULL
`
```

---

### Problem: Database Locks

**Error**:
```
ERROR: deadlock detected
```

**Diagnosis**:
```sql
-- Check locks
SELECT 
  l.pid,
  l.locktype,
  l.relation::regclass,
  l.mode,
  a.query
FROM pg_locks l
JOIN pg_stat_activity a ON a.pid = l.pid
WHERE NOT l.granted;
```

**Solution**:
```bash
# Kill blocking query
psql -h localhost -U system -d promenade_dev -c "SELECT pg_terminate_backend(<PID>);"

# Or restart PostgreSQL
make docker-down
make docker-up
```

---

## 4. API Errors

### Problem: 400 Bad Request - Validation Error

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email: must be a valid email address"
  }
}
```

**Common Causes**:

#### Cause 1: Invalid Email Format
```bash
# BAD
curl -X POST ... -d '{"email":"not-an-email"}'

# GOOD
curl -X POST ... -d '{"email":"user@example.com"}'
```

#### Cause 2: Missing Required Field
```bash
# BAD
curl -X POST ... -d '{"name":"John"}'

# GOOD
curl -X POST ... -d '{"name":"John","email":"john@example.com","password":"Pass123"}'
```

#### Cause 3: Invalid Password
```bash
# BAD (too short)
curl -X POST ... -d '{"password":"123"}'

# GOOD (min 8 chars, letter + digit)
curl -X POST ... -d '{"password":"SecurePass123"}'
```

---

### Problem: 404 Not Found

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "NOT_FOUND",
    "message": "customer not found"
  }
}
```

**Diagnosis**:
```bash
# Verify resource exists
curl -X GET http://localhost:8081/api/v1/customer-mgmt/customers/$CUSTOMER_ID \
  -H "Authorization: Bearer $TOKEN"

# Check ID format (should be UUID v7)
echo $CUSTOMER_ID
# Should be: 01JGABC123DEF456GHI789JKL0
```

**Common Mistakes**:
- Wrong UUID format
- Resource was deleted (soft delete)
- Wrong endpoint URL

---

### Problem: 500 Internal Server Error

**Error Response**:
```json
{
  "status": "error",
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "internal server error"
  }
}
```

**Diagnosis**:
```bash
# Check server logs
docker logs promenade_api

# Look for panic or error stack trace
```

**Common Causes**:
- Nil pointer dereference
- Database constraint violation
- Unhandled edge case

**Solution**: Report bug with request details and logs

---

### Problem: Request Timeout

**Error**:
```
curl: (28) Operation timed out after 30000 milliseconds
```

**Diagnosis**:
```bash
# Check server health
curl http://localhost:8081/health

# Check database connection
curl http://localhost:8081/health/db
```

**Common Causes**:
- Database query too slow
- Server overloaded
- Network issue

**Solutions**:

#### Solution 1: Increase Timeout
```bash
curl --max-time 60 http://localhost:8081/api/v1/customers
```

#### Solution 2: Add Pagination
```bash
# BAD: Fetch all customers (slow)
curl http://localhost:8081/api/v1/customer-mgmt/customers

# GOOD: Paginate results
curl "http://localhost:8081/api/v1/customer-mgmt/customers?page=1&page_size=20"
```

---

## 5. Performance Issues

### Problem: Memory Leak

**Symptoms**:
- Memory usage grows over time
- Eventually causes OOM (Out of Memory)

**Diagnosis**:
```bash
# Monitor memory usage
docker stats promenade_api

# Go profiling
curl http://localhost:8081/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**Common Causes**:
- Goroutine leak (missing context cancellation)
- Database connection leak (missing Close())
- Cache without TTL

**Prevention**:
```go
// Always use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Always close database connections
db, err := sqlx.Open("postgres", dsn)
defer db.Close()

// Always set cache TTL
cache.Set(ctx, key, value, 1*time.Hour)  // Not forever!
```

---

### Problem: High CPU Usage

**Diagnosis**:
```bash
# CPU profiling
curl http://localhost:8081/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

**Common Causes**:
- Inefficient algorithm (O(n²) instead of O(n))
- Too many goroutines
- JSON marshaling in tight loop

**Solutions**:
- Use proper data structures (map instead of array for lookups)
- Limit goroutine pool size
- Cache expensive computations

---

## 6. Docker & Infrastructure

### Problem: Docker Container Won't Start

**Error**:
```
ERROR: Cannot start service postgres: driver failed
```

**Diagnosis**:
```bash
# Check Docker daemon
docker ps

# Check container logs
docker logs promenade_postgres

# Check disk space
df -h
```

**Solutions**:

#### Solution 1: Restart Docker
```bash
# macOS
killall Docker && open /Applications/Docker.app

# Linux
sudo systemctl restart docker
```

#### Solution 2: Remove Old Containers
```bash
make docker-clean
make docker-up
```

#### Solution 3: Prune Docker System
```bash
# WARNING: Removes all unused containers, networks, images
docker system prune -a --volumes
```

---

### Problem: Port Conflict

**Error**:
```
ERROR: Port 5432 is already allocated
```

**Diagnosis**:
```bash
# Find process using port
lsof -i :5432
```

**Solution**:
```bash
# Kill process
kill -9 <PID>

# Or change port in docker-compose.dev.yml
ports:
  - "5433:5432"  # Use 5433 instead
```

---

## 7. Testing Problems

### Problem: Tests Fail Randomly

**Symptoms**:
- Tests pass locally, fail in CI
- Tests fail intermittently

**Common Causes**:

#### Cause 1: Race Condition

**Solution**: Run with race detector
```bash
go test -race ./...
```

#### Cause 2: Shared State

**Bad Test**:
```go
// BAD: Shared global variable
var testDB *sqlx.DB

func TestA(t *testing.T) {
  testDB.Exec("INSERT ...")
}

func TestB(t *testing.T) {
  testDB.Exec("INSERT ...")  // Conflict!
}
```

**Good Test**:
```go
// GOOD: Isolated database per test
func TestA(t *testing.T) {
  db := integration.SetupTestDB(t)
  defer db.Close()
  db.Exec("INSERT ...")
}
```

---

### Problem: Test Database Connection Failed

**Error**:
```
ERROR: Failed to connect to test database
```

**Check**:
```bash
# Verify test DB running
docker ps | grep promenade_test_postgres

# Start test database
make test-db-start
```

---

## 8. Development Workflow

### Problem: Go Module Issues

**Error**:
```
go: module requires Go 1.24 or later
```

**Solution**:
```bash
# Check Go version
go version

# Update Go (macOS)
brew upgrade go

# Update Go (Linux)
sudo snap refresh go --classic
```

---

### Problem: Import Path Errors

**Error**:
```
cannot find package "github.com/basilex/promenade/pkg/uuidv7"
```

**Solution**:
```bash
# Tidy modules
go mod tidy

# Download dependencies
go mod download

# Verify modules
go mod verify
```

---

### Problem: Git Push Rejected

**Error**:
```
! [rejected] dev -> dev (non-fast-forward)
```

**Solution**:
```bash
# Pull latest changes
git pull origin dev

# Resolve conflicts
git add .
git commit -m "Merge conflicts resolved"

# Push again
git push origin dev
```

---

## Quick Troubleshooting Checklist

### When API Request Fails

1.  Check server is running: `curl http://localhost:8081/health`
2.  Verify database connection: `curl http://localhost:8081/health/db`
3.  Check token validity: `echo $TOKEN`
4.  Verify endpoint URL: Correct path and method?
5.  Check request body: Valid JSON format?
6.  Review server logs: `docker logs promenade_api`
7.  Check rate limits: Wait 60 seconds
8.  Verify user permissions: `GET /users/me`

### When Tests Fail

1.  Run `make test-unit` first (fastest)
2.  Check test database: `make test-db-start`
3.  Clean test cache: `go clean -testcache`
4.  Run with race detector: `go test -race ./...`
5.  Check CI logs on GitHub Actions
6.  Verify workspace config: `make workspace`

### When Server Won't Start

1.  Check workspace config: `make workspace`
2.  Verify database running: `make docker-ps`
3.  Check port availability: `lsof -i :8081`
4.  Review config file: `cat config/app.postgres-dev.yaml`
5.  Check migrations: `make migrate-status`
6.  Try fresh start: `make dev-fresh`

---

## Getting Help

### Documentation

- **[Quick Start Guide](quick-start.md)** - Getting started
- **[Authentication Flow](authentication-flow.md)** - JWT issues
- **[Common Use Cases](common-use-cases.md)** - API workflows
- **[API Reference](../reference/api-reference.md)** - Complete API docs
- **[Testing Guide](../reference/testing-patterns.md)** - Testing patterns

### Community

- **GitHub Issues**: [github.com/basilex/promenade/issues](https://github.com/basilex/promenade/issues)
- **Discussions**: [github.com/basilex/promenade/discussions](https://github.com/basilex/promenade/discussions)
- **Email**: alexander.vasilenko@gmail.com

### Reporting Bugs

Include:
1. **Environment**: OS, Go version, database driver
2. **Steps to reproduce**: Exact commands/API calls
3. **Expected behavior**: What should happen
4. **Actual behavior**: What actually happened
5. **Logs**: Server logs, error messages, stack traces
6. **Config**: Relevant config file sections (redact secrets!)

---

## Summary

### Top 5 Issues

1. **401 Unauthorized** → Check token, refresh if expired
2. **403 Forbidden** → Verify user roles and permissions
3. **404 Not Found** → Check resource ID and endpoint URL
4. **429 Rate Limit** → Wait 60 seconds, add delays in scripts
5. **500 Internal Error** → Check server logs, report bug

### Prevention Tips

- Always use HTTPS in production
- Store tokens securely (not in localStorage)
- Add timeouts to all requests
- Paginate large result sets
- Index frequently queried columns
- Monitor server metrics
- Run tests before pushing
- Use environment-specific configs

---

**Version**: 0.1.0  
**Last Updated**: January 5, 2026  
**Status**: Production-ready
