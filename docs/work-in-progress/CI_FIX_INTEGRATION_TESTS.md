# CI Fix: Integration Tests Skip in Short Mode

**Date**: December 29, 2025  
**Commit**: e254fd4  
**Status**: ✅ Fixed

## Problem

GitHub Actions CI was failing during `make test-unit` step with error:

```
Failed to connect to test database: dial tcp [::1]:5433: connection refused
FAIL
make: *** [Makefile.test.mk:13: test-unit] Error 1
Error: Process completed with exit code 2.
```

**Root Cause**: Integration tests were running during `go test -v -short ./...` because they didn't check `testing.Short()` flag.

## Solution

Added `testing.Short()` skip check to all integration test functions:

```go
func TestCustomerRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	// ... rest of test
}
```

**Files Modified**: 10 integration test files
- `test/integration/contexts/customer-mgmt/customer/repository_test.go`
- `test/integration/contexts/identity/contact/repository_test.go`
- `test/integration/contexts/identity/permission/repository_test.go`
- `test/integration/contexts/identity/profile/repository_test.go`
- `test/integration/contexts/identity/role/repository_test.go`
- `test/integration/contexts/identity/user/repository_test.go`
- `test/integration/contexts/shared/country/repository_test.go`
- `test/integration/contexts/shared/currency/repository_test.go`
- `test/integration/contexts/shared/language/repository_test.go`
- `test/integration/contexts/shared/timezone/repository_test.go`

**Tool Created**: `scripts/add_integration_test_skip.py` - Python script to automate adding skip checks

## Verification

```bash
# Unit tests (with short flag) - all pass
make test-unit
# Result: 54 packages PASS, 0 FAIL

# Integration tests (without short flag) - all pass
make test-integration
# Result: All integration tests run with real DB
```

## CI Pipeline Flow

| Step                | Command               | Tests Run                          |
|---------------------|----------------------|------------------------------------|
| **test-unit**       | `go test -short ./...` | Unit + Smoke (no integration)      |
| **test-smoke**      | Custom command       | Smoke only (mock-based)            |
| **test-integration**| Custom command       | Integration only (with test DB)    |

## Impact

- ✅ CI now passes on `make test-unit` step
- ✅ Unit tests run in ~5 seconds
- ✅ Integration tests only run when explicitly requested
- ✅ No false positives from missing test database

## Related Documentation

- [Testing Guide](../test/README.md)
- [Testing Patterns](../docs/guides/testing-patterns.md)
- [Makefile.test.mk](../Makefile.test.mk)

---

**Last Updated**: December 29, 2025
