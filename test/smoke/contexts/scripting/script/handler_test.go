package script_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	scriptHTTP "github.com/basilex/promenade/internal/contexts/scripting/script/adapter/http"
	scriptAggregate "github.com/basilex/promenade/internal/contexts/scripting/script/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockScriptUseCase implements IScriptUseCase with function fields for testing
type MockScriptUseCase struct {
	CreateScriptFunc        func(ctx context.Context, name, description, code string, createdBy uuidv7.UUID) (*scriptAggregate.Script, error)
	GetScriptFunc           func(ctx context.Context, scriptID uuidv7.UUID) (*scriptAggregate.Script, error)
	GetScriptByNameFunc     func(ctx context.Context, name string) (*scriptAggregate.Script, error)
	UpdateScriptFunc        func(ctx context.Context, scriptID uuidv7.UUID, code string) error
	DeleteScriptFunc        func(ctx context.Context, scriptID uuidv7.UUID) error
	ListScriptsFunc         func(ctx context.Context, status scriptAggregate.ScriptStatus, limit, offset int) ([]*scriptAggregate.Script, int, error)
	ListAllScriptsFunc      func(ctx context.Context, limit, offset int) ([]*scriptAggregate.Script, int, error)
	ActivateScriptFunc      func(ctx context.Context, scriptID uuidv7.UUID) error
	DeactivateScriptFunc    func(ctx context.Context, scriptID uuidv7.UUID) error
	ExecuteScriptFunc       func(ctx context.Context, name string, params map[string]interface{}, executedBy uuidv7.UUID) (interface{}, error)
	ValidateScriptFunc      func(ctx context.Context, code string) error
	ArchiveScriptFunc       func(ctx context.Context, scriptID uuidv7.UUID) error
	UpdateMetadataFunc      func(ctx context.Context, scriptID uuidv7.UUID, key string, value string) error
	GetExecutionHistoryFunc func(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*scriptAggregate.ScriptExecution, int, error)
	GetRecentExecutionsFunc func(ctx context.Context, limit int) ([]*scriptAggregate.ScriptExecution, error)
	GetExecutionDetailsFunc func(ctx context.Context, executionID uuidv7.UUID) (*scriptAggregate.ScriptExecution, error)
	ListScriptVersionsFunc  func(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*scriptAggregate.ScriptVersion, int, error)
}

// Implement IScriptUseCase interface methods
func (m *MockScriptUseCase) CreateScript(ctx context.Context, name, description, code string, createdBy uuidv7.UUID) (*scriptAggregate.Script, error) {
	if m.CreateScriptFunc != nil {
		return m.CreateScriptFunc(ctx, name, description, code, createdBy)
	}
	return fakeScript(), nil
}

func (m *MockScriptUseCase) GetScript(ctx context.Context, scriptID uuidv7.UUID) (*scriptAggregate.Script, error) {
	if m.GetScriptFunc != nil {
		return m.GetScriptFunc(ctx, scriptID)
	}
	return fakeScript(), nil
}

func (m *MockScriptUseCase) GetScriptByName(ctx context.Context, name string) (*scriptAggregate.Script, error) {
	if m.GetScriptByNameFunc != nil {
		return m.GetScriptByNameFunc(ctx, name)
	}
	return fakeScript(), nil
}

func (m *MockScriptUseCase) UpdateScript(ctx context.Context, scriptID uuidv7.UUID, code string) error {
	if m.UpdateScriptFunc != nil {
		return m.UpdateScriptFunc(ctx, scriptID, code)
	}
	return nil
}

func (m *MockScriptUseCase) DeleteScript(ctx context.Context, scriptID uuidv7.UUID) error {
	if m.DeleteScriptFunc != nil {
		return m.DeleteScriptFunc(ctx, scriptID)
	}
	return nil
}

func (m *MockScriptUseCase) ListScripts(ctx context.Context, status scriptAggregate.ScriptStatus, limit, offset int) ([]*scriptAggregate.Script, int, error) {
	if m.ListScriptsFunc != nil {
		return m.ListScriptsFunc(ctx, status, limit, offset)
	}
	return []*scriptAggregate.Script{fakeScript()}, 1, nil
}

func (m *MockScriptUseCase) ListAllScripts(ctx context.Context, limit, offset int) ([]*scriptAggregate.Script, int, error) {
	if m.ListAllScriptsFunc != nil {
		return m.ListAllScriptsFunc(ctx, limit, offset)
	}
	return []*scriptAggregate.Script{fakeScript()}, 1, nil
}

func (m *MockScriptUseCase) UpdateScriptMetadata(ctx context.Context, scriptID uuidv7.UUID, key string, value string) error {
	if m.UpdateMetadataFunc != nil {
		return m.UpdateMetadataFunc(ctx, scriptID, key, value)
	}
	return nil
}

func (m *MockScriptUseCase) ExecuteScript(ctx context.Context, name string, params map[string]interface{}, executedBy uuidv7.UUID) (interface{}, error) {
	if m.ExecuteScriptFunc != nil {
		return m.ExecuteScriptFunc(ctx, name, params, executedBy)
	}
	return map[string]interface{}{"result": 4}, nil
}

func (m *MockScriptUseCase) GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*scriptAggregate.ScriptExecution, int, error) {
	if m.GetExecutionHistoryFunc != nil {
		return m.GetExecutionHistoryFunc(ctx, scriptID, limit, offset)
	}
	return []*scriptAggregate.ScriptExecution{}, 0, nil
}

func (m *MockScriptUseCase) GetRecentExecutions(ctx context.Context, limit int) ([]*scriptAggregate.ScriptExecution, error) {
	if m.GetRecentExecutionsFunc != nil {
		return m.GetRecentExecutionsFunc(ctx, limit)
	}
	return []*scriptAggregate.ScriptExecution{}, nil
}

func (m *MockScriptUseCase) GetExecutionDetails(ctx context.Context, executionID uuidv7.UUID) (*scriptAggregate.ScriptExecution, error) {
	if m.GetExecutionDetailsFunc != nil {
		return m.GetExecutionDetailsFunc(ctx, executionID)
	}
	return nil, fmt.Errorf("execution not found")
}

func (m *MockScriptUseCase) ListScriptVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*scriptAggregate.ScriptVersion, int, error) {
	if m.ListScriptVersionsFunc != nil {
		return m.ListScriptVersionsFunc(ctx, scriptID, limit, offset)
	}
	return []*scriptAggregate.ScriptVersion{}, 0, nil
}

func (m *MockScriptUseCase) ActivateScript(ctx context.Context, scriptID uuidv7.UUID) error {
	if m.ActivateScriptFunc != nil {
		return m.ActivateScriptFunc(ctx, scriptID)
	}
	return nil
}

func (m *MockScriptUseCase) DeactivateScript(ctx context.Context, scriptID uuidv7.UUID) error {
	if m.DeactivateScriptFunc != nil {
		return m.DeactivateScriptFunc(ctx, scriptID)
	}
	return nil
}

func (m *MockScriptUseCase) ArchiveScript(ctx context.Context, scriptID uuidv7.UUID) error {
	if m.ArchiveScriptFunc != nil {
		return m.ArchiveScriptFunc(ctx, scriptID)
	}
	return nil
}

func (m *MockScriptUseCase) ValidateScript(ctx context.Context, code string) error {
	if m.ValidateScriptFunc != nil {
		return m.ValidateScriptFunc(ctx, code)
	}
	return nil
}

// Helper functions to create fake data
func fakeScript() *scriptAggregate.Script {
	s, _ := scriptAggregate.NewScript("test-script", "return 2 + 2", scriptAggregate.ScriptTypeCustom)
	return s
}

// Test cases

func TestScriptHandler_CreateScript_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		CreateScriptFunc: func(ctx context.Context, name, description, code string, createdBy uuidv7.UUID) (*scriptAggregate.Script, error) {
			return fakeScript(), nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.POST("/scripts", handler.CreateScript)

	body := map[string]interface{}{
		"name":        "test-script",
		"description": "Test description",
		"code":        "return 2 + 2",
		"script_type": "validation",
		"entity_type": "customer",
	}

	w := smoke.MakeRequest(t, router, "POST", "/scripts", body)
	smoke.AssertSuccessResponse(t, w, 201)
}

func TestScriptHandler_CreateScript_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{}
	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.POST("/scripts", handler.CreateScript)

	body := map[string]interface{}{
		"name": "", // Empty name should fail validation
	}

	w := smoke.MakeRequest(t, router, "POST", "/scripts", body)
	smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
}

func TestScriptHandler_GetScript_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		GetScriptFunc: func(ctx context.Context, scriptID uuidv7.UUID) (*scriptAggregate.Script, error) {
			return fakeScript(), nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.GET("/scripts/:id", handler.GetScript)

	w := smoke.MakeRequest(t, router, "GET", "/scripts/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_GetScriptByName_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		GetScriptByNameFunc: func(ctx context.Context, name string) (*scriptAggregate.Script, error) {
			return fakeScript(), nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.GET("/scripts/name/:name", handler.GetScriptByName)

	w := smoke.MakeRequest(t, router, "GET", "/scripts/name/test-script", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_ExecuteScript_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		ExecuteScriptFunc: func(ctx context.Context, name string, params map[string]interface{}, executedBy uuidv7.UUID) (interface{}, error) {
			return map[string]interface{}{"result": 4}, nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.POST("/scripts/:name/execute", handler.ExecuteScript)

	body := map[string]interface{}{
		"params":      map[string]interface{}{"x": 10},
		"executed_by": smoke.FakeUUID(),
	}

	w := smoke.MakeRequest(t, router, "POST", "/scripts/test-script/execute", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_ValidateScript_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		ValidateScriptFunc: func(ctx context.Context, code string) error {
			return nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.POST("/scripts/validate", handler.ValidateScript)

	body := map[string]interface{}{
		"code": "return 2 + 2",
	}

	w := smoke.MakeRequest(t, router, "POST", "/scripts/validate", body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_ListScripts_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		ListScriptsFunc: func(ctx context.Context, status scriptAggregate.ScriptStatus, limit, offset int) ([]*scriptAggregate.Script, int, error) {
			return []*scriptAggregate.Script{fakeScript()}, 1, nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.GET("/scripts", handler.ListScripts)

	w := smoke.MakeRequest(t, router, "GET", "/scripts?status=active&limit=10&offset=0", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_UpdateScript_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		UpdateScriptFunc: func(ctx context.Context, scriptID uuidv7.UUID, code string) error {
			return nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.PUT("/scripts/:id", handler.UpdateScript)

	body := map[string]interface{}{
		"code":        "return 3 + 3",
		"description": "Updated description",
	}

	w := smoke.MakeRequest(t, router, "PUT", "/scripts/"+smoke.FakeUUID(), body)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_DeleteScript_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		DeleteScriptFunc: func(ctx context.Context, scriptID uuidv7.UUID) error {
			return nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.DELETE("/scripts/:id", handler.DeleteScript)

	w := smoke.MakeRequest(t, router, "DELETE", "/scripts/"+smoke.FakeUUID(), nil)
	// DELETE returns 204 No Content (correct REST standard)
	if w.Code != 204 {
		t.Errorf("Expected status code 204, got %d", w.Code)
	}
}

func TestScriptHandler_ActivateScript_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		ActivateScriptFunc: func(ctx context.Context, scriptID uuidv7.UUID) error {
			return nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.POST("/scripts/:id/activate", handler.ActivateScript)

	w := smoke.MakeRequest(t, router, "POST", "/scripts/"+smoke.FakeUUID()+"/activate", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_GetExecutionHistory_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		GetExecutionHistoryFunc: func(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*scriptAggregate.ScriptExecution, int, error) {
			return []*scriptAggregate.ScriptExecution{}, 0, nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.GET("/scripts/:id/executions", handler.GetExecutionHistory)

	w := smoke.MakeRequest(t, router, "GET", "/scripts/"+smoke.FakeUUID()+"/executions?limit=10&offset=0", nil)
	smoke.AssertSuccessResponse(t, w, 200)
}

func TestScriptHandler_ListScriptVersions_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockScriptUseCase{
		ListScriptVersionsFunc: func(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*scriptAggregate.ScriptVersion, int, error) {
			return []*scriptAggregate.ScriptVersion{fakeScriptVersion()}, 1, nil
		},
	}

	handler := scriptHTTP.NewScriptHandler(mockUC)
	router.GET("/scripts/:id/versions", handler.ListScriptVersions)

	w := smoke.MakeRequest(t, router, "GET", "/scripts/"+smoke.FakeUUID()+"/versions?page=1&page_size=20", nil)

	smoke.AssertSuccessResponse(t, w, 200)
}

func fakeScriptVersion() *scriptAggregate.ScriptVersion {
	return &scriptAggregate.ScriptVersion{
		ID:        uuidv7.New(),
		ScriptID:  uuidv7.New(),
		Version:   1,
		Code:      "return 2 + 2",
		Metadata:  jsonstore.NewField(map[string]string{"source": "test"}),
		ChangeLog: "initial",
		CreatedAt: time.Now(),
	}
}
