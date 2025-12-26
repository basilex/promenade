package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// ============================================================================
// Test Setup Helpers
// ============================================================================

func setupWorkflowDefinitionTest(t *testing.T) (*integration.TestDB, *postgres.WorkflowDefinitionRepository, uuidv7.UUID) {
	t.Helper()

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := postgres.NewWorkflowDefinitionRepository(testDB.DB)

	// Clean workflows tables specifically
	_, _ = testDB.DB.Exec("TRUNCATE TABLE workflows_definitions CASCADE")

	// Create test user for foreign key
	testUserID := createTestUser(t, testDB)

	return testDB, repo.(*postgres.WorkflowDefinitionRepository), testUserID
}

func createTestUser(t *testing.T, testDB *integration.TestDB) uuidv7.UUID {
	t.Helper()

	userID := uuidv7.New()
	_, err := testDB.DB.Exec(`
		INSERT INTO core_users (id, email, password, name, status)
		VALUES ($1, $2, $3, $4, 'active')
		ON CONFLICT (email) DO NOTHING
	`, userID, "workflow-test@promenade.com", "test-hash", "Workflow Test User")
	require.NoError(t, err)

	return userID
}

func createTestDefinition(name string, status entity.WorkflowDefinitionStatus, userID uuidv7.UUID) *entity.WorkflowDefinition {
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

	def := entity.NewWorkflowDefinition(
		name,
		name+" Display",
		"Test workflow definition",
		schema,
		userID, // Use provided userID
	)
	def.Status = status
	def.Category = "test"
	// Don't set Tags for now - array handling needs special treatment

	return def
}

// ============================================================================
// Create Tests
// ============================================================================

func TestWorkflowDefinitionRepository_Create_Success(t *testing.T) {
	testDB, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("test_workflow", entity.WorkflowDefinitionStatusDraft, userID)

	err := repo.Create(ctx, definition)

	assert.NoError(t, err)

	// Verify creation
	var count int
	err = testDB.DB.Get(&count, "SELECT COUNT(*) FROM workflows_definitions WHERE id = $1", definition.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestWorkflowDefinitionRepository_Create_WithAllFields(t *testing.T) {
	testDB, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("full_workflow", entity.WorkflowDefinitionStatusActive, userID)
	definition.Category = "business_process"
	// Tags commented out for now - array handling needs pq.Array

	err := repo.Create(ctx, definition)
	assert.NoError(t, err)

	// Verify all fields
	var saved entity.WorkflowDefinition
	err = testDB.DB.Get(&saved, `
		SELECT * FROM workflows_definitions 
		WHERE id = $1 AND deleted_at IS NULL
	`, definition.ID)
	
	require.NoError(t, err)
	assert.Equal(t, definition.Name, saved.Name)
	assert.Equal(t, definition.DisplayName, saved.DisplayName)
	assert.Equal(t, definition.Status, saved.Status)
	assert.Equal(t, definition.Category, saved.Category)
	assert.Equal(t, definition.Definition.Data.InitialState, saved.Definition.Data.InitialState)
	assert.Len(t, saved.Definition.Data.States, 3)
	assert.Len(t, saved.Definition.Data.Transitions, 2)
}

// ============================================================================
// GetByID Tests
// ============================================================================

func TestWorkflowDefinitionRepository_GetByID_Success(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("test_get", entity.WorkflowDefinitionStatusActive, userID)
	err := repo.Create(ctx, definition)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, definition.ID)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, definition.ID, result.ID)
	assert.Equal(t, definition.Name, result.Name)
	assert.Equal(t, definition.Status, result.Status)
}

func TestWorkflowDefinitionRepository_GetByID_NotFound(t *testing.T) {
	_, repo, _ := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	nonExistentID := uuidv7.New()
	result, err := repo.GetByID(ctx, nonExistentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestWorkflowDefinitionRepository_GetByID_IgnoresDeleted(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("test_deleted", entity.WorkflowDefinitionStatusDraft, userID)
	err := repo.Create(ctx, definition)
	require.NoError(t, err)

	// Soft delete
	err = repo.Delete(ctx, definition.ID)
	require.NoError(t, err)

	// Try to get deleted definition
	result, err := repo.GetByID(ctx, definition.ID)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// ============================================================================
// GetByName Tests
// ============================================================================

func TestWorkflowDefinitionRepository_GetByName_Success(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	name := "unique_workflow"
	definition := createTestDefinition(name, entity.WorkflowDefinitionStatusActive, userID)
	err := repo.Create(ctx, definition)
	require.NoError(t, err)

	result, err := repo.GetByName(ctx, name)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, name, result.Name)
}

func TestWorkflowDefinitionRepository_GetByName_NotFound(t *testing.T) {
	_, repo, _ := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	result, err := repo.GetByName(ctx, "nonexistent_workflow")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestWorkflowDefinitionRepository_GetByName_ReturnsLatestVersion(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	name := "versioned_workflow"
	
	// Create version 1
	def1 := createTestDefinition(name, entity.WorkflowDefinitionStatusActive, userID)
	def1.Version = 1
	err := repo.Create(ctx, def1)
	require.NoError(t, err)

	// Create version 2 (should be returned)
	def2 := createTestDefinition(name, entity.WorkflowDefinitionStatusActive, userID)
	def2.Version = 2
	err = repo.Create(ctx, def2)
	require.NoError(t, err)

	result, err := repo.GetByName(ctx, name)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Version)
}

// ============================================================================
// Update Tests
// ============================================================================

func TestWorkflowDefinitionRepository_Update_Success(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("test_update", entity.WorkflowDefinitionStatusDraft, userID)
	err := repo.Create(ctx, definition)
	require.NoError(t, err)

	// Update fields
	definition.DisplayName = "Updated Display Name"
	definition.Description = "Updated description"
	definition.Status = entity.WorkflowDefinitionStatusActive
	definition.Category = "updated_category"
	definition.UpdatedAt = time.Now()

	err = repo.Update(ctx, definition)

	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, definition.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Display Name", updated.DisplayName)
	assert.Equal(t, "Updated description", updated.Description)
	assert.Equal(t, entity.WorkflowDefinitionStatusActive, updated.Status)
	assert.Equal(t, "updated_category", updated.Category)
}

func TestWorkflowDefinitionRepository_Update_NotFound(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("nonexistent", entity.WorkflowDefinitionStatusDraft, userID)
	definition.ID = uuidv7.New() // Non-existent ID

	err := repo.Update(ctx, definition)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestWorkflowDefinitionRepository_Update_CannotUpdateDeleted(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("test_update_deleted", entity.WorkflowDefinitionStatusDraft, userID)
	err := repo.Create(ctx, definition)
	require.NoError(t, err)

	// Soft delete
	err = repo.Delete(ctx, definition.ID)
	require.NoError(t, err)

	// Try to update
	definition.DisplayName = "Should Not Update"
	err = repo.Update(ctx, definition)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ============================================================================
// Delete Tests
// ============================================================================

func TestWorkflowDefinitionRepository_Delete_Success(t *testing.T) {
	testDB, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("test_delete", entity.WorkflowDefinitionStatusDraft, userID)
	err := repo.Create(ctx, definition)
	require.NoError(t, err)

	err = repo.Delete(ctx, definition.ID)

	assert.NoError(t, err)

	// Verify soft delete
	var deletedAt *time.Time
	err = testDB.DB.Get(&deletedAt, `
		SELECT deleted_at FROM workflows_definitions WHERE id = $1
	`, definition.ID)
	assert.NoError(t, err)
	assert.NotNil(t, deletedAt)
}

func TestWorkflowDefinitionRepository_Delete_NotFound(t *testing.T) {
	_, repo, _ := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	err := repo.Delete(ctx, uuidv7.New())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestWorkflowDefinitionRepository_Delete_IdempotentDelete(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	definition := createTestDefinition("test_idempotent_delete", entity.WorkflowDefinitionStatusDraft, userID)
	err := repo.Create(ctx, definition)
	require.NoError(t, err)

	// First delete
	err = repo.Delete(ctx, definition.ID)
	assert.NoError(t, err)

	// Second delete should fail (already deleted)
	err = repo.Delete(ctx, definition.ID)
	assert.Error(t, err)
}

// ============================================================================
// ListByStatus Tests
// ============================================================================

func TestWorkflowDefinitionRepository_ListByStatus_Success(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	// Create definitions with different statuses
	active1 := createTestDefinition("active_1", entity.WorkflowDefinitionStatusActive, userID)
	active2 := createTestDefinition("active_2", entity.WorkflowDefinitionStatusActive, userID)
	draft := createTestDefinition("draft_1", entity.WorkflowDefinitionStatusDraft, userID)

	require.NoError(t, repo.Create(ctx, active1))
	require.NoError(t, repo.Create(ctx, active2))
	require.NoError(t, repo.Create(ctx, draft))

	results, err := repo.ListByStatus(ctx, entity.WorkflowDefinitionStatusActive, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, results, 2)
	for _, def := range results {
		assert.Equal(t, entity.WorkflowDefinitionStatusActive, def.Status)
	}
}

func TestWorkflowDefinitionRepository_ListByStatus_WithPagination(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	// Create 5 active definitions
	for i := 0; i < 5; i++ {
		def := createTestDefinition(fmt.Sprintf("active_%d", i), entity.WorkflowDefinitionStatusActive, userID)
		require.NoError(t, repo.Create(ctx, def))
	}

	// Get first page
	page1, err := repo.ListByStatus(ctx, entity.WorkflowDefinitionStatusActive, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, page1, 2)

	// Get second page
	page2, err := repo.ListByStatus(ctx, entity.WorkflowDefinitionStatusActive, 2, 2)
	assert.NoError(t, err)
	assert.Len(t, page2, 2)

	// Verify different results
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestWorkflowDefinitionRepository_ListByStatus_EmptyResult(t *testing.T) {
	_, repo, _ := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	results, err := repo.ListByStatus(ctx, entity.WorkflowDefinitionStatusActive, 10, 0)

	assert.NoError(t, err)
	assert.Empty(t, results)
}

// ============================================================================
// CountByStatus Tests
// ============================================================================

func TestWorkflowDefinitionRepository_CountByStatus_Success(t *testing.T) {
	_, repo, userID := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	// Create definitions
	for i := 0; i < 3; i++ {
		def := createTestDefinition(fmt.Sprintf("active_%d", i), entity.WorkflowDefinitionStatusActive, userID)
		require.NoError(t, repo.Create(ctx, def))
	}
	draft := createTestDefinition("draft_1", entity.WorkflowDefinitionStatusDraft, userID)
	require.NoError(t, repo.Create(ctx, draft))

	count, err := repo.CountByStatus(ctx, entity.WorkflowDefinitionStatusActive)

	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestWorkflowDefinitionRepository_CountByStatus_Zero(t *testing.T) {
	_, repo, _ := setupWorkflowDefinitionTest(t)
	ctx := context.Background()

	count, err := repo.CountByStatus(ctx, entity.WorkflowDefinitionStatusActive)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
