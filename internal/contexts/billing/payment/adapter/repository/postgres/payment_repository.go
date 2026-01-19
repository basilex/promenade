package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	paymenterrors "github.com/basilex/promenade/internal/contexts/billing/payment"
	"github.com/basilex/promenade/internal/contexts/billing/payment/aggregate"
	"github.com/basilex/promenade/internal/contexts/billing/payment/usecase"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

type paymentRepository struct {
	db *sqlx.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *sqlx.DB) usecase.IPaymentRepository {
	return &paymentRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *paymentRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes a query that returns a single row
func (r *paymentRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes a query that returns multiple rows
func (r *paymentRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query that doesn't return rows
func (r *paymentRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query
func (r *paymentRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

// paymentRow represents a payment database row
type paymentRow struct {
	ID              string         `db:"id"`
	PaymentNo       string         `db:"payment_no"`
	TransactionID   sql.NullString `db:"transaction_id"`
	CustomerID      string         `db:"customer_id"`
	InvoiceID       sql.NullString `db:"invoice_id"`
	Amount          int64          `db:"amount"`
	Currency        string         `db:"currency"`
	Method          string         `db:"method"`
	CardLast4       sql.NullString `db:"card_last4"`
	CardBrand       sql.NullString `db:"card_brand"`
	BankAccount     sql.NullString `db:"bank_account"`
	PaymentProvider sql.NullString `db:"payment_provider"`
	Status          string         `db:"status"`
	FailureReason   sql.NullString `db:"failure_reason"`
	ProcessedAt     time.Time      `db:"processed_at"`
	RefundedAt      sql.NullTime   `db:"refunded_at"`
	RefundedAmount  sql.NullInt64  `db:"refunded_amount"`
	ProcessedBy     sql.NullString `db:"processed_by"`
	Notes           sql.NullString `db:"notes"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	DeletedAt       sql.NullTime   `db:"deleted_at"`
}

// toNullString converts a string to sql.NullString
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// toEntity converts paymentRow to Payment entity
func (r *paymentRow) toEntity() (*aggregate.Payment, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payment ID: %w", err)
	}

	customerID, err := uuidv7.Parse(r.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse customer ID: %w", err)
	}

	var invoiceID *uuidv7.UUID
	if r.InvoiceID.Valid && r.InvoiceID.String != "" {
		parsed, err := uuidv7.Parse(r.InvoiceID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse invoice ID: %w", err)
		}
		invoiceID = &parsed
	}

	amount, err := valueobject.NewMoney(r.Amount, r.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create money: %w", err)
	}

	var refundedAmount *valueobject.Money
	if r.RefundedAmount.Valid {
		refunded, err := valueobject.NewMoney(r.RefundedAmount.Int64, r.Currency)
		if err != nil {
			return nil, fmt.Errorf("failed to create refunded money: %w", err)
		}
		refundedAmount = &refunded
	}

	var refundedAt *time.Time
	if r.RefundedAt.Valid {
		refundedAt = &r.RefundedAt.Time
	}

	var deletedAt *time.Time
	if r.DeletedAt.Valid {
		deletedAt = &r.DeletedAt.Time
	}

	p := &aggregate.Payment{
		PaymentNo:       r.PaymentNo,
		TransactionID:   r.TransactionID.String,
		CustomerID:      customerID,
		InvoiceID:       invoiceID,
		Amount:          amount,
		Method:          aggregate.PaymentMethod(r.Method),
		Status:          aggregate.PaymentStatus(r.Status),
		CardLast4:       r.CardLast4.String,
		CardBrand:       r.CardBrand.String,
		BankAccount:     r.BankAccount.String,
		PaymentProvider: r.PaymentProvider.String,
		FailureReason:   r.FailureReason.String,
		ProcessedAt:     r.ProcessedAt,
		RefundedAt:      refundedAt,
		RefundedAmount:  refundedAmount,
		ProcessedBy:     r.ProcessedBy.String,
		Notes:           r.Notes.String,
	}

	// Set BaseAggregate fields via setters (not direct access)
	p.ID = id
	p.CreatedAt = r.CreatedAt
	p.UpdatedAt = r.UpdatedAt
	if deletedAt != nil {
		p.DeletedAt = deletedAt
	}

	return p, nil
}

// Create creates a new payment
func (r *paymentRepository) Create(ctx context.Context, p *aggregate.Payment) error {
	query := `
		INSERT INTO billing_payments (
			id, payment_no, transaction_id, customer_id, invoice_id,
			amount, currency, method, card_last4, card_brand,
			bank_account, payment_provider, status, failure_reason,
			processed_at, refunded_at, refunded_amount, processed_by,
			notes, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`

	var invoiceID interface{}
	if p.InvoiceID != nil {
		invoiceID = p.InvoiceID.String()
	}

	var refundedAt interface{}
	if p.RefundedAt != nil {
		refundedAt = *p.RefundedAt
	}

	var refundedAmount interface{}
	if p.RefundedAmount != nil {
		refundedAmount = p.RefundedAmount.Amount
	}

	var deletedAt interface{}
	if p.DeletedAt != nil {
		deletedAt = *p.DeletedAt
	}

	_, err := r.Exec(ctx, query,
		p.GetID().String(), p.PaymentNo, toNullString(p.TransactionID), p.CustomerID.String(), invoiceID,
		p.Amount.Amount, p.Amount.Currency, string(p.Method), toNullString(p.CardLast4), toNullString(p.CardBrand),
		toNullString(p.BankAccount), toNullString(p.PaymentProvider), string(p.Status), toNullString(p.FailureReason),
		p.ProcessedAt, refundedAt, refundedAmount, toNullString(p.ProcessedBy),
		toNullString(p.Notes), p.GetCreatedAt(), p.GetUpdatedAt(), deletedAt,
	)

	return err
}

// GetByID retrieves a payment by ID
func (r *paymentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Payment, error) {
	var row paymentRow
	query := `SELECT * FROM billing_payments WHERE id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, paymenterrors.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return row.toEntity()
}

// GetByPaymentNo retrieves a payment by payment number
func (r *paymentRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*aggregate.Payment, error) {
	var row paymentRow
	query := `SELECT * FROM billing_payments WHERE payment_no = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, paymentNo); err != nil {
		if err == sql.ErrNoRows {
			return nil, paymenterrors.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by number: %w", err)
	}

	return row.toEntity()
}

// GetByTransactionID retrieves a payment by transaction ID
func (r *paymentRepository) GetByTransactionID(ctx context.Context, transactionID string) (*aggregate.Payment, error) {
	var row paymentRow
	query := `SELECT * FROM billing_payments WHERE transaction_id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, transactionID); err != nil {
		if err == sql.ErrNoRows {
			return nil, paymenterrors.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("failed to get payment by transaction ID: %w", err)
	}

	return row.toEntity()
}

// Update updates a payment
func (r *paymentRepository) Update(ctx context.Context, p *aggregate.Payment) error {
	query := `
		UPDATE billing_payments SET
			payment_no = $1, transaction_id = $2, customer_id = $3, invoice_id = $4,
			amount = $5, currency = $6, method = $7, card_last4 = $8, card_brand = $9,
			bank_account = $10, payment_provider = $11, status = $12, failure_reason = $13,
			processed_at = $14, refunded_at = $15, refunded_amount = $16, processed_by = $17,
			notes = $18, updated_at = $19
		WHERE id = $20 AND deleted_at IS NULL
	`

	var invoiceID interface{}
	if p.InvoiceID != nil {
		invoiceID = p.InvoiceID.String()
	}

	var refundedAt interface{}
	if p.RefundedAt != nil {
		refundedAt = *p.RefundedAt
	}

	var refundedAmount interface{}
	if p.RefundedAmount != nil {
		refundedAmount = p.RefundedAmount.Amount
	}

	result, err := r.Exec(ctx, query,
		p.PaymentNo, toNullString(p.TransactionID), p.CustomerID.String(), invoiceID,
		p.Amount.Amount, p.Amount.Currency, string(p.Method), toNullString(p.CardLast4), toNullString(p.CardBrand),
		toNullString(p.BankAccount), toNullString(p.PaymentProvider), string(p.Status), toNullString(p.FailureReason),
		p.ProcessedAt, refundedAt, refundedAmount, toNullString(p.ProcessedBy),
		toNullString(p.Notes), time.Now(), p.GetID().String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return paymenterrors.ErrPaymentNotFound
	}

	return nil
}

// Delete soft deletes a payment
func (r *paymentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE billing_payments SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return paymenterrors.ErrPaymentNotFound
	}

	return nil
}

// List retrieves all payments with pagination
func (r *paymentRepository) List(ctx context.Context, page, pageSize int) ([]*aggregate.Payment, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT * FROM billing_payments 
		WHERE deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	var rows []paymentRow
	if err := r.Select(ctx, &rows, query, pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	payments := make([]*aggregate.Payment, len(rows))
	for i, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		payments[i] = p
	}

	return payments, nil
}

// ListByCustomerID retrieves payments for a customer
func (r *paymentRepository) ListByCustomerID(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Payment, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT * FROM billing_payments 
		WHERE customer_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	var rows []paymentRow
	if err := r.Select(ctx, &rows, query, customerID.String(), pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list payments by customer: %w", err)
	}

	payments := make([]*aggregate.Payment, len(rows))
	for i, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		payments[i] = p
	}

	return payments, nil
}

// ListByInvoiceID retrieves payments for an invoice
func (r *paymentRepository) ListByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]*aggregate.Payment, error) {
	query := `
		SELECT * FROM billing_payments 
		WHERE invoice_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC
	`

	var rows []paymentRow
	if err := r.Select(ctx, &rows, query, invoiceID.String()); err != nil {
		return nil, fmt.Errorf("failed to list payments by invoice: %w", err)
	}

	payments := make([]*aggregate.Payment, len(rows))
	for i, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		payments[i] = p
	}

	return payments, nil
}

// ListByStatus retrieves payments by status
func (r *paymentRepository) ListByStatus(ctx context.Context, status aggregate.PaymentStatus, page, pageSize int) ([]*aggregate.Payment, error) {
	offset := (page - 1) * pageSize
	query := `
		SELECT * FROM billing_payments 
		WHERE status = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	var rows []paymentRow
	if err := r.Select(ctx, &rows, query, string(status), pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list payments by status: %w", err)
	}

	payments := make([]*aggregate.Payment, len(rows))
	for i, row := range rows {
		p, err := row.toEntity()
		if err != nil {
			return nil, err
		}
		payments[i] = p
	}

	return payments, nil
}

// GetTotalByCustomer gets total payment amount for a customer
func (r *paymentRepository) GetTotalByCustomer(ctx context.Context, customerID uuidv7.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM billing_payments 
		WHERE customer_id = $1 AND status = 'completed' AND deleted_at IS NULL
	`

	var total int64
	if err := r.Get(ctx, &total, query, customerID.String()); err != nil {
		return 0, fmt.Errorf("failed to get total by customer: %w", err)
	}

	return total, nil
}

// GetTotalByInvoice gets total payment amount for an invoice
func (r *paymentRepository) GetTotalByInvoice(ctx context.Context, invoiceID uuidv7.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM billing_payments 
		WHERE invoice_id = $1 AND status = 'completed' AND deleted_at IS NULL
	`

	var total int64
	if err := r.Get(ctx, &total, query, invoiceID.String()); err != nil {
		return 0, fmt.Errorf("failed to get total by invoice: %w", err)
	}

	return total, nil
}
