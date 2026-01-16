# Script Context - Domain Errors & Testing Guide

## Overview

**Script Context** implements LUA script management with **Phase 2 Domain Errors pattern** - centralized error constants, zero inline fmt.Errorf, and type-safe error handling via errors.Is().

---

## Domain Errors Pattern (Phase 2)

### errors.go - Central Error Registry

All domain errors are defined as constants in `errors.go` (21 total):

**Repository Errors** (2):
- `ErrScriptNotFound` - Script doesn't exist
- `ErrScriptAlreadyExists` - Duplicate script name

**Business Logic Errors** (10):
- `ErrInvalidScriptName` - Name validation failed
- `ErrEmptyCode` - Code cannot be empty
- `ErrInvalidScriptType` - Unknown script type
- `ErrScriptNotActive` - Script is inactive
- `ErrScriptNotValidated` - Validation required before activation
- `ErrInvalidMetadataKey` - Metadata key validation failed
- `ErrMetadataKeyRequired` - Key cannot be empty
- `ErrMetadataValueRequired` - Value cannot be empty
- `ErrMetadataNotFound` - Key doesn't exist
- `ErrDeleteNonExistentMetadata` - Cannot delete missing key

**Technical Operation Errors** (7):
- `ErrScriptCreateFailed` - Creation failed
- `ErrScriptUpdateFailed` - Update failed
- `ErrScriptDeleteFailed` - Deletion failed
- `ErrScriptListFailed` - List operation failed
- `ErrScriptMetadataFailed` - Metadata operation failed
- `ErrScriptActivateFailed` - Activation failed
- `ErrScriptDeactivateFailed` - Deactivation failed

**Execution Errors** (2):
- `ErrExecutionAlreadyFinished` - Cannot update finished execution
- `ErrExecutionUpdateFailed` - Update failed

### Usage Pattern

**Domain Layer** (entity.go, usecase.go, execution.go):
```go
//  OLD - NEVER DO THIS
return nil, fmt.Errorf("script not found")

//  NEW - ALWAYS USE CONSTANTS
return nil, ErrScriptNotFound
```

**Handler Layer** (handler.go):
```go
// Map domain errors to HTTP codes via errors.Is()
scr, err := h.usecase.GetScript(ctx, name)
if err != nil {
    if errors.Is(err, ErrScriptNotFound) {
        response.NotFound(c, "Script not found")  // 404
        return
    }
    if errors.Is(err, ErrScriptNotActive) {
        response.BadRequest(c, "Script is not active")  // 400
        return
    }
    response.InternalError(c, "Failed to retrieve script")  // 500
    return
}
```

**Test Layer**:
```go
// Type-safe error checking
_, err := usecase.GetScript(ctx, "missing")
assert.Error(t, err)
assert.True(t, errors.Is(err, ErrScriptNotFound))
```

---

## API Changes (Phase 2 Entity Refactoring)

### NewScript Signature

**OLD** (before Phase 2):
```go
s, err := script.NewScript("my_script", "return 2 + 2")
```

**NEW** (Phase 2+):
```go
s, err := script.NewScript("my_script", "return 2 + 2", script.ScriptTypeValidation)
// Parameters: name, code, scriptType
// ScriptType options:
//   ScriptTypeValidation   - Input validation scripts
//   ScriptTypeWorkflow     - Business workflow automation
//   ScriptTypeReport       - Report generation
//   ScriptTypePricing      - Dynamic pricing calculations
//   ScriptTypeNotification - Notification triggers
//   ScriptTypeAutomation   - General automation tasks
//   ScriptTypeCustom       - Custom/miscellaneous scripts
```

### Metadata Access

**OLD** (direct map access):
```go
author := script.Metadata["author"]
script.Metadata["version"] = "1.0.0"
```

**NEW** (jsonstore.Field with .Get() method):
```go
// Read metadata
metadata := script.Metadata.Get()
author := metadata["author"]

// Modify metadata (use entity methods)
script.UpdateMetadata("version", "1.0.0")
script.DeleteMetadata("author")
```

---

## Testing Strategy

### 3-Tier Test Architecture

**1. Unit Tests** (`usecase_test.go`):
- **Location**: Same directory as code
- **Purpose**: Test business logic in isolation
- **Dependencies**: No database, mocked repository
- **Coverage**: All usecase methods (17 total)
- **Run**: `go test ./internal/contexts/scripting/script -v`
- **Status**:  PASSING (exit code 0)

**Example**:
```go
func TestUseCase_CreateScript(t *testing.T) {
    mockRepo := &MockRepository{
        CreateFunc: func(ctx context.Context, s *Script) error {
            return nil
        },
    }
    uc := NewUseCase(mockRepo)
    
    s, err := uc.CreateScript(ctx, "test", "return 1", ScriptTypeGeneral)
    assert.NoError(t, err)
    assert.Equal(t, "test", s.Name)
}
```

**2. Smoke Tests** (`test/smoke/contexts/scripting/script/handler_test.go`):
- **Location**: `test/smoke/` mirror path
- **Purpose**: HTTP handler validation (80/20 rule)
- **Dependencies**: MockScriptUseCase with function fields
- **Coverage**: HTTP status codes, response format, routing
- **Run**: `go test ./test/smoke/contexts/scripting/... -v`
- **Key Pattern**: Test endpoints with minimal setup, no database

**Example**:
```go
type MockScriptUseCase struct {
    GetScriptFunc func(ctx context.Context, name string) (*Script, error)
}

func TestScriptHandler_GetByName_Success(t *testing.T) {
    router := smoke.SetupRouter()
    
    mockUC := &MockScriptUseCase{
        GetScriptFunc: func(ctx context.Context, name string) (*Script, error) {
            return fakeScript(), nil
        },
    }
    
    handler := NewScriptHandler(mockUC)
    router.GET("/scripts/:name", handler.GetByName)
    
    resp := smoke.MakeRequest(t, router, "GET", "/scripts/test", nil)
    smoke.AssertSuccessResponse(t, resp, 200)
}
```

**3. Integration Tests** (`test/integration/contexts/scripting/script/repository_test.go`):
- **Location**: `test/integration/` mirror path
- **Purpose**: Full E2E with real PostgreSQL database
- **Dependencies**: integration.SetupTestDB(t), testDB.WithTransaction
- **Coverage**: Repository CRUD, List, Metadata, Execution tracking, Pagination
- **Run**: `go test ./test/integration/contexts/scripting/... -v`
- **Status**:  FIXED (Phase 2 API migration complete)

**Test Functions**:
1. `TestScriptRepository_CRUD` - Create, Read, Update, Delete
2. `TestScriptRepository_List` - List with status filter
3. `TestScriptRepository_Metadata` - Metadata management
4. `TestScriptRepository_Execution` - Execution tracking
5. `TestScriptRepository_Errors` - Error scenarios
6. `TestScriptRepository_Pagination` - Pagination logic

**Example**:
```go
func TestScriptRepository_CRUD(t *testing.T) {
    testDB := integration.SetupTestDB(t)
    repo := postgres.NewScriptRepository(testDB.DB)
    
    testDB.WithTransaction(t, func(ctx context.Context) {
        // Create
        s, err := script.NewScript("test", "return 1", script.ScriptTypeGeneral)
        require.NoError(t, err)
        require.NoError(t, repo.Create(ctx, s))
        
        // Read
        found, err := repo.GetByID(ctx, s.ID)
        require.NoError(t, err)
        assert.Equal(t, s.Name, found.Name)
        
        // Update
        found.Description = "Updated"
        require.NoError(t, repo.Update(ctx, found))
        
        // Delete
        require.NoError(t, repo.Delete(ctx, s.ID))
    })
}
```

---

## Swagger/OpenAPI Documentation

All HTTP endpoints are documented with swaggo annotations:

```go
// @Summary Execute script
// @Description Execute LUA script by name with parameters
// @Tags scripts
// @Accept json
// @Produce json
// @Param name path string true "Script name"
// @Param request body dto.ExecuteScriptRequest false "Execution parameters"
// @Success 200 {object} dto.ExecutionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /scripts/{name}/execute [post]
func (h *ScriptHandler) ExecuteScript(c *gin.Context) {
    // Implementation
}
```

**Generate docs**: `make swagger-generate`  
**View docs**: http://localhost:8081/api/docs/index.html (after `make dev`)

---

## Common Issues & Solutions

### Issue: "not enough arguments in call to script.NewScript"

**Cause**: NewScript signature changed in Phase 2, now requires ScriptType as 3rd parameter

**Solution**:
```go
//  OLD
s, err := script.NewScript("test", "return 1")

//  NEW
s, err := script.NewScript("test", "return 1", script.ScriptTypeCustom)
// Use appropriate ScriptType: Validation, Workflow, Report, Pricing, Notification, Automation, Custom
```

### Issue: "cannot index script.Metadata"

**Cause**: Metadata changed from map to jsonstore.Field[map[string]string]

**Solution**:
```go
//  OLD
author := script.Metadata["author"]
script.Metadata["version"] = "1.0.0"

//  NEW
metadata := script.Metadata.Get()
author := metadata["author"]

// Or use entity methods
script.UpdateMetadata("version", "1.0.0")
script.DeleteMetadata("author")
```

### Issue: ST1001 lint warning (dot imports)

**Cause**: `handler.go` uses dot import for script package

**Current**: Accepted with comment (improves readability in handlers)
```go
// Using dot import for cleaner error handling in HTTP layer
//
//nolint:staticcheck // dot import improves readability
. "github.com/basilex/promenade/internal/contexts/scripting/script"
```

**Alternative**: Remove dot import and add `script.` prefix to all error constants (100+ occurrences)

---

## Best Practices

### DO

 Use domain error constants from `errors.go`  
 Map errors via `errors.Is()` in handlers  
 Use `NewScript(name, code, scriptType)` with all 3 parameters  
 Access metadata via `.Get()` method or entity methods  
 Write smoke tests for all HTTP handlers  
 Write integration tests for repository methods  
 Use `testDB.WithTransaction` for automatic rollback

### DON'T

 Use `fmt.Errorf()` in domain layer (use constants only)  
 Use string comparison for errors (use `errors.Is()`)  
 Direct map access on `script.Metadata` (use `.Get()` or methods)  
 Forget ScriptType parameter in `NewScript()`  
 Skip transaction wrapper in integration tests

---

## Related Documentation

- [Main README](../../../../../README.md) - Project overview
- [Documentation Index](../../../../../docs/INDEX.md) - All guides
- [Testing Guide](../../../../../test/README.md) - Testing infrastructure
- [Smoke Tests Guide](../../../../../test/smoke/README.md) - HTTP handler testing
- [Domain Errors Guide](../../../../../docs/guides/domain-errors.md) - Error handling patterns

---

**Status**:  Production-ready (Phase 2 complete)  
**Test Coverage**: 3-tier (unit + smoke + integration)  
**Domain Errors**: 21 constants, 0 fmt.Errorf  
**Last Updated**: January 13, 2026
