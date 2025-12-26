package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
	"github.com/basilex/promenade/internal/modules/workflows/domain/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// WorkflowInstanceRepository implements IWorkflowInstanceRepository
type WorkflowInstanceRepository struct {
	*BaseRepository
}

// NewWorkflowInstanceRepository creates a new workflow instance repository
func NewWorkflowInstanceRepository(db *sqlx.DB) repository.IWorkflowInstanceRepository {
	return &WorkflowInstanceRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new workflow instance
func (r *WorkflowInstanceRepository) Create(ctx context.Context, instance *entity.WorkflowInstance) error {
	query := `
		INSERT INTO workflows_instances (
			id, definition_id, definition_version, external_reference, status, current_state, 
			previous_state, context, input, output, error_message, error_details,
			priority, parent_instance_id, started_at, started_by, assigned_to, due_date,
			tags, retry_count, state_entered_at, state_timeout_at, created_at, updated_at
		) VALUES (
			:id, :definition_id, :definition_version, :external_reference, :status, :current_state,
			:previous_state, :context, :input, :output, :error_message, :error_details,
			:priority, :parent_instance_id, :started_at, :started_by, :assigned_to, :due_date,
			:tags, :retry_count, :state_entered_at, :state_timeout_at, :created_at, :updated_at
		)
	`
	params := map[string]interface{}{
		"id":                  instance.ID,
		"definition_id":       instance.DefinitionID,
		"definition_version":  instance.DefinitionVersion,
		"external_reference":  instance.ExternalReference,
		"status":              instance.Status,
		"current_state":       instance.CurrentState,
		"previous_state":      instance.PreviousState,
		"context":             instance.Context,
		"input":               instance.Input,
		"output":              instance.Output,
		"error_message":       instance.ErrorMessage,
		"error_details":       instance.ErrorDetails,
		"priority":            instance.Priority,
		"parent_instance_id":  instance.ParentInstanceID,
		"started_at":          instance.StartedAt,
		"started_by":          instance.StartedBy,
		"assigned_to":         instance.AssignedTo,
		"due_date":            instance.DueDate,
		"tags":                pq.StringArray(instance.Tags),
		"retry_count":         instance.RetryCount,
		"state_entered_at":    instance.StateEnteredAt,
		"state_timeout_at":    instance.StateTimeoutAt,
		"created_at":          instance.CreatedAt,
		"updated_at":          instance.UpdatedAt,
	}
	_, err := r.NamedExec(ctx, query, params)
	return err
}

// Update updates an existing workflow instance
func (r *WorkflowInstanceRepository) Update(ctx context.Context, instance *entity.WorkflowInstance) error {
	query := `
		UPDATE workflows_instances SET
			status = :status,
			current_state = :current_state,
			previous_state = :previous_state,
			context = :context,
			output = :output,
			error_message = :error_message,
			error_details = :error_details,
			priority = :priority,
			assigned_to = :assigned_to,
			due_date = :due_date,
			retry_count = :retry_count,
			started_at = :started_at,
			completed_at = :completed_at,
			state_entered_at = :state_entered_at,
			state_timeout_at = :state_timeout_at,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL
	`
	
	params := map[string]interface{}{
		"id":               instance.ID,
		"status":           instance.Status,
		"current_state":    instance.CurrentState,
		"previous_state":   instance.PreviousState,
		"context":          instance.Context,
		"output":           instance.Output,
		"error_message":    instance.ErrorMessage,
		"error_details":    instance.ErrorDetails,
		"priority":         instance.Priority,
		"assigned_to":      instance.AssignedTo,
		"due_date":         instance.DueDate,
		"retry_count":      instance.RetryCount,
		"started_at":       instance.StartedAt,
		"completed_at":     instance.CompletedAt,
		"state_entered_at": instance.StateEnteredAt,
		"state_timeout_at": instance.StateTimeoutAt,
		"updated_at":       instance.UpdatedAt,
	}
	
	result, err := r.NamedExec(ctx, query, params)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("workflow instance not found or already deleted")
	}

	return nil
}

// Delete soft-deletes a workflow instance
func (r *WorkflowInstanceRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE workflows_instances
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL
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
		return fmt.Errorf("workflow instance not found or already deleted")
	}

	return nil
}

// GetByID retrieves a workflow instance by ID
func (r *WorkflowInstanceRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowInstance, error) {
	query := `
		SELECT 
			id, definition_id, definition_version, external_reference, status, 
			current_state, previous_state, context, input, output, error_message, error_details,
			priority, parent_instance_id, started_at, started_by, assigned_to, due_date,
			tags, retry_count, state_entered_at, state_timeout_at, created_at, updated_at, deleted_at
		FROM workflows_instances
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	var instance entity.WorkflowInstance
	var tags pq.StringArray
	
	err := r.getExecutor(ctx).QueryRowxContext(ctx, query, id).Scan(
		&instance.ID, &instance.DefinitionID, &instance.DefinitionVersion, &instance.ExternalReference, &instance.Status,
		&instance.CurrentState, &instance.PreviousState, &instance.Context, &instance.Input, &instance.Output, &instance.ErrorMessage, &instance.ErrorDetails,
		&instance.Priority, &instance.ParentInstanceID, &instance.StartedAt, &instance.StartedBy, &instance.AssignedTo, &instance.DueDate,
		&tags, &instance.RetryCount, &instance.StateEnteredAt, &instance.StateTimeoutAt, &instance.CreatedAt, &instance.UpdatedAt, &instance.DeletedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow instance not found")
		}
		return nil, err
	}
	
	instance.Tags = []string(tags)
	return &instance, nil
}

// GetByExternalReference retrieves a workflow instance by external reference
func (r *WorkflowInstanceRepository) GetByExternalReference(ctx context.Context, externalRef string) (*entity.WorkflowInstance, error) {
	var instance entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE external_reference = $1 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &instance, query, externalRef)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow instance not found")
		}
		return nil, err
	}
	return &instance, nil
}

// ListByDefinition lists workflow instances by definition ID
func (r *WorkflowInstanceRepository) ListByDefinition(ctx context.Context, definitionID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE definition_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &instances, query, definitionID, limit, offset)
	return instances, err
}

// ListByStatus lists workflow instances by status
func (r *WorkflowInstanceRepository) ListByStatus(ctx context.Context, status entity.WorkflowInstanceStatus, limit, offset int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &instances, query, status, limit, offset)
	return instances, err
}

// ListByAssignee lists workflow instances by assignee
func (r *WorkflowInstanceRepository) ListByAssignee(ctx context.Context, assigneeID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE assigned_to = $1 AND deleted_at IS NULL
		ORDER BY priority DESC, created_at ASC
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &instances, query, assigneeID, limit, offset)
	return instances, err
}

// ListByCreator lists workflow instances by creator
func (r *WorkflowInstanceRepository) ListByCreator(ctx context.Context, creatorID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE started_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &instances, query, creatorID, limit, offset)
	return instances, err
}

// ListOverdue lists workflow instances that are overdue
func (r *WorkflowInstanceRepository) ListOverdue(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE due_at < CURRENT_TIMESTAMP
			AND status NOT IN ('completed', 'failed', 'cancelled', 'timed_out')
			AND deleted_at IS NULL
		ORDER BY due_at ASC
		LIMIT $1 OFFSET $2
	`
	err := r.Select(ctx, &instances, query, limit, offset)
	return instances, err
}

// ListTimedOut lists workflow instances that have timed out
func (r *WorkflowInstanceRepository) ListTimedOut(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE timeout_at < CURRENT_TIMESTAMP
			AND status NOT IN ('completed', 'failed', 'cancelled', 'timed_out')
			AND deleted_at IS NULL
		ORDER BY timeout_at ASC
		LIMIT $1 OFFSET $2
	`
	err := r.Select(ctx, &instances, query, limit, offset)
	return instances, err
}

// ListActive lists all active workflow instances
func (r *WorkflowInstanceRepository) ListActive(ctx context.Context, limit, offset int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE status IN ('pending', 'running', 'waiting')
			AND deleted_at IS NULL
		ORDER BY priority DESC, created_at ASC
		LIMIT $1 OFFSET $2
	`
	err := r.Select(ctx, &instances, query, limit, offset)
	return instances, err
}

// CountByStatus counts workflow instances by status
func (r *WorkflowInstanceRepository) CountByStatus(ctx context.Context, status entity.WorkflowInstanceStatus) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_instances
		WHERE status = $1 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &count, query, status)
	return count, err
}

// CountByDefinition counts workflow instances by definition
func (r *WorkflowInstanceRepository) CountByDefinition(ctx context.Context, definitionID uuidv7.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_instances
		WHERE definition_id = $1 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &count, query, definitionID)
	return count, err
}

// CountActiveByDefinition counts active workflow instances (pending, running, waiting) for a definition
func (r *WorkflowInstanceRepository) CountActiveByDefinition(ctx context.Context, definitionID uuidv7.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_instances
		WHERE definition_id = $1
			AND status IN ('pending', 'running', 'waiting')
			AND deleted_at IS NULL
	`
	err := r.Get(ctx, &count, query, definitionID)
	return count, err
}

// CountByAssignee counts workflow instances by assignee
func (r *WorkflowInstanceRepository) CountByAssignee(ctx context.Context, assigneeID uuidv7.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_instances
		WHERE assigned_to = $1 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &count, query, assigneeID)
	return count, err
}

// CountActive counts all active workflow instances
func (r *WorkflowInstanceRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_instances
		WHERE status IN ('pending', 'running', 'waiting')
			AND deleted_at IS NULL
	`
	err := r.Get(ctx, &count, query)
	return count, err
}

// GetInstancesForProcessing retrieves instances ready for processing (internal helper)
func (r *WorkflowInstanceRepository) GetInstancesForProcessing(ctx context.Context, limit int) ([]*entity.WorkflowInstance, error) {
	var instances []*entity.WorkflowInstance
	query := `
		SELECT * FROM workflows_instances
		WHERE status = 'pending'
			AND deleted_at IS NULL
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	err := r.Select(ctx, &instances, query, limit)
	return instances, err
}

// UpdateStatusBatch updates status for multiple instances (internal helper for batch operations)
func (r *WorkflowInstanceRepository) UpdateStatusBatch(ctx context.Context, ids []uuidv7.UUID, status entity.WorkflowInstanceStatus, errorMsg *string) error {
	if len(ids) == 0 {
		return nil
	}

	query := `
		UPDATE workflows_instances
		SET status = $1,
			error_message = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ANY($3) AND deleted_at IS NULL
	`
	
	// Convert UUIDs to string array for PostgreSQL
	strIDs := make([]string, len(ids))
	for i, id := range ids {
		strIDs[i] = id.String()
	}
	
	_, err := r.Exec(ctx, query, status, errorMsg, strIDs)
	return err
}
