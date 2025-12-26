# Workflows Module - Test Implementation Plan

##  Мета

Досягти 80%+ test coverage для комерційного workflows модуля через поетапне тестування всіх layers.

---

## Phase 1: Entity Layer Extended Tests (30 min)

### 1.1 WorkflowDefinition - розширені тести

**Файл:** `domain/entity/workflow_definition_test.go`

**Додати тести:**

```go
// Lifecycle methods
TestWorkflowDefinition_Deprecate           // active → deprecated
TestWorkflowDefinition_Archive             // any → archived
TestWorkflowDefinition_Lifecycle_Invalid   // invalid transitions

// Schema validation
TestWorkflowSchema_Validate_EmptyStates
TestWorkflowSchema_Validate_NoInitialState
TestWorkflowSchema_Validate_InitialStateNotFound
TestWorkflowSchema_Validate_InvalidTransition
TestWorkflowSchema_Validate_CircularTransition
TestWorkflowSchema_Validate_UnreachableStates

// Edge cases
TestNewWorkflowDefinition_EmptyName
TestNewWorkflowDefinition_LongDescription
TestWorkflowDefinition_Tags_Manipulation
TestWorkflowDefinition_Metadata_JSON
```

### 1.2 WorkflowInstance - розширені тести

**Файл:** `domain/entity/workflow_instance_test.go`

**Додати тести:**

```go
// Helper methods
TestWorkflowInstance_IsActive              // running/waiting = true
TestWorkflowInstance_IsCompleted           // completed/failed/cancelled = true
TestWorkflowInstance_IsTerminal            // can't transition further

// State timeout
TestWorkflowInstance_SetStateTimeout
TestWorkflowInstance_Timeout_Expired

// Due date
TestWorkflowInstance_SetDueDate
TestWorkflowInstance_IsOverdue

// Sub-workflows
TestWorkflowInstance_SetParent
TestWorkflowInstance_HasParent

// External reference
TestWorkflowInstance_SetExternalReference
TestWorkflowInstance_GetByExternalRef

// Tags
TestWorkflowInstance_AddTag
TestWorkflowInstance_RemoveTag
TestWorkflowInstance_HasTag
```

### 1.3 WorkflowStep - NEW

**Файл:** `domain/entity/workflow_step_test.go` (створити новий)

```go
TestNewWorkflowStep
TestWorkflowStep_MarkCompleted
TestWorkflowStep_MarkFailed
TestWorkflowStep_RecordOutput
TestWorkflowStep_Duration
```

### 1.4 WorkflowVariable - NEW

**Файл:** `domain/entity/workflow_variable_test.go` (створити новий)

```go
TestNewWorkflowVariable
TestWorkflowVariable_UpdateValue
TestWorkflowVariable_TypeValidation
TestWorkflowVariable_EncryptSensitive
```

**Оцінка часу:** 30 хвилин  
**Результат:** ~30 нових тестів, entity layer 90%+ coverage

---

## Phase 2: UseCase Layer Tests (2-3 hours)

### 2.1 Виправити існуючі моки (30 min)

**Файл:** `usecase/mocks_test.go`

**Додати методи до MockWorkflowDefinitionRepository:**

```go
GetByNameAndVersion(ctx, name, version) (*entity.WorkflowDefinition, error)
ListByStatus(ctx, status, limit, offset) ([]*entity.WorkflowDefinition, error)
ListByCategory(ctx, category, limit, offset) ([]*entity.WorkflowDefinition, error)
ListByCreator(ctx, creatorID, limit, offset) ([]*entity.WorkflowDefinition, error)
ListVersions(ctx, name) ([]*entity.WorkflowDefinition, error)
CountByStatus(ctx, status) (int64, error)
CountByCategory(ctx, category) (int64, error)
Search(ctx, query, limit, offset) ([]*entity.WorkflowDefinition, error)
```

**Додати методи до MockWorkflowInstanceRepository:**

```go
ListByStatus(ctx, status, limit, offset) ([]*entity.WorkflowInstance, error)
ListByUser(ctx, userID, limit, offset) ([]*entity.WorkflowInstance, error)
CountByDefinition(ctx, definitionID) (int64, error)
GetActiveByDefinition(ctx, definitionID) ([]*entity.WorkflowInstance, error)
ListByExternalRef(ctx, ref) ([]*entity.WorkflowInstance, error)
```

### 2.2 WorkflowDefinitionUseCase тести (1 hour)

**Файл:** `usecase/workflow_definition_usecase_test.go`

**Виправити + додати тести:**

```go
// CREATE операції
TestCreate_Success                          //  вже є, виправити назву
TestCreate_DuplicateName                    // NEW - перевірка на duplicate
TestCreate_ValidationError                  // NEW - invalid schema
TestCreate_RepositoryError                  //  вже є

// READ операції
TestGetByID_Success                         //  вже є
TestGetByID_NotFound                        //  вже є
TestGetByName_Success                       // NEW
TestGetByName_NotFound                      // NEW

// UPDATE операції
TestActivate_Success                        //  вже є
TestActivate_AlreadyActive                  // NEW
TestActivate_NotFound                       // NEW
TestUpdate_Success                          // NEW
TestUpdate_NotFound                         // NEW

// DELETE операції
TestDelete_Success                          //  вже є
TestDelete_NotFound                         // NEW
TestDelete_HasActiveInstances               // NEW - business rule

// LIST операції
TestListByStatus_Success                    // NEW
TestListByCategory_Success                  // NEW
TestListByCreator_Success                   // NEW
TestSearch_Success                          // NEW
TestSearch_NoResults                        // NEW
```

### 2.3 WorkflowInstanceUseCase тести (1-1.5 hours)

**Файл:** `usecase/workflow_instance_usecase_test.go` (створити новий)

```go
// START операції
TestStartInstance_Success
TestStartInstance_DefinitionNotActive       // can't start from draft/deprecated
TestStartInstance_DefinitionNotFound
TestStartInstance_InvalidInput

// STATE TRANSITIONS
TestTransitionState_Success
TestTransitionState_InvalidTransition       // transition not allowed
TestTransitionState_NotRunning              // can't transition if paused/completed
TestTransitionState_DefinitionMismatch      // state doesn't exist in definition

// COMPLETE операції
TestCompleteInstance_Success
TestCompleteInstance_NotRunning
TestCompleteInstance_AlreadyCompleted

// FAIL операції
TestFailInstance_Success
TestFailInstance_WithRetry                  // increment retry, restart
TestFailInstance_MaxRetriesExceeded         // mark as failed permanently

// CANCEL операції
TestCancelInstance_Success
TestCancelInstance_AlreadyCompleted

// PAUSE/RESUME операції
TestPauseInstance_Success
TestPauseInstance_NotRunning
TestResumeInstance_Success
TestResumeInstance_NotPaused

// READ операції
TestGetByID_Success
TestGetByID_NotFound
TestListByDefinition_Success
TestListByUser_Success
TestCountByDefinition_Success

// BUSINESS LOGIC
TestTimeout_AutomaticHandling               // background job handling
TestRetry_ExponentialBackoff
TestStateTimeout_Warning                    // approaching timeout
```

**Оцінка часу:** 2-3 години  
**Результат:** ~60 нових тестів, UseCase layer 80%+ coverage

---

## Phase 3: Repository Integration Tests (3-4 hours)

### 3.1 Test Database Setup (30 min)

**Файл:** `test/integration/workflows_test_helper.go`

```go
// Setup test database on port 5433
func SetupWorkflowsTestDB(t *testing.T) *sqlx.DB
func TeardownWorkflowsTestDB(t *testing.T, db *sqlx.DB)
func CleanWorkflowsTables(t *testing.T, db *sqlx.DB)

// Fixtures
func CreateTestDefinition(t *testing.T, db *sqlx.DB) *entity.WorkflowDefinition
func CreateTestInstance(t *testing.T, db *sqlx.DB, defID uuid.UUID) *entity.WorkflowInstance
```

### 3.2 WorkflowDefinitionRepository тести (1 hour)

**Файл:** `adapter/repository/postgres/workflow_definition_repository_test.go`

```go
TestCreate_Success
TestCreate_Duplicate                        // unique constraint
TestGetByID_Success
TestGetByID_NotFound
TestGetByName_Success                       // latest active version
TestGetByName_NotFound
TestGetByNameAndVersion_Success
TestUpdate_Success
TestUpdate_NotFound
TestDelete_Success                          // soft delete
TestDelete_VerifyDeletedAt
TestListByStatus_Pagination
TestListByCategory_Success
TestListVersions_OrderByVersion
TestCountByStatus_Success
TestSearch_ByName
TestSearch_ByDescription
TestTransactions_Rollback                   // важливо!
```

### 3.3 WorkflowInstanceRepository тести (1 hour)

**Файл:** `adapter/repository/postgres/workflow_instance_repository_test.go`

```go
TestCreate_Success
TestCreate_WithForeignKey                   // definition_id exists
TestGetByID_Success
TestGetByID_NotFound
TestUpdate_Success
TestDelete_Success                          // soft delete
TestListByDefinition_Pagination
TestListByStatus_Success
TestListByUser_Success
TestCountByDefinition_Success
TestGetActiveByDefinition_Success
TestListByExternalRef_Success
TestTransactions_MultipleUpdates
TestConcurrency_OptimisticLocking          // важливо для state machine!
```

### 3.4 Інші Repositories (1-1.5 hours)

**Файли:**

- `workflow_step_repository_test.go`
- `workflow_event_repository_test.go`
- `workflow_variable_repository_test.go`

Базові тести для кожного (~10-15 tests per repo)

**Оцінка часу:** 3-4 години  
**Результат:** ~80 нових тестів, Repository layer 70%+ coverage

---

## Phase 4: Handler Tests (2-3 hours)

### 4.1 Test Helpers (30 min)

**Файл:** `adapter/http/handler/test_helpers.go`

```go
// Mock gin.Context
func MockGinContext(method, path string, body interface{}) *gin.Context

// Mock JWT auth
func MockAuthContext(userID uuid.UUID) *gin.Context

// Response assertions
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int)
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int, expectedError string)
```

### 4.2 WorkflowDefinitionHandler тести (1 hour)

**Файл:** `adapter/http/handler/workflow_definition_handler_test.go`

```go
// POST /definitions
TestCreate_Success_201
TestCreate_ValidationError_400
TestCreate_Unauthorized_401
TestCreate_DuplicateName_409

// GET /definitions/:id
TestGetByID_Success_200
TestGetByID_NotFound_404

// PUT /definitions/:id
TestUpdate_Success_200
TestUpdate_NotFound_404
TestUpdate_ValidationError_400

// POST /definitions/:id/activate
TestActivate_Success_200
TestActivate_NotFound_404
TestActivate_AlreadyActive_409

// DELETE /definitions/:id
TestDelete_Success_204
TestDelete_NotFound_404
TestDelete_HasActiveInstances_409

// GET /definitions
TestList_Success_200
TestList_Pagination
TestList_FilterByStatus
TestList_FilterByCategory
TestList_Search
```

### 4.3 WorkflowInstanceHandler тести (1 hour)

**Файл:** `adapter/http/handler/workflow_instance_handler_test.go`

```go
// POST /instances
TestStartInstance_Success_201
TestStartInstance_ValidationError_400
TestStartInstance_DefinitionNotActive_409

// GET /instances/:id
TestGetByID_Success_200
TestGetByID_NotFound_404

// POST /instances/:id/transition
TestTransition_Success_200
TestTransition_InvalidTransition_400
TestTransition_NotRunning_409

// POST /instances/:id/complete
TestComplete_Success_200
TestComplete_NotRunning_409

// POST /instances/:id/fail
TestFail_Success_200
TestFail_WithRetry_200

// POST /instances/:id/cancel
TestCancel_Success_200
TestCancel_AlreadyCompleted_409

// POST /instances/:id/pause
TestPause_Success_200

// POST /instances/:id/resume
TestResume_Success_200
TestResume_NotPaused_409

// GET /instances
TestList_Success_200
TestList_ByDefinition
TestList_ByUser
TestList_Pagination
```

### 4.4 DTO Validation тести (30 min)

**Файл:** `adapter/http/dto/workflow_dto_test.go`

```go
TestCreateDefinitionRequest_Validation
TestUpdateDefinitionRequest_Validation
TestStartInstanceRequest_Validation
TestTransitionRequest_Validation
```

**Оцінка часу:** 2-3 години  
**Результат:** ~60 нових тестів, Handler layer 75%+ coverage

---

## Phase 5: Integration Tests (2 hours)

### 5.1 End-to-End Scenarios

**Файл:** `test/integration/workflows_e2e_test.go`

```go
// Happy path
TestE2E_CompleteWorkflow
  1. Create definition
  2. Activate definition
  3. Start instance
  4. Transition through states
  5. Complete successfully
  6. Verify final state

// Pause/Resume
TestE2E_PauseResumeWorkflow
  1. Create + activate definition
  2. Start instance
  3. Transition to middle state
  4. Pause
  5. Resume
  6. Complete

// Failure scenarios
TestE2E_WorkflowFailure
  1. Create + activate definition
  2. Start instance
  3. Transition to failing state
  4. Fail with error details
  5. Verify error recorded

// Retry logic
TestE2E_RetryMechanism
  1. Create definition with retry policy
  2. Start instance
  3. Fail (trigger retry)
  4. Verify retry count incremented
  5. Eventually succeed or max retries

// Cancellation
TestE2E_CancelWorkflow
  1. Start instance
  2. User cancels
  3. Verify cancelled state
  4. Cannot resume

// Multiple instances
TestE2E_MultipleInstances
  1. Create definition
  2. Start 5 instances
  3. Each in different states
  4. Verify independent execution

// Sub-workflows
TestE2E_ParentChildWorkflow
  1. Create parent definition
  2. Create child definition
  3. Start parent
  4. Parent spawns child
  5. Child completes
  6. Parent continues

// Timeout
TestE2E_TimeoutHandling
  1. Create definition with state timeout
  2. Start instance
  3. Wait for timeout
  4. Verify auto-timeout

// Version upgrade
TestE2E_VersionUpgrade
  1. Create definition v1
  2. Start instances
  3. Create definition v2 (same name)
  4. New instances use v2
  5. Old instances continue on v1
```

**Оцінка часу:** 2 години  
**Результат:** ~10 scenarios, critical paths validated

---

## Phase 6: Smoke Tests (1 hour)

### 6.1 Module Health

**Файл:** `test/smoke/workflows_smoke_test.go`

```go
TestModuleLoads
TestMigrationsApplied
TestHealthEndpoint
TestLicenseValidation              // commercial module!
TestBasicCRUD
TestDatabaseConnectivity
TestEventBusConnectivity
```

**Оцінка часу:** 1 година  
**Результат:** ~7 smoke tests, deployment validation

---

##  Підсумок по phases

| Phase                        | Time            | Tests          | Coverage Goal    |
| ---------------------------- | --------------- | -------------- | ---------------- |
| **Phase 1:** Entity Extended | 30 min          | +30            | 90%+             |
| **Phase 2:** UseCase         | 2-3 hours       | +60            | 80%+             |
| **Phase 3:** Repository      | 3-4 hours       | +80            | 70%+             |
| **Phase 4:** Handler         | 2-3 hours       | +60            | 75%+             |
| **Phase 5:** Integration     | 2 hours         | +10            | E2E validation   |
| **Phase 6:** Smoke           | 1 hour          | +7             | Deployment ready |
| **TOTAL**                    | **10-13 hours** | **~247 tests** | **80%+ overall** |

---

##  Milestone Goals

### Milestone 1: UseCase Layer Complete (3 hours)

-  All mocks implement full interfaces
-  All UseCase methods tested
-  80%+ UseCase coverage
-  **Ready for:** Code review

### Milestone 2: Repository Layer Complete (6-7 hours)

-  All repositories tested
-  Transaction scenarios covered
-  70%+ Repository coverage
-  **Ready for:** Integration testing

### Milestone 3: Handler Layer Complete (8-10 hours)

-  All endpoints tested
-  Validation scenarios covered
-  75%+ Handler coverage
-  **Ready for:** API documentation

### Milestone 4: Full Coverage (10-13 hours)

-  Integration tests pass
-  Smoke tests pass
-  80%+ overall coverage
-  **Ready for:** Commercial release

---

##  Critical Tests (Must Have)

### State Machine Logic

-  All valid transitions tested
-  All invalid transitions rejected
-  Concurrent state updates handled

### Data Integrity

-  Foreign key constraints validated
-  Soft deletes work correctly
-  Transactions rollback properly

### Business Rules

-  Can't start instance from inactive definition
-  Can't delete definition with active instances
-  Retry limits enforced
-  Timeout handling works

### Security

-  Authorization checked on all endpoints
-  User can only see their workflows
-  Admin can see all workflows

---

##  Next Immediate Action

**Розпочати з Phase 2.1:**

1. Відкрити `usecase/mocks_test.go`
2. Додати відсутні методи до моків
3. Перекомпілювати тести
4. Виправити назви методів у тестах
5. Переконатись що всі UseCase тести компілюються

**Команда для перевірки:**

```bash
go test ./internal/modules/workflows/usecase -v -count=1
```

**Очікуваний результат:**

-  Компіляція успішна
-  9-15 тестів проходять
-  UseCase layer базово покрито

---

**Створено:** 26 грудня 2025  
**Оновлено:** Never  
**Статус:**  Planning Complete, Ready to Execute
