# Test Audit Report - Phase 1, Step 1.1

**Date**: December 29, 2025  
**Status**: Analysis Complete  
**Next**: Decide what to keep/delete

---

## Summary

| Category              | Count | Verdict       | Action             |
|-----------------------|-------|---------------|--------------------|
| **Smoke tests**       | 47    | 🔴 DELETE ALL | Mock-based, no value |
| **Integration tests** | 78    | 🟡 OPTIMIZE   | Keep ~30 critical  |
| **Unit tests**        | 239   | 🟡 CLEAN UP   | Keep ~50 critical  |
| **Package tests**     | 186   | 🟢 KEEP ALL   | Core functionality |
| **TOTAL**             | 550   | →             | Target: ~266 tests |

---

## 1. Smoke Tests (47 tests) - 🔴 DELETE ALL

**Location**: `test/smoke/contexts/`

### Files to DELETE:

```
test/smoke/contexts/customer-mgmt/customer/handler_test.go       (18 tests)
test/smoke/contexts/identity/contact/handler_test.go             (7 tests)
test/smoke/contexts/identity/permission/handler_test.go          (2 tests)
test/smoke/contexts/identity/profile/handler_test.go             (8 tests)
test/smoke/contexts/identity/role/handler_test.go                (7 tests)
test/smoke/contexts/identity/user/handler_test.go                (1 test)
test/smoke/contexts/shared/country/handler_test.go               (1 test)
test/smoke/contexts/shared/currency/handler_test.go              (1 test)
test/smoke/contexts/shared/language/handler_test.go              (1 test)
test/smoke/contexts/shared/timezone/handler_test.go              (1 test)
```

**Why delete?**
- Mock-based handlers don't test real integration
- Only test HTTP status codes (can verify manually)
- Don't catch SQL errors, validation bugs, etc.
- Maintenance overhead without value

**Command:**
```bash
rm -rf test/smoke/
```

---

## 2. Integration Tests (78 tests) - 🟡 OPTIMIZE

**Location**: `test/integration/contexts/`

### Current state:
```
customer-mgmt/customer/repository_test.go     (14 tests)
identity/contact/repository_test.go           (9 tests)
identity/permission/repository_test.go        (8 tests)
identity/profile/repository_test.go           (6 tests)
identity/role/repository_test.go              (8 tests)
identity/user/repository_test.go              (9 tests)
shared/country/repository_test.go             (6 tests)
shared/currency/repository_test.go            (6 tests)
shared/language/repository_test.go            (6 tests)
shared/timezone/repository_test.go            (6 tests)
```

### Optimization plan:

**Goal**: 78 → ~30 tests (consolidate + table-driven)

**Keep these operations (per repository):**
- ✅ Create (basic + validation)
- ✅ GetByID (found + not found)
- ✅ Update (basic)
- ✅ Delete (soft delete)
- ✅ List (pagination)

**Remove:**
- ❌ Multiple GetByXxx variations (keep 1-2)
- ❌ Duplicate validation tests (keep in entity tests)
- ❌ Edge cases already covered in unit tests

**Example consolidation:**

BEFORE (9 tests):
```
TestContactRepository_Create
TestContactRepository_GetByID
TestContactRepository_GetByUserID
TestContactRepository_GetByEmail
TestContactRepository_Update
TestContactRepository_Delete
TestContactRepository_List
TestContactRepository_SetPrimary
TestContactRepository_VerifyContact
```

AFTER (3-4 tests):
```
TestContactRepository_CRUD        // Create, Read, Update, Delete
TestContactRepository_Queries     // GetByID, GetByUserID, List
TestContactRepository_Operations  // SetPrimary, Verify
```

---

## 3. Unit Tests (239 tests) - 🟡 CLEAN UP

**Location**: `internal/contexts/*/`

### Breakdown by context:

#### Customer Management (43 tests)
```
customer/entity_test.go           (13 tests)
customer/usecase_test.go          (8 tests)
customer/dto_test.go              (7 tests)
customer/adapter/.../handler_test.go (15 tests)
```

**Action**: Keep entity + usecase (21 tests), consider removing DTO tests (trivial)

#### Identity - Contact (26 tests)
```
contact/entity_test.go            (13 tests)
contact/usecase_test.go           (7 tests)
contact/dto_test.go               (2 tests)
contact/adapter/.../handler_test.go (4 tests)
```

**Action**: Keep entity + usecase (20 tests)

#### Identity - Permission (17 tests)
```
permission/entity_test.go         (6 tests)
permission/usecase_test.go        (7 tests)
permission/dto_test.go            (2 tests)
permission/adapter/.../handler_test.go (2 tests)
```

**Action**: Keep entity + usecase (13 tests)

#### Identity - Profile (40 tests)
```
profile/entity_test.go            (11 tests)
profile/usecase_test.go           (11 tests)
profile/dto_test.go               (8 tests)
profile/adapter/.../handler_test.go (10 tests)
```

**Action**: Keep entity + usecase (22 tests), DTO tests trivial

#### Identity - Role (17 tests)
```
role/entity_test.go               (3 tests)
role/usecase_test.go              (7 tests)
role/dto_test.go                  (2 tests)
role/adapter/.../handler_test.go  (5 tests)
```

**Action**: Keep entity + usecase (10 tests)

#### Identity - User (81 tests!)
```
user/entity_test.go               (39 tests)
user/usecase_test.go              (15 tests)
user/dto_test.go                  (9 tests)
user/adapter/.../handler_test.go  (18 tests)
```

**Action**: Keep critical entity tests (15), usecase (10), total ~25

#### Shared Contexts (15 tests)
```
country/entity_test.go + usecase_test.go
currency/entity_test.go + usecase_test.go
language/entity_test.go + usecase_test.go
timezone/entity_test.go + usecase_test.go
```

**Action**: Keep all (reference data is simple)

### Clean-up rules:

**DELETE:**
- ❌ DTO tests (trivial JSON marshaling)
- ❌ Handler unit tests (covered by integration)
- ❌ Getter/Setter tests (no logic)
- ❌ Constructor tests without validation
- ❌ Tests with single assertion

**KEEP:**
- ✅ Entity validation logic
- ✅ Business rules (use case logic)
- ✅ Value object validation
- ✅ Error handling

**Target**: 239 → ~100 tests

---

## 4. Package Tests (186 tests) - 🟢 KEEP ALL

**Location**: `pkg/*/`

### Breakdown:
```
pkg/aggregate/aggregate_test.go              (5 tests)
pkg/bus/event_test.go                        (8 tests)
pkg/bus/factory_test.go                      (13 tests)
pkg/bus/topics_test.go                       (2 tests)
pkg/bus/memory/*_test.go                     (21 tests)
pkg/bus/redis/*_test.go                      (16 tests)
pkg/jsonb/jsonb_test.go                      (8 tests)
pkg/jwt/jwt_test.go                          (18 tests)
pkg/jwt/middleware_test.go                   (7 tests)
pkg/logger/logger_test.go                    (15 tests)
pkg/migration/manager_test.go                (8 tests)
pkg/response/response_test.go                (13 tests)
pkg/saga/saga_test.go + orchestrator_test.go (28 tests)
pkg/uuidv7/uuidv7_test.go                    (10 tests)
pkg/valueobject/*_test.go                    (45+ tests)
```

**Action**: KEEP ALL - these are core utilities, well-tested

---

## Execution Plan

### Phase 1.2: Delete smoke tests (Day 1)
```bash
rm -rf test/smoke/
git rm -r test/smoke/
# Update Makefile.test.mk (remove test-smoke target)
# Update .github/workflows/ci.yml (remove smoke step)
```

### Phase 1.3: Optimize integration tests (Day 2-3)
1. Consolidate each repository to 3-4 tests
2. Use table-driven approach
3. Target: 78 → 30 tests

### Phase 1.4: Clean unit tests (Day 4-5)
1. Delete DTO tests
2. Delete handler unit tests
3. Keep only business logic
4. Target: 239 → 100 tests

### Phase 1.5: Update CI (Day 6)
1. Remove testing.Short() checks
2. Separate unit/integration jobs
3. Clean test infrastructure

---

## Expected Results

### Before:
```
Total: 550 tests
- Smoke: 47 (DELETE)
- Integration: 78 (→ 30)
- Unit: 239 (→ 100)
- Package: 186 (KEEP)
CI time: ~40s
```

### After:
```
Total: 316 tests
- Smoke: 0 (deleted)
- Integration: 30 (optimized)
- Unit: 100 (cleaned)
- Package: 186 (kept)
CI time: ~15s (unit only)
```

---

## Next Steps

1. ✅ Review this report
2. ⏳ Confirm approach
3. ⏳ Start Phase 1.2 (delete smoke tests)

---

**Status**: Waiting for approval  
**Date**: December 29, 2025
