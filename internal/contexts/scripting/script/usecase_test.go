package script

import (
	"context"
	"errors"
	"testing"

	"github.com/basilex/promenade/pkg/scripting"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of IRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, script *Script) error {
	args := m.Called(ctx, script)
	return args.Error(0)
}

func (m *MockRepository) Update(ctx context.Context, script *Script) error {
	args := m.Called(ctx, script)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, scriptID uuidv7.UUID) error {
	args := m.Called(ctx, scriptID)
	return args.Error(0)
}

func (m *MockRepository) CreateVersion(ctx context.Context, version *ScriptVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockRepository) ListVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*ScriptVersion, int, error) {
	args := m.Called(ctx, scriptID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*ScriptVersion), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetByID(ctx context.Context, scriptID uuidv7.UUID) (*Script, error) {
	args := m.Called(ctx, scriptID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Script), args.Error(1)
}

func (m *MockRepository) GetByName(ctx context.Context, name string) (*Script, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Script), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, status ScriptStatus, limit, offset int) ([]*Script, int, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Script), args.Int(1), args.Error(2)
}

func (m *MockRepository) ListAll(ctx context.Context, limit, offset int) ([]*Script, int, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*Script), args.Int(1), args.Error(2)
}

func (m *MockRepository) CreateExecution(ctx context.Context, execution *ScriptExecution) error {
	args := m.Called(ctx, execution)
	return args.Error(0)
}

func (m *MockRepository) GetExecutionByID(ctx context.Context, executionID uuidv7.UUID) (*ScriptExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ScriptExecution), args.Error(1)
}

func (m *MockRepository) GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*ScriptExecution, int, error) {
	args := m.Called(ctx, scriptID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*ScriptExecution), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetRecentExecutions(ctx context.Context, limit int) ([]*ScriptExecution, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ScriptExecution), args.Error(1)
}

// MockEngine is a mock implementation of LUA engine
type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) Execute(ctx context.Context, code string, params map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, code, params)
	return args.Get(0), args.Error(1)
}

func (m *MockEngine) Validate(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

// Helper to create engine interface from mock
func (m *MockEngine) AsEngine() *scripting.Engine {
	// For testing, we'll just return nil and test the interface separately
	return nil
}

// ============================================================================
// ExecuteScript Tests
// ============================================================================

func TestUseCase_ExecuteScript(t *testing.T) {
	ctx := context.Background()
	executedBy := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		script := &Script{
			Name:   "test_script",
			Code:   "return 2 + 2",
			Status: ScriptStatusActive,
		}
		script.ID = uuidv7.New()

		params := map[string]interface{}{"x": 10}

		repo.On("GetByName", ctx, "test_script").Return(script, nil)
		engine.On("Execute", ctx, script.Code, params).Return(4, nil)
		repo.On("CreateExecution", ctx, mock.Anything).Return(nil)

		result, err := uc.ExecuteScript(ctx, "test_script", params, executedBy)

		require.NoError(t, err)
		assert.Equal(t, 4, result)
		repo.AssertExpectations(t)
		engine.AssertExpectations(t)
	})

	t.Run("script not found", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		repo.On("GetByName", ctx, "nonexistent").Return(nil, errors.New("not found"))

		result, err := uc.ExecuteScript(ctx, "nonexistent", nil, executedBy)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "script not found")
	})

	t.Run("script not active", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		script := &Script{
			Name:   "inactive_script",
			Status: ScriptStatusInactive,
		}

		repo.On("GetByName", ctx, "inactive_script").Return(script, nil)

		result, err := uc.ExecuteScript(ctx, "inactive_script", nil, executedBy)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "script is not active")
	})

	t.Run("execution error", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		script := &Script{
			Name:   "failing_script",
			Code:   "error('something went wrong')",
			Status: ScriptStatusActive,
		}
		script.ID = uuidv7.New()

		repo.On("GetByName", ctx, "failing_script").Return(script, nil)
		engine.On("Execute", ctx, script.Code, mock.Anything).Return(nil, errors.New("execution error"))
		repo.On("CreateExecution", ctx, mock.Anything).Return(nil)

		result, err := uc.ExecuteScript(ctx, "failing_script", nil, executedBy)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "script execution failed")
	})
}

// ============================================================================
// ValidateScript Tests
// ============================================================================

func TestUseCase_ValidateScript(t *testing.T) {
	ctx := context.Background()

	t.Run("valid script", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		code := "return 2 + 2"
		engine.On("Validate", code).Return(nil)

		err := uc.ValidateScript(ctx, code)

		assert.NoError(t, err)
		engine.AssertExpectations(t)
	})

	t.Run("empty code", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		err := uc.ValidateScript(ctx, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "script code cannot be empty")
	})

	t.Run("syntax error", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		code := "return 2 +"
		engine.On("Validate", code).Return(errors.New("syntax error"))

		err := uc.ValidateScript(ctx, code)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "syntax validation failed")
	})
}

// ============================================================================
// CreateScript Tests
// ============================================================================

func TestUseCase_CreateScript(t *testing.T) {
	ctx := context.Background()
	createdBy := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		code := "return 2 + 2"
		engine.On("Validate", code).Return(nil)
		repo.On("Create", ctx, mock.Anything).Return(nil)
		repo.On("CreateVersion", ctx, mock.Anything).Return(nil)

		script, err := uc.CreateScript(ctx, "test_script", "Test script", code, createdBy)

		require.NoError(t, err)
		assert.NotNil(t, script)
		assert.Equal(t, "test_script", script.Name)
		assert.Equal(t, "Test script", script.Description)
		assert.Equal(t, code, script.Code)
		assert.Equal(t, ScriptStatusDraft, script.Status)
	})

	t.Run("validation error", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		code := "return 2 +"
		engine.On("Validate", code).Return(errors.New("syntax error"))

		script, err := uc.CreateScript(ctx, "bad_script", "", code, createdBy)

		assert.Error(t, err)
		assert.Nil(t, script)
		assert.ErrorIs(t, err, ErrScriptSyntaxInvalid)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		code := "return 2 + 2"
		engine.On("Validate", code).Return(nil)
		repo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

		script, err := uc.CreateScript(ctx, "test_script", "", code, createdBy)

		assert.Error(t, err)
		assert.Nil(t, script)
		assert.ErrorIs(t, err, ErrScriptCreateFailed)
	})
}

// ============================================================================
// UpdateScript Tests
// ============================================================================

func TestUseCase_UpdateScript(t *testing.T) {
	ctx := context.Background()
	scriptID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		existingScript := &Script{
			Name:    "test_script",
			Code:    "return 2 + 2",
			Version: 1,
		}
		existingScript.ID = uuidv7.New()

		newCode := "return 3 + 3"
		engine.On("Validate", newCode).Return(nil)
		repo.On("GetByID", ctx, scriptID).Return(existingScript, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)
		repo.On("CreateVersion", ctx, mock.Anything).Return(nil)

		err := uc.UpdateScript(ctx, scriptID, newCode)

		assert.NoError(t, err)
		assert.Equal(t, 2, existingScript.Version) // Version incremented
		repo.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		newCode := "return 3 +"
		engine.On("Validate", newCode).Return(errors.New("syntax error"))

		err := uc.UpdateScript(ctx, scriptID, newCode)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrScriptSyntaxInvalid)
	})

	t.Run("script not found", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		newCode := "return 3 + 3"
		engine.On("Validate", newCode).Return(nil)
		repo.On("GetByID", ctx, scriptID).Return(nil, errors.New("not found"))

		err := uc.UpdateScript(ctx, scriptID, newCode)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "script not found")
	})
}

// ============================================================================
// Status Transition Tests
// ============================================================================

func TestUseCase_ActivateScript(t *testing.T) {
	ctx := context.Background()
	scriptID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		script := &Script{
			Name:   "test_script",
			Status: ScriptStatusDraft,
		}
		script.ID = uuidv7.New()

		repo.On("GetByID", ctx, scriptID).Return(script, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.ActivateScript(ctx, scriptID)

		assert.NoError(t, err)
		assert.Equal(t, ScriptStatusActive, script.Status)
	})

	t.Run("script not found", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		repo.On("GetByID", ctx, scriptID).Return(nil, errors.New("not found"))

		err := uc.ActivateScript(ctx, scriptID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "script not found")
	})
}

func TestUseCase_DeactivateScript(t *testing.T) {
	ctx := context.Background()
	scriptID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		script := &Script{
			Name:   "test_script",
			Status: ScriptStatusActive,
		}
		script.ID = uuidv7.New()

		repo.On("GetByID", ctx, scriptID).Return(script, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.DeactivateScript(ctx, scriptID)

		assert.NoError(t, err)
		assert.Equal(t, ScriptStatusInactive, script.Status)
	})
}

func TestUseCase_ArchiveScript(t *testing.T) {
	ctx := context.Background()
	scriptID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		script := &Script{
			Name:   "test_script",
			Status: ScriptStatusInactive,
		}
		script.ID = uuidv7.New()

		repo.On("GetByID", ctx, scriptID).Return(script, nil)
		repo.On("Update", ctx, mock.Anything).Return(nil)

		err := uc.ArchiveScript(ctx, scriptID)

		assert.NoError(t, err)
		assert.Equal(t, ScriptStatusArchived, script.Status)
	})
}

// ============================================================================
// List and Get Tests
// ============================================================================

func TestUseCase_GetScript(t *testing.T) {
	ctx := context.Background()
	scriptID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		expectedScript := &Script{Name: "test_script"}
		repo.On("GetByID", ctx, scriptID).Return(expectedScript, nil)

		script, err := uc.GetScript(ctx, scriptID)

		assert.NoError(t, err)
		assert.Equal(t, expectedScript, script)
	})

	t.Run("not found", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		repo.On("GetByID", ctx, scriptID).Return(nil, errors.New("not found"))

		script, err := uc.GetScript(ctx, scriptID)

		assert.Error(t, err)
		assert.Nil(t, script)
	})
}

func TestUseCase_ListScripts(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		expectedScripts := []*Script{
			{Name: "script1"},
			{Name: "script2"},
		}

		repo.On("List", ctx, ScriptStatusActive, 10, 0).Return(expectedScripts, 2, nil)

		scripts, total, err := uc.ListScripts(ctx, ScriptStatusActive, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, scripts, 2)
	})
}

// ============================================================================
// Execution History Tests
// ============================================================================

func TestUseCase_GetExecutionHistory(t *testing.T) {
	ctx := context.Background()
	scriptID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		repo := new(MockRepository)
		engine := new(MockEngine)
		uc := NewUseCase(repo, engine)

		expectedExecutions := []*ScriptExecution{
			{ScriptName: "test_script"},
			{ScriptName: "test_script"},
		}

		repo.On("GetExecutionHistory", ctx, scriptID, 10, 0).Return(expectedExecutions, 2, nil)

		executions, total, err := uc.GetExecutionHistory(ctx, scriptID, 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, executions, 2)
	})
}
