# Local CI Validation

**Run GitHub Actions checks locally** before pushing to avoid failed builds and save time.

---

## Overview

GitHub Actions builds take time and resources. Failed linting or tests waste:
- **Time**: 5-10 minutes per build
- **Resources**: GitHub Actions minutes
- **Productivity**: Context switching while waiting

**Solution**: Run the same checks locally with Makefile targets before pushing.

---

## Quick Start

```bash
# Before every push (REQUIRED)
make pre-push

# Or run checks individually
make ci-lint        # golangci-lint (0 issues)
make ci-test        # All tests + race detector
make ci-build       # Test compilation
```

**Result**: ✅ All CI checks pass locally → push with confidence

---

## Available Commands

### make pre-push

Run **all CI checks** (lint + test + build):

```bash
make pre-push
```

**Use before**: Every `git push` to ensure CI will pass

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

---

### make ci-lint

Run **golangci-lint** with same configuration as CI:

```bash
make ci-lint
```

**Checks**:
- errcheck: Unchecked error returns
- staticcheck: Code quality issues
- unused: Dead code
- gosimple: Simplification suggestions
- 40+ other linters

**Timeout**: 5 minutes (same as GitHub Actions)

---

### make ci-test

Run **all tests** with race detector:

```bash
make ci-test
```

**Runs**:
1. Unit tests (~5s)
2. Integration tests with DB (~14s)
3. Race detector (~40s)

---

### make ci-build

Test **compilation** without running:

```bash
make ci-build
```

Verifies code builds successfully (catches import/syntax errors).

---

## Recommended Workflow

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

**If lint fails**, fix issues before committing.

### 3. Before Push

```bash
# Run full CI check
make pre-push
```

### 4. Push with Confidence

```bash
git push origin dev
```

CI will pass because you already validated locally! ✅

---

## Common Lint Fixes

### errcheck (unchecked errors)

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

### staticcheck (code simplification)

```go
// ❌ Bad (QF1008: unnecessary embedded field selector)
user, err := row.userRow.toEntity()

// ✅ Good
user, err := row.toEntity()
```

### unused (dead code)

```go
// ❌ Bad (function never used)
func parseOptionalUUID(s *string) (*uuidv7.UUID, error) {
    // ... implementation
}

// ✅ Good - delete the function entirely
```

---

## Troubleshooting

### Lint Hangs/Timeouts

**Solution**: Clear cache and update:

```bash
golangci-lint cache clean
brew upgrade golangci-lint
make ci-lint
```

### Integration Tests Fail

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

## Git Hooks (Optional)

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

Now every `git push` runs validation automatically!

---

## CI Comparison

| Check Type     | Local Command       | GitHub Actions | Duration |
|----------------|---------------------|----------------|----------|
| **Lint**       | `make ci-lint`      | `lint` job     | ~30s     |
| **Unit Tests** | `make test-unit`    | `test` job     | ~5s      |
| **Integration**| `make test-integration` | `test` job | ~14s     |
| **Race Detector** | `go test -race ./...` | `test` job | ~40s  |
| **Build**      | `make ci-build`     | `build` job    | ~10s     |
| **TOTAL**      | `make pre-push`     | All jobs       | ~1m40s   |

**Time Savings**: 4+ minutes per failed push by catching issues early

---

## Benefits

✅ **Faster Feedback**: Catch issues in seconds, not minutes  
✅ **Save Resources**: Reduce GitHub Actions usage  
✅ **Better Code Quality**: Lint catches issues before review  
✅ **Productivity**: No context switching while waiting for CI  
✅ **Confidence**: Push knowing CI will pass

---

## Related Documentation

- [Testing Patterns](/guide/testing-patterns) - Four-tier testing strategy
- [Development Workflow](/guide/development-workflow) - Daily development process
- [Contributing](/guide/contributing) - How to contribute

---

**Version**: 1.0.0  
**Status**: Production-ready  
**Maintainer**: Promenade Team
