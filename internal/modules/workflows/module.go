package workflows

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/modules/workflows/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/workflows/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/workflows/usecase"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/module"
	moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

// WorkflowsModule implements the workflows business module
type WorkflowsModule struct {
	*module.BaseModule
	db                *sqlx.DB
	config            *moduleconfig.Config // Module's own config
	authMiddleware    *middleware.AuthMiddleware
	definitionUseCase usecase.IWorkflowDefinitionUseCase
	instanceUseCase   usecase.IWorkflowInstanceUseCase
	definitionHandler *handler.WorkflowDefinitionHandler
	instanceHandler   *handler.WorkflowInstanceHandler
}

// NewModule creates a new workflows module
func NewModule() module.IModule {
	return &WorkflowsModule{
		BaseModule: module.NewBaseModule(module.Metadata{
			Name:        "workflows",
			DisplayName: "Workflow Management",
			Version:     "1.0.0",
			Author:      "Promenade Team",
			Description: "BPMN-inspired workflow engine with state machine and event sourcing",
			License:     "Commercial",
			Tags:        []string{"workflows", "bpmn", "state-machine", "orchestration"},
		}),
	}
}

// Initialize initializes the workflows module
func (m *WorkflowsModule) Initialize(ctx context.Context, core *module.Core) error {
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	slog.Info("Initializing workflows module")

	m.db = core.DB

	// Load module's own configuration
	cfg, err := moduleconfig.Load("internal/modules/workflows/config", "promenade")
	if err != nil {
		slog.Warn("Failed to load workflows module config, using fallback defaults", "error", err)
		// Use fallback configuration with safe defaults
		m.config = getDefaultWorkflowsConfig()
		slog.Info("Using fallback workflows configuration",
			"max_concurrent_instances", 100,
			"worker_pool_size", 4,
		)
	} else {
		m.config = cfg
		slog.Info("Workflows module config loaded",
			"version", cfg.Module.Version,
			"enabled", cfg.Module.Enabled,
		)
	}

	// Ensure config is not nil (should never happen with fallback, but be defensive)
	if m.config == nil {
		return fmt.Errorf("workflows config not loaded - cannot initialize module")
	}

	// Initialize repositories
	definitionRepo := postgres.NewWorkflowDefinitionRepository(m.db)
	instanceRepo := postgres.NewWorkflowInstanceRepository(m.db)
	// Initialize middleware
	m.authMiddleware = middleware.NewAuthMiddleware(core.JWT)
	// Initialize use cases
	m.definitionUseCase = usecase.NewWorkflowDefinitionUseCase(definitionRepo, instanceRepo)
	m.instanceUseCase = usecase.NewWorkflowInstanceUseCase(instanceRepo, definitionRepo)

	// Initialize handlers
	m.definitionHandler = handler.NewWorkflowDefinitionHandler(m.definitionUseCase)
	m.instanceHandler = handler.NewWorkflowInstanceHandler(m.instanceUseCase)

	// Note: Basic CRUD operations ready. Advanced features (execution engine,
	// license validation, purge policies) will be implemented in future iterations.

	return nil
}

// getDefaultWorkflowsConfig returns fallback configuration with safe defaults
func getDefaultWorkflowsConfig() *moduleconfig.Config {
	return &moduleconfig.Config{
		Module: moduleconfig.ModuleSection{
			Name:    "workflows",
			Version: "1.0.0",
			Enabled: true,
		},
		Settings: map[string]any{
			"execution": map[string]any{
				"max_concurrent_instances":  100,
				"step_timeout_default":      "5m",
				"instance_timeout_default":  "24h",
				"enable_parallel_execution": true,
				"worker_pool_size":          4,
			},
			"retry": map[string]any{
				"max_attempts":     3,
				"initial_interval": "1s",
				"multiplier":       2.0,
				"max_interval":     "1m",
			},
			"activities": map[string]any{
				"http_timeout":   "30s",
				"email_enabled":  false,
				"script_engine":  "javascript",
				"script_timeout": "10s",
			},
			"monitoring": map[string]any{
				"enable_metrics":      false,
				"enable_tracing":      false,
				"slow_step_threshold": "10s",
			},
			"license_key":      "",
			"license_required": false,
		},
		Purge: moduleconfig.PurgeSection{
			Enabled: true,
			Settings: map[string]map[string]any{
				"completed_instances": {"retention_days": 90},
				"failed_instances":    {"retention_days": 180},
				"steps":               {"retention_days": 90},
				"variables":           {"retention_days": 90},
				"events":              {"retention_days": 60},
			},
		},
	}
}

// RegisterRoutes registers HTTP routes for the workflows module
func (m *WorkflowsModule) RegisterRoutes(router *gin.RouterGroup) {
	// Health check (must be before parameterized routes)
	router.GET("/workflows/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"data": gin.H{
				"status":  "healthy",
				"module":  m.Metadata().Name,
				"version": m.Metadata().Version,
			},
		})
	})

	// Apply auth middleware to all workflows routes except health
	workflows := router.Group("/workflows")
	workflows.Use(m.authMiddleware.RequireAuth())
	{
		// Definitions
		workflows.GET("/definitions", m.definitionHandler.ListDefinitions)          // NEW
		workflows.POST("/definitions", m.definitionHandler.CreateDefinition)
		workflows.GET("/definitions/:id", m.definitionHandler.GetDefinition)
		workflows.PUT("/definitions/:id", m.definitionHandler.UpdateDefinition)     // NEW
		workflows.DELETE("/definitions/:id", m.definitionHandler.DeleteDefinition)  // NEW
		workflows.GET("/definitions/name/:name", m.definitionHandler.GetDefinitionByName)
		workflows.PUT("/definitions/:id/activate", m.definitionHandler.ActivateDefinition)

		// Instances
		workflows.POST("/instances", m.instanceHandler.StartWorkflow)
		workflows.GET("/instances", m.instanceHandler.ListInstances)
		workflows.GET("/instances/:id", m.instanceHandler.GetInstance)
		workflows.PUT("/instances/:id", m.instanceHandler.UpdateInstance)           // NEW
		workflows.PUT("/instances/:id/pause", m.instanceHandler.PauseInstance)
		workflows.PUT("/instances/:id/resume", m.instanceHandler.ResumeInstance)
		workflows.PUT("/instances/:id/cancel", m.instanceHandler.CancelInstance)
	}
}

// RegisterPermissions returns RBAC permissions for the workflows module
func (m *WorkflowsModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "workflows:definitions", Action: "read", Description: "View workflow definitions"},
		{Resource: "workflows:definitions", Action: "create", Description: "Create workflow definitions"},
		{Resource: "workflows:definitions", Action: "update", Description: "Update workflow definitions"},
		{Resource: "workflows:definitions", Action: "delete", Description: "Delete workflow definitions"},
		{Resource: "workflows:definitions", Action: "activate", Description: "Activate workflow definitions"},
		{Resource: "workflows:instances", Action: "read", Description: "View workflow instances"},
		{Resource: "workflows:instances", Action: "create", Description: "Start workflow instances"},
		{Resource: "workflows:instances", Action: "update", Description: "Update workflow instances"},
		{Resource: "workflows:instances", Action: "cancel", Description: "Cancel workflow instances"},
		{Resource: "workflows:tasks", Action: "read", Description: "View assigned tasks"},
		{Resource: "workflows:tasks", Action: "complete", Description: "Complete tasks"},
	}
}

// Start starts background workers for the workflows module
func (m *WorkflowsModule) Start(ctx context.Context) error {
	log := logger.FromContext(ctx)
	log.Info("Starting workflows module")
	// TODO: Start workflow execution engine
	// TODO: Start event processing worker
	// TODO: Start timeout checker
	log.Info("Workflows module started successfully")
	return nil
}

// Stop gracefully stops the workflows module
func (m *WorkflowsModule) Stop(ctx context.Context) error {
	// TODO: Stop background workers
	return nil
}

// HealthCheck checks the health of the workflows module
func (m *WorkflowsModule) HealthCheck(ctx context.Context) error {
	if err := m.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	return nil
}
