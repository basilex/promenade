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

// StandardAccounts holds account IDs for automatic journal entries
type StandardAccounts struct {
	BankAccount         uuidv7.UUID // 311 - Bank Accounts
	CashRegister        uuidv7.UUID // 301 - Cash Register
	AccountsReceivable  uuidv7.UUID // 361 - Accounts Receivable
	SalesRevenue        uuidv7.UUID // 702 - Sales Revenue
	OperatingExpenses   uuidv7.UUID // 902 - Operating Expenses
	AccountsPayable     uuidv7.UUID // 631 - Accounts Payable
}

// BankEventHandler handles banking events and creates journal entries
type BankEventHandler struct {
	journalEntryUseCase usecase.IJournalEntryUseCase
	accounts            StandardAccounts
}

// NewBankEventHandler creates a new bank event handler
func NewBankEventHandler(
	journalEntryUseCase usecase.IJournalEntryUseCase,
	accounts StandardAccounts,
) *BankEventHandler {
	return &BankEventHandler{
		journalEntryUseCase: journalEntryUseCase,
		accounts:            accounts,
	}
}

// HandleTransactionRecorded handles bank transaction recorded event
func (h *BankEventHandler) HandleTransactionRecorded(ctx context.Context, event bus.Event) error {
	logger.Info("Processing bank transaction for accounting",
		slog.String("event_type", event.Type()),
		slog.String("aggregate_id", event.AggregateID().String()),
	)

	// Parse event data from metadata payload
	var eventData struct {
		TransactionID    string `json:"transaction_id"`
		AccountID        string `json:"account_id"`
		AmountCents      int64  `json:"amount_cents"`
		CurrencyCode     string `json:"currency_code"`
		Direction        string `json:"direction"` // "credit" or "debit"
		Description      string `json:"description"`
		TransactionDate  string `json:"transaction_date"`
		CounterpartyName string `json:"counterparty_name"`
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

	// Parse transaction ID
	txID, err := uuidv7.Parse(eventData.TransactionID)
	if err != nil {
		logger.Error("Invalid transaction ID", "error", err)
		return err
	}

	// Parse transaction date
	txDate, err := time.Parse(time.RFC3339, eventData.TransactionDate)
	if err != nil {
		txDate = time.Now()
	}

	// Create journal entry description
	description := fmt.Sprintf("Bank transaction: %s", eventData.Description)
	if eventData.CounterpartyName != "" {
		description = fmt.Sprintf("%s (from/to: %s)", description, eventData.CounterpartyName)
	}

	// Determine organization ID (TODO: get from event or context)
	organizationID := event.AggregateID() // Using transaction's aggregate ID for now
	systemUserID := uuidv7.New()          // TODO: Use actual system user ID

	// Create journal entry using usecase
	entry, err := h.journalEntryUseCase.CreateJournalEntry(
		ctx,
		organizationID,
		txDate,
		description,
		aggregate.SourceTypeBankTransaction,
		&txID,
		systemUserID,
	)
	if err != nil {
		logger.Error("Failed to create journal entry", "error", err)
		return nil // Don't fail the operation, just log
	}

	// Add lines based on transaction direction
	var addErr error
	if eventData.Direction == "credit" {
		// Incoming payment: Dr 311 (Bank) / Cr 702 (Sales Revenue)
		if _, addErr = h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.BankAccount, eventData.AmountCents, 0, eventData.CurrencyCode, "Bank receipt", systemUserID); addErr != nil {
			logger.Error("Failed to add debit line", "error", addErr)
			return nil
		}
		if _, addErr = h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.SalesRevenue, 0, eventData.AmountCents, eventData.CurrencyCode, "Sales revenue", systemUserID); addErr != nil {
			logger.Error("Failed to add credit line", "error", addErr)
			return nil
		}
	} else {
		// Outgoing payment: Dr 902 (Operating Expenses) / Cr 311 (Bank)
		if _, addErr = h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.OperatingExpenses, eventData.AmountCents, 0, eventData.CurrencyCode, "Operating expense", systemUserID); addErr != nil {
			logger.Error("Failed to add debit line", "error", addErr)
			return nil
		}
		if _, addErr = h.journalEntryUseCase.AddLine(ctx, entry.ID, h.accounts.BankAccount, 0, eventData.AmountCents, eventData.CurrencyCode, "Bank payment", systemUserID); addErr != nil {
			logger.Error("Failed to add credit line", "error", addErr)
			return nil
		}
	}

	// Auto-post entry (in production, might want manual review)
	postedEntry, err := h.journalEntryUseCase.PostEntry(ctx, entry.ID, systemUserID)
	if err != nil {
		logger.Warn("Failed to post journal entry, left as draft", "error", err)
		return nil
	}

	logger.Info("Successfully created and posted journal entry for bank transaction",
		slog.String("entry_id", postedEntry.ID.String()),
		slog.String("transaction_id", txID.String()),
	)

	return nil
}

// HandleTransactionReconciled handles bank transaction reconciled event
func (h *BankEventHandler) HandleTransactionReconciled(ctx context.Context, event bus.Event) error {
	logger.Info("Bank transaction reconciled, accounting entry already exists",
		slog.String("aggregate_id", event.AggregateID().String()),
	)
	// No additional action needed - entry was already created on transaction recorded
	return nil
}
