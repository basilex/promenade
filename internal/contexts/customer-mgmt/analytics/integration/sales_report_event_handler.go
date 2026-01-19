package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/analytics/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/infrastructure/database"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// SalesReportEventHandler materializes sales report read models from events.
type SalesReportEventHandler struct {
	repo      *postgres.SalesReportRepository
	txManager database.TransactionManager
}

// NewSalesReportEventHandler creates a new sales report event handler.
func NewSalesReportEventHandler(repo *postgres.SalesReportRepository, txManager database.TransactionManager) *SalesReportEventHandler {
	return &SalesReportEventHandler{
		repo:      repo,
		txManager: txManager,
	}
}

// orderConfirmedEvent mirrors the order.confirmed payload (cross-context DTO).
type orderConfirmedEvent struct {
	OrderID     uuidv7.UUID      `json:"order_id"`
	CustomerID  uuidv7.UUID      `json:"customer_id"`
	Items       []orderEventItem `json:"items"`
	Currency    string           `json:"currency"`
	ConfirmedBy uuidv7.UUID      `json:"confirmed_by"`
}

type orderEventItem struct {
	ProductID uuidv7.UUID `json:"product_id"`
	SKU       string      `json:"sku"`
	Quantity  int         `json:"quantity"`
	UnitPrice int64       `json:"unit_price_cents"`
}

// RegisterHandlers registers analytics handlers on the event bus.
func (h *SalesReportEventHandler) RegisterHandlers(eventBus bus.IBus) error {
	if err := eventBus.Subscribe(bus.TopicOrderConfirmed, h.HandleOrderConfirmed); err != nil {
		return fmt.Errorf("failed to subscribe to order.confirmed: %w", err)
	}
	if err := eventBus.Subscribe(bus.TopicOrderCancelled, h.HandleOrderCancelled); err != nil {
		return fmt.Errorf("failed to subscribe to order.cancelled: %w", err)
	}
	if err := eventBus.Subscribe(bus.TopicOrderFulfilled, h.HandleOrderFulfilled); err != nil {
		return fmt.Errorf("failed to subscribe to order.fulfilled: %w", err)
	}
	return nil
}

// HandleOrderConfirmed builds sales report read model entries for confirmed orders.
func (h *SalesReportEventHandler) HandleOrderConfirmed(ctx context.Context, event bus.Event) error {
	log := logger.FromContext(ctx)

	var payload orderConfirmedEvent
	if err := h.parseEventMetadata(event, &payload); err != nil {
		log.Error("Failed to parse order.confirmed event",
			slog.String("event_type", event.Type()),
			slog.Any("error", err),
		)
		return fmt.Errorf("failed to parse order.confirmed event: %w", err)
	}

	if payload.OrderID == uuidv7.Nil || payload.CustomerID == uuidv7.Nil {
		return fmt.Errorf("invalid order.confirmed event: missing order_id or customer_id")
	}
	if payload.Currency == "" {
		return fmt.Errorf("invalid order.confirmed event: missing currency")
	}
	if len(payload.Items) == 0 {
		return fmt.Errorf("invalid order.confirmed event: no items")
	}

	confirmedAt := event.OccurredAt()
	if confirmedAt.IsZero() {
		confirmedAt = time.Now().UTC()
	}

	items := make([]postgres.SalesReportItemRow, 0, len(payload.Items))
	totalCents := int64(0)
	for _, item := range payload.Items {
		lineTotal := item.UnitPrice * int64(item.Quantity)
		totalCents += lineTotal
		items = append(items, postgres.SalesReportItemRow{
			ID:             uuidv7.New(),
			OrderID:        payload.OrderID,
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPrice,
			TotalCents:     lineTotal,
			UpdatedAt:      confirmedAt,
		})
	}

	var managerID *uuidv7.UUID
	if payload.ConfirmedBy != uuidv7.Nil {
		managerID = &payload.ConfirmedBy
	}

	row := &postgres.SalesReportOrderRow{
		OrderID:     payload.OrderID,
		CustomerID:  payload.CustomerID,
		Currency:    payload.Currency,
		TotalCents:  totalCents,
		Status:      "confirmed",
		ManagerID:   managerID,
		ConfirmedAt: confirmedAt,
		UpdatedAt:   confirmedAt,
	}

	if err := h.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.UpsertOrder(txCtx, row); err != nil {
			return fmt.Errorf("upsert sales order: %w", err)
		}
		if err := h.repo.ReplaceItems(txCtx, payload.OrderID, items); err != nil {
			return fmt.Errorf("replace sales order items: %w", err)
		}
		return nil
	}); err != nil {
		log.Error("Failed to materialize sales report order",
			slog.Any("error", err),
			slog.String("order_id", payload.OrderID.String()),
		)
		return err
	}

	log.Info("Sales report order materialized",
		slog.String("order_id", payload.OrderID.String()),
	)
	return nil
}

// HandleOrderCancelled updates read model status for cancelled orders.
func (h *SalesReportEventHandler) HandleOrderCancelled(ctx context.Context, event bus.Event) error {
	return h.updateStatus(ctx, event, "cancelled")
}

// HandleOrderFulfilled updates read model status for fulfilled orders.
func (h *SalesReportEventHandler) HandleOrderFulfilled(ctx context.Context, event bus.Event) error {
	return h.updateStatus(ctx, event, "fulfilled")
}

func (h *SalesReportEventHandler) updateStatus(ctx context.Context, event bus.Event, status string) error {
	log := logger.FromContext(ctx)
	orderID := event.AggregateID()
	if orderID == uuidv7.Nil {
		return fmt.Errorf("invalid event: missing aggregate_id")
	}

	updatedAt := event.OccurredAt()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	if err := h.repo.UpdateOrderStatus(ctx, orderID, status, updatedAt); err != nil {
		log.Error("Failed to update sales report order status",
			slog.Any("error", err),
			slog.String("order_id", orderID.String()),
		)
		return err
	}

	log.Info("Sales report order status updated",
		slog.String("order_id", orderID.String()),
		slog.String("status", status),
	)
	return nil
}

func (h *SalesReportEventHandler) parseEventMetadata(event bus.Event, target interface{}) error {
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
