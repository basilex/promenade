package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// SalesReportOrderRow represents a sales order read model row.
type SalesReportOrderRow struct {
	OrderID     uuidv7.UUID
	CustomerID  uuidv7.UUID
	Currency    string
	TotalCents  int64
	Status      string
	ManagerID   *uuidv7.UUID
	ConfirmedAt time.Time
	UpdatedAt   time.Time
}

// SalesReportItemRow represents a sales order line item read model row.
type SalesReportItemRow struct {
	ID            uuidv7.UUID
	OrderID       uuidv7.UUID
	ProductID     uuidv7.UUID
	Quantity      int
	UnitPriceCents int64
	TotalCents    int64
	CategoryID    *uuidv7.UUID
	ManagerID     *uuidv7.UUID
	UpdatedAt     time.Time
}

// SalesReportRepository writes to analytics read model tables.
type SalesReportRepository struct {
	*BaseRepository
}

// NewSalesReportRepository creates a new SalesReportRepository.
func NewSalesReportRepository(db *sqlx.DB) *SalesReportRepository {
	return &SalesReportRepository{BaseRepository: NewBaseRepository(db)}
}

// UpsertOrder inserts or updates a sales order read model record.
func (r *SalesReportRepository) UpsertOrder(ctx context.Context, row *SalesReportOrderRow) error {
	query := `
		INSERT INTO analytics_sales_orders (
			order_id,
			customer_id,
			currency_code,
			total_cents,
			status,
			manager_id,
			confirmed_at,
			created_at,
			updated_at
		) VALUES (
			:order_id,
			:customer_id,
			:currency_code,
			:total_cents,
			:status,
			:manager_id,
			:confirmed_at,
			CURRENT_TIMESTAMP,
			:updated_at
		)
		ON CONFLICT (order_id) DO UPDATE SET
			customer_id = EXCLUDED.customer_id,
			currency_code = EXCLUDED.currency_code,
			total_cents = EXCLUDED.total_cents,
			status = EXCLUDED.status,
			manager_id = EXCLUDED.manager_id,
			confirmed_at = EXCLUDED.confirmed_at,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.NamedExec(ctx, query, map[string]interface{}{
		"order_id":      row.OrderID.String(),
		"customer_id":   row.CustomerID.String(),
		"currency_code": row.Currency,
		"total_cents":   row.TotalCents,
		"status":        row.Status,
		"manager_id":    uuidPtrToString(row.ManagerID),
		"confirmed_at":  row.ConfirmedAt,
		"updated_at":    row.UpdatedAt,
	})
	return err
}

// UpdateOrderStatus updates the status for a sales order read model record.
func (r *SalesReportRepository) UpdateOrderStatus(ctx context.Context, orderID uuidv7.UUID, status string, updatedAt time.Time) error {
	query := `
		UPDATE analytics_sales_orders
		SET status = $1, updated_at = $2
		WHERE order_id = $3
	`

	_, err := r.Exec(ctx, query, status, updatedAt, orderID.String())
	return err
}

// ReplaceItems removes existing items for the order and inserts new ones.
func (r *SalesReportRepository) ReplaceItems(ctx context.Context, orderID uuidv7.UUID, items []SalesReportItemRow) error {
	deleteQuery := `DELETE FROM analytics_sales_order_items WHERE order_id = $1`
	if _, err := r.Exec(ctx, deleteQuery, orderID.String()); err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO analytics_sales_order_items (
			id,
			order_id,
			product_id,
			quantity,
			unit_price_cents,
			total_cents,
			category_id,
			manager_id,
			created_at,
			updated_at
		) VALUES (
			:id,
			:order_id,
			:product_id,
			:quantity,
			:unit_price_cents,
			:total_cents,
			:category_id,
			:manager_id,
			CURRENT_TIMESTAMP,
			:updated_at
		)
	`

	for _, item := range items {
		if _, err := r.NamedExec(ctx, insertQuery, map[string]interface{}{
			"id":              item.ID.String(),
			"order_id":        item.OrderID.String(),
			"product_id":      item.ProductID.String(),
			"quantity":        item.Quantity,
			"unit_price_cents": item.UnitPriceCents,
			"total_cents":     item.TotalCents,
			"category_id":     uuidPtrToString(item.CategoryID),
			"manager_id":      uuidPtrToString(item.ManagerID),
			"updated_at":      item.UpdatedAt,
		}); err != nil {
			return err
		}
	}

	return nil
}

func uuidPtrToString(id *uuidv7.UUID) interface{} {
	if id == nil || *id == uuidv7.Nil {
		return nil
	}
	return id.String()
}

// SalesReportFilters defines query filters for sales report
type SalesReportFilters struct {
	StartDate  *time.Time
	EndDate    *time.Time
	ManagerID  *uuidv7.UUID
	CustomerID *uuidv7.UUID
	Status     *string
	Limit      int
	Offset     int
}

// SalesReportOrderSummary represents an order summary with item count
type SalesReportOrderSummary struct {
	OrderID     uuidv7.UUID
	CustomerID  uuidv7.UUID
	ManagerID   *uuidv7.UUID
	TotalCents  int64
	Currency    string
	Status      string
	ItemCount   int
	ConfirmedAt time.Time
	UpdatedAt   time.Time
}

// GetSalesReport retrieves sales report with filters and aggregations
func (r *SalesReportRepository) GetSalesReport(ctx context.Context, filters SalesReportFilters) ([]SalesReportOrderSummary, int, int64, error) {
	query := `
		SELECT 
			o.order_id,
			o.customer_id,
			o.manager_id,
			o.total_cents,
			o.currency_code,
			o.status,
			COUNT(i.id) as item_count,
			o.confirmed_at,
			o.updated_at
		FROM analytics_sales_orders o
		LEFT JOIN analytics_sales_order_items i ON o.order_id = i.order_id
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	// Apply filters
	if filters.StartDate != nil {
		query += ` AND o.confirmed_at >= $` + formatArgPos(argPos)
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		query += ` AND o.confirmed_at <= $` + formatArgPos(argPos)
		args = append(args, *filters.EndDate)
		argPos++
	}

	if filters.ManagerID != nil {
		query += ` AND o.manager_id = $` + formatArgPos(argPos)
		args = append(args, filters.ManagerID.String())
		argPos++
	}

	if filters.CustomerID != nil {
		query += ` AND o.customer_id = $` + formatArgPos(argPos)
		args = append(args, filters.CustomerID.String())
		argPos++
	}

	if filters.Status != nil {
		query += ` AND o.status = $` + formatArgPos(argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	query += `
		GROUP BY o.order_id, o.customer_id, o.manager_id, o.total_cents, 
		         o.currency_code, o.status, o.confirmed_at, o.updated_at
		ORDER BY o.confirmed_at DESC
		LIMIT $` + formatArgPos(argPos) + ` OFFSET $` + formatArgPos(argPos+1)

	args = append(args, filters.Limit, filters.Offset)

	// Execute query
	rows := []struct {
		OrderID     string     `db:"order_id"`
		CustomerID  string     `db:"customer_id"`
		ManagerID   *string    `db:"manager_id"`
		TotalCents  int64      `db:"total_cents"`
		Currency    string     `db:"currency_code"`
		Status      string     `db:"status"`
		ItemCount   int        `db:"item_count"`
		ConfirmedAt time.Time  `db:"confirmed_at"`
		UpdatedAt   time.Time  `db:"updated_at"`
	}{}

	if err := r.Select(ctx, &rows, query, args...); err != nil {
		return nil, 0, 0, err
	}

	// Map to domain objects
	orders := make([]SalesReportOrderSummary, 0, len(rows))
	totalValue := int64(0)

	for _, row := range rows {
		orderID, _ := uuidv7.Parse(row.OrderID)
		customerID, _ := uuidv7.Parse(row.CustomerID)

		var managerID *uuidv7.UUID
		if row.ManagerID != nil {
			mid, _ := uuidv7.Parse(*row.ManagerID)
			managerID = &mid
		}

		orders = append(orders, SalesReportOrderSummary{
			OrderID:     orderID,
			CustomerID:  customerID,
			ManagerID:   managerID,
			TotalCents:  row.TotalCents,
			Currency:    row.Currency,
			Status:      row.Status,
			ItemCount:   row.ItemCount,
			ConfirmedAt: row.ConfirmedAt,
			UpdatedAt:   row.UpdatedAt,
		})

		totalValue += row.TotalCents
	}

	// Get total count (without limit/offset)
	totalCount := len(orders)
	if filters.Limit > 0 && len(orders) == filters.Limit {
		// There might be more, need accurate count
		countQuery := `SELECT COUNT(DISTINCT o.order_id) FROM analytics_sales_orders o WHERE 1=1`
		countArgs := []interface{}{}
		countPos := 1

		if filters.StartDate != nil {
			countQuery += ` AND o.confirmed_at >= $` + formatArgPos(countPos)
			countArgs = append(countArgs, *filters.StartDate)
			countPos++
		}

		if filters.EndDate != nil {
			countQuery += ` AND o.confirmed_at <= $` + formatArgPos(countPos)
			countArgs = append(countArgs, *filters.EndDate)
			countPos++
		}

		if filters.ManagerID != nil {
			countQuery += ` AND o.manager_id = $` + formatArgPos(countPos)
			countArgs = append(countArgs, filters.ManagerID.String())
			countPos++
		}

		if filters.CustomerID != nil {
			countQuery += ` AND o.customer_id = $` + formatArgPos(countPos)
			countArgs = append(countArgs, filters.CustomerID.String())
			countPos++
		}

		if filters.Status != nil {
			countQuery += ` AND o.status = $` + formatArgPos(countPos)
			countArgs = append(countArgs, *filters.Status)
		}

		if err := r.Get(ctx, &totalCount, countQuery, countArgs...); err != nil {
			return orders, len(orders), totalValue, nil // Fallback to current count
		}
	}

	return orders, totalCount, totalValue, nil
}

func formatArgPos(pos int) string {
	return string(rune('0' + pos))
}
