package contract_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	contractRepo "github.com/basilex/promenade/internal/contexts/order-mgmt/contract/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestCustomer inserts a customer directly into the database for testing
func createTestCustomer(ctx context.Context, tx *sqlx.Tx) (uuidv7.UUID, error) {
	customerID := uuidv7.New()
	assignedTo := uuidv7.New()
	_, err := tx.ExecContext(ctx, `
		INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
	return customerID, err
}

// createTestOrder inserts an order directly into the database for testing
func createTestOrder(ctx context.Context, tx *sqlx.Tx, customerID uuidv7.UUID) (uuidv7.UUID, error) {
	orderID := uuidv7.New()
	_, err := tx.ExecContext(ctx, `
		INSERT INTO order_orders (id, order_number, customer_id, total_amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`, orderID, "ORD-2026-000001", customerID, 0.0, "USD", "pending")
	return orderID, err
}

func TestContractRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := contractRepo.NewContractRepository(testDB.DB)

		// Create dependencies
		customerID, err := createTestCustomer(ctx, tx)
		require.NoError(t, err)
		orderID, err := createTestOrder(ctx, tx, customerID)
		require.NoError(t, err)

		// Create test contract
		c := contract.NewContract(orderID, customerID, "Test contract terms")

		err = repo.Create(ctx, c)
		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, c.GetID())
	})
}

func TestContractRepository_GetByID(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	// Create and retrieve
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	c := contract.NewContract(orderID, customerID, "Test contract terms")
	require.NoError(t, repo.Create(ctx, c))

	retrieved, err := repo.GetByID(ctx, c.GetID())
	require.NoError(t, err)
	assert.Equal(t, c.GetID(), retrieved.GetID())
	assert.Equal(t, orderID, retrieved.OrderID)
	assert.Equal(t, customerID, retrieved.CustomerID)
	assert.Equal(t, contract.ContractStatusDraft, retrieved.Status)
	assert.Equal(t, "Test contract terms", retrieved.Terms)
	assert.Equal(t, 1, retrieved.Version)
}

func TestContractRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuidv7.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "contract not found")
}

func TestContractRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	// Create contract
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	c := contract.NewContract(orderID, customerID, "Test contract terms")
	require.NoError(t, repo.Create(ctx, c))

	// Submit for signature
	require.NoError(t, c.SubmitForSignature())
	require.NoError(t, repo.Update(ctx, c))

	// Verify update
	retrieved, err := repo.GetByID(ctx, c.GetID())
	require.NoError(t, err)
	assert.Equal(t, contract.ContractStatusPendingSignature, retrieved.Status)
}

func TestContractRepository_Delete(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	// Create contract
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	c := contract.NewContract(orderID, customerID, "Test contract terms")
	require.NoError(t, repo.Create(ctx, c))

	// Soft delete
	err := repo.Delete(ctx, c.GetID())
	require.NoError(t, err)

	// Verify not found
	_, err = repo.GetByID(ctx, c.GetID())
	require.Error(t, err)
}

func TestContractRepository_List(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	// Create multiple contracts
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	for i := 0; i < 3; i++ {
		c := contract.NewContract(orderID, customerID, "Test contract terms")
		require.NoError(t, repo.Create(ctx, c))
	}

	// List all
	contracts, total, err := repo.List(ctx, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, contracts, 3)
}

func TestContractRepository_List_Pagination(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	// Create 5 contracts
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	for i := 0; i < 5; i++ {
		c := contract.NewContract(orderID, customerID, "Test contract terms")
		require.NoError(t, repo.Create(ctx, c))
	}

	// First page
	contracts, total, err := repo.List(ctx, 0, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, contracts, 2)

	// Second page
	contracts, total, err = repo.List(ctx, 2, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, contracts, 2)
}

func TestContractRepository_GetByOrder(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	orderID1 := uuidv7.New()
	orderID2 := uuidv7.New()
	customerID := uuidv7.New()

	// Create contracts for different orders
	c1 := contract.NewContract(orderID1, customerID, "Contract 1")
	require.NoError(t, repo.Create(ctx, c1))

	c2 := contract.NewContract(orderID1, customerID, "Contract 2")
	require.NoError(t, repo.Create(ctx, c2))

	c3 := contract.NewContract(orderID2, customerID, "Contract 3")
	require.NoError(t, repo.Create(ctx, c3))

	// Get contracts for orderID1
	contracts, err := repo.GetByOrder(ctx, orderID1)
	require.NoError(t, err)
	assert.Len(t, contracts, 2)
}

func TestContractRepository_GetByCustomer(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID1 := uuidv7.New()
	customerID2 := uuidv7.New()

	// Create contracts for different customers
	c1 := contract.NewContract(orderID, customerID1, "Contract 1")
	require.NoError(t, repo.Create(ctx, c1))

	c2 := contract.NewContract(orderID, customerID1, "Contract 2")
	require.NoError(t, repo.Create(ctx, c2))

	c3 := contract.NewContract(orderID, customerID2, "Contract 3")
	require.NoError(t, repo.Create(ctx, c3))

	// Get contracts for customerID1
	contracts, err := repo.GetByCustomer(ctx, customerID1)
	require.NoError(t, err)
	assert.Len(t, contracts, 2)
}

func TestContractRepository_GetActiveContracts(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()

	// Create draft contract
	c1 := contract.NewContract(orderID, customerID, "Contract 1")
	require.NoError(t, repo.Create(ctx, c1))

	// Create active contract
	c2 := contract.NewContract(orderID, customerID, "Contract 2")
	require.NoError(t, c2.SubmitForSignature())
	require.NoError(t, c2.Sign("John Doe", "john@example.com", "sig-123"))
	require.NoError(t, repo.Create(ctx, c2))

	// Get active contracts
	contracts, err := repo.GetActiveContracts(ctx)
	require.NoError(t, err)
	assert.Len(t, contracts, 1)
	assert.Equal(t, contract.ContractStatusActive, contracts[0].Status)
}

func TestContractRepository_ListByStatus(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()

	// Create contracts with different statuses
	c1 := contract.NewContract(orderID, customerID, "Contract 1")
	require.NoError(t, repo.Create(ctx, c1))

	c2 := contract.NewContract(orderID, customerID, "Contract 2")
	require.NoError(t, c2.SubmitForSignature())
	require.NoError(t, repo.Create(ctx, c2))

	c3 := contract.NewContract(orderID, customerID, "Contract 3")
	require.NoError(t, c3.SubmitForSignature())
	require.NoError(t, repo.Create(ctx, c3))

	// List by status
	contracts, total, err := repo.ListByStatus(ctx, contract.ContractStatusPendingSignature, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, contracts, 2)
}

func TestContractRepository_ListByCustomer(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()

	// Create contracts
	for i := 0; i < 3; i++ {
		c := contract.NewContract(orderID, customerID, "Contract terms")
		require.NoError(t, repo.Create(ctx, c))
	}

	// List by customer
	contracts, total, err := repo.ListByCustomer(ctx, customerID, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, contracts, 3)
}

func TestContractRepository_ListExpiringSoon(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := contractRepo.NewContractRepository(db.DB)
	ctx := context.Background()

	orderID := uuidv7.New()
	customerID := uuidv7.New()

	// Create active contract expiring in 5 days
	c1 := contract.NewContract(orderID, customerID, "Contract 1")
	require.NoError(t, c1.SubmitForSignature())
	require.NoError(t, c1.Sign("John Doe", "john@example.com", "sig-123"))
	expiresAt := time.Now().AddDate(0, 0, 5)
	require.NoError(t, c1.SetExpirationDate(expiresAt))
	require.NoError(t, repo.Create(ctx, c1))

	// Create active contract expiring in 20 days (outside range)
	c2 := contract.NewContract(orderID, customerID, "Contract 2")
	require.NoError(t, c2.SubmitForSignature())
	require.NoError(t, c2.Sign("Jane Doe", "jane@example.com", "sig-456"))
	expiresAt2 := time.Now().AddDate(0, 0, 20)
	require.NoError(t, c2.SetExpirationDate(expiresAt2))
	require.NoError(t, repo.Create(ctx, c2))

	// List expiring in 7 days
	contracts, err := repo.ListExpiringSoon(ctx, 7)
	require.NoError(t, err)
	assert.Len(t, contracts, 1)
	assert.Equal(t, c1.GetID(), contracts[0].GetID())
}
