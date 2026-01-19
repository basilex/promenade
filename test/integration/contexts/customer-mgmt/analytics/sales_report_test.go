package analytics_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/dto"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/usecase"
	customerRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/adapter/repository/postgres"
	customerAggregate "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer/aggregate"
	orderRepo "github.com/basilex/promenade/internal/contexts/order-mgmt/order/adapter/repository/postgres"
	orderAggregate "github.com/basilex/promenade/internal/contexts/order-mgmt/order/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/basilex/promenade/test/integration"
)

// TestSalesReportUseCase_GetSalesReport_WithRealData tests sales report with actual order data
func TestSalesReportUseCase_GetSalesReport_WithRealData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	// Setup test data
	customer1ID, customer2ID := setupTestData(t, testDB.DB, ctx)

	// Create orders
	order1 := createTestOrder(t, testDB.DB, ctx, customer1ID, 10000, "confirmed") // $100
	order2 := createTestOrder(t, testDB.DB, ctx, customer1ID, 20000, "paid")      // $200
	order3 := createTestOrder(t, testDB.DB, ctx, customer2ID, 15000, "confirmed") // $150
	order4 := createTestOrder(t, testDB.DB, ctx, customer2ID, 5000, "cancelled")  // $50 (should not count)

	// Setup analytics
	analyticsRepo := postgres.NewSalesReportRepository(testDB.DB)
	analyticsUC := usecase.NewSalesReportUseCase(analyticsRepo)

	t.Run("Get all orders without filters", func(t *testing.T) {
		filters := dto.SalesReportRequest{
			Limit: 100,
		}

		report, err := analyticsUC.GetSalesReport(ctx, filters)
		require.NoError(t, err)
		require.NotNil(t, report)

		// Should return all orders (including cancelled)
		assert.GreaterOrEqual(t, report.TotalCount, 4)
		assert.GreaterOrEqual(t, len(report.Orders), 4)

		// Verify order IDs are present
		orderIDs := make(map[string]bool)
		for _, order := range report.Orders {
			orderIDs[order.OrderID] = true
		}
		assert.True(t, orderIDs[order1.String()])
		assert.True(t, orderIDs[order2.String()])
		assert.True(t, orderIDs[order3.String()])
		assert.True(t, orderIDs[order4.String()])
	})

	t.Run("Filter by status - confirmed only", func(t *testing.T) {
		status := "confirmed"
		filters := dto.SalesReportRequest{
			Status: &status,
			Limit:  100,
		}

		report, err := analyticsUC.GetSalesReport(ctx, filters)
		require.NoError(t, err)

		// Should return only confirmed orders (order1 and order3)
		assert.GreaterOrEqual(t, report.TotalCount, 2)
		assert.Equal(t, int64(25000), report.TotalValue) // $100 + $150

		for _, order := range report.Orders {
			assert.Equal(t, "confirmed", order.Status)
		}
	})

	t.Run("Filter by customer", func(t *testing.T) {
		customerIDStr := customer1ID.String()
		filters := dto.SalesReportRequest{
			CustomerID: &customerIDStr,
			Limit:      100,
		}

		report, err := analyticsUC.GetSalesReport(ctx, filters)
		require.NoError(t, err)

		// Should return only customer1 orders (order1 and order2)
		assert.GreaterOrEqual(t, report.TotalCount, 2)
		assert.Equal(t, int64(30000), report.TotalValue) // $100 + $200

		for _, order := range report.Orders {
			assert.Equal(t, customer1ID.String(), order.CustomerID)
		}
	})

	t.Run("Filter by date range", func(t *testing.T) {
		now := time.Now()
		startDate := now.Add(-1 * time.Hour)
		endDate := now.Add(1 * time.Hour)

		filters := dto.SalesReportRequest{
			StartDate: &startDate,
			EndDate:   &endDate,
			Limit:     100,
		}

		report, err := analyticsUC.GetSalesReport(ctx, filters)
		require.NoError(t, err)

		// Should return orders within date range
		assert.GreaterOrEqual(t, report.TotalCount, 4)

		for _, order := range report.Orders {
			assert.True(t, order.ConfirmedAt.After(startDate) || order.ConfirmedAt.Equal(startDate))
			assert.True(t, order.ConfirmedAt.Before(endDate) || order.ConfirmedAt.Equal(endDate))
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		// First page
		filters := dto.SalesReportRequest{
			Limit:  2,
			Offset: 0,
		}

		report1, err := analyticsUC.GetSalesReport(ctx, filters)
		require.NoError(t, err)
		assert.Equal(t, 2, len(report1.Orders))

		// Second page
		filters.Offset = 2
		report2, err := analyticsUC.GetSalesReport(ctx, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(report2.Orders), 2)

		// Verify different orders
		if len(report2.Orders) >= 2 {
			assert.NotEqual(t, report1.Orders[0].OrderID, report2.Orders[0].OrderID)
		}
	})

	t.Run("Combined filters", func(t *testing.T) {
		customerIDStr := customer1ID.String()
		status := "paid"
		filters := dto.SalesReportRequest{
			CustomerID: &customerIDStr,
			Status:     &status,
			Limit:      100,
		}

		report, err := analyticsUC.GetSalesReport(ctx, filters)
		require.NoError(t, err)

		// Should return only paid orders for customer1 (order2)
		assert.GreaterOrEqual(t, report.TotalCount, 1)
		assert.Equal(t, int64(20000), report.TotalValue) // $200

		for _, order := range report.Orders {
			assert.Equal(t, customer1ID.String(), order.CustomerID)
			assert.Equal(t, "paid", order.Status)
		}
	})
}

// TestSalesReportUseCase_GetSalesReport_EmptyDatabase tests behavior with no data
func TestSalesReportUseCase_GetSalesReport_EmptyDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsRepo := postgres.NewSalesReportRepository(testDB.DB)
	analyticsUC := usecase.NewSalesReportUseCase(analyticsRepo)

	filters := dto.SalesReportRequest{
		Limit: 100,
	}

	report, err := analyticsUC.GetSalesReport(ctx, filters)
	require.NoError(t, err)
	require.NotNil(t, report)

	// Should return empty results without error
	assert.Equal(t, 0, report.TotalCount)
	assert.Equal(t, int64(0), report.TotalValue)
	assert.Equal(t, 0, len(report.Orders))
}

// TestSalesReportUseCase_GetSalesReport_InvalidFilters tests validation
func TestSalesReportUseCase_GetSalesReport_InvalidFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testDB := integration.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	analyticsRepo := postgres.NewSalesReportRepository(testDB.DB)
	analyticsUC := usecase.NewSalesReportUseCase(analyticsRepo)

	t.Run("Invalid customer UUID", func(t *testing.T) {
		invalidUUID := "not-a-uuid"
		filters := dto.SalesReportRequest{
			CustomerID: &invalidUUID,
			Limit:      100,
		}

		_, err := analyticsUC.GetSalesReport(ctx, filters)
		assert.Error(t, err)
	})

	t.Run("Invalid manager UUID", func(t *testing.T) {
		invalidUUID := "not-a-uuid"
		filters := dto.SalesReportRequest{
			ManagerID: &invalidUUID,
			Limit:     100,
		}

		_, err := analyticsUC.GetSalesReport(ctx, filters)
		assert.Error(t, err)
	})
}

// setupTestData creates test customers
func setupTestData(t *testing.T, db *sqlx.DB, ctx context.Context) (uuidv7.UUID, uuidv7.UUID) {
	t.Helper()

	// Create manager user (for customer assignment)
	managerID := uuidv7.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO identity_users (id, email, password_hash, status, email_verified)
		VALUES ($1, $2, $3, $4, $5)
	`, managerID, fmt.Sprintf("manager_%s@example.com", uuidv7.New().String()[:8]), "hash", "active", true)
	require.NoError(t, err)

	// Create customers
	custRepo := customerRepo.NewCustomerRepository(db)

	customer1, err := customerAggregate.NewCustomer(
		"Customer 1",
		fmt.Sprintf("customer1_%s@example.com", uuidv7.New().String()[:8]),
		"website",
		managerID,
	)
	require.NoError(t, err)
	err = custRepo.Create(ctx, customer1)
	require.NoError(t, err)

	customer2, err := customerAggregate.NewCustomer(
		"Customer 2",
		fmt.Sprintf("customer2_%s@example.com", uuidv7.New().String()[:8]),
		"referral",
		managerID,
	)
	require.NoError(t, err)
	err = custRepo.Create(ctx, customer2)
	require.NoError(t, err)

	return customer1.ID, customer2.ID
}

// createTestOrder creates a test order with given parameters
func createTestOrder(t *testing.T, db *sqlx.DB, ctx context.Context, customerID uuidv7.UUID, totalCents int64, status string) uuidv7.UUID {
	t.Helper()

	ordRepo := orderRepo.NewOrderRepository(db)

	// Create order
	order, err := orderAggregate.NewOrder(customerID, "USD")
	require.NoError(t, err)

	// Create a product for the line item
	productID := uuidv7.New()
	_, err = db.ExecContext(ctx, `
		INSERT INTO warehouse_products (id, sku, name, unit_price_cents, currency, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, productID, fmt.Sprintf("SKU-%s", productID.String()[:8]), "Test Product", totalCents, "USD", "active")
	require.NoError(t, err)

	// Add line item using correct API (productID, quantity, Money)
	unitPrice := valueobject.Money{Amount: totalCents, Currency: "USD"}
	err = order.AddLine(productID, 1, unitPrice)
	require.NoError(t, err)

	// Set status using correct methods
	switch status {
	case "confirmed":
		err = order.Confirm()
		require.NoError(t, err)
	case "paid":
		err = order.Confirm()
		require.NoError(t, err)
		// "paid" status may require saga/billing context - for now just mark as processing
		err = order.StartProcessing()
		require.NoError(t, err)
	case "fulfilled":
		err = order.Confirm()
		require.NoError(t, err)
		err = order.StartProcessing()
		require.NoError(t, err)
		err = order.MarkFulfilled()
		require.NoError(t, err)
	case "cancelled":
		err = order.Cancel()
		require.NoError(t, err)
	}

	err = ordRepo.Create(ctx, order)
	require.NoError(t, err)

	return order.ID
}
