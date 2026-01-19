package order_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/repository/postgres"
	orderAggregate "github.com/basilex/promenade/internal/contexts/order-mgmt/order/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

func TestOrderRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewOrderRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New() // Sales rep ID (mock)

		_, err := tx.ExecContext(ctx, "INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		o, err := orderAggregate.NewOrder(customerID, "USD")
		require.NoError(t, err)
		require.NoError(t, repo.Create(ctx, o))
		assert.NotEqual(t, uuidv7.UUID{}, o.ID)

		retrieved, err := repo.GetByID(ctx, o.ID)
		require.NoError(t, err)
		assert.Equal(t, o.ID, retrieved.ID)
		assert.Equal(t, customerID, retrieved.CustomerID)

		byNumber, err := repo.GetByOrderNumber(ctx, o.OrderNumber)
		require.NoError(t, err)
		assert.Equal(t, o.ID, byNumber.ID)

		productID := uuidv7.New()
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		require.NoError(t, o.AddLine(productID, 2, unitPrice))
		require.NoError(t, o.Confirm())
		require.NoError(t, repo.Update(ctx, o))

		updated, _ := repo.GetByID(ctx, o.ID)
		assert.Equal(t, orderAggregate.OrderStatusConfirmed, updated.Status)

		require.NoError(t, repo.Delete(ctx, o.ID))
		_, err = repo.GetByID(ctx, o.ID)
		assert.Error(t, err)
	})
}

func TestOrderRepository_ListByCustomerID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewOrderRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New() // Sales rep ID (mock)

		_, err := tx.ExecContext(ctx, "INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			o, _ := orderAggregate.NewOrder(customerID, "USD")
			require.NoError(t, repo.Create(ctx, o))
		}

		// Query orders for customer
		orders, total, err := repo.ListByCustomerID(ctx, customerID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, orders, 3)
		assert.Equal(t, int64(3), total)
	})
}

func TestOrderRepository_ListByStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewOrderRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New() // Sales rep ID (mock)

		_, err := tx.ExecContext(ctx, "INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		o1, _ := orderAggregate.NewOrder(customerID, "USD")
		require.NoError(t, repo.Create(ctx, o1))

		o2, _ := orderAggregate.NewOrder(customerID, "USD")
		productID := uuidv7.New()
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		require.NoError(t, o2.AddLine(productID, 1, unitPrice))
		require.NoError(t, o2.Confirm())
		require.NoError(t, repo.Create(ctx, o2))

		pending, total, err := repo.ListByStatus(ctx, orderAggregate.OrderStatusPending, 1, 10)
		require.NoError(t, err)
		assert.Len(t, pending, 1)
		assert.Equal(t, int64(1), total)
	})
}

func TestOrderRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewOrderRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New() // Sales rep ID (mock)

		_, err := tx.ExecContext(ctx, "INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		for i := 0; i < 5; i++ {
			o, _ := orderAggregate.NewOrder(customerID, "USD")
			require.NoError(t, repo.Create(ctx, o))
		}

		// Query orders with pagination
		orders, total, err := repo.List(ctx, 1, 3)
		require.NoError(t, err)
		assert.Len(t, orders, 3)
		assert.Equal(t, int64(5), total)
	})
}

func TestOrderRepository_OrderWithLines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewOrderRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New() // Sales rep ID (mock)

		_, err := tx.ExecContext(ctx, "INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		o, _ := orderAggregate.NewOrder(customerID, "USD")

		productID1 := uuidv7.New()
		unitPrice1, _ := valueobject.NewMoney(10000, "USD")
		require.NoError(t, o.AddLine(productID1, 2, unitPrice1))

		productID2 := uuidv7.New()
		unitPrice2, _ := valueobject.NewMoney(5000, "USD")
		require.NoError(t, o.AddLine(productID2, 3, unitPrice2))

		require.NoError(t, repo.Create(ctx, o))

		retrieved, err := repo.GetByID(ctx, o.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.Lines, 2)
		assert.Equal(t, int64(35000), retrieved.Total.Amount)
	})
}

func TestOrderRepository_OrderLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewOrderRepository(testDB.DB)
		customerID := uuidv7.New()
		assignedTo := uuidv7.New() // Sales rep ID (mock)

		_, err := tx.ExecContext(ctx, "INSERT INTO customer_customers (id, name, email, status, tier, source, assigned_to) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			customerID, "Test Customer", "customer_"+customerID.String()+"@test.com", "customer", "free", "direct", assignedTo)
		require.NoError(t, err)

		o, _ := orderAggregate.NewOrder(customerID, "USD")
		productID := uuidv7.New()
		unitPrice, _ := valueobject.NewMoney(10000, "USD")
		require.NoError(t, o.AddLine(productID, 1, unitPrice))
		require.NoError(t, repo.Create(ctx, o))
		assert.Equal(t, orderAggregate.OrderStatusPending, o.Status)

		require.NoError(t, o.Confirm())
		require.NoError(t, repo.Update(ctx, o))
		retrieved, _ := repo.GetByID(ctx, o.ID)
		assert.Equal(t, orderAggregate.OrderStatusConfirmed, retrieved.Status)

		require.NoError(t, o.StartProcessing())
		require.NoError(t, repo.Update(ctx, o))
		retrieved, _ = repo.GetByID(ctx, o.ID)
		assert.Equal(t, orderAggregate.OrderStatusProcessing, retrieved.Status)

		require.NoError(t, o.MarkFulfilled())
		require.NoError(t, repo.Update(ctx, o))
		retrieved, _ = repo.GetByID(ctx, o.ID)
		assert.Equal(t, orderAggregate.OrderStatusFulfilled, retrieved.Status)
	})
}
