package audit

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/basilex/promenade/internal/modules/audit/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/audit/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/audit/purge"
	"github.com/basilex/promenade/internal/modules/audit/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/license"
	"github.com/basilex/promenade/pkg/module"
	moduleconfig "github.com/basilex/promenade/pkg/module/config"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// AuditModule represents the audit module
type AuditModule struct {
	config     *moduleconfig.Config // Use standard config like other modules
	db         *sqlx.DB
	handler    *handler.AuditEventHandler
	useCase    usecase.IAuditEventUseCase
	license    *license.License
	licenseKey string
}

// New creates a new audit module instance
func New() module.IModule {
	return &AuditModule{}
}

// Metadata returns module information
func (m *AuditModule) Metadata() module.Metadata {
	return module.Metadata{
		Name:        "audit",
		DisplayName: "Audit Logging",
		Version:     "1.0.0",
		Description: "Immutable audit logging with cryptographic signatures",
		Author:      "Promenade Team",
		License:     "Commercial",
		Tags:        []string{"audit", "compliance", "security"},
	}
}

// Dependencies returns module dependencies
func (m *AuditModule) Dependencies() []string {
	return []string{} // No dependencies
}

// Initialize initializes the audit module
func (m *AuditModule) Initialize(ctx context.Context, core *module.Core) error {
	slog.Info("Initializing audit module")

	m.db = core.DB

	// Load module's own configuration (standard approach like other modules)
	cfg, err := moduleconfig.Load("internal/modules/audit/config", "promenade")
	if err != nil {
		slog.Warn("Failed to load audit module config, using fallback defaults", "error", err)
		cfg = getDefaultAuditConfig()
		slog.Info("Using fallback audit configuration")
	}

	// Ensure config is not nil
	if cfg == nil {
		return fmt.Errorf("audit config not loaded - cannot initialize module")
	}

	m.config = cfg
	slog.Info("Audit module config loaded",
		"version", cfg.Module.Version,
		"enabled", cfg.Module.Enabled,
	)

	// Check if module is enabled
	if !m.config.Module.Enabled {
		slog.Info("Audit module is disabled")
		return nil
	}

	// Validate license if required
	licenseRequired := false
	if moduleSettings, ok := m.config.Settings["module"].(map[string]any); ok {
		if req, ok := moduleSettings["license_required"].(bool); ok {
			licenseRequired = req
		}
	}

	if licenseRequired {
		// Get license key from env or config
		licenseKey := os.Getenv("AUDIT_LICENSE_KEY")
		if licenseKey == "" {
			if licenseConfig, ok := m.config.Settings["license"].(map[string]any); ok {
				if key, ok := licenseConfig["key"].(string); ok {
					licenseKey = key
				}
			}
		}
		m.licenseKey = licenseKey

		if err := m.validateLicense(); err != nil {
			return fmt.Errorf("license validation failed: %w", err)
		}
		slog.Info("Audit module license validated",
			"tier", m.license.Tier,
			"expires", m.license.ExpiryDate.Format("2006-01-02"),
			"days_until_expiry", m.license.DaysUntilExpiry())
	}

	// Get signature secret from env or config
	secret := os.Getenv("AUDIT_SECRET")
	if secret == "" {
		if audit, ok := m.config.Settings["audit"].(map[string]any); ok {
			if s, ok := audit["signature_secret"].(string); ok {
				secret = s
			}
		}
	}
	if secret == "" {
		return fmt.Errorf("audit signature secret not configured")
	}

	// Initialize repository and use case
	repo := postgres.NewAuditEventRepository(m.db)
	m.useCase = usecase.NewAuditEventUseCase(repo, secret)
	m.handler = handler.NewAuditEventHandler(m.useCase)

	// Register purge handler if retention is enabled
	var retentionEnabled bool
	var retentionDays int

	if retention, ok := m.config.Settings["retention"].(map[string]any); ok {
		if enabled, ok := retention["enabled"].(bool); ok {
			retentionEnabled = enabled
		}
		if days, ok := retention["days"].(int); ok {
			retentionDays = days
		}
	}

	if retentionEnabled && retentionDays > 0 {
		if err := purge.RegisterPurgeHandlers(m.useCase, retentionDays); err != nil {
			slog.Warn("Failed to register audit purge handler", "error", err)
		} else {
			slog.Info("Audit purge handler registered", "retention_days", retentionDays)
		}
	}

	slog.Info("Audit module initialized successfully")
	return nil
}

// RegisterRoutes registers HTTP routes
func (m *AuditModule) RegisterRoutes(router *gin.RouterGroup) {
	if !m.config.Module.Enabled {
		return
	}
	m.registerRoutes(router)
}

// RegisterMigrations returns database migrations
func (m *AuditModule) RegisterMigrations() []module.Migration {
	return []module.Migration{}
}

// RegisterEventHandlers subscribes to domain events
func (m *AuditModule) RegisterEventHandlers(eventBus bus.IBus) error {
	return nil // No event subscriptions needed
}

// RegisterPermissions returns RBAC permissions
func (m *AuditModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "audit", Action: "read", Description: "View audit logs"},
		{Resource: "audit", Action: "create", Description: "Create audit entries"},
		{Resource: "audit", Action: "verify", Description: "Verify audit signatures"},
	}
}

// Start starts the audit module
func (m *AuditModule) Start(ctx context.Context) error {
	return nil
}

// Stop stops the audit module
func (m *AuditModule) Stop(ctx context.Context) error {
	return nil
}

// HealthCheck performs module health check
func (m *AuditModule) HealthCheck(ctx context.Context) error {
	if !m.config.Module.Enabled {
		return nil
	}

	// Check license expiry if required
	licenseRequired := false
	if moduleSettings, ok := m.config.Settings["module"].(map[string]any); ok {
		if req, ok := moduleSettings["license_required"].(bool); ok {
			licenseRequired = req
		}
	}

	gracePeriodDays := 0
	if licenseConfig, ok := m.config.Settings["license"].(map[string]any); ok {
		if grace, ok := licenseConfig["grace_period_days"].(int); ok {
			gracePeriodDays = grace
		}
	}

	if licenseRequired && m.license != nil {
		if m.license.IsExpired() {
			daysExpired := -m.license.DaysUntilExpiry()
			if daysExpired > gracePeriodDays {
				return fmt.Errorf("license expired %d days ago", daysExpired)
			}
		}
	}

	if err := m.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// validateLicense validates module license
func (m *AuditModule) validateLicense() error {
	if m.licenseKey == "" {
		return fmt.Errorf("license key not provided (set AUDIT_LICENSE_KEY environment variable)")
	}

	lic, err := license.Parse(m.licenseKey)
	if err != nil {
		return fmt.Errorf("failed to parse license: %w", err)
	}

	secret := os.Getenv("LICENSE_SECRET")
	if secret == "" {
		return fmt.Errorf("LICENSE_SECRET environment variable not set")
	}

	// Get grace period from config
	gracePeriodDays := 0
	if licenseConfig, ok := m.config.Settings["license"].(map[string]any); ok {
		if grace, ok := licenseConfig["grace_period_days"].(int); ok {
			gracePeriodDays = grace
		}
	}

	if err := lic.Validate(secret, m.Metadata().Name, gracePeriodDays); err != nil {
		return fmt.Errorf("license validation failed: %w", err)
	}

	m.license = lic
	return nil
}

// registerRoutes registers HTTP routes
func (m *AuditModule) registerRoutes(router *gin.RouterGroup) {
	auditGroup := router.Group("/audit")
	{
		auditGroup.POST("/events", m.handler.CreateAuditEvent)
		auditGroup.GET("/events", m.handler.ListAuditEvents)
		auditGroup.GET("/events/:id", m.handler.GetAuditEvent)
		auditGroup.GET("/events/:id/verify", m.handler.VerifyAuditEvent)
		auditGroup.GET("/events/entity/:entity_type/:entity_id", m.handler.GetAuditEventsByEntity)
	}
}

// getDefaultAuditConfig returns fallback configuration with safe defaults
func getDefaultAuditConfig() *moduleconfig.Config {
	return &moduleconfig.Config{
		Module: moduleconfig.ModuleSection{
			Name:    "audit",
			Version: "1.0.0",
			Enabled: true,
		},
		Settings: map[string]any{
			"module": map[string]any{
				"license_required": false,
				"license_key":      "",
			},
			"signature": map[string]any{
				"algorithm": "sha256",
				"secret":    "", // Must be set via AUDIT_SIGNATURE_SECRET env var
			},
			"license": map[string]any{
				"secret":            "",
				"grace_period_days": 30,
			},
		},
		Purge: moduleconfig.PurgeSection{
			Enabled: false, // Audit logs should NEVER be purged (compliance)
		},
	}
}
