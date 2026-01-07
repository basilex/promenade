package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// OrderEventHandler handles order lifecycle events for warehouse operations
type OrderEventHandler struct {
	reservationService IReservationService
}

// NewOrderEventHandler creates a new order event handler
func NewOrderEventHandler(reservationService IReservationService) *OrderEventHandler {
	return &OrderEventHandler{
		reservationService: reservationService,
	}
}

// OrderConfirmedEvent represents the order.confirmed domain event
type OrderConfirmedEvent struct {
	OrderID    uuidv7.UUID   `json:"order_id"`
	CustomerID uuidv7.UUID   `json:"customer_id"`
	Items      []OrderItem   `json:"items"`
	ConfirmedBy uuidv7.UUID  `json:"confirmed_by"`
}

// OrderCancelledEvent represents the order.cancelled domain event
type OrderCancelledEvent struct {
	OrderID     uuidv7.UUID   `json:"order_id"`
	Items       []OrderItem   `json:"items"`
	Reason      string        `json:"reason"`
	CancelledBy uuidv7.UUID   `json:"cancelled_by"`
}

// OrderFulfilledEvent represents the order.fulfilled domain event
type OrderFulfilledEvent struct {
	OrderID     uuidv7.UUID   `json:"order_id"`
	Items       []OrderItem   `json:"items"`
	FulfilledBy uuidv7.UUID   `json:"fulfilled_by"`
}

// HandleOrderConfirmed reserves stock when an order is confirmed
func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event bus.Event) error {
	log := logger.FromContext(ctx)

	// Parse event metadata
	var orderEvent OrderConfirmedEvent
	if err := h.parseEventMetadata(event, &orderEvent); err != nil {
		log.Error("Failed to parse order.confirmed event",
			slog.String("event_type", event.Type()),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to parse order.confirmed event: %w", err)
	}

	// Validate event data
	if orderEvent.OrderID == uuidv7.Nil {
		return fmt.Errorf("invalid order.confirmed event: missing order_id")
	}
	if orderEvent.ConfirmedBy == uuidv7.Nil {
		return fmt.Errorf("invalid order.confirmed event: missing confirmed_by")
	}
	if len(orderEvent.Items) == 0 {
		return fmt.Errorf("invalid order.confirmed event: no items")
	}

	log.Info("Reserving stock for confirmed order",
		slog.String("order_id", orderEvent.OrderID.String()),
		slog.Int("item_count", len(orderEvent.Items)),
	)

	// Reserve stock for order
	if err := h.reservationService.ReserveForOrder(ctx, orderEvent.OrderID, orderEvent.Items, orderEvent.ConfirmedBy); err != nil {
		log.Error("Failed to reserve stock for order",
			slog.String("order_id", orderEvent.OrderID.String()),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to reserve stock for order %s: %w", orderEvent.OrderID, err)
	}

	log.Info("Stock reserved successfully",
		slog.String("order_id", orderEvent.OrderID.String()),
	)

	return nil
}

// HandleOrderCancelled releases stock when an order is cancelled
func (h *OrderEventHandler) HandleOrderCancelled(ctx context.Context, event bus.Event) error {
	log := logger.FromContext(ctx)

	// Parse event metadata
	var orderEvent OrderCancelledEvent
	if err := h.parseEventMetadata(event, &orderEvent); err != nil {
		log.Error("Failed to parse order.cancelled event",
			slog.String("event_type", event.Type()),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to parse order.cancelled event: %w", err)
	}

	// Validate event data
	if orderEvent.OrderID == uuidv7.Nil {
		return fmt.Errorf("invalid order.cancelled event: missing order_id")
	}
	if orderEvent.CancelledBy == uuidv7.Nil {
		return fmt.Errorf("invalid order.cancelled event: missing cancelled_by")
	}
	if len(orderEvent.Items) == 0 {
		return fmt.Errorf("invalid order.cancelled event: no items")
	}

	log.Info("Releasing stock for cancelled order",
		slog.String("order_id", orderEvent.OrderID.String()),
		slog.String("reason", orderEvent.Reason),
	)

	// Release stock for order (idempotent, best effort)
	if err := h.reservationService.ReleaseForOrder(ctx, orderEvent.OrderID, orderEvent.Items, orderEvent.CancelledBy); err != nil {
		log.Warn("Failed to release stock for cancelled order",
			slog.String("order_id", orderEvent.OrderID.String()),
			slog.Any("error", err),
		)
		// Don't return error - release is best effort
	}

	log.Info("Stock release processed",
		slog.String("order_id", orderEvent.OrderID.String()),
	)

	return nil
}

// HandleOrderFulfilled commits stock when an order is fulfilled
func (h *OrderEventHandler) HandleOrderFulfilled(ctx context.Context, event bus.Event) error {
	log := logger.FromContext(ctx)

	// Parse event metadata
	var orderEvent OrderFulfilledEvent
	if err := h.parseEventMetadata(event, &orderEvent); err != nil {
		log.Error("Failed to parse order.fulfilled event",
			slog.String("event_type", event.Type()),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to parse order.fulfilled event: %w", err)
	}

	// Validate event data
	if orderEvent.OrderID == uuidv7.Nil {
		return fmt.Errorf("invalid order.fulfilled event: missing order_id")
	}
	if orderEvent.FulfilledBy == uuidv7.Nil {
		return fmt.Errorf("invalid order.fulfilled event: missing fulfilled_by")
	}
	if len(orderEvent.Items) == 0 {
		return fmt.Errorf("invalid order.fulfilled event: no items")
	}

	log.Info("Committing stock for fulfilled order",
		slog.String("order_id", orderEvent.OrderID.String()),
		slog.Int("item_count", len(orderEvent.Items)),
	)

	// Commit stock for order
	if err := h.reservationService.CommitForOrder(ctx, orderEvent.OrderID, orderEvent.Items, orderEvent.FulfilledBy); err != nil {
		log.Error("Failed to commit stock for order",
			slog.String("order_id", orderEvent.OrderID.String()),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to commit stock for order %s: %w", orderEvent.OrderID, err)
	}

	log.Info("Stock committed successfully",
		slog.String("order_id", orderEvent.OrderID.String()),
	)

	return nil
}

// parseEventMetadata extracts and unmarshals event metadata into target struct
func (h *OrderEventHandler) parseEventMetadata(event bus.Event, target interface{}) error {
	metadata := event.Metadata()
	if metadata == nil {
		return fmt.Errorf("event metadata is nil")
	}

	// Get JSON string from metadata
	jsonStr, ok := metadata["payload"]
	if !ok {
		return fmt.Errorf("event metadata missing 'payload' field")
	}

	// Unmarshal JSON string into target struct
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

// RegisterHandlers registers all order event handlers with the event bus
func (h *OrderEventHandler) RegisterHandlers(eventBus bus.IBus) error {
	// Register order.confirmed handler
	if err := eventBus.Subscribe("order.confirmed", h.HandleOrderConfirmed); err != nil {
		return fmt.Errorf("failed to subscribe to order.confirmed: %w", err)
	}

	// Register order.cancelled handler
	if err := eventBus.Subscribe("order.cancelled", h.HandleOrderCancelled); err != nil {
		return fmt.Errorf("failed to subscribe to order.cancelled: %w", err)
	}

	// Register order.fulfilled handler
	if err := eventBus.Subscribe("order.fulfilled", h.HandleOrderFulfilled); err != nil {
		return fmt.Errorf("failed to subscribe to order.fulfilled: %w", err)
	}

	return nil
}
