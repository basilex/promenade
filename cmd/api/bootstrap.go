package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/internal/infrastructure/health"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/migration"

	"github.com/basilex/promenade/internal/contexts/warehouse/integration"
	"github.com/basilex/promenade/internal/contexts/warehouse/inventory"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
)

// App holds all application dependencies
type App struct {
	Config             *config.AppConfig
	DB                *sqlx.DB
	RedisClient       *redis.Client
	CacheClient       cache.ICache
	JWTManager        *jwt.Manager
	TokenRevoker      *jwt.TokenRevoker
	EventBus           bus.IBus
	HealthChecker     *health.Checker
	OrderEventHandler *integration.OrderEventHandler
}

// Bootstrap initializes all application dependencies
func Bootstrap(cfg *config.AppConfig) (*App, error) {
	app := &App{Config: cfg}

	// Initialize logger
	if err := initLogger(cfg); err != nil {
		return nil, err
	}

	logger.Info("Starting Promenade Platform",
		slog.String("environment", cfg.App.Environment),
		slog.String("version", cfg.App.Version),
	)

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		return nil, err
	}
	app.DB = db

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, err
	}

	// Initialize Redis
	redisClient := initRedis(cfg)
	app.RedisClient = redisClient

	// Initialize cache
	cacheClient, err := initCache(cfg, redisClient)
	if err != nil {
		return nil, err
	}
	app.CacheClient = cacheClient

	// Initialize JWT and Token Revoker
	jwtManager, tokenRevoker := initAuth(cfg, redisClient)
	app.JWTManager = jwtManager
	app.TokenRevoker = tokenRevoker

	// Initialize Event Bus
	eventBus, err := initEventBus(cfg)
	if err != nil {
		return nil, err
	}
	app.EventBus = eventBus

	// Initialize Warehouse Integration
	orderEventHandler, err := initWarehouseIntegration(db, eventBus)
	if err != nil {
		return nil, err
	}
	app.OrderEventHandler = orderEventHandler

	// Initialize Health Checker
	healthChecker := health.NewChecker(db, redisClient, eventBus, cfg.App.Version)
	app.HealthChecker = healthChecker

	return app, nil
}

// initLogger initializes the application logger
func initLogger(cfg *config.AppConfig) error {
	logFormat := "text"
	if cfg.App.Environment == "production" {
		logFormat = "json"
	}

	logLevel := "info"
	if cfg.App.Environment == "development" {
		logLevel = "debug"
	}

	logger.Init(logger.Config{
		Level:      logLevel,
		Format:     logFormat,
		AddSource:  cfg.App.Environment == "development",
		TimeFormat: time.RFC3339,
	})

	return nil
}

// initDatabase initializes database connection (PostgreSQL or SQLite)
func initDatabase(cfg *config.AppConfig) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	switch cfg.Database.Driver {
	case "postgres":
		db, err = database.NewPostgresConnection(&cfg.Database.Postgres)
		if err != nil {
			logger.Fatal("Failed to connect to PostgreSQL", slog.Any("error", err))
			return nil, err
		}
		logger.Info("PostgreSQL connected successfully")

	case "sqlite":
		db, err = database.NewSQLiteConnection(&cfg.Database.SQLite)
		if err != nil {
			logger.Fatal("Failed to connect to SQLite", slog.Any("error", err))
			return nil, err
		}
		logger.Info("SQLite connected successfully", slog.String("path", cfg.Database.SQLite.Path))

	default:
		logger.Fatal("Unsupported database driver",
			slog.String("driver", cfg.Database.Driver),
			slog.String("supported", "postgres, sqlite"),
		)
		return nil, err
	}

	return db, nil
}

// runMigrations runs all database migrations
func runMigrations(db *sqlx.DB) error {
	logger.Info("Running database migrations...")
	migrationManager := migration.NewManager(db, "migrations")
	ctx := context.Background()

	namespaces := []string{"core", "shared", "identity", "customer-mgmt", "order-mgmt", "billing", "warehouse"}
	for _, ns := range namespaces {
		if err := migrationManager.MigrateNamespace(ctx, ns); err != nil {
			logger.Fatal("Failed to run migrations",
				slog.String("namespace", ns),
				slog.Any("error", err),
			)
			return err
		}
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// initRedis initializes Redis client (for token revocation, bus, and cache)
func initRedis(cfg *config.AppConfig) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:       cfg.Database.Redis.Addr,
		Password:   cfg.Database.Redis.Password,
		DB:         cfg.Database.Redis.Databases.Revocation,
		PoolSize:   cfg.Database.Redis.PoolSize,
		MaxRetries: cfg.Database.Redis.MaxRetries,
	})

	// Test connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Warn("Redis connection failed, some features will be disabled", slog.Any("error", err))
		return nil // Graceful degradation
	}

	logger.Info("Redis connected successfully",
		slog.String("addr", cfg.Database.Redis.Addr),
		slog.Int("revocation_db", cfg.Database.Redis.Databases.Revocation),
	)

	return redisClient
}

// initCache initializes cache layer
func initCache(cfg *config.AppConfig, redisClient *redis.Client) (cache.ICache, error) {
	if redisClient == nil {
		// Fallback to no-op cache when Redis unavailable
		cacheClient, _ := cache.NewCache(&cache.Config{Enabled: false, Adapter: "noop"}, nil)
		logger.Warn("Cache disabled (Redis unavailable)")
		return cacheClient, nil
	}

	// Parse cache config
	cacheConfig, err := cfg.Cache.ToCacheConfig()
	if err != nil {
		logger.Fatal("Failed to parse cache config", slog.Any("error", err))
		return nil, err
	}

	// Create Redis client for cache (separate DB)
	cacheRedisClient := redis.NewClient(&redis.Options{
		Addr:       cfg.Database.Redis.Addr,
		Password:   cfg.Database.Redis.Password,
		DB:         cfg.Database.Redis.Databases.Cache,
		PoolSize:   cfg.Database.Redis.PoolSize,
		MaxRetries: cfg.Database.Redis.MaxRetries,
	})

	cacheClient, err := cache.NewCache(cacheConfig, cacheRedisClient)
	if err != nil {
		logger.Fatal("Failed to initialize cache", slog.Any("error", err))
		return nil, err
	}

	logger.Info("Cache initialized",
		slog.String("adapter", cacheConfig.Adapter),
		slog.String("prefix", cacheConfig.Prefix),
		slog.Bool("enabled", cacheConfig.Enabled),
	)

	return cacheClient, nil
}

// initAuth initializes JWT Manager and Token Revoker
func initAuth(cfg *config.AppConfig, redisClient *redis.Client) (*jwt.Manager, *jwt.TokenRevoker) {
	// Initialize JWT Manager
	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            cfg.JWT.Secret,
		AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
		RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
		Issuer:               cfg.JWT.Issuer,
	})

	logger.Info("JWT Manager initialized",
		slog.String("issuer", cfg.JWT.Issuer),
		slog.Duration("access_token_duration", cfg.JWT.AccessTokenDuration),
		slog.Duration("refresh_token_duration", cfg.JWT.RefreshTokenDuration),
	)

	// Initialize Token Revoker (if Redis is available)
	var tokenRevoker *jwt.TokenRevoker
	if redisClient != nil {
		tokenRevoker = jwt.NewTokenRevoker(redisClient)
		logger.Info("Token Revoker initialized with Redis")
	} else {
		logger.Warn("Token Revoker disabled (Redis unavailable)")
	}

	return jwtManager, tokenRevoker
}

// initEventBus initializes Event Bus
func initEventBus(cfg *config.AppConfig) (bus.IBus, error) {
	eventBus, err := bus.NewBus(cfg.Bus, cfg.Database.Redis)
	if err != nil {
		logger.Fatal("Failed to initialize event bus", slog.Any("error", err))
		return nil, err
	}

	logger.Info("Event Bus initialized",
		slog.String("adapter", cfg.Bus.Adapter),
		slog.Int("worker_pool_size", cfg.Bus.WorkerPoolSize),
	)

	return eventBus, nil
}

// initWarehouseIntegration initializes Warehouse Integration (ReservationService + OrderEventHandler)
func initWarehouseIntegration(db *sqlx.DB, eventBus bus.IBus) (*integration.OrderEventHandler, error) {
	// Initialize Inventory Use Case
	invRepo := inventoryRepo.NewInventoryRepository(db)
	inventoryUC := inventory.NewUseCase(invRepo)

	// Initialize Reservation Service
	reservationService := integration.NewReservationService(inventoryUC)

	// Initialize Order Event Handler
	orderEventHandler := integration.NewOrderEventHandler(reservationService)

	// Register event handlers
	if err := orderEventHandler.RegisterHandlers(eventBus); err != nil {
		logger.Fatal("Failed to register order event handlers", slog.Any("error", err))
		return nil, err
	}

	logger.Info("Warehouse Integration initialized",
		slog.String("component", "ReservationService + OrderEventHandler"),
		slog.Int("event_handlers", 3), // order.confirmed, order.cancelled, order.fulfilled
	)

	return orderEventHandler, nil
}

// Close gracefully closes all application dependencies
func (app *App) Close() {
	ctx := context.Background()

	// Close event bus first
	if app.EventBus != nil {
		closeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := app.EventBus.Close(closeCtx); err != nil {
			logger.Error("Failed to close event bus", slog.Any("error", err))
		}
	}

	// Close cache (has its own Redis client)
	if app.CacheClient != nil {
		if err := app.CacheClient.Close(ctx); err != nil {
			// Ignore "client is closed" error as cache closes its own Redis client
			if err.Error() != "redis: client is closed" {
				logger.Error("Failed to close cache client", slog.Any("error", err))
			}
		}
	}

	// Close Redis (for JWT revocation)
	if app.RedisClient != nil {
		if err := app.RedisClient.Close(); err != nil {
			// Ignore "client is closed" error if already closed by cache
			if err.Error() != "redis: client is closed" {
				logger.Error("Failed to close Redis client", slog.Any("error", err))
			}
		}
	}

	// Close database last
	if app.DB != nil {
		if err := app.DB.Close(); err != nil {
			logger.Error("Failed to close database connection", slog.Any("error", err))
		}
	}
}
