package integration

import (
	"context"

	accountUseCase "github.com/basilex/promenade/internal/contexts/accounting/account/usecase"
	fiscalPeriodUseCase "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/usecase"
	journalEntryUseCase "github.com/basilex/promenade/internal/contexts/accounting/journalentry/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// AccountingEventHandler orchestrates all accounting event handlers
type AccountingEventHandler struct {
	bankHandler    *BankEventHandler
	billingHandler *BillingEventHandler
	fiscalHandler  *FiscalEventHandler
}

// NewAccountingEventHandler creates a new accounting event handler
func NewAccountingEventHandler(
	journalEntryUC journalEntryUseCase.IJournalEntryUseCase,
	accountUC accountUseCase.IAccountUseCase,
	fiscalPeriodUC fiscalPeriodUseCase.IFiscalPeriodUseCase,
) *AccountingEventHandler {
	// TODO: Load standard account IDs from configuration or database
	// For now, using placeholder IDs - these should be configured during setup
	accounts := StandardAccounts{
		BankAccount:        uuidv7.New(), // 311 - Bank Accounts
		CashRegister:       uuidv7.New(), // 301 - Cash Register
		AccountsReceivable: uuidv7.New(), // 361 - Accounts Receivable
		SalesRevenue:       uuidv7.New(), // 702 - Sales Revenue
		OperatingExpenses:  uuidv7.New(), // 902 - Operating Expenses
		AccountsPayable:    uuidv7.New(), // 631 - Accounts Payable
	}

	return &AccountingEventHandler{
		bankHandler:    NewBankEventHandler(journalEntryUC, accounts),
		billingHandler: NewBillingEventHandler(journalEntryUC, accounts),
		fiscalHandler:  NewFiscalEventHandler(journalEntryUC, accounts),
	}
}

// RegisterHandlers registers all event handlers with the event bus
func (h *AccountingEventHandler) RegisterHandlers(eventBus bus.IBus) error {
	// Bank transaction events
	if err := eventBus.Subscribe("bank.transaction.recorded", h.handleBankTransactionRecorded); err != nil {
		return err
	}

	// Billing events
	if err := eventBus.Subscribe("billing.invoice.generated", h.handleInvoiceGenerated); err != nil {
		return err
	}

	if err := eventBus.Subscribe("billing.payment.received", h.handlePaymentReceived); err != nil {
		return err
	}

	// Fiscal events
	if err := eventBus.Subscribe("fiscal.receipt.created", h.handleReceiptCreated); err != nil {
		return err
	}

	logger.Info("Accounting event handlers registered",
		"handlers", []string{
			"bank.transaction.recorded",
			"billing.invoice.generated",
			"billing.payment.received",
			"fiscal.receipt.created",
		},
	)

	return nil
}

// Bank transaction event handlers
func (h *AccountingEventHandler) handleBankTransactionRecorded(ctx context.Context, event bus.Event) error {
	return h.bankHandler.HandleTransactionRecorded(ctx, event)
}

// Billing event handlers
func (h *AccountingEventHandler) handleInvoiceGenerated(ctx context.Context, event bus.Event) error {
	return h.billingHandler.HandleInvoiceGenerated(ctx, event)
}

func (h *AccountingEventHandler) handlePaymentReceived(ctx context.Context, event bus.Event) error {
	return h.billingHandler.HandlePaymentReceived(ctx, event)
}

// Fiscal event handlers
func (h *AccountingEventHandler) handleReceiptCreated(ctx context.Context, event bus.Event) error {
	return h.fiscalHandler.HandleReceiptCreated(ctx, event)
}
