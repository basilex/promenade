package script

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for script repository
type IRepository interface {
	// Script CRUD operations
	Create(ctx context.Context, script *Script) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*Script, error)
	GetByName(ctx context.Context, name string) (*Script, error)
	Update(ctx context.Context, script *Script) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// Script versioning
	CreateVersion(ctx context.Context, version *ScriptVersion) error
	ListVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*ScriptVersion, int, error)

	// Script queries
	List(ctx context.Context, status ScriptStatus, limit, offset int) ([]*Script, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]*Script, int, error)

	// Script execution tracking
	CreateExecution(ctx context.Context, execution *ScriptExecution) error
	GetExecutionByID(ctx context.Context, id uuidv7.UUID) (*ScriptExecution, error)
	GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*ScriptExecution, int, error)
	GetRecentExecutions(ctx context.Context, limit int) ([]*ScriptExecution, error)
}
