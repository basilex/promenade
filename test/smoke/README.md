# Smoke Tests

Smoke tests verify that critical API functionality works correctly in a running environment. These are **black-box integration tests** that make real HTTP requests to the API.

## What Are Smoke Tests?

Smoke tests are quick, essential tests that verify the application's core functionality:

- ✅ Can the API start?
- ✅ Are endpoints accessible?
- ✅ Do critical workflows work end-to-end?
- ✅ Is the database connected?

**Not covered by smoke tests:**

- ❌ Edge cases
- ❌ Performance/load testing
- ❌ Unit test logic
- ❌ Error handling details

## Quick Start

### Prerequisites

- **curl** - HTTP client
- **jq** - JSON processor
- Running API instance (port 8081 by default)

```bash
# Check if tools are installed
which curl && which jq
```

### Running Tests

```bash
# 1. Start API (if not already running)
make dev

# 2. Run smoke tests (in another terminal)
make test-smoke

# Or run directly
cd test/smoke
./smoke_test.sh
```

### Expected Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   Promenade API - Smoke Tests
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  API:         http://localhost:8081
  Environment: dev
  Date:        2025-12-25 14:30:00

ℹ Checking required tools...
✓ All required tools found
ℹ Waiting for API to be ready...
✓ API is ready!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Running Smoke Tests
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Health Check
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ℹ Testing GET /health...
✓ Health check passed
✓ Test passed: 01_health

...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Test Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Total tests:  5
  Passed:       5
  Failed:       0

✓ All smoke tests passed!
```

## Test Structure

### Tests (in order)

1. **01_health.sh** - Health check endpoint
2. **02_auth.sh** - Authentication flow (register → login → get me → logout)
3. **03_posts.sh** - Posts CRUD operations
4. **04_profiles.sh** - Profiles CRUD operations
5. **05_analytics.sh** - Analytics metrics

### Configuration

**config.sh** - Test configuration:

- API URL (default: `http://localhost:8081`)
- Test credentials
- HTTP timeouts

**helpers.sh** - Helper functions:

- HTTP wrappers (GET, POST, PUT, DELETE)
- JSON assertions
- Colored output
- Cleanup utilities

## Configuration

### Environment Variables

```bash
# API URL (default: http://localhost:8081)
export PROMENADE_API_URL="http://localhost:8081"

# Environment name (optional)
export PROMENADE_ENV="dev"
```

### Custom API URL

```bash
# Test against staging
PROMENADE_API_URL="https://staging.example.com" ./smoke_test.sh

# Test against production
PROMENADE_API_URL="https://api.example.com" ./smoke_test.sh
```

## Running Individual Tests

Each test can be run independently:

```bash
cd test/smoke

# Run single test
./tests/01_health.sh

# Run specific tests
./tests/02_auth.sh
./tests/03_posts.sh
```

**Note:** Tests 03-05 require authentication, so test 02 will run automatically if needed.

## Writing New Tests

### Template

```bash
#!/bin/bash

# Smoke Test: My Feature

source "$(dirname "$0")/../helpers.sh"

test_my_feature() {
    print_section "My Feature"

    # Test logic here
    print_info "Testing GET /my-endpoint..."
    local response=$(http_get "$API_BASE/my-endpoint" 200)

    if [ $? -ne 0 ]; then
        print_error "Test failed"
        return 1
    fi

    assert_json_field "$response" ".status" "success" || return 1

    print_success "Test passed"
    return 0
}

# Run test
test_my_feature
exit $?
```

### Add to Runner

Edit `smoke_test.sh`:

```bash
run_test "tests/06_my_feature.sh"
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Smoke Tests

on: [push, pull_request]

jobs:
  smoke-tests:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.25"

      - name: Install dependencies
        run: |
          sudo apt-get update
          sudo apt-get install -y jq

      - name: Run migrations
        run: make migrate

      - name: Build API
        run: make build

      - name: Start API
        run: ./bin/promenade &
        env:
          ENVIRONMENT: test

      - name: Wait for API
        run: sleep 5

      - name: Run smoke tests
        run: make test-smoke

      - name: Stop API
        run: pkill promenade
```

## Troubleshooting

### API Not Ready

```bash
# Check if API is running
curl http://localhost:8081/api/v1/health

# Check logs
tail -f /path/to/logs/promenade.log
```

### Test Failures

```bash
# Run with verbose output
bash -x ./test/smoke/tests/01_health.sh

# Check individual requests
curl -v http://localhost:8081/api/v1/health
```

### Permission Denied

```bash
# Make scripts executable
chmod +x test/smoke/*.sh
chmod +x test/smoke/tests/*.sh
```

### Missing Dependencies

```bash
# macOS
brew install curl jq

# Ubuntu/Debian
apt-get install curl jq

# CentOS/RHEL
yum install curl jq
```

## Best Practices

1. **Keep tests fast** - Smoke tests should run in < 2 minutes
2. **Test critical paths only** - Not every endpoint needs a smoke test
3. **Use unique test data** - Timestamps in test emails/titles prevent conflicts
4. **Clean up after tests** - Delete test data to avoid pollution
5. **Make tests idempotent** - Should pass even if run multiple times
6. **Use assertions** - Verify response structure, not just status codes
7. **Fail fast** - Stop on first failure to save time

## What's Next?

After smoke tests pass, consider:

- **Load Testing** - See `test/stress/` for wrk-based stress tests
- **End-to-End Tests** - Cypress/Playwright for UI testing
- **Security Testing** - OWASP ZAP, Burp Suite
- **Chaos Testing** - Simulate failures

## Support

- **Issues** - Report bugs in smoke tests
- **Questions** - Ask in team chat
- **Improvements** - Submit PR with new tests

---

**Created:** December 25, 2025
**Maintained by:** Promenade DevOps Team
