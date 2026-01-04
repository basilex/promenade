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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var orderID interface{}
	if inv.OrderID != nil {
		orderID = inv.OrderID.String()
	}

	var paidDate interface{}
	if inv.PaidDate != nil {
		paidDate = inv.PaidDate
	}

	_, err := r.Exec(ctx, query,
		inv.ID.String(),
		inv.InvoiceNo,
		inv.CustomerID.String(),
		orderID,
		inv.SubtotalAmount.Amount,
		inv.TaxAmount.Amount,
		inv.TotalAmount.Amount,
		inv.Currency,
		inv.IssueDate,
		inv.DueDate,
		paidDate,
		string(inv.Status),
		inv.CreatedAt,
		inv.UpdatedAt,
	)

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
		WHERE id = ? AND deleted_at IS NULL`

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
		WHERE invoice_no = ? AND deleted_at IS NULL`

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
		SET invoice_no = ?, customer_id = ?, order_id = ?,
			subtotal_amount = ?, tax_amount = ?, total_amount = ?, currency = ?,
			issue_date = ?, due_date = ?, paid_date = ?, status = ?,
			updated_at = ?
		WHERE id = ? AND deleted_at IS NULL`

	var orderID interface{}
	if inv.OrderID != nil {
		orderID = inv.OrderID.String()
	}

	var paidDate interface{}
	if inv.PaidDate != nil {
		paidDate = inv.PaidDate
	}

	result, err := r.Exec(ctx, query,
		inv.InvoiceNo,
		inv.CustomerID.String(),
		orderID,
		inv.SubtotalAmount.Amount,
		inv.TaxAmount.Amount,
		inv.TotalAmount.Amount,
		inv.Currency,
		inv.IssueDate,
		inv.DueDate,
		paidDate,
		string(inv.Status),
		inv.UpdatedAt,
		inv.ID.String(),
	)

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
	query := `UPDATE billing_invoices SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`

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
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *invoiceRepository) ListByOrder(ctx context.Context, orderID uuidv7.UUID) ([]*invoice.Invoice, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *invoiceRepository) ListByStatus(ctx context.Context, status invoice.InvoiceStatus, page, pageSize int) ([]*invoice.Invoice, int, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *invoiceRepository) ListOverdue(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *invoiceRepository) List(ctx context.Context, page, pageSize int) ([]*invoice.Invoice, int, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *invoiceRepository) CountByStatus(ctx context.Context, status invoice.InvoiceStatus) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *invoiceRepository) GetTotalRevenue(ctx context.Context, from, to time.Time) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *invoiceRepository) CreateLine(ctx context.Context, line *invoice.InvoiceLine) error {
	query := `
		INSERT INTO billing_invoice_lines (
			id, invoice_id, description, quantity, unit_price, amount,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.Exec(ctx, query,
		line.ID.String(),
		line.InvoiceID.String(),
		line.Description,
		line.Quantity,
		line.UnitPrice.Amount,
		line.Amount.Amount,
		line.CreatedAt,
		line.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create invoice line: %w", err)
	}

	return nil
}

func (r *invoiceRepository) DeleteLine(ctx context.Context, lineID uuidv7.UUID) error {
	query := `DELETE FROM billing_invoice_lines WHERE id = ?`

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
		WHERE invoice_id = ?
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
	query := `SELECT EXISTS(SELECT 1 FROM billing_invoices WHERE invoice_no = ? AND deleted_at IS NULL)`
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
		WHERE invoice_no LIKE ?
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
