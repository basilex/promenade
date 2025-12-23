package analytics

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/basilex/promenade/internal/modules/analytics/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/analytics/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/analytics/license"
	"github.com/basilex/promenade/internal/modules/analytics/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/module"
	"github.com/gin-gonic/gin"
)

// AnalyticsModule implements the module.Module interface
type AnalyticsModule struct {
	core           *module.Core
	config         *Config
	license        *license.License
	logger         *slog.Logger
	metricsHandler *handler.MetricsHandler
}

// Config holds analytics module configuration
type Config struct {
	LicenseRequired    bool   `yaml:"license_required"`
	LicenseKey         string `yaml:"license_key"`
	ValidateExpiry     bool   `yaml:"validate_expiry"`
	ValidateSignature  bool   `yaml:"validate_signature"`
	GracePeriodDays    int    `yaml:"grace_period_days"`
	ValidateOnRequest  bool   `yaml:"validate_on_request"`
	MetricsRetention   int    `yaml:"metrics_retention_days"`
	MaxReportsPerUser  int    `yaml:"max_reports_per_user"`
	MaxDashboards      int    `yaml:"max_dashboards_per_user"`
}

// New creates a new analytics module instance
func New() module.Module {
	return &AnalyticsModule{
		logger: slog.Default().With("module", "analytics"),
	}
}

// Metadata returns module information
func (m *AnalyticsModule) Metadata() module.Metadata {
	return module.Metadata{
		Name:        "analytics",
		Version:     "1.0.0",
		Description: "Advanced analytics and reporting (Commercial)",
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
	// TODO: Load from config/modules/analytics/config.yaml
	m.config = &Config{
		LicenseRequired:   false, // Set via configuration
		ValidateExpiry:    true,
		ValidateSignature: true,
		GracePeriodDays:   7,
		MetricsRetention:  90,
		MaxReportsPerUser: 10,
		MaxDashboards:     5,
	}

	// Validate license if required
	if m.config.LicenseRequired {
		if err := m.validateLicense(); err != nil {
			return fmt.Errorf("license validation failed: %w", err)
		}
		m.logger.Info("License validated successfully",
			"tier", m.license.Tier,
			"expiry", m.license.ExpiryDate.Format("2006-01-02"),
			"days_remaining", m.license.DaysUntilExpiry())
	} else {
		m.logger.Warn("License validation disabled (development mode)")
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

// validateLicense validates the module license
func (m *AnalyticsModule) validateLicense() error {
	// Get license key from config or environment
	licenseKey := m.config.LicenseKey
	if licenseKey == "" {
		licenseKey = os.Getenv("ANALYTICS_LICENSE_KEY")
	}

	if licenseKey == "" {
		return fmt.Errorf("license key is required but not provided")
	}

	// Get secret from environment
	secret := os.Getenv("LICENSE_SECRET")
	if secret == "" {
		secret = "default-dev-secret-change-in-production"
		m.logger.Warn("Using default license secret (not for production)")
	}

	// Parse license
	lic, err := license.Parse(licenseKey)
	if err != nil {
		return fmt.Errorf("failed to parse license: %w", err)
	}

	// Validate license
	err = lic.Validate(secret, m.Metadata().Name, m.config.GracePeriodDays)
	if err != nil {
		return fmt.Errorf("license validation failed: %w", err)
	}

	m.license = lic
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
		status := "healthy"
		licenseInfo := make(map[string]interface{})

		if m.license != nil {
			licenseInfo["tier"] = m.license.Tier
			licenseInfo["expiry"] = m.license.ExpiryDate.Format("2006-01-02")
			licenseInfo["days_remaining"] = m.license.DaysUntilExpiry()
			licenseInfo["expired"] = m.license.IsExpired()

			if m.license.IsExpired() && m.config.LicenseRequired {
				status = "degraded"
			}
		}

		c.JSON(200, gin.H{
			"status":  status,
			"module":  m.Metadata().Name,
			"version": m.Metadata().Version,
			"license": licenseInfo,
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
func (m *AnalyticsModule) RegisterEventHandlers(eventBus bus.Bus) error {
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
	// Check license status
	if m.license != nil {
		if m.license.IsExpired() {
			daysExpired := -m.license.DaysUntilExpiry()
			if daysExpired > m.config.GracePeriodDays {
				return fmt.Errorf("license expired %d days ago (grace period exceeded)", daysExpired)
			}
		}
	}

	// TODO: Check database connectivity
	// TODO: Check background workers status

	return nil
}
