# Test Coverage Matrix - Quick Reference

**Last Updated**: January 5, 2026

| Context | Aggregate | Entity Test | UseCase Test | Integration UseCase | Integration Repo | Status |
|---------|-----------|-------------|--------------|---------------------|------------------|---------|
| **Identity** | user |  |  |  |  |  Complete |
| | contact |  |  |  |  |  Complete |
| | profile |  |  |  |  |  Complete |
| | role |  |  |  |  |  Complete |
| | permission |  |  |  |  |  Complete |
| **Customer-mgmt** | customer |  |  |  |  |  Complete |
| | company |  |  |  | - |  Complete |
| | deal |  |  |  | - |  Complete |
| | **interaction** |  |  |  **MISSING** |  |  **Incomplete** |
| | analytics |  | - |  | - |  Complete |
| **Billing** | invoice |  |  | - |  |  Complete |
| | payment |  |  | - |  |  Complete |
| | subscription |  |  | - |  |  Complete |
| **Order-mgmt** | order |  |  | - |  |  Complete |
| **Shared** | country |  |  | - |  |  Complete |
| | currency |  |  | - |  |  Complete |
| | language |  |  | - |  |  Complete |
| | timezone |  |  | - |  |  Complete |

**Legend**:
-  = File exists and in correct location
-  = File missing (needs creation)
- - = Not required (simpler business logic)

## Summary

- **Total Aggregates**: 17
- **Entity Tests**: 17/17 (100%) 
- **UseCase Tests**: 17/17 (100%) 
- **Integration UseCase Tests**: 11/12 (92%) 
- **Integration Repo Tests**: 17/17 (100%) 

## Critical Finding

**Missing File**: `test/integration/contexts/customer-mgmt/interaction/usecase_test.go`

This is the ONLY missing test file across all 17 aggregates. All other tests are present and properly organized.

## File Locations

### Unit Tests (internal/contexts/)
```
internal/contexts/{context}/{aggregate}/
 entity_test.go        ← Tests domain entity logic
 usecase_test.go       ← Tests business logic with mocks
 adapter/http/dto_test.go  ← Tests DTO conversion (optional)
```

### Integration Tests (test/integration/contexts/)
```
test/integration/contexts/{context}/{aggregate}/
 usecase_test.go       ← Full E2E with real database
 repository_test.go    ← CRUD operations with real database
```

### Additional Tests
```
test/smoke/contexts/      ← HTTP handler validation (15 files)
test/benchmark/contexts/  ← Performance tests (2 files)
pkg/                      ← Package unit tests (27 files)
```

## Next Action

**Priority 1**: Complete `test/integration/contexts/customer-mgmt/interaction/usecase_test.go`
- Template: Use `deal/usecase_test.go` (24 tests)
- Methods: 14 UseCase methods to test
- Dependencies: Customer, User, Company UseCases

**See**: [Complete Audit Report](test-coverage-audit-complete.md) for detailed analysis
