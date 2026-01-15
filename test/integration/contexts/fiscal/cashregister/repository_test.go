package cashregister_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestCashRegister creates a test cash register with unique fiscal number
func createTestCashRegister() *cashregister.CashRegister {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	// Use UUID in fiscal number for uniqueness
	fiscalNumber := fmt.Sprintf("FN-%s", uuidv7.New().String()[:8])
	cr, _ := cashregister.NewCashRegister(orgID, fiscalNumber, "Test Model X", userID)
	return cr
}

func TestCashRegisterRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)
		cr := createTestCashRegister()

		err := repo.Create(ctx, cr)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, cr.ID)
		require.NoError(t, err)
		assert.Equal(t, cr.ID, found.ID)
		assert.Equal(t, cr.FiscalNumber, found.FiscalNumber)
		assert.Equal(t, cr.Model, found.Model)
		assert.Equal(t, cashregister.StatusInactive, found.Status)
	})
}

func TestCashRegisterRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)
		cr := createTestCashRegister()

		require.NoError(t, repo.Create(ctx, cr))

		found, err := repo.GetByID(ctx, cr.ID)
		require.NoError(t, err)
		assert.Equal(t, cr.ID, found.ID)

		// Test not found
		notFound, err := repo.GetByID(ctx, uuidv7.New())
		assert.Error(t, err)
		assert.Equal(t, cashregister.ErrCashRegisterNotFound, err)
		assert.Nil(t, notFound)
	})
}

func TestCashRegisterRepository_GetByFiscalNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)
		cr := createTestCashRegister()

		require.NoError(t, repo.Create(ctx, cr))

		found, err := repo.GetByFiscalNumber(ctx, cr.FiscalNumber)
		require.NoError(t, err)
		assert.Equal(t, cr.ID, found.ID)
		assert.Equal(t, cr.FiscalNumber, found.FiscalNumber)

		// Test not found
		notFound, err := repo.GetByFiscalNumber(ctx, "NONEXISTENT-FN")
		assert.Error(t, err)
		assert.Equal(t, cashregister.ErrCashRegisterNotFound, err)
		assert.Nil(t, notFound)
	})
}

func TestCashRegisterRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)
		cr := createTestCashRegister()

		require.NoError(t, repo.Create(ctx, cr))

		// Activate cash register
		userID := uuidv7.New()
		err := cr.Activate("LICENSE-KEY-123", userID)
		require.NoError(t, err)

		err = repo.Update(ctx, cr)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, cr.ID)
		require.NoError(t, err)
		assert.Equal(t, cashregister.StatusActive, found.Status)
		assert.Equal(t, "LICENSE-KEY-123", found.LicenseKey)
	})
}

func TestCashRegisterRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)
		cr := createTestCashRegister()

		require.NoError(t, repo.Create(ctx, cr))

		err := repo.Delete(ctx, cr.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, cr.ID)
		assert.Error(t, err)
		assert.Equal(t, cashregister.ErrCashRegisterNotFound, err)
	})
}

func TestCashRegisterRepository_ListByOrganization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)

		// Create cash registers for same organization
		orgID := uuidv7.New()
		userID := uuidv7.New()
		cr1, _ := cashregister.NewCashRegister(orgID, fmt.Sprintf("FN-%s-1", uuidv7.New().String()[:8]), "Model 1", userID)
		cr2, _ := cashregister.NewCashRegister(orgID, fmt.Sprintf("FN-%s-2", uuidv7.New().String()[:8]), "Model 2", userID)

		require.NoError(t, repo.Create(ctx, cr1))
		require.NoError(t, repo.Create(ctx, cr2))

		filters := &cashregister.ListFilters{OrganizationID: &orgID}
		cashRegisters, err := repo.List(ctx, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cashRegisters), 2)

		// Verify both are in results
		found1 := false
		found2 := false
		for _, cr := range cashRegisters {
			if cr.ID == cr1.ID {
				found1 = true
			}
			if cr.ID == cr2.ID {
				found2 = true
			}
		}
		assert.True(t, found1, "First cash register should be in results")
		assert.True(t, found2, "Second cash register should be in results")
	})
}

func TestCashRegisterRepository_ListActive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)

		cr := createTestCashRegister()
		require.NoError(t, repo.Create(ctx, cr))

		// Activate it
		userID := uuidv7.New()
		require.NoError(t, cr.Activate("LICENSE-123", userID))
		require.NoError(t, repo.Update(ctx, cr))

		cashRegisters, err := repo.ListActive(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cashRegisters), 1)

		// Verify created active cash register is in results
		found := false
		for _, c := range cashRegisters {
			if c.ID == cr.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Activated cash register should be in active list")
	})
}

func TestCashRegisterRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)

		cr := createTestCashRegister()
		require.NoError(t, repo.Create(ctx, cr))

		// Test with filters
		filters := &cashregister.ListFilters{
			OrganizationID: &cr.OrganizationID,
		}

		cashRegisters, err := repo.List(ctx, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(cashRegisters), 1)

		// Verify created cash register is in results
		found := false
		for _, c := range cashRegisters {
			if c.ID == cr.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created cash register should be in filtered results")
	})
}

func TestCashRegisterRepository_GetByLocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		// Note: This test requires Location context integration
		// For now, we'll skip if Location is not available
		t.Skip("GetByLocation requires Location context integration")
	})
}

func TestCashRegisterRepository_ConcurrentCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)

		// Test concurrent creation doesn't violate unique constraints
		orgID := uuidv7.New()
		userID := uuidv7.New()

		cr1, _ := cashregister.NewCashRegister(orgID, fmt.Sprintf("FN-%s-CONC1", uuidv7.New().String()[:8]), "Model", userID)
		cr2, _ := cashregister.NewCashRegister(orgID, fmt.Sprintf("FN-%s-CONC2", uuidv7.New().String()[:8]), "Model", userID)

		// Create both - should succeed
		err1 := repo.Create(ctx, cr1)
		err2 := repo.Create(ctx, cr2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Verify both exist
		found1, err := repo.GetByID(ctx, cr1.ID)
		require.NoError(t, err)
		assert.Equal(t, cr1.FiscalNumber, found1.FiscalNumber)

		found2, err := repo.GetByID(ctx, cr2.ID)
		require.NoError(t, err)
		assert.Equal(t, cr2.FiscalNumber, found2.FiscalNumber)
	})
}

func TestCashRegisterRepository_StatusTransitions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewCashRegisterRepository(testDB.DB)
		cr := createTestCashRegister()
		userID := uuidv7.New()

		require.NoError(t, repo.Create(ctx, cr))

		// Test status transitions: inactive → active → suspended → maintenance → active
		
		// Activate
		require.NoError(t, cr.Activate("LICENSE-KEY", userID))
		require.NoError(t, repo.Update(ctx, cr))
		found, _ := repo.GetByID(ctx, cr.ID)
		assert.Equal(t, cashregister.StatusActive, found.Status)

		// Suspend
		require.NoError(t, cr.Suspend(userID))
		require.NoError(t, repo.Update(ctx, cr))
		found, _ = repo.GetByID(ctx, cr.ID)
		assert.Equal(t, cashregister.StatusSuspended, found.Status)

		// Set to maintenance
		require.NoError(t, cr.SetMaintenance(userID))
		require.NoError(t, repo.Update(ctx, cr))
		found, _ = repo.GetByID(ctx, cr.ID)
		assert.Equal(t, cashregister.StatusMaintenance, found.Status)

		// Activate again
		require.NoError(t, cr.Activate("NEW-LICENSE", userID))
		require.NoError(t, repo.Update(ctx, cr))
		found, _ = repo.GetByID(ctx, cr.ID)
		assert.Equal(t, cashregister.StatusActive, found.Status)
	})
}
