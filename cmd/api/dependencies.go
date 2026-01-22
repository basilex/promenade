package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/cache/noop"
	"github.com/basilex/promenade/pkg/logger"

	// Warehouse Context
	"github.com/basilex/promenade/internal/contexts/warehouse/integration"
	inventoryRepo "github.com/basilex/promenade/internal/contexts/warehouse/inventory/adapter/repository/postgres"
	inventoryUseCase "github.com/basilex/promenade/internal/contexts/warehouse/inventory/usecase"

	// Analytics (Customer-Mgmt Context)
	analyticsHandler "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/adapter/http"
	salesReportRepo "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/adapter/repository/postgres"
	analyticsIntegration "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/integration"
	analyticsUseCase "github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/usecase"

	// Banking Context
	bankAccountRepo "github.com/basilex/promenade/internal/contexts/banking/bankaccount/adapter/repository/postgres"
	bankAccountUseCase "github.com/basilex/promenade/internal/contexts/banking/bankaccount/usecase"
	bankTransactionRepo "github.com/basilex/promenade/internal/contexts/banking/banktransaction/adapter/repository/postgres"
	bankTransactionUseCase "github.com/basilex/promenade/internal/contexts/banking/banktransaction/usecase"

	// Accounting Context
	accountRepo "github.com/basilex/promenade/internal/contexts/accounting/account/adapter/repository/postgres"
	accountCache "github.com/basilex/promenade/internal/contexts/accounting/account/cache"
	accountUseCase "github.com/basilex/promenade/internal/contexts/accounting/account/usecase"
	accountingAudit "github.com/basilex/promenade/internal/contexts/accounting/audit"
	budgetRepo "github.com/basilex/promenade/internal/contexts/accounting/budget/adapter/repository/postgres"
	budgetUseCase "github.com/basilex/promenade/internal/contexts/accounting/budget/usecase"
	costCenterRepo "github.com/basilex/promenade/internal/contexts/accounting/costcenter/adapter/repository/postgres"
	costCenterUseCase "github.com/basilex/promenade/internal/contexts/accounting/costcenter/usecase"
	fiscalPeriodRepo "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/adapter/repository/postgres"
	fiscalPeriodCache "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/cache"
	fiscalPeriodUseCase "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/usecase"
	accountingIntegration "github.com/basilex/promenade/internal/contexts/accounting/integration"
	journalEntryRepo "github.com/basilex/promenade/internal/contexts/accounting/journalentry/adapter/repository/postgres"
	journalEntryUseCase "github.com/basilex/promenade/internal/contexts/accounting/journalentry/usecase"
	reconciliationRepo "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/adapter/repository/postgres"
	reconciliationUseCase "github.com/basilex/promenade/internal/contexts/accounting/reconciliation/usecase"
	taxCodeRepo "github.com/basilex/promenade/internal/contexts/accounting/taxcode/adapter/repository/postgres"
	taxCodeCache "github.com/basilex/promenade/internal/contexts/accounting/taxcode/cache"
	taxCodeUseCase "github.com/basilex/promenade/internal/contexts/accounting/taxcode/usecase"

	// Shared Context (Reference Data)
	countryRepositoryFactory "github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository"
	countryUseCase "github.com/basilex/promenade/internal/contexts/shared/country/usecase"
)

// Dependencies holds all application domain dependencies (use cases, event handlers)
type Dependencies struct {
	// Shared Context (Reference Data)
	Shared *SharedUseCases

	// Banking Context
	Banking *BankingUseCases

	// Accounting Context
	Accounting *AccountingUseCases

	// Warehouse Context
	Warehouse *WarehouseIntegration

	// Analytics Context
	Analytics *AnalyticsComponents
}

// SharedUseCases holds all Shared Context use cases
type SharedUseCases struct {
	Country countryUseCase.ICountryUseCase
	// Currency, Language, Timezone will be added when implemented
}

// BankingUseCases holds all Banking Context use cases
type BankingUseCases struct {
	BankAccount     bankAccountUseCase.IBankAccountUseCase
	BankTransaction bankTransactionUseCase.IBankTransactionUseCase
}

// AccountingUseCases holds all Accounting Context use cases and event handlers
type AccountingUseCases struct {
	Account        accountUseCase.IAccountUseCase
	JournalEntry   journalEntryUseCase.IJournalEntryUseCase
	FiscalPeriod   fiscalPeriodUseCase.IFiscalPeriodUseCase
	TaxCode        taxCodeUseCase.ITaxCodeUseCase
	Budget         budgetUseCase.IBudgetUseCase
	CostCenter     costCenterUseCase.ICostCenterUseCase
	Reconciliation reconciliationUseCase.IReconciliationUseCase
	EventHandler   *accountingIntegration.AccountingEventHandler
}

// WarehouseIntegration holds Warehouse Integration components
type WarehouseIntegration struct {
	OrderEventHandler *integration.OrderEventHandler
}

// AnalyticsComponents holds Analytics components (event handler + query handler)
type AnalyticsComponents struct {
	EventHandler *analyticsIntegration.SalesReportEventHandler
	HTTPHandler  *analyticsHandler.SalesReportHandler
}

// InitRepositories initializes all domain dependencies in one place
// This centralizes all repository/use case initialization, keeping bootstrap.go clean
func InitRepositories(db *sqlx.DB, cfg *config.AppConfig, eventBus bus.IBus, cacheClient cache.ICache) (*Dependencies, error) {
	deps := &Dependencies{}

	// Initialize Shared Context (Reference Data)
	shared, err := initSharedContext(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize shared context: %w", err)
	}
	deps.Shared = shared

	// Initialize Banking Context
	banking, err := initBankingContext(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize banking context: %w", err)
	}
	deps.Banking = banking

	// Initialize Accounting Context
	accounting, err := initAccountingContext(db, eventBus)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize accounting context: %w", err)
	}
	deps.Accounting = accounting

	// Initialize Warehouse Integration
	warehouse, err := initWarehouseContext(db, eventBus)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize warehouse integration: %w", err)
	}
	deps.Warehouse = warehouse

	// Initialize Analytics
	analytics, err := initAnalyticsContext(db, eventBus)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize analytics: %w", err)
	}
	deps.Analytics = analytics

	logger.Info("All domain dependencies initialized successfully")
	return deps, nil
}

// initSharedContext initializes Shared Context with all reference data aggregates
func initSharedContext(db *sqlx.DB, cfg *config.AppConfig) (*SharedUseCases, error) {
	driver := cfg.Database.Driver

	// Initialize Country aggregate (multi-database support via factory)
	countryRepo, err := countryRepositoryFactory.NewCountryRepository(db, driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create country repository: %w", err)
	}
	countryUC := countryUseCase.NewCountryUseCase(countryRepo, noop.NewNoOpCache())

	// TODO: Initialize Currency, Language, Timezone aggregates when factories are ready
	// currencyRepo, err := currencyRepositoryFactory.NewCurrencyRepository(db, driver)
	// languageRepo, err := languageRepositoryFactory.NewLanguageRepository(db, driver)
	// timezoneRepo, err := timezoneRepositoryFactory.NewTimezoneRepository(db, driver)

	logger.Info("Shared Context initialized",
		"driver", driver,
		"aggregates", 1, // Will be 4 when all are implemented
	)

	return &SharedUseCases{
		Country: countryUC,
		// Currency: currencyUC,
		// Language: languageUC,
		// Timezone: timezoneUC,
	}, nil
}

// initBankingContext initializes Banking Context (Bank Accounts + Bank Transactions)
func initBankingContext(db *sqlx.DB) (*BankingUseCases, error) {
	// Initialize Bank Account Use Case
	bankAccountRepository := bankAccountRepo.NewBankAccountRepository(db)
	bankAccountUC := bankAccountUseCase.NewBankAccountUseCase(bankAccountRepository)

	// Initialize Bank Transaction Use Case
	bankTransactionRepository := bankTransactionRepo.NewBankTransactionRepository(db)
	bankTransactionUC := bankTransactionUseCase.NewBankTransactionUseCase(bankTransactionRepository)

	logger.Info("Banking Context initialized",
		"aggregates", 2,
	)

	return &BankingUseCases{
		BankAccount:     bankAccountUC,
		BankTransaction: bankTransactionUC,
	}, nil
}

// initAccountingContext initializes Accounting Context (7 aggregates + event handlers)
func initAccountingContext(db *sqlx.DB, eventBus bus.IBus) (*AccountingUseCases, error) {
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
	reconciliationUC := reconciliationUseCase.NewReconciliationUseCase(reconciliationRepository, auditLogger)

	// Initialize Accounting Event Handler (for bank/billing/fiscal integration)
	accountingEventHandler := accountingIntegration.NewAccountingEventHandler(
		journalEntryUC,
		accountUC,
		fiscalPeriodUC,
	)

	// Register event handlers
	if err := accountingEventHandler.RegisterHandlers(eventBus); err != nil {
		return nil, fmt.Errorf("failed to register accounting event handlers: %w", err)
	}

	logger.Info("Accounting Context initialized",
		"aggregates", 7,
		"caches", 3,
		"audit_enabled", true,
		"event_sourcing_enabled", true,
	)

	return &AccountingUseCases{
		Account:        accountUC,
		JournalEntry:   journalEntryUC,
		FiscalPeriod:   fiscalPeriodUC,
		TaxCode:        taxCodeUC,
		Budget:         budgetUC,
		CostCenter:     costCenterUC,
		Reconciliation: reconciliationUC,
		EventHandler:   accountingEventHandler,
	}, nil
}

// initWarehouseContext initializes Warehouse Integration (ReservationService + OrderEventHandler)
func initWarehouseContext(db *sqlx.DB, eventBus bus.IBus) (*WarehouseIntegration, error) {
	// Initialize Inventory Use Case
	invRepo := inventoryRepo.NewInventoryRepository(db)
	inventoryUC := inventoryUseCase.NewInventoryUseCase(invRepo)

	// Initialize Reservation Service
	reservationService := integration.NewReservationService(inventoryUC)

	// Initialize Order Event Handler
	orderEventHandler := integration.NewOrderEventHandler(reservationService)

	// Register event handlers
	if err := orderEventHandler.RegisterHandlers(eventBus); err != nil {
		return nil, fmt.Errorf("failed to register order event handlers: %w", err)
	}

	logger.Info("Warehouse Integration initialized",
		"event_handlers", 3, // order.confirmed, order.cancelled, order.fulfilled
	)

	return &WarehouseIntegration{
		OrderEventHandler: orderEventHandler,
	}, nil
}

// initAnalyticsContext initializes sales report analytics (event handler + query handler)
func initAnalyticsContext(db *sqlx.DB, eventBus bus.IBus) (*AnalyticsComponents, error) {
	repo := salesReportRepo.NewSalesReportRepository(db)
	txManager := database.NewTransactionManager(db)

	// Event handler (write side - materializes read model)
	eventHandler := analyticsIntegration.NewSalesReportEventHandler(repo, txManager)
	if err := eventHandler.RegisterHandlers(eventBus); err != nil {
		return nil, fmt.Errorf("failed to register sales report event handlers: %w", err)
	}

	// Query handler (read side - serves HTTP requests)
	useCase := analyticsUseCase.NewSalesReportUseCase(repo)
	httpHandler := analyticsHandler.NewSalesReportHandler(useCase)

	logger.Info("Sales Report Analytics initialized",
		"event_handlers", 3,
	)

	return &AnalyticsComponents{
		EventHandler: eventHandler,
		HTTPHandler:  httpHandler,
	}, nil
}
