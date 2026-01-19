package script_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/scripting/script/adapter/repository/postgres"
	scriptAggregate "github.com/basilex/promenade/internal/contexts/scripting/script/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestScriptRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewScriptRepository(testDB.DB)

		// Create
		scriptName := fmt.Sprintf("test_script_%s", uuidv7.New().String())
		s, err := scriptAggregate.NewScript(scriptName, "return 2 + 2", scriptAggregate.ScriptTypeCustom)
		require.NoError(t, err)
		s.Description = "Test description"
		require.NoError(t, repo.Create(ctx, s))
		assert.NotEqual(t, uuidv7.UUID{}, s.ID)

		// GetByID
		found, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		assert.Equal(t, scriptName, found.Name)
		assert.Equal(t, "Test description", found.Description)
		assert.Equal(t, "return 2 + 2", found.Code)
		assert.Equal(t, scriptAggregate.ScriptStatusDraft, found.Status)

		// Version snapshot
		version, err := scriptAggregate.NewScriptVersion(found, "initial version", nil)
		require.NoError(t, err)
		require.NoError(t, repo.CreateVersion(ctx, version))

		versions, total, err := repo.ListVersions(ctx, found.ID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, versions, 1)
		assert.Equal(t, found.ID, versions[0].ScriptID)
		assert.Equal(t, found.Version, versions[0].Version)

		// GetByName
		foundByName, err := repo.GetByName(ctx, scriptName)
		require.NoError(t, err)
		assert.Equal(t, s.ID, foundByName.ID)

		// Update
		_ = s.UpdateCode("return 4 + 4")
		_ = s.Activate()
		require.NoError(t, repo.Update(ctx, s))
		updated, _ := repo.GetByID(ctx, s.ID)
		assert.Equal(t, "return 4 + 4", updated.Code)
		assert.Equal(t, scriptAggregate.ScriptStatusActive, updated.Status)

		// Delete
		require.NoError(t, repo.Delete(ctx, s.ID))
		_, err = repo.GetByID(ctx, s.ID)
		assert.Error(t, err)
	})
}

func TestScriptRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewScriptRepository(testDB.DB)

		// Create 5 scripts (3 active, 2 draft)
		uuid := uuidv7.New().String()
		for i := range 5 {
			scriptName := fmt.Sprintf("script_%d_%s", i, uuid)
			s, _ := scriptAggregate.NewScript(scriptName, "return "+string(rune('0'+i)), scriptAggregate.ScriptTypeCustom)
			s.Description = "Test"
			if i < 3 {
				_ = s.Activate()
			}
			require.NoError(t, repo.Create(ctx, s))
		}

		// List active scripts
		activeScripts, total, err := repo.List(ctx, scriptAggregate.ScriptStatusActive, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 3)
		assert.GreaterOrEqual(t, len(activeScripts), 3)

		// List draft scripts
		draftScripts, total, err := repo.List(ctx, scriptAggregate.ScriptStatusDraft, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 2)
		assert.GreaterOrEqual(t, len(draftScripts), 2)

		// ListAll with pagination
		allScripts, total, err := repo.ListAll(ctx, 3, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 5)
		assert.LessOrEqual(t, len(allScripts), 3) // Limited to 3
	})
}

func TestScriptRepository_Metadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewScriptRepository(testDB.DB)

		// Create script with metadata
		scriptName := fmt.Sprintf("meta_script_%s", uuidv7.New().String())
		s, _ := scriptAggregate.NewScript(scriptName, "return 1", scriptAggregate.ScriptTypeCustom)
		s.Description = "Test"
		s.UpdateMetadata("author", "John Doe")
		s.UpdateMetadata("version", "1.0.0")
		require.NoError(t, repo.Create(ctx, s))

		// Retrieve and verify metadata
		found, err := repo.GetByID(ctx, s.ID)
		require.NoError(t, err)
		metadata := found.Metadata.Get()
		assert.Equal(t, "John Doe", metadata["author"])
		assert.Equal(t, "1.0.0", metadata["version"])

		// Update metadata
		found.UpdateMetadata("version", "2.0.0")
		found.DeleteMetadata("author")
		require.NoError(t, repo.Update(ctx, found))

		// Verify changes
		updated, _ := repo.GetByID(ctx, s.ID)
		updatedMetadata := updated.Metadata.Get()
		_, authorExists := updatedMetadata["author"]
		assert.False(t, authorExists, "author key should be deleted")
		assert.Equal(t, "2.0.0", updatedMetadata["version"])
	})
}

func TestScriptRepository_Execution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewScriptRepository(testDB.DB)

		// Create script
		scriptName := fmt.Sprintf("exec_script_%s", uuidv7.New().String())
		s, _ := scriptAggregate.NewScript(scriptName, "return 2 + 2", scriptAggregate.ScriptTypeCustom)
		s.Description = "Test"
		require.NoError(t, repo.Create(ctx, s))

		// Create execution record
		executedBy := uuidv7.New()
		exec := scriptAggregate.NewScriptExecution(s.ID, s.Name, executedBy)
		exec.SetInput(map[string]interface{}{"x": 2, "y": 2})
		exec.SetResult(4, 150)
		require.NoError(t, repo.CreateExecution(ctx, exec))

		// GetExecutionByID
		foundExec, err := repo.GetExecutionByID(ctx, exec.ID)
		require.NoError(t, err)
		assert.Equal(t, s.ID, foundExec.ScriptID)
		assert.Equal(t, s.Name, foundExec.ScriptName)
		assert.Equal(t, 150, foundExec.DurationMs)
		assert.True(t, foundExec.IsSuccess())

		// Create another execution with error
		exec2 := scriptAggregate.NewScriptExecution(s.ID, s.Name, executedBy)
		exec2.SetError(fmt.Errorf("script timeout"), 100)
		require.NoError(t, repo.CreateExecution(ctx, exec2))

		// GetExecutionHistory
		history, total, err := repo.GetExecutionHistory(ctx, s.ID, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 2)
		assert.GreaterOrEqual(t, len(history), 2)

		// GetRecentExecutions
		recent, err := repo.GetRecentExecutions(ctx, 5)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(recent), 2)
	})
}

func TestScriptRepository_Errors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewScriptRepository(testDB.DB)

		// GetByID - not found
		_, err := repo.GetByID(ctx, uuidv7.New())
		assert.Error(t, err)

		// GetByName - not found
		_, err = repo.GetByName(ctx, "nonexistent_script")
		assert.Error(t, err)

		// Update - not found
		s, _ := scriptAggregate.NewScript("test", "return 1", scriptAggregate.ScriptTypeCustom)
		s.Description = "Test"
		err = repo.Update(ctx, s)
		assert.Error(t, err)

		// Delete - not found
		err = repo.Delete(ctx, uuidv7.New())
		assert.Error(t, err)

		// GetExecutionByID - not found
		_, err = repo.GetExecutionByID(ctx, uuidv7.New())
		assert.Error(t, err)
	})
}

func TestScriptRepository_Pagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewScriptRepository(testDB.DB)

		// Create 10 scripts
		uuid := uuidv7.New().String()
		for i := range 10 {
			scriptName := fmt.Sprintf("page_script_%d_%s", i, uuid)
			s, _ := scriptAggregate.NewScript(scriptName, "return "+string(rune('0'+i)), scriptAggregate.ScriptTypeCustom)
			s.Description = "Test"
			_ = s.Activate()
			require.NoError(t, repo.Create(ctx, s))
		}

		// Page 1 (5 items)
		page1, total, err := repo.List(ctx, scriptAggregate.ScriptStatusActive, 5, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 10)
		assert.LessOrEqual(t, len(page1), 5)

		// Page 2 (5 items)
		page2, _, err := repo.List(ctx, scriptAggregate.ScriptStatusActive, 5, 5)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(page2), 5)

		// Ensure no overlap
		if len(page1) > 0 && len(page2) > 0 {
			assert.NotEqual(t, page1[0].ID, page2[0].ID)
		}
	})
}
