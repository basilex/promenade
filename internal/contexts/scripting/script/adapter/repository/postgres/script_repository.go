package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/scripting/script"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// scriptRepository implements script.IRepository
type scriptRepository struct {
	*BaseRepository
}

// NewScriptRepository creates a new script repository
func NewScriptRepository(db *sqlx.DB) script.IRepository {
	return &scriptRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// scriptRow is the database row representation
type scriptRow struct {
	ID          string                             `db:"id"`
	Name        string                             `db:"name"`
	Description sql.NullString                     `db:"description"`
	Code        string                             `db:"code"`
	Version     int                                `db:"version"`
	Status      string                             `db:"status"`
	ScriptType  sql.NullString                     `db:"script_type"`
	EntityType  sql.NullString                     `db:"entity_type"`
	Metadata    jsonstore.Field[map[string]string] `db:"metadata"`
	CreatedBy   sql.NullString                     `db:"created_by"`
	CreatedAt   time.Time                          `db:"created_at"`
	UpdatedAt   time.Time                          `db:"updated_at"`
	DeletedAt   sql.NullTime                       `db:"deleted_at"`
}

// executionRow is the database row for script execution
type executionRow struct {
	ID           string         `db:"id"`
	ScriptID     string         `db:"script_id"`
	ScriptName   string         `db:"script_name"`
	InputParams  sql.NullString `db:"input_params"`
	OutputResult sql.NullString `db:"output_result"`
	Error        sql.NullString `db:"error"`
	DurationMs   int            `db:"duration_ms"`
	ExecutedBy   string         `db:"executed_by"`
	ExecutedAt   time.Time      `db:"executed_at"`
}

// versionRow is the database row for script version snapshot
type versionRow struct {
	ID        string                             `db:"id"`
	ScriptID  string                             `db:"script_id"`
	Version   int                                `db:"version"`
	Code      string                             `db:"code"`
	Metadata  jsonstore.Field[map[string]string] `db:"metadata"`
	ChangeLog sql.NullString                     `db:"change_log"`
	CreatedBy sql.NullString                     `db:"created_by"`
	CreatedAt time.Time                          `db:"created_at"`
}

func (r *versionRow) toEntity() (*script.ScriptVersion, error) {
	scriptID, err := uuidv7.Parse(r.ScriptID)
	if err != nil {
		return nil, fmt.Errorf("invalid script ID: %w", err)
	}

	version := &script.ScriptVersion{
		ScriptID:  scriptID,
		Version:   r.Version,
		Code:      r.Code,
		Metadata:  r.Metadata,
		CreatedAt: r.CreatedAt,
	}

	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid version ID: %w", err)
	}
	version.ID = id

	if r.ChangeLog.Valid {
		version.ChangeLog = r.ChangeLog.String
	}
	if r.CreatedBy.Valid {
		createdBy, err := uuidv7.Parse(r.CreatedBy.String)
		if err == nil {
			version.CreatedBy = &createdBy
		}
	}

	return version, nil
}

// toEntity converts scriptRow to domain entity
func (r *scriptRow) toEntity() (*script.Script, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid script ID: %w", err)
	}

	s := &script.Script{
		Name:     r.Name,
		Code:     r.Code,
		Version:  r.Version,
		Status:   script.ScriptStatus(r.Status),
		Metadata: r.Metadata, // No manual JSON parsing needed!
	}

	// Set BaseAggregate fields directly
	s.ID = id
	s.CreatedAt = r.CreatedAt
	s.UpdatedAt = r.UpdatedAt

	// Optional fields
	if r.Description.Valid {
		s.Description = r.Description.String
	}

	if r.ScriptType.Valid {
		s.ScriptType = script.ScriptType(r.ScriptType.String)
	}

	if r.EntityType.Valid {
		s.EntityType = r.EntityType.String
	}

	if r.DeletedAt.Valid {
		s.DeletedAt = &r.DeletedAt.Time
	}

	return s, nil
}

// toRow converts domain entity to database row
func toRow(s *script.Script) (*scriptRow, error) {
	row := &scriptRow{
		ID:        s.GetID().String(),
		Name:      s.Name,
		Code:      s.Code,
		Version:   s.Version,
		Status:    string(s.Status),
		Metadata:  s.Metadata, // No manual JSON marshaling needed!
		CreatedAt: s.GetCreatedAt(),
		UpdatedAt: s.GetUpdatedAt(),
	}

	// Optional description
	if s.Description != "" {
		row.Description = sql.NullString{String: s.Description, Valid: true}
	}

	// Optional script type
	if s.ScriptType != "" {
		row.ScriptType = sql.NullString{String: string(s.ScriptType), Valid: true}
	}

	// Optional entity type
	if s.EntityType != "" {
		row.EntityType = sql.NullString{String: s.EntityType, Valid: true}
	}

	if s.DeletedAt != nil {
		row.DeletedAt = sql.NullTime{Time: *s.DeletedAt, Valid: true}
	}

	return row, nil
}

// toExecutionEntity converts executionRow to domain entity
func (r *executionRow) toEntity() (*script.ScriptExecution, error) {
	id, err := uuidv7.Parse(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	scriptID, err := uuidv7.Parse(r.ScriptID)
	if err != nil {
		return nil, fmt.Errorf("invalid script ID: %w", err)
	}

	executedBy, err := uuidv7.Parse(r.ExecutedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid executed_by ID: %w", err)
	}

	exec := &script.ScriptExecution{
		ID:         id,
		ScriptID:   scriptID,
		ScriptName: r.ScriptName,
		DurationMs: r.DurationMs,
		ExecutedBy: executedBy,
		ExecutedAt: r.ExecutedAt,
	}

	// Parse input params JSON
	if r.InputParams.Valid {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(r.InputParams.String), &params); err != nil {
			return nil, fmt.Errorf("failed to parse input params: %w", err)
		}
		exec.InputParams = params
	}

	// Parse output result JSON
	if r.OutputResult.Valid {
		var result interface{}
		if err := json.Unmarshal([]byte(r.OutputResult.String), &result); err != nil {
			return nil, fmt.Errorf("failed to parse output result: %w", err)
		}
		exec.OutputResult = result
	}

	// Error message
	if r.Error.Valid {
		exec.Error = &r.Error.String
	}

	return exec, nil
}

// toExecutionRow converts domain entity to database row
func toExecutionRow(e *script.ScriptExecution) (*executionRow, error) {
	row := &executionRow{
		ID:         e.ID.String(),
		ScriptID:   e.ScriptID.String(),
		ScriptName: e.ScriptName,
		DurationMs: e.DurationMs,
		ExecutedBy: e.ExecutedBy.String(),
		ExecutedAt: e.ExecutedAt,
	}

	// Marshal input params to JSON
	if len(e.InputParams) > 0 {
		paramsJSON, err := json.Marshal(e.InputParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input params: %w", err)
		}
		row.InputParams = sql.NullString{String: string(paramsJSON), Valid: true}
	}

	// Marshal output result to JSON
	if e.OutputResult != nil {
		resultJSON, err := json.Marshal(e.OutputResult)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal output result: %w", err)
		}
		row.OutputResult = sql.NullString{String: string(resultJSON), Valid: true}
	}

	// Error message
	if e.Error != nil {
		row.Error = sql.NullString{String: *e.Error, Valid: true}
	}

	return row, nil
}

func toVersionRow(v *script.ScriptVersion) *versionRow {
	row := &versionRow{
		ID:        v.ID.String(),
		ScriptID:  v.ScriptID.String(),
		Version:   v.Version,
		Code:      v.Code,
		Metadata:  v.Metadata,
		CreatedAt: v.CreatedAt,
	}

	if v.ChangeLog != "" {
		row.ChangeLog = sql.NullString{String: v.ChangeLog, Valid: true}
	}
	if v.CreatedBy != nil {
		row.CreatedBy = sql.NullString{String: v.CreatedBy.String(), Valid: true}
	}

	return row
}

// Create inserts a new script
func (r *scriptRepository) Create(ctx context.Context, s *script.Script) error {
	row, err := toRow(s)
	if err != nil {
		return fmt.Errorf("failed to convert to row: %w", err)
	}

	query := `
		INSERT INTO scripting_scripts (
			id, name, description, code, version, status, 
			script_type, entity_type, metadata,
			created_by, created_at, updated_at, deleted_at
		) VALUES (
			:id, :name, :description, :code, :version, :status,
			:script_type, :entity_type, :metadata,
			:created_by, :created_at, :updated_at, :deleted_at
		)`

	if _, err := r.NamedExec(ctx, query, row); err != nil {
		return fmt.Errorf("failed to create script: %w", err)
	}

	return nil
}

// CreateVersion inserts a new script version snapshot
func (r *scriptRepository) CreateVersion(ctx context.Context, v *script.ScriptVersion) error {
	if v == nil {
		return fmt.Errorf("script version is nil")
	}
	row := toVersionRow(v)
	query := `
		INSERT INTO scripting_script_versions (
			id, script_id, version, code, metadata,
			change_log, created_by, created_at
		) VALUES (
			:id, :script_id, :version, :code, :metadata,
			:change_log, :created_by, :created_at
		)`

	if _, err := r.NamedExec(ctx, query, row); err != nil {
		return fmt.Errorf("failed to create script version: %w", err)
	}

	return nil
}

// ListVersions retrieves script versions for a script ID
func (r *scriptRepository) ListVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*script.ScriptVersion, int, error) {
	countQuery := `SELECT COUNT(*) FROM scripting_script_versions WHERE script_id = $1`
	var total int
	if err := r.Get(ctx, &total, countQuery, scriptID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to count script versions: %w", err)
	}

	query := `SELECT * FROM scripting_script_versions
		WHERE script_id = $1
		ORDER BY version DESC
		LIMIT $2 OFFSET $3`

	var rows []versionRow
	if err := r.Select(ctx, &rows, query, scriptID.String(), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list script versions: %w", err)
	}

	versions := make([]*script.ScriptVersion, 0, len(rows))
	for _, row := range rows {
		version, err := row.toEntity()
		if err != nil {
			return nil, 0, err
		}
		versions = append(versions, version)
	}

	return versions, total, nil
}

// GetByID retrieves a script by ID
func (r *scriptRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*script.Script, error) {
	var row scriptRow
	query := `
		SELECT * FROM scripting_scripts
		WHERE id = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("script not found")
		}
		return nil, fmt.Errorf("failed to get script: %w", err)
	}

	return row.toEntity()
}

// GetByName retrieves a script by name
func (r *scriptRepository) GetByName(ctx context.Context, name string) (*script.Script, error) {
	var row scriptRow
	query := `
		SELECT * FROM scripting_scripts
		WHERE name = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &row, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("script not found")
		}
		return nil, fmt.Errorf("failed to get script by name: %w", err)
	}

	return row.toEntity()
}

// Update updates an existing script
func (r *scriptRepository) Update(ctx context.Context, s *script.Script) error {
	row, err := toRow(s)
	if err != nil {
		return fmt.Errorf("failed to convert to row: %w", err)
	}

	query := `
		UPDATE scripting_scripts SET
			name = :name,
			description = :description,
			code = :code,
			version = :version,
			status = :status,
			script_type = :script_type,
			entity_type = :entity_type,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.NamedExec(ctx, query, row)
	if err != nil {
		return fmt.Errorf("failed to update script: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("script not found or already deleted")
	}

	return nil
}

// Delete soft deletes a script
func (r *scriptRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE scripting_scripts SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete script: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("script not found or already deleted")
	}

	return nil
}

// List retrieves scripts by status with pagination
func (r *scriptRepository) List(ctx context.Context, status script.ScriptStatus, limit, offset int) ([]*script.Script, int, error) {
	var rows []scriptRow

	query := `
		SELECT * FROM scripting_scripts
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, string(status), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list scripts: %w", err)
	}

	// Count total matching records
	var total int
	countQuery := `
		SELECT COUNT(*) FROM scripting_scripts
		WHERE status = $1 AND deleted_at IS NULL`

	if err := r.Get(ctx, &total, countQuery, string(status)); err != nil {
		return nil, 0, fmt.Errorf("failed to count scripts: %w", err)
	}

	// Convert rows to entities
	scripts := make([]*script.Script, 0, len(rows))
	for _, row := range rows {
		s, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row: %w", err)
		}
		scripts = append(scripts, s)
	}

	return scripts, total, nil
}

// ListAll retrieves all scripts with pagination
func (r *scriptRepository) ListAll(ctx context.Context, limit, offset int) ([]*script.Script, int, error) {
	var rows []scriptRow

	query := `
		SELECT * FROM scripting_scripts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	if err := r.Select(ctx, &rows, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to list all scripts: %w", err)
	}

	// Count total records
	var total int
	countQuery := `SELECT COUNT(*) FROM scripting_scripts WHERE deleted_at IS NULL`

	if err := r.Get(ctx, &total, countQuery); err != nil {
		return nil, 0, fmt.Errorf("failed to count all scripts: %w", err)
	}

	// Convert rows to entities
	scripts := make([]*script.Script, 0, len(rows))
	for _, row := range rows {
		s, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row: %w", err)
		}
		scripts = append(scripts, s)
	}

	return scripts, total, nil
}

// CreateExecution inserts a new script execution record
func (r *scriptRepository) CreateExecution(ctx context.Context, e *script.ScriptExecution) error {
	row, err := toExecutionRow(e)
	if err != nil {
		return fmt.Errorf("failed to convert to row: %w", err)
	}

	query := `
		INSERT INTO scripting_script_executions (
			id, script_id, script_name, input_params, output_result,
			error, duration_ms, executed_by, executed_at
		) VALUES (
			:id, :script_id, :script_name, :input_params, :output_result,
			:error, :duration_ms, :executed_by, :executed_at
		)`

	if _, err := r.NamedExec(ctx, query, row); err != nil {
		return fmt.Errorf("failed to create execution: %w", err)
	}

	return nil
}

// GetExecutionByID retrieves a script execution by ID
func (r *scriptRepository) GetExecutionByID(ctx context.Context, id uuidv7.UUID) (*script.ScriptExecution, error) {
	var row executionRow
	query := `SELECT * FROM scripting_script_executions WHERE id = $1`

	if err := r.Get(ctx, &row, query, id.String()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("execution not found")
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return row.toEntity()
}

// GetExecutionHistory retrieves execution history for a script with pagination
func (r *scriptRepository) GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*script.ScriptExecution, int, error) {
	var rows []executionRow

	query := `
		SELECT * FROM scripting_script_executions
		WHERE script_id = $1
		ORDER BY executed_at DESC
		LIMIT $2 OFFSET $3`

	if err := r.Select(ctx, &rows, query, scriptID.String(), limit, offset); err != nil {
		return nil, 0, fmt.Errorf("failed to get execution history: %w", err)
	}

	// Count total executions for this script
	var total int
	countQuery := `SELECT COUNT(*) FROM scripting_script_executions WHERE script_id = $1`

	if err := r.Get(ctx, &total, countQuery, scriptID.String()); err != nil {
		return nil, 0, fmt.Errorf("failed to count executions: %w", err)
	}

	// Convert rows to entities
	executions := make([]*script.ScriptExecution, 0, len(rows))
	for _, row := range rows {
		e, err := row.toEntity()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert row: %w", err)
		}
		executions = append(executions, e)
	}

	return executions, total, nil
}

// GetRecentExecutions retrieves most recent executions across all scripts
func (r *scriptRepository) GetRecentExecutions(ctx context.Context, limit int) ([]*script.ScriptExecution, error) {
	var rows []executionRow

	query := `
		SELECT * FROM scripting_script_executions
		ORDER BY executed_at DESC
		LIMIT $1`

	if err := r.Select(ctx, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", err)
	}

	// Convert rows to entities
	executions := make([]*script.ScriptExecution, 0, len(rows))
	for _, row := range rows {
		e, err := row.toEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert row: %w", err)
		}
		executions = append(executions, e)
	}

	return executions, nil
}
