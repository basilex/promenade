package postgres_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/test/integration"
)

func setupWorkflowVariableTest(t *testing.T) (*integration.TestDB, *postgres.WorkflowVariableRepository, *entity.WorkflowInstance) {
	t.Helper()

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewWorkflowVariableRepository(testDB.DB)

	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_variables CASCADE")
	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_instances CASCADE")
	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_definitions CASCADE")

	testUserID := createTestUser(t, testDB)
	testDefinition := createTestWorkflowDefinition(t, testDB, testUserID)
	testInstance := createTestWorkflowInstance(t, testDB, testDefinition, testUserID)

	return testDB, repo.(*postgres.WorkflowVariableRepository), testInstance
}

func createTestVariable(instance *entity.WorkflowInstance, name string, value map[string]interface{}) *entity.WorkflowVariable {
	valueJSON, _ := json.Marshal(value)
	variable := entity.NewWorkflowVariable(
		instance.ID,
		name,
		json.RawMessage(valueJSON),
		"object",
		entity.WorkflowVariableScopeGlobal,
	)
	return variable
}

func TestWorkflowVariableRepository_Create_Success(t *testing.T) {
	testDB, repo, instance := setupWorkflowVariableTest(t)
	ctx := context.Background()

	variable := createTestVariable(instance, "order_total", map[string]interface{}{"amount": 100.50, "currency": "USD"})
	err := repo.Create(ctx, variable)
	require.NoError(t, err, "Create should not return error")

	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM workflows_variables WHERE id = $1", variable.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Variable should be created in database")
}

func TestWorkflowVariableRepository_GetByID_Success(t *testing.T) {
	_, repo, instance := setupWorkflowVariableTest(t)
	ctx := context.Background()

	variable := createTestVariable(instance, "customer_id", map[string]interface{}{"id": "12345"})
	err := repo.Create(ctx, variable)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, variable.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, variable.ID, found.ID)
	assert.Equal(t, "customer_id", found.Name)
	assert.Equal(t, "object", found.Type)
}

func TestWorkflowVariableRepository_GetByName_Success(t *testing.T) {
	_, repo, instance := setupWorkflowVariableTest(t)
	ctx := context.Background()

	variable := createTestVariable(instance, "status", map[string]interface{}{"value": "approved"})
	err := repo.Create(ctx, variable)
	require.NoError(t, err)

	found, err := repo.GetByName(ctx, instance.ID, "status")
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "status", found.Name)
	assert.Equal(t, instance.ID, found.InstanceID)
}

func TestWorkflowVariableRepository_Update_Success(t *testing.T) {
	_, repo, instance := setupWorkflowVariableTest(t)
	ctx := context.Background()

	variable := createTestVariable(instance, "counter", map[string]interface{}{"value": 1})
	err := repo.Create(ctx, variable)
	require.NoError(t, err)

	newValueJSON, _ := json.Marshal(map[string]interface{}{"value": 2})
	variable.Update(json.RawMessage(newValueJSON), nil)
	err = repo.Update(ctx, variable)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, variable.ID)
	require.NoError(t, err)
	
	var valueMap map[string]interface{}
	err = json.Unmarshal(found.Value, &valueMap)
	require.NoError(t, err)
	assert.Equal(t, float64(2), valueMap["value"])
}

func TestWorkflowVariableRepository_Delete_Success(t *testing.T) {
	testDB, repo, instance := setupWorkflowVariableTest(t)
	ctx := context.Background()

	variable := createTestVariable(instance, "temp", map[string]interface{}{"data": "test"})
	err := repo.Create(ctx, variable)
	require.NoError(t, err)

	err = repo.Delete(ctx, variable.ID)
	require.NoError(t, err)

	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM workflows_variables WHERE id = $1", variable.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Variable should be deleted")
}

func TestWorkflowVariableRepository_ListByInstance_Success(t *testing.T) {
	_, repo, instance := setupWorkflowVariableTest(t)
	ctx := context.Background()

	var1 := createTestVariable(instance, "var1", map[string]interface{}{"value": 1})
	var2 := createTestVariable(instance, "var2", map[string]interface{}{"value": 2})
	var3 := createTestVariable(instance, "var3", map[string]interface{}{"value": 3})

	require.NoError(t, repo.Create(ctx, var1))
	require.NoError(t, repo.Create(ctx, var2))
	require.NoError(t, repo.Create(ctx, var3))

	variables, err := repo.ListByInstance(ctx, instance.ID)
	require.NoError(t, err)
	assert.Len(t, variables, 3)

	names := []string{variables[0].Name, variables[1].Name, variables[2].Name}
	assert.Contains(t, names, "var1")
	assert.Contains(t, names, "var2")
	assert.Contains(t, names, "var3")
}

func TestWorkflowVariableRepository_ListByScope_Success(t *testing.T) {
	_, repo, instance := setupWorkflowVariableTest(t)
	ctx := context.Background()

	globalVar := createTestVariable(instance, "global_var", map[string]interface{}{"value": "global"})
	globalVar.Scope = entity.WorkflowVariableScopeGlobal
	require.NoError(t, repo.Create(ctx, globalVar))

	stateVar := createTestVariable(instance, "state_var", map[string]interface{}{"value": "state"})
	stateVar.Scope = entity.WorkflowVariableScopeState
	stateName := "processing"
	stateVar.StateName = &stateName
	require.NoError(t, repo.Create(ctx, stateVar))

	globalVars, err := repo.ListByScope(ctx, instance.ID, entity.WorkflowVariableScopeGlobal)
	require.NoError(t, err)
	assert.Len(t, globalVars, 1)
	assert.Equal(t, "global_var", globalVars[0].Name)

	stateVars, err := repo.ListByScope(ctx, instance.ID, entity.WorkflowVariableScopeState)
	require.NoError(t, err)
	assert.Len(t, stateVars, 1)
	assert.Equal(t, "state_var", stateVars[0].Name)
}
