# Phase 3: Contract Aggregate - Progress Checkpoint

**Date**: January 8, 2026  
**Status**: Task 3.2  COMPLETE → Starting Task 3.3 (UseCase)  
**Overall Phase Progress**: 2/5 tasks complete (40%) → 3/5 in progress (60%)  
**Strategic Priority**: Contract → Saga → LUA (business logic before infrastructure)

---

## Quick Summary

**What's Working**:
- Task 3.1: Contract Entity (COMPLETE, committed `a0048b1`)
-  Task 3.2: Contract Repository (COMPLETE, committed `f78a281`)
  -  Repository compiles with 0 errors
  -  ALL 13/13 integration tests PASSING (100%)
  -  Helper functions (createTestCustomer, createTestOrder) working
  -  WithTransaction pattern applied to all tests
  -  FK dependencies handled correctly
  -  SQL INTERVAL syntax fixed in ListExpiringSoon

**What's Next**:
-  Task 3.3: Contract UseCase (business logic layer, 2-3 hours)

**Strategic Plan (January 8, 2026)**:
1. Complete Task 3.2 (12 tests) - TODAY
2. Task 3.3: Contract UseCase (2-3 hours)
3. Task 3.4: HTTP API (2-3 hours)
4. Task 3.5: Migrations (30 min)
5. Phase 4: Fulfillment Saga (Week 6-7)
6. Return to LUA Week 2-3 (Script Storage + UI Metadata)

**Rationale**: Business logic (Contract, Saga) takes priority over infrastructure (LUA storage). Core engine already operational with 33 tests and 10 endpoints.

---

## Detailed Progress Report

### Task 3.1: Contract Entity  COMPLETE

**Status**: 100% complete, committed in `a0048b1`

**Deliverables**:
-  Contract aggregate root entity (194 lines)
-  ContractStatus enum (draft, pending, active, completed, terminated, renewed)
-  Business rules implemented:
  - Draft → Pending transition (requires signature + terms)
  - Pending → Active transition (requires all party signatures)
  - Auto-expiration checks
  - Termination with reason
  - Contract renewal
-  40 passing entity tests (100% coverage)
-  Error definitions (ErrContractNotFound, ErrInvalidContractTransition)

**Files**:
- `internal/contexts/order-mgmt/contract/entity.go`
- `internal/contexts/order-mgmt/contract/entity_test.go`

**Git**: Committed in `a0048b1`

---

### Task 3.2: Contract Repository  98% COMPLETE

**Status**: Code complete, 2/14 tests passing, 12 tests need FK fixes

#### What Was Achieved This Session

**1. Infrastructure Pattern Resolution** 
- **Problem**: Used non-existent `database.Executor` interface
- **Solution**: Use `database.GetTx(ctx)` + `sqlx.ExtContext` pattern
- **Result**: BaseRepository working correctly

**2. Entity API Corrections** 
- **Problem**: Assumed BaseAggregate had SetID(), SetDeletedAt(), GetDeletedAt() methods
- **Discovery**: BaseAggregate has public fields (ID, DeletedAt *time.Time)
- **Solution**: Use direct field access (`c.ID = id`, `c.DeletedAt = &time`)
- **Result**: All entity access patterns correct

**3. Missing Entity Errors** 
- **Problem**: Repository needed `ErrContractNotFound` for not found conditions
- **Solution**: Added error definitions to entity.go
- **Result**: Proper domain error handling

**4. BaseRepository.NamedExec Fix** 
- **Problem**: Used `executor.NamedExecContext()` method (doesn't exist)
- **Solution**: Use `sqlx.NamedExecContext(ctx, executor, query, arg)` helper function
- **Result**: Named queries work correctly

**5. SQL Placeholder Conversion** 
- **Problem**: All queries used `?` placeholders (wrong for PostgreSQL)
- **Solution**: Converted all placeholders to `$1, $2, $3` format
- **Result**: SQL syntax correct for PostgreSQL

**6. Test FK Dependencies**  PARTIAL
- **Problem**: Tests created contracts with fake Order/Customer IDs (FK violation)
- **Solution**: Created helper functions to insert dependencies
- **Result**: 
  -  `TestContractRepository_Create` - PASSES
  -  `TestContractRepository_GetByID_NotFound` - PASSES  
  -  12 remaining tests need same FK setup pattern

#### Implementation Details

**BaseRepository** (`base_repository.go` - 54 lines):
```go
// Correct infrastructure pattern
func (r *BaseRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
    if tx, ok := database.GetTx(ctx); ok && tx != nil {
        return tx
    }
    return r.db
}

// Fixed NamedExec - using helper function, not method
func (r *BaseRepository) NamedExec(ctx context.Context, query string, arg interface{}) error {
    _, err := sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
    return err
}
```

**Contract Repository** (`contract_repository.go` - 516 lines):
-  All 11 methods implemented
-  PostgreSQL placeholders ($1, $2, $3)
-  Direct field access (c.ID, c.OrderID, c.Status, c.DeletedAt)
-  Compiles with 0 errors

**Methods Implemented**:
1. Create(ctx, *Contract) error
2. GetByID(ctx, UUID) (*Contract, error)
3. Update(ctx, *Contract) error
4. Delete(ctx, UUID) error
5. List(ctx, offset, limit) ([]*Contract, int, error)
6. GetByOrder(ctx, UUID) ([]*Contract, error)
7. GetByCustomer(ctx, UUID) ([]*Contract, error)
8. GetActiveContracts(ctx) ([]*Contract, error)
9. ListByStatus(ctx, status, offset, limit) ([]*Contract, int, error)
10. ListByCustomer(ctx, customerID, offset, limit) ([]*Contract, int, error)
11. ListExpiringSoon(ctx, days) ([]*Contract, error)

**Test Helpers** (in `repository_test.go`):
```go
func createTestCustomer(ctx context.Context, tx *sqlx.Tx) (uuidv7.UUID, error) {
    customerID := uuidv7.New()
    assignedTo := uuidv7.New()
    _, err := tx.ExecContext(ctx, `
        INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", 
       "customer", "free", "direct", assignedTo)
    return customerID, err
}

func createTestOrder(ctx context.Context, tx *sqlx.Tx, customerID uuidv7.UUID) (uuidv7.UUID, error) {
    orderID := uuidv7.New()
    _, err := tx.ExecContext(ctx, `
        INSERT INTO order_orders (id, order_number, customer_id, total_amount, currency, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    `, orderID, "ORD-2026-000001", customerID, 0.0, "USD", "pending")
    return orderID, err
}
```

#### What Remains for Task 3.2

**Mechanical Test Fixes** (~10 minutes work):

12 tests need FK dependency setup. Pattern to apply:

```go
// BEFORE (current broken pattern)
func TestContractRepository_GetByID(t *testing.T) {
    db := integration.SetupTestDBWithCleanTables(t)
    repo := contractRepo.NewContractRepository(db.DB)
    ctx := context.Background()

    orderID := uuidv7.New()  //  Fake ID - FK violation
    customerID := uuidv7.New()  //  Fake ID - FK violation
    c := contract.NewContract(orderID, customerID, "Test terms")
    require.NoError(t, repo.Create(ctx, c))
    // ... rest of test
}

// AFTER (working pattern from Create test)
func TestContractRepository_GetByID(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    testDB := integration.SetupTestDB(t)
    testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
        repo := contractRepo.NewContractRepository(testDB.DB)
        
        //  Create real dependencies
        customerID, err := createTestCustomer(ctx, tx)
        require.NoError(t, err)
        orderID, err := createTestOrder(ctx, tx, customerID)
        require.NoError(t, err)
        
        //  Use real IDs
        c := contract.NewContract(orderID, customerID, "Test terms")
        require.NoError(t, repo.Create(ctx, c))
        // ... rest of test
    })
}
```

**Tests Needing Fixes**:
1.  TestContractRepository_GetByID
2.  TestContractRepository_Update
3.  TestContractRepository_Delete
4.  TestContractRepository_List
5.  TestContractRepository_List_Pagination
6.  TestContractRepository_GetByOrder
7.  TestContractRepository_GetByCustomer
8.  TestContractRepository_GetActiveContracts
9.  TestContractRepository_ListByStatus
10.  TestContractRepository_ListByCustomer
11.  TestContractRepository_ListExpiringSoon
12. (1 more if exists)

**Already Working**:
-  TestContractRepository_Create
-  TestContractRepository_GetByID_NotFound (no FK needed)

---

## Lessons Learned

### Infrastructure Patterns

**1. BaseAggregate API** (Critical Discovery):
- **ID**: Public field, NOT `SetID()` method
- **DeletedAt**: Public pointer field `*time.Time`, NOT `SetDeletedAt()/GetDeletedAt()` methods
- **CreatedAt/UpdatedAt**: Have both `SetCreatedAt()` method AND `GetCreatedAt()` method
- **Version**: Public field

**Correct Usage**:
```go
//  CORRECT
c.ID = id                           // Direct field
c.DeletedAt = &time                 // Direct pointer
c.SetCreatedAt(time)                // Method exists
timestamp := c.GetCreatedAt()       // Method exists

//  WRONG
c.SetID(id)                         // Method doesn't exist
c.SetDeletedAt(time)                // Method doesn't exist
timestamp := c.GetDeletedAt()       // Method doesn't exist
```

**2. sqlx Helpers vs Methods**:
- Some operations are **helper functions**: `sqlx.NamedExecContext(ctx, executor, query, arg)`
- Some operations are **methods**: `executor.ExecContext(ctx, query, args...)`
- Always check documentation for correct signature

**3. PostgreSQL Placeholders**:
- Use `$1, $2, $3` (NOT `?`) for positional parameters
- Use `:name` for named parameters with `NamedExec`
- Never mix placeholder styles in same query

### Testing Patterns

**1. FK Dependencies**:
- Integration tests MUST create real foreign key records
- Use helper functions (`createTestCustomer`, `createTestOrder`)
- Insert dependencies in transaction before testing aggregate

**2. Transaction Pattern**:
```go
testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
    // All operations in this block use same transaction
    // Auto-rollback on test end
})
```

**3. Test Organization**:
-  Helper functions at top of test file
-  Each test self-contained with own dependencies
-  Use `testing.Short()` skip for long integration tests

---

## Files Modified (Uncommitted)

### Implementation Files
1. `internal/contexts/order-mgmt/contract/entity.go` (194 lines)
   - Added: `ErrContractNotFound`, `ErrInvalidContractTransition`
   
2. `internal/contexts/order-mgmt/contract/adapter/repository/postgres/base_repository.go` (54 lines)
   - Fixed: `NamedExec` to use `sqlx.NamedExecContext` helper

3. `internal/contexts/order-mgmt/contract/adapter/repository/postgres/contract_repository.go` (516 lines)
   - Fixed: All SQL placeholders (`?` → `$1, $2, $3`)
   - Fixed: Entity field access (direct fields instead of Set/Get methods)
   - Status: Compiles with 0 errors 

### Test Files
4. `test/integration/contexts/order-mgmt/contract/repository_test.go` (333 lines)
   - Added: Helper functions (`createTestCustomer`, `createTestOrder`)
   - Fixed: `TestContractRepository_Create` - PASSES 
   - Fixed: Transaction pattern in Create test
   - TODO: Fix 12 remaining tests with same pattern

### Migration Files
5. `migrations/order-mgmt/000003_add_contracts.up.sql`
   - Applied to database 

---

## How to Continue

### Option 1: Finish Task 3.2 Tests (Recommended, ~10 minutes)

**Action**: Apply FK dependency pattern to 12 remaining tests

**Steps**:
1. Open `test/integration/contexts/order-mgmt/contract/repository_test.go`
2. For each failing test, replace setup with:
   ```go
   testDB := integration.SetupTestDB(t)
   testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
       repo := contractRepo.NewContractRepository(testDB.DB)
       customerID, _ := createTestCustomer(ctx, tx)
       orderID, _ := createTestOrder(ctx, tx, customerID)
       // ... existing test code with real IDs
   })
   ```
3. Run: `go test ./test/integration/contexts/order-mgmt/contract -v`
4. Commit: `git commit -m "feat(contract): complete repository integration tests (Task 3.2)"`

**Result**: Task 3.2 100% complete

### Option 2: Move to Task 3.3 (Skip test fixes for now)

**Action**: Start Contract UseCase implementation

**Steps**:
1. Create `internal/contexts/order-mgmt/contract/usecase.go`
2. Define `IContractUseCase` interface
3. Implement business operations
4. Create `usecase_test.go` with mocks
5. Circle back to repository tests later

**Result**: Progress to Task 3.3 (60% Phase 3)

---

## Phase 3 Roadmap Remaining

### Task 3.3: Contract UseCase (Not Started)
**Estimate**: 2-3 hours

**Deliverables**:
- IContractUseCase interface (8-10 business methods)
- UseCase implementation with business logic
- Unit tests with repository mocks (40+ tests)

**Methods**:
- CreateContract(ctx, orderID, customerID, terms) (*Contract, error)
- GetContract(ctx, contractID) (*Contract, error)
- UpdateContract(ctx, contractID, terms, termsURL) error
- SignContract(ctx, contractID, signedByName, signedByEmail) error
- ActivateContract(ctx, contractID) error
- CompleteContract(ctx, contractID) error
- TerminateContract(ctx, contractID, reason) error
- RenewContract(ctx, contractID) (*Contract, error)

### Task 3.4: Contract HTTP API (Not Started)
**Estimate**: 3-4 hours

**Deliverables**:
- HTTP handlers (12 endpoints)
- DTOs (request/response)
- Integration tests (24+ tests)
- Smoke tests (12+ tests)

**Endpoints**:
1. POST /api/v1/contracts - Create contract
2. GET /api/v1/contracts/:id - Get by ID
3. PUT /api/v1/contracts/:id - Update contract
4. DELETE /api/v1/contracts/:id - Delete contract
5. GET /api/v1/contracts - List contracts
6. GET /api/v1/orders/:orderId/contracts - Get by order
7. GET /api/v1/customers/:customerId/contracts - Get by customer
8. GET /api/v1/contracts/active - Get active contracts
9. POST /api/v1/contracts/:id/sign - Sign contract
10. POST /api/v1/contracts/:id/activate - Activate contract
11. POST /api/v1/contracts/:id/complete - Complete contract
12. POST /api/v1/contracts/:id/terminate - Terminate contract

### Task 3.5: Documentation (Not Started)
**Estimate**: 1 hour

**Deliverables**:
- Update `internal/contexts/order-mgmt/contract/README.md`
- Add Swagger annotations to handlers
- Update Phase 3 roadmap in `docs/roadmap/`

---

## Key Decisions Made

1. **BaseAggregate Usage**: Confirmed public fields for ID and DeletedAt, not getter/setter methods
2. **Infrastructure Pattern**: Use `database.GetTx(ctx)` + `sqlx.ExtContext` (matches all other repositories)
3. **SQL Placeholders**: PostgreSQL format `$1, $2, $3` for positional, `:name` for named
4. **Test Dependencies**: Create real FK records with helper functions (not fake UUIDs)
5. **Error Handling**: Domain errors defined in entity package (ErrContractNotFound)

---

## Next Session Start Here

**Quick Restart**:
```bash
cd /Users/basilex/Workspace/src/promenade

# Check current state
git status

# Option A: Finish repository tests
code test/integration/contexts/order-mgmt/contract/repository_test.go
# Apply FK pattern to 12 tests, run: make test-integration

# Option B: Start UseCase
touch internal/contexts/order-mgmt/contract/usecase.go
touch internal/contexts/order-mgmt/contract/usecase_test.go
# Define IContractUseCase interface

# Option C: Check what's working
go test ./test/integration/contexts/order-mgmt/contract -v -run="Create|NotFound"
# Should see 2 PASS
```

**Recommended**: Option A (finish tests) then commit before starting UseCase

---

## Statistics

**Phase 3 Progress**: 40% (2/5 tasks)
-  Task 3.1: Entity (100%)
-  Task 3.2: Repository (98%)
- ⏳ Task 3.3: UseCase (0%)
- ⏳ Task 3.4: HTTP API (0%)
- ⏳ Task 3.5: Documentation (0%)

**Code Written**:
- Entity: 194 lines + 40 tests
- Repository: 516 lines + helpers
- Tests: 2 passing, 12 need fixes

**Time Spent**: ~4 hours (infrastructure debugging + implementation)

**Overall Project**: 29% complete (2/7 phases)

---

## References

- [Phase 3 Roadmap](../roadmap/PHASE3_CONTRACT_AGGREGATE.md)
- [Order Management Guide](../concepts/order-management.md)
- [Clean Architecture Summary](../concepts/clean-architecture.md)
- [Testing Patterns](../guides/testing-patterns.md)

**Last Updated**: January 7, 2026, 16:05  
**Next Review**: When resuming Phase 3 work
