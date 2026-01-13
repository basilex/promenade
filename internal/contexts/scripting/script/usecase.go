package script

import (
	"context"
	"log/slog"
	"time"

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

// IScriptUseCase defines business operations for script management
type IScriptUseCase interface {
	// ExecuteScript executes a script by name with parameters
	ExecuteScript(ctx context.Context, scriptName string, params map[string]interface{}, executedBy uuidv7.UUID) (interface{}, error)

	// ValidateScript validates LUA script syntax
	ValidateScript(ctx context.Context, code string) error

	// CreateScript creates a new script
	CreateScript(ctx context.Context, name, description, code string, createdBy uuidv7.UUID) (*Script, error)

	// UpdateScript updates script code (increments version)
	UpdateScript(ctx context.Context, scriptID uuidv7.UUID, code string) error

	// UpdateScriptMetadata updates script metadata
	UpdateScriptMetadata(ctx context.Context, scriptID uuidv7.UUID, key string, value string) error

	// DeleteScript soft-deletes a script
	DeleteScript(ctx context.Context, scriptID uuidv7.UUID) error

	// GetScript retrieves a script by ID
	GetScript(ctx context.Context, scriptID uuidv7.UUID) (*Script, error)

	// GetScriptByName retrieves a script by name
	GetScriptByName(ctx context.Context, name string) (*Script, error)

	// ListScripts retrieves scripts with filtering
	ListScripts(ctx context.Context, status ScriptStatus, limit, offset int) ([]*Script, int, error)

	// ListAllScripts retrieves all scripts with pagination
	ListAllScripts(ctx context.Context, limit, offset int) ([]*Script, int, error)

	// ActivateScript changes status to active
	ActivateScript(ctx context.Context, scriptID uuidv7.UUID) error

	// DeactivateScript changes status to inactive
	DeactivateScript(ctx context.Context, scriptID uuidv7.UUID) error

	// ArchiveScript archives a script
	ArchiveScript(ctx context.Context, scriptID uuidv7.UUID) error

	// GetExecutionHistory retrieves execution history for a script
	GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*ScriptExecution, int, error)

	// GetRecentExecutions retrieves recent executions across all scripts
	GetRecentExecutions(ctx context.Context, limit int) ([]*ScriptExecution, error)

	// GetExecutionDetails retrieves details of a single execution
	GetExecutionDetails(ctx context.Context, executionID uuidv7.UUID) (*ScriptExecution, error)
}

// useCase implements IScriptUseCase
type useCase struct {
	repo   IRepository
	engine IScriptEngine
	logger *slog.Logger
}

// NewUseCase creates a new script use case
func NewUseCase(repo IRepository, engine IScriptEngine) IScriptUseCase {
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
		return nil, ErrScriptNotFound
	}

	// Check if script is active
	if script.Status != ScriptStatusActive {
		return nil, ErrScriptNotActive
	}

	// Create execution record
	execution := NewScriptExecution(script.GetID(), script.Name, executedBy)
	execution.SetInput(params)

	startTime := time.Now()

	// Execute script with engine
	result, err := uc.engine.Execute(ctx, script.Code, params)
	duration := time.Since(startTime)
	durationMs := int(duration.Milliseconds())

	// Update execution record with result or error
	if err != nil {
		execution.SetError(err, durationMs)
		log.Error("Script execution failed",
			slog.String("script_name", scriptName),
			slog.String("script_id", script.GetID().String()),
			slog.Int("duration_ms", durationMs),
			slog.Any("error", err),
		)
	} else {
		execution.SetResult(result, durationMs)
		log.Info("Script executed successfully",
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
		return nil, ErrScriptExecutionFailed
	}

	return result, nil
}

// ValidateScript validates LUA script syntax
func (uc *useCase) ValidateScript(ctx context.Context, code string) error {
	if code == "" {
		return ErrScriptCodeEmpty
	}

	// Use engine to validate syntax
	if err := uc.engine.Validate(code); err != nil {
		return ErrScriptSyntaxInvalid
	}

	return nil
}

// CreateScript creates a new script
func (uc *useCase) CreateScript(ctx context.Context, name, description, code string, createdBy uuidv7.UUID) (*Script, error) {
	log := logger.FromContext(ctx)

	// Validate script code
	if err := uc.ValidateScript(ctx, code); err != nil {
		return nil, ErrScriptSyntaxInvalid
	}

	// Create script entity (default to custom type)
	script, err := NewScript(name, code, ScriptTypeCustom)
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
		return nil, ErrScriptCreateFailed
	}

	log.Info("Script created successfully",
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
		return ErrScriptSyntaxInvalid
	}

	// Load existing script
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return ErrScriptNotFound
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
		return ErrScriptUpdateFailed
	}

	log.Info("Script updated successfully",
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
		return ErrScriptNotFound
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
		return ErrScriptUpdateFailed
	}

	log.Info("Script metadata updated",
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
		return ErrScriptDeleteFailed
	}

	log.Info("Script deleted successfully",
		slog.String("script_id", scriptID.String()),
	)

	return nil
}

// GetScript retrieves a script by ID
func (uc *useCase) GetScript(ctx context.Context, scriptID uuidv7.UUID) (*Script, error) {
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return nil, ErrScriptNotFound
	}
	return script, nil
}

// GetScriptByName retrieves a script by name
func (uc *useCase) GetScriptByName(ctx context.Context, name string) (*Script, error) {
	script, err := uc.repo.GetByName(ctx, name)
	if err != nil {
		return nil, ErrScriptNotFound
	}
	return script, nil
}

// ListScripts retrieves scripts with filtering
func (uc *useCase) ListScripts(ctx context.Context, status ScriptStatus, limit, offset int) ([]*Script, int, error) {
	scripts, total, err := uc.repo.List(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, ErrScriptListFailed
	}
	return scripts, total, nil
}

// ListAllScripts retrieves all scripts with pagination
func (uc *useCase) ListAllScripts(ctx context.Context, limit, offset int) ([]*Script, int, error) {
	scripts, total, err := uc.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, ErrScriptListFailed
	}
	return scripts, total, nil
}

// ActivateScript changes status to active
func (uc *useCase) ActivateScript(ctx context.Context, scriptID uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	// Load script
	script, err := uc.repo.GetByID(ctx, scriptID)
	if err != nil {
		return ErrScriptNotFound
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
		return ErrScriptUpdateFailed
	}

	log.Info("Script activated",
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
		return ErrScriptNotFound
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
		return ErrScriptUpdateFailed
	}

	log.Info("Script deactivated",
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
		return ErrScriptNotFound
	}

	// Archive
	script.Archive()

	// Save changes
	if err := uc.repo.Update(ctx, script); err != nil {
		log.Error("Failed to archive script",
			slog.String("script_id", scriptID.String()),
			slog.Any("error", err),
		)
		return ErrScriptUpdateFailed
	}

	log.Info("Script archived",
		slog.String("script_id", scriptID.String()),
	)

	return nil
}

// GetExecutionHistory retrieves execution history for a script
func (uc *useCase) GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*ScriptExecution, int, error) {
	executions, total, err := uc.repo.GetExecutionHistory(ctx, scriptID, limit, offset)
	if err != nil {
		return nil, 0, ErrScriptExecutionNotFound
	}
	return executions, total, nil
}

// GetRecentExecutions retrieves recent executions across all scripts
func (uc *useCase) GetRecentExecutions(ctx context.Context, limit int) ([]*ScriptExecution, error) {
	executions, err := uc.repo.GetRecentExecutions(ctx, limit)
	if err != nil {
		return nil, ErrScriptExecutionNotFound
	}
	return executions, nil
}

// GetExecutionDetails retrieves details of a single execution
func (uc *useCase) GetExecutionDetails(ctx context.Context, executionID uuidv7.UUID) (*ScriptExecution, error) {
	execution, err := uc.repo.GetExecutionByID(ctx, executionID)
	if err != nil {
		return nil, ErrScriptExecutionNotFound
	}
	return execution, nil
}
