package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type InvoiceRepository struct {
	*BaseRepository
}

func NewInvoiceRepository(db *sqlx.DB) repository.IInvoiceRepository {
	return &InvoiceRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *InvoiceRepository) Create(ctx context.Context, invoice *entity.Invoice) error {
	query := `
		INSERT INTO billing_invoices (
			id, subscription_id, invoice_number, status,
			subtotal_amount, tax_amount, total_amount, amount_paid, amount_due,
			currency, due_date, paid_at, created_at, updated_at
		) VALUES (
			:id, :subscription_id, :invoice_number, :status,
			:subtotal_amount, :tax_amount, :total_amount, :amount_paid, :amount_due,
			:currency, :due_date, :paid_at, :created_at, :updated_at
		)`
	return r.NamedExec(ctx, query, invoice)
}

func (r *InvoiceRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Invoice, error) {
	var invoice entity.Invoice
	query := `SELECT * FROM billing_invoices WHERE id = $1 AND deleted_at IS NULL`
	err := r.Get(ctx, &invoice, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &invoice, err
}

func (r *InvoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entity.Invoice, error) {
	var invoice entity.Invoice
	query := `SELECT * FROM billing_invoices WHERE invoice_number = $1 AND deleted_at IS NULL`
	err := r.Get(ctx, &invoice, query, invoiceNumber)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &invoice, err
}

func (r *InvoiceRepository) GetBySubscriptionID(ctx context.Context, subscriptionID uuidv7.UUID) ([]*entity.Invoice, error) {
	var invoices []*entity.Invoice
	query := `SELECT * FROM billing_invoices WHERE subscription_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	err := r.Select(ctx, &invoices, query, subscriptionID)
	return invoices, err
}

func (r *InvoiceRepository) GetOverdue(ctx context.Context, limit int) ([]*entity.Invoice, error) {
	var invoices []*entity.Invoice
	now := time.Now()
	query := `
		SELECT * FROM billing_invoices 
		WHERE status = 'open'
		AND due_date < $1
		AND deleted_at IS NULL
		ORDER BY due_date ASC
		LIMIT $2`
	err := r.Select(ctx, &invoices, query, now, limit)
	return invoices, err
}

func (r *InvoiceRepository) List(ctx context.Context, status *entity.InvoiceStatus, limit, offset int) ([]*entity.Invoice, error) {
	var invoices []*entity.Invoice
	query := `SELECT * FROM billing_invoices WHERE deleted_at IS NULL`
	args := []interface{}{}
	
	if status != nil {
		query += ` AND status = $1`
		args = append(args, *status)
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}
	
	err := r.Select(ctx, &invoices, query, args...)
	return invoices, err
}

func (r *InvoiceRepository) Count(ctx context.Context, status *entity.InvoiceStatus) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM billing_invoices WHERE deleted_at IS NULL`
	args := []interface{}{}
	
	if status != nil {
		query += ` AND status = $1`
		args = append(args, *status)
	}
	
	err := r.Get(ctx, &count, query, args...)
	return count, err
}

func (r *InvoiceRepository) Update(ctx context.Context, invoice *entity.Invoice) error {
	query := `
		UPDATE billing_invoices SET
			status = :status,
			subtotal_amount = :subtotal_amount,
			tax_amount = :tax_amount,
			total_amount = :total_amount,
			amount_paid = :amount_paid,
			amount_due = :amount_due,
			due_date = :due_date,
			paid_at = :paid_at,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`
	return r.NamedExec(ctx, query, invoice)
}

func (r *InvoiceRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE billing_invoices SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	return r.Exec(ctx, query, id)
}
