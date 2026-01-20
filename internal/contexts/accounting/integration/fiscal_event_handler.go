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

// FiscalEventHandler handles fiscal events and creates journal entries
type FiscalEventHandler struct {
	journalEntryUseCase usecase.IJournalEntryUseCase
	accounts            StandardAccounts
}

// NewFiscalEventHandler creates a new fiscal event handler
func NewFiscalEventHandler(
	journalEntryUseCase usecase.IJournalEntryUseCase,
	accounts StandardAccounts,
) *FiscalEventHandler {
	return &FiscalEventHandler{
		journalEntryUseCase: journalEntryUseCase,
		accounts:            accounts,
	}
}

// HandleReceiptCreated handles fiscal receipt created event
func (h *FiscalEventHandler) HandleReceiptCreated(ctx context.Context, event bus.Event) error {
	logger.Info("Processing fiscal receipt for accounting",
		slog.String("event_type", event.Type()),
		slog.String("aggregate_id", event.AggregateID().String()),
	)

	// Parse event data from metadata payload
	var eventData struct {
		ReceiptID    string `json:"receipt_id"`
		OrderID      string `json:"order_id"`
		AmountCents  int64  `json:"amount_cents"`
		CurrencyCode string `json:"currency_code"`
		Description  string `json:"description"`
		CreatedAt    string `json:"created_at"`
		CashierID    string `json:"cashier_id"`
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

	// Parse receipt ID
	receiptID, err := uuidv7.Parse(eventData.ReceiptID)
	if err != nil {
		logger.Error("Invalid receipt ID", "error", err)
		return err
	}

	// Parse created date
	createdAt, err := time.Parse(time.RFC3339, eventData.CreatedAt)
	if err != nil {
		createdAt = time.Now()
	}

	// Create journal entry: Dr 301 (Cash Register) / Cr 702 (Sales Revenue)
	description := fmt.Sprintf("Fiscal receipt: %s", eventData.Description)

	// Determine organization ID (TODO: get from event or context)
	organizationID := event.AggregateID() // Using receipt's aggregate ID for now
	systemUserID := uuidv7.New()          // TODO: Use actual cashier ID from event

	if eventData.CashierID != "" {
		if cashierID, err := uuidv7.Parse(eventData.CashierID); err == nil {
			systemUserID = cashierID
		}
	}

	// Create journal entry using usecase
	entry, err := h.journalEntryUseCase.CreateJournalEntry(
		ctx,
		organizationID,
		createdAt,
		description,
		aggregate.SourceTypeReceipt,
		&receiptID,
		systemUserID,
	)
	if err != nil {
		logger.Error("Failed to create journal entry", "error", err)
		return err
	}

	// Add debit line: Dr 301 (Cash Register)
	if _, err := h.journalEntryUseCase.AddLine(
		ctx,
		entry.ID,
		h.accounts.CashRegister,
		eventData.AmountCents,
		0,
		eventData.CurrencyCode,
		"Dr 301 Cash in national currency",
		systemUserID,
	); err != nil {
		logger.Error("Failed to add debit line to journal entry", "error", err)
		return err
	}

	// Add credit line: Cr 702 (Sales Revenue)
	if _, err := h.journalEntryUseCase.AddLine(
		ctx,
		entry.ID,
		h.accounts.SalesRevenue,
		0,
		eventData.AmountCents,
		eventData.CurrencyCode,
		"Cr 702 Sales Revenue",
		systemUserID,
	); err != nil {
		logger.Error("Failed to add credit line to journal entry", "error", err)
		return err
	}

	// Auto-post the entry
	if _, err := h.journalEntryUseCase.PostEntry(ctx, entry.ID, systemUserID); err != nil {
		logger.Error("Failed to post journal entry", "error", err)
		// Log but don't fail - entry is created in draft state
		logger.Warn("Journal entry created but not posted",
			slog.String("entry_id", entry.ID.String()),
		)
	}

	logger.Info("Fiscal receipt processed successfully",
		slog.String("receipt_id", receiptID.String()),
		slog.String("entry_id", entry.ID.String()),
		slog.Int64("amount_cents", eventData.AmountCents),
	)

	return nil
}
