package analytics

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/basilex/promenade/internal/modules/analytics/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/analytics/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/analytics/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/module"
	moduleconfig "github.com/basilex/promenade/pkg/module/config"
	"github.com/gin-gonic/gin"
)

// AnalyticsModule implements the module.IModule interface
type AnalyticsModule struct {
	core           *module.Core
	config         *Config
	logger         *slog.Logger
	metricsHandler *handler.MetricsHandler
}

// Config holds analytics module configuration
type Config struct {
	MetricsRetention  int `yaml:"metrics_retention_days"`
	MaxReportsPerUser int `yaml:"max_reports_per_user"`
	MaxDashboards     int `yaml:"max_dashboards_per_user"`
}

// New creates a new analytics module instance
func New() module.IModule {
	return &AnalyticsModule{
		logger: slog.Default().With("module", "analytics"),
	}
}

// Metadata returns module information
func (m *AnalyticsModule) Metadata() module.Metadata {
	return module.Metadata{
		Name:        "analytics",
		Version:     "1.0.0",
		Description: "Analytics and reporting (Free)",
	}
}

// Dependencies returns list of module names this module depends on
func (m *AnalyticsModule) Dependencies() []string {
	// Analytics doesn't depend on other modules
	return nil
}

// Initialize initializes the module with configuration
func (m *AnalyticsModule) Initialize(ctx context.Context, core *module.Core) error {
	m.logger.Info("Initializing analytics module", "version", m.Metadata().Version)
	m.core = core

	// Load module-specific configuration
	cfg, err := moduleconfig.Load("internal/modules/analytics/config", "promenade")
	if err != nil {
		m.logger.Error("Failed to load analytics config", "error", err)
		return fmt.Errorf("failed to load analytics config: %w", err)
	}

	{
		// Parse config into our Config struct (all values required)
		var metricsRetention, maxReportsPerUser, maxDashboards int

		if analytics, ok := cfg.Settings["analytics"].(map[string]any); ok {
			if metrics, ok := analytics["metrics"].(map[string]any); ok {
				if v, ok := metrics["retention_days"].(int); ok {
					metricsRetention = v
				}
			}
			if reports, ok := analytics["reports"].(map[string]any); ok {
				if v, ok := reports["max_per_user"].(int); ok {
					maxReportsPerUser = v
				}
			}
			if dashboards, ok := analytics["dashboards"].(map[string]any); ok {
				if v, ok := dashboards["max_per_user"].(int); ok {
					maxDashboards = v
				}
			}
		}

		m.config = &Config{
			MetricsRetention:  metricsRetention,
			MaxReportsPerUser: maxReportsPerUser,
			MaxDashboards:     maxDashboards,
		}

		// Validate required config values
		if metricsRetention == 0 || maxReportsPerUser == 0 || maxDashboards == 0 {
			return fmt.Errorf("analytics config incomplete - check config/analytics/config.*.yaml")
		}

		m.logger.Info("Analytics config loaded",
			"metrics_retention", metricsRetention,
			"max_reports_per_user", maxReportsPerUser,
			"max_dashboards", maxDashboards,
		)
	}

	// Initialize repositories
	metricRepo := postgres.NewMetricRepository(core.DB)

	// Initialize use cases
	analyticsUC := usecase.NewAnalyticsUseCase(metricRepo, m.logger)

	// Initialize handlers
	m.metricsHandler = handler.NewMetricsHandler(analyticsUC, m.logger)

	m.logger.Info("Analytics module initialized successfully")
	return nil
}

// Start starts the module
func (m *AnalyticsModule) Start(ctx context.Context) error {
	m.logger.Info("Starting analytics module")

	// TODO: Start background workers (metric aggregation, scheduled reports)

	return nil
}

// Stop stops the module
func (m *AnalyticsModule) Stop(ctx context.Context) error {
	m.logger.Info("Stopping analytics module")

	// TODO: Stop background workers gracefully

	return nil
}

// RegisterRoutes registers HTTP routes
func (m *AnalyticsModule) RegisterRoutes(router *gin.RouterGroup) {
	m.logger.Info("Registering analytics routes")

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"module":  m.Metadata().Name,
			"version": m.Metadata().Version,
		})
	})

	// Metrics endpoints
	router.POST("/metrics", m.metricsHandler.CollectMetric)
	router.GET("/metrics", m.metricsHandler.ListMetrics)
	router.GET("/metrics/latest", m.metricsHandler.GetLatestMetric)
	router.GET("/metrics/:id", m.metricsHandler.GetMetric)

	// TODO: Add reports endpoints
	// router.POST("/reports", handler.CreateReport)
	// router.GET("/reports", handler.ListReports)

	// TODO: Add dashboards endpoints
	// router.POST("/dashboards", handler.CreateDashboard)
	// router.GET("/dashboards", handler.ListDashboards)
}

// RegisterMigrations returns database migrations
func (m *AnalyticsModule) RegisterMigrations() []module.Migration {
	// Migrations are in migrations/analytics/ directory
	// They will be auto-discovered by the migration system
	return nil
}

// RegisterEventHandlers subscribes to domain events
func (m *AnalyticsModule) RegisterEventHandlers(eventBus bus.IBus) error {
	m.logger.Info("Registering analytics event handlers")

	// TODO: Subscribe to user events
	// eventBus.Subscribe(ctx, bus.TopicUserRegistered, func(event *bus.Event) {
	//     // Track user registration metric
	// })

	// TODO: Subscribe to post events
	// eventBus.Subscribe(ctx, bus.TopicPostCreated, func(event *bus.Event) {
	//     // Track post creation metric
	// })

	return nil
}

// RegisterPermissions registers RBAC permissions
func (m *AnalyticsModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "analytics:metrics", Action: "view", Description: "View analytics metrics"},
		{Resource: "analytics:metrics", Action: "create", Description: "Create analytics metrics"},
		{Resource: "analytics:reports", Action: "view", Description: "View analytics reports"},
		{Resource: "analytics:reports", Action: "create", Description: "Create analytics reports"},
		{Resource: "analytics:reports", Action: "export", Description: "Export analytics reports"},
		{Resource: "analytics:dashboards", Action: "view", Description: "View dashboards"},
		{Resource: "analytics:dashboards", Action: "create", Description: "Create dashboards"},
		{Resource: "analytics:dashboards", Action: "manage", Description: "Manage dashboards"},
	}
}

// HealthCheck returns the module's health status
func (m *AnalyticsModule) HealthCheck(ctx context.Context) error {
	// Check database connectivity by pinging
	if err := m.core.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Note: Background workers health check will be added when workers are implemented

	return nil
}
