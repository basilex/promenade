package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/order"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// orderRepository implements order.IRepository
type orderRepository struct {
	*BaseRepository
}

// NewOrderRepository creates a new PostgreSQL order repository
func NewOrderRepository(db *sqlx.DB) order.IRepository {
	return &orderRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// orderRow represents a database row for orders table
type orderRow struct {
	ID          uuidv7.UUID    `db:"id"`
	OrderNumber string         `db:"order_number"`
	CustomerID  uuidv7.UUID    `db:"customer_id"`
	CompanyID   *uuidv7.UUID   `db:"company_id"`
	TotalAmount float64        `db:"total_amount"`
	Currency    string         `db:"currency"`
	Status      string         `db:"status"`
	OrderDate   sql.NullTime   `db:"order_date"`
	ConfirmedAt *sql.NullTime  `db:"confirmed_at"`
	FulfilledAt *sql.NullTime  `db:"fulfilled_at"`
	CancelledAt *sql.NullTime  `db:"cancelled_at"`
	ContractID  *uuidv7.UUID   `db:"contract_id"`
	InvoiceID   *uuidv7.UUID   `db:"invoice_id"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	DeletedAt   *sql.NullTime  `db:"deleted_at"`
}

// orderLineRow represents a database row for order_lines table
type orderLineRow struct {
	ID          uuidv7.UUID `db:"id"`
	OrderID     uuidv7.UUID `db:"order_id"`
	ProductID   uuidv7.UUID `db:"product_id"`
	Quantity    int         `db:"quantity"`
	UnitPrice   float64     `db:"unit_price"`
	TotalAmount float64     `db:"total_amount"`
	Currency    string      `db:"currency"`
}

// toEntity converts orderRow to order.Order
func (r *orderRow) toEntity() (*order.Order, error) {
	o := &order.Order{
		ID:          r.ID,
		OrderNumber: r.OrderNumber,
		CustomerID:  r.CustomerID,
		CompanyID:   r.CompanyID,
		Total: valueobject.Money{
			Amount:   int64(r.TotalAmount * 100),
			Currency: r.Currency,
		},
		Currency:   r.Currency,
		Status:     order.OrderStatus(r.Status),
		ContractID: r.ContractID,
		InvoiceID:  r.InvoiceID,
		Lines:      []order.OrderLine{}, // Will be loaded separately
	}

	if r.OrderDate.Valid {
		o.OrderDate = r.OrderDate.Time
	}

	if r.ConfirmedAt != nil && r.ConfirmedAt.Valid {
		t := r.ConfirmedAt.Time
		o.ConfirmedAt = &t
	}

	if r.FulfilledAt != nil && r.FulfilledAt.Valid {
		t := r.FulfilledAt.Time
		o.FulfilledAt = &t
	}

	if r.CancelledAt != nil && r.CancelledAt.Valid {
		t := r.CancelledAt.Time
		o.CancelledAt = &t
	}

	if r.CreatedAt.Valid {
		o.CreatedAt = r.CreatedAt.Time
	}

	if r.UpdatedAt.Valid {
		o.UpdatedAt = r.UpdatedAt.Time
	}

	if r.DeletedAt != nil && r.DeletedAt.Valid {
		t := r.DeletedAt.Time
		o.DeletedAt = &t
	}

	return o, nil
}

// toLineEntity converts orderLineRow to order.OrderLine
func (r *orderLineRow) toLineEntity() order.OrderLine {
	return order.OrderLine{
		ID:        r.ID,
		OrderID:   r.OrderID,
		ProductID: r.ProductID,
		Quantity:  r.Quantity,
		UnitPrice: valueobject.Money{
			Amount:   int64(r.UnitPrice * 100),
			Currency: r.Currency,
		},
		Total: valueobject.Money{
			Amount:   int64(r.TotalAmount * 100),
			Currency: r.Currency,
		},
	}
}

// Create creates a new order
func (r *orderRepository) Create(ctx context.Context, o *order.Order) error {
	query := `
		INSERT INTO order_mgmt_orders (
			id, order_number, customer_id, company_id,
			total_amount, currency, status,
			order_date, confirmed_at, fulfilled_at, cancelled_at,
			contract_id, invoice_id,
			created_at, updated_at
		) VALUES (
			:id, :order_number, :customer_id, :company_id,
			:total_amount, :currency, :status,
			:order_date, :confirmed_at, :fulfilled_at, :cancelled_at,
			:contract_id, :invoice_id,
			:created_at, :updated_at
		)
	`

	_, err := r.NamedExec(ctx, query, map[string]interface{}{
		"id":            o.ID,
		"order_number":  o.OrderNumber,
		"customer_id":   o.CustomerID,
		"company_id":    o.CompanyID,
		"total_amount":  float64(o.Total.Amount) / 100,
		"currency":      o.Currency,
		"status":        string(o.Status),
		"order_date":    o.OrderDate,
		"confirmed_at":  o.ConfirmedAt,
		"fulfilled_at":  o.FulfilledAt,
		"cancelled_at":  o.CancelledAt,
		"contract_id":   o.ContractID,
		"invoice_id":    o.InvoiceID,
		"created_at":    o.CreatedAt,
		"updated_at":    o.UpdatedAt,
	})

	return err
}

// GetByID retrieves an order by its ID
func (r *orderRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*order.Order, error) {
	query := `
		SELECT id, order_number, customer_id, company_id,
			   total_amount, currency, status,
			   order_date, confirmed_at, fulfilled_at, cancelled_at,
			   contract_id, invoice_id,
			   created_at, updated_at, deleted_at
		FROM order_mgmt_orders
		WHERE id = $1 AND deleted_at IS NULL
	`

	var row orderRow
	if err := r.Get(ctx, &row, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, order.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return row.toEntity()
}

// GetByOrderNumber retrieves an order by its order number
func (r *orderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*order.Order, error) {
	query := `
		SELECT id, order_number, customer_id, company_id,
			   total_amount, currency, status,
			   order_date, confirmed_at, fulfilled_at, cancelled_at,
			   contract_id, invoice_id,
			   created_at, updated_at, deleted_at
		FROM order_mgmt_orders
		WHERE order_number = $1 AND deleted_at IS NULL
	`

	var row orderRow
	if err := r.Get(ctx, &row, query, orderNumber); err != nil {
		if err == sql.ErrNoRows {
			return nil, order.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing order
func (r *orderRepository) Update(ctx context.Context, o *order.Order) error {
	query := `
		UPDATE order_mgmt_orders
		SET order_number = :order_number,
			customer_id = :customer_id,
			company_id = :company_id,
			total_amount = :total_amount,
			currency = :currency,
			status = :status,
			confirmed_at = :confirmed_at,
			fulfilled_at = :fulfilled_at,
			cancelled_at = :cancelled_at,
			contract_id = :contract_id,
			invoice_id = :invoice_id,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	_, err := r.NamedExec(ctx, query, map[string]interface{}{
		"id":           o.ID,
		"order_number": o.OrderNumber,
		"customer_id":  o.CustomerID,
		"company_id":   o.CompanyID,
		"total_amount": float64(o.Total.Amount) / 100,
		"currency":     o.Currency,
		"status":       string(o.Status),
		"confirmed_at": o.ConfirmedAt,
		"fulfilled_at": o.FulfilledAt,
		"cancelled_at": o.CancelledAt,
		"contract_id":  o.ContractID,
		"invoice_id":   o.InvoiceID,
		"updated_at":   o.UpdatedAt,
	})

	return err
}

// Delete soft-deletes an order
func (r *orderRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE order_mgmt_orders
		SET deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.Exec(ctx, query, id)
	return err
}

// ListByCustomerID retrieves all orders for a customer
func (r *orderRepository) ListByCustomerID(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*order.Order, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM order_mgmt_orders
		WHERE customer_id = $1 AND deleted_at IS NULL
	`
	if err := r.Get(ctx, &total, countQuery, customerID); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get orders
	query := `
		SELECT id, order_number, customer_id, company_id,
			   total_amount, currency, status,
			   order_date, confirmed_at, fulfilled_at, cancelled_at,
			   contract_id, invoice_id,
			   created_at, updated_at, deleted_at
		FROM order_mgmt_orders
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []orderRow
	if err := r.Select(ctx, &rows, query, customerID, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	orders := make([]*order.Order, 0, len(rows))
	for _, row := range rows {
		o, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

// ListByStatus retrieves all orders with a specific status
func (r *orderRepository) ListByStatus(ctx context.Context, status order.OrderStatus, page, pageSize int) ([]*order.Order, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM order_mgmt_orders
		WHERE status = $1 AND deleted_at IS NULL
	`
	if err := r.Get(ctx, &total, countQuery, string(status)); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get orders
	query := `
		SELECT id, order_number, customer_id, company_id,
			   total_amount, currency, status,
			   order_date, confirmed_at, fulfilled_at, cancelled_at,
			   contract_id, invoice_id,
			   created_at, updated_at, deleted_at
		FROM order_mgmt_orders
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []orderRow
	if err := r.Select(ctx, &rows, query, string(status), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	orders := make([]*order.Order, 0, len(rows))
	for _, row := range rows {
		o, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

// List retrieves all orders with pagination
func (r *orderRepository) List(ctx context.Context, page, pageSize int) ([]*order.Order, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM order_mgmt_orders
		WHERE deleted_at IS NULL
	`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get orders
	query := `
		SELECT id, order_number, customer_id, company_id,
			   total_amount, currency, status,
			   order_date, confirmed_at, fulfilled_at, cancelled_at,
			   contract_id, invoice_id,
			   created_at, updated_at, deleted_at
		FROM order_mgmt_orders
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var rows []orderRow
	if err := r.Select(ctx, &rows, query, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	orders := make([]*order.Order, 0, len(rows))
	for _, row := range rows {
		o, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

// GetLines retrieves all line items for an order
func (r *orderRepository) GetLines(ctx context.Context, orderID uuidv7.UUID) ([]order.OrderLine, error) {
	query := `
		SELECT id, order_id, product_id, quantity,
			   unit_price, total_amount, currency
		FROM order_mgmt_order_lines
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	var rows []orderLineRow
	if err := r.Select(ctx, &rows, query, orderID); err != nil {
		return nil, fmt.Errorf("failed to get order lines: %w", err)
	}

	lines := make([]order.OrderLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, row.toLineEntity())
	}

	return lines, nil
}

// CreateLine creates a new order line
func (r *orderRepository) CreateLine(ctx context.Context, line *order.OrderLine) error {
	query := `
		INSERT INTO order_mgmt_order_lines (
			id, order_id, product_id, quantity,
			unit_price, total_amount, currency
		) VALUES (
			:id, :order_id, :product_id, :quantity,
			:unit_price, :total_amount, :currency
		)
	`

	_, err := r.NamedExec(ctx, query, map[string]interface{}{
		"id":           line.ID,
		"order_id":     line.OrderID,
		"product_id":   line.ProductID,
		"quantity":     line.Quantity,
		"unit_price":   float64(line.UnitPrice.Amount) / 100,
		"total_amount": float64(line.Total.Amount) / 100,
		"currency":     line.UnitPrice.Currency,
	})

	return err
}

// UpdateLine updates an order line
func (r *orderRepository) UpdateLine(ctx context.Context, line *order.OrderLine) error {
	query := `
		UPDATE order_mgmt_order_lines
		SET quantity = :quantity,
			unit_price = :unit_price,
			total_amount = :total_amount
		WHERE id = :id
	`

	_, err := r.NamedExec(ctx, query, map[string]interface{}{
		"id":           line.ID,
		"quantity":     line.Quantity,
		"unit_price":   float64(line.UnitPrice.Amount) / 100,
		"total_amount": float64(line.Total.Amount) / 100,
	})

	return err
}

// DeleteLine deletes an order line
func (r *orderRepository) DeleteLine(ctx context.Context, lineID uuidv7.UUID) error {
	query := `
		DELETE FROM order_mgmt_order_lines
		WHERE id = $1
	`

	_, err := r.Exec(ctx, query, lineID)
	return err
}
