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

// WorkflowEventRepository implements IWorkflowEventRepository
type WorkflowEventRepository struct {
	*BaseRepository
}

// NewWorkflowEventRepository creates a new workflow event repository
func NewWorkflowEventRepository(db *sqlx.DB) repository.IWorkflowEventRepository {
	return &WorkflowEventRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new workflow event
func (r *WorkflowEventRepository) Create(ctx context.Context, event *entity.WorkflowEvent) error {
	query := `
		INSERT INTO workflows_events (
			id, instance_id, type, name, payload, triggered_by, source,
			processed, processed_at, created_at
		) VALUES (
			:id, :instance_id, :type, :name, :payload, :triggered_by, :source,
			:processed, :processed_at, :created_at
		)
	`
	_, err := r.NamedExec(ctx, query, event)
	return err
}

// Update updates an existing workflow event
func (r *WorkflowEventRepository) Update(ctx context.Context, event *entity.WorkflowEvent) error {
	query := `
		UPDATE workflows_events SET
			processed = :processed,
			processed_at = :processed_at
		WHERE id = :id
	`
	result, err := r.NamedExec(ctx, query, event)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("workflow event not found")
	}

	return nil
}

// GetByID retrieves a workflow event by ID
func (r *WorkflowEventRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowEvent, error) {
	var event entity.WorkflowEvent
	query := `
		SELECT * FROM workflows_events
		WHERE id = $1
	`
	err := r.Get(ctx, &event, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow event not found")
		}
		return nil, err
	}
	return &event, nil
}

// ListByInstance lists all events for a specific instance
func (r *WorkflowEventRepository) ListByInstance(ctx context.Context, instanceID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowEvent, error) {
	var events []*entity.WorkflowEvent
	query := `
		SELECT * FROM workflows_events
		WHERE instance_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &events, query, instanceID, limit, offset)
	return events, err
}

// ListUnprocessed lists all unprocessed events (for event processing worker)
func (r *WorkflowEventRepository) ListUnprocessed(ctx context.Context, limit int) ([]*entity.WorkflowEvent, error) {
	var events []*entity.WorkflowEvent
	query := `
		SELECT * FROM workflows_events
		WHERE processed = FALSE
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	err := r.Select(ctx, &events, query, limit)
	return events, err
}

// ListByType lists events by type
func (r *WorkflowEventRepository) ListByType(ctx context.Context, eventType entity.WorkflowEventType, limit, offset int) ([]*entity.WorkflowEvent, error) {
	var events []*entity.WorkflowEvent
	query := `
		SELECT * FROM workflows_events
		WHERE type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &events, query, eventType, limit, offset)
	return events, err
}

// CountByInstance counts events for a specific instance
func (r *WorkflowEventRepository) CountByInstance(ctx context.Context, instanceID uuidv7.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_events
		WHERE instance_id = $1
	`
	err := r.Get(ctx, &count, query, instanceID)
	return count, err
}

// CountUnprocessed counts unprocessed events
func (r *WorkflowEventRepository) CountUnprocessed(ctx context.Context) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_events
		WHERE processed = FALSE
	`
	err := r.Get(ctx, &count, query)
	return count, err
}

// GetPendingScheduledEvents retrieves scheduled events that are ready to process
func (r *WorkflowEventRepository) GetPendingScheduledEvents(ctx context.Context, limit int) ([]*entity.WorkflowEvent, error) {
	var events []*entity.WorkflowEvent
	query := `
		SELECT * FROM workflows_events
		WHERE status = 'pending'
			AND scheduled_at IS NOT NULL
			AND scheduled_at <= CURRENT_TIMESTAMP
		ORDER BY scheduled_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	err := r.Select(ctx, &events, query, limit)
	return events, err
}

// MarkAsProcessed marks an event as processed
func (r *WorkflowEventRepository) MarkAsProcessed(ctx context.Context, eventID uuidv7.UUID) error {
	query := `
		UPDATE workflows_events
		SET status = 'processed',
			processed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.Exec(ctx, query, eventID)
	return err
}

// MarkAsFailed marks an event as failed
func (r *WorkflowEventRepository) MarkAsFailed(ctx context.Context, eventID uuidv7.UUID, errorMsg string) error {
	query := `
		UPDATE workflows_events
		SET status = 'failed',
			error_message = $2,
			processed_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.Exec(ctx, query, eventID, errorMsg)
	return err
}

// DeleteOldProcessedEvents deletes processed events older than specified days (cleanup)
func (r *WorkflowEventRepository) DeleteOldProcessedEvents(ctx context.Context, daysOld int) (int64, error) {
	query := `
		DELETE FROM workflows_events
		WHERE status = 'processed'
			AND processed_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * $1
	`
	result, err := r.Exec(ctx, query, daysOld)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetEventsByInstanceAndType retrieves events by instance and type
func (r *WorkflowEventRepository) GetEventsByInstanceAndType(ctx context.Context, instanceID uuidv7.UUID, eventType entity.WorkflowEventType) ([]*entity.WorkflowEvent, error) {
	var events []*entity.WorkflowEvent
	query := `
		SELECT * FROM workflows_events
		WHERE instance_id = $1 AND event_type = $2
		ORDER BY created_at DESC
	`
	err := r.Select(ctx, &events, query, instanceID, eventType)
	return events, err
}
