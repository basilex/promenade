package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	subscriptionerrors "github.com/basilex/promenade/internal/contexts/billing/subscription"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/aggregate"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/usecase"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// subscriptionRepository implements usecase.ISubscriptionRepository using PostgreSQL
type subscriptionRepository struct {
	db *sqlx.DB
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *sqlx.DB) usecase.ISubscriptionRepository {
	return &subscriptionRepository{
		db: db,
	}
}

// getExecutor returns either transaction or regular connection from context
func (r *subscriptionRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes a query that returns a single row
func (r *subscriptionRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes a query that returns multiple rows
func (r *subscriptionRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes a query that doesn't return rows
func (r *subscriptionRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query
func (r *subscriptionRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

// subscriptionRow represents the database row structure
type subscriptionRow struct {
	ID                        string         `db:"id"`
	SubscriptionNo            string         `db:"subscription_no"`
	CustomerID                string         `db:"customer_id"`
	PlanID                    string         `db:"plan_id"`
	Status                    string         `db:"status"`
	BillingPeriod             string         `db:"billing_period"`
	Currency                  string         `db:"currency"`
	AmountCents               int64          `db:"amount_cents"`
	StartDate                 sql.NullTime   `db:"start_date"`
	EndDate                   sql.NullTime   `db:"end_date"`
	RenewalDate               sql.NullTime   `db:"renewal_date"`
	TrialEndDate              sql.NullTime   `db:"trial_end_date"`
	CancelledAt               sql.NullTime   `db:"cancelled_at"`
	CancelReason              sql.NullString `db:"cancel_reason"`
	CancellationEffectiveDate sql.NullTime   `db:"cancellation_effective_date"`
	Metadata                  string         `db:"metadata"`
	CreatedAt                 sql.NullTime   `db:"created_at"`
	UpdatedAt                 sql.NullTime   `db:"updated_at"`
	DeletedAt                 sql.NullTime   `db:"deleted_at"`
}

// toEntity converts database row to subscription entity
func (r *subscriptionRow) toEntity() (*aggregate.Subscription, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid subscription ID: %w", err)
	}

	customerID, err := uuidv7.Parse(r.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	money, err := valueobject.NewMoney(r.AmountCents, r.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid money value: %w", err)
	}

	// Parse metadata using jsonstore
	metadataField := jsonstore.Field[map[string]string]{}
	if r.Metadata != "" {
		if err := metadataField.UnmarshalJSON([]byte(r.Metadata)); err != nil {
			return nil, fmt.Errorf("invalid metadata JSON: %w", err)
		}
	}

	sub := &aggregate.Subscription{
		SubscriptionNo: r.SubscriptionNo,
		CustomerID:     customerID,
		PlanID:         r.PlanID,
		Status:         aggregate.SubscriptionStatus(r.Status),
		BillingPeriod:  aggregate.BillingPeriod(r.BillingPeriod),
		Currency:       r.Currency,
		Amount:         money,
		Metadata:       metadataField,
	}

	// Set BaseAggregate fields directly
	sub.ID = id
	if r.CreatedAt.Valid {
		sub.CreatedAt = r.CreatedAt.Time
	}
	if r.UpdatedAt.Valid {
		sub.UpdatedAt = r.UpdatedAt.Time
	}
	if r.DeletedAt.Valid {
		sub.DeletedAt = &r.DeletedAt.Time
	}

	// Set dates
	if r.StartDate.Valid {
		sub.StartDate = r.StartDate.Time
	}
	if r.EndDate.Valid {
		sub.EndDate = &r.EndDate.Time
	}
	if r.RenewalDate.Valid {
		sub.RenewalDate = r.RenewalDate.Time
	}
	if r.TrialEndDate.Valid {
		sub.TrialEndDate = &r.TrialEndDate.Time
	}
	if r.CancelledAt.Valid {
		sub.CancelledAt = &r.CancelledAt.Time
	}
	if r.CancelReason.Valid {
		sub.CancelReason = r.CancelReason.String
	}
	if r.CancellationEffectiveDate.Valid {
		sub.CancellationEffectiveDate = &r.CancellationEffectiveDate.Time
	}

	return sub, nil
}

// fromEntity converts subscription entity to database row
func fromEntity(sub *aggregate.Subscription) (*subscriptionRow, error) {
	// Marshal metadata using jsonstore
	metadataJSON, err := sub.Metadata.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	row := &subscriptionRow{
		ID:             sub.GetID().String(),
		SubscriptionNo: sub.SubscriptionNo,
		CustomerID:     sub.CustomerID.String(),
		PlanID:         sub.PlanID,
		Status:         string(sub.Status),
		BillingPeriod:  string(sub.BillingPeriod),
		Currency:       sub.Currency,
		AmountCents:    sub.Amount.Amount,
		Metadata:       string(metadataJSON),
	}

	// Set dates
	row.StartDate = sql.NullTime{Time: sub.StartDate, Valid: !sub.StartDate.IsZero()}
	row.RenewalDate = sql.NullTime{Time: sub.RenewalDate, Valid: !sub.RenewalDate.IsZero()}

	if sub.EndDate != nil {
		row.EndDate = sql.NullTime{Time: *sub.EndDate, Valid: true}
	}
	if sub.TrialEndDate != nil {
		row.TrialEndDate = sql.NullTime{Time: *sub.TrialEndDate, Valid: true}
	}
	if sub.CancelledAt != nil {
		row.CancelledAt = sql.NullTime{Time: *sub.CancelledAt, Valid: true}
	}
	if sub.CancelReason != "" {
		row.CancelReason = sql.NullString{String: sub.CancelReason, Valid: true}
	}
	if sub.CancellationEffectiveDate != nil {
		row.CancellationEffectiveDate = sql.NullTime{Time: *sub.CancellationEffectiveDate, Valid: true}
	}

	// Set audit fields
	row.CreatedAt = sql.NullTime{Time: sub.GetCreatedAt(), Valid: !sub.GetCreatedAt().IsZero()}
	row.UpdatedAt = sql.NullTime{Time: sub.GetUpdatedAt(), Valid: !sub.GetUpdatedAt().IsZero()}
	if sub.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *sub.DeletedAt, Valid: true}
	}

	return row, nil
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *aggregate.Subscription) error {
	row, err := fromEntity(sub)
	if err != nil {
		return fmt.Errorf("failed to convert entity: %w", err)
	}

	query := `
		INSERT INTO billing_subscriptions (
			id, subscription_no, customer_id, plan_id, status, billing_period,
			currency, amount_cents, start_date, end_date, renewal_date, trial_end_date,
			cancelled_at, cancel_reason, cancellation_effective_date, metadata,
			created_at, updated_at, deleted_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	_, err = r.Exec(ctx, query,
		row.ID, row.SubscriptionNo, row.CustomerID, row.PlanID, row.Status, row.BillingPeriod,
		row.Currency, row.AmountCents, row.StartDate, row.EndDate, row.RenewalDate, row.TrialEndDate,
		row.CancelledAt, row.CancelReason, row.CancellationEffectiveDate, row.Metadata,
		row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
	query := `
		SELECT id, subscription_no, customer_id, plan_id, status, billing_period,
		       currency, amount_cents, start_date, end_date, renewal_date, trial_end_date,
		       cancelled_at, cancel_reason, cancellation_effective_date, metadata,
		       created_at, updated_at, deleted_at
		FROM billing_subscriptions
		WHERE id = $1 AND deleted_at IS NULL`

	var row subscriptionRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if err == sql.ErrNoRows {
			return nil, subscriptionerrors.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return row.toEntity()
}

func (r *subscriptionRepository) Update(ctx context.Context, sub *aggregate.Subscription) error {
	row, err := fromEntity(sub)
	if err != nil {
		return fmt.Errorf("failed to convert entity: %w", err)
	}

	query := `
		UPDATE billing_subscriptions
		SET subscription_no = $2, customer_id = $3, plan_id = $4, status = $5, billing_period = $6,
		    currency = $7, amount_cents = $8, start_date = $9, end_date = $10, renewal_date = $11,
		    trial_end_date = $12, cancelled_at = $13, cancel_reason = $14,
		    cancellation_effective_date = $15, metadata = $16, updated_at = $17
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query,
		row.ID, row.SubscriptionNo, row.CustomerID, row.PlanID, row.Status, row.BillingPeriod,
		row.Currency, row.AmountCents, row.StartDate, row.EndDate, row.RenewalDate, row.TrialEndDate,
		row.CancelledAt, row.CancelReason, row.CancellationEffectiveDate, row.Metadata, row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return subscriptionerrors.ErrSubscriptionNotFound
	}

	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE billing_subscriptions
		SET deleted_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return subscriptionerrors.ErrSubscriptionNotFound
	}

	return nil
}

func (r *subscriptionRepository) ListSubscriptions(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT id, subscription_no, customer_id, plan_id, status, billing_period,
		       currency, amount_cents, start_date, end_date, renewal_date, trial_end_date,
		       cancelled_at, cancel_reason, cancellation_effective_date, metadata,
		       created_at, updated_at, deleted_at
		FROM billing_subscriptions
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	var rows []subscriptionRow
	if err := r.Select(ctx, &rows, query, pageSize, offset); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	subscriptions := make([]*aggregate.Subscription, 0, len(rows))
	for _, row := range rows {
		sub, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert row: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*aggregate.Subscription, error) {
	query := `
		SELECT id, subscription_no, customer_id, plan_id, status, billing_period,
		       currency, amount_cents, start_date, end_date, renewal_date, trial_end_date,
		       cancelled_at, cancel_reason, cancellation_effective_date, metadata,
		       created_at, updated_at, deleted_at
		FROM billing_subscriptions
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var rows []subscriptionRow
	if err := r.Select(ctx, &rows, query, customerID.String()); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by customer: %w", err)
	}

	subscriptions := make([]*aggregate.Subscription, 0, len(rows))
	for _, row := range rows {
		sub, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert row: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) ListByStatus(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error) {
	query := `
		SELECT id, subscription_no, customer_id, plan_id, status, billing_period,
		       currency, amount_cents, start_date, end_date, renewal_date, trial_end_date,
		       cancelled_at, cancel_reason, cancellation_effective_date, metadata,
		       created_at, updated_at, deleted_at
		FROM billing_subscriptions
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	var rows []subscriptionRow
	if err := r.Select(ctx, &rows, query, string(status)); err != nil {
		return nil, fmt.Errorf("failed to list subscriptions by status: %w", err)
	}

	subscriptions := make([]*aggregate.Subscription, 0, len(rows))
	for _, row := range rows {
		sub, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert row: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *subscriptionRepository) CountByStatus(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM billing_subscriptions
		WHERE status = $1 AND deleted_at IS NULL`

	var count int64
	if err := r.Get(ctx, &count, query, string(status)); err != nil {
		return 0, fmt.Errorf("failed to count subscriptions: %w", err)
	}

	return count, nil
}

func (r *subscriptionRepository) GetTotalRevenue(ctx context.Context) (int64, error) {
	query := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN billing_period = 'monthly' THEN amount_cents
				WHEN billing_period = 'quarterly' THEN amount_cents / 3
				WHEN billing_period = 'yearly' THEN amount_cents / 12
				ELSE 0
			END
		), 0) as total_mrr
		FROM billing_subscriptions
		WHERE status IN ('active', 'trial') AND deleted_at IS NULL`

	var totalMRR int64
	if err := r.Get(ctx, &totalMRR, query); err != nil {
		return 0, fmt.Errorf("failed to calculate total revenue: %w", err)
	}

	return totalMRR, nil
}
