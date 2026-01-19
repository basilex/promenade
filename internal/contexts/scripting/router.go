package scripting

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	scripthttp "github.com/basilex/promenade/internal/contexts/scripting/script/adapter/http"
	scriptRepo "github.com/basilex/promenade/internal/contexts/scripting/script/adapter/repository/postgres"
	scriptUseCase "github.com/basilex/promenade/internal/contexts/scripting/script/usecase"
	"github.com/basilex/promenade/pkg/scripting"
)

// Router handles routing for the Scripting context
type Router struct {
	scriptHandler *scripthttp.ScriptHandler
}

// NewRouter creates a new scripting router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Script aggregate
	scriptRepository := scriptRepo.NewScriptRepository(db)

	// Initialize LUA engine with default config and standard library
	config := scripting.DefaultConfig()
	stdlib := scripting.NewStandardLibrary(context.TODO(), nil, nil, nil, nil) // TODO: Pass real dependencies
	engine := scripting.NewEngine(config, stdlib)

	// Initialize use case with repository and engine
	useCase := scriptUseCase.NewScriptUseCase(scriptRepository, engine)
	scriptHandler := scripthttp.NewScriptHandler(useCase)

	return &Router{
		scriptHandler: scriptHandler,
	}
}

// RegisterRoutes registers all scripting context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	scripting := api.Group("/scripting")
	{
		// Script validation endpoint (no ID required)
		scripting.POST("/scripts/validate", r.scriptHandler.ValidateScript)

		// Script CRUD endpoints
		scripts := scripting.Group("/scripts")
		{
			scripts.POST("", r.scriptHandler.CreateScript)
			scripts.GET("", r.scriptHandler.ListScripts)
			scripts.GET("/:id", r.scriptHandler.GetScript)
			scripts.PUT("/:id", r.scriptHandler.UpdateScript)
			scripts.DELETE("/:id", r.scriptHandler.DeleteScript)

			// Script by name
			scripts.GET("/name/:name", r.scriptHandler.GetScriptByName)

			// Script execution
			scripts.POST("/name/:name/execute", r.scriptHandler.ExecuteScript)

			// Script status management
			scripts.POST("/:id/activate", r.scriptHandler.ActivateScript)
			scripts.POST("/:id/deactivate", r.scriptHandler.DeactivateScript)
			scripts.POST("/:id/archive", r.scriptHandler.ArchiveScript)

			// Execution history
			scripts.GET("/:id/executions", r.scriptHandler.GetExecutionHistory)

			// Version history
			scripts.GET("/:id/versions", r.scriptHandler.ListScriptVersions)
		}

		// Recent executions endpoint (outside of scripts group)
		executions := scripting.Group("/scripts/executions")
		{
			executions.GET("/recent", r.scriptHandler.GetRecentExecutions)
			executions.GET("/:execution_id", r.scriptHandler.GetExecutionDetails)
		}
	}
}
