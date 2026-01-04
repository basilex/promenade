package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/billing/invoice"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

type invoiceRepository struct {
	*BaseRepository
}

func NewInvoiceRepository(db *sqlx.DB) invoice.IRepository {
	return &invoiceRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

type invoiceRow struct {
	ID             string         `db:"id"`
	InvoiceNo      string         `db:"invoice_no"`
	CustomerID     string         `db:"customer_id"`
	OrderID        sql.NullString `db:"order_id"`
	SubtotalAmount int64          `db:"subtotal_amount"`
	TaxAmount      int64          `db:"tax_amount"`
	TotalAmount    int64          `db:"total_amount"`
	Currency       string         `db:"currency"`
	IssueDate      time.Time      `db:"issue_date"`
	DueDate        time.Time      `db:"due_date"`
	PaidDate       sql.NullTime   `db:"paid_date"`
	Status         string         `db:"status"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      sql.NullTime   `db:"deleted_at"`
}

type invoiceLineRow struct {
	ID          string    `db:"id"`
	InvoiceID   string    `db:"invoice_id"`
	Description string    `db:"description"`
	Quantity    int       `db:"quantity"`
	UnitPrice   int64     `db:"unit_price"`
	Amount      int64     `db:"amount"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (r *invoiceRepository) Create(ctx context.Context, inv *invoice.Invoice) error {
	query := `
		INSERT INTO billing_invoices (
			id, invoice_no, customer_id, order_id,
			subtotal_amount, tax_amount, total_amount, currency,
			issue_date, due_date, paid_date, status,
			created_at, updated_at
		) VALUES (
			:id, :invoice_no, :customer_id, :order_id,
			:subtotal_amount, :tax_amount, :total_amount, :currency,
			:issue_date, :due_date, :paid_date, :status,
			:created_at, :updated_at
		)`

	_, err := r.NamedExec(ctx, query, map[string]interface{}{
		"id":               inv.ID.String(),
		"invoice_no":       inv.InvoiceNo,
		"customer_id":      inv.CustomerID.String(),
		"order_id":         inv.OrderID,
		"subtotal_amount":  inv.SubtotalAmount.Amount,
		"tax_amount":       inv.TaxAmount.Amount,
		"total_amount":     inv.TotalAmount.Amount,
		"currency":         inv.Currency,
		"issue_date":       inv.IssueDate,
		"due_date":         inv.DueDate,
		"paid_date":        inv.PaidDate,
		"status":           string(inv.Status),
		"created_at":       inv.CreatedAt,
		"updated_at":       inv.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	for _, line := range inv.Lines {
		if err := r.CreateLine(ctx, &line); err != nil {
			return fmt.Errorf("failed to create invoice line: %w", err)
		}
	}

	return nil
}

func (r *invoiceRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*invoice.Invoice, error) {
	query := `
		SELECT id, invoice_no, customer_id, order_id,
			   subtotal_amount, tax_amount, total_amount, currency,
			   issue_date, due_date, paid_date, status,
			   created_at, updated_at, deleted_at
		FROM billing_invoices
		WHERE id = $1 AND deleted_at IS NULL`

	var row invoiceRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, invoice.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	inv, err := r.rowToEntity(&row)
	if err != nil {
		return nil, err
	}

	lines, err := r.GetLinesByInvoiceID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice lines: %w", err)
	}
	inv.Lines = lines

	return inv, nil
}

func (r *invoiceRepository) GetByInvoiceNo(ctx context.Context, invoiceNo string) (*invoice.Invoice, error) {
	query := `
		SELECT id, invoice_no, customer_id, order_id,
			   subtotal_amount, tax_amount, total_amount, currency,
			   issue_date, due_date, paid_date, status,
			   created_at, updated_at, deleted_at
		FROM billing_invoices
		WHERE invoice_no = $1 AND deleted_at IS NULL`

	var row invoiceRow
	if err := r.Get(ctx, &row, query, invoiceNo); err != nil {
		if err == sql.ErrNoRows {
			return nil, invoice.ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	inv, err := r.rowToEntity(&row)
	if err != nil {
		return nil, err
	}

	id, _ := uuidv7.Parse(row.ID)
	lines, err := r.GetLinesByInvoiceID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load invoice lines: %w", err)
	}
	inv.Lines = lines

	return inv, nil
}

func (r *invoiceRepository) Update(ctx context.Context, inv *invoice.Invoice) error {
	query := `
		UPDATE billing_invoices
		SET invoice_no = :invoice_no, customer_id = :customer_id, order_id = :order_id,
			subtotal_amount = :subtotal_amount, tax_amount = :tax_amount, total_amount = :total_amount, currency = :currency,
			issue_date = :issue_date, due_date = :due_date, paid_date = :paid_date, status = :status,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.NamedExec(ctx, query, map[string]interface{}{
		"invoice_no":       inv.InvoiceNo,
		"customer_id":      inv.CustomerID.String(),
		"order_id":         inv.OrderID,
		"subtotal_amount":  inv.SubtotalAmount.Amount,
		"tax_amount":       inv.TaxAmount.Amount,
		"total_amount":     inv.TotalAmount.Amount,
		"currency":         inv.Currency,
		"issue_date":       inv.IssueDate,
		"due_date":         inv.DueDate,
		"paid_date":        inv.PaidDate,
		"status":           string(inv.Status),
		"updated_at":       inv.UpdatedAt,
		"id":               inv.ID.String(),
	})

	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return invoice.ErrInvoiceNotFound
	}

	return nil
}

func (r *invoiceRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE billing_invoices SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return invoice.ErrInvoiceNotFound
	}

	return nil
}

func (r *invoiceRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*invoice.Invoice, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM billing_invoices
		WHERE customer_id = $1 AND deleted_at IS NULL
	`
	if err := r.Get(ctx, &total, countQuery, customerID); err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Get invoices
	query := `
		SELECT id, invoice_no, customer_id, order_id,
			   due_date, paid_date, status,
			   subtotal_amount, tax_amount, total_amount, currency,
			   created_at, updated_at, deleted_at
		FROM billing_invoices
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []invoiceRow
	if err := r.Select(ctx, &rows, query, customerID, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by customer: %w", err)
	}

	invoices := make([]*invoice.Invoice, 0, len(rows))
	for _, row := range rows {
		inv, err := r.rowToEntity(&row)
		if err != nil {
			return nil, 0, err
		}

		// Load line items
		lines, err := r.GetLinesByInvoiceID(ctx, inv.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		inv.Lines = lines

		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *invoiceRepository) ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*invoice.Invoice, error) {
	query := `
		SELECT id, invoice_no, customer_id, order_id,
			   due_date, paid_date, status,
			   subtotal_amount, tax_amount, total_amount, currency,
			   created_at, updated_at, deleted_at
		FROM billing_invoices
		WHERE order_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	var rows []invoiceRow
	if err := r.Select(ctx, &rows, query, orderID); err != nil {
		return nil, fmt.Errorf("failed to list invoices by order: %w", err)
	}

	invoices := make([]*invoice.Invoice, 0, len(rows))
	for _, row := range rows {
		inv, err := r.rowToEntity(&row)
		if err != nil {
			return nil, err
		}

		// Load line items
		lines, err := r.GetLinesByInvoiceID(ctx, inv.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		inv.Lines = lines

		invoices = append(invoices, inv)
	}

	return invoices, nil
}

func (r *invoiceRepository) ListByStatus(ctx context.Context, status invoice.InvoiceStatus, page, pageSize int) ([]*invoice.Invoice, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM billing_invoices
		WHERE status = $1 AND deleted_at IS NULL
	`
	if err := r.Get(ctx, &total, countQuery, string(status)); err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Get invoices
	query := `
		SELECT id, invoice_no, customer_id, order_id,
			   due_date, paid_date, status,
			   subtotal_amount, tax_amount, total_amount, currency,
			   created_at, updated_at, deleted_at
		FROM billing_invoices
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []invoiceRow
	if err := r.Select(ctx, &rows, query, string(status), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by status: %w", err)
	}

	invoices := make([]*invoice.Invoice, 0, len(rows))
	for _, row := range rows {
		inv, err := r.rowToEntity(&row)
		if err != nil {
			return nil, 0, err
		}

		// Load line items
		lines, err := r.GetLinesByInvoiceID(ctx, inv.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		inv.Lines = lines

		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *invoiceRepository) ListOverdue(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM billing_invoices
		WHERE status = $1
		  AND due_date < NOW()
		  AND deleted_at IS NULL
	`
	if err := r.Get(ctx, &total, countQuery, string(invoice.InvoiceStatusSent)); err != nil {
		return nil, 0, fmt.Errorf("failed to count overdue invoices: %w", err)
	}

	// Get invoices
	query := `
		SELECT id, invoice_number, customer_id, order_id,
			   due_date, paid_date, status,
			   subtotal_amount, tax_amount, total_amount, currency,
			   created_at, updated_at, deleted_at
		FROM billing_invoices
		WHERE status = $1
		  AND due_date < NOW()
		  AND deleted_at IS NULL
		ORDER BY due_date ASC
		LIMIT $2 OFFSET $3
	`

	var rows []invoiceRow
	if err := r.Select(ctx, &rows, query, string(invoice.InvoiceStatusSent), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list overdue invoices: %w", err)
	}

	invoices := make([]*invoice.Invoice, 0, len(rows))
	for _, row := range rows {
		inv, err := r.rowToEntity(&row)
		if err != nil {
			return nil, 0, err
		}

		// Load line items
		lines, err := r.GetLinesByInvoiceID(ctx, inv.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		inv.Lines = lines

		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *invoiceRepository) List(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*)
		FROM billing_invoices
		WHERE deleted_at IS NULL
	`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Get invoices
	query := `
		SELECT id, invoice_number, customer_id, order_id,
			   due_date, paid_date, status,
			   subtotal_amount, tax_amount, total_amount, currency,
			   created_at, updated_at, deleted_at
		FROM billing_invoices
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var rows []invoiceRow
	if err := r.Select(ctx, &rows, query, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}

	invoices := make([]*invoice.Invoice, 0, len(rows))
	for _, row := range rows {
		inv, err := r.rowToEntity(&row)
		if err != nil {
			return nil, 0, err
		}

		// Load line items
		lines, err := r.GetLinesByInvoiceID(ctx, inv.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load invoice lines: %w", err)
		}
		inv.Lines = lines

		invoices = append(invoices, inv)
	}

	return invoices, total, nil
}

func (r *invoiceRepository) CountByStatus(ctx context.Context, status invoice.InvoiceStatus) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM billing_invoices
		WHERE status = $1 AND deleted_at IS NULL
	`

	if err := r.Get(ctx, &count, query, string(status)); err != nil {
		return 0, fmt.Errorf("failed to count invoices by status: %w", err)
	}

	return count, nil
}

func (r *invoiceRepository) GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error) {
	var total sql.NullInt64
	query := `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM billing_invoices
		WHERE status = $1
		  AND paid_date BETWEEN $2 AND $3
		  AND deleted_at IS NULL
	`

	if err := r.Get(ctx, &total, query, string(invoice.InvoiceStatusPaid), from, to); err != nil {
		return 0, fmt.Errorf("failed to calculate total revenue: %w", err)
	}

	if !total.Valid {
		return 0, nil
	}

	return total.Int64, nil
}

func (r *invoiceRepository) CreateLine(ctx context.Context, line *invoice.InvoiceLine) error {
	query := `
		INSERT INTO billing_invoice_lines (
			id, invoice_id, description, quantity, unit_price, amount,
			created_at, updated_at
		) VALUES (
			:id, :invoice_id, :description, :quantity, :unit_price, :amount,
			:created_at, :updated_at
		)`

	_, err := r.NamedExec(ctx, query, map[string]interface{}{
		"id":          line.ID.String(),
		"invoice_id":  line.InvoiceID.String(),
		"description": line.Description,
		"quantity":    line.Quantity,
		"unit_price":  line.UnitPrice.Amount,
		"amount":      line.Amount.Amount,
		"created_at":  line.CreatedAt,
		"updated_at":  line.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create invoice line: %w", err)
	}

	return nil
}

func (r *invoiceRepository) DeleteLine(ctx context.Context, lineID uuidv7.UUID) error {
	query := `DELETE FROM billing_invoice_lines WHERE id = $1`

	result, err := r.Exec(ctx, query, lineID.String())
	if err != nil {
		return fmt.Errorf("failed to delete invoice line: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return invoice.ErrInvoiceLineNotFound
	}

	return nil
}

func (r *invoiceRepository) GetLinesByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]invoice.InvoiceLine, error) {
	query := `
		SELECT id, invoice_id, description, quantity, unit_price, amount,
			   created_at, updated_at
		FROM billing_invoice_lines
		WHERE invoice_id = $1
		ORDER BY created_at ASC`

	var rows []invoiceLineRow
	if err := r.Select(ctx, &rows, query, invoiceID.String()); err != nil {
		return nil, fmt.Errorf("failed to get invoice lines: %w", err)
	}

	lines := make([]invoice.InvoiceLine, 0, len(rows))
	for _, row := range rows {
		line, err := r.lineRowToEntity(&row)
		if err != nil {
			return nil, err
		}
		lines = append(lines, *line)
	}

	return lines, nil
}

func (r *invoiceRepository) ExistsByInvoiceNo(ctx context.Context, invoiceNo string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM billing_invoices WHERE invoice_no = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.Get(ctx, &exists, query, invoiceNo)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return exists, err
}

func (r *invoiceRepository) GenerateInvoiceNumber(ctx context.Context) (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("INV-%d-", year)

	query := `
		SELECT invoice_no
		FROM billing_invoices
		WHERE invoice_no LIKE $1
		ORDER BY invoice_no DESC
		LIMIT 1`

	var lastNo string
	err := r.Get(ctx, &lastNo, query, prefix+"%")
	if err == sql.ErrNoRows {
		return fmt.Sprintf("%s%06d", prefix, 1), nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to generate invoice number: %w", err)
	}

	var sequence int
	_, err = fmt.Sscanf(lastNo, prefix+"%d", &sequence)
	if err != nil {
		return "", fmt.Errorf("failed to parse last invoice number: %w", err)
	}

	return fmt.Sprintf("%s%06d", prefix, sequence+1), nil
}

func (r *invoiceRepository) rowToEntity(row *invoiceRow) (*invoice.Invoice, error) {
	id, err := uuidv7.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	customerID, err := uuidv7.Parse(row.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	var orderID *uuidv7.UUID
	if row.OrderID.Valid {
		oid, err := uuidv7.Parse(row.OrderID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid order ID: %w", err)
		}
		orderID = &oid
	}

	var paidDate *time.Time
	if row.PaidDate.Valid {
		paidDate = &row.PaidDate.Time
	}

	subtotal, _ := valueobject.NewMoney(row.SubtotalAmount, row.Currency)
	tax, _ := valueobject.NewMoney(row.TaxAmount, row.Currency)
	total, _ := valueobject.NewMoney(row.TotalAmount, row.Currency)

	return &invoice.Invoice{
		ID:             id,
		InvoiceNo:      row.InvoiceNo,
		CustomerID:     customerID,
		OrderID:        orderID,
		SubtotalAmount: subtotal,
		TaxAmount:      tax,
		TotalAmount:    total,
		Currency:       row.Currency,
		IssueDate:      row.IssueDate,
		DueDate:        row.DueDate,
		PaidDate:       paidDate,
		Status:         invoice.InvoiceStatus(row.Status),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		Lines:          []invoice.InvoiceLine{},
	}, nil
}

func (r *invoiceRepository) lineRowToEntity(row *invoiceLineRow) (*invoice.InvoiceLine, error) {
	id, err := uuidv7.Parse(row.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid line ID: %w", err)
	}

	invoiceID, err := uuidv7.Parse(row.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	unitPrice, _ := valueobject.NewMoney(row.UnitPrice, "USD")
	amount, _ := valueobject.NewMoney(row.Amount, "USD")

	return &invoice.InvoiceLine{
		ID:          id,
		InvoiceID:   invoiceID,
		Description: row.Description,
		Quantity:    row.Quantity,
		UnitPrice:   unitPrice,
		Amount:      amount,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}
