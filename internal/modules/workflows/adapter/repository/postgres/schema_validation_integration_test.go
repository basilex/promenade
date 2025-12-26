package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// ============================================================================
// Schema Validation Integration Tests
// ============================================================================
// These tests verify that Schema.Validate() works correctly with real database
// operations and that invalid schemas are rejected properly.

func setupSchemaValidationTest(t *testing.T) (*integration.TestDB, *postgres.WorkflowDefinitionRepository, uuidv7.UUID) {
	t.Helper()

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewWorkflowDefinitionRepository(testDB.DB)

	// Clean workflows tables
	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_definitions CASCADE")

	// Create test user
	userID := uuidv7.New()
	_, err := testDB.DB.Exec(`
		INSERT INTO core_users (id, email, password, name, status)
		VALUES ($1, $2, $3, $4, 'active')
		ON CONFLICT (email) DO NOTHING
	`, userID, "schema-test@promenade.com", "test-hash", "Schema Test User")
	require.NoError(t, err)

	return testDB, repo.(*postgres.WorkflowDefinitionRepository), userID
}

// ============================================================================
// Invalid Schema Tests
// ============================================================================

func TestSchemaValidation_InvalidSchema_NoStates(t *testing.T) {
	testDB, _, userID := setupSchemaValidationTest(t)
	defer testDB.Cleanup()

	// Create workflow with no states
	schema := entity.WorkflowSchema{
		InitialState: "start",
		States:       []entity.WorkflowState{}, // Empty states
		Transitions:  []entity.WorkflowTransition{},
	}

	def := entity.NewWorkflowDefinition(
		"invalid_no_states",
		"Invalid No States",
		"Test invalid schema",
		schema,
		userID,
	)

	// Validate should fail
	err := def.Validate()

	// Schema validation happens in entity layer
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state") // Should mention states in error
}

func TestSchemaValidation_InvalidSchema_MissingInitialState(t *testing.T) {
	testDB, _, userID := setupSchemaValidationTest(t)
	defer testDB.Cleanup()

	// Create workflow with initial state not in states list
	schema := entity.WorkflowSchema{
		InitialState: "nonexistent", // Not in states list
		States: []entity.WorkflowState{
			{Name: "start", Type: "start", IsFinal: false},
			{Name: "completed", Type: "final", IsFinal: true},
		},
		Transitions: []entity.WorkflowTransition{
			{From: "start", To: "completed", Event: "finish"},
		},
	}

	def := entity.NewWorkflowDefinition(
		"invalid_initial_state",
		"Invalid Initial State",
		"Test invalid schema",
		schema,
		userID,
	)

	// Validate should fail
	err := def.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "initial state") // Should mention initial state
}

func TestSchemaValidation_InvalidSchema_InvalidTransition(t *testing.T) {
	testDB, _, userID := setupSchemaValidationTest(t)
	defer testDB.Cleanup()

	// Create workflow with transition referencing nonexistent state
	schema := entity.WorkflowSchema{
		InitialState: "start",
		States: []entity.WorkflowState{
			{Name: "start", Type: "start", IsFinal: false},
			{Name: "completed", Type: "final", IsFinal: true},
		},
		Transitions: []entity.WorkflowTransition{
			{From: "start", To: "nonexistent", Event: "invalid"}, // "nonexistent" not in states
		},
	}

	def := entity.NewWorkflowDefinition(
		"invalid_transition",
		"Invalid Transition",
		"Test invalid schema",
		schema,
		userID,
	)

	// Validate should fail
	err := def.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transition") // Should mention transition
}

func TestSchemaValidation_InvalidSchema_UnreachableState(t *testing.T) {
	testDB, _, userID := setupSchemaValidationTest(t)
	defer testDB.Cleanup()

	// Create workflow with unreachable state
	schema := entity.WorkflowSchema{
		InitialState: "start",
		States: []entity.WorkflowState{
			{Name: "start", Type: "start", IsFinal: false},
			{Name: "processing", Type: "activity", IsFinal: false},
			{Name: "orphan", Type: "activity", IsFinal: false}, // Unreachable
			{Name: "completed", Type: "final", IsFinal: true},
		},
		Transitions: []entity.WorkflowTransition{
			{From: "start", To: "processing", Event: "begin"},
			{From: "processing", To: "completed", Event: "finish"},
			// No transition leads to "orphan" state
		},
	}

	def := entity.NewWorkflowDefinition(
		"unreachable_state",
		"Unreachable State",
		"Test invalid schema",
		schema,
		userID,
	)

	// Validate should fail
	err := def.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unreachable") // Should mention unreachability
}

func TestSchemaValidation_InvalidSchema_Deadlock(t *testing.T) {
	testDB, _, userID := setupSchemaValidationTest(t)
	defer testDB.Cleanup()

	// Create workflow with deadlock (cycle without exit to terminal state)
	schema := entity.WorkflowSchema{
		InitialState: "start",
		States: []entity.WorkflowState{
			{Name: "start", Type: "start", IsFinal: false},
			{Name: "state_a", Type: "activity", IsFinal: false},
			{Name: "state_b", Type: "activity", IsFinal: false},
			{Name: "state_c", Type: "activity", IsFinal: false},
			{Name: "completed", Type: "final", IsFinal: true},
		},
		Transitions: []entity.WorkflowTransition{
			{From: "start", To: "state_a", Event: "begin"},
			{From: "state_a", To: "state_b", Event: "next"},
			{From: "state_b", To: "state_c", Event: "next"},
			{From: "state_c", To: "state_a", Event: "retry"}, // Cycle back
			// No transition from cycle to "completed" - DEADLOCK
		},
	}

	def := entity.NewWorkflowDefinition(
		"deadlock_cycle",
		"Deadlock Cycle",
		"Test invalid schema",
		schema,
		userID,
	)

	// Validate should fail
	err := def.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unreachable") // Deadlock detected as unreachable state
}

// ============================================================================
// Valid Complex Schema Tests
// ============================================================================

func TestSchemaValidation_ComplexWorkflow_Valid(t *testing.T) {
	testDB, repo, userID := setupSchemaValidationTest(t)
	defer testDB.Cleanup()

	ctx := context.Background()

	// Create complex but valid workflow
	schema := entity.WorkflowSchema{
		InitialState: "draft",
		States: []entity.WorkflowState{
			{Name: "draft", Type: "start", IsFinal: false},
			{Name: "review", Type: "activity", IsFinal: false},
			{Name: "revision", Type: "activity", IsFinal: false},
			{Name: "approved", Type: "activity", IsFinal: false},
			{Name: "rejected", Type: "final", IsFinal: true},
			{Name: "published", Type: "final", IsFinal: true},
		},
		Transitions: []entity.WorkflowTransition{
			{From: "draft", To: "review", Event: "submit"},
			{From: "review", To: "approved", Event: "approve"},
			{From: "review", To: "revision", Event: "request_changes"},
			{From: "review", To: "rejected", Event: "reject"},
			{From: "revision", To: "review", Event: "resubmit"},
			{From: "approved", To: "published", Event: "publish"},
		},
	}

	def := entity.NewWorkflowDefinition(
		"complex_valid",
		"Complex Valid Workflow",
		"Test complex valid schema",
		schema,
		userID,
	)

	// Validate should pass
	err := def.Validate()
	require.NoError(t, err)

	// Should create successfully
	err = repo.Create(ctx, def)
	require.NoError(t, err)
	assert.NotEmpty(t, def.ID)

	// Verify it was stored
	retrieved, err := repo.GetByID(ctx, def.ID)
	require.NoError(t, err)
	assert.Equal(t, "complex_valid", retrieved.Name)
	assert.Equal(t, 6, len(retrieved.Definition.Data.States))
	assert.Equal(t, 6, len(retrieved.Definition.Data.Transitions))
}

func TestSchemaValidation_ComplexWorkflow_WithRetryLoop(t *testing.T) {
	testDB, repo, userID := setupSchemaValidationTest(t)
	defer testDB.Cleanup()

	ctx := context.Background()

	// Create workflow with retry loop BUT with exit to terminal state
	schema := entity.WorkflowSchema{
		InitialState: "pending",
		States: []entity.WorkflowState{
			{Name: "pending", Type: "start", IsFinal: false},
			{Name: "processing", Type: "activity", IsFinal: false},
			{Name: "retrying", Type: "activity", IsFinal: false},
			{Name: "failed", Type: "final", IsFinal: true},
			{Name: "completed", Type: "final", IsFinal: true},
		},
		Transitions: []entity.WorkflowTransition{
			{From: "pending", To: "processing", Event: "start"},
			{From: "processing", To: "completed", Event: "success"},
			{From: "processing", To: "retrying", Event: "error"},
			{From: "retrying", To: "processing", Event: "retry"}, // Cycle
			{From: "retrying", To: "failed", Event: "give_up"},   // EXIT from cycle
		},
	}

	def := entity.NewWorkflowDefinition(
		"retry_loop_valid",
		"Retry Loop Valid",
		"Test valid retry loop",
		schema,
		userID,
	)

	// Validate should pass (cycle has exit to terminal state)
	err := def.Validate()
	require.NoError(t, err)

	// Should create successfully
	err = repo.Create(ctx, def)
	require.NoError(t, err)
	assert.NotEmpty(t, def.ID)

	// Verify it was stored
	retrieved, err := repo.GetByID(ctx, def.ID)
	require.NoError(t, err)
	assert.Equal(t, "retry_loop_valid", retrieved.Name)
	assert.Equal(t, 5, len(retrieved.Definition.Data.States))
	assert.Equal(t, 5, len(retrieved.Definition.Data.Transitions))
}
