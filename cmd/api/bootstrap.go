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
	"github.com/basilex/promenade/pkg/fiscal/checkbox"
	"github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/migration"
	"github.com/basilex/promenade/pkg/scheduler"

	"github.com/basilex/promenade/internal/contexts/warehouse/integration"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	inventoryUseCase "github.com/basilex/promenade/internal/contexts/warehouse/inventory/usecase"

	analyticsHandler "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/adapter/http"
	salesReportRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/adapter/repository/postgres"
	analyticsIntegration "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/integration"
	analyticsUseCase "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/usecase"

	bankAccountRepo "github.com/basilex/promenade/internal/contexts/banking/bankaccount/adapter/repository/postgres"
	bankAccountUseCase "github.com/basilex/promenade/internal/contexts/banking/bankaccount/usecase"
	bankTransactionRepo "github.com/basilex/promenade/internal/contexts/banking/banktransaction/adapter/repository/postgres"
	bankTransactionUseCase "github.com/basilex/promenade/internal/contexts/banking/banktransaction/usecase"

	accountingAudit "github.com/basilex/promenade/internal/contexts/accounting/audit"
	accountRepo "github.com/basilex/promenade/internal/contexts/accounting/account/adapter/repository/postgres"
	accountCache "github.com/basilex/promenade/internal/contexts/accounting/account/cache"
	accountUseCase "github.com/basilex/promenade/internal/contexts/accounting/account/usecase"
	journalEntryRepo "github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/repository/postgres"
	journalEntryUseCase "github.com/basilex/promenade/internal/contexts/accounting/journalentry/usecase"
	fiscalPeriodRepo "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/adapter/repository/postgres"
	fiscalPeriodCache "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/cache"
	fiscalPeriodUseCase "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/usecase"
	taxCodeRepo "github.com/basilex/promenade/internal/contexts/accounting/taxcode/adapter/repository/postgres"
	taxCodeCache "github.com/basilex/promenade/internal/contexts/accounting/taxcode/cache"
	taxCodeUseCase "github.com/basilex/promenade/internal/contexts/accounting/taxcode/usecase"
	budgetRepo "github.com/basilex/promenade/internal/contexts/accounting/budget/adapter/repository/postgres"
	budgetUseCase "github.com/basilex/promenade/internal/contexts/accounting/budget/usecase"
	costCenterRepo "github.com/basilex/promenade/internal/contexts/accounting/costcenter/adapter/repository/postgres"
	costCenterUseCase "github.com/basilex/promenade/internal/contexts/accounting/costcenter/usecase"
	reconciliationRepo "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/adapter/repository/postgres"
	reconciliationUseCase "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/usecase"
	accountingIntegration "github.com/basilex/promenade/internal/contexts/accounting/integration"

	cashregisterRepo "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/repository/postgres"
	fiscalIntegration "github.com/basilex/promenade/internal/contexts/fiscal/integration"
	receiptPrinter "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/printer"
	receiptRepo "github.com/basilex/promenade/internal/contexts/fiscal/receipt/adapter/repository/postgres"
	receiptUseCase "github.com/basilex/promenade/internal/contexts/fiscal/receipt/usecase"
)

// App holds all application dependencies
type App struct {
	DB                      *sqlx.DB
	Config                  *config.AppConfig
	RedisClient             *redis.Client
	CacheClient             cache.ICache
	JWTManager              *jwt.Manager
	TokenRevoker            *jwt.TokenRevoker
	EventBus                bus.IBus
	HealthChecker           *health.Checker
	OrderEventHandler       *integration.OrderEventHandler
	FiscalOrderEventHandler *fiscalIntegration.OrderEventHandler
	SalesReportEventHandler *analyticsIntegration.SalesReportEventHandler
	SalesReportHandler        *analyticsHandler.SalesReportHandler
	BankAccountUseCase        bankAccountUseCase.IBankAccountUseCase
	BankTransactionUseCase    bankTransactionUseCase.IBankTransactionUseCase
	AccountUseCase            accountUseCase.IAccountUseCase
	JournalEntryUseCase       journalEntryUseCase.IJournalEntryUseCase
	FiscalPeriodUseCase       fiscalPeriodUseCase.IFiscalPeriodUseCase
	TaxCodeUseCase            taxCodeUseCase.ITaxCodeUseCase
	BudgetUseCase             budgetUseCase.IBudgetUseCase
	CostCenterUseCase    costCenterUseCase.ICostCenterUseCase
	ReconciliationUseCase reconciliationUseCase.IReconciliationUseCase
	AccountingEventHandler *accountingIntegration.AccountingEventHandler
	Scheduler                 *scheduler.Engine
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

	// Initialize Fiscal Integration
	fiscalOrderEventHandler, receiptUC, printerEnabled, checkboxClient, err := initFiscalIntegration(db, eventBus, cfg)
	if err != nil {
		return nil, err
	}
	app.FiscalOrderEventHandler = fiscalOrderEventHandler

	// Initialize Sales Report Analytics (event handler + query handler)
	salesReportEventHandler, salesReportHandler, err := initSalesReportAnalytics(db, eventBus)
	if err != nil {
		return nil, err
	}
	app.SalesReportEventHandler = salesReportEventHandler
	app.SalesReportHandler = salesReportHandler

	// Initialize Banking Context
	bankAccountUC, bankTransactionUC, err := initBanking(db)
	if err != nil {
		return nil, err
	}
	app.BankAccountUseCase = bankAccountUC
	app.BankTransactionUseCase = bankTransactionUC

	// Initialize Accounting Context
	accountUC, journalEntryUC, fiscalPeriodUC, taxCodeUC, budgetUC, costCenterUC, reconciliationUC, accountingEventHandler, err := initAccounting(db, eventBus)
	if err != nil {
		return nil, err
	}
	app.AccountUseCase = accountUC
	app.JournalEntryUseCase = journalEntryUC
	app.FiscalPeriodUseCase = fiscalPeriodUC
	app.TaxCodeUseCase = taxCodeUC
	app.BudgetUseCase = budgetUC
	app.CostCenterUseCase = costCenterUC
	app.ReconciliationUseCase = reconciliationUC
	app.AccountingEventHandler = accountingEventHandler

	// Initialize Scheduler (Fiscal retries, etc.)
	schedulerEngine, err := initScheduler(cfg, db, receiptUC, printerEnabled, checkboxClient)
	if err != nil {
		return nil, err
	}
	app.Scheduler = schedulerEngine

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

	namespaces := []string{"core", "shared", "identity", "customer-mgmt", "order-mgmt", "billing", "banking", "accounting", "warehouse", "scripting", "fiscal", "ui"}
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
	inventoryUC := inventoryUseCase.NewInventoryUseCase(invRepo)

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

// initFiscalIntegration initializes Fiscal Integration (Order → Receipt auto-print)
func initFiscalIntegration(db *sqlx.DB, eventBus bus.IBus, cfg *config.AppConfig) (*fiscalIntegration.OrderEventHandler, receiptUseCase.IReceiptUseCase, bool, *checkbox.Client, error) {
	cashRegisterRepository := cashregisterRepo.NewCashRegisterRepository(db)
	receiptRepository := receiptRepo.NewReceiptRepository(db)

	var printer receiptUseCase.IPrinter
	printerEnabled := false

	var pdfPrinter receiptUseCase.IPrinter
	if cfg.Fiscal.PDFOutputDir != "" {
		pdfPrinter = receiptPrinter.NewPDFPrinter(cfg.Fiscal.PDFOutputDir)
	}

	var checkboxClient *checkbox.Client
	checkboxCfg := cfg.Fiscal.Checkbox
	if checkboxCfg.APIKey != "" {
		checkboxClient = checkbox.NewClient(&checkbox.Config{
			APIKey:  checkboxCfg.APIKey,
			Sandbox: checkboxCfg.Sandbox,
			Timeout: checkboxCfg.Timeout,
		})
	}

	if checkboxClient != nil {
		checkboxPrinter := receiptPrinter.NewCheckboxPrinter(checkboxClient)
		if pdfPrinter != nil {
			printer = receiptUseCase.NewMultiPrinter(checkboxPrinter, pdfPrinter)
		} else {
			printer = checkboxPrinter
		}
	} else if pdfPrinter != nil {
		printer = pdfPrinter
	}

	if printer != nil {
		printerEnabled = true
	}

	receiptUseCase := receiptUseCase.NewReceiptUseCase(receiptRepository, printer)
	orderEventHandler := fiscalIntegration.NewOrderEventHandler(receiptUseCase, cashRegisterRepository, printerEnabled)

	if err := orderEventHandler.RegisterHandlers(eventBus); err != nil {
		logger.Fatal("Failed to register fiscal order event handlers", slog.Any("error", err))
		return nil, nil, false, nil, err
	}

	logger.Info("Fiscal Integration initialized",
		slog.String("component", "OrderEventHandler"),
		slog.Bool("printer_enabled", printerEnabled),
	)

	return orderEventHandler, receiptUseCase, printerEnabled, checkboxClient, nil
}

// initSalesReportAnalytics initializes sales report analytics (event handler + query handler)
func initSalesReportAnalytics(db *sqlx.DB, eventBus bus.IBus) (*analyticsIntegration.SalesReportEventHandler, *analyticsHandler.SalesReportHandler, error) {
	repo := salesReportRepo.NewSalesReportRepository(db)
	txManager := database.NewTransactionManager(db)

	// Event handler (write side - materializes read model)
	eventHandler := analyticsIntegration.NewSalesReportEventHandler(repo, txManager)
	if err := eventHandler.RegisterHandlers(eventBus); err != nil {
		logger.Fatal("Failed to register sales report event handlers", slog.Any("error", err))
		return nil, nil, err
	}

	// Query handler (read side - serves HTTP requests)
	useCase := analyticsUseCase.NewSalesReportUseCase(repo)
	httpHandler := analyticsHandler.NewSalesReportHandler(useCase)

	logger.Info("Sales Report Analytics initialized",
		slog.String("component", "SalesReportEventHandler + HTTP Handler"),
		slog.Int("event_handlers", 3),
	)

	return eventHandler, httpHandler, nil
}

// initScheduler initializes scheduler engine and registers fiscal retry jobs
func initScheduler(cfg *config.AppConfig, db *sqlx.DB, receiptUC receiptUseCase.IReceiptUseCase, printerEnabled bool, checkboxClient *checkbox.Client) (*scheduler.Engine, error) {
	if !cfg.Scheduler.Enabled {
		logger.Info("Scheduler disabled")
		return nil, nil
	}

	engine, err := scheduler.NewEngine(cfg.Scheduler)
	if err != nil {
		logger.Fatal("Failed to initialize scheduler", slog.Any("error", err))
		return nil, err
	}

	// Start engine before registering jobs
	if err := engine.Start(); err != nil {
		logger.Fatal("Failed to start scheduler", slog.Any("error", err))
		return nil, err
	}

	if receiptUC != nil && printerEnabled {
		if err := fiscalIntegration.RegisterReceiptRetryJob(engine, receiptUC, cfg.Fiscal.RetryCron); err != nil {
			logger.Fatal("Failed to register fiscal receipt retry job", slog.Any("error", err))
			return nil, err
		}
	} else {
		logger.Info("Fiscal receipt retry job not registered (printer disabled)")
	}

	if checkboxClient != nil {
		cashRegisterRepository := cashregisterRepo.NewCashRegisterRepository(db)
		if err := fiscalIntegration.RegisterShiftJobs(engine, cashRegisterRepository, checkboxClient, cfg.Fiscal.ShiftOpenCron, cfg.Fiscal.ShiftCloseCron); err != nil {
			logger.Fatal("Failed to register fiscal shift jobs", slog.Any("error", err))
			return nil, err
		}
	} else {
		logger.Info("Shift jobs not registered (checkbox client not configured)")
	}

	logger.Info("Scheduler initialized")
	return engine, nil
}

// initBanking initializes Banking Context (Bank Accounts + Bank Transactions)
func initBanking(db *sqlx.DB) (bankAccountUseCase.IBankAccountUseCase, bankTransactionUseCase.IBankTransactionUseCase, error) {
	// Initialize Bank Account Use Case
	bankAccountRepository := bankAccountRepo.NewBankAccountRepository(db)
	bankAccountUC := bankAccountUseCase.NewBankAccountUseCase(bankAccountRepository)

	// Initialize Bank Transaction Use Case
	bankTransactionRepository := bankTransactionRepo.NewBankTransactionRepository(db)
	bankTransactionUC := bankTransactionUseCase.NewBankTransactionUseCase(bankTransactionRepository)

	logger.Info("Banking Context initialized",
		slog.String("component", "BankAccount + BankTransaction"),
	)

	return bankAccountUC, bankTransactionUC, nil
}

// initAccounting initializes Accounting Context (7 aggregates + event handlers)
func initAccounting(db *sqlx.DB, eventBus bus.IBus) (
	accountUseCase.IAccountUseCase,
	journalEntryUseCase.IJournalEntryUseCase,
	fiscalPeriodUseCase.IFiscalPeriodUseCase,
	taxCodeUseCase.ITaxCodeUseCase,
	budgetUseCase.IBudgetUseCase,
	costCenterUseCase.ICostCenterUseCase,
	reconciliationUseCase.IReconciliationUseCase,
	*accountingIntegration.AccountingEventHandler,
	error,
) {
	// Initialize Audit Logger
	auditLogger := accountingAudit.NewAuditLogger(db)

	// Initialize Account Use Case with Cache
	accountRepository := accountRepo.NewAccountRepository(db)
	accountCacheInstance := accountCache.NewAccountCache(accountRepository)
	accountUC := accountUseCase.NewAccountUseCase(accountRepository, accountCacheInstance, auditLogger)

	// Initialize Fiscal Period Use Case with Cache
	fiscalPeriodRepository := fiscalPeriodRepo.NewFiscalPeriodRepository(db)
	fiscalPeriodCacheInstance := fiscalPeriodCache.NewFiscalPeriodCache(fiscalPeriodRepository)
	fiscalPeriodUC := fiscalPeriodUseCase.NewFiscalPeriodUseCase(fiscalPeriodRepository, fiscalPeriodCacheInstance, auditLogger)

	// Initialize Tax Code Use Case with Cache
	taxCodeRepository := taxCodeRepo.NewTaxCodeRepository(db)
	taxCodeCacheInstance := taxCodeCache.NewTaxCodeCache(taxCodeRepository)
	taxCodeUC := taxCodeUseCase.NewTaxCodeUseCase(taxCodeRepository, taxCodeCacheInstance, auditLogger)

	// Initialize Journal Entry Use Case with Event Store
	journalEntryRepository := journalEntryRepo.NewJournalEntryRepository(db)
	eventStore := accountingAudit.NewEventStore(db)
	journalEntryUC := journalEntryUseCase.NewJournalEntryUseCase(journalEntryRepository, eventStore, auditLogger)

	// Initialize Budget Use Case
	budgetRepository := budgetRepo.NewBudgetRepository(db)
	budgetUC := budgetUseCase.NewBudgetUseCase(budgetRepository, auditLogger)

	// Initialize Cost Center Use Case
	costCenterRepository := costCenterRepo.NewCostCenterRepository(db)
	costCenterUC := costCenterUseCase.NewCostCenterUseCase(costCenterRepository, auditLogger)

	// Initialize Reconciliation Use Case
	reconciliationRepository := reconciliationRepo.NewReconciliationRepository(db)
	reconciliationUC := reconciliationUseCase.NewReconciliationUseCase(
		reconciliationRepository,
		auditLogger,
	)

	// Initialize Accounting Event Handler (for bank/billing/fiscal integration)
	accountingEventHandler := accountingIntegration.NewAccountingEventHandler(
		journalEntryUC,
		accountUC,
		fiscalPeriodUC,
	)

	// Register event handlers
	if err := accountingEventHandler.RegisterHandlers(eventBus); err != nil {
		logger.Fatal("Failed to register accounting event handlers", slog.Any("error", err))
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}

	logger.Info("Accounting Context initialized",
		slog.String("component", "7 Aggregates + Event Integration"),
		slog.Int("use_cases", 7),
		slog.Int("caches", 3),
		slog.Bool("audit_enabled", true),
		slog.Bool("event_sourcing_enabled", true),
	)

	return accountUC, journalEntryUC, fiscalPeriodUC, taxCodeUC, budgetUC, costCenterUC, reconciliationUC, accountingEventHandler, nil
}

// Close gracefully closes all application dependencies
func (app *App) Close() {
	ctx := context.Background()

	// Stop scheduler before closing dependencies
	if app.Scheduler != nil {
		if err := app.Scheduler.Stop(); err != nil {
			if err != scheduler.ErrEngineNotRunning {
				logger.Error("Failed to stop scheduler", slog.Any("error", err))
			}
		}
	}

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
