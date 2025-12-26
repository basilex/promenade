package postgres_test

import (
	"context"
	"encoding/json"
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

func setupWorkflowStepTest(t *testing.T) (*integration.TestDB, *postgres.WorkflowStepRepository, *entity.WorkflowInstance) {
	t.Helper()

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewWorkflowStepRepository(testDB.DB)

	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_steps CASCADE")
	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_instances CASCADE")
	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_definitions CASCADE")

	testUserID := createTestUser(t, testDB)
	testDefinition := createTestWorkflowDefinition(t, testDB, testUserID)
	testInstance := createTestWorkflowInstance(t, testDB, testDefinition, testUserID)

	return testDB, repo.(*postgres.WorkflowStepRepository), testInstance
}

func createTestWorkflowInstance(t *testing.T, testDB *integration.TestDB, definition *entity.WorkflowDefinition, userID uuidv7.UUID) *entity.WorkflowInstance {
	t.Helper()

	input := jsonb.Map{"order_id": "12345"}
	instance := entity.NewWorkflowInstance(definition.ID, definition.Version, "start", input, userID)

	instanceRepo := postgres.NewWorkflowInstanceRepository(testDB.DB)
	err := instanceRepo.Create(context.Background(), instance)
	require.NoError(t, err)

	return instance
}

func createTestStep(instance *entity.WorkflowInstance, stepNumber int) *entity.WorkflowStep {
	step := entity.NewWorkflowStep(
		instance.ID,
		instance.DefinitionID,
		stepNumber,
		entity.WorkflowStepTypeActivity,
		"processing",
	)
	step.ActivityName = stringPtr("process_order")
	step.Input = json.RawMessage(`{"item": "test"}`)
	step.Output = json.RawMessage(`{}`)
	step.ErrorDetails = json.RawMessage(`{}`)
	return step
}

func TestWorkflowStepRepository_Create_Success(t *testing.T) {
	testDB, repo, instance := setupWorkflowStepTest(t)
	ctx := context.Background()

	step := createTestStep(instance, 1)
	err := repo.Create(ctx, step)
	require.NoError(t, err, "Create should not return error")

	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM workflows_steps WHERE id = $1", step.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Step should be created in database")
}

func TestWorkflowStepRepository_GetByID_Success(t *testing.T) {
	_, repo, instance := setupWorkflowStepTest(t)
	ctx := context.Background()

	step := createTestStep(instance, 1)
	err := repo.Create(ctx, step)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, step.ID)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, step.ID, result.ID)
	assert.Equal(t, step.InstanceID, result.InstanceID)
	assert.Equal(t, step.StepNumber, result.StepNumber)
	assert.Equal(t, entity.WorkflowStepStatusPending, result.Status)
}

func TestWorkflowStepRepository_Update_Success(t *testing.T) {
	_, repo, instance := setupWorkflowStepTest(t)
	ctx := context.Background()

	step := createTestStep(instance, 1)
	err := repo.Create(ctx, step)
	require.NoError(t, err)

	step.Status = entity.WorkflowStepStatusCompleted
	output := json.RawMessage(`{"result": "success"}`)
	step.Output = output
	now := time.Now()
	step.CompletedAt = &now

	err = repo.Update(ctx, step)
	assert.NoError(t, err)

	updated, err := repo.GetByID(ctx, step.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.WorkflowStepStatusCompleted, updated.Status)
	assert.NotNil(t, updated.CompletedAt)
}

func TestWorkflowStepRepository_ListByInstanceID_Success(t *testing.T) {
	_, repo, instance := setupWorkflowStepTest(t)
	ctx := context.Background()

	step1 := createTestStep(instance, 1)
	step2 := createTestStep(instance, 2)
	step3 := createTestStep(instance, 3)

	err := repo.Create(ctx, step1)
	require.NoError(t, err)
	err = repo.Create(ctx, step2)
	require.NoError(t, err)
	err = repo.Create(ctx, step3)
	require.NoError(t, err)

	steps, err := repo.ListByInstance(ctx, instance.ID, 100, 0)
	assert.NoError(t, err)
	assert.Len(t, steps, 3)

	assert.Equal(t, 1, steps[0].StepNumber)
	assert.Equal(t, 2, steps[1].StepNumber)
	assert.Equal(t, 3, steps[2].StepNumber)
}

func TestWorkflowStepRepository_GetLatestByInstance_Success(t *testing.T) {
	_, repo, instance := setupWorkflowStepTest(t)
	ctx := context.Background()

	step1 := createTestStep(instance, 1)
	step2 := createTestStep(instance, 2)
	step3 := createTestStep(instance, 3)

	err := repo.Create(ctx, step1)
	require.NoError(t, err)
	err = repo.Create(ctx, step2)
	require.NoError(t, err)
	err = repo.Create(ctx, step3)
	require.NoError(t, err)

	lastStep, err := repo.GetLatestByInstance(ctx, instance.ID)
	assert.NoError(t, err)
	require.NotNil(t, lastStep)
	assert.Equal(t, 3, lastStep.StepNumber)
	assert.Equal(t, step3.ID, lastStep.ID)
}

func TestWorkflowStepRepository_GetLatestByInstance_NoSteps(t *testing.T) {
	_, repo, instance := setupWorkflowStepTest(t)
	ctx := context.Background()

	lastStep, err := repo.GetLatestByInstance(ctx, instance.ID)
	assert.NoError(t, err)
	assert.Nil(t, lastStep)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
