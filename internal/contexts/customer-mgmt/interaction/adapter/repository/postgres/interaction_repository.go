package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// interactionRepository implements interaction.IRepository using PostgreSQL
type interactionRepository struct {
	*BaseRepository
}

// NewInteractionRepository creates a new PostgreSQL interaction repository
func NewInteractionRepository(db *sqlx.DB) interaction.IRepository {
	return &interactionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// interactionRow represents a database row for interaction
type interactionRow struct {
	ID               uuidv7.UUID    `db:"id"`
	CustomerID       uuidv7.UUID    `db:"customer_id"`
	CompanyID        *uuidv7.UUID   `db:"company_id"`
	Type             string         `db:"type"`
	Direction        string         `db:"direction"`
	Outcome          *string        `db:"outcome"`
	Subject          string                          `db:"subject"`
	Description      string                          `db:"description"`
	CreatedBy        uuidv7.UUID                     `db:"created_by"`
	Attendees        jsonstore.Field[[]uuidv7.UUID]  `db:"attendees"`
	StartedAt        time.Time                       `db:"started_at"`
	EndedAt          *time.Time     `db:"ended_at"`
	DurationSec      *int           `db:"duration_sec"`
	FollowUpRequired bool           `db:"follow_up_required"`
	FollowUpDate     *time.Time     `db:"follow_up_date"`
	FollowUpNotes    string         `db:"follow_up_notes"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
	DeletedAt        *time.Time     `db:"deleted_at"`
}

// interactionRowWithRelations represents a database row with related entity names (for List queries)
// Optimization: Avoids N+1 query problem by using LEFT JOIN + ARRAY_AGG
type interactionRowWithRelations struct {
	interactionRow
	CustomerName  *string `db:"customer_name"`   // Customer name from LEFT JOIN
	CompanyName   *string `db:"company_name"`    // Company name from LEFT JOIN
	CreatedByName *string `db:"created_by_name"` // User name from LEFT JOIN
}

// toEntity converts database row to domain entity
func (r *interactionRow) toEntity() (*interaction.Interaction, error) {
	// Get attendees from jsonstore.Field
	attendees := r.Attendees.Get()

	inter := &interaction.Interaction{
		CustomerID:       r.CustomerID,
		CompanyID:        r.CompanyID,
		Type:             interaction.InteractionType(r.Type),
		Direction:        interaction.InteractionDirection(r.Direction),
		Outcome:          nil,
		Subject:          r.Subject,
		Description:      r.Description,
		CreatedBy:        r.CreatedBy,
		Attendees:        attendees,
		StartedAt:        r.StartedAt,
		EndedAt:          r.EndedAt,
		DurationSec:      r.DurationSec,
		FollowUpRequired: r.FollowUpRequired,
		FollowUpDate:     r.FollowUpDate,
		FollowUpNotes:    r.FollowUpNotes,
}

// Set BaseAggregate fields
inter.ID = r.ID
inter.CreatedAt = r.CreatedAt
inter.UpdatedAt = r.UpdatedAt
if r.DeletedAt != nil {
	inter.DeletedAt = r.DeletedAt
}

if r.Outcome != nil {
	outcome := interaction.InteractionOutcome(*r.Outcome)
	inter.Outcome = &outcome
}

return inter, nil
}

// toRow converts domain entity to database row
func toRow(inter *interaction.Interaction) *interactionRow {
	row := &interactionRow{
		ID:               inter.GetID(),
		CustomerID:       inter.CustomerID,
		CompanyID:        inter.CompanyID,
		Type:             string(inter.Type),
		Direction:        string(inter.Direction),
		Outcome:          nil,
		Subject:          inter.Subject,
		Description:      inter.Description,
		CreatedBy:        inter.CreatedBy,
		StartedAt:        inter.StartedAt,
		EndedAt:          inter.EndedAt,
		DurationSec:      inter.DurationSec,
		FollowUpRequired: inter.FollowUpRequired,
		FollowUpDate:     inter.FollowUpDate,
		FollowUpNotes:    inter.FollowUpNotes,
		CreatedAt:        inter.GetCreatedAt(),
		UpdatedAt:        inter.GetUpdatedAt(),
		DeletedAt:        inter.DeletedAt,
	}

	// Set attendees using jsonstore.Field
	row.Attendees.Set(inter.Attendees)

	if inter.Outcome != nil {
		outcome := string(*inter.Outcome)
		row.Outcome = &outcome
	}

	return row
}

// Create creates a new interaction
func (r *interactionRepository) Create(ctx context.Context, inter *interaction.Interaction) error {
	row := toRow(inter)

	query := `
		INSERT INTO customer_interactions (
			id, customer_id, company_id, type, direction, outcome,
			subject, description, created_by, attendees,
			started_at, ended_at, duration_sec,
			follow_up_required, follow_up_date, follow_up_notes,
			created_at, updated_at
		) VALUES (
			:id, :customer_id, :company_id, :type, :direction, :outcome,
			:subject, :description, :created_by, :attendees,
			:started_at, :ended_at, :duration_sec,
			:follow_up_required, :follow_up_date, :follow_up_notes,
			:created_at, :updated_at
		)`

	_, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to insert interaction: %w", err)
	}

	return nil
}

// GetByID retrieves an interaction by ID
func (r *interactionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*interaction.Interaction, error) {
	var row interactionRow
	query := `
		SELECT 
			id, customer_id, company_id, type, direction, outcome,
			subject, description, created_by, attendees,
			started_at, ended_at, duration_sec,
			follow_up_required, follow_up_date, follow_up_notes,
			created_at, updated_at, deleted_at
		FROM customer_interactions
		WHERE id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, interaction.ErrInteractionNotFound
		}
		return nil, fmt.Errorf("failed to get interaction: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing interaction
func (r *interactionRepository) Update(ctx context.Context, inter *interaction.Interaction) error {
	inter.UpdatedAt = time.Now()
	row := toRow(inter)

	query := `
		UPDATE customer_interactions SET
			company_id = :company_id,
			outcome = :outcome,
			subject = :subject,
			description = :description,
			attendees = :attendees,
			ended_at = :ended_at,
			duration_sec = :duration_sec,
			follow_up_required = :follow_up_required,
			follow_up_date = :follow_up_date,
			follow_up_notes = :follow_up_notes,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update interaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return interaction.ErrInteractionNotFound
	}

	return nil
}

// Delete soft deletes an interaction
func (r *interactionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE customer_interactions 
		SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete interaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return interaction.ErrInteractionNotFound
	}

	return nil
}

// ListByCustomer retrieves interactions for a customer
func (r *interactionRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*interaction.Interaction, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM customer_interactions 
		WHERE customer_id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &total, countQuery, customerID); err != nil {
		return nil, 0, fmt.Errorf("failed to count interactions: %w", err)
	}

	if total == 0 {
		return []*interaction.Interaction{}, 0, nil
	}

	// Get page with LEFT JOIN to avoid N+1 (loads customer, company, user names in single query)
	var rows []interactionRowWithRelations
	query := `
		SELECT 
			i.id, i.customer_id, i.company_id, i.type, i.direction, i.outcome,
			i.subject, i.description, i.created_by, i.attendees,
			i.started_at, i.ended_at, i.duration_sec,
			i.follow_up_required, i.follow_up_date, i.follow_up_notes,
			i.created_at, i.updated_at, i.deleted_at,
			c.name AS customer_name,
			NULL AS company_name,
			u.email AS created_by_name
		FROM customer_interactions i
		LEFT JOIN customer_customers c ON i.customer_id = c.id
		
		LEFT JOIN identity_users u ON i.created_by = u.id
		WHERE i.customer_id = $1 AND i.deleted_at IS NULL
		ORDER BY i.started_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, customerID, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list interactions: %w", err)
	}

	interactions := make([]*interaction.Interaction, len(rows))
	for i, row := range rows {
		inter, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		interactions[i] = inter
	}

	return interactions, total, nil
}

// ListByCompany retrieves interactions for a company
func (r *interactionRepository) ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*interaction.Interaction, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM customer_interactions 
		WHERE company_id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &total, countQuery, companyID); err != nil {
		return nil, 0, fmt.Errorf("failed to count interactions: %w", err)
	}

	if total == 0 {
		return []*interaction.Interaction{}, 0, nil
	}

	// Get page with LEFT JOIN to avoid N+1
	var rows []interactionRowWithRelations
	query := `
		SELECT 
			i.id, i.customer_id, i.company_id, i.type, i.direction, i.outcome,
			i.subject, i.description, i.created_by, i.attendees,
			i.started_at, i.ended_at, i.duration_sec,
			i.follow_up_required, i.follow_up_date, i.follow_up_notes,
			i.created_at, i.updated_at, i.deleted_at,
			c.name AS customer_name,
			NULL AS company_name,
			u.email AS created_by_name
		FROM customer_interactions i
		LEFT JOIN customer_customers c ON i.customer_id = c.id
		LEFT JOIN identity_users u ON i.created_by = u.id
		WHERE i.company_id = $1 AND i.deleted_at IS NULL
		ORDER BY i.started_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, companyID, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list interactions: %w", err)
	}

	interactions := make([]*interaction.Interaction, len(rows))
	for i, row := range rows {
		inter, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		interactions[i] = inter
	}

	return interactions, total, nil
}

// ListByType retrieves interactions by type
func (r *interactionRepository) ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*interaction.Interaction, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM customer_interactions 
		WHERE type = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &total, countQuery, interactionType); err != nil {
		return nil, 0, fmt.Errorf("failed to count interactions: %w", err)
	}

	if total == 0 {
		return []*interaction.Interaction{}, 0, nil
	}

	// Get page with LEFT JOIN to avoid N+1
	var rows []interactionRowWithRelations
	query := `
		SELECT 
			i.id, i.customer_id, i.company_id, i.type, i.direction, i.outcome,
			i.subject, i.description, i.created_by, i.attendees,
			i.started_at, i.ended_at, i.duration_sec,
			i.follow_up_required, i.follow_up_date, i.follow_up_notes,
			i.created_at, i.updated_at, i.deleted_at,
			c.name AS customer_name,
			NULL AS company_name,
			u.email AS created_by_name
		FROM customer_interactions i
		LEFT JOIN customer_customers c ON i.customer_id = c.id
		
		LEFT JOIN identity_users u ON i.created_by = u.id
		WHERE i.type = $1 AND i.deleted_at IS NULL
		ORDER BY i.started_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, interactionType, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list interactions: %w", err)
	}

	interactions := make([]*interaction.Interaction, len(rows))
	for i, row := range rows {
		inter, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		interactions[i] = inter
	}

	return interactions, total, nil
}

// ListByCreatedBy retrieves interactions created by a user
func (r *interactionRepository) ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*interaction.Interaction, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM customer_interactions 
		WHERE created_by = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &total, countQuery, createdBy); err != nil {
		return nil, 0, fmt.Errorf("failed to count interactions: %w", err)
	}

	if total == 0 {
		return []*interaction.Interaction{}, 0, nil
	}

	// Get page with LEFT JOIN to avoid N+1
	var rows []interactionRowWithRelations
	query := `
		SELECT 
			i.id, i.customer_id, i.company_id, i.type, i.direction, i.outcome,
			i.subject, i.description, i.created_by, i.attendees,
			i.started_at, i.ended_at, i.duration_sec,
			i.follow_up_required, i.follow_up_date, i.follow_up_notes,
			i.created_at, i.updated_at, i.deleted_at,
			c.name AS customer_name,
			NULL AS company_name,
			u.email AS created_by_name
		FROM customer_interactions i
		LEFT JOIN customer_customers c ON i.customer_id = c.id
		
		LEFT JOIN identity_users u ON i.created_by = u.id
		WHERE i.created_by = $1 AND i.deleted_at IS NULL
		ORDER BY i.started_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, createdBy, pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list interactions: %w", err)
	}

	interactions := make([]*interaction.Interaction, len(rows))
	for i, row := range rows {
		inter, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		interactions[i] = inter
	}

	return interactions, total, nil
}

// ListPendingFollowUps retrieves interactions with pending follow-ups
func (r *interactionRepository) ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*interaction.Interaction, int64, error) {
	offset := (page - 1) * pageSize

	// Count total
	var total int64
	countQuery := `
		SELECT COUNT(*) 
		FROM customer_interactions 
		WHERE follow_up_required = true 
		  AND (follow_up_date IS NULL OR follow_up_date <= $1)
		  AND deleted_at IS NULL`

	if err := r.Get(ctx, &total, countQuery, time.Now()); err != nil {
		return nil, 0, fmt.Errorf("failed to count interactions: %w", err)
	}

	if total == 0 {
		return []*interaction.Interaction{}, 0, nil
	}

	// Get page with LEFT JOIN to avoid N+1
	var rows []interactionRowWithRelations
	query := `
		SELECT 
			i.id, i.customer_id, i.company_id, i.type, i.direction, i.outcome,
			i.subject, i.description, i.created_by, i.attendees,
			i.started_at, i.ended_at, i.duration_sec,
			i.follow_up_required, i.follow_up_date, i.follow_up_notes,
			i.created_at, i.updated_at, i.deleted_at,
			c.name AS customer_name,
			NULL AS company_name,
			u.email AS created_by_name
		FROM customer_interactions i
		LEFT JOIN customer_customers c ON i.customer_id = c.id
		
		LEFT JOIN identity_users u ON i.created_by = u.id
		WHERE i.follow_up_required = true 
		  AND (i.follow_up_date IS NULL OR i.follow_up_date <= $1)
		  AND i.deleted_at IS NULL
		ORDER BY i.follow_up_date ASC NULLS FIRST, i.started_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, time.Now(), pageSize, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list interactions: %w", err)
	}

	interactions := make([]*interaction.Interaction, len(rows))
	for i, row := range rows {
		inter, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		interactions[i] = inter
	}

	return interactions, total, nil
}
