package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// BillingEventHandler handles billing events and creates journal entries
type BillingEventHandler struct {
	journalEntryUseCase usecase.IJournalEntryUseCase
	accounts            StandardAccounts
}

// NewBillingEventHandler creates a new billing event handler
func NewBillingEventHandler(
	journalEntryUseCase usecase.IJournalEntryUseCase,
	accounts StandardAccounts,
) *BillingEventHandler {
	return &BillingEventHandler{
		journalEntryUseCase: journalEntryUseCase,
		accounts:            accounts,
	}
}

// HandleInvoiceGenerated handles invoice generated event
func (h *BillingEventHandler) HandleInvoiceGenerated(ctx context.Context, event bus.Event) error {
	logger.Info("Processing invoice for accounting",
		slog.String("event_type", event.Type()),
		slog.String("aggregate_id", event.AggregateID().String()),
	)

	// Parse event data from metadata payload
	var eventData struct {
		InvoiceID    string `json:"invoice_id"`
		CustomerID   string `json:"customer_id"`
		AmountCents  int64  `json:"amount_cents"`
		CurrencyCode string `json:"currency_code"`
		Description  string `json:"description"`
		IssueDate    string `json:"issue_date"`
	}

	// Get payload from event metadata
	metadata := event.Metadata()
	if metadata == nil {
		logger.Error("Event metadata is nil")
		return fmt.Errorf("event metadata is nil")
	}

	payloadStr, ok := metadata["payload"]
	if !ok {
		logger.Error("Event metadata missing 'payload' field")
		return fmt.Errorf("event metadata missing 'payload' field")
	}

	if err := json.Unmarshal([]byte(payloadStr), &eventData); err != nil {
		logger.Error("Failed to unmarshal event payload", "error", err)
		return err
	}

	// Parse invoice ID
	invoiceID, err := uuidv7.Parse(eventData.InvoiceID)
	if err != nil {
		logger.Error("Invalid invoice ID", "error", err)
		return err
	}

	// Parse issue date
	issueDate, err := time.Parse(time.RFC3339, eventData.IssueDate)
	if err != nil {
		issueDate = time.Now()
	}

	// Create journal entry: Dr 361 (Accounts Receivable) / Cr 702 (Sales Revenue)
	description := fmt.Sprintf("Invoice generated: %s", eventData.Description)

	// Determine organization ID (TODO: get from event or context)
	organizationID := event.AggregateID() // Using invoice's aggregate ID for now
	systemUserID := uuidv7.New()          // TODO: Use actual system user ID

	// Create journal entry using usecase
	entry, err := h.journalEntryUseCase.CreateJournalEntry(
		ctx,
		organizationID,
		issueDate,
		description,
		aggregate.SourceTypeInvoice,
		&invoiceID,
		systemUserID,
	)
	if err != nil {
		logger.Error("Failed to create journal entry", "error", err)
		return nil
	}

	// Debit: Accounts Receivable (customer owes us)
	if _, err := h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.AccountsReceivable, eventData.AmountCents, 0, eventData.CurrencyCode, "Invoice issued", systemUserID); err != nil {
		logger.Error("Failed to add debit line", "error", err)
		return nil
	}

	// Credit: Sales Revenue (we earned revenue)
	if _, err := h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.SalesRevenue, 0, eventData.AmountCents, eventData.CurrencyCode, "Sales revenue recognized", systemUserID); err != nil {
		logger.Error("Failed to add credit line", "error", err)
		return nil
	}

	// Auto-post entry
	postedEntry, err := h.journalEntryUseCase.PostEntry(ctx, entry.ID, systemUserID)
	if err != nil {
		logger.Warn("Failed to post journal entry, left as draft", "error", err)
		return nil
	}

	logger.Info("Successfully created journal entry for invoice",
		slog.String("entry_id", postedEntry.ID.String()),
		slog.String("invoice_id", invoiceID.String()),
	)

	return nil
}

// HandlePaymentReceived handles payment received event
func (h *BillingEventHandler) HandlePaymentReceived(ctx context.Context, event bus.Event) error {
	logger.Info("Processing payment for accounting",
		slog.String("event_type", event.Type()),
		slog.String("aggregate_id", event.AggregateID().String()),
	)

	// Parse event data from metadata payload
	var eventData struct {
		PaymentID    string `json:"payment_id"`
		InvoiceID    string `json:"invoice_id"`
		AmountCents  int64  `json:"amount_cents"`
		CurrencyCode string `json:"currency_code"`
		PaymentDate  string `json:"payment_date"`
		Method       string `json:"method"`
	}

	// Get payload from event metadata
	metadata := event.Metadata()
	if metadata == nil {
		logger.Error("Event metadata is nil")
		return fmt.Errorf("event metadata is nil")
	}

	payloadStr, ok := metadata["payload"]
	if !ok {
		logger.Error("Event metadata missing 'payload' field")
		return fmt.Errorf("event metadata missing 'payload' field")
	}

	if err := json.Unmarshal([]byte(payloadStr), &eventData); err != nil {
		logger.Error("Failed to unmarshal event payload", "error", err)
		return err
	}

	// Parse payment ID
	paymentID, err := uuidv7.Parse(eventData.PaymentID)
	if err != nil {
		logger.Error("Invalid payment ID", "error", err)
		return err
	}

	// Parse payment date
	paymentDate, err := time.Parse(time.RFC3339, eventData.PaymentDate)
	if err != nil {
		paymentDate = time.Now()
	}

	// Create journal entry: Dr 311 (Bank) / Cr 361 (Accounts Receivable)
	description := fmt.Sprintf("Payment received: %s", eventData.Method)

	// Determine organization ID (TODO: get from event or context)
	organizationID := event.AggregateID() // Using payment's aggregate ID for now
	systemUserID := uuidv7.New()          // TODO: Use actual system user ID

	// Create journal entry using usecase
	entry, err := h.journalEntryUseCase.CreateJournalEntry(
		ctx,
		organizationID,
		paymentDate,
		description,
		aggregate.SourceTypePayment,
		&paymentID,
		systemUserID,
	)
	if err != nil {
		logger.Error("Failed to create journal entry", "error", err)
		return nil
	}

	// Debit: Bank Account (we received money)
	if _, err := h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.BankAccount, eventData.AmountCents, 0, eventData.CurrencyCode, "Payment received", systemUserID); err != nil {
		logger.Error("Failed to add debit line", "error", err)
		return nil
	}

	// Credit: Accounts Receivable (customer paid their debt)
	if _, err := h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.AccountsReceivable, 0, eventData.AmountCents, eventData.CurrencyCode, "Payment applied to receivable", systemUserID); err != nil {
		logger.Error("Failed to add credit line", "error", err)
		return nil
	}

	// Auto-post entry
	postedEntry, err := h.journalEntryUseCase.PostEntry(ctx, entry.ID, systemUserID)
	if err != nil {
		logger.Warn("Failed to post journal entry, left as draft", "error", err)
		return nil
	}

	logger.Info("Successfully created journal entry for payment",
		slog.String("entry_id", postedEntry.ID.String()),
		slog.String("payment_id", paymentID.String()),
	)

	return nil
}
