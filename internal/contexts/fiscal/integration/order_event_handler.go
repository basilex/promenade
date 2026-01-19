package integration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/repository"
	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	receiptusecase "github.com/basilex/promenade/internal/contexts/fiscal/receipt/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// OrderEventHandler handles order lifecycle events for fiscal receipts.
type OrderEventHandler struct {
	receiptUC        receiptusecase.IReceiptUseCase
	cashRegisterRepo repository.ICashRegisterRepository
	printerEnabled   bool
}

// NewOrderEventHandler creates a new fiscal order event handler.
func NewOrderEventHandler(receiptUC receiptusecase.IReceiptUseCase, cashRegisterRepo repository.ICashRegisterRepository, printerEnabled bool) *OrderEventHandler {
	return &OrderEventHandler{
		receiptUC:        receiptUC,
		cashRegisterRepo: cashRegisterRepo,
		printerEnabled:   printerEnabled,
	}
}

// OrderConfirmedEvent represents the order.confirmed domain event payload.
type OrderConfirmedEvent struct {
	OrderID     uuidv7.UUID `json:"order_id"`
	CustomerID  uuidv7.UUID `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	Currency    string      `json:"currency"`
	ConfirmedBy uuidv7.UUID `json:"confirmed_by"`
}

// OrderItem represents an order line item for fiscal receipts.
type OrderItem struct {
	ProductID uuidv7.UUID `json:"product_id"`
	SKU       string      `json:"sku"`
	Quantity  int         `json:"quantity"`
	UnitPrice int64       `json:"unit_price_cents"`
}

// HandleOrderConfirmed creates and optionally prints a fiscal receipt when an order is confirmed.
func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event bus.Event) error {
	log := logger.FromContext(ctx)

	var orderEvent OrderConfirmedEvent
	if err := h.parseEventMetadata(event, &orderEvent); err != nil {
		log.Error("Failed to parse order.confirmed event", slog.Any("error", err))
		return fmt.Errorf("failed to parse order.confirmed event: %w", err)
	}

	if orderEvent.OrderID == uuidv7.Nil {
		return fmt.Errorf("invalid order.confirmed event: missing order_id")
	}
	if orderEvent.ConfirmedBy == uuidv7.Nil {
		return fmt.Errorf("invalid order.confirmed event: missing confirmed_by")
	}
	if orderEvent.Currency == "" {
		return fmt.Errorf("invalid order.confirmed event: missing currency")
	}
	if len(orderEvent.Items) == 0 {
		return fmt.Errorf("invalid order.confirmed event: no items")
	}

	cashRegisters, err := h.cashRegisterRepo.ListActive(ctx)
	if err != nil {
		log.Error("Failed to list active cash registers", slog.Any("error", err))
		return fmt.Errorf("failed to list active cash registers: %w", err)
	}
	if len(cashRegisters) == 0 {
		return fmt.Errorf("no active cash registers available")
	}

	cashRegister := cashRegisters[0]
	lines := make([]aggregate.ReceiptLine, 0, len(orderEvent.Items))
	for _, item := range orderEvent.Items {
		name := item.SKU
		if name == "" {
			name = fmt.Sprintf("Product %s", item.ProductID.String())
		}
		lines = append(lines, aggregate.ReceiptLine{
			Name:       name,
			Quantity:   item.Quantity,
			PriceCents: item.UnitPrice,
			TaxRate:    0,
		})
	}

	rec, err := h.receiptUC.CreateReceipt(
		ctx,
		cashRegister.GetID(),
		orderEvent.OrderID,
		aggregate.PaymentTypeCard,
		aggregate.ReceiptTypeSale,
		orderEvent.Currency,
		lines,
		orderEvent.ConfirmedBy,
	)
	if err != nil {
		if errors.Is(err, receipterrors.ErrReceiptAlreadyExists) {
			log.Info("Receipt already exists for order", slog.String("order_id", orderEvent.OrderID.String()))
			return nil
		}
		log.Error("Failed to create receipt from order.confirmed", slog.Any("error", err))
		return err
	}

	if !h.printerEnabled {
		log.Info("Receipt created without auto-print (printer not configured)", slog.String("receipt_id", rec.GetID().String()))
		return nil
	}

	if _, err := h.receiptUC.PrintReceipt(ctx, rec.GetID(), orderEvent.ConfirmedBy); err != nil {
		log.Error("Failed to auto-print receipt", slog.Any("error", err))
		return err
	}

	log.Info("Receipt created and printed", slog.String("receipt_id", rec.GetID().String()), slog.String("order_id", orderEvent.OrderID.String()))
	return nil
}

// RegisterHandlers registers fiscal handlers on the event bus.
func (h *OrderEventHandler) RegisterHandlers(eventBus bus.IBus) error {
	if err := eventBus.Subscribe(bus.TopicOrderConfirmed, h.HandleOrderConfirmed); err != nil {
		return fmt.Errorf("failed to subscribe to order.confirmed: %w", err)
	}
	return nil
}

func (h *OrderEventHandler) parseEventMetadata(event bus.Event, target interface{}) error {
	metadata := event.Metadata()
	if metadata == nil {
		return fmt.Errorf("event metadata is nil")
	}

	jsonStr, ok := metadata["payload"]
	if !ok {
		return fmt.Errorf("event metadata missing 'payload' field")
	}

	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}
