package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// receiptRepository implements receipt.IRepository using PostgreSQL
type receiptRepository struct {
	*BaseRepository
}

// NewReceiptRepository creates a new PostgreSQL receipt repository
func NewReceiptRepository(db *sqlx.DB) receipt.IRepository {
	return &receiptRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

type receiptRow struct {
	ID                 string                                 `db:"id"`
	Version            int                                    `db:"version"`
	CashRegisterID     string                                 `db:"cash_register_id"`
	OrderID            string                                 `db:"order_id"`
	PaymentType        string                                 `db:"payment_type"`
	ReceiptType        string                                 `db:"receipt_type"`
	Currency           string                                 `db:"currency"`
	TotalAmount        int64                                  `db:"total_amount"`
	TaxAmount          int64                                  `db:"tax_amount"`
	FiscalNumber       sql.NullString                         `db:"fiscal_number"`
	FiscalURL          sql.NullString                         `db:"fiscal_url"`
	QRCode             sql.NullString                         `db:"qr_code"`
	ProviderReceiptID  sql.NullString                         `db:"provider_receipt_id"`
	PrintedAt          sql.NullTime                           `db:"printed_at"`
	CancelledAt        sql.NullTime                           `db:"cancelled_at"`
	CancellationReason sql.NullString                         `db:"cancellation_reason"`
	Lines              jsonstore.Field[[]receipt.ReceiptLine] `db:"lines"`
	CreatedBy          string                                 `db:"created_by"`
	LastUpdatedBy      string                                 `db:"last_updated_by"`
	Status             string                                 `db:"status"`
	CreatedAt          time.Time                              `db:"created_at"`
	UpdatedAt          time.Time                              `db:"updated_at"`
	DeletedAt          sql.NullTime                           `db:"deleted_at"`
}

func (r *receiptRow) toEntity() (*receipt.Receipt, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, receipt.ErrReceiptCreateFailed
	}

	cashRegisterID, err := uuidv7.Parse(r.CashRegisterID)
	if err != nil {
		return nil, receipt.ErrReceiptCreateFailed
	}

	orderID, err := uuidv7.Parse(r.OrderID)
	if err != nil {
		return nil, receipt.ErrReceiptCreateFailed
	}

	createdBy, err := uuidv7.Parse(r.CreatedBy)
	if err != nil {
		return nil, receipt.ErrReceiptCreateFailed
	}

	lastUpdatedBy, err := uuidv7.Parse(r.LastUpdatedBy)
	if err != nil {
		return nil, receipt.ErrReceiptCreateFailed
	}

	rec := &receipt.Receipt{
		CashRegisterID: cashRegisterID,
		OrderID:        orderID,
		PaymentType:    receipt.PaymentType(r.PaymentType),
		ReceiptType:    receipt.ReceiptType(r.ReceiptType),
		Currency:       r.Currency,
		TotalAmount:    r.TotalAmount,
		TaxAmount:      r.TaxAmount,
		Lines:          r.Lines,
		CreatedBy:      createdBy,
		LastUpdatedBy:  lastUpdatedBy,
		Status:         receipt.ReceiptStatus(r.Status),
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

func fromEntity(rec *receipt.Receipt) *receiptRow {
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
func (r *receiptRepository) Create(ctx context.Context, rec *receipt.Receipt) error {
	query := `
		INSERT INTO fiscal_receipts (
			id, version, cash_register_id, order_id, payment_type, receipt_type,
			currency, total_amount, tax_amount, fiscal_number, fiscal_url, qr_code, provider_receipt_id,
			printed_at, cancelled_at, cancellation_reason, lines, created_by, last_updated_by,
			status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	row := fromEntity(rec)
	_, err := r.Exec(ctx, query,
		row.ID, row.Version, row.CashRegisterID, row.OrderID, row.PaymentType, row.ReceiptType,
		row.Currency, row.TotalAmount, row.TaxAmount, row.FiscalNumber, row.FiscalURL, row.QRCode, row.ProviderReceiptID,
		row.PrintedAt, row.CancelledAt, row.CancellationReason, row.Lines, row.CreatedBy, row.LastUpdatedBy,
		row.Status, row.CreatedAt, row.UpdatedAt,
	)
	if err != nil {
		return receipt.ErrReceiptCreateFailed
	}

	return nil
}

// GetByID retrieves receipt by ID
func (r *receiptRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*receipt.Receipt, error) {
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
			return nil, receipt.ErrReceiptNotFound
		}
		return nil, receipt.ErrReceiptCreateFailed
	}

	return row.toEntity()
}

// GetByOrderID retrieves receipt by order ID
func (r *receiptRepository) GetByOrderID(ctx context.Context, orderID uuidv7.UUID) (*receipt.Receipt, error) {
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
			return nil, receipt.ErrReceiptNotFound
		}
		return nil, receipt.ErrReceiptCreateFailed
	}

	return row.toEntity()
}

// List retrieves receipts with filters
func (r *receiptRepository) List(ctx context.Context, filters *receipt.ListFilters) ([]*receipt.Receipt, error) {
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
		return nil, receipt.ErrReceiptListFailed
	}

	receipts := make([]*receipt.Receipt, len(rows))
	for i := range rows {
		rec, err := rows[i].toEntity()
		if err != nil {
			return nil, receipt.ErrReceiptCreateFailed
		}
		receipts[i] = rec
	}

	return receipts, nil
}

// Update updates receipt
func (r *receiptRepository) Update(ctx context.Context, rec *receipt.Receipt) error {
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
		return receipt.ErrReceiptUpdateFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return receipt.ErrReceiptUpdateFailed
	}
	if rowsAffected == 0 {
		return receipt.ErrReceiptNotFound
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
		return receipt.ErrReceiptDeleteFailed
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return receipt.ErrReceiptDeleteFailed
	}
	if rowsAffected == 0 {
		return receipt.ErrReceiptNotFound
	}

	return nil
}
