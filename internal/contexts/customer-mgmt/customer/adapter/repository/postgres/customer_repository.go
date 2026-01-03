package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// customerRepository implements customer.ICustomerRepository
type customerRepository struct {
	*BaseRepository
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *sqlx.DB) customer.ICustomerRepository {
	return &customerRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// customerRow represents database row
type customerRow struct {
	ID              string               `db:"id"`
	UserID          sql.NullString       `db:"user_id"`
	CompanyID       sql.NullString       `db:"company_id"`
	Name            string               `db:"name"`
	Email           string               `db:"email"`
	Phone           sql.NullString       `db:"phone"`
	Status          string                       `db:"status"`
	Tier            string                       `db:"tier"`
	Source          string                       `db:"source"`
	AssignedTo      string                       `db:"assigned_to"`
	Tags            jsonstore.Field[[]string]    `db:"tags"`
	CreatedAt       time.Time                    `db:"created_at"`
	UpdatedAt       time.Time            `db:"updated_at"`
	LastContactedAt sql.NullTime         `db:"last_contacted_at"`
	ConvertedAt     sql.NullTime         `db:"converted_at"`
	ChurnedAt       sql.NullTime         `db:"churned_at"`
	ChurnReason     sql.NullString       `db:"churn_reason"`
	DeletedAt       sql.NullTime         `db:"deleted_at"`
}

// toEntity converts database row to domain entity
func (r *customerRow) toEntity() (*customer.Customer, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	assignedTo, err := uuidv7.Parse(r.AssignedTo)
	if err != nil {
		return nil, fmt.Errorf("invalid assigned_to ID: %w", err)
	}

	email, err := valueobject.NewEmail(r.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	c := &customer.Customer{
		Name:       r.Name,
		Email:      email,
		Status:     customer.CustomerStatus(r.Status),
		Tier:       customer.CustomerTier(r.Tier),
		Source:     r.Source,
		AssignedTo: assignedTo,
	}
	
	// Set BaseAggregate fields
	c.BaseAggregate.ID = id
	c.BaseAggregate.CreatedAt = r.CreatedAt
	c.BaseAggregate.UpdatedAt = r.UpdatedAt

	// Optional UserID
	if r.UserID.Valid {
		userID, err := uuidv7.Parse(r.UserID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id: %w", err)
		}
		c.UserID = &userID
	}

	// Optional CompanyID
	if r.CompanyID.Valid {
		companyID, err := uuidv7.Parse(r.CompanyID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid company_id: %w", err)
		}
		c.CompanyID = &companyID
	}

	// Optional Phone
	if r.Phone.Valid {
		phone, err := valueobject.NewPhone(r.Phone.String)
		if err != nil {
			return nil, fmt.Errorf("invalid phone: %w", err)
		}
		c.Phone = &phone
	}

	// Tags (JSON array stored as TEXT)
	if !r.Tags.IsNull() {
		c.Tags = r.Tags.Get()
	}

	// Optional timestamps
	if r.LastContactedAt.Valid {
		c.LastContactedAt = &r.LastContactedAt.Time
	}
	if r.ConvertedAt.Valid {
		c.ConvertedAt = &r.ConvertedAt.Time
	}
	if r.ChurnedAt.Valid {
		c.ChurnedAt = &r.ChurnedAt.Time
	}
	if r.ChurnReason.Valid {
		c.ChurnReason = r.ChurnReason.String
	}
	if r.DeletedAt.Valid {
		c.DeletedAt = &r.DeletedAt.Time
	}

	return c, nil
}

// toRow converts domain entity to database row
func toRow(c *customer.Customer) (*customerRow, error) {
	row := &customerRow{
		ID:         c.GetID().String(),
		Name:       c.Name,
		Email:      c.Email.Value(),
		Status:     string(c.Status),
		Tier:       string(c.Tier),
		Source:     c.Source,
		AssignedTo: c.AssignedTo.String(),
		CreatedAt:  c.GetCreatedAt(),
		UpdatedAt:  c.GetUpdatedAt(),
	}

	// Optional UserID
	if c.UserID != nil {
		row.UserID = sql.NullString{String: c.UserID.String(), Valid: true}
	}

	// Optional CompanyID
	if c.CompanyID != nil {
		row.CompanyID = sql.NullString{String: c.CompanyID.String(), Valid: true}
	}

	// Optional Phone
	if c.Phone != nil {
		row.Phone = sql.NullString{String: c.Phone.Value(), Valid: true}
	}

	// Tags (JSON array stored as TEXT)
	row.Tags.Set(c.Tags)

	// Optional timestamps
	if c.LastContactedAt != nil {
		row.LastContactedAt = sql.NullTime{Time: *c.LastContactedAt, Valid: true}
	}
	if c.ConvertedAt != nil {
		row.ConvertedAt = sql.NullTime{Time: *c.ConvertedAt, Valid: true}
	}
	if c.ChurnedAt != nil {
		row.ChurnedAt = sql.NullTime{Time: *c.ChurnedAt, Valid: true}
	}
	if c.ChurnReason != "" {
		row.ChurnReason = sql.NullString{String: c.ChurnReason, Valid: true}
	}
	if c.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *c.DeletedAt, Valid: true}
	}

	return row, nil
}

// Create inserts a new customer
func (r *customerRepository) Create(ctx context.Context, c *customer.Customer) error {
	row, err := toRow(c)
	if err != nil {
		return fmt.Errorf("failed to convert to row: %w", err)
	}

	query := `
		INSERT INTO customer_customers (
			id, user_id, company_id, name, email, phone,
			status, tier, source, assigned_to, tags,
			created_at, updated_at, last_contacted_at, converted_at,
			churned_at, churn_reason, deleted_at
		) VALUES (
			:id, :user_id, :company_id, :name, :email, :phone,
			:status, :tier, :source, :assigned_to, :tags,
			:created_at, :updated_at, :last_contacted_at, :converted_at,
			:churned_at, :churn_reason, :deleted_at
		)`

	if _, err := r.NamedExec(ctx, query, row); err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by ID
func (r *customerRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*customer.Customer, error) {
	var row customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return row.toEntity()
}

// GetByEmail retrieves a customer by email
func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	var row customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE email = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return row.toEntity()
}

// GetByUserID retrieves a customer by linked user ID
func (r *customerRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*customer.Customer, error) {
	var row customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE user_id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, userID.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer by user ID: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing customer
func (r *customerRepository) Update(ctx context.Context, c *customer.Customer) error {
	row, err := toRow(c)
	if err != nil {
		return fmt.Errorf("failed to convert to row: %w", err)
	}

	query := `
		UPDATE customer_customers SET
			user_id = :user_id,
			company_id = :company_id,
			name = :name,
			email = :email,
			phone = :phone,
			status = :status,
			tier = :tier,
			source = :source,
			assigned_to = :assigned_to,
			tags = :tags,
			updated_at = :updated_at,
			last_contacted_at = :last_contacted_at,
			converted_at = :converted_at,
			churned_at = :churned_at,
			churn_reason = :churn_reason
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("customer not found or already deleted")
	}

	return nil
}

// Delete soft-deletes a customer
func (r *customerRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE customer_customers
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("customer not found or already deleted")
	}

	return nil
}

// ExistsByEmail checks if a customer with given email exists
func (r *customerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM customer_customers
			WHERE email = $1 AND deleted_at IS NULL
		)`

	if err := r.Get(ctx, &exists, query, email); err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

// ListByAssignedTo retrieves all customers assigned to a sales rep
func (r *customerRepository) ListByAssignedTo(ctx context.Context, repID uuidv7.UUID, limit, offset int) ([]*customer.Customer, int, error) {
	var rows []customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE assigned_to = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, repID.String(), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list customers by assigned to: %w", err)
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM customer_customers
		WHERE assigned_to = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, repID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	customers := make([]*customer.Customer, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, total, nil
}

// ListByCompanyID retrieves all customers for a company (B2B)
func (r *customerRepository) ListByCompanyID(ctx context.Context, companyID uuidv7.UUID) ([]*customer.Customer, error) {
	var rows []customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE company_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`

	if err := r.Select(ctx, &rows, query, companyID.String()); err != nil {
		return nil, fmt.Errorf("failed to list customers by company: %w", err)
	}

	customers := make([]*customer.Customer, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, nil
}

// ListByStatus retrieves customers by status
func (r *customerRepository) ListByStatus(ctx context.Context, status customer.CustomerStatus, limit, offset int) ([]*customer.Customer, int, error) {
	var rows []customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, string(status), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list customers by status: %w", err)
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM customer_customers
		WHERE status = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, string(status)); err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	customers := make([]*customer.Customer, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, total, nil
}

// ListByTier retrieves customers by tier
func (r *customerRepository) ListByTier(ctx context.Context, tier customer.CustomerTier, limit, offset int) ([]*customer.Customer, int, error) {
	var rows []customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE tier = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, string(tier), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list customers by tier: %w", err)
	}

	// Get total count
	var total int
	countQuery := `
		SELECT COUNT(*) FROM customer_customers
		WHERE tier = $1 AND deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery, string(tier)); err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	customers := make([]*customer.Customer, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, total, nil
}

// List retrieves customers with pagination
func (r *customerRepository) List(ctx context.Context, limit, offset int) ([]*customer.Customer, int, error) {
	var rows []customerRow
	query := `
		SELECT * FROM customer_customers
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	if err := r.Select(ctx, &rows, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM customer_customers WHERE deleted_at IS NULL`
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	customers := make([]*customer.Customer, 0, len(rows))
	for _, row := range rows {
		c, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row to entity: %w", err)
		}
		customers = append(customers, c)
	}

	return customers, total, nil
}

// CountByStatus counts customers by status
func (r *customerRepository) CountByStatus(ctx context.Context, status customer.CustomerStatus) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM customer_customers
		WHERE status = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &count, query, string(status)); err != nil {
		return 0, fmt.Errorf("failed to count by status: %w", err)
	}

	return count, nil
}

// CountByTier counts customers by tier
func (r *customerRepository) CountByTier(ctx context.Context, tier customer.CustomerTier) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM customer_customers
		WHERE tier = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &count, query, string(tier)); err != nil {
		return 0, fmt.Errorf("failed to count by tier: %w", err)
	}

	return count, nil
}

// CountByAllStatuses returns counts grouped by all statuses (optimized, single query)
func (r *customerRepository) CountByAllStatuses(ctx context.Context) (map[customer.CustomerStatus]int, error) {
	type statusCount struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}

	var results []statusCount
	query := `
		SELECT status, COUNT(*) as count
		FROM customer_customers
		WHERE deleted_at IS NULL
		GROUP BY status`

	if err := r.Select(ctx, &results, query); err != nil {
		return nil, fmt.Errorf("failed to count by all statuses: %w", err)
	}

	counts := make(map[customer.CustomerStatus]int)
	for _, result := range results {
		counts[customer.CustomerStatus(result.Status)] = result.Count
	}

	return counts, nil
}

// CountByAllTiers returns counts grouped by all tiers (optimized, single query)
func (r *customerRepository) CountByAllTiers(ctx context.Context) (map[customer.CustomerTier]int, error) {
	type tierCount struct {
		Tier  string `db:"tier"`
		Count int    `db:"count"`
	}

	var results []tierCount
	query := `
		SELECT tier, COUNT(*) as count
		FROM customer_customers
		WHERE deleted_at IS NULL
		GROUP BY tier`

	if err := r.Select(ctx, &results, query); err != nil {
		return nil, fmt.Errorf("failed to count by all tiers: %w", err)
	}

	counts := make(map[customer.CustomerTier]int)
	for _, result := range results {
		counts[customer.CustomerTier(result.Tier)] = result.Count
	}

	return counts, nil
}
