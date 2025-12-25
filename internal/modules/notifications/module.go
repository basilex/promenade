package notifications

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/modules/notifications/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/notifications/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/notifications/usecase"
	"github.com/basilex/promenade/pkg/license"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/module"
)

// Module implements the IModule interface for notifications
type Module struct {
	*module.BaseModule
	handler *handler.NotificationHandler
	useCase usecase.INotificationUseCase
}

// New creates a new notifications module
func New() module.IModule {
	return &Module{
		BaseModule: module.NewBaseModule(module.Metadata{
			Name:        "notifications",
			DisplayName: "Notifications",
			Version:     "1.0.0",
			Author:      "Promenade Team",
			Description: "Multi-channel notification system with user preferences",
			License:     "Commercial",
			Tags:        []string{"notifications", "email", "sms", "push"},
		}),
	}
}

// Initialize sets up the module with core infrastructure
func (m *Module) Initialize(ctx context.Context, core *module.Core) error {
	// Call base initialization
	if err := m.BaseModule.Initialize(ctx, core); err != nil {
		return err
	}

	// Validate commercial license
	licenseKey := os.Getenv("NOTIFICATIONS_LICENSE_KEY")
	if licenseKey == "" {
		// Allow running without license in development
		if core.Config.App.Environment == "development" || core.Config.App.Environment == "test" {
			logger.Warn("Notifications module running without license (dev/test mode)")
		} else {
			return fmt.Errorf("notifications module requires a license key (set NOTIFICATIONS_LICENSE_KEY env var)")
		}
	} else {
		// Parse and validate license
		lic, err := license.Parse(licenseKey)
		if err != nil {
			return fmt.Errorf("failed to parse license: %w", err)
		}

		// Validate against module name and secret (use default for now)
		secret := os.Getenv("LICENSE_SECRET")
		if secret == "" {
			secret = "default-secret-change-in-production"
		}

		if err := lic.Validate(secret, "notifications", 7); err != nil {
			return fmt.Errorf("license validation failed: %w", err)
		}

		logger.Info("Notifications module license validated",
			"tier", lic.Tier,
			"expires_at", lic.ExpiryDate.Format("2006-01-02"))
	}

	// Initialize repositories
	notifRepo := postgres.NewNotificationRepository(core.DB)
	prefRepo := postgres.NewUserPreferenceRepository(core.DB)

	// Initialize use case
	m.useCase = usecase.NewNotificationUseCase(notifRepo, prefRepo, core.EventBus)

	// Initialize HTTP handler
	m.handler = handler.NewNotificationHandler(m.useCase)

	logger.Info("Notifications module initialized successfully")

	return nil
}

// RegisterRoutes registers HTTP routes for this module
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	// Health check (must be before parameterized routes)
	router.GET("/notifications/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"module":  m.Metadata().Name,
			"version": m.Metadata().Version,
		})
	})

	// Create notifications group
	notifGroup := router.Group("/notifications")
	{
		// All routes require authentication (middleware registered in main.go)
		// No need to manually check for auth middleware here

		// Notification routes
		notifGroup.POST("", m.handler.SendNotification)
		notifGroup.GET("", m.handler.GetNotifications)
		notifGroup.GET("/unread-count", m.handler.GetUnreadCount)
		notifGroup.GET("/:id", m.handler.GetNotification)
		notifGroup.POST("/:id/opened", m.handler.MarkAsOpened)

		// Preference routes
		notifGroup.GET("/preferences", m.handler.GetPreferences)
		notifGroup.PUT("/preferences", m.handler.UpdatePreferences)
	}

	logger.Info("Notifications module routes registered")
}

// RegisterPermissions returns RBAC permissions for this module
func (m *Module) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{
			Resource:    "notifications",
			Action:      "read",
			Description: "View own notifications",
		},
		{
			Resource:    "notifications",
			Action:      "create",
			Description: "Send notifications",
		},
		{
			Resource:    "notifications",
			Action:      "update",
			Description: "Update notification status",
		},
		{
			Resource:    "notifications.preferences",
			Action:      "read",
			Description: "View notification preferences",
		},
		{
			Resource:    "notifications.preferences",
			Action:      "update",
			Description: "Update notification preferences",
		},
	}
}

// Dependencies returns modules this module depends on
func (m *Module) Dependencies() []string {
	return []string{} // No dependencies
}

// Start is called after all modules are initialized
func (m *Module) Start(ctx context.Context) error {
	logger.Info("Notifications module started")
	return nil
}

// Stop is called during graceful shutdown
func (m *Module) Stop(ctx context.Context) error {
	logger.Info("Notifications module stopped")
	return nil
}

// HealthCheck returns the module's health status
func (m *Module) HealthCheck(ctx context.Context) error {
	// TODO: Check if database is accessible, event bus is running
	return nil
}

// RegisterMigrations returns database migrations for this module
func (m *Module) RegisterMigrations() []module.Migration {
	return []module.Migration{
		{
			Version:     1,
			Description: "Create notifications tables",
			// Migration files are in migrations/notifications/ directory
			// They will be auto-loaded by migration manager
		},
	}
}
