package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/repository"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// receiptRepository implements aggregate.IRepository using PostgreSQL
type receiptRepository struct {
	db *sqlx.DB
}

// NewReceiptRepository creates a new PostgreSQL receipt receiptRepository
func NewReceiptRepository(db *sqlx.DB) repository.IReceiptRepository {
	return &receiptRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *receiptRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *receiptRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *receiptRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query without returning rows
func (r *receiptRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query without returning rows
func (r *receiptRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

type receiptRow struct {
	ID                 string                                   `db:"id"`
	Version            int                                      `db:"version"`
	CashRegisterID     string                                   `db:"cash_register_id"`
	OrderID            string                                   `db:"order_id"`
	PaymentType        string                                   `db:"payment_type"`
	ReceiptType        string                                   `db:"receipt_type"`
	Currency           string                                   `db:"currency"`
	TotalAmount        int64                                    `db:"total_amount"`
	TaxAmount          int64                                    `db:"tax_amount"`
	FiscalNumber       sql.NullString                           `db:"fiscal_number"`
	FiscalURL          sql.NullString                           `db:"fiscal_url"`
	QRCode             sql.NullString                           `db:"qr_code"`
	ProviderReceiptID  sql.NullString                           `db:"provider_receipt_id"`
	PrintedAt          sql.NullTime                             `db:"printed_at"`
	CancelledAt        sql.NullTime                             `db:"cancelled_at"`
	CancellationReason sql.NullString                           `db:"cancellation_reason"`
	Lines              jsonstore.Field[[]aggregate.ReceiptLine] `db:"lines"`
	CreatedBy          string                                   `db:"created_by"`
	LastUpdatedBy      string                                   `db:"last_updated_by"`
	Status             string                                   `db:"status"`
	CreatedAt          time.Time                                `db:"created_at"`
	UpdatedAt          time.Time                                `db:"updated_at"`
	DeletedAt          sql.NullTime                             `db:"deleted_at"`
}

func (r *receiptRow) toEntity() (*aggregate.Receipt, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	cashRegisterID, err := uuidv7.Parse(r.CashRegisterID)
	if err != nil {
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	orderID, err := uuidv7.Parse(r.OrderID)
	if err != nil {
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	createdBy, err := uuidv7.Parse(r.CreatedBy)
	if err != nil {
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	lastUpdatedBy, err := uuidv7.Parse(r.LastUpdatedBy)
	if err != nil {
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	rec := &aggregate.Receipt{
		CashRegisterID: cashRegisterID,
		OrderID:        orderID,
		PaymentType:    aggregate.PaymentType(r.PaymentType),
		ReceiptType:    aggregate.ReceiptType(r.ReceiptType),
		Currency:       r.Currency,
		TotalAmount:    r.TotalAmount,
		TaxAmount:      r.TaxAmount,
		Lines:          r.Lines,
		CreatedBy:      createdBy,
		LastUpdatedBy:  lastUpdatedBy,
		Status:         aggregate.ReceiptStatus(r.Status),
	}

	rec.ID = id
	rec.Version = r.Version
	rec.CreatedAt = r.CreatedAt
	rec.UpdatedAt = r.UpdatedAt

	if r.FiscalNumber.Valid {
		rec.FiscalNumber = r.FiscalNumber.String
	}
	if r.FiscalURL.Valid {
		rec.FiscalURL = r.FiscalURL.String
	}
	if r.QRCode.Valid {
		rec.QRCode = r.QRCode.String
	}
	if r.ProviderReceiptID.Valid {
		rec.ProviderReceiptID = r.ProviderReceiptID.String
	}
	if r.PrintedAt.Valid {
		rec.PrintedAt = &r.PrintedAt.Time
	}
	if r.CancelledAt.Valid {
		rec.CancelledAt = &r.CancelledAt.Time
	}
	if r.CancellationReason.Valid {
		rec.CancellationReason = r.CancellationReason.String
	}
	if r.DeletedAt.Valid {
		rec.DeletedAt = &r.DeletedAt.Time
	}

	return rec, nil
}

func fromEntity(rec *aggregate.Receipt) *receiptRow {
	row := &receiptRow{
		ID:             rec.GetID().String(),
		Version:        rec.GetVersion(),
		CashRegisterID: rec.CashRegisterID.String(),
		OrderID:        rec.OrderID.String(),
		PaymentType:    string(rec.PaymentType),
		ReceiptType:    string(rec.ReceiptType),
		Currency:       rec.Currency,
		TotalAmount:    rec.TotalAmount,
		TaxAmount:      rec.TaxAmount,
		Lines:          rec.Lines,
		CreatedBy:      rec.CreatedBy.String(),
		LastUpdatedBy:  rec.LastUpdatedBy.String(),
		Status:         string(rec.Status),
		CreatedAt:      rec.CreatedAt,
		UpdatedAt:      rec.UpdatedAt,
	}

	if rec.FiscalNumber != "" {
		row.FiscalNumber = sql.NullString{String: rec.FiscalNumber, Valid: true}
	}
	if rec.FiscalURL != "" {
		row.FiscalURL = sql.NullString{String: rec.FiscalURL, Valid: true}
	}
	if rec.QRCode != "" {
		row.QRCode = sql.NullString{String: rec.QRCode, Valid: true}
	}
	if rec.ProviderReceiptID != "" {
		row.ProviderReceiptID = sql.NullString{String: rec.ProviderReceiptID, Valid: true}
	}
	if rec.PrintedAt != nil {
		row.PrintedAt = sql.NullTime{Time: *rec.PrintedAt, Valid: true}
	}
	if rec.CancelledAt != nil {
		row.CancelledAt = sql.NullTime{Time: *rec.CancelledAt, Valid: true}
	}
	if rec.CancellationReason != "" {
		row.CancellationReason = sql.NullString{String: rec.CancellationReason, Valid: true}
	}
	if rec.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *rec.DeletedAt, Valid: true}
	}

	return row
}

// Create creates a new receipt
func (r *receiptRepository) Create(ctx context.Context, rec *aggregate.Receipt) error {
	query := `
		INSERT INTO fiscal_receipts (
			id, version, cash_register_id, order_id, payment_type, receipt_type,
			currency, total_amount, tax_amount, fiscal_number, fiscal_url, qr_code, provider_receipt_id,
			printed_at, cancelled_at, cancellation_reason, lines, created_by, last_updated_by,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)`

	row := fromEntity(rec)
	_, err := r.Exec(ctx, query,
		row.ID, row.Version, row.CashRegisterID, row.OrderID, row.PaymentType, row.ReceiptType,
		row.Currency, row.TotalAmount, row.TaxAmount, row.FiscalNumber, row.FiscalURL, row.QRCode, row.ProviderReceiptID,
		row.PrintedAt, row.CancelledAt, row.CancellationReason, row.Lines, row.CreatedBy, row.LastUpdatedBy,
		row.Status, row.CreatedAt, row.UpdatedAt,
	)
	if err != nil {
		return receipterrors.ErrReceiptCreateFailed
	}

	return nil
}

// GetByID retrieves receipt by ID
func (r *receiptRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Receipt, error) {
	query := `
		SELECT id, version, cash_register_id, order_id, payment_type, receipt_type,
		       currency, total_amount, tax_amount, fiscal_number, fiscal_url, qr_code, provider_receipt_id,
		       printed_at, cancelled_at, cancellation_reason, lines, created_by, last_updated_by,
		       status, created_at, updated_at, deleted_at
		FROM fiscal_receipts
		WHERE id = $1 AND deleted_at IS NULL`

	var row receiptRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, receipterrors.ErrReceiptNotFound
		}
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	return row.toEntity()
}

// GetByOrderID retrieves receipt by order ID
func (r *receiptRepository) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Receipt, error) {
	query := `
		SELECT id, version, cash_register_id, order_id, payment_type, receipt_type,
		       currency, total_amount, tax_amount, fiscal_number, fiscal_url, qr_code, provider_receipt_id,
		       printed_at, cancelled_at, cancellation_reason, lines, created_by, last_updated_by,
		       status, created_at, updated_at, deleted_at
		FROM fiscal_receipts
		WHERE order_id = $1 AND deleted_at IS NULL`

	var row receiptRow
	if err := r.Get(ctx, &row, query, orderID.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, receipterrors.ErrReceiptNotFound
		}
		return nil, receipterrors.ErrReceiptCreateFailed
	}

	return row.toEntity()
}

// List retrieves receipts with filters
func (r *receiptRepository) List(ctx context.Context, filters *repository.ListFilters) ([]*aggregate.Receipt, error) {
	query := `
		SELECT id, version, cash_register_id, order_id, payment_type, receipt_type,
		       currency, total_amount, tax_amount, fiscal_number, fiscal_url, qr_code, provider_receipt_id,
		       printed_at, cancelled_at, cancellation_reason, lines, created_by, last_updated_by,
		       status, created_at, updated_at, deleted_at
		FROM fiscal_receipts
		WHERE deleted_at IS NULL`

	args := []interface{}{}
	argPos := 1

	if filters != nil && filters.CashRegisterID != nil {
		query += fmt.Sprintf(" AND cash_register_id = $%d", argPos)
		args = append(args, filters.CashRegisterID.String())
		argPos++
	}
	if filters != nil && filters.OrderID != nil {
		query += fmt.Sprintf(" AND order_id = $%d", argPos)
		args = append(args, filters.OrderID.String())
		argPos++
	}
	if filters != nil && filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, string(*filters.Status))
	}

	query += " ORDER BY created_at DESC"

	var rows []receiptRow
	if err := r.Select(ctx, &rows, query, args...); err != nil {
		return nil, receipterrors.ErrReceiptListFailed
	}

	receipts := make([]*aggregate.Receipt, len(rows))
	for i := range rows {
		rec, err := rows[i].toEntity()
		if err != nil {
			return nil, receipterrors.ErrReceiptCreateFailed
		}
		receipts[i] = rec
	}

	return receipts, nil
}

// Update updates receipt
func (r *receiptRepository) Update(ctx context.Context, rec *aggregate.Receipt) error {
	query := `
		UPDATE fiscal_receipts
		SET version = $1,
		    cash_register_id = $2,
		    order_id = $3,
		    payment_type = $4,
		    receipt_type = $5,
		    currency = $6,
		    total_amount = $7,
		    tax_amount = $8,
		    fiscal_number = $9,
		    fiscal_url = $10,
		    qr_code = $11,
		    provider_receipt_id = $12,
		    printed_at = $13,
		    cancelled_at = $14,
		    cancellation_reason = $15,
		    lines = $16,
		    created_by = $17,
		    last_updated_by = $18,
		    status = $19,
		    updated_at = $20
		WHERE id = $21 AND deleted_at IS NULL`

	row := fromEntity(rec)
	result, err := r.Exec(ctx, query,
		row.Version, row.CashRegisterID, row.OrderID, row.PaymentType, row.ReceiptType,
		row.Currency, row.TotalAmount, row.TaxAmount, row.FiscalNumber, row.FiscalURL, row.QRCode,
		row.ProviderReceiptID, row.PrintedAt, row.CancelledAt, row.CancellationReason, row.Lines, row.CreatedBy, row.LastUpdatedBy,
		row.Status, row.UpdatedAt, row.ID,
	)
	if err != nil {
		return receipterrors.ErrReceiptUpdateFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return receipterrors.ErrReceiptUpdateFailed
	}
	if rowsAffected == 0 {
		return receipterrors.ErrReceiptNotFound
	}

	rec.Version = rec.GetVersion() + 1
	return nil
}

// Delete soft-deletes receipt
func (r *receiptRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE fiscal_receipts
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, time.Now(), id.String())
	if err != nil {
		return receipterrors.ErrReceiptDeleteFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return receipterrors.ErrReceiptDeleteFailed
	}
	if rowsAffected == 0 {
		return receipterrors.ErrReceiptNotFound
	}

	return nil
}
