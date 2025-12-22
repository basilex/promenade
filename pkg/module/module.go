package module

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/pkg/bus"
	jwtpkg "github.com/basilex/promenade/pkg/jwt"
)

// Module defines the interface that all Promenade modules must implement.
// This enables a plugin architecture where business logic can be packaged
// as independent, reusable modules.
type Module interface {
	// Metadata returns module information
	Metadata() Metadata

	// Dependencies returns list of module names this module depends on
	Dependencies() []string

	// Initialize is called once during application startup
	// Core provides access to shared infrastructure (DB, EventBus, JWT, etc.)
	Initialize(ctx context.Context, core *Core) error

	// RegisterRoutes registers HTTP routes for this module
	// The router group is scoped to the module (e.g., /api/v1/warehouse)
	RegisterRoutes(router *gin.RouterGroup)

	// RegisterMigrations returns database migrations for this module
	// Migrations are applied in order during startup
	RegisterMigrations() []Migration

	// RegisterEventHandlers subscribes to domain events
	// This enables async communication between modules
	RegisterEventHandlers(bus bus.Bus) error

	// RegisterPermissions returns RBAC permissions for this module
	// These are automatically inserted into the database
	RegisterPermissions() []Permission

	// Start is called after all modules are initialized
	// Use this for background workers, cron jobs, etc.
	Start(ctx context.Context) error

	// Stop is called during graceful shutdown
	Stop(ctx context.Context) error

	// HealthCheck returns the module's health status
	HealthCheck(ctx context.Context) error
}

// Metadata contains module information
type Metadata struct {
	Name        string   // Unique module identifier (e.g., "warehouse")
	DisplayName string   // Human-readable name (e.g., "Warehouse Management")
	Version     string   // Semantic version (e.g., "1.2.0")
	Author      string   // Module author/vendor
	Description string   // Short description
	License     string   // License type (e.g., "MIT", "Commercial")
	Tags        []string // Categories (e.g., ["inventory", "logistics"])
}

// Core provides shared infrastructure to modules
type Core struct {
	// Database connection
	DB *sqlx.DB

	// Event bus for async communication
	EventBus bus.Bus

	// JWT manager for authentication
	JWT *jwtpkg.JWTManager

	// Application configuration (core config only)
	Config *config.AppConfig

	// Structured logger
	Logger *slog.Logger

	// Module registry (for inter-module communication)
	Registry *Registry
}

// Migration represents a database migration for a module
type Migration struct {
	Version     int    // Migration version (incremental)
	Description string // Human-readable description
	Up          string // SQL to apply migration
	Down        string // SQL to rollback migration
}

// Permission represents an RBAC permission for a module
type Permission struct {
	Resource    string // Resource name (e.g., "warehouse:items")
	Action      string // Action (e.g., "create", "read", "update", "delete")
	Description string // Human-readable description
}

// BaseModule provides default implementations for optional methods
// Embed this in your module to avoid implementing every method
type BaseModule struct {
	meta Metadata
	core *Core
}

func NewBaseModule(meta Metadata) *BaseModule {
	return &BaseModule{meta: meta}
}

func (m *BaseModule) Metadata() Metadata {
	return m.meta
}

func (m *BaseModule) Dependencies() []string {
	return []string{} // No dependencies by default
}

func (m *BaseModule) Initialize(ctx context.Context, core *Core) error {
	m.core = core
	return nil
}

func (m *BaseModule) RegisterRoutes(router *gin.RouterGroup) {
	// No routes by default
}

func (m *BaseModule) RegisterMigrations() []Migration {
	return []Migration{} // No migrations by default
}

func (m *BaseModule) RegisterEventHandlers(bus bus.Bus) error {
	return nil // No event handlers by default
}

func (m *BaseModule) RegisterPermissions() []Permission {
	return []Permission{} // No permissions by default
}

func (m *BaseModule) Start(ctx context.Context) error {
	return nil // Nothing to start by default
}

func (m *BaseModule) Stop(ctx context.Context) error {
	return nil // Nothing to stop by default
}

func (m *BaseModule) HealthCheck(ctx context.Context) error {
	return nil // Healthy by default
}

// GetCore returns the core infrastructure (for use in embedded BaseModule)
func (m *BaseModule) GetCore() *Core {
	return m.core
}
