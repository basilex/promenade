package usecase

import (
	"github.com/basilex/promenade/internal/contexts/order-mgmt/order/aggregate"
	ordererrors "github.com/basilex/promenade/internal/contexts/order-mgmt/order"
	"github.com/basilex/promenade/internal/contexts/order-mgmt/order/repository"
)

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IOrderUseCase defines the interface for order business logic
type IOrderUseCase interface {
	// CreateOrder creates a new order
	CreateOrder(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, currency string) (*aggregate.Order, error)

	// GetOrder retrieves an order by ID
	GetOrder(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error)

	// GetOrderByNumber retrieves an order by order number
	GetOrderByNumber(ctx context.Context, orderNumber string) (*aggregate.Order, error)

	// AddOrderLine adds a line item to an order
	AddOrderLine(ctx context.Context, orderID, productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) (*aggregate.Order, error)

	// RemoveOrderLine removes a line item from an order
	RemoveOrderLine(ctx context.Context, orderID, lineID uuidv7.UUID) (*aggregate.Order, error)

	// UpdateOrderLineQuantity updates the quantity of a line item
	UpdateOrderLineQuantity(ctx context.Context, orderID, lineID uuidv7.UUID, quantity int) (*aggregate.Order, error)

	// ConfirmOrder confirms an order
	ConfirmOrder(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error)

	// StartProcessing moves order to processing status
	StartProcessing(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error)

	// MarkFulfilled marks order as fulfilled
	MarkFulfilled(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error)

	// CancelOrder cancels an order
	CancelOrder(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error)

	// ListOrders retrieves all orders with pagination
	ListOrders(ctx context.Context, page, pageSize int) ([]*aggregate.Order, int64, error)

	// ListOrdersByCustomer retrieves all orders for a customer
	ListOrdersByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Order, int64, error)

	// ListOrdersByStatus retrieves all orders with specific status
	ListOrdersByStatus(ctx context.Context, status aggregate.OrderStatus, page, pageSize int) ([]*aggregate.Order, int64, error)
}

// orderUseCase implements IOrderUseCase
type orderUseCase struct {
	repo     repository.IOrderRepository
	eventBus bus.IBus
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(repo repository.IOrderRepository, eventBus bus.IBus) IOrderUseCase {
	return &orderUseCase{
		repo:     repo,
		eventBus: eventBus,
	}
}

// CreateOrder creates a new order
func (uc *orderUseCase) CreateOrder(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, currency string) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Create new order
	order, err := aggregate.NewOrder(customerID, currency)
	if err != nil {
		log.Error("Failed to create order entity", slog.Any("error", err))
		return nil, err
	}

	// Set company ID if provided
	if companyID != nil {
		order.CompanyID = companyID
	}

	// Persist order
	if err := uc.repo.Create(ctx, order); err != nil {
		log.Error("Failed to persist order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderPersistFailed
	}

	log.Info("aggregate.Order created successfully",
		slog.String("order_id", order.ID.String()),
		slog.String("order_number", order.OrderNumber),
		slog.String("customer_id", customerID.String()),
	)

	return order, nil
}

// GetOrder retrieves an order by ID
func (uc *orderUseCase) GetOrder(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		log.Error("Failed to get order", slog.Any("error", err), slog.String("order_id", orderID.String()))
		return nil, ordererrors.ErrOrderGetFailed
	}

	// Load order lines
	lines, err := uc.repo.GetLines(ctx, orderID)
	if err != nil {
		log.Error("Failed to get order lines", slog.Any("error", err))
		return nil, ordererrors.ErrOrderLinesFetchFailed
	}
	order.Lines = lines

	return order, nil
}

// GetOrderByNumber retrieves an order by order number
func (uc *orderUseCase) GetOrderByNumber(ctx context.Context, orderNumber string) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	order, err := uc.repo.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		log.Error("Failed to get order by number", slog.Any("error", err), slog.String("order_number", orderNumber))
		return nil, ordererrors.ErrOrderGetFailed
	}

	// Load order lines
	lines, err := uc.repo.GetLines(ctx, order.ID)
	if err != nil {
		log.Error("Failed to get order lines", slog.Any("error", err))
		return nil, ordererrors.ErrOrderLinesFetchFailed
	}
	order.Lines = lines

	return order, nil
}

// AddOrderLine adds a line item to an order
func (uc *orderUseCase) AddOrderLine(ctx context.Context, orderID, productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Add line to order (business logic)
	if err := order.AddLine(productID, quantity, unitPrice); err != nil {
		log.Error("Failed to add line to order", slog.Any("error", err))
		return nil, err
	}

	// Persist new line
	newLine := order.Lines[len(order.Lines)-1]
	if err := uc.repo.CreateLine(ctx, &newLine); err != nil {
		log.Error("Failed to persist order line", slog.Any("error", err))
		return nil, ordererrors.ErrOrderLineCreateFailed
	}

	// Update order total
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderUpdateFailed
	}

	log.Info("aggregate.Order line added successfully",
		slog.String("order_id", orderID.String()),
		slog.String("line_id", newLine.ID.String()),
	)

	return order, nil
}

// RemoveOrderLine removes a line item from an order
func (uc *orderUseCase) RemoveOrderLine(ctx context.Context, orderID, lineID uuidv7.UUID) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Remove line from order (business logic)
	if err := order.RemoveLine(lineID); err != nil {
		log.Error("Failed to remove line from order", slog.Any("error", err))
		return nil, err
	}

	// Delete line from database
	if err := uc.repo.DeleteLine(ctx, lineID); err != nil {
		log.Error("Failed to delete order line", slog.Any("error", err))
		return nil, ordererrors.ErrOrderLineDeleteFailed
	}

	// Update order total
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderUpdateFailed
	}

	log.Info("aggregate.Order line removed successfully",
		slog.String("order_id", orderID.String()),
		slog.String("line_id", lineID.String()),
	)

	return order, nil
}

// UpdateOrderLineQuantity updates the quantity of a line item
func (uc *orderUseCase) UpdateOrderLineQuantity(ctx context.Context, orderID, lineID uuidv7.UUID, quantity int) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Update line quantity (business logic)
	if err := order.UpdateLineQuantity(lineID, quantity); err != nil {
		log.Error("Failed to update line quantity", slog.Any("error", err))
		return nil, err
	}

	// Find and update the line
	for _, line := range order.Lines {
		if line.ID == lineID {
			if err := uc.repo.UpdateLine(ctx, &line); err != nil {
				log.Error("Failed to update order line", slog.Any("error", err))
				return nil, ordererrors.ErrOrderLineUpdateFailed
			}
			break
		}
	}

	// Update order total
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderUpdateFailed
	}

	log.Info("aggregate.Order line quantity updated successfully",
		slog.String("order_id", orderID.String()),
		slog.String("line_id", lineID.String()),
		slog.Int("quantity", quantity),
	)

	return order, nil
}

// ConfirmOrder confirms an order
func (uc *orderUseCase) ConfirmOrder(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Confirm order (business logic)
	if err := order.Confirm(); err != nil {
		log.Error("Failed to confirm order", slog.Any("error", err))
		return nil, err
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderUpdateFailed
	}

	log.Info("aggregate.Order confirmed successfully", slog.String("order_id", orderID.String()))
	uc.publishOrderConfirmed(ctx, order)

	return order, nil
}

// StartProcessing moves order to processing status
func (uc *orderUseCase) StartProcessing(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Start processing (business logic)
	if err := order.StartProcessing(); err != nil {
		log.Error("Failed to start processing", slog.Any("error", err))
		return nil, err
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderUpdateFailed
	}

	log.Info("aggregate.Order processing started", slog.String("order_id", orderID.String()))

	return order, nil
}

// MarkFulfilled marks order as fulfilled
func (uc *orderUseCase) MarkFulfilled(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Mark fulfilled (business logic)
	if err := order.MarkFulfilled(); err != nil {
		log.Error("Failed to mark fulfilled", slog.Any("error", err))
		return nil, err
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderUpdateFailed
	}

	log.Info("aggregate.Order marked as fulfilled", slog.String("order_id", orderID.String()))
	uc.publishOrderFulfilled(ctx, order)

	return order, nil
}

// CancelOrder cancels an order
func (uc *orderUseCase) CancelOrder(ctx context.Context, orderID uuidv7.UUID) (*aggregate.Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Cancel order (business logic)
	if err := order.Cancel(); err != nil {
		log.Error("Failed to cancel order", slog.Any("error", err))
		return nil, err
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, ordererrors.ErrOrderUpdateFailed
	}

	log.Info("aggregate.Order cancelled successfully", slog.String("order_id", orderID.String()))
	uc.publishOrderCancelled(ctx, order, "")

	return order, nil
}

// ListOrders retrieves all orders with pagination
func (uc *orderUseCase) ListOrders(ctx context.Context, page, pageSize int) ([]*aggregate.Order, int64, error) {
	orders, total, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, ordererrors.ErrOrderListFailed
	}

	// Load lines for each order
	for _, order := range orders {
		lines, err := uc.repo.GetLines(ctx, order.ID)
		if err != nil {
			return nil, 0, ordererrors.ErrOrderLinesFetchFailed
		}
		order.Lines = lines
	}

	return orders, total, nil
}

// ListOrdersByCustomer retrieves all orders for a customer
func (uc *orderUseCase) ListOrdersByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Order, int64, error) {
	orders, total, err := uc.repo.ListByCustomerID(ctx, customerID, page, pageSize)
	if err != nil {
		return nil, 0, ordererrors.ErrOrderListFailed
	}

	// Load lines for each order
	for _, order := range orders {
		lines, err := uc.repo.GetLines(ctx, order.ID)
		if err != nil {
			return nil, 0, ordererrors.ErrOrderLinesFetchFailed
		}
		order.Lines = lines
	}

	return orders, total, nil
}

// ListOrdersByStatus retrieves all orders with specific status
func (uc *orderUseCase) ListOrdersByStatus(ctx context.Context, status aggregate.OrderStatus, page, pageSize int) ([]*aggregate.Order, int64, error) {
	orders, total, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, ordererrors.ErrOrderListFailed
	}

	// Load lines for each order
	for _, order := range orders {
		lines, err := uc.repo.GetLines(ctx, order.ID)
		if err != nil {
			return nil, 0, ordererrors.ErrOrderLinesFetchFailed
		}
		order.Lines = lines
	}

	return orders, total, nil
}

type orderEventItem struct {
	ProductID uuidv7.UUID `json:"product_id"`
	SKU       string      `json:"sku"`
	Quantity  int         `json:"quantity"`
	UnitPrice int64       `json:"unit_price_cents"`
}

type orderConfirmedEvent struct {
	OrderID     uuidv7.UUID      `json:"order_id"`
	CustomerID  uuidv7.UUID      `json:"customer_id"`
	Items       []orderEventItem `json:"items"`
	Currency    string           `json:"currency"`
	ConfirmedBy uuidv7.UUID      `json:"confirmed_by"`
}

type orderCancelledEvent struct {
	OrderID     uuidv7.UUID      `json:"order_id"`
	Items       []orderEventItem `json:"items"`
	Reason      string           `json:"reason"`
	CancelledBy uuidv7.UUID      `json:"cancelled_by"`
}

type orderFulfilledEvent struct {
	OrderID     uuidv7.UUID      `json:"order_id"`
	Items       []orderEventItem `json:"items"`
	FulfilledBy uuidv7.UUID      `json:"fulfilled_by"`
}

func (uc *orderUseCase) publishOrderConfirmed(ctx context.Context, order *aggregate.Order) {
	uc.publishOrderEvent(ctx, bus.TopicOrderConfirmed, order.GetID(), orderConfirmedEvent{
		OrderID:     order.GetID(),
		CustomerID:  order.CustomerID,
		Items:       uc.buildOrderItems(order),
		Currency:    order.Currency,
		ConfirmedBy: order.CustomerID,
	})
}

func (uc *orderUseCase) publishOrderCancelled(ctx context.Context, order *aggregate.Order, reason string) {
	uc.publishOrderEvent(ctx, bus.TopicOrderCancelled, order.GetID(), orderCancelledEvent{
		OrderID:     order.GetID(),
		Items:       uc.buildOrderItems(order),
		Reason:      reason,
		CancelledBy: order.CustomerID,
	})
}

func (uc *orderUseCase) publishOrderFulfilled(ctx context.Context, order *aggregate.Order) {
	uc.publishOrderEvent(ctx, bus.TopicOrderFulfilled, order.GetID(), orderFulfilledEvent{
		OrderID:     order.GetID(),
		Items:       uc.buildOrderItems(order),
		FulfilledBy: order.CustomerID,
	})
}

func (uc *orderUseCase) publishOrderEvent(ctx context.Context, topic string, aggregateID uuidv7.UUID, payload interface{}) {
	if uc.eventBus == nil {
		return
	}

	log := logger.FromContext(ctx)

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Error("Failed to marshal order event payload", slog.Any("error", err), slog.String("topic", topic))
		return
	}

	event := bus.NewBaseEvent(topic, aggregateID)
	event.Meta["payload"] = string(payloadJSON)

	if err := uc.eventBus.Publish(ctx, topic, event); err != nil {
		log.Error("Failed to publish order event", slog.Any("error", err), slog.String("topic", topic))
	}
}

func (uc *orderUseCase) buildOrderItems(order *aggregate.Order) []orderEventItem {
	items := make([]orderEventItem, 0, len(order.Lines))
	for _, line := range order.Lines {
		items = append(items, orderEventItem{
			ProductID: line.ProductID,
			SKU:       "",
			Quantity:  line.Quantity,
			UnitPrice: line.UnitPrice.Amount,
		})
	}
	return items
}
