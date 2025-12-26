package billing

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/modules/billing/adapter/http/handler"
	"github.com/basilex/promenade/internal/modules/billing/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/billing/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/module"
	moduleconfig "github.com/basilex/promenade/pkg/module/config"
	"github.com/basilex/promenade/pkg/purge"
)

// BillingModule implements module.IModule for billing functionality
type BillingModule struct {
	db                  *sqlx.DB
	eventBus            bus.IBus
	config              *moduleconfig.Config // Module's own config
	planUseCase         usecase.IPlanUseCase
	planHandler         *handler.PlanHandler
	subscriptionUseCase usecase.ISubscriptionUseCase
	subscriptionHandler *handler.SubscriptionHandler
	invoiceUseCase      usecase.IInvoiceUseCase
	invoiceHandler      *handler.InvoiceHandler
	paymentUseCase      usecase.IPaymentUseCase
	paymentHandler      *handler.PaymentHandler
}

// New creates a new billing module instance
func New() module.IModule {
	return &BillingModule{}
}

// Metadata returns module information
func (m *BillingModule) Metadata() module.Metadata {
	return module.Metadata{
		Name:        "billing",
		DisplayName: "Billing Management",
		Version:     "1.0.0",
		Description: "Complete billing solution: plans, subscriptions, invoices, payments",
		Author:      "Promenade Team",
		License:     "Commercial",
		Tags:        []string{"billing", "subscription", "payment", "invoice", "commercial"},
	}
}

// Dependencies returns list of module dependencies
func (m *BillingModule) Dependencies() []string {
	return []string{} // No dependencies
}

// Initialize sets up the module components
func (m *BillingModule) Initialize(ctx context.Context, core *module.Core) error {
	slog.Info("Initializing billing module")

	m.db = core.DB
	m.eventBus = core.EventBus

	// Load module's own configuration
	cfg, err := moduleconfig.Load("internal/modules/billing/config", "promenade")
	if err != nil {
		slog.Warn("Failed to load billing module config, using defaults", "error", err)
		// Continue with defaults
	} else {
		m.config = cfg
		slog.Info("Billing module config loaded",
			"version", cfg.Module.Version,
			"enabled", cfg.Module.Enabled,
		)
	}

	// Get retention policies from config (for purge system)
	// These MUST be configured - no fallbacks (legal compliance)
	if m.config == nil {
		return fmt.Errorf("billing config not loaded - cannot initialize module")
	}

	// Read purge retention days from config (required)
	invoicesRetentionDays := m.config.GetRetentionDays("billing_invoices", 0)
	paymentsRetentionDays := m.config.GetRetentionDays("billing_payments", 0)
	subscriptionsRetentionDays := m.config.GetRetentionDays("billing_subscriptions", 0)

	if invoicesRetentionDays == 0 || paymentsRetentionDays == 0 || subscriptionsRetentionDays == 0 {
		return fmt.Errorf("billing retention policies not configured - check config/billing/config.*.yaml")
	}

	slog.Info("Billing purge policies configured",
		"invoices_retention_days", invoicesRetentionDays,
		"payments_retention_days", paymentsRetentionDays,
		"subscriptions_retention_days", subscriptionsRetentionDays,
	)

	// Initialize repositories
	planRepo := postgres.NewPlanRepository(m.db)
	subscriptionRepo := postgres.NewSubscriptionRepository(m.db)
	invoiceRepo := postgres.NewInvoiceRepository(m.db)
	paymentRepo := postgres.NewPaymentRepository(m.db)

	// Initialize use cases
	m.planUseCase = usecase.NewPlanUseCase(planRepo, subscriptionRepo, m.eventBus)
	m.subscriptionUseCase = usecase.NewSubscriptionUseCase(
		subscriptionRepo,
		planRepo,
		invoiceRepo,
		m.eventBus,
	)
	m.invoiceUseCase = usecase.NewInvoiceUseCase(
		invoiceRepo,
		paymentRepo,
		subscriptionRepo,
		m.eventBus,
	)
	m.paymentUseCase = usecase.NewPaymentUseCase(
		paymentRepo,
		invoiceRepo,
		m.eventBus,
	)

	// Initialize handlers
	m.planHandler = handler.NewPlanHandler(m.planUseCase)
	m.subscriptionHandler = handler.NewSubscriptionHandler(m.subscriptionUseCase)
	m.invoiceHandler = handler.NewInvoiceHandler(m.invoiceUseCase)
	m.paymentHandler = handler.NewPaymentHandler(m.paymentUseCase)

	// Register purge policies for billing data (long retention for legal compliance)
	// Invoices: 7 years retention (legal requirement for financial records)
	invoicePolicy := purge.RetentionPolicy{
		EntityName:    "billing_invoices",
		RetentionDays: invoicesRetentionDays,
		Enabled:       true,
	}
	if err := purge.DefaultPolicyRegistry.RegisterPolicy(invoicePolicy); err != nil {
		slog.Error("Failed to register invoice retention policy", "error", err)
		return err
	}
	slog.Info("Registered purge for billing_invoices", "retention_days", invoicesRetentionDays)

	// Payments: 7 years retention (legal requirement)
	paymentPolicy := purge.RetentionPolicy{
		EntityName:    "billing_payments",
		RetentionDays: paymentsRetentionDays,
		Enabled:       true,
	}
	if err := purge.DefaultPolicyRegistry.RegisterPolicy(paymentPolicy); err != nil {
		slog.Error("Failed to register payment retention policy", "error", err)
		return err
	}
	slog.Info("Registered purge for billing_payments", "retention_days", paymentsRetentionDays)

	// Subscriptions: 3 years retention for canceled/expired subscriptions
	subscriptionPolicy := purge.RetentionPolicy{
		EntityName:    "billing_subscriptions",
		RetentionDays: subscriptionsRetentionDays,
		Enabled:       true,
		// Note: Purge handler should filter by status = 'canceled' OR status = 'expired' in its SQL query
	}
	if err := purge.DefaultPolicyRegistry.RegisterPolicy(subscriptionPolicy); err != nil {
		slog.Error("Failed to register subscription retention policy", "error", err)
		return err
	}
	slog.Info("Registered purge for billing_subscriptions",
		"retention_days", subscriptionsRetentionDays,
		"condition", "canceled or expired")

	slog.Info("Billing module initialized successfully")
	return nil
}

// RegisterRoutes registers HTTP routes for this module
func (m *BillingModule) RegisterRoutes(router *gin.RouterGroup) {
	slog.Info("Registering billing module routes")

	// Health check (must be before parameterized routes)
	router.GET("/billing/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"module":  m.Metadata().Name,
			"version": m.Metadata().Version,
		})
	})

	// Billing routes
	billingGroup := router.Group("/billing")
	{
		// Plan routes
		plansGroup := billingGroup.Group("/plans")
		{
			// Public routes
			plansGroup.GET("", m.planHandler.ListPlans)
			plansGroup.GET("/:id", m.planHandler.GetPlan)

			// Admin routes (protected by middleware in handler)
			plansGroup.POST("", m.planHandler.CreatePlan)
			plansGroup.PUT("/:id", m.planHandler.UpdatePlan)
			plansGroup.POST("/:id/activate", m.planHandler.ActivatePlan)
			plansGroup.POST("/:id/deactivate", m.planHandler.DeactivatePlan)
			plansGroup.DELETE("/:id", m.planHandler.DeletePlan)
		}

		// Subscription routes
		subscriptionsGroup := billingGroup.Group("/subscriptions")
		{
			subscriptionsGroup.POST("", m.subscriptionHandler.CreateSubscription)
			subscriptionsGroup.GET("/:id", m.subscriptionHandler.GetSubscription)
			subscriptionsGroup.GET("/me", m.subscriptionHandler.GetMySubscriptions)
			subscriptionsGroup.GET("/me/active", m.subscriptionHandler.GetMyActiveSubscription)
			subscriptionsGroup.GET("", m.subscriptionHandler.ListSubscriptions) // Admin
			subscriptionsGroup.POST("/:id/cancel", m.subscriptionHandler.CancelSubscription)
			subscriptionsGroup.POST("/:id/upgrade", m.subscriptionHandler.UpgradeSubscription)
		}

		// Invoice routes
		invoicesGroup := billingGroup.Group("/invoices")
		{
			invoicesGroup.POST("", m.invoiceHandler.CreateInvoice) // Admin
			invoicesGroup.GET("/:id", m.invoiceHandler.GetInvoice)
			invoicesGroup.GET("/me", m.invoiceHandler.GetMyInvoices)
			invoicesGroup.GET("", m.invoiceHandler.ListInvoices)                  // Admin
			invoicesGroup.POST("/:id/finalize", m.invoiceHandler.FinalizeInvoice) // Admin
			invoicesGroup.POST("/:id/void", m.invoiceHandler.VoidInvoice)         // Admin
		}

		// Payment routes
		paymentsGroup := billingGroup.Group("/payments")
		{
			paymentsGroup.POST("", m.paymentHandler.CreatePayment)
			paymentsGroup.GET("/:id", m.paymentHandler.GetPayment)
			paymentsGroup.GET("/me", m.paymentHandler.GetMyPayments)
			paymentsGroup.GET("", m.paymentHandler.ListPayments)                  // Admin
			paymentsGroup.POST("/:id/complete", m.paymentHandler.CompletePayment) // Admin/Webhook
			paymentsGroup.POST("/:id/refund", m.paymentHandler.RefundPayment)     // Admin
		}
	}

	slog.Info("Billing module routes registered successfully")
}

// RegisterMigrations returns migrations for this module
func (m *BillingModule) RegisterMigrations() []module.Migration {
	return []module.Migration{
		{
			Version:     1,
			Description: "Create billing_plans table",
			Up:          "-- Migration managed in migrations/billing/000001_billing_plans.up.sql",
			Down:        "-- Migration managed in migrations/billing/000001_billing_plans.down.sql",
		},
		{
			Version:     2,
			Description: "Create billing_subscriptions table",
			Up:          "-- Migration managed in migrations/billing/000002_billing_subscriptions.up.sql",
			Down:        "-- Migration managed in migrations/billing/000002_billing_subscriptions.down.sql",
		},
		{
			Version:     3,
			Description: "Create billing_invoices table",
			Up:          "-- Migration managed in migrations/billing/000003_billing_invoices.up.sql",
			Down:        "-- Migration managed in migrations/billing/000003_billing_invoices.down.sql",
		},
		{
			Version:     4,
			Description: "Create billing_payments table",
			Up:          "-- Migration managed in migrations/billing/000004_billing_payments.up.sql",
			Down:        "-- Migration managed in migrations/billing/000004_billing_payments.down.sql",
		},
	}
}

// RegisterEventHandlers subscribes to events
func (m *BillingModule) RegisterEventHandlers(eventBus bus.IBus) error {
	slog.Info("Registering billing module event handlers")

	// Event handlers will be implemented when background workers are ready
	// Planned events:
	// - billing.subscription.expiring -> Send renewal reminders
	// - billing.payment.failed -> Handle payment failures
	// - billing.invoice.generated -> Send invoice emails
	// - user.deleted -> Cancel subscriptions

	return nil
}

// RegisterPermissions returns permissions required by this module
func (m *BillingModule) RegisterPermissions() []module.Permission {
	return []module.Permission{
		// Plan permissions
		{Resource: "billing:plans", Action: "create", Description: "Create billing plans (admin)"},
		{Resource: "billing:plans", Action: "read", Description: "View billing plans"},
		{Resource: "billing:plans", Action: "update", Description: "Update billing plans (admin)"},
		{Resource: "billing:plans", Action: "delete", Description: "Delete billing plans (admin)"},
		{Resource: "billing:plans", Action: "manage", Description: "Full plan management (admin)"},

		// Subscription permissions
		{Resource: "billing:subscriptions", Action: "create", Description: "Create subscriptions"},
		{Resource: "billing:subscriptions", Action: "read", Description: "View own subscriptions"},
		{Resource: "billing:subscriptions", Action: "update", Description: "Update own subscriptions"},
		{Resource: "billing:subscriptions", Action: "cancel", Description: "Cancel own subscriptions"},
		{Resource: "billing:subscriptions", Action: "manage", Description: "Manage all subscriptions (admin)"},

		// Invoice permissions
		{Resource: "billing:invoices", Action: "create", Description: "Create invoices (admin)"},
		{Resource: "billing:invoices", Action: "read", Description: "View own invoices"},
		{Resource: "billing:invoices", Action: "manage", Description: "Manage all invoices (admin)"},

		// Payment permissions
		{Resource: "billing:payments", Action: "create", Description: "Create payments"},
		{Resource: "billing:payments", Action: "read", Description: "View own payments"},
		{Resource: "billing:payments", Action: "refund", Description: "Refund payments (admin)"},
		{Resource: "billing:payments", Action: "manage", Description: "Manage all payments (admin)"},
	}
}

// Start starts the module background tasks
func (m *BillingModule) Start(ctx context.Context) error {
	slog.Info("Starting billing module")

	// TODO: Start background workers:
	// - Subscription renewal checker
	// - Invoice overdue processor
	// - Payment reminder sender
	// - Trial expiration checker

	return nil
}

// Stop gracefully stops the module
func (m *BillingModule) Stop(ctx context.Context) error {
	slog.Info("Stopping billing module")

	// TODO: Stop background workers gracefully

	return nil
}

// HealthCheck checks module health
func (m *BillingModule) HealthCheck(ctx context.Context) error {
	if m.db == nil {
		return module.ErrModuleNotInitialized
	}

	if err := m.db.PingContext(ctx); err != nil {
		return err
	}

	// Check that billing tables exist and are accessible
	var count int
	err := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM billing_plans WHERE deleted_at IS NULL").Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}
