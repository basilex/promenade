package profiles

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/profiles/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/profiles/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/profiles/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/module"
	moduleconfig "github.com/basilex/promenade/pkg/module/config"
)

// ProfilesModule implements module.IModule for user profiles and contacts functionality
type ProfilesModule struct {
	db             *sqlx.DB
	eventBus       bus.IBus
	config         *moduleconfig.Config
	profileUseCase usecase.IUserProfileUseCase
	profileHandler *handler.UserProfileHandler
	contactUseCase usecase.IUserContactUseCase
	contactHandler *handler.UserContactHandler
}

// New creates a new profiles module instance
func New() module.IModule {
	return &ProfilesModule{}
}

// Metadata returns module information
func (m *ProfilesModule) Metadata() module.Metadata {
	return module.Metadata{
		Name:        "profiles",
		DisplayName: "User Profiles & Contacts",
		Version:     "1.0.0",
		Description: "User profile management with contact information (email, phone, telegram, etc.)",
		Author:      "Promenade Team",
		License:     "MIT",
		Tags:        []string{"profiles", "contacts", "user-data"},
	}
}

// Dependencies returns list of module dependencies
func (m *ProfilesModule) Dependencies() []string {
	return []string{} // No dependencies
}

// Initialize sets up the module components
func (m *ProfilesModule) Initialize(ctx context.Context, core *module.Core) error {
	slog.Info("Initializing profiles module")

	m.db = core.DB
	m.eventBus = core.EventBus

	// Load module's own configuration
	cfg, err := moduleconfig.Load("internal/modules/profiles/config", "promenade")
	if err != nil {
		slog.Warn("Failed to load profiles module config, using defaults", "error", err)
	} else {
		m.config = cfg
		slog.Info("Profiles module config loaded",
			"version", cfg.IModule.Version,
			"enabled", cfg.IModule.Enabled,
		)
	}

	// Get module-specific settings from config
	maxSocialLinks := 10
	maxContactsPerUser := 5

	if m.config != nil {
		if profiles, ok := m.config.Settings["profiles"].(map[string]any); ok {
			if v, ok := profiles["max_social_links"].(int); ok {
				maxSocialLinks = v
			}
		}

		if contacts, ok := m.config.Settings["contacts"].(map[string]any); ok {
			if v, ok := contacts["max_per_user"].(int); ok {
				maxContactsPerUser = v
			}
		}
	}

	slog.Info("Profiles module settings",
		"max_social_links", maxSocialLinks,
		"max_contacts_per_user", maxContactsPerUser,
	)

	// Initialize repositories
	profileRepo := postgres.NewUserProfileRepository(m.db)
	contactRepo := postgres.NewUserContactRepository(m.db)

	// Initialize use cases
	m.profileUseCase = usecase.NewUserProfileUseCase(profileRepo)
	m.contactUseCase = usecase.NewUserContactUseCase(contactRepo)

	// Initialize handlers
	m.profileHandler = handler.NewUserProfileHandler(m.profileUseCase)
	m.contactHandler = handler.NewUserContactHandler(m.contactUseCase)

	slog.Info("Profiles module initialized successfully")
	return nil
}

// RegisterRoutes registers HTTP routes for this module
func (m *ProfilesModule) RegisterRoutes(router *gin.RouterGroup) {
	slog.Info("Registering profiles module routes")

	// Profile routes
	profilesGroup := router.Group("/profiles")
	{
		profilesGroup.GET("", m.profileHandler.ListProfiles)
		profilesGroup.POST("", m.profileHandler.CreateProfile)
		profilesGroup.GET("/me", m.profileHandler.GetMyProfile)
		profilesGroup.GET("/search", m.profileHandler.SearchProfiles)
		profilesGroup.GET("/nickname/:nickname", m.profileHandler.GetProfileByNickname)
		profilesGroup.GET("/:userId", m.profileHandler.GetProfile)
		profilesGroup.PUT("/:userId", m.profileHandler.UpdateProfile)
		profilesGroup.DELETE("/:userId", m.profileHandler.DeleteProfile)
		profilesGroup.POST("/:userId/ban", m.profileHandler.BanProfile)
		profilesGroup.POST("/:userId/unban", m.profileHandler.UnbanProfile)
		profilesGroup.POST("/:userId/verify", m.profileHandler.VerifyProfile)
		profilesGroup.POST("/:userId/unverify", m.profileHandler.UnverifyProfile)
	}

	// Contact routes
	contactsGroup := router.Group("/contacts")
	{
		contactsGroup.GET("", m.contactHandler.GetUserContacts)
		contactsGroup.POST("", m.contactHandler.CreateContact)
		contactsGroup.GET("/type/:type", m.contactHandler.GetContactsByType)
		contactsGroup.GET("/primary", m.contactHandler.GetPrimaryContact)
		contactsGroup.GET("/:id", m.contactHandler.GetContact)
		contactsGroup.PUT("/:id", m.contactHandler.UpdateContact)
		contactsGroup.DELETE("/:id", m.contactHandler.DeleteContact)
		contactsGroup.POST("/:id/set-primary", m.contactHandler.SetPrimaryContact)
		contactsGroup.POST("/:id/toggle-active", m.contactHandler.ToggleContactActive)
	}

	slog.Info("Profiles module routes registered successfully")
}

// RegisterMigrations returns migrations for this module
func (m *ProfilesModule) RegisterMigrations() []module.Migration {
	return []module.Migration{
		{
			Version:     4,
			Description: "Create user_contacts table",
			Up:          "-- Migration managed in migrations/000004_create_user_contacts.up.sql",
			Down:        "-- Migration managed in migrations/000004_create_user_contacts.down.sql",
		},
		{
			Version:     5,
			Description: "Create user_profiles table",
			Up:          "-- Migration managed in migrations/000005_create_user_profiles.up.sql",
			Down:        "-- Migration managed in migrations/000005_create_user_profiles.down.sql",
		},
	}
}

// RegisterEventHandlers subscribes to events
func (m *ProfilesModule) RegisterEventHandlers(eventBus bus.IBus) error {
	slog.Info("Registering profiles module event handlers")
	return nil
}

// RegisterPermissions returns permissions required by this module
func (m *ProfilesModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		{Resource: "profiles", Action: "create", Description: "Create user profiles"},
		{Resource: "profiles", Action: "read", Description: "View user profiles"},
		{Resource: "profiles", Action: "update", Description: "Update own profile"},
		{Resource: "profiles", Action: "delete", Description: "Delete own profile"},
		{Resource: "profiles", Action: "manage", Description: "Manage all profiles (admin)"},
		{Resource: "contacts", Action: "create", Description: "Create contacts"},
		{Resource: "contacts", Action: "read", Description: "View own contacts"},
		{Resource: "contacts", Action: "update", Description: "Update own contacts"},
		{Resource: "contacts", Action: "delete", Description: "Delete own contacts"},
		{Resource: "contacts", Action: "verify", Description: "Verify contact ownership"},
		{Resource: "contacts", Action: "manage", Description: "Manage all contacts (admin)"},
	}
}

// Start starts the module background tasks
func (m *ProfilesModule) Start(ctx context.Context) error {
	slog.Info("Starting profiles module")
	return nil
}

// Stop gracefully stops the module
func (m *ProfilesModule) Stop(ctx context.Context) error {
	slog.Info("Stopping profiles module")
	return nil
}

// HealthCheck checks module health
func (m *ProfilesModule) HealthCheck(ctx context.Context) error {
	if m.db == nil {
		return module.ErrModuleNotInitialized
	}

	if err := m.db.PingContext(ctx); err != nil {
		return err
	}

	var count int
	err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM profiles_profiles").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}
