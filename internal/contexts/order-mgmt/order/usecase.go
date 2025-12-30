package order

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// IUseCase defines the interface for order business logic
type IUseCase interface {
	// CreateOrder creates a new order
	CreateOrder(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, currency string) (*Order, error)

	// GetOrder retrieves an order by ID
	GetOrder(ctx context.Context, orderID uuidv7.UUID) (*Order, error)

	// GetOrderByNumber retrieves an order by order number
	GetOrderByNumber(ctx context.Context, orderNumber string) (*Order, error)

	// AddOrderLine adds a line item to an order
	AddOrderLine(ctx context.Context, orderID, productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) (*Order, error)

	// RemoveOrderLine removes a line item from an order
	RemoveOrderLine(ctx context.Context, orderID, lineID uuidv7.UUID) (*Order, error)

	// UpdateOrderLineQuantity updates the quantity of a line item
	UpdateOrderLineQuantity(ctx context.Context, orderID, lineID uuidv7.UUID, quantity int) (*Order, error)

	// ConfirmOrder confirms an order
	ConfirmOrder(ctx context.Context, orderID uuidv7.UUID) (*Order, error)

	// StartProcessing moves order to processing status
	StartProcessing(ctx context.Context, orderID uuidv7.UUID) (*Order, error)

	// MarkFulfilled marks order as fulfilled
	MarkFulfilled(ctx context.Context, orderID uuidv7.UUID) (*Order, error)

	// CancelOrder cancels an order
	CancelOrder(ctx context.Context, orderID uuidv7.UUID) (*Order, error)

	// ListOrders retrieves all orders with pagination
	ListOrders(ctx context.Context, page, pageSize int) ([]*Order, int64, error)

	// ListOrdersByCustomer retrieves all orders for a customer
	ListOrdersByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Order, int64, error)

	// ListOrdersByStatus retrieves all orders with specific status
	ListOrdersByStatus(ctx context.Context, status OrderStatus, page, pageSize int) ([]*Order, int64, error)
}

// useCase implements IUseCase
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new order use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateOrder creates a new order
func (uc *useCase) CreateOrder(ctx context.Context, customerID uuidv7.UUID, companyID *uuidv7.UUID, currency string) (*Order, error) {
	log := logger.FromContext(ctx)

	// Create new order
	order, err := NewOrder(customerID, currency)
	if err != nil {
		log.Error("Failed to create order entity", slog.Any("error", err))
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Set company ID if provided
	if companyID != nil {
		order.CompanyID = companyID
	}

	// Persist order
	if err := uc.repo.Create(ctx, order); err != nil {
		log.Error("Failed to persist order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to persist order: %w", err)
	}

	log.Info("Order created successfully",
		slog.String("order_id", order.ID.String()),
		slog.String("order_number", order.OrderNumber),
		slog.String("customer_id", customerID.String()),
	)

	return order, nil
}

// GetOrder retrieves an order by ID
func (uc *useCase) GetOrder(ctx context.Context, orderID uuidv7.UUID) (*Order, error) {
	log := logger.FromContext(ctx)

	order, err := uc.repo.GetByID(ctx, orderID)
	if err != nil {
		log.Error("Failed to get order", slog.Any("error", err), slog.String("order_id", orderID.String()))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Load order lines
	lines, err := uc.repo.GetLines(ctx, orderID)
	if err != nil {
		log.Error("Failed to get order lines", slog.Any("error", err))
		return nil, fmt.Errorf("failed to get order lines: %w", err)
	}
	order.Lines = lines

	return order, nil
}

// GetOrderByNumber retrieves an order by order number
func (uc *useCase) GetOrderByNumber(ctx context.Context, orderNumber string) (*Order, error) {
	log := logger.FromContext(ctx)

	order, err := uc.repo.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		log.Error("Failed to get order by number", slog.Any("error", err), slog.String("order_number", orderNumber))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Load order lines
	lines, err := uc.repo.GetLines(ctx, order.ID)
	if err != nil {
		log.Error("Failed to get order lines", slog.Any("error", err))
		return nil, fmt.Errorf("failed to get order lines: %w", err)
	}
	order.Lines = lines

	return order, nil
}

// AddOrderLine adds a line item to an order
func (uc *useCase) AddOrderLine(ctx context.Context, orderID, productID uuidv7.UUID, quantity int, unitPrice valueobject.Money) (*Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Add line to order (business logic)
	if err := order.AddLine(productID, quantity, unitPrice); err != nil {
		log.Error("Failed to add line to order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to add line: %w", err)
	}

	// Persist new line
	newLine := order.Lines[len(order.Lines)-1]
	if err := uc.repo.CreateLine(ctx, &newLine); err != nil {
		log.Error("Failed to persist order line", slog.Any("error", err))
		return nil, fmt.Errorf("failed to persist line: %w", err)
	}

	// Update order total
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	log.Info("Order line added successfully",
		slog.String("order_id", orderID.String()),
		slog.String("line_id", newLine.ID.String()),
	)

	return order, nil
}

// RemoveOrderLine removes a line item from an order
func (uc *useCase) RemoveOrderLine(ctx context.Context, orderID, lineID uuidv7.UUID) (*Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Remove line from order (business logic)
	if err := order.RemoveLine(lineID); err != nil {
		log.Error("Failed to remove line from order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to remove line: %w", err)
	}

	// Delete line from database
	if err := uc.repo.DeleteLine(ctx, lineID); err != nil {
		log.Error("Failed to delete order line", slog.Any("error", err))
		return nil, fmt.Errorf("failed to delete line: %w", err)
	}

	// Update order total
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	log.Info("Order line removed successfully",
		slog.String("order_id", orderID.String()),
		slog.String("line_id", lineID.String()),
	)

	return order, nil
}

// UpdateOrderLineQuantity updates the quantity of a line item
func (uc *useCase) UpdateOrderLineQuantity(ctx context.Context, orderID, lineID uuidv7.UUID, quantity int) (*Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Update line quantity (business logic)
	if err := order.UpdateLineQuantity(lineID, quantity); err != nil {
		log.Error("Failed to update line quantity", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update quantity: %w", err)
	}

	// Find and update the line
	for _, line := range order.Lines {
		if line.ID == lineID {
			if err := uc.repo.UpdateLine(ctx, &line); err != nil {
				log.Error("Failed to update order line", slog.Any("error", err))
				return nil, fmt.Errorf("failed to update line: %w", err)
			}
			break
		}
	}

	// Update order total
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	log.Info("Order line quantity updated successfully",
		slog.String("order_id", orderID.String()),
		slog.String("line_id", lineID.String()),
		slog.Int("quantity", quantity),
	)

	return order, nil
}

// ConfirmOrder confirms an order
func (uc *useCase) ConfirmOrder(ctx context.Context, orderID uuidv7.UUID) (*Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Confirm order (business logic)
	if err := order.Confirm(); err != nil {
		log.Error("Failed to confirm order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to confirm order: %w", err)
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	log.Info("Order confirmed successfully", slog.String("order_id", orderID.String()))

	return order, nil
}

// StartProcessing moves order to processing status
func (uc *useCase) StartProcessing(ctx context.Context, orderID uuidv7.UUID) (*Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Start processing (business logic)
	if err := order.StartProcessing(); err != nil {
		log.Error("Failed to start processing", slog.Any("error", err))
		return nil, fmt.Errorf("failed to start processing: %w", err)
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	log.Info("Order processing started", slog.String("order_id", orderID.String()))

	return order, nil
}

// MarkFulfilled marks order as fulfilled
func (uc *useCase) MarkFulfilled(ctx context.Context, orderID uuidv7.UUID) (*Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Mark fulfilled (business logic)
	if err := order.MarkFulfilled(); err != nil {
		log.Error("Failed to mark fulfilled", slog.Any("error", err))
		return nil, fmt.Errorf("failed to mark fulfilled: %w", err)
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	log.Info("Order marked as fulfilled", slog.String("order_id", orderID.String()))

	return order, nil
}

// CancelOrder cancels an order
func (uc *useCase) CancelOrder(ctx context.Context, orderID uuidv7.UUID) (*Order, error) {
	log := logger.FromContext(ctx)

	// Get order
	order, err := uc.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Cancel order (business logic)
	if err := order.Cancel(); err != nil {
		log.Error("Failed to cancel order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// Update order
	if err := uc.repo.Update(ctx, order); err != nil {
		log.Error("Failed to update order", slog.Any("error", err))
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	log.Info("Order cancelled successfully", slog.String("order_id", orderID.String()))

	return order, nil
}

// ListOrders retrieves all orders with pagination
func (uc *useCase) ListOrders(ctx context.Context, page, pageSize int) ([]*Order, int64, error) {
	orders, total, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	// Load lines for each order
	for _, order := range orders {
		lines, err := uc.repo.GetLines(ctx, order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load lines for order %s: %w", order.ID.String(), err)
		}
		order.Lines = lines
	}

	return orders, total, nil
}

// ListOrdersByCustomer retrieves all orders for a customer
func (uc *useCase) ListOrdersByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Order, int64, error) {
	orders, total, err := uc.repo.ListByCustomerID(ctx, customerID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders by customer: %w", err)
	}

	// Load lines for each order
	for _, order := range orders {
		lines, err := uc.repo.GetLines(ctx, order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load lines for order %s: %w", order.ID.String(), err)
		}
		order.Lines = lines
	}

	return orders, total, nil
}

// ListOrdersByStatus retrieves all orders with specific status
func (uc *useCase) ListOrdersByStatus(ctx context.Context, status OrderStatus, page, pageSize int) ([]*Order, int64, error) {
	orders, total, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders by status: %w", err)
	}

	// Load lines for each order
	for _, order := range orders {
		lines, err := uc.repo.GetLines(ctx, order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load lines for order %s: %w", order.ID.String(), err)
		}
		order.Lines = lines
	}

	return orders, total, nil
}
