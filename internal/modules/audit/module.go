package audit

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/basilex/promenade/internal/modules/audit/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/audit/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/audit/purge"
	"github.com/basilex/promenade/internal/modules/audit/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/license"
	"github.com/basilex/promenade/pkg/module"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v3"
)

const (
	ModuleName = "audit"
	Version    = "1.0.0"
)

// Module represents the audit module
type Module struct {
	config     *Config
	db         *sqlx.DB
	handler    *handler.AuditEventHandler
	useCase    usecase.AuditEventUseCase
	license    *license.License
	licenseKey string
}

// Config holds audit module configuration
type Config struct {
	Module struct {
		Enabled         bool `yaml:"enabled"`
		LicenseRequired bool `yaml:"license_required"`
	} `yaml:"module"`
	Audit struct {
		SignatureSecret string `yaml:"signature_secret"`
	} `yaml:"audit"`
	License struct {
		Key        string `yaml:"key"`
		Validation struct {
			CheckOnStartup bool `yaml:"check_on_startup"`
			CheckOnRequest bool `yaml:"check_on_request"`
		} `yaml:"validation"`
		GracePeriodDays int `yaml:"grace_period_days"`
	} `yaml:"license"`
	Retention struct {
		Enabled bool `yaml:"enabled"`
		Days    int  `yaml:"days"`
	} `yaml:"retention"`
	Purge struct {
		Enabled  bool   `yaml:"enabled"`
		Schedule string `yaml:"schedule"`
	} `yaml:"purge"`
}

// New creates a new audit module instance
func New() module.Module {
	return &Module{}
}

// Metadata returns module information
func (m *Module) Metadata() module.Metadata {
	return module.Metadata{
		Name:        ModuleName,
		DisplayName: "Audit Logging",
		Version:     Version,
		Description: "Immutable audit logging with cryptographic signatures",
		Author:      "Promenade Team",
		License:     "Commercial",
		Tags:        []string{"audit", "compliance", "security"},
	}
}

// Name returns module name
func (m *Module) Name() string {
	return ModuleName
}

// Version returns module version
func (m *Module) Version() string {
	return Version
}

// Dependencies returns module dependencies
func (m *Module) Dependencies() []string {
	return []string{} // No dependencies
}

// Initialize initializes the audit module
func (m *Module) Initialize(ctx context.Context, core *module.Core) error {
	m.db = core.DB

	if err := m.loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !m.config.Module.Enabled {
		slog.Info("Audit module is disabled")
		return nil
	}

	if m.config.Module.LicenseRequired {
		if err := m.validateLicense(); err != nil {
			return fmt.Errorf("license validation failed: %w", err)
		}
		slog.Info("Audit module license validated",
			"tier", m.license.Tier,
			"expires", m.license.ExpiryDate.Format("2006-01-02"),
			"days_until_expiry", m.license.DaysUntilExpiry())
	}

	secret := os.Getenv("AUDIT_SECRET")
	if secret == "" {
		secret = m.config.Audit.SignatureSecret
	}
	if secret == "" {
		return fmt.Errorf("audit signature secret not configured")
	}

	repo := postgres.NewAuditEventRepository(m.db)
	m.useCase = usecase.NewAuditEventUseCase(repo, secret)
	m.handler = handler.NewAuditEventHandler(m.useCase)

	// Register purge handler if retention is enabled
	if m.config.Retention.Enabled && m.config.Retention.Days > 0 {
		if err := purge.RegisterPurgeHandlers(m.useCase, m.config.Retention.Days); err != nil {
			slog.Warn("Failed to register audit purge handler", "error", err)
		} else {
			slog.Info("Audit purge handler registered", "retention_days", m.config.Retention.Days)
		}
	}

	slog.Info("Audit module initialized", "version", Version)
	return nil
}

// RegisterRoutes registers HTTP routes
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	if !m.config.Module.Enabled {
		return
	}
	m.registerRoutes(router)
}

// RegisterMigrations returns database migrations
func (m *Module) RegisterMigrations() []module.Migration {
	return []module.Migration{}
}

// RegisterEventHandlers subscribes to domain events
func (m *Module) RegisterEventHandlers(eventBus bus.Bus) error {
	return nil // No event subscriptions needed
}

// RegisterPermissions returns RBAC permissions
func (m *Module) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "audit", Action: "read", Description: "View audit logs"},
		{Resource: "audit", Action: "create", Description: "Create audit entries"},
		{Resource: "audit", Action: "verify", Description: "Verify audit signatures"},
	}
}

// Start starts the audit module
func (m *Module) Start(ctx context.Context) error {
	return nil
}

// Stop stops the audit module
func (m *Module) Stop(ctx context.Context) error {
	return nil
}

// HealthCheck performs module health check
func (m *Module) HealthCheck(ctx context.Context) error {
	if !m.config.Module.Enabled {
		return nil
	}

	if m.config.Module.LicenseRequired && m.license != nil {
		if m.license.IsExpired() {
			daysExpired := -m.license.DaysUntilExpiry()
			if daysExpired > m.config.License.GracePeriodDays {
				return fmt.Errorf("license expired %d days ago", daysExpired)
			}
		}
	}

	if err := m.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// loadConfig loads module configuration
func (m *Module) loadConfig() error {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	configPath := filepath.Join("internal", "modules", ModuleName, "config", fmt.Sprintf("config.%s.yaml", env))
	if env == "development" {
		configPath = filepath.Join("internal", "modules", ModuleName, "config", "config.dev.yaml")
	} else if env == "test" {
		configPath = filepath.Join("internal", "modules", ModuleName, "config", "config.test.yaml")
	} else if env == "production" {
		configPath = filepath.Join("internal", "modules", ModuleName, "config", "config.prod.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	m.config = &Config{}
	if err := yaml.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if key := os.Getenv("AUDIT_LICENSE_KEY"); key != "" {
		m.licenseKey = key
	} else {
		m.licenseKey = m.config.License.Key
	}

	return nil
}

// validateLicense validates module license
func (m *Module) validateLicense() error {
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

	if err := lic.Validate(secret, ModuleName, m.config.License.GracePeriodDays); err != nil {
		return fmt.Errorf("license validation failed: %w", err)
	}

	m.license = lic
	return nil
}

// registerRoutes registers HTTP routes
func (m *Module) registerRoutes(router *gin.RouterGroup) {
	auditGroup := router.Group("/audit")
	{
		auditGroup.POST("/events", m.handler.CreateAuditEvent)
		auditGroup.GET("/events", m.handler.ListAuditEvents)
		auditGroup.GET("/events/:id", m.handler.GetAuditEvent)
		auditGroup.GET("/events/:id/verify", m.handler.VerifyAuditEvent)
		auditGroup.GET("/events/entity/:entity_type/:entity_id", m.handler.GetAuditEventsByEntity)
	}
}
