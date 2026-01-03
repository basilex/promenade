# Link Audit Report - January 3, 2026

## Summary

Comprehensive check of all markdown links in Promenade Platform documentation.

**Date**: January 3, 2026  
**Tool**: scripts/check-links.py  
**Files Checked**: 120+ markdown files  
**Broken Links Found**: 200+

---

## Categories of Issues

### 1. Work-in-Progress Files (High Priority)

**Files with most broken links:**

- ~~`docs/work-in-progress/INDEX_OLD.md`~~ - **REMOVED** ✅
- `docs/work-in-progress/GAPS_AND_TODOS_OLD.md`
- `docs/work-in-progress/DB_AGNOSTIC_REFACTORING.md`
- `docs/work-in-progress/README.md`
- `docs/work-in-progress/CI_FIX_INTEGRATION_TESTS.md`

**Action**: Review and either fix or archive these files.

---

### 2. Documentation Structure Changes

Many links reference old documentation structure:

**Old Paths** (no longer exist):
- `docs/CLEAN_ARCHITECTURE_SUMMARY.md` → `docs/concepts/clean-architecture.md`
- `docs/TESTING_PATTERNS.md` → `docs/guides/testing-patterns.md`
- `docs/RBAC.md` → `docs/guides/rbac.md`
- `docs/RATE_LIMITING.md` → `docs/guides/rate-limiting.md`
- `docs/HEALTH_CHECKS.md` → `docs/guides/health-checks.md`
- `docs/BUS_TEST_COVERAGE.md` → `docs/reference/bus-test-coverage.md`
- `docs/README.md` → `docs/INDEX.md`

**Affected Files**: ~40 files

**Action**: Update all references to new paths.

---

### 3. Website Internal Links (Medium Priority)

Website uses VitePress routing (`/guide/`, `/concepts/`, `/packages/`):

**Pattern**: `[Text](/guide/page)` or `[Text](/concepts/page)`

**Issues**:
- Links don't resolve during markdown link check (expected)
- These are VitePress routes, not file paths
- Valid at runtime but fail static validation

**Affected Files**: All `website/*.md` files (~50 files)

**Action**: Create separate validation for VitePress routes or exclude from check.

---

### 4. False Positives

**Generic Type Parameters** incorrectly detected as links:

```markdown
[T any] - detected as link to "T any"
```

**Examples**:
- `website/packages/response.md`: `[T any]` in function signatures
- `docs/work-in-progress/DB_AGNOSTIC_REFACTORING.md`: `value T [T any]`

**Affected Files**: 5 files

**Action**: Improve regex in check-links.py to exclude `[T any]` patterns.

---

### 5. Missing Referenced Files

**Non-existent files referenced**:

1. `ORDER_MGMT_VERIFICATION.md` (root) - Referenced in `docs/concepts/order-management.md`
2. `docs/concepts/cqrs.md` - Referenced in `internal/contexts/customer-mgmt/analytics/README.md`
3. `docs/reference/api-reference.md` - Referenced in multiple places
4. `website/reference/configuration.md` - Referenced in `website/guide/production-deployment.md`
5. `website/guide/webhooks.md` - Referenced in `website/reference/api-reference.md`

**Action**: Create these missing files or remove references.

---

## Quick Fix Plan

### Phase 1: Immediate Fixes ✅

- [x] Remove `docs/work-in-progress/INDEX_OLD.md`
- [x] Sync `README.md` with `website/index.md`
- [x] Create link checker scripts

### Phase 2: High Priority (Next)

- [ ] Update all `docs/` files to use new structure paths
  - Replace `CLEAN_ARCHITECTURE_SUMMARY.md` → `concepts/clean-architecture.md`
  - Replace `TESTING_PATTERNS.md` → `guides/testing-patterns.md`
  - Replace `RBAC.md` → `guides/rbac.md`
  - Replace `BUS_TEST_COVERAGE.md` → `reference/bus-test-coverage.md`

- [ ] Fix `internal/contexts/shared/README.md` links:
  ```diff
  - [Clean Architecture](../../../docs/CLEAN_ARCHITECTURE_SUMMARY.md)
  + [Clean Architecture](../../../docs/concepts/clean-architecture.md)
  ```

- [ ] Fix `internal/contexts/customer-mgmt/analytics/README.md` links:
  ```diff
  - [Customer Management](../../../docs/concepts/customer-management.md)
  + [Customer Management Guide](../../../../docs/concepts/customer-management.md)
  ```

### Phase 3: Medium Priority

- [ ] Create missing referenced files:
  - `docs/reference/api-reference.md`
  - `docs/concepts/cqrs.md`
  - `website/reference/configuration.md`

- [ ] Review work-in-progress files:
  - Archive or fix `GAPS_AND_TODOS_OLD.md`
  - Fix or remove `DB_AGNOSTIC_REFACTORING.md`

### Phase 4: Low Priority

- [ ] Improve check-links.py:
  - Exclude VitePress routes from validation
  - Filter out `[T any]` generic patterns
  - Add whitelist for known external routes

- [ ] Document VitePress routing:
  - Create mapping: `/guide/page` → `website/guide/page.md`
  - Validate VitePress routes separately

---

## Commands for Cleanup

### Find and Replace Pattern

```bash
# Update all CLEAN_ARCHITECTURE_SUMMARY.md references
find docs -name "*.md" -type f -exec sed -i '' 's|CLEAN_ARCHITECTURE_SUMMARY\.md|concepts/clean-architecture.md|g' {} +

# Update all TESTING_PATTERNS.md references
find docs -name "*.md" -type f -exec sed -i '' 's|TESTING_PATTERNS\.md|guides/testing-patterns.md|g' {} +

# Update all RBAC.md references
find docs -name "*.md" -type f -exec sed -i '' 's|RBAC\.md|guides/rbac.md|g' {} +
```

### Run Link Check

```bash
# Python version (more accurate)
python3 scripts/check-links.py

# Bash version (faster but less accurate)
./scripts/check-links.sh
```

---

## Statistics

| Category | Count | Status |
|----------|-------|--------|
| **Total Files Checked** | 120+ | ✅ |
| **Broken Links Found** | 200+ | 🔴 |
| **Files Removed** | 1 | ✅ |
| **Scripts Created** | 2 | ✅ |
| **Immediate Fixes** | 3 | ✅ |
| **Remaining Issues** | 195+ | 🔧 |

---

## Next Steps

1. ✅ **Done**: Remove INDEX_OLD.md, sync README, create link checker
2. **Next**: Run Phase 2 (update docs/ structure references)
3. **Then**: Create missing files (Phase 3)
4. **Finally**: Improve validation script (Phase 4)

---

**Status**: Initial audit complete, scripts ready, cleanup in progress  
**Last Updated**: January 3, 2026, 19:30 UTC  
**Maintainer**: Promenade Team
