package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// dealRepository implements deal.IRepository using PostgreSQL
type dealRepository struct {
	*BaseRepository
}

// NewDealRepository creates a new PostgreSQL deal repository
func NewDealRepository(db *sqlx.DB) deal.IRepository {
	return &dealRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// dealRow represents a database row for the customer_deals table
type dealRow struct {
	ID                string         `db:"id"`
	CustomerID        string         `db:"customer_id"`
	CompanyID         sql.NullString `db:"company_id"`
	Name              string         `db:"name"`
	Description       sql.NullString `db:"description"`
	ValueCents        int64          `db:"value_cents"`
	Currency          string         `db:"currency"`
	Stage             string         `db:"stage"`
	Probability       int            `db:"probability"`
	Source            string         `db:"source"`
	ExpectedCloseDate time.Time      `db:"expected_close_date"`
	ActualCloseDate   sql.NullTime   `db:"actual_close_date"`
	AssignedTo        string         `db:"assigned_to"`
	CloseReason       sql.NullString `db:"close_reason"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	DeletedAt         sql.NullTime   `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *dealRow) toEntity() (*deal.Deal, error) {
	// Parse Deal ID
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid deal ID: %w", err)
	}

	// Parse Customer ID
	customerID, err := uuidv7.Parse(r.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	// Parse Assigned To
	assignedTo, err := uuidv7.Parse(r.AssignedTo)
	if err != nil {
		return nil, fmt.Errorf("invalid assigned to ID: %w", err)
	}

	// Create Money value object
	money, err := valueobject.NewMoney(r.ValueCents, r.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid money value: %w", err)
	}

	// Create Deal aggregate
	d := &deal.Deal{
		CustomerID:        customerID,
		Name:              r.Name,
		Value:             money,
		Currency:          r.Currency,
		Stage:             deal.DealStage(r.Stage),
		Probability:       r.Probability,
		Source:            deal.DealSource(r.Source),
		ExpectedCloseDate: r.ExpectedCloseDate,
		AssignedTo:        assignedTo,
	}
	
	// Set BaseAggregate fields
	d.BaseAggregate.ID = id
	d.BaseAggregate.CreatedAt = r.CreatedAt
	d.BaseAggregate.UpdatedAt = r.UpdatedAt

	// Parse optional Company ID
	if r.CompanyID.Valid {
		companyID, err := uuidv7.Parse(r.CompanyID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid company ID: %w", err)
		}
		d.CompanyID = &companyID
	}

	// Parse optional Description
	if r.Description.Valid {
		d.Description = r.Description.String
	}

	// Parse optional Actual Close Date
	if r.ActualCloseDate.Valid {
		d.ActualCloseDate = &r.ActualCloseDate.Time
	}

	// Parse optional Close Reason
	if r.CloseReason.Valid {
		d.CloseReason = r.CloseReason.String
	}

	// Parse optional Deleted At
	if r.DeletedAt.Valid {
		d.DeletedAt = &r.DeletedAt.Time
	}

	return d, nil
}

// fromEntity converts domain entity to database row
func fromEntity(d *deal.Deal) *dealRow {
	row := &dealRow{
		ID:                d.GetID().String(),
		CustomerID:        d.CustomerID.String(),
		Name:              d.Name,
		ValueCents:        d.Value.Amount,
		Currency:          d.Currency,
		Stage:             string(d.Stage),
		Probability:       d.Probability,
		Source:            string(d.Source),
		ExpectedCloseDate: d.ExpectedCloseDate,
		AssignedTo:        d.AssignedTo.String(),
		CreatedAt:         d.GetCreatedAt(),
		UpdatedAt:         d.GetUpdatedAt(),
	}

	// Optional Company ID
	if d.CompanyID != nil {
		row.CompanyID = sql.NullString{String: d.CompanyID.String(), Valid: true}
	}

	// Optional Description
	if d.Description != "" {
		row.Description = sql.NullString{String: d.Description, Valid: true}
	}

	// Optional Actual Close Date
	if d.ActualCloseDate != nil {
		row.ActualCloseDate = sql.NullTime{Time: *d.ActualCloseDate, Valid: true}
	}

	// Optional Close Reason
	if d.CloseReason != "" {
		row.CloseReason = sql.NullString{String: d.CloseReason, Valid: true}
	}

	// Optional Deleted At
	if d.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *d.DeletedAt, Valid: true}
	}

	return row
}

// Create creates a new deal
func (r *dealRepository) Create(ctx context.Context, d *deal.Deal) error {
	row := fromEntity(d)

	query := `
		INSERT INTO customer_deals (
			id, customer_id, company_id, name, description,
			value_cents, currency, stage, probability, source,
			expected_close_date, actual_close_date, assigned_to,
			close_reason, created_at, updated_at
		) VALUES (
			:id, :customer_id, :company_id, :name, :description,
			:value_cents, :currency, :stage, :probability, :source,
			:expected_close_date, :actual_close_date, :assigned_to,
			:close_reason, :created_at, :updated_at
		)
	`

	_, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to create deal: %w", err)
	}

	return nil
}

// GetByID retrieves a deal by ID
func (r *dealRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*deal.Deal, error) {
	query := `
		SELECT * FROM customer_deals
		WHERE id = $1 AND deleted_at IS NULL
	`

	var row dealRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("deal not found")
		}
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	return row.toEntity()
}

// Update updates a deal
func (r *dealRepository) Update(ctx context.Context, d *deal.Deal) error {
	row := fromEntity(d)

	query := `
		UPDATE customer_deals SET
			customer_id = :customer_id,
			company_id = :company_id,
			name = :name,
			description = :description,
			value_cents = :value_cents,
			currency = :currency,
			stage = :stage,
			probability = :probability,
			source = :source,
			expected_close_date = :expected_close_date,
			actual_close_date = :actual_close_date,
			assigned_to = :assigned_to,
			close_reason = :close_reason,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`

	result, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update deal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deal not found")
	}

	return nil
}

// Delete soft deletes a deal
func (r *dealRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE customer_deals
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.Exec(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to delete deal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("deal not found")
	}

	return nil
}

// List returns paginated deals
func (r *dealRepository) List(ctx context.Context, page, pageSize int) ([]*deal.Deal, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM customer_deals WHERE deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM customer_deals
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var rows []dealRow
	if err := r.Select(ctx, &rows, query, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list deals: %w", err)
	}

	deals := make([]*deal.Deal, 0, len(rows))
	for _, row := range rows {
		d, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		deals = append(deals, d)
	}

	return deals, total, nil
}

// ListByStage returns deals in a specific stage
func (r *dealRepository) ListByStage(ctx context.Context, stage deal.DealStage, page, pageSize int) ([]*deal.Deal, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM customer_deals WHERE stage = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, string(stage)); err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM customer_deals
		WHERE stage = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []dealRow
	if err := r.Select(ctx, &rows, query, string(stage), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list deals by stage: %w", err)
	}

	deals := make([]*deal.Deal, 0, len(rows))
	for _, row := range rows {
		d, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		deals = append(deals, d)
	}

	return deals, total, nil
}

// ListByCustomer returns deals for a specific customer
func (r *dealRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*deal.Deal, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM customer_deals WHERE customer_id = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, customerID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM customer_deals
		WHERE customer_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []dealRow
	if err := r.Select(ctx, &rows, query, customerID.String(), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list deals by customer: %w", err)
	}

	deals := make([]*deal.Deal, 0, len(rows))
	for _, row := range rows {
		d, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		deals = append(deals, d)
	}

	return deals, total, nil
}

// ListByCompany returns deals for a specific company
func (r *dealRepository) ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*deal.Deal, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM customer_deals WHERE company_id = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, companyID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM customer_deals
		WHERE company_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []dealRow
	if err := r.Select(ctx, &rows, query, companyID.String(), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list deals by company: %w", err)
	}

	deals := make([]*deal.Deal, 0, len(rows))
	for _, row := range rows {
		d, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		deals = append(deals, d)
	}

	return deals, total, nil
}

// ListByAssignedTo returns deals assigned to a sales rep
func (r *dealRepository) ListByAssignedTo(ctx context.Context, userID uuidv7.UUID, page, pageSize int) ([]*deal.Deal, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM customer_deals WHERE assigned_to = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, userID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM customer_deals
		WHERE assigned_to = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []dealRow
	if err := r.Select(ctx, &rows, query, userID.String(), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list deals by assigned to: %w", err)
	}

	deals := make([]*deal.Deal, 0, len(rows))
	for _, row := range rows {
		d, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		deals = append(deals, d)
	}

	return deals, total, nil
}

// ListBySource returns deals from a specific source
func (r *dealRepository) ListBySource(ctx context.Context, source deal.DealSource, page, pageSize int) ([]*deal.Deal, int64, error) {
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM customer_deals WHERE source = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, string(source)); err != nil {
		return nil, 0, fmt.Errorf("failed to count deals: %w", err)
	}

	// Get paginated results
	query := `
		SELECT * FROM customer_deals
		WHERE source = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []dealRow
	if err := r.Select(ctx, &rows, query, string(source), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list deals by source: %w", err)
	}

	deals := make([]*deal.Deal, 0, len(rows))
	for _, row := range rows {
		d, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		deals = append(deals, d)
	}

	return deals, total, nil
}

// GetPipelineStats returns deal counts by stage
func (r *dealRepository) GetPipelineStats(ctx context.Context) (map[deal.DealStage]int64, error) {
	query := `
		SELECT stage, COUNT(*) as count
		FROM customer_deals
		WHERE deleted_at IS NULL
		GROUP BY stage
	`

	type statsRow struct {
		Stage string `db:"stage"`
		Count int64  `db:"count"`
	}

	var rows []statsRow
	if err := r.Select(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed to get pipeline stats: %w", err)
	}

	stats := make(map[deal.DealStage]int64)
	for _, row := range rows {
		stats[deal.DealStage(row.Stage)] = row.Count
	}

	return stats, nil
}

// GetTotalValue returns total value of all active deals
func (r *dealRepository) GetTotalValue(ctx context.Context) (int64, error) {
	query := `
		SELECT COALESCE(SUM(value_cents), 0) as total
		FROM customer_deals
		WHERE deleted_at IS NULL
		AND stage NOT IN ('closed_won', 'closed_lost')
	`

	var total int64
	if err := r.Get(ctx, &total, query); err != nil {
		return 0, fmt.Errorf("failed to get total value: %w", err)
	}

	return total, nil
}

// GetWonDeals returns won deals count and value
func (r *dealRepository) GetWonDeals(ctx context.Context) (int64, int64, error) {
	query := `
		SELECT COUNT(*) as count, COALESCE(SUM(value_cents), 0) as value
		FROM customer_deals
		WHERE stage = 'closed_won' AND deleted_at IS NULL
	`

	type wonRow struct {
		Count int64 `db:"count"`
		Value int64 `db:"value"`
	}

	var row wonRow
	if err := r.Get(ctx, &row, query); err != nil {
		return 0, 0, fmt.Errorf("failed to get won deals: %w", err)
	}

	return row.Count, row.Value, nil
}

// Exists checks if a deal exists
func (r *dealRepository) Exists(ctx context.Context, id uuidv7.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM customer_deals WHERE id = $1 AND deleted_at IS NULL)`

	var exists bool
	if err := r.Get(ctx, &exists, query, id.String()); err != nil {
		return false, fmt.Errorf("failed to check deal existence: %w", err)
	}

	return exists, nil
}
