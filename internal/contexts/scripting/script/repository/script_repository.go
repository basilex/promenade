package repository

import (
	"context"

	"github.com/basilex/promenade/internal/contexts/scripting/script/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IRepository defines the interface for script repository
type IRepository interface {
	// aggregate.Script CRUD operations
	Create(ctx context.Context, script *aggregate.Script) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Script, error)
	GetByName(ctx context.Context, name string) (*aggregate.Script, error)
	Update(ctx context.Context, script *aggregate.Script) error
	Delete(ctx context.Context, id uuidv7.UUID) error

	// aggregate.Script versioning
	CreateVersion(ctx context.Context, version *aggregate.ScriptVersion) error
	ListVersions(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptVersion, int, error)

	// aggregate.Script queries
	List(ctx context.Context, status aggregate.ScriptStatus, limit, offset int) ([]*aggregate.Script, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]*aggregate.Script, int, error)

	// aggregate.Script execution tracking
	CreateExecution(ctx context.Context, execution *aggregate.ScriptExecution) error
	GetExecutionByID(ctx context.Context, id uuidv7.UUID) (*aggregate.ScriptExecution, error)
	GetExecutionHistory(ctx context.Context, scriptID uuidv7.UUID, limit, offset int) ([]*aggregate.ScriptExecution, int, error)
	GetRecentExecutions(ctx context.Context, limit int) ([]*aggregate.ScriptExecution, error)
}
