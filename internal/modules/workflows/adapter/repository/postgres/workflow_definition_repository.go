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

// WorkflowDefinitionRepository implements IWorkflowDefinitionRepository
type WorkflowDefinitionRepository struct {
	*BaseRepository
}

// NewWorkflowDefinitionRepository creates a new workflow definition repository
func NewWorkflowDefinitionRepository(db *sqlx.DB) repository.IWorkflowDefinitionRepository {
	return &WorkflowDefinitionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new workflow definition
func (r *WorkflowDefinitionRepository) Create(ctx context.Context, definition *entity.WorkflowDefinition) error {
	query := `
		INSERT INTO workflows_definitions (
			id, name, display_name, description, version, status, category, tags,
			definition, input_schema, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13
		)
	`

	_, err := r.Exec(ctx, query,
		definition.ID, definition.Name, definition.DisplayName, definition.Description,
		definition.Version, definition.Status, definition.Category, definition.Tags,
		definition.Definition, definition.InputSchema,
		definition.CreatedBy, definition.CreatedAt, definition.UpdatedAt,
	)
	return err
}

// Update updates an existing workflow definition
func (r *WorkflowDefinitionRepository) Update(ctx context.Context, definition *entity.WorkflowDefinition) error {
	query := `
		UPDATE workflows_definitions SET
			display_name = $1,
			description = $2,
			status = $3,
			category = $4,
			tags = $5,
			definition = $6,
			input_schema = $7,
			updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	result, err := r.Exec(ctx, query,
		definition.DisplayName, definition.Description, definition.Status,
		definition.Category, definition.Tags, definition.Definition,
		definition.InputSchema, definition.UpdatedAt, definition.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("workflow definition not found or already deleted")
	}

	return nil
}

// Delete soft-deletes a workflow definition
func (r *WorkflowDefinitionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	query := `
		UPDATE workflows_definitions
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
		return fmt.Errorf("workflow definition not found or already deleted")
	}

	return nil
}

// GetByID retrieves a workflow definition by ID
func (r *WorkflowDefinitionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.WorkflowDefinition, error) {
	var definition entity.WorkflowDefinition
	query := `
		SELECT * FROM workflows_definitions
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &definition, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow definition not found")
		}
		return nil, err
	}

	return &definition, nil
}

// GetByName retrieves a workflow definition by name (latest active version)
func (r *WorkflowDefinitionRepository) GetByName(ctx context.Context, name string) (*entity.WorkflowDefinition, error) {
	var definition entity.WorkflowDefinition
	query := `
		SELECT * FROM workflows_definitions
		WHERE name = $1 AND status = 'active' AND deleted_at IS NULL
		ORDER BY version DESC
		LIMIT 1
	`
	err := r.Get(ctx, &definition, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow definition not found")
		}
		return nil, err
	}

	return &definition, nil
}

// GetByNameAndVersion retrieves a specific version of a workflow definition
func (r *WorkflowDefinitionRepository) GetByNameAndVersion(ctx context.Context, name string, version int) (*entity.WorkflowDefinition, error) {
	var definition entity.WorkflowDefinition
	query := `
		SELECT * FROM workflows_definitions
		WHERE name = $1 AND version = $2 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &definition, query, name, version)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow definition not found")
		}
		return nil, err
	}

	return &definition, nil
}

// ListByStatus lists workflow definitions by status
func (r *WorkflowDefinitionRepository) ListByStatus(ctx context.Context, status entity.WorkflowDefinitionStatus, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	var definitions []*entity.WorkflowDefinition
	query := `
		SELECT * FROM workflows_definitions
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	if err := r.Select(ctx, &definitions, query, status, limit, offset); err != nil {
		return nil, err
	}
	return definitions, nil
}

// ListByCategory lists workflow definitions by category
func (r *WorkflowDefinitionRepository) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	var definitions []*entity.WorkflowDefinition
	query := `
		SELECT * FROM workflows_definitions
		WHERE category = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	if err := r.Select(ctx, &definitions, query, category, limit, offset); err != nil {
		return nil, err
	}
	return definitions, nil
}

// ListByCreator lists workflow definitions by creator
func (r *WorkflowDefinitionRepository) ListByCreator(ctx context.Context, creatorID uuidv7.UUID, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	var definitions []*entity.WorkflowDefinition
	query := `
		SELECT * FROM workflows_definitions
		WHERE created_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	if err := r.Select(ctx, &definitions, query, creatorID, limit, offset); err != nil {
		return nil, err
	}
	return definitions, nil
}

// ListVersions lists all versions of a workflow definition
func (r *WorkflowDefinitionRepository) ListVersions(ctx context.Context, name string) ([]*entity.WorkflowDefinition, error) {
	var definitions []*entity.WorkflowDefinition
	query := `
		SELECT * FROM workflows_definitions
		WHERE name = $1 AND deleted_at IS NULL
		ORDER BY version DESC
	`
	if err := r.Select(ctx, &definitions, query, name); err != nil {
		return nil, err
	}
	return definitions, nil
}

// CountByStatus counts workflow definitions by status
func (r *WorkflowDefinitionRepository) CountByStatus(ctx context.Context, status entity.WorkflowDefinitionStatus) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_definitions
		WHERE status = $1 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &count, query, status)
	return count, err
}

// CountByCategory counts workflow definitions by category
func (r *WorkflowDefinitionRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) FROM workflows_definitions
		WHERE category = $1 AND deleted_at IS NULL
	`
	err := r.Get(ctx, &count, query, category)
	return count, err
}

// Search searches workflow definitions by name or description
func (r *WorkflowDefinitionRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entity.WorkflowDefinition, error) {
	var definitions []*entity.WorkflowDefinition
	searchQuery := `
		SELECT * FROM workflows_definitions
		WHERE (name ILIKE $1 OR display_name ILIKE $1 OR description ILIKE $1)
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	searchPattern := "%" + query + "%"
	if err := r.Select(ctx, &definitions, searchQuery, searchPattern, limit, offset); err != nil {
		return nil, err
	}
	return definitions, nil
}
