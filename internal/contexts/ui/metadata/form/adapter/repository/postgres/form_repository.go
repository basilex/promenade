package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/ui/metadata/form/aggregate"
	"github.com/basilex/promenade/internal/contexts/ui/metadata/form/usecase"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// formRepository implements usecase.IFormRepository
type formRepository struct {
	db *sqlx.DB
}

// NewFormRepository creates a new form repository
func NewFormRepository(db *sqlx.DB) usecase.IFormRepository {
	return &formRepository{db: db}
}

// getExecutor returns either transaction or regular connection from context
func (r *formRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return r.db
}

// Get executes query and scans single row
func (r *formRepository) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.GetContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Select executes query and scans multiple rows
func (r *formRepository) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return sqlx.SelectContext(ctx, r.getExecutor(ctx), dest, query, args...)
}

// Exec executes query without returning rows
func (r *formRepository) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return r.getExecutor(ctx).ExecContext(ctx, query, args...)
}

// NamedExec executes a named query
func (r *formRepository) NamedExec(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return sqlx.NamedExecContext(ctx, r.getExecutor(ctx), query, arg)
}

type formRow struct {
	ID          string                                        `db:"id"`
	FormID      string                                        `db:"form_id"`
	EntityType  string                                        `db:"entity_type"`
	Name        string                                        `db:"name"`
	Description sql.NullString                                `db:"description"`
	Layout      jsonstore.Field[map[string]any]               `db:"layout"`
	Fields      jsonstore.Field[[]map[string]any]             `db:"fields"`
	Validation  jsonstore.Field[map[string]any]               `db:"validation"`
	Events      jsonstore.Field[map[string]any]               `db:"events"`
	Permissions jsonstore.Field[map[string]any]               `db:"permissions"`
	I18n        jsonstore.Field[map[string]map[string]string] `db:"i18n"`
	Version     int                                           `db:"version"`
	IsActive    bool                                          `db:"is_active"`
	TenantID    sql.NullString                                `db:"tenant_id"`
	CreatedBy   sql.NullString                                `db:"created_by"`
	CreatedAt   time.Time                                     `db:"created_at"`
	UpdatedAt   time.Time                                     `db:"updated_at"`
	DeletedAt   sql.NullTime                                  `db:"deleted_at"`
}

func (r *formRow) toEntity() (*aggregate.FormDefinition, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	entity := &aggregate.FormDefinition{
		FormID:      r.FormID,
		EntityType:  r.EntityType,
		Name:        r.Name,
		Layout:      r.Layout,
		Fields:      r.Fields,
		Validation:  r.Validation,
		Events:      r.Events,
		Permissions: r.Permissions,
		I18n:        r.I18n,
		IsActive:    r.IsActive,
	}

	entity.ID = id
	entity.Version = r.Version
	entity.CreatedAt = r.CreatedAt
	entity.UpdatedAt = r.UpdatedAt

	if r.Description.Valid {
		entity.Description = r.Description.String
	}

	if r.TenantID.Valid {
		tenantID, err := uuidv7.Parse(r.TenantID.String)
		if err == nil {
			entity.TenantID = &tenantID
		}
	}

	if r.CreatedBy.Valid {
		createdBy, err := uuidv7.Parse(r.CreatedBy.String)
		if err == nil {
			entity.CreatedBy = &createdBy
		}
	}

	if r.DeletedAt.Valid {
		entity.DeletedAt = &r.DeletedAt.Time
	}

	return entity, nil
}

func toRow(f *aggregate.FormDefinition) *formRow {
	row := &formRow{
		ID:          f.GetID().String(),
		FormID:      f.FormID,
		EntityType:  f.EntityType,
		Name:        f.Name,
		Layout:      f.Layout,
		Fields:      f.Fields,
		Validation:  f.Validation,
		Events:      f.Events,
		Permissions: f.Permissions,
		I18n:        f.I18n,
		Version:     f.Version,
		IsActive:    f.IsActive,
		CreatedAt:   f.GetCreatedAt(),
		UpdatedAt:   f.GetUpdatedAt(),
	}

	if f.Description != "" {
		row.Description = sql.NullString{String: f.Description, Valid: true}
	}

	if f.TenantID != nil {
		row.TenantID = sql.NullString{String: f.TenantID.String(), Valid: true}
	}

	if f.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: f.CreatedBy.String(), Valid: true}
	}

	if f.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *f.DeletedAt, Valid: true}
	}

	return row
}

func (r *formRepository) Create(ctx context.Context, formEntity *aggregate.FormDefinition) error {
	row := toRow(formEntity)
	query := `INSERT INTO ui_form_definitions (
		id, form_id, entity_type, name, description,
		layout, fields, validation, events, permissions, i18n,
		version, is_active, tenant_id, created_by,
		created_at, updated_at, deleted_at
	) VALUES (
		$1, $2, $3, $4, $5,
		$6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15,
		$16, $17, $18
	)`

	_, err := r.Exec(ctx, query,
		row.ID, row.FormID, row.EntityType, row.Name, row.Description,
		row.Layout, row.Fields, row.Validation, row.Events, row.Permissions, row.I18n,
		row.Version, row.IsActive, row.TenantID, row.CreatedBy,
		row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	)
	return err
}

func (r *formRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error) {
	query := `SELECT * FROM ui_form_definitions WHERE id = $1 AND deleted_at IS NULL`
	var row formRow
	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		return nil, err
	}
	return row.toEntity()
}

func (r *formRepository) GetByFormID(ctx context.Context, formID string) (*aggregate.FormDefinition, error) {
	query := `SELECT * FROM ui_form_definitions WHERE form_id = $1 AND deleted_at IS NULL`
	var row formRow
	if err := r.Get(ctx, &row, query, formID); err != nil {
		return nil, err
	}
	return row.toEntity()
}

func (r *formRepository) Update(ctx context.Context, formEntity *aggregate.FormDefinition) error {
	row := toRow(formEntity)
	query := `UPDATE ui_form_definitions SET
		entity_type = $1,
		name = $2,
		description = $3,
		layout = $4,
		fields = $5,
		validation = $6,
		events = $7,
		permissions = $8,
		i18n = $9,
		version = $10,
		is_active = $11,
		tenant_id = $12,
		updated_at = $13
		WHERE id = $14 AND deleted_at IS NULL`

	_, err := r.Exec(ctx, query,
		row.EntityType,
		row.Name,
		row.Description,
		row.Layout,
		row.Fields,
		row.Validation,
		row.Events,
		row.Permissions,
		row.I18n,
		row.Version,
		row.IsActive,
		row.TenantID,
		row.UpdatedAt,
		row.ID,
	)
	return err
}

func (r *formRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `UPDATE ui_form_definitions SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.Exec(ctx, query, time.Now(), id.String())
	return err
}

func (r *formRepository) List(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
	if entityType == "" {
		return r.ListAll(ctx, limit, offset)
	}

	countQuery := `SELECT COUNT(*) FROM ui_form_definitions WHERE entity_type = $1 AND deleted_at IS NULL`
	var total int
	if err := r.Get(ctx, &total, countQuery, entityType); err != nil {
		return nil, 0, err
	}

	query := `SELECT * FROM ui_form_definitions
		WHERE entity_type = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var rows []formRow
	if err := r.Select(ctx, &rows, query, entityType, limit, offset); err != nil {
		return nil, 0, err
	}

	forms := make([]*aggregate.FormDefinition, 0, len(rows))
	for _, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		forms = append(forms, entity)
	}

	return forms, total, nil
}

func (r *formRepository) ListAll(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
	countQuery := `SELECT COUNT(*) FROM ui_form_definitions WHERE deleted_at IS NULL`
	var total int
	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}

	query := `SELECT * FROM ui_form_definitions
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	var rows []formRow
	if err := r.Select(ctx, &rows, query, limit, offset); err != nil {
		return nil, 0, err
	}

	forms := make([]*aggregate.FormDefinition, 0, len(rows))
	for _, row := range rows {
		entity, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		forms = append(forms, entity)
	}

	return forms, total, nil
}
