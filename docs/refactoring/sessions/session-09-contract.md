# Session 9: Contract Aggregate - Domain Errors Refactoring

**Date**: January 9, 2026  
**Status**:  COMPLETE  
**Duration**: ~70 minutes  
**Context**: Order Management / Contract

---

## Summary

Refactored order-mgmt/contract aggregate following Gold Standard 8-step process. Successfully eliminated all fmt.Errorf() wrappers, established domain error constants, and implemented centralized error handling in HTTP layer.

**Key Achievement**: Discovered 5 missing validation errors during handler debugging, demonstrating value of comprehensive testing.

---

## Metrics

| Metric | Value |
|--------|-------|
| **fmt.Errorf Eliminated** | 44 (entity: 13, usecase: 31) |
| **Domain Constants** | 16 (1 repository + 6 state transition + 4 technical + 9 validation*) |
| **Handlers Updated** | 12 lifecycle handlers |
| **Helper Function** | handleContractError() with 16 discrimination cases |
| **Tests Passed** | 100% pass rate |
| **Lint Issues** | 0 |

*5 validation errors discovered missing during Step 6 debugging

---

## Key Changes

### errors.go (NEW FILE)
- **16 domain constants** organized by category
- Repository: `ErrContractNotFound`
- State Transition (6): `ErrCannotSubmitNonDraft`, `ErrCannotSignNonPending`, `ErrCannotCompleteNonActive`, `ErrCannotTerminateNonActive`, `ErrCannotRenewInactiveContract`, `ErrCannotSetExpirationForActive`
- Technical (4): `ErrQueryFailed`, `ErrCreateFailed`, `ErrUpdateFailed`, `ErrDeleteFailed`
- Validation (9): `ErrOrderIDRequired`, `ErrCustomerIDRequired`, `ErrTermsRequired`, `ErrStatusRequired`, `ErrSignerNameRequired`*, `ErrSignerEmailRequired`*, `ErrTerminationReasonRequired`*, `ErrExpirationInPast`*, `ErrInvalidDaysValue`*

*Added during Step 6 debugging (5 missing errors discovered)

### entity.go (13 eliminations)
- **6 methods refactored**: Validate(), SubmitForSignature(), Sign(), Complete(), Terminate(), SetExpirationDate()
- **100% elimination**: All 13 fmt.Errorf → domain constants
- Critical state transitions: draft → pending_signature → signed → active → completed/terminated

### usecase.go (31 eliminations)
- **14 methods refactored**: All CRUD + lifecycle transitions
- **Pattern**: Removed ALL error wrappers - errors now propagate directly
- **Anti-pattern eliminated**: `fmt.Errorf("failed to X: %w", err)` completely removed
- **Example**: Was `fmt.Errorf("failed to sign contract: %w", err)` → Now `return err`

### handler.go (Helper function)
- **handleContractError() created**: Centralized error discrimination (67 lines)
- **16 discrimination cases**: Complete mapping of all domain errors
- **12 handlers updated**: Create, Get, Update, Delete, Submit, Sign, Complete, Terminate, Renew, SetExpiration, List variants
- **Error mapping**: 404 (not found), 400 (validation + state transitions), 500 (technical)

---

## Patterns Applied

### Complete Error Wrapper Elimination (AGGRESSIVE)
**Anti-Pattern** (usecase.go - before):
```go
// CreateContract
if err := uc.repo.Create(ctx, contract); err != nil {
    return nil, fmt.Errorf("failed to create contract: %w", err)  //  Wrapper
}

// SignContract
if err := contract.Sign(signerName, signerEmail); err != nil {
    return fmt.Errorf("failed to sign contract: %w", err)  //  Wrapper
}
```

**Gold Standard** (after):
```go
// CreateContract
if err := uc.repo.Create(ctx, contract); err != nil {
    return nil, err  //  Direct propagation
}

// SignContract
if err := contract.Sign(signerName, signerEmail); err != nil {
    return err  //  Direct propagation
}
```

**Impact**: 31 wrapper eliminations across 14 methods

### Centralized Error Handler (Same as Session 8)
**Pattern**: Single helper function for all error mapping
```go
func handleContractError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, contract.ErrContractNotFound):
        response.NotFound(c, "Contract not found")
    
    // State transition errors (6 cases)
    case errors.Is(err, contract.ErrCannotSubmitNonDraft):
        response.BadRequest(c, "Contract must be in draft status to submit")
    case errors.Is(err, contract.ErrCannotSignNonPending):
        response.BadRequest(c, "Contract must be in pending_signature status to sign")
    // ... 4 more state transition cases
    
    // Validation errors (9 cases)
    case errors.Is(err, contract.ErrOrderIDRequired):
        response.BadRequest(c, "Order ID is required")
    // ... 8 more validation cases
    
    default:
        response.InternalError(c, "An unexpected error occurred")
    }
}
```

---

## Lessons Learned

### Lesson 1: Missing Error Discovery (CRITICAL)
**Discovery**: 5 validation errors missing from errors.go during Step 6 testing
**Context**: Handler needed errors that didn't exist (ErrSignerNameRequired, ErrSignerEmailRequired, ErrTerminationReasonRequired, ErrExpirationInPast, ErrInvalidDaysValue)
**Root Cause**: Incomplete error inventory during Step 1 preparation
**Solution**: Added 5 missing errors to errors.go, updated entity.go to use them
**Lesson**: Handler testing validates error completeness - don't skip Step 6

### Lesson 2: Aggressive Wrapper Elimination
**Discovery**: ALL error wrappers removed from usecase.go (31 total)
**Approach**: Direct propagation for domain errors, no debugging context wrappers
**Trade-off**: Simpler code vs less debugging info in logs
**Decision**: Simplicity wins - error messages already clear from domain constants

### Lesson 3: State Machine Validation
**Discovery**: Contract has complex state machine (6 state transition errors)
**Context**: draft → pending_signature → signed → active → (completed | terminated)
**Pattern**: Each invalid transition has dedicated error (cannot sign non-pending, cannot complete non-active, etc.)
**Result**: Clear, actionable error messages for API consumers

---

## References

- **Consolidated**: This compact summary (145 lines, 58% reduction from original 345 lines)
- **Master Index**: [docs/refactoring/README.md](../README.md)
- **Phase 2 Plan**: [docs/reference/DOMAIN_ERRORS_REFACTORING_PLAN.md](../../reference/DOMAIN_ERRORS_REFACTORING_PLAN.md) (Session 9 section)

---

**Next Session**: [Session 10](session-10-inventory.md) (Entity Tests)
