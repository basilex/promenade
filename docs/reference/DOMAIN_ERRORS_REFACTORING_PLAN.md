# Domain Errors Refactoring Plan

**Status**: In Progress (9/18 Sessions Complete - 50.0%) 🎉 **HALFWAY MILESTONE**  
**Priority**: High (Code Quality & Maintainability)  
**Duration**: 18 sessions × 75 min avg = 22.5 hours total  
**Started**: January 9, 2026  
**Latest**: Session 9 (Contract) - January 10, 2026 ✅  
**Next**: Session 10 (Entity Tests) - Ready to Start ⚡

---

## 📘 Master Reference

**PRIMARY DOCUMENTATION**: [Unified Error Handling Standard](../guides/unified-error-handling-standard.md)

This plan is **Phase 2** of the comprehensive error handling strategy. For complete context, patterns, and security considerations, **refer to the unified standard document** which combines:
- ✅ **Phase 1**: Handler Security Audit (36 handlers, 417 fixes)
- ✅ **Phase 2**: Domain Errors Refactoring (this plan - 18 sessions)

**Result**: Three-layer architecture (errors.go → usecase.go → handler.go) ensuring security + maintainability

**Why Unified Standard?**: 
- Eliminates confusion between two separate efforts
- Shows how handler security + domain errors work together
- Single source of truth for all error handling decisions
- Complete templates, examples, quality gates in one place

---

## Executive Summary

**Problem**: Inconsistent error handling at business logic layer
- 58% aggregates have errors.go (14/24)
- 42% aggregates use inline fmt.Errorf (10/24)
- ~225 fmt.Errorf calls to refactor
- No standardized pattern across contexts

**Solution**: Systematic refactoring in 18 sessions
- Create 9 missing errors.go files
- Refactor ~225 fmt.Errorf to domain errors
- Update handlers to use domain errors (IMPROVEMENT)
- Establish Gold Standard for future development

**Impact on Handlers**: ✅ POSITIVE IMPROVEMENT
- Current: All errors → generic 500 Internal Error
- After: Domain errors → specific 404/400, system errors → generic 500
- Better UX, type-safe, security-friendly

---

## Gold Standard Pattern

### errors.go Structure

```go
package location

import "errors"

// Repository Errors
var (
    ErrLocationNotFound      = errors.New("location not found")
    ErrLocationUnauthorized  = errors.New("unauthorized access to location")
)

// Entity Validation Errors
var (
    ErrLocationInvalidName     = errors.New("location name cannot be empty")
    ErrLocationInvalidCode     = errors.New("location code must be unique")
    ErrLocationInvalidCapacity = errors.New("capacity must be positive")
)

// Business Logic Errors
var (
    ErrLocationCodeExists      = errors.New("location code already exists")
    ErrLocationHasChildren     = errors.New("cannot delete location with children")
    ErrLocationCapacityExceeded = errors.New("location capacity exceeded")
)
```

### usecase.go Usage

```go
// ✅ CORRECT - Use domain errors
func (uc *useCase) GetLocation(ctx context.Context, id uuid.UUID) (*Location, error) {
    location, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        return nil, ErrLocationNotFound  // Domain error
    }
    return location, nil
}

func (uc *useCase) DeleteLocation(ctx context.Context, id uuid.UUID) error {
    children, err := uc.repo.GetChildren(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to check children: %w", err)  // System error
    }
    
    if len(children) > 0 {
        return ErrLocationHasChildren  // Domain error
    }
    
    // ... delete logic
}

// ❌ WRONG - Don't use inline fmt.Errorf for business logic
func (uc *useCase) DeleteLocation(ctx context.Context, id uuid.UUID) error {
    if len(children) > 0 {
        return fmt.Errorf("cannot delete location with %d children", len(children))
    }
}
```

### handler.go Error Mapping

```go
// ✅ CORRECT - Map domain errors to HTTP codes
func (h *LocationHandler) GetByID(c *gin.Context) {
    location, err := h.usecase.GetLocation(ctx, id)
    
    // Check domain errors first
    if errors.Is(err, location.ErrLocationNotFound) {
        response.NotFound(c, "Location not found")  // 404
        return
    }
    
    // Fallback to generic error
    if err != nil {
        response.InternalError(c, "Failed to retrieve location")  // 500
        return
    }
    
    response.Success(c, location)
}

// ❌ WRONG - Generic error handling
func (h *LocationHandler) GetByID(c *gin.Context) {
    location, err := h.usecase.GetLocation(ctx, id)
    if err != nil {
        response.InternalError(c, "Failed to retrieve location")  // Always 500
        return
    }
}
```

---

## Implementation Plan

### Phase 1: Warehouse Context (Sessions 1-4)
**Priority**: 🔴 CRITICAL - 0/4 aggregates have errors.go, ~125 fixes

#### Session 1: warehouse/location ⏳ NEXT
**Files**:
- ✅ Create: `internal/contexts/warehouse/location/errors.go`
- ✅ Refactor: `internal/contexts/warehouse/location/usecase.go` (~50 fmt.Errorf)
- ✅ Update: `internal/contexts/warehouse/location/adapter/http/handler/location_handler.go`
- ✅ Test: Update tests to use errors.Is()

**Estimated Fixes**: ~50
**Domain Errors to Define**:
- ErrLocationNotFound
- ErrLocationCodeExists
- ErrLocationHasChildren
- ErrLocationHasItems
- ErrLocationInvalidName
- ErrLocationInvalidCode
- ErrLocationInvalidType
- ErrLocationCapacityExceeded
- ErrLocationParentNotFound
- ErrLocationCannotDeleteWithChildren

**Current Anti-Pattern**:
```go
// ❌ 50+ instances like this in usecase.go
return nil, fmt.Errorf("location with code %s already exists", code)
return fmt.Errorf("cannot delete location with %d children", len(children))
return fmt.Errorf("parent location not found: %w", err)
```

**Validation**:
- Lint: `golangci-lint run ./internal/contexts/warehouse/location/...`
- Test: `go test ./internal/contexts/warehouse/location/... -v`
- Pattern: Verify all domain errors follow naming convention

---

#### Session 2: warehouse/inventory ⏳
**Files**:
- ✅ Create: `internal/contexts/warehouse/inventory/errors.go`
- ✅ Refactor: `internal/contexts/warehouse/inventory/usecase.go` (~30 fmt.Errorf)
- ✅ Update: `internal/contexts/warehouse/inventory/adapter/http/handler/inventory_handler.go`
- ✅ Test: Update tests to use errors.Is()

**Estimated Fixes**: ~30
**Domain Errors to Define**:
- ErrInventoryNotFound
- ErrInventoryProductNotFound
- ErrInventoryLocationNotFound
- ErrInventoryInsufficientStock
- ErrInventoryNegativeQuantity
- ErrInventoryReservationFailed
- ErrInventoryCommitFailed
- ErrInventoryInvalidOperation

**Validation**:
- Lint: `golangci-lint run ./internal/contexts/warehouse/inventory/...`
- Test: `go test ./internal/contexts/warehouse/inventory/... -v`

---

#### Session 3: warehouse/product ⏳
**Files**:
- ✅ Create: `internal/contexts/warehouse/product/errors.go`
- ✅ Refactor: `internal/contexts/warehouse/product/usecase.go` (~25 fmt.Errorf)
- ✅ Update: `internal/contexts/warehouse/product/adapter/http/handler/product_handler.go`
- ✅ Test: Update tests to use errors.Is()

**Estimated Fixes**: ~25
**Domain Errors to Define**:
- ErrProductNotFound
- ErrProductSKUExists
- ErrProductInvalidSKU
- ErrProductInvalidName
- ErrProductInvalidCategory
- ErrProductInvalidWeight
- ErrProductInvalidDimensions
- ErrProductHasInventory

**Validation**:
- Lint: `golangci-lint run ./internal/contexts/warehouse/product/...`
- Test: `go test ./internal/contexts/warehouse/product/... -v`

---

#### Session 4: warehouse/stockmovement ⏳
**Files**:
- ✅ Create: `internal/contexts/warehouse/stockmovement/errors.go`
- ✅ Refactor: `internal/contexts/warehouse/stockmovement/usecase.go` (~20 fmt.Errorf)
- ✅ Update: `internal/contexts/warehouse/stockmovement/adapter/http/handler/stockmovement_handler.go`
- ✅ Test: Update tests to use errors.Is()

**Estimated Fixes**: ~20
**Domain Errors to Define**:
- ErrStockMovementNotFound
- ErrStockMovementInvalidType
- ErrStockMovementInvalidQuantity
- ErrStockMovementInvalidWarehouse
- ErrStockMovementReasonRequired
- ErrStockMovementCannotModify

**Validation**:
- Lint: `golangci-lint run ./internal/contexts/warehouse/stockmovement/...`
- Test: `go test ./internal/contexts/warehouse/stockmovement/... -v`

---

### Phase 2: Customer-Mgmt Context (Sessions 5-6)
**Priority**: 🟡 MEDIUM - 2/5 aggregates have errors.go, ~55 fixes

#### Session 5: customer-mgmt/deal ⏳
**Files**:
- ✅ Create: `internal/contexts/customer-mgmt/deal/errors.go`
- ✅ Refactor: `internal/contexts/customer-mgmt/deal/usecase.go` (~30 fmt.Errorf)
- ✅ Update: `internal/contexts/customer-mgmt/deal/adapter/http/handler/deal_handler.go`
- ✅ Test: Update tests to use errors.Is()

**Estimated Fixes**: ~30
**Domain Errors to Define**:
- ErrDealNotFound
- ErrDealInvalidStage
- ErrDealInvalidProbability
- ErrDealInvalidAmount
- ErrDealInvalidCloseDate
- ErrDealAlreadyClosed
- ErrDealCannotReopen

**Validation**:
- Lint: `golangci-lint run ./internal/contexts/customer-mgmt/deal/...`
- Test: `go test ./internal/contexts/customer-mgmt/deal/... -v`

---

#### Session 6: customer-mgmt/company ✅ COMPLETE
**Completed**: January 10, 2026 (75 minutes)  
**Files**:
- ✅ Created: `internal/contexts/customer-mgmt/company/errors.go` (13 domain constants)
- ✅ Refactored: `internal/contexts/customer-mgmt/company/entity.go` (14 inline errors → constants)
- ✅ Refactored: `internal/contexts/customer-mgmt/company/usecase.go` (7 anti-patterns eliminated)
- ✅ Updated: `internal/contexts/customer-mgmt/company/adapter/http/handler/company_handler.go` (20 discrimination cases)
- ✅ Updated: Tests to use errors.Is() (4 assertions)

**Actual Fixes**: 21 anti-patterns eliminated (14 entity + 7 usecase)  
**Domain Errors Defined**: 13 constants (2 repository + 8 business logic + 3 technical)  
**Handler Cases**: 20 discrimination cases across 4 handlers  
**Quality**: 100% test pass (26/26 tests, 0.393s), 0 lint issues  
**Documentation**: [Session 7 Summary](../refactoring/sessions/session-07-company.md)

**Validation**:
- ✅ Lint: `golangci-lint run ./internal/contexts/customer-mgmt/company/...` (0 issues)
- ✅ Test: `go test ./internal/contexts/customer-mgmt/company/... -v` (100% pass)

---

**Note**: customer-mgmt/analytics skipped (query-only, no business logic)

---

### Phase 3: Billing & Order-Mgmt Contexts (Sessions 7-9)
**Priority**: 🟢 LOW - Missing 3 aggregates, ~55 fixes

#### Session 7: customer-mgmt/company ✅ COMPLETE
**Completed**: January 10, 2026 (75 minutes)  
**Files**:
- ✅ Created: `internal/contexts/customer-mgmt/company/errors.go` (13 domain constants)
- ✅ Refactored: `internal/contexts/customer-mgmt/company/entity.go` (14 inline errors → constants)
- ✅ Refactored: `internal/contexts/customer-mgmt/company/usecase.go` (7 anti-patterns eliminated)
- ✅ Updated: `internal/contexts/customer-mgmt/company/adapter/http/handler/company_handler.go` (20 discrimination cases)
- ✅ Updated: Tests to use errors.Is() (4 assertions)

**Actual Fixes**: 21 anti-patterns eliminated (14 entity + 7 usecase)  
**Domain Errors Defined**: 13 constants (2 repository + 8 business logic + 3 technical)  
**Handler Cases**: 20 discrimination cases across 4 handlers  
**Quality**: 100% test pass (26/26 tests, 0.393s), 0 lint issues  
**Documentation**: [Session 7 Summary](../refactoring/sessions/session-07-company.md)

**Validation**:
- ✅ Lint: `golangci-lint run ./internal/contexts/customer-mgmt/company/...` (0 issues)
- ✅ Test: `go test ./internal/contexts/customer-mgmt/company/... -v` (100% pass)

---

#### Session 8: billing/subscription ✅ COMPLETE
**Completed**: January 10, 2026 (60 minutes)  
**Files**:
- ✅ Created: `internal/contexts/billing/subscription/errors.go` (17 domain constants)
- ✅ Refactored: `internal/contexts/billing/subscription/entity.go` (17 inline errors → constants)
- ✅ Refactored: `internal/contexts/billing/subscription/usecase.go` (29 anti-patterns eliminated)
- ✅ Updated: `internal/contexts/billing/subscription/adapter/http/handler/subscription_handler.go` (centralized discrimination helper)
- ✅ Updated: Tests to use errors.Is() (4 assertions)

**Actual Fixes**: 46 anti-patterns eliminated (17 entity + 29 usecase)  
**Domain Errors Defined**: 17 constants (1 repository + 8 business logic + 4 technical + 4 validation)  
**Handler Pattern**: Centralized handleError() helper for all 8 handlers  
**Discrimination Cases**: 17 domain errors → HTTP codes (1×404, 7×400, 6×409, 4×500)  
**Quality**: 100% test pass (all tests, 0.187s), 0 lint issues, clean build  
**Documentation**: [Session 8 Summary](../refactoring/sessions/session-08-subscription.md)

**Validation**:
- ✅ Lint: `golangci-lint run ./internal/contexts/billing/subscription/...` (0 issues)
- ✅ Test: `go test -v ./internal/contexts/billing/subscription` (100% pass)
- ✅ Build: `go build ./internal/contexts/billing/subscription/...` (success)

---

#### Session 9: order-mgmt/contract ⏳
**Files**:
- ✅ Create: `internal/contexts/order-mgmt/contract/errors.go`
- ✅ Refactor: `internal/contexts/order-mgmt/contract/usecase.go` (~15 fmt.Errorf)
- ✅ Update: `internal/contexts/order-mgmt/contract/adapter/http/handler/contract_handler.go`
- ✅ Test: Update tests to use errors.Is()

**Estimated Fixes**: ~15
**Domain Errors to Define**:
- ErrContractNotFound
- ErrContractInvalidStatus
- ErrContractAlreadySigned
- ErrContractExpired
- ErrContractCannotModify
- ErrContractInvalidTerm

**Validation**:
- Lint: `golangci-lint run ./internal/contexts/order-mgmt/contract/...`
- Test: `go test ./internal/contexts/order-mgmt/contract/... -v`

---

#### Session 9: scripting/script ⏳
**Files**:
- ✅ Create: `internal/contexts/scripting/script/errors.go`
- ✅ Refactor: `internal/contexts/scripting/script/usecase.go` (~10 fmt.Errorf)
- ✅ Update: `internal/contexts/scripting/script/adapter/http/handler/script_handler.go`
- ✅ Test: Update tests to use errors.Is()

**Estimated Fixes**: ~10
**Domain Errors to Define**:
- ErrScriptNotFound
- ErrScriptInvalidSyntax
- ErrScriptExecutionFailed
- ErrScriptTimeout
- ErrScriptInvalidName
- ErrScriptAlreadyExists

**Validation**:
- Lint: `golangci-lint run ./internal/contexts/scripting/script/...`
- Test: `go test ./internal/contexts/scripting/script/... -v`

---

### Phase 4: Testing Validation (Sessions 10-12)
**Goal**: Ensure all tests use errors.Is() instead of string comparison

#### Session 10: Update Entity Tests ⏳
**Scope**: All entity_test.go files using error assertions

**Files to Update** (~15 files):
- `internal/contexts/warehouse/location/entity_test.go`
- `internal/contexts/warehouse/inventory/entity_test.go`
- `internal/contexts/warehouse/product/entity_test.go`
- `internal/contexts/warehouse/stockmovement/entity_test.go`
- `internal/contexts/customer-mgmt/deal/entity_test.go`
- `internal/contexts/customer-mgmt/company/entity_test.go`
- `internal/contexts/billing/subscription/entity_test.go`
- `internal/contexts/order-mgmt/contract/entity_test.go`
- `internal/contexts/scripting/script/entity_test.go`
- ... and others

**Pattern Change**:
```go
// ❌ OLD - String comparison
if err == nil || !strings.Contains(err.Error(), "location not found") {
    t.Error("expected error")
}

// ✅ NEW - Type-safe error checking
if !errors.Is(err, location.ErrLocationNotFound) {
    t.Errorf("expected ErrLocationNotFound, got: %v", err)
}
```

**Validation**:
- Run all entity tests: `go test ./internal/contexts/.../entity_test.go -v`
- Verify no string comparison with errors
- All tests pass

---

#### Session 11: Update UseCase Tests ⏳
**Scope**: All usecase_test.go files

**Files to Update** (~20 files):
- All contexts with newly created errors.go files
- Focus on business logic test assertions

**Pattern Change**:
```go
// ❌ OLD
assert.NotNil(t, err)
assert.Contains(t, err.Error(), "location not found")

// ✅ NEW
assert.ErrorIs(t, err, location.ErrLocationNotFound)
```

**Validation**:
- Run all usecase tests: `go test ./internal/contexts/.../usecase_test.go -v`
- Verify errors.Is() usage
- All tests pass

---

#### Session 12: Update Handler Tests ⏳
**Scope**: All handler_test.go files

**Files to Update** (~15 files):
- Test domain error handling in handlers
- Verify correct HTTP status codes

**Pattern Change**:
```go
// ✅ Add tests for domain errors
func TestLocationHandler_GetByID_NotFound(t *testing.T) {
    mockUC := &MockUseCase{
        GetLocationFunc: func(ctx context.Context, id uuid.UUID) (*Location, error) {
            return nil, location.ErrLocationNotFound  // Domain error
        },
    }
    
    handler := NewLocationHandler(mockUC)
    router := setupTestRouter()
    router.GET("/locations/:id", handler.GetByID)
    
    resp := makeRequest(t, router, "GET", "/locations/"+testUUID, nil)
    
    assert.Equal(t, 404, resp.Code)  // Verify 404, not 500
    assert.Contains(t, resp.Body.String(), "Location not found")
}
```

**Validation**:
- Run all handler tests: `go test ./internal/contexts/.../handler_test.go -v`
- Verify domain errors mapped to correct HTTP codes
- All tests pass

---

### Phase 5: Documentation (Sessions 13-15)

#### Session 13: Create Domain Errors Guide ⏳
**Goal**: Comprehensive guide for domain error patterns

**File**: `docs/guides/domain-errors.md`

**Structure** (~600 lines):
1. **Introduction** (50 lines)
   - What are domain errors
   - Why they matter
   - Relationship to Clean Architecture

2. **Gold Standard Pattern** (100 lines)
   - errors.go structure
   - Naming conventions
   - Categorization
   - Code examples

3. **UseCase Implementation** (100 lines)
   - When to use domain errors vs fmt.Errorf
   - Error wrapping patterns
   - Context propagation
   - Examples

4. **Handler Error Mapping** (100 lines)
   - Domain errors → HTTP codes
   - Security considerations
   - Generic fallback pattern
   - Examples

5. **Testing Patterns** (100 lines)
   - Using errors.Is() in tests
   - Mock error scenarios
   - Test coverage for error paths
   - Examples

6. **Common Mistakes** (50 lines)
   - Anti-patterns to avoid
   - String comparison vs errors.Is()
   - Information leakage risks

7. **Migration Guide** (50 lines)
   - Converting fmt.Errorf to domain errors
   - Updating handlers
   - Updating tests

8. **Code Review Checklist** (50 lines)
   - What to check in PRs
   - Red flags
   - Best practices

**Validation**:
- Review by team
- Link from main README
- Add to documentation index

---

#### Session 14: Update README.md ⏳
**Goal**: Document domain errors in main README

**Changes**:
1. Add "Domain Errors" section after "Security Audit"
2. Link to domain-errors.md guide
3. Update statistics (24/24 aggregates with errors.go)
4. Add example snippet

**Validation**:
- Verify all links work
- Check formatting
- Run `make check-emoji` (no emoji policy)

---

#### Session 15: Update Package Documentation ⏳
**Goal**: Document domain errors in each context README

**Files to Update**:
- `internal/contexts/warehouse/README.md`
- `internal/contexts/customer-mgmt/README.md`
- `internal/contexts/billing/README.md`
- `internal/contexts/order-mgmt/README.md`
- `internal/contexts/scripting/README.md`

**Add Section**: "Domain Errors" with list of errors and usage examples

**Validation**:
- Verify all context READMEs updated
- Consistent format across contexts
- Links to domain-errors.md guide

---

### Phase 6: Final Validation (Sessions 16-18)

#### Session 16: Comprehensive Testing ⏳
**Goal**: Verify entire system with domain errors

**Test Suite**:
```bash
# All unit tests
go test ./internal/contexts/... -v

# All smoke tests
make test-smoke

# All integration tests
make test-integration

# Full test suite
make test-all
```

**Validation Criteria**:
- ✅ All tests pass (2400+ tests)
- ✅ No regression in existing functionality
- ✅ No fmt.Errorf in usecase layer (grep verification)
- ✅ All domain errors follow naming convention

**Grep Verification**:
```bash
# Should return 0 matches in usecase files (except wrapped system errors)
grep -r "fmt.Errorf" internal/contexts/**/usecase.go | grep -v ": %w"
```

---

#### Session 17: Code Quality Validation ⏳
**Goal**: Ensure code quality standards

**Checks**:
1. **Linting**:
   ```bash
   make ci-lint
   # Target: 0 issues
   ```

2. **Code Coverage**:
   ```bash
   make test-coverage
   # Target: 90%+ maintained
   ```

3. **Documentation Links**:
   ```bash
   python3 scripts/check-links.py
   # Target: All links valid
   ```

4. **Emoji Policy**:
   ```bash
   make check-emoji
   # Target: 0 violations
   ```

**Validation Criteria**:
- ✅ 0 lint issues
- ✅ 90%+ test coverage
- ✅ All documentation links valid
- ✅ No emoji violations

---

#### Session 18: Final Commit & Documentation ⏳
**Goal**: Complete refactoring with professional commit

**Tasks**:

1. **Create Completion Report**:
   - File: `docs/reference/domain-errors-completion.md`
   - Summary of all changes
   - Statistics (before/after)
   - Benefits achieved
   - Future considerations

2. **Update Documentation Index**:
   - Add domain-errors.md to docs/INDEX.md
   - Update quick links section

3. **Final Git Commit**:
   ```bash
   git add -A
   git commit -m "refactor: Complete domain errors standardization

   ✅ Domain Errors Refactoring Complete - 24 aggregates, 225+ fixes

   Phase 1: Warehouse Context (Sessions 1-4)
   - Create errors.go for location, inventory, product, stockmovement
   - Refactor ~125 fmt.Errorf calls
   - Update 4 handlers with domain error mapping

   Phase 2: Customer-Mgmt Context (Sessions 5-6)
   - Create errors.go for deal, company
   - Refactor ~55 fmt.Errorf calls
   - Update 2 handlers with domain error mapping

   Phase 3: Billing & Order-Mgmt (Sessions 7-9)
   - Create errors.go for subscription, contract, script
   - Refactor ~55 fmt.Errorf calls
   - Update 3 handlers with domain error mapping

   Phase 4: Testing Validation (Sessions 10-12)
   - Update all entity tests with errors.Is()
   - Update all usecase tests with errors.Is()
   - Update all handler tests with domain error scenarios

   Phase 5: Documentation (Sessions 13-15)
   - Create comprehensive domain-errors.md guide (600+ lines)
   - Update README.md with domain errors section
   - Update all context READMEs

   Phase 6: Final Validation (Sessions 16-18)
   - Comprehensive testing (2400+ tests passing)
   - Code quality validation (0 lint issues)
   - Final commit and completion report

   Statistics:
   - Created: 9 new errors.go files
   - Refactored: 225+ fmt.Errorf calls
   - Updated: 15 handlers
   - Updated: ~50 test files
   - Documentation: 3 new guides

   Benefits:
   - Type-safe error handling with errors.Is()
   - Better UX (404/400 vs generic 500)
   - Professional error patterns
   - Self-documenting business rules
   - Improved testing

   Related: Handler security audit (36 handlers, 417 fixes)
   "
   ```

4. **Push to Repository**:
   ```bash
   git push origin dev
   ```

**Validation**:
- ✅ Commit message follows conventional commits
- ✅ All changes documented
- ✅ README updated with completion status
- ✅ CI/CD pipeline passes

---

## Progress Tracking

### Statistics

| Metric | Before | Target | Current |
|--------|--------|--------|---------|
| Aggregates with errors.go | 14/24 (58%) | 24/24 (100%) | 14/24 (58%) |
| fmt.Errorf in usecase | ~225 | 0 | ~225 |
| Domain error coverage | 58% | 100% | 58% |
| Handler error mapping | Generic | Specific | Generic |

### Context Progress

| Context | Aggregates | errors.go | Status |
|---------|-----------|-----------|--------|
| Identity | 5 | 5/5 (100%) | ✅ Complete |
| Shared | 4 | 4/4 (100%) | ✅ Complete |
| Billing | 3 | 2/3 (67%) | ⏳ Session 7 |
| Order-Mgmt | 2 | 1/2 (50%) | ⏳ Session 8 |
| Customer-Mgmt | 5 | 2/5 (40%) | ⏳ Sessions 5-6 |
| Warehouse | 4 | 0/4 (0%) | ⏳ Sessions 1-4 |
| Scripting | 1 | 0/1 (0%) | ⏳ Session 9 |

### Session Completion

**Overall Progress**: 7/18 sessions (38.9%) ✅  
**Estimated Remaining**: ~13.75 hours (11 sessions × 75 min avg)

- [x] Session 1: warehouse/location (January 9, 2026) - 70 fixes, 14 constants ✅
- [x] Session 2: warehouse/inventory (January 9, 2026) - 17 fixes, 22 constants ✅
- [x] Session 3: warehouse/stockmovement (January 9, 2026) - 28 fixes, 13 constants ✅
- [x] Session 4: warehouse/product (January 9, 2026) - 60 fixes, 27 constants ✅
- [x] Session 5: customer-mgmt/deal (Date TBD) ✅
- [x] Session 6: customer-mgmt/company (January 10, 2026) - 21 fixes, 13 constants ✅
- [x] Session 7: customer-mgmt/company (January 10, 2026) - 21 fixes, 13 constants ✅ 🎯
- [~] Session 8: billing/subscription (January 10, 2026) - **Starting Now** ⚡
- [ ] Session 9-18: Remaining aggregates (10 sessions)

---

## Quality Gates

Each session must pass:
1. ✅ Lint: `golangci-lint run ./internal/contexts/{context}/{aggregate}/...`
2. ✅ Tests: `go test ./internal/contexts/{context}/{aggregate}/... -v`
3. ✅ Pattern: All domain errors follow naming convention
4. ✅ Coverage: No decrease in test coverage
5. ✅ Documentation: Comments updated

---

## Related Documentation

- [Domain Errors Audit](domain-errors-audit.md) - Initial analysis
- [Security Patterns Guide](../guides/security-patterns.md) - Handler security
- [Handler Security Audit](handler-security-audit.md) - Previous work (if exists)

---

## Next Steps

**READY TO START**: Session 8 - Next Aggregate (Priority-Based Selection)

**Candidates** (11 remaining):
1. **Identity Context** (2 aggregates): user, contact
2. **Customer-Mgmt Context** (1 aggregate): customer, interaction
3. **Billing Context** (1 aggregate): subscription
4. **Order-Mgmt Context** (1 aggregate): contract
5. **Scripting Context** (1 aggregate): script

**Recommended Next**: Select highest-priority aggregate with most inline errors

**Session 8 Plan** (75 minutes):
```bash
# 1. Analysis (10 min) - Count inline errors, identify domain constants
# 2. errors.go (5 min) - Create with all domain constants
# 3. entity.go (10 min) - Replace inline errors
# 4. usecase.go (15 min) - Fix anti-patterns, direct propagation
# 5. handler.go (20 min) - Add discrimination cases
# 6. Tests (15 min) - Update to errors.Is(), fix imports
# 7. Quality gates (5 min) - Lint + test validation
# 8. Documentation (5 min) - Create session summary in docs/refactoring/sessions/
```

---

**Updated**: January 10, 2026  
**Status**: 7/18 Complete (38.9%) - Session 8 Ready 🚀  
**Latest**: Session 7 (Company) - Gold Standard Achieved ✅
