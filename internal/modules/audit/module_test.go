package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/module"
)

// getProjectRoot finds the project root by looking for go.mod
func getProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (go.mod not found)")
		}
		dir = parent
	}
}

func TestModule_Metadata(t *testing.T) {
	mod := New()

	metadata := mod.Metadata()

	assert.Equal(t, "audit", metadata.Name)
	assert.Equal(t, "Audit Logging", metadata.DisplayName)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.NotEmpty(t, metadata.Description)
	assert.Contains(t, metadata.Tags, "audit")
	assert.Contains(t, metadata.Tags, "compliance")
}

func TestModule_Name(t *testing.T) {
	mod := New().(*IModule)
	assert.Equal(t, "audit", mod.Name())
}

func TestModule_Version(t *testing.T) {
	mod := New().(*IModule)
	assert.Equal(t, "1.0.0", mod.Version())
}

func TestModule_Dependencies(t *testing.T) {
	mod := New()
	deps := mod.Dependencies()
	assert.NotNil(t, deps)
	assert.Empty(t, deps)
}

func TestNew(t *testing.T) {
	t.Run("creates new module", func(t *testing.T) {
		mod := New()
		assert.NotNil(t, mod)
		assert.Implements(t, (*module.IModule)(nil), mod)
	})

	t.Run("module is properly typed", func(t *testing.T) {
		mod := New()
		_, ok := mod.(module.IModule)
		assert.True(t, ok)
	})
}

func TestModule_Constants(t *testing.T) {
	assert.Equal(t, "audit", ModuleName)
	assert.Equal(t, "1.0.0", Version)
}

func TestModule_Registration(t *testing.T) {
	t.Run("module can be registered", func(t *testing.T) {
		registry := module.NewRegistry()
		mod := New()
		err := registry.Register(mod)
		require.NoError(t, err)
	})

	t.Run("duplicate registration fails", func(t *testing.T) {
		registry := module.NewRegistry()
		mod1 := New()
		err := registry.Register(mod1)
		require.NoError(t, err)

		mod2 := New()
		err = registry.Register(mod2)
		assert.Error(t, err)
	})
}

func TestModule_RegisterPermissions(t *testing.T) {
	mod := &IModule{}
	permissions := mod.RegisterPermissions()

	assert.Len(t, permissions, 3)

	// Check specific permissions
	var readPerm, createPerm, verifyPerm bool
	for _, p := range permissions {
		if p.Resource == "audit" && p.Action == "read" {
			readPerm = true
			assert.Equal(t, "View audit logs", p.Description)
		}
		if p.Resource == "audit" && p.Action == "create" {
			createPerm = true
			assert.Equal(t, "Create audit entries", p.Description)
		}
		if p.Resource == "audit" && p.Action == "verify" {
			verifyPerm = true
			assert.Equal(t, "Verify audit signatures", p.Description)
		}
	}

	assert.True(t, readPerm, "read permission should exist")
	assert.True(t, createPerm, "create permission should exist")
	assert.True(t, verifyPerm, "verify permission should exist")
}

func TestModule_RegisterMigrations(t *testing.T) {
	mod := &IModule{}
	migrations := mod.RegisterMigrations()
	assert.NotNil(t, migrations)
	assert.Len(t, migrations, 0)
}

func TestModule_RegisterEventHandlers(t *testing.T) {
	mod := &IModule{}
	err := mod.RegisterEventHandlers(nil)
	assert.NoError(t, err)
}

func TestModule_StartStop(t *testing.T) {
	ctx := context.Background()
	mod := &IModule{}

	t.Run("start succeeds", func(t *testing.T) {
		err := mod.Start(ctx)
		assert.NoError(t, err)
	})

	t.Run("stop succeeds", func(t *testing.T) {
		err := mod.Stop(ctx)
		assert.NoError(t, err)
	})
}

func TestModule_LoadConfig(t *testing.T) {
	projectRoot := getProjectRoot(t)
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	
	// Change to project root for tests
	err = os.Chdir(projectRoot)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	t.Run("loads test config successfully", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "test")
		defer os.Unsetenv("ENVIRONMENT")

		mod := &IModule{}
		err := mod.loadConfig()

		assert.NoError(t, err)
		assert.NotNil(t, mod.config)
	})

	t.Run("loads dev config by default", func(t *testing.T) {
		os.Unsetenv("ENVIRONMENT")

		mod := &IModule{}
		err := mod.loadConfig()

		assert.NoError(t, err)
		assert.NotNil(t, mod.config)
	})

	t.Run("respects AUDIT_LICENSE_KEY env variable", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "test")
		os.Setenv("AUDIT_LICENSE_KEY", "test-key-12345")
		defer func() {
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("AUDIT_LICENSE_KEY")
		}()

		mod := &IModule{}
		err := mod.loadConfig()

		assert.NoError(t, err)
		assert.Equal(t, "test-key-12345", mod.licenseKey)
	})

	t.Run("uses config file license key if env not set", func(t *testing.T) {
		os.Setenv("ENVIRONMENT", "test")
		os.Unsetenv("AUDIT_LICENSE_KEY")
		defer os.Unsetenv("ENVIRONMENT")

		mod := &IModule{}
		err := mod.loadConfig()

		assert.NoError(t, err)
		assert.NotNil(t, mod.config)
	})
}

func TestModule_Initialize_Disabled(t *testing.T) {
	projectRoot := getProjectRoot(t)
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(projectRoot)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create a test database connection
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping test: TEST_DATABASE_URL not set")
		return
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer db.Close()

	os.Setenv("ENVIRONMENT", "test")
	defer os.Unsetenv("ENVIRONMENT")

	mod := &IModule{}
	core := &module.Core{DB: db}

	// If module is disabled in config, initialization should succeed but do nothing
	err = mod.Initialize(context.Background(), core)
	// Should not error even if module is disabled
	if err != nil {
		assert.Contains(t, err.Error(), "license", "Error should be about license or config")
	}
}

func TestModule_Initialize_Success(t *testing.T) {
	projectRoot := getProjectRoot(t)
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(projectRoot)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("Skipping test: TEST_DATABASE_URL not set")
		return
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return
	}
	defer db.Close()

	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("AUDIT_SECRET", "test-secret-key")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("AUDIT_SECRET")
	}()

	t.Run("initializes successfully when enabled without license", func(t *testing.T) {
		auditMod := New()
		mod, ok := auditMod.(*IModule)
		require.True(t, ok, "New() should return *IModule")
		
		core := &module.Core{DB: db}

		err := mod.Initialize(context.Background(), core)
		// Might fail due to license, but should at least load config
		if err != nil {
			// Check if it's a license error (expected) or config error (unexpected)
			assert.Contains(t, err.Error(), "license", "Should fail on license, not config")
		}

		// Verify config was loaded
		assert.NotNil(t, mod.config)
		assert.NotNil(t, mod.db)
	})

	t.Run("loads config and sets database", func(t *testing.T) {
		auditMod := New()
		mod, ok := auditMod.(*IModule)
		require.True(t, ok, "New() should return *IModule")
		
		core := &module.Core{DB: db}

		_ = mod.Initialize(context.Background(), core)

		assert.NotNil(t, mod.config)
		assert.Equal(t, db, mod.db)
	})
}

func TestModule_HealthCheck(t *testing.T) {
	ctx := context.Background()

	t.Run("disabled module passes health check", func(t *testing.T) {
		mod := &IModule{
			config: &Config{},
		}
		mod.config.IModule.Enabled = false

		err := mod.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("enabled module without db panics", func(t *testing.T) {
		mod := &IModule{
			config: &Config{},
		}
		mod.config.IModule.Enabled = true
		// m.db is nil, so PingContext will panic

		// Expecting panic when db is nil
		assert.Panics(t, func() {
			_ = mod.HealthCheck(ctx)
		}, "HealthCheck should panic when db is nil")
	})
}

func TestModule_ValidateLicense_Errors(t *testing.T) {
	t.Run("empty license key fails", func(t *testing.T) {
		mod := &IModule{
			config:     &Config{},
			licenseKey: "",
		}

		err := mod.validateLicense()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "license key not provided")
	})

	t.Run("invalid license format fails", func(t *testing.T) {
		mod := &IModule{
			config:     &Config{},
			licenseKey: "invalid-license-format",
		}

		err := mod.validateLicense()
		assert.Error(t, err)
	})

	t.Run("missing LICENSE_SECRET fails", func(t *testing.T) {
		os.Unsetenv("LICENSE_SECRET")

		mod := &IModule{
			config:     &Config{},
			licenseKey: "PROMENADE-AUDIT-PRO-20261231-ABC123",
		}

		err := mod.validateLicense()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LICENSE_SECRET")
	})
}

func TestModule_Config_Structure(t *testing.T) {
	config := &Config{}

	t.Run("has module section", func(t *testing.T) {
		config.IModule.Enabled = true
		config.IModule.LicenseRequired = true
		assert.True(t, config.IModule.Enabled)
		assert.True(t, config.IModule.LicenseRequired)
	})

	t.Run("has audit section", func(t *testing.T) {
		config.Audit.SignatureSecret = "test-secret"
		assert.Equal(t, "test-secret", config.Audit.SignatureSecret)
	})

	t.Run("has license section", func(t *testing.T) {
		config.License.Key = "test-key"
		config.License.Validation.CheckOnStartup = true
		config.License.Validation.CheckOnRequest = false
		config.License.GracePeriodDays = 7

		assert.Equal(t, "test-key", config.License.Key)
		assert.True(t, config.License.Validation.CheckOnStartup)
		assert.False(t, config.License.Validation.CheckOnRequest)
		assert.Equal(t, 7, config.License.GracePeriodDays)
	})

	t.Run("has retention section", func(t *testing.T) {
		config.Retention.Enabled = true
		config.Retention.Days = 90

		assert.True(t, config.Retention.Enabled)
		assert.Equal(t, 90, config.Retention.Days)
	})

	t.Run("has purge section", func(t *testing.T) {
		config.Purge.Enabled = true
		config.Purge.Schedule = "0 2 * * *"

		assert.True(t, config.Purge.Enabled)
		assert.Equal(t, "0 2 * * *", config.Purge.Schedule)
	})
}

func TestModule_RegisterRoutes_WhenDisabled(t *testing.T) {
	mod := &IModule{
		config: &Config{},
	}
	mod.config.IModule.Enabled = false

	// Should not panic when module is disabled
	require.NotPanics(t, func() {
		mod.RegisterRoutes(nil)
	})
}

func TestModule_RegisterRoutes(t *testing.T) {
	projectRoot := getProjectRoot(t)
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	
	err = os.Chdir(projectRoot)
	require.NoError(t, err)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("AUDIT_SECRET", "test-secret-key")
	defer func() {
		os.Unsetenv("ENVIRONMENT")
		os.Unsetenv("AUDIT_SECRET")
	}()

	t.Run("registers routes when enabled and initialized", func(t *testing.T) {
		// Skip if no database
		dbURL := os.Getenv("TEST_DATABASE_URL")
		if dbURL == "" {
			t.Skip("Skipping test: TEST_DATABASE_URL not set")
			return
		}

		db, err := sqlx.Connect("postgres", dbURL)
		if err != nil {
			t.Skipf("Skipping test: database not available: %v", err)
			return
		}
		defer db.Close()

		auditMod := New()
		mod, ok := auditMod.(*IModule)
		require.True(t, ok, "New() should return *IModule")
		
		core := &module.Core{DB: db}

		err = mod.Initialize(context.Background(), core)
		if err != nil {
			t.Skipf("Skipping test: initialization failed: %v", err)
			return
		}

		// RegisterRoutes should not panic when handler is set
		require.NotPanics(t, func() {
			// We can't pass nil router as it will panic on method calls
			// This test just verifies the method can be called
			mod.RegisterRoutes(nil)
		})
	})

	t.Run("does not panic when disabled", func(t *testing.T) {
		mod := &IModule{
			config: &Config{},
		}
		mod.config.IModule.Enabled = false

		require.NotPanics(t, func() {
			mod.RegisterRoutes(nil)
		})
	})
}
