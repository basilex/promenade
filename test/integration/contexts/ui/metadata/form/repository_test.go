package form_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	formRepo "github.com/basilex/promenade/internal/contexts/ui/metadata/form/adapter/repository/postgres"
	formAggregate "github.com/basilex/promenade/internal/contexts/ui/metadata/form/aggregate"
	"github.com/basilex/promenade/test/integration"
)

func TestFormRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := formRepo.NewFormRepository(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, _ *sqlx.Tx) {
		// Create
		entity := newFormDefinition(t, "customer_form", "customer")
		entity.Description = "Customer form definition"

		err := repo.Create(ctx, entity)
		require.NoError(t, err, "Create should succeed")

		// Read
		found, err := repo.GetByID(ctx, entity.GetID())
		require.NoError(t, err, "GetByID should succeed")
		assert.Equal(t, entity.FormID, found.FormID)
		assert.Equal(t, entity.EntityType, found.EntityType)
		assert.Equal(t, entity.Name, found.Name)
		assert.Equal(t, entity.Description, found.Description)

		// Update
		err = found.UpdateMetadata(
			"Updated Form",
			"Updated description",
			map[string]any{"type": "two_column"},
			[]map[string]any{{"name": "phone", "type": "text"}},
			nil,
			nil,
			nil,
			nil,
			false,
			"customer",
			found.TenantID,
		)
		require.NoError(t, err)

		err = repo.Update(ctx, found)
		require.NoError(t, err, "Update should succeed")

		updated, err := repo.GetByID(ctx, entity.GetID())
		require.NoError(t, err)
		assert.Equal(t, "Updated Form", updated.Name)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, 2, updated.Version)
		assert.False(t, updated.IsActive)

		// Delete
		err = repo.Delete(ctx, entity.GetID())
		require.NoError(t, err, "Delete should succeed")

		deleted, err := repo.GetByID(ctx, entity.GetID())
		assert.Error(t, err, "GetByID should fail after delete")
		assert.Nil(t, deleted)
	})
}

func TestFormRepository_ListAndGetByFormID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDBWithCleanTables(t)
	repo := formRepo.NewFormRepository(testDB.DB)

	testDB.WithTransaction(t, func(ctx context.Context, _ *sqlx.Tx) {
		// Create test data
		customerForm := newFormDefinition(t, "customer_form", "customer")
		orderForm := newFormDefinition(t, "order_form", "order")

		require.NoError(t, repo.Create(ctx, customerForm))
		require.NoError(t, repo.Create(ctx, orderForm))

		// GetByFormID
		found, err := repo.GetByFormID(ctx, "customer_form")
		require.NoError(t, err)
		assert.Equal(t, customerForm.FormID, found.FormID)

		// List by entity type
		forms, total, err := repo.List(ctx, "customer", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, forms, 1)
		assert.Equal(t, "customer", forms[0].EntityType)

		// List all
		all, totalAll, err := repo.ListAll(ctx, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, totalAll)
		assert.Len(t, all, 2)
	})
}

func newFormDefinition(t *testing.T, formID, entityType string) *formAggregate.FormDefinition {
	t.Helper()

	entity, err := formAggregate.NewFormDefinition(
		formID,
		entityType,
		"Test Form",
		map[string]any{"type": "single_column"},
		[]map[string]any{{"name": "email", "type": "text"}},
	)
	require.NoError(t, err)

	// Ensure deterministic values for optional fields
	entity.TenantID = nil
	entity.CreatedBy = nil

	return entity
}
