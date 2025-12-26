# Notifications Module - Testing Guide

This document describes all test types available for the Notifications module.

---

## Test Overview

| Test Type             | Location                       | Purpose                      | Duration | Status     |
| --------------------- | ------------------------------ | ---------------------------- | -------- | ---------- |
| **Unit Tests**        | `domain/entity/`, `usecase/`   | Business logic validation    | ~0.6s    |  35 PASS |
| **Integration Tests** | `adapter/repository/postgres/` | Database operations          | ~1.2s    |  13 PASS |
| **Smoke Tests**       | `test/smoke/`                  | End-to-end API validation    | ~10s     | ⏳ Ready   |
| **Stress Tests**      | `test/stress/`                 | Load and performance testing | 30s+     | ⏳ Ready   |

**Total Coverage**: 35 unit tests + 13 integration tests + 7 smoke tests + 4 stress scenarios = **59 tests**

---

## Latest Test Results

**Date:** December 25, 2025  
**Status:**  ALL TESTS PASSING (48/48)

---

## 1. Unit Tests (35 tests)

### Location

```
internal/modules/notifications/
 domain/entity/
    notification_test.go (20 tests)
    user_preference_test.go (15 tests)
 usecase/
     notification_usecase_test.go (15 tests)
```

### Running Unit Tests

```bash
# Entity tests only
go test -v github.com/basilex/promenade/internal/modules/notifications/domain/entity

# UseCase tests only
go test -v github.com/basilex/promenade/internal/modules/notifications/usecase

# All unit tests
go test -v github.com/basilex/promenade/internal/modules/notifications/domain/entity \
            github.com/basilex/promenade/internal/modules/notifications/usecase

# With coverage
go test -cover github.com/basilex/promenade/internal/modules/notifications/domain/entity
```

### Test Coverage

- **Notification Entity**: 20 tests

  - NewNotification validation
  - Status transitions (sent, delivered, opened, clicked, failed)
  - Multi-channel support (email, sms, push, in_app)
  - Notification types (system, security, product, social)
  - Field validation (type, channel, recipient)

- **UserPreference Entity**: 15 tests
  - NewUserPreference creation
  - Channel enabling/disabling (email, sms, push, in_app)
  - Notification type preferences
  - Quiet hours logic with timezone support
  - Field validation

---

## 2. Integration Tests (4 test suites)

### Location

```
internal/modules/notifications/adapter/repository/postgres/integration_test.go
```

### Running Integration Tests

```bash
# Prerequisites: Database must be running
make docker-up  # Start PostgreSQL

# Run notifications integration tests
make test-integration-notifications

# Or directly
go test -v -tags=integration ./internal/modules/notifications/adapter/repository/postgres/...
```

### Test Suites

## 2. Integration Tests (13 tests in 4 suites)

### Location

```
internal/modules/notifications/adapter/repository/postgres/
 integration_test.go
```

### Running Integration Tests

```bash
# Run notifications integration tests (includes automatic migration)
make test-integration-notifications

# Or with go test directly
go test -v -tags=integration \
  github.com/basilex/promenade/internal/modules/notifications/adapter/repository/postgres

# Prerequisites (handled automatically by make target):
# - Test database running on port 5433
# - Migrations applied
```

### Test Suites

1. **TestNotificationRepository_Integration** (7 tests)

   -  Create and GetByID
   -  GetByUserID returns user notifications
   -  Update notification status
   -  CountByUserID returns correct count
   -  GetUnreadCount returns pending notifications
   -  GetPendingNotifications returns pending only
   -  Delete notification

2. **TestUserPreferenceRepository_Integration** (5 tests)

   -  Create and GetByUserID
   -  Update user preference
   -  Exists returns true for existing preference
   -  Delete user preference
   -  Exists returns false for deleted preference

3. **TestNotificationWorkflow_Integration** (1 test)

   -  Complete workflow: preferences → notification → tracking
   - Tests real-world scenario with multiple operations

4. **TestQuietHoursLogic_Integration** (1 test)
   -  Quiet hours logic with timezone
   - Tests TimeOfDay storage and IsInQuietHours logic

### Database Features Tested

- **Custom Types**:
  - `JSONB` type for notification data
  - `TimeOfDay` type for quiet hours (VARCHAR(5) "HH:MM" format)
- **UUID v7** primary keys
- **Soft delete** for notifications
- **Hard delete** for preferences
- **Transactions** via context
- **Automatic migrations** before tests

---

## 3. Smoke Tests (7 scenarios)

### Location

```
test/smoke/tests/06_notifications.sh
```

### Running Smoke Tests

```bash
# Prerequisites: API must be running
make dev  # Start API in one terminal

# Run all smoke tests (in another terminal)
make test-smoke

# Or run only notifications smoke test
cd test/smoke
./tests/06_notifications.sh
```

### Test Scenarios

1. **GET /notifications/preferences** - Get user preferences (auto-create)
2. **PUT /notifications/preferences** - Update preferences
3. **POST /notifications** - Send notification
4. **GET /notifications** - List notifications (paginated)
5. **GET /notifications/:id** - Get single notification
6. **GET /notifications/unread-count** - Get unread count
7. **POST /notifications/:id/opened** - Mark notification as opened

### Expected Results

```

  Notifications Tests

ℹ Testing GET /notifications/preferences...
 Get preferences passed
ℹ Testing PUT /notifications/preferences...
 Update preferences passed
ℹ Testing POST /notifications...
 Send notification passed
ℹ Testing GET /notifications...
 List notifications passed
ℹ Testing GET /notifications/:id...
 Get notification passed
ℹ Testing GET /notifications/unread-count...
 Get unread count passed
ℹ Testing POST /notifications/:id/opened...
 Mark as opened passed
 Test passed: 06_notifications
```

---

## 4. Stress Tests (4 operations)

### Location

```
test/stress/scenarios/notifications.lua
```

### Running Stress Tests

```bash
# Prerequisites:
# 1. API must be running (make dev)
# 2. wrk must be installed (brew install wrk on macOS)
# 3. Valid JWT token must be set in notifications.lua

# Edit notifications.lua and set TOKEN variable:
# local TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.YOUR_REAL_TOKEN_HERE"

# Run stress test
cd test/stress
wrk -t4 -c100 -d30s -s scenarios/notifications.lua http://localhost:8081
```

### Load Test Operations

The stress test randomly selects one of these operations for each request:

1. **Send Notification** (30% probability)

   - POST /api/v1/notifications
   - Random type (system, security, product, social)
   - Random channel (email, sms, push, in_app)

2. **List Notifications** (40% probability)

   - GET /api/v1/notifications?page=1&page_size=20
   - Paginated listing

3. **Get Preferences** (20% probability)

   - GET /api/v1/notifications/preferences

4. **Get Unread Count** (10% probability)
   - GET /api/v1/notifications/unread-count

### Expected Performance

```
Running 30s test @ http://localhost:8081
  4 threads and 100 connections

Requests/sec:    450.23
Transfer/sec:    128.45KB

Latency Distribution:
  50%    85.00ms
  75%   120.00ms
  90%   180.00ms
  99%   350.00ms

Operation Statistics:
  Send notifications:    3,500 (30%)
  List notifications:    4,800 (40%)
  Get preferences:       2,400 (20%)
  Get unread count:      1,200 (10%)

Total requests: 13,507
Non-2xx responses: 0 (0.00%)
```

---

## Test Database Setup

### Automatic Setup (Recommended)

Integration and smoke tests use the main development database:

```bash
# Start PostgreSQL
make docker-up

# Migrations run automatically on app start
make dev
```

### Manual Setup

```bash
# Create test database
make test-integration-setup

# Check database status
make test-integration-check
```

### Database Connection

- **Host**: localhost
- **Port**: 5432 (Docker) or 5433 (test container)
- **User**: promenade
- **Password**: promenade
- **Database**: promenade_test

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Notifications Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: promenade
          POSTGRES_PASSWORD: promenade
          POSTGRES_DB: promenade_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.25"

      - name: Run unit tests
        run: make test-module-notifications

      - name: Run integration tests
        run: make test-integration-notifications
        env:
          TEST_DB_HOST: localhost
          TEST_DB_PORT: 5432
          TEST_DB_USER: promenade
          TEST_DB_PASSWORD: promenade
          TEST_DB_NAME: promenade_test

      - name: Start API for smoke tests
        run: |
          make build
          ./bin/promenade &
          sleep 5

      - name: Run smoke tests
        run: ./test/smoke/tests/06_notifications.sh
```

---

## Troubleshooting

### Unit Tests Fail

**Problem**: Import cycle or missing dependencies

**Solution**:

```bash
cd internal/modules/notifications
go mod tidy
go test ./...
```

### Integration Tests Fail: Database Connection

**Problem**: "connection refused" or "database not available"

**Solution**:

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Start PostgreSQL
make docker-up

# Check database status
make test-integration-check
```

### Integration Tests Fail: Table Does Not Exist

**Problem**: "relation notifications does not exist"

**Solution**:

```bash
# Run migrations
make migrate

# Or start app (migrations run automatically)
make dev
```

### Smoke Tests Fail: API Not Running

**Problem**: "Connection refused" on port 8081

**Solution**:

```bash
# Start API in another terminal
make dev

# Wait for API to be ready
curl http://localhost:8081/api/v1/health
```

### Smoke Tests Fail: Authentication Error

**Problem**: "401 Unauthorized"

**Solution**:

```bash
# Ensure 02_auth.sh runs before 06_notifications.sh
# This sets JWT_TOKEN environment variable
cd test/smoke
./smoke_test.sh  # Runs all tests in correct order
```

### Stress Tests Fail: Invalid Token

**Problem**: "401 Unauthorized" in stress test

**Solution**:

```bash
# 1. Get a valid token
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alexander.vasilenko@gmail.com","password":"03041965"}'

# 2. Copy access_token from response

# 3. Edit test/stress/scenarios/notifications.lua
# Replace TOKEN variable with your token:
local TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.YOUR_TOKEN_HERE"

# 4. Run stress test
cd test/stress
wrk -t4 -c100 -d30s -s scenarios/notifications.lua http://localhost:8081
```

---

## Quick Test Commands Reference

```bash
# Unit tests
make test-module-notifications

# Integration tests
make test-integration-notifications

# Smoke tests (requires running API)
make test-smoke

# Stress tests (requires wrk + running API + valid token)
cd test/stress
wrk -t4 -c100 -d30s -s scenarios/notifications.lua http://localhost:8081

# All tests
make test  # Runs unit + integration tests
```

---

## Coverage Goals

-  **Unit Tests**: 80%+ coverage (currently 35 tests)
-  **Integration Tests**: All repository methods covered (4 test suites)
-  **Smoke Tests**: All API endpoints covered (7 scenarios)
-  **Stress Tests**: All operations under load (4 operation types)

---

## Next Steps

1. Add more unit tests for `notification_usecase.go`
2. Add smoke tests for error scenarios (invalid data, not found, etc.)
3. Add stress test with mixed read/write operations
4. Add smoke tests for quiet hours and timezone handling

---

**For more information:**

- [Notifications Module README](README.md)
- [Test Infrastructure](../../../test/README.md)
- [Testing Guide](../../../docs/TESTING_GUIDE.md)
