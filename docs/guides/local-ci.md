# Local CI Validation

**Run GitHub Actions checks locally** before pushing to avoid failed builds and save CI resources.

---

## Problem

GitHub Actions builds take time and resources. Failed linting or tests waste:
- **Time**: 5-10 minutes per build
- **Resources**: GitHub Actions minutes
- **Productivity**: Context switching while waiting

---

## Solution

**Makefile targets** that replicate exact GitHub Actions workflow locally.

---

## Available Commands

### ci-check

Run **all CI checks** (lint + test + build):

```bash
make ci-check
```

This is equivalent to running:
```bash
make ci-lint && make ci-test && make ci-build
```

**Use before**: `git push` to ensure CI will pass

---

### ci-lint

Run **golangci-lint** with same configuration as CI:

```bash
make ci-lint
```

**Checks**:
- errcheck: Unchecked error returns
- staticcheck: Code quality issues
- unused: Dead code
- gosimple: Simplification suggestions
- And 40+ other linters

**Timeout**: 5 minutes (same as CI)

---

### ci-test

Run **all tests** with race detector:

```bash
make ci-test
```

**Runs**:
1. `make test-unit` - Unit tests (~5s)
2. `make test-integration` - Integration tests with DB (~14s)
3. `go test -race ./...` - Race detector (~40s)

---

### ci-build

Test **compilation** without running:

```bash
make ci-build
```

Verifies code builds successfully (catches import/syntax errors).

---

### pre-push

**Alias for ci-check**:

```bash
make pre-push
```

Mnemonic command to run before `git push`.

---

## Workflow

### 1. Development

```bash
# Make code changes
vim internal/contexts/customer-mgmt/interaction/entity.go

# Run quick unit tests
make test-unit
```

### 2. Before Commit

```bash
# Run lint to catch issues early
make ci-lint
```

**If lint fails**, fix issues before committing:
```bash
# Example: Add error handling
_ = db.Close()  # Instead of: db.Close()
```

### 3. Before Push

```bash
# Run full CI check
make pre-push
```

**Output**:
```
🔍 Running golangci-lint...
✅ Lint passed

🧪 Running unit tests...
✅ Unit tests passed

🧪 Running integration tests...
✅ Integration tests passed

🧪 Running race detector...
✅ Race detector passed

🔨 Testing build...
✅ Build passed

🎉 All CI checks passed! Safe to push.
```

### 4. Push

```bash
git push origin dev
```

CI will pass because you already validated locally! ✅

---

## Common Issues

### Issue: Lint finds 21 errors

**Problem**:
```bash
make ci-lint
# Returns: 21 issues (14 errcheck, 4 staticcheck, 3 unused)
```

**Solution**: Fix systematically by category:

#### errcheck (unchecked errors)

```go
// ❌ Bad
db.Close()

// ✅ Good
_ = db.Close()  // Explicit ignore
// or
if err := db.Close(); err != nil {
    log.Error("Failed to close DB", err)
}
```

#### staticcheck (code simplification)

```go
// ❌ Bad (QF1008: unnecessary embedded field selector)
user, err := row.userRow.toEntity()

// ✅ Good
user, err := row.toEntity()
```

#### unused (dead code)

```go
// ❌ Bad (function never used)
func parseOptionalUUID(s *string) (*uuidv7.UUID, error) {
    // ... implementation
}

// ✅ Good - delete the function entirely
```

---

### Issue: Integration tests fail

**Problem**:
```bash
make ci-test
# Integration tests fail: connection refused
```

**Solution**: Start test database:

```bash
make test-db-start
make ci-test
```

Or use automatic DB management:
```bash
make test-integration  # Starts DB automatically
```

---

### Issue: Lint hangs/times out

**Problem**:
```bash
make ci-lint
# (hangs for > 5 minutes)
```

**Solutions**:

1. **Clear cache**:
   ```bash
   golangci-lint cache clean
   make ci-lint
   ```

2. **Update golangci-lint**:
   ```bash
   brew upgrade golangci-lint
   ```

3. **Run specific linters**:
   ```bash
   golangci-lint run --disable-all --enable=errcheck
   ```

---

## Git Hooks (Optional)

### Pre-commit Hook

**Automatically run lint** before every commit:

```bash
# .git/hooks/pre-commit
#!/bin/bash
make ci-lint || exit 1
```

```bash
chmod +x .git/hooks/pre-commit
```

### Pre-push Hook

**Automatically run full CI** before every push:

```bash
# .git/hooks/pre-push
#!/bin/bash
make ci-check || exit 1
```

```bash
chmod +x .git/hooks/pre-push
```

---

## CI Comparison

| Check Type     | Local Command       | GitHub Actions Job | Duration |
|----------------|---------------------|-------------------|----------|
| **Lint**       | `make ci-lint`      | `lint` job        | ~30s     |
| **Unit Tests** | `make test-unit`    | `test` job        | ~5s      |
| **Integration**| `make test-integration` | `test` job    | ~14s     |
| **Race Detector** | `go test -race ./...` | `test` job   | ~40s     |
| **Build**      | `make ci-build`     | `build` job       | ~10s     |
| **Total**      | `make ci-check`     | All jobs          | ~60s     |

**GitHub Actions overhead**: +3-5 minutes (queue time, VM startup)

---

## Implementation Details

### Makefile.dev.mk

```makefile
ci-check: ci-lint ci-test ci-build  ## Run all CI checks locally (lint + test + build)

ci-lint:  ## Run linters (same as CI)
	@echo "🔍 Running golangci-lint..."
	@golangci-lint run --timeout=5m || (echo "❌ Lint failed" && exit 1)
	@echo "✅ Lint passed"

ci-test:  ## Run all tests (same as CI)
	@echo "🧪 Running unit tests..."
	@make test-unit || (echo "❌ Unit tests failed" && exit 1)
	@echo "✅ Unit tests passed"
	@echo ""
	@echo "🧪 Running integration tests..."
	@make test-integration || (echo "❌ Integration tests failed" && exit 1)
	@echo "✅ Integration tests passed"
	@echo ""
	@echo "🧪 Running race detector..."
	@go test -race ./... > /dev/null 2>&1 || (echo "❌ Race detector failed" && exit 1)
	@echo "✅ Race detector passed"

ci-build:  ## Test build (same as CI)
	@echo "🔨 Testing build..."
	@make build > /dev/null || (echo "❌ Build failed" && exit 1)
	@echo "✅ Build passed"

pre-push: ci-check  ## Alias for ci-check (run before git push)
	@echo ""
	@echo "🎉 All CI checks passed! Safe to push."
```

---

## Statistics

**Promenade Local CI** (after lint fixes):

| Metric                 | Value           |
|------------------------|-----------------|
| **Total Tests**        | 240+ tests      |
| **Lint Issues Fixed**  | 26 issues       |
| **golangci-lint**      | 0 issues ✅     |
| **Local CI Duration**  | ~60s            |
| **GitHub CI Duration** | ~5min (with VM) |
| **Time Saved**         | 4min per push   |

**ROI**: 10 failed pushes = 40 minutes saved

---

## Best Practices

### DO

✅ **Run `make ci-lint`** after every significant change  
✅ **Run `make pre-push`** before every push  
✅ **Fix lint issues** immediately (don't accumulate)  
✅ **Use `_ = err`** for intentional error ignores  
✅ **Add comments** when ignoring errors (`// Safe to ignore`)

### DON'T

❌ **Don't push without running** `make ci-check`  
❌ **Don't ignore lint warnings** ("I'll fix later")  
❌ **Don't disable linters** without good reason  
❌ **Don't commit commented code** (triggers unused)  
❌ **Don't skip tests** to "save time"

---

## Troubleshooting

### Lint passes locally but fails on CI

**Causes**:
1. **Different golangci-lint version**
2. **Different Go version**
3. **OS-specific issues**

**Solution**:
```bash
# Check versions match CI
go version  # Should be 1.24.0
golangci-lint version  # Should be 1.64.8+
```

---

### Tests pass locally but fail on CI

**Causes**:
1. **Race conditions** (use `-race` flag)
2. **Database state** (CI has clean DB)
3. **Timing issues** (CI slower)

**Solution**:
```bash
# Run with race detector
go test -race ./...

# Clean database before tests
make test-db-start
make test-integration
```

---

## Related Documentation

- [Testing Guide](../test/README.md) - Testing structure
- [Testing Patterns](testing-patterns.md) - Comprehensive testing guide
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Contribution workflow

---

**Last Updated**: 2025-12-29  
**Status**: Production-ready  
**Lint Issues Fixed**: 26 (errcheck: 17, staticcheck: 6, unused: 3)  
**Maintainer**: Promenade Team
