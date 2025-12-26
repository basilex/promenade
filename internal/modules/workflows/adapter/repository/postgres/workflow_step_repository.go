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

// WorkflowStepRepository implements IWorkflowStepRepository
type WorkflowStepRepository struct {
	*BaseRepository
}

// NewWorkflowStepRepository creates a new workflow step repository
func NewWorkflowStepRepository(db *sqlx.DB) repository.IWorkflowStepRepository {
	return &WorkflowStepRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new workflow step
func (r *WorkflowStepRepository) Create(ctx context.Context, step *entity.WorkflowStep) error {
	query := `
		INSERT INTO workflows_steps (
			id, instance_id, definition_id, step_number, type, status, state_name, 
			activity_name, event, input, output, error_message, error_details,
			retry_count, duration, executed_by, started_at, completed_at, created_at, updated_at
		) VALUES (
			:id, :instance_id, :definition_id, :step_number, :type, :status, :state_name,
			:activity_name, :event, :input, :output, :error_message, :error_details,
			:retry_count, :duration, :executed_by, :started_at, :completed_at, :created_at, :updated_at
		)
	`
	_, err := r.NamedExec(ctx, query, step)
	return err
}

// Update updates an existing workflow step
func (r *WorkflowStepRepository) Update(ctx context.Context, step *entity.WorkflowStep) error {
	query := `
		UPDATE workflows_steps SET
			status = :status,
			output = :output,
			error_message = :error_message,
			error_details = :error_details,
			retry_count = :retry_count,
			duration = :duration,
			completed_at = :completed_at,
			updated_at = :updated_at
		WHERE id = :id
	`
	result, err := r.NamedExec(ctx, query, step)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("workflow step not found")
	}

	return nil
}

// GetByID retrieves a workflow step by ID
func (r *WorkflowStepRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowStep, error) {
	var step entity.WorkflowStep
	query := `
		SELECT * FROM workflows_steps
		WHERE id = $1
	`
	err := r.Get(ctx, &step, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow step not found")
		}
		return nil, err
	}
	return &step, nil
}

// ListByInstance lists workflow steps for a specific instance
func (r *WorkflowStepRepository) ListByInstance(ctx context.Context, instanceID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowStep, error) {
	var steps []*entity.WorkflowStep
	query := `
		SELECT * FROM workflows_steps
		WHERE instance_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	err := r.Select(ctx, &steps, query, instanceID, limit, offset)
	return steps, err
}

// ListByInstanceAndState lists workflow steps for a specific instance and state
func (r *WorkflowStepRepository) ListByInstanceAndState(ctx context.Context, instanceID uuidv7.UUID, state string) ([]*entity.WorkflowStep, error) {
	var steps []*entity.WorkflowStep
	query := `
		SELECT * FROM workflows_steps
		WHERE instance_id = $1 AND state = $2
		ORDER BY created_at ASC
	`
	err := r.Select(ctx, &steps, query, instanceID, state)
	return steps, err
}

// GetLatestByInstance retrieves the latest step for an instance
func (r *WorkflowStepRepository) GetLatestByInstance(ctx context.Context, instanceID uuidv7.UUID) (*entity.WorkflowStep, error) {
	var step entity.WorkflowStep
	query := `
		SELECT * FROM workflows_steps
		WHERE instance_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := r.Get(ctx, &step, query, instanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No steps is not an error
		}
		return nil, err
	}
	return &step, nil
}

// CountByInstance counts workflow steps for a specific instance
func (r *WorkflowStepRepository) CountByInstance(ctx context.Context, instanceID uuidv7.UUID) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_steps
		WHERE instance_id = $1
	`
	err := r.Get(ctx, &count, query, instanceID)
	return count, err
}

// CountByStatus counts workflow steps by status for a specific instance
func (r *WorkflowStepRepository) CountByStatus(ctx context.Context, instanceID uuidv7.UUID, status entity.WorkflowStepStatus) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_steps
		WHERE instance_id = $1 AND status = $2
	`
	err := r.Get(ctx, &count, query, instanceID, status)
	return count, err
}

// GetAverageDuration calculates average duration for steps of a specific type in a definition
func (r *WorkflowStepRepository) GetAverageDuration(ctx context.Context, definitionID uuidv7.UUID, stepType entity.WorkflowStepType) (int64, error) {
	var avgDuration sql.NullInt64
	query := `
		SELECT AVG(duration_ms) FROM workflows_steps
		WHERE definition_id = $1 AND step_type = $2 AND status = 'completed' AND duration_ms IS NOT NULL
	`
	err := r.Get(ctx, &avgDuration, query, definitionID, stepType)
	if err != nil {
		return 0, err
	}
	if !avgDuration.Valid {
		return 0, nil
	}
	return avgDuration.Int64, nil
}

// GetStepExecutionHistory retrieves execution history for a specific step name (analytics)
func (r *WorkflowStepRepository) GetStepExecutionHistory(ctx context.Context, stepName string, limit int) ([]*entity.WorkflowStep, error) {
	var steps []*entity.WorkflowStep
	query := `
		SELECT * FROM workflows_steps
		WHERE step_name = $1 AND status = 'completed'
		ORDER BY created_at DESC
		LIMIT $2
	`
	err := r.Select(ctx, &steps, query, stepName, limit)
	return steps, err
}

// GetFailedSteps retrieves all failed steps for monitoring
func (r *WorkflowStepRepository) GetFailedSteps(ctx context.Context, limit, offset int) ([]*entity.WorkflowStep, error) {
	var steps []*entity.WorkflowStep
	query := `
		SELECT * FROM workflows_steps
		WHERE status = 'failed'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := r.Select(ctx, &steps, query, limit, offset)
	return steps, err
}

// GetStepDurationStats retrieves duration statistics for a specific step name
func (r *WorkflowStepRepository) GetStepDurationStats(ctx context.Context, stepName string) (*entity.StepDurationStats, error) {
	type stats struct {
		MinDuration sql.NullInt64 `db:"min_duration"`
		MaxDuration sql.NullInt64 `db:"max_duration"`
		AvgDuration sql.NullInt64 `db:"avg_duration"`
		TotalSteps  int64         `db:"total_steps"`
	}
	
	var s stats
	query := `
		SELECT 
			MIN(duration_ms) as min_duration,
			MAX(duration_ms) as max_duration,
			AVG(duration_ms) as avg_duration,
			COUNT(*) as total_steps
		FROM workflows_steps
		WHERE step_name = $1 AND status = 'completed' AND duration_ms IS NOT NULL
	`
	err := r.Get(ctx, &s, query, stepName)
	if err != nil {
		return nil, err
	}
	
	result := &entity.StepDurationStats{
		StepName:   stepName,
		TotalSteps: s.TotalSteps,
	}
	
	if s.MinDuration.Valid {
		result.MinDuration = s.MinDuration.Int64
	}
	if s.MaxDuration.Valid {
		result.MaxDuration = s.MaxDuration.Int64
	}
	if s.AvgDuration.Valid {
		result.AvgDuration = s.AvgDuration.Int64
	}
	
	return result, nil
}
