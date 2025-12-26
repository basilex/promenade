package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func setupWorkflowInstanceTest(t *testing.T) (*integration.TestDB, *postgres.WorkflowInstanceRepository, uuidv7.UUID, *entity.WorkflowDefinition) {
	t.Helper()

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewWorkflowInstanceRepository(testDB.DB)

	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_instances CASCADE")
	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_definitions CASCADE")

	testUserID := createTestUser(t, testDB)
	testDefinition := createTestWorkflowDefinition(t, testDB, testUserID)

	return testDB, repo.(*postgres.WorkflowInstanceRepository), testUserID, testDefinition
}

func createTestWorkflowDefinition(t *testing.T, testDB *integration.TestDB, userID uuidv7.UUID) *entity.WorkflowDefinition {
	t.Helper()

	schema := entity.WorkflowSchema{
		InitialState: "start",
		States: []entity.WorkflowState{
			{Name: "start", Type: "start", IsFinal: false},
			{Name: "processing", Type: "activity", IsFinal: false},
			{Name: "completed", Type: "final", IsFinal: true},
		},
		Transitions: []entity.WorkflowTransition{
			{From: "start", To: "processing", Event: "begin"},
			{From: "processing", To: "completed", Event: "finish"},
		},
	}

	def := entity.NewWorkflowDefinition("test_workflow", "Test Workflow", "Test workflow", schema, userID)
	def.Status = entity.WorkflowDefinitionStatusActive

	defRepo := postgres.NewWorkflowDefinitionRepository(testDB.DB)
	err := defRepo.Create(context.Background(), def)
	require.NoError(t, err)

	return def
}

func createTestInstance(definition *entity.WorkflowDefinition, userID uuidv7.UUID) *entity.WorkflowInstance {
	input := jsonb.Map{
		"order_id":  "12345",
		"customer":  "John Doe",
		"amount":    100.50,
		"priority":  "high",
	}
	return entity.NewWorkflowInstance(definition.ID, definition.Version, "start", input, userID)
}

func TestWorkflowInstanceRepository_Create_Success(t *testing.T) {
	testDB, repo, userID, definition := setupWorkflowInstanceTest(t)
	ctx := context.Background()

	instance := createTestInstance(definition, userID)
	err := repo.Create(ctx, instance)
	assert.NoError(t, err)

	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM workflows_instances WHERE id = $1", instance.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestWorkflowInstanceRepository_GetByID_Success(t *testing.T) {
	_, repo, userID, definition := setupWorkflowInstanceTest(t)
	ctx := context.Background()

	instance := createTestInstance(definition, userID)
	err := repo.Create(ctx, instance)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, instance.ID)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, instance.ID, result.ID)
}

func TestWorkflowInstanceRepository_Update_Success(t *testing.T) {
	_, repo, userID, definition := setupWorkflowInstanceTest(t)
	ctx := context.Background()

	instance := createTestInstance(definition, userID)
	err := repo.Create(ctx, instance)
	require.NoError(t, err)

	instance.Status = entity.WorkflowInstanceStatusRunning
	err = repo.Update(ctx, instance)
	assert.NoError(t, err)
	
	updated, err := repo.GetByID(ctx, instance.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.WorkflowInstanceStatusRunning, updated.Status)
}

func TestWorkflowInstanceRepository_Delete_Success(t *testing.T) {
	testDB, repo, userID, definition := setupWorkflowInstanceTest(t)
	ctx := context.Background()

	instance := createTestInstance(definition, userID)
	err := repo.Create(ctx, instance)
	require.NoError(t, err)

	err = repo.Delete(ctx, instance.ID)
	assert.NoError(t, err)

	var deletedAt *time.Time
	err = testDB.DB.Get(&deletedAt, "SELECT deleted_at FROM workflows_instances WHERE id = $1", instance.ID)
	assert.NoError(t, err)
	assert.NotNil(t, deletedAt)
}

