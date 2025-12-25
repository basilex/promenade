package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type PaymentRepository struct {
	*BaseRepository
}

func NewPaymentRepository(db *sqlx.DB) repository.IPaymentRepository {
	return &PaymentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *PaymentRepository) Create(ctx context.Context, payment *entity.Payment) error {
	query := `
		INSERT INTO billing_payments (
			id, invoice_id, user_id, transaction_id, amount, currency,
			status, payment_method, payment_details, failure_code, failure_message,
			refunded_amount, created_at, updated_at
		) VALUES (
			:id, :invoice_id, :user_id, :transaction_id, :amount, :currency,
			:status, :payment_method, :payment_details, :failure_code, :failure_message,
			:refunded_amount, :created_at, :updated_at
		)`
	return r.NamedExec(ctx, query, payment)
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Payment, error) {
	var payment entity.Payment
	query := `SELECT * FROM billing_payments WHERE id = $1 AND deleted_at IS NULL`
	err := r.Get(ctx, &payment, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &payment, err
}

func (r *PaymentRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error) {
	var payment entity.Payment
	query := `SELECT * FROM billing_payments WHERE transaction_id = $1 AND deleted_at IS NULL`
	err := r.Get(ctx, &payment, query, transactionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &payment, err
}

func (r *PaymentRepository) GetByInvoiceID(ctx context.Context, invoiceID uuidv7.UUID) ([]*entity.Payment, error) {
	var payments []*entity.Payment
	query := `SELECT * FROM billing_payments WHERE invoice_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	err := r.Select(ctx, &payments, query, invoiceID)
	return payments, err
}

func (r *PaymentRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) ([]*entity.Payment, error) {
	var payments []*entity.Payment
	query := `SELECT * FROM billing_payments WHERE user_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	err := r.Select(ctx, &payments, query, userID)
	return payments, err
}

func (r *PaymentRepository) List(ctx context.Context, status *entity.PaymentStatus, limit, offset int) ([]*entity.Payment, error) {
	var payments []*entity.Payment
	query := `SELECT * FROM billing_payments WHERE deleted_at IS NULL`
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
	
	err := r.Select(ctx, &payments, query, args...)
	return payments, err
}

func (r *PaymentRepository) Count(ctx context.Context, status *entity.PaymentStatus) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM billing_payments WHERE deleted_at IS NULL`
	args := []interface{}{}
	
	if status != nil {
		query += ` AND status = $1`
		args = append(args, *status)
	}
	
	err := r.Get(ctx, &count, query, args...)
	return count, err
}

func (r *PaymentRepository) Update(ctx context.Context, payment *entity.Payment) error {
	query := `
		UPDATE billing_payments SET
			status = :status,
			payment_details = :payment_details,
			failure_code = :failure_code,
			failure_message = :failure_message,
			refunded_amount = :refunded_amount,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`
	return r.NamedExec(ctx, query, payment)
}

func (r *PaymentRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE billing_payments SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	return r.Exec(ctx, query, id)
}
