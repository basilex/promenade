package usecase

import (
	"context"
	"log/slog"
	"time"

	scripterrors "github.com/basilex/promenade/internal/contexts/scripting/script"
	"github.com/basilex/promenade/internal/contexts/scripting/script/aggregate"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IScriptEngine defines the interface for LUA script execution
type IScriptEngine interface {
	// Execute executes LUA code with parameters and returns result
	Execute(ctx context.Context, code string, params map[string]interface{}) (interface{}, error)

	// Validate validates LUA code syntax
	Validate(code string) error
}

// IScriptRepository defines the repository interface for script persistence
type IScriptRepository interface {
	// Script CRUD operations
	Create(ctx context.Context, script *aggregate.Script) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Script, error)
	GetByName(ctx context.Context, name string) (*aggregate.Script, error)
	Update(ctx context.Context, script *aggregate.Script) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Script versioning
	CreateVersion(ctx context.Context, version *aggregate.ScriptVersion) error
	ListVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptVersion, int, error)

	// Script queries
	List(ctx context.Context, status aggregate.ScriptStatus, limit, offset int) ([]*aggregate.Script, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]*aggregate.Script, int, error)

	// Script execution tracking
	CreateExecution(ctx context.Context, execution *aggregate.ScriptExecution) error
	GetExecutionByID(ctx context.Context, id uuidv7.UUID) (*aggregate.ScriptExecution, error)
	GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptExecution, int, error)
	GetRecentExecutions(ctx context.Context, limit int) ([]*aggregate.ScriptExecution, error)
}

// IScriptUseCase defines business operations for script management
type IScriptUseCase interface {
	// ExecuteScript executes a script by name with parameters
	ExecuteScript(ctx context.Context, scriptName string, params map[string]interface{}, executedBy uuidv7.UUID) (interface{}, error)

	// ValidateScript validates LUA script syntax
	ValidateScript(ctx context.Context, code string) error

	// CreateScript creates a new script
	CreateScript(ctx context.Context, name, description, code string, createdBy uuidv7.UUID) (*aggregate.Script, error)

	// UpdateScript updates script code (increments version)
	UpdateScript(ctx context.Context, scriptID uuidv7.UUID, code string) error

	// UpdateScriptMetadata updates script metadata
	UpdateScriptMetadata(ctx context.Context, scriptID uuidv7.UUID, key string, value string) error

	// DeleteScript soft-deletes a script
	DeleteScript(ctx context.Context, scriptID uuidv7.UUID) error

	// GetScript retrieves a script by ID
	GetScript(ctx context.Context, scriptID uuidv7.UUID) (*aggregate.Script, error)

	// GetScriptByName retrieves a script by name
	GetScriptByName(ctx context.Context, name string) (*aggregate.Script, error)

	// ListScripts retrieves scripts with filtering
	ListScripts(ctx context.Context, status aggregate.ScriptStatus, limit, offset int) ([]*aggregate.Script, int, error)

	// ListAllScripts retrieves all scripts with pagination
	ListAllScripts(ctx context.Context, limit, offset int) ([]*aggregate.Script, int, error)

	// ActivateScript changes status to active
	ActivateScript(ctx context.Context, scriptID uuidv7.UUID) error

	// DeactivateScript changes status to inactive
	DeactivateScript(ctx context.Context, scriptID uuidv7.UUID) error

	// ArchiveScript archives a script
	ArchiveScript(ctx context.Context, scriptID uuidv7.UUID) error

	// GetExecutionHistory retrieves execution history for a script
	GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptExecution, int, error)

	// GetRecentExecutions retrieves recent executions across all scripts
	GetRecentExecutions(ctx context.Context, limit int) ([]*aggregate.ScriptExecution, error)

	// GetExecutionDetails retrieves details of a single execution
	GetExecutionDetails(ctx context.Context, executionID uuidv7.UUID) (*aggregate.ScriptExecution, error)

	// ListScriptVersions retrieves versions for a script
	ListScriptVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptVersion, int, error)
}

// useCase implements IScriptUseCase
type useCase struct {
	repo   IScriptRepository
	engine IScriptEngine
	logger *slog.Logger
}

// NewScriptUseCase creates a new script use case
func NewScriptUseCase(repo IScriptRepository, engine IScriptEngine) IScriptUseCase {
	return &useCase{
		repo:   repo,
		engine: engine,
		logger: slog.Default(),
	}
}

// ExecuteScript executes a script by name with parameters
func (uc *useCase) ExecuteScript(ctx context.Context, scriptName string, params map[string]interface{}, executedBy uuidv7.UUID) (interface{}, error) {
	log := logger.FromContext(ctx)

	// Load script from repository
	script, err := uc.repo.GetByName(ctx, scriptName)
	if err != nil {
		log.Error("Failed to load script",
			slog.String("script_name", scriptName),
			slog.Any("error", err),
		)
		return nil, scripterrors.ErrScriptNotFound
	}

	// Check if script is active
	if script.Status != aggregate.ScriptStatusActive {
		return nil, scripterrors.ErrScriptNotActive
	}

	// Create execution record
	execution := aggregate.NewScriptExecution(script.GetID(), script.Name, executedBy)
	execution.SetInput(params)

	startTime := time.Now()

	// Execute script with engine
	result, err := uc.engine.Execute(ctx, script.Code, params)
	duration := time.Since(startTime)
	durationMs := int(duration.Milliseconds())

	// Update execution record with result or error
	if err != nil {
		execution.SetError(err, durationMs)
		log.Error("aggregate.Script execution failed",
			slog.String("script_name", scriptName),
			slog.String("script_id", script.GetID().String()),
			slog.Int("duration_ms", durationMs),
			slog.Any("error", err),
		)
	} else {
		execution.SetResult(result, durationMs)
		log.Info("aggregate.Script executed successfully",
			slog.String("script_name", scriptName),
			slog.String("script_id", script.GetID().String()),
			slog.Int("duration_ms", durationMs),
		)
	}

	// Save execution record (log errors but don't fail)
	if saveErr := uc.repo.CreateExecution(ctx, execution); saveErr != nil {
		log.Error("Failed to save execution record",
			slog.String("execution_id", execution.ID.String()),
			slog.Any("error", saveErr),
		)
	}

	// Return execution result or error
	if err != nil {
		return nil, scripterrors.ErrScriptExecutionFailed
	}

	return result, nil
}

// ValidateScript validates LUA script syntax
func (uc *useCase) ValidateScript(ctx context.Context, code string) error {
	if code == "" {
		return scripterrors.ErrScriptCodeEmpty
	}

	// Use engine to validate syntax
	if err := uc.engine.Validate(code); err != nil {
		return scripterrors.ErrScriptSyntaxInvalid
	}

	return nil
}

// CreateScript creates a new script
func (uc *useCase) CreateScript(ctx context.Context, name, description, code string, createdBy uuidv7.UUID) (*aggregate.Script, error) {
	log := logger.FromContext(ctx)

	// Validate script code
	if err := uc.ValidateScript(ctx, code); err != nil {
		return nil, scripterrors.ErrScriptSyntaxInvalid
	}

	// Create script entity (default to custom type)
	script, err := aggregate.NewScript(name, code, aggregate.ScriptTypeCustom)
	if err != nil {
		return nil, err // Domain error from entity
	}

	script.Description = description
	script.UpdateMetadata("created_by", createdBy.String())

	// Save to repository
	if err := uc.repo.Create(ctx, script); err != nil {
		log.Error("Failed to create script",
			slog.String("script_name", name),
			slog.Any("error", err),
		)
		return nil, scripterrors.ErrScriptCreateFailed
	}

	// Store version snapshot (log only on failure)
	version, err := aggregate.NewScriptVersion(script, "initial version", &createdBy)
	if err == nil {
		if saveErr := uc.repo.CreateVersion(ctx, version); saveErr != nil {
			log.Error("Failed to create script version snapshot",
				slog.String("script_id", script.GetID().String()),
				slog.Any("error", saveErr),
			)
		}
	}

	log.Info("aggregate.Script created successfully",
		slog.String("script_id", script.GetID().String()),
		slog.String("script_name", name),
		slog.String("created_by", createdBy.String()),
	)

	return script, nil
}

// UpdateScript updates script code (increments version)
func (uc *useCase) UpdateScript(ctx context.Context, scriptID uuidv7.UUID, code string) error {
	log := logger.FromContext(ctx)

	// Validate new code
	if err := uc.ValidateScript(ctx, code); err != nil {
		return scripterrors.ErrScriptSyntaxInvalid
	}

	// Load existing script
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return scripterrors.ErrScriptNotFound
	}

	// Update code (increments version)
	if err := script.UpdateCode(code); err != nil {
		return err // Domain error from entity
	}

	// Save changes
	if err := uc.repo.Update(ctx, script); err != nil {
		log.Error("Failed to update script",
			slog.String("script_id", scriptID.String()),
			slog.Any("error", err),
		)
		return scripterrors.ErrScriptUpdateFailed
	}

	// Store version snapshot (log only on failure)
	version, err := aggregate.NewScriptVersion(script, "code update", nil)
	if err == nil {
		if saveErr := uc.repo.CreateVersion(ctx, version); saveErr != nil {
			log.Error("Failed to create script version snapshot",
				slog.String("script_id", scriptID.String()),
				slog.Any("error", saveErr),
			)
		}
	}

	log.Info("aggregate.Script updated successfully",
		slog.String("script_id", scriptID.String()),
		slog.Int("new_version", script.Version),
	)

	return nil
}

// UpdateScriptMetadata updates script metadata
func (uc *useCase) UpdateScriptMetadata(ctx context.Context, scriptID uuidv7.UUID, key string, value string) error {
	log := logger.FromContext(ctx)

	// Load script
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return scripterrors.ErrScriptNotFound
	}

	// Update metadata
	script.UpdateMetadata(key, value)

	// Save changes
	if err := uc.repo.Update(ctx, script); err != nil {
		log.Error("Failed to update metadata",
			slog.String("script_id", scriptID.String()),
			slog.String("key", key),
			slog.Any("error", err),
		)
		return scripterrors.ErrScriptUpdateFailed
	}

	log.Info("aggregate.Script metadata updated",
		slog.String("script_id", scriptID.String()),
		slog.String("key", key),
	)

	return nil
}

// DeleteScript soft-deletes a script
func (uc *useCase) DeleteScript(ctx context.Context, scriptID uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	if err := uc.repo.Delete(ctx, scriptID); err != nil {
		log.Error("Failed to delete script",
			slog.String("script_id", scriptID.String()),
			slog.Any("error", err),
		)
		return scripterrors.ErrScriptDeleteFailed
	}

	log.Info("aggregate.Script deleted successfully",
		slog.String("script_id", scriptID.String()),
	)

	return nil
}

// GetScript retrieves a script by ID
func (uc *useCase) GetScript(ctx context.Context, scriptID uuidv7.UUID) (*aggregate.Script, error) {
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return nil, scripterrors.ErrScriptNotFound
	}
	return script, nil
}

// GetScriptByName retrieves a script by name
func (uc *useCase) GetScriptByName(ctx context.Context, name string) (*aggregate.Script, error) {
	script, err := uc.repo.GetByName(ctx, name)
	if err != nil {
		return nil, scripterrors.ErrScriptNotFound
	}
	return script, nil
}

// ListScripts retrieves scripts with filtering
func (uc *useCase) ListScripts(ctx context.Context, status aggregate.ScriptStatus, limit, offset int) ([]*aggregate.Script, int, error) {
	scripts, total, err := uc.repo.List(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, scripterrors.ErrScriptListFailed
	}
	return scripts, total, nil
}

// ListAllScripts retrieves all scripts with pagination
func (uc *useCase) ListAllScripts(ctx context.Context, limit, offset int) ([]*aggregate.Script, int, error) {
	scripts, total, err := uc.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, scripterrors.ErrScriptListFailed
	}
	return scripts, total, nil
}

// ActivateScript changes status to active
func (uc *useCase) ActivateScript(ctx context.Context, scriptID uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	// Load script
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return scripterrors.ErrScriptNotFound
	}

	// Activate
	if err := script.Activate(); err != nil {
		return err // Domain error from entity
	}

	// Save changes
	if err := uc.repo.Update(ctx, script); err != nil {
		log.Error("Failed to activate script",
			slog.String("script_id", scriptID.String()),
			slog.Any("error", err),
		)
		return scripterrors.ErrScriptUpdateFailed
	}

	log.Info("aggregate.Script activated",
		slog.String("script_id", scriptID.String()),
	)

	return nil
}

// DeactivateScript changes status to inactive
func (uc *useCase) DeactivateScript(ctx context.Context, scriptID uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	// Load script
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return scripterrors.ErrScriptNotFound
	}

	// Deactivate
	if err := script.Deactivate(); err != nil {
		return err // Domain error from entity
	}

	// Save changes
	if err := uc.repo.Update(ctx, script); err != nil {
		log.Error("Failed to deactivate script",
			slog.String("script_id", scriptID.String()),
			slog.Any("error", err),
		)
		return scripterrors.ErrScriptUpdateFailed
	}

	log.Info("aggregate.Script deactivated",
		slog.String("script_id", scriptID.String()),
	)

	return nil
}

// ArchiveScript archives a script
func (uc *useCase) ArchiveScript(ctx context.Context, scriptID uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	// Load script
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return scripterrors.ErrScriptNotFound
	}

	// Archive
	script.Archive()

	// Save changes
	if err := uc.repo.Update(ctx, script); err != nil {
		log.Error("Failed to archive script",
			slog.String("script_id", scriptID.String()),
			slog.Any("error", err),
		)
		return scripterrors.ErrScriptUpdateFailed
	}

	log.Info("aggregate.Script archived",
		slog.String("script_id", scriptID.String()),
	)

	return nil
}

// GetExecutionHistory retrieves execution history for a script
func (uc *useCase) GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptExecution, int, error) {
	executions, total, err := uc.repo.GetExecutionHistory(ctx, scriptID, limit, offset)
	if err != nil {
		return nil, 0, scripterrors.ErrScriptExecutionNotFound
	}
	return executions, total, nil
}

// GetRecentExecutions retrieves recent executions across all scripts
func (uc *useCase) GetRecentExecutions(ctx context.Context, limit int) ([]*aggregate.ScriptExecution, error) {
	executions, err := uc.repo.GetRecentExecutions(ctx, limit)
	if err != nil {
		return nil, scripterrors.ErrScriptExecutionNotFound
	}
	return executions, nil
}

// GetExecutionDetails retrieves details of a single execution
func (uc *useCase) GetExecutionDetails(ctx context.Context, executionID uuidv7.UUID) (*aggregate.ScriptExecution, error) {
	execution, err := uc.repo.GetExecutionByID(ctx, executionID)
	if err != nil {
		return nil, scripterrors.ErrScriptExecutionNotFound
	}
	return execution, nil
}

// ListScriptVersions retrieves versions for a script
func (uc *useCase) ListScriptVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptVersion, int, error) {
	if _, err := uc.repo.GetByID(ctx, scriptID); err != nil {
		return nil, 0, scripterrors.ErrScriptNotFound
	}

	versions, total, err := uc.repo.ListVersions(ctx, scriptID, limit, offset)
	if err != nil {
		return nil, 0, scripterrors.ErrScriptListFailed
	}

	return versions, total, nil
}
