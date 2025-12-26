package postgres_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/test/integration"
)

func setupWorkflowEventTest(t *testing.T) (*integration.TestDB, *postgres.WorkflowEventRepository, *entity.WorkflowInstance) {
	t.Helper()

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewWorkflowEventRepository(testDB.DB)

	testDB.DB.Exec("TRUNCATE TABLE workflows_events CASCADE")
	testDB.DB.Exec("TRUNCATE TABLE workflows_instances CASCADE")
	testDB.DB.Exec("TRUNCATE TABLE workflows_definitions CASCADE")

	testUserID := createTestUser(t, testDB)
	testDefinition := createTestWorkflowDefinition(t, testDB, testUserID)
	testInstance := createTestWorkflowInstance(t, testDB, testDefinition, testUserID)

	return testDB, repo.(*postgres.WorkflowEventRepository), testInstance
}

func createTestEvent(instance *entity.WorkflowInstance, eventType entity.WorkflowEventType, name string) *entity.WorkflowEvent {
	payload, _ := json.Marshal(map[string]interface{}{"data": "test"})
	event := entity.NewWorkflowEvent(instance.ID, eventType, name, json.RawMessage(payload))
	return event
}

func TestWorkflowEventRepository_Create_Success(t *testing.T) {
	testDB, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	event := createTestEvent(instance, entity.WorkflowEventTypeSignal, "order_paid")
	err := repo.Create(ctx, event)
	require.NoError(t, err, "Create should not return error")

	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM workflows_events WHERE id = $1", event.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Event should be created in database")
}

func TestWorkflowEventRepository_GetByID_Success(t *testing.T) {
	_, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	event := createTestEvent(instance, entity.WorkflowEventTypeWebhook, "payment_received")
	err := repo.Create(ctx, event)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, event.ID)
	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, event.ID, found.ID)
	assert.Equal(t, "payment_received", found.Name)
	assert.Equal(t, entity.WorkflowEventTypeWebhook, found.Type)
	assert.False(t, found.Processed)
}

func TestWorkflowEventRepository_Update_MarkProcessed(t *testing.T) {
	_, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	event := createTestEvent(instance, entity.WorkflowEventTypeManual, "approval_submitted")
	err := repo.Create(ctx, event)
	require.NoError(t, err)

	event.MarkAsProcessed()
	err = repo.Update(ctx, event)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, event.ID)
	require.NoError(t, err)
	assert.True(t, found.Processed)
	assert.NotNil(t, found.ProcessedAt)
}

func TestWorkflowEventRepository_ListByInstance_Success(t *testing.T) {
	_, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	event1 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "event1")
	event2 := createTestEvent(instance, entity.WorkflowEventTypeMessage, "event2")
	event3 := createTestEvent(instance, entity.WorkflowEventTypeTimer, "event3")

	require.NoError(t, repo.Create(ctx, event1))
	require.NoError(t, repo.Create(ctx, event2))
	require.NoError(t, repo.Create(ctx, event3))

	events, err := repo.ListByInstance(ctx, instance.ID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, events, 3)
}

func TestWorkflowEventRepository_ListUnprocessed_Success(t *testing.T) {
	_, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	event1 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "unprocessed1")
	event2 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "unprocessed2")
	event3 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "processed")

	require.NoError(t, repo.Create(ctx, event1))
	require.NoError(t, repo.Create(ctx, event2))
	require.NoError(t, repo.Create(ctx, event3))

	event3.MarkAsProcessed()
	require.NoError(t, repo.Update(ctx, event3))

	unprocessed, err := repo.ListUnprocessed(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, unprocessed, 2)
	
	for _, e := range unprocessed {
		assert.False(t, e.Processed, "All returned events should be unprocessed")
	}
}

func TestWorkflowEventRepository_ListByType_Success(t *testing.T) {
	_, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	webhookEvent1 := createTestEvent(instance, entity.WorkflowEventTypeWebhook, "webhook1")
	webhookEvent2 := createTestEvent(instance, entity.WorkflowEventTypeWebhook, "webhook2")
	signalEvent := createTestEvent(instance, entity.WorkflowEventTypeSignal, "signal1")

	require.NoError(t, repo.Create(ctx, webhookEvent1))
	require.NoError(t, repo.Create(ctx, webhookEvent2))
	require.NoError(t, repo.Create(ctx, signalEvent))

	webhookEvents, err := repo.ListByType(ctx, entity.WorkflowEventTypeWebhook, 10, 0)
	require.NoError(t, err)
	assert.Len(t, webhookEvents, 2)
	
	for _, e := range webhookEvents {
		assert.Equal(t, entity.WorkflowEventTypeWebhook, e.Type)
	}
}

func TestWorkflowEventRepository_CountByInstance_Success(t *testing.T) {
	testDB, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	event1 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "event1")
	event2 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "event2")

	require.NoError(t, repo.Create(ctx, event1))
	require.NoError(t, repo.Create(ctx, event2))

	// Get definition and user from first instance
	var definition entity.WorkflowDefinition
	err := testDB.DB.Get(&definition, "SELECT * FROM workflows_definitions WHERE id = $1", instance.DefinitionID)
	require.NoError(t, err)
	
	// Create another instance for the same definition
	input := jsonb.Map{"test": "other-instance"}
	otherInstance := entity.NewWorkflowInstance(definition.ID, definition.Version, "start", input, definition.CreatedBy)
	instanceRepo := postgres.NewWorkflowInstanceRepository(testDB.DB)
	require.NoError(t, instanceRepo.Create(ctx, otherInstance))
	
	event3 := createTestEvent(otherInstance, entity.WorkflowEventTypeSignal, "event3")
	require.NoError(t, repo.Create(ctx, event3))

	count, err := repo.CountByInstance(ctx, instance.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestWorkflowEventRepository_CountUnprocessed_Success(t *testing.T) {
	_, repo, instance := setupWorkflowEventTest(t)
	ctx := context.Background()

	event1 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "unprocessed1")
	event2 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "unprocessed2")
	event3 := createTestEvent(instance, entity.WorkflowEventTypeSignal, "processed")

	require.NoError(t, repo.Create(ctx, event1))
	require.NoError(t, repo.Create(ctx, event2))
	require.NoError(t, repo.Create(ctx, event3))

	event3.MarkAsProcessed()
	require.NoError(t, repo.Update(ctx, event3))

	count, err := repo.CountUnprocessed(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}
