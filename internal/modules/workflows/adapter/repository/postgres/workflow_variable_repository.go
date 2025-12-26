package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/internal/modules/workflows/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowVariableRepository implements IWorkflowVariableRepository
type WorkflowVariableRepository struct {
	*BaseRepository
}

// NewWorkflowVariableRepository creates a new workflow variable repository
func NewWorkflowVariableRepository(db *sqlx.DB) repository.IWorkflowVariableRepository {
	return &WorkflowVariableRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new workflow variable
func (r *WorkflowVariableRepository) Create(ctx context.Context, variable *entity.WorkflowVariable) error {
	query := `
		INSERT INTO workflows_variables (
			id, instance_id, name, value, type, scope, state_name, set_by, set_at, created_at, updated_at
		) VALUES (
			:id, :instance_id, :name, :value, :type, :scope, :state_name, :set_by, :set_at, :created_at, :updated_at
		)
	`
	_, err := r.NamedExec(ctx, query, variable)
	return err
}

// Update updates an existing workflow variable
func (r *WorkflowVariableRepository) Update(ctx context.Context, variable *entity.WorkflowVariable) error {
	query := `
		UPDATE workflows_variables SET
			value = :value,
			type = :type,
			scope = :scope,
			state_name = :state_name,
			set_by = :set_by,
			set_at = :set_at,
			updated_at = :updated_at
		WHERE id = :id
	`
	result, err := r.NamedExec(ctx, query, variable)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("workflow variable not found")
	}

	return nil
}

// Delete deletes a workflow variable (hard delete)
func (r *WorkflowVariableRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		DELETE FROM workflows_variables
		WHERE id = $1
	`
	result, err := r.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("workflow variable not found")
	}

	return nil
}

// GetByID retrieves a workflow variable by ID
func (r *WorkflowVariableRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowVariable, error) {
	var variable entity.WorkflowVariable
	query := `
		SELECT * FROM workflows_variables
		WHERE id = $1
	`
	err := r.Get(ctx, &variable, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow variable not found")
		}
		return nil, err
	}
	return &variable, nil
}

// GetByName retrieves a workflow variable by instance ID and name
func (r *WorkflowVariableRepository) GetByName(ctx context.Context, instanceID uuidv7.UUID, name string) (*entity.WorkflowVariable, error) {
	var variable entity.WorkflowVariable
	query := `
		SELECT * FROM workflows_variables
		WHERE instance_id = $1 AND name = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.Get(ctx, &variable, query, instanceID, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow variable not found")
		}
		return nil, err
	}
	return &variable, nil
}

// ListByInstance lists all variables for a specific instance
func (r *WorkflowVariableRepository) ListByInstance(ctx context.Context, instanceID uuidv7.UUID) ([]*entity.WorkflowVariable, error) {
	var variables []*entity.WorkflowVariable
	query := `
		SELECT * FROM workflows_variables
		WHERE instance_id = $1
		ORDER BY name ASC, created_at DESC
	`
	err := r.Select(ctx, &variables, query, instanceID)
	return variables, err
}

// ListByScope lists variables by scope
func (r *WorkflowVariableRepository) ListByScope(ctx context.Context, instanceID uuidv7.UUID, scope entity.WorkflowVariableScope) ([]*entity.WorkflowVariable, error) {
	var variables []*entity.WorkflowVariable
	query := `
		SELECT * FROM workflows_variables
		WHERE instance_id = $1 AND scope = $2
		ORDER BY name ASC
	`
	err := r.Select(ctx, &variables, query, instanceID, scope)
	return variables, err
}

// ListByState lists variables for a specific state
func (r *WorkflowVariableRepository) ListByState(ctx context.Context, instanceID uuidv7.UUID, state string) ([]*entity.WorkflowVariable, error) {
	var variables []*entity.WorkflowVariable
	query := `
		SELECT * FROM workflows_variables
		WHERE instance_id = $1 AND (state = $2 OR scope = 'global')
		ORDER BY name ASC
	`
	err := r.Select(ctx, &variables, query, instanceID, state)
	return variables, err
}

// DeleteByInstance deletes all variables for a specific instance
func (r *WorkflowVariableRepository) DeleteByInstance(ctx context.Context, instanceID uuidv7.UUID) error {
	query := `
		DELETE FROM workflows_variables
		WHERE instance_id = $1
	`
	_, err := r.Exec(ctx, query, instanceID)
	return err
}

// DeleteByScope deletes all variables of a specific scope
func (r *WorkflowVariableRepository) DeleteByScope(ctx context.Context, instanceID uuidv7.UUID, scope entity.WorkflowVariableScope) error {
	query := `
		DELETE FROM workflows_variables
		WHERE instance_id = $1 AND scope = $2
	`
	_, err := r.Exec(ctx, query, instanceID, scope)
	return err
}

// GetInstanceContext retrieves all global and state-specific variables as a context map
func (r *WorkflowVariableRepository) GetInstanceContext(ctx context.Context, instanceID uuidv7.UUID, currentState string) (map[string]interface{}, error) {
	var variables []*entity.WorkflowVariable
	query := `
		SELECT * FROM workflows_variables
		WHERE instance_id = $1 
			AND (scope = 'global' OR (scope = 'state' AND state = $2))
		ORDER BY name ASC
	`
	err := r.Select(ctx, &variables, query, instanceID, currentState)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, v := range variables {
		result[v.Name] = v.Value
	}
	
	return result, nil
}

// UpsertVariable creates or updates a variable (upsert operation)
func (r *WorkflowVariableRepository) UpsertVariable(ctx context.Context, variable *entity.WorkflowVariable) error {
	query := `
		INSERT INTO workflows_variables (
			id, instance_id, name, value, scope, state, is_encrypted, created_at, updated_at
		) VALUES (
			:id, :instance_id, :name, :value, :scope, :state, :is_encrypted, :created_at, :updated_at
		)
		ON CONFLICT (instance_id, name, scope, COALESCE(state, ''))
		DO UPDATE SET
			value = EXCLUDED.value,
			is_encrypted = EXCLUDED.is_encrypted,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.NamedExec(ctx, query, variable)
	return err
}

// CountByInstance counts variables for a specific instance
func (r *WorkflowVariableRepository) CountByInstance(ctx context.Context, instanceID uuidv7.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_variables
		WHERE instance_id = $1
	`
	err := r.Get(ctx, &count, query, instanceID)
	return count, err
}
