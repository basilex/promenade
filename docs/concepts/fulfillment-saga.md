# Fulfillment Saga

**Distributed transaction orchestration** for order fulfillment process across multiple bounded contexts.

---

## Overview

Fulfillment Saga coordinates the complete order fulfillment process, ensuring **data consistency** across distributed services without distributed transactions. It implements the **Saga pattern** with compensating actions for failure scenarios.

**Pattern**: Orchestration-based Saga (centralized coordinator)

---

## File Structure

```
internal/contexts/order-mgmt/fulfillment/
 saga/                                    # Saga core components
    fulfillment_saga.go                # Saga aggregate entity
    fulfillment_saga_test.go
    orchestrator.go                     # Saga coordinator
    orchestrator_test.go
    step.go                             # Step interface
    state.go                            # State management
    repository.go                       # ISagaRepository interface
    postgres_repository.go             # PostgreSQL implementation
    repository_test.go
 payment_step.go                         # Payment step handler
 payment_step_test.go
 inventory_step.go                       # Inventory step handler
 inventory_step_test.go
 shipping_step.go                        # Shipping step handler
 shipping_step_test.go

migrations/order-mgmt/
 000004_add_fulfillment_sagas.up.sql    # Database schema
 000004_add_fulfillment_sagas.down.sql
```

---

## Business Flow

### Happy Path (Success Scenario)

```
Order Created → Order Confirmed → Fulfillment Started
    ↓
1. Payment Processing
     Validate payment method
     Reserve funds (authorization)
     Capture payment
    ↓
2. Inventory Reservation
     Check stock availability
     Reserve items
     Commit reservation
    ↓
3. Shipping Coordination
     Create shipment
     Notify carrier
     Generate tracking number
    ↓
Order Fulfilled → Customer Notified
```

### Failure Scenarios (Compensating Actions)

**Scenario 1: Payment Fails**
```
Order → Payment  (declined)
    ↓
Compensate: Cancel order
    ↓
Order Status: cancelled (payment_failed)
```

**Scenario 2: Inventory Unavailable**
```
Order → Payment  → Inventory  (out of stock)
    ↓
Compensate:
    1. Refund payment
    2. Cancel order
    ↓
Order Status: cancelled (inventory_unavailable)
```

**Scenario 3: Shipping Fails**
```
Order → Payment  → Inventory  → Shipping  (carrier error)
    ↓
Compensate:
    1. Release inventory reservation
    2. Refund payment
    3. Cancel order
    ↓
Order Status: cancelled (shipping_failed)
```

---

## Saga Architecture

### Components

1. **FulfillmentSaga** (Aggregate Root)
   - Orchestrates distributed transaction
   - Tracks saga state and progress
   - Triggers compensation on failure

2. **SagaOrchestrator** (Coordinator)
   - Executes saga steps sequentially
   - Handles step results (success/failure)
   - Invokes compensating actions

3. **Step Handlers** (Business Logic)
   - PaymentStepHandler
   - InventoryStepHandler
   - ShippingStepHandler

4. **Compensation Handlers**
   - RefundPaymentHandler
   - ReleaseInventoryHandler
   - CancelShipmentHandler

### State Machine

```
                  Pending
                     ↓
            Payment Processing
                /         \
               /           \
        Payment OK    Payment Failed
             ↓              ↓
    Inventory Processing  Compensating
          /       \            ↓
         /         \        Cancelled
   Inventory OK  Inventory Failed
        ↓              ↓
   Shipping       Compensating
      /  \             ↓
     /    \         Cancelled
 Success  Failed
    ↓       ↓
Completed  Compensating
             ↓
          Cancelled
```

### Saga States

| State | Description |
|-------|-------------|
| `pending` | Saga created, not started |
| `payment_processing` | Processing payment |
| `inventory_processing` | Reserving inventory |
| `shipping_processing` | Creating shipment |
| `compensating` | Rolling back changes |
| `completed` | All steps successful |
| `cancelled` | Saga failed, compensated |

---

## Implementation Design

### 1. FulfillmentSaga Aggregate

**Location**: `internal/contexts/order-mgmt/fulfillment/saga/fulfillment_saga.go`

```go
package saga

import (
    "time"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/basilex/promenade/pkg/jsonstore"
)

type FulfillmentSagaState string

const (
    FulfillmentSagaStatePending             FulfillmentSagaState = "pending"
    FulfillmentSagaStatePaymentProcessing   FulfillmentSagaState = "payment_processing"
    FulfillmentSagaStateInventoryProcessing FulfillmentSagaState = "inventory_processing"
    FulfillmentSagaStateShippingProcessing  FulfillmentSagaState = "shipping_processing"
    FulfillmentSagaStateCompensating        FulfillmentSagaState = "compensating"
    FulfillmentSagaStateCompleted           FulfillmentSagaState = "completed"
    FulfillmentSagaStateCancelled           FulfillmentSagaState = "cancelled"
)

type FulfillmentSaga struct {
    ID         uuidv7.UUID
    OrderID    uuidv7.UUID
    CustomerID uuidv7.UUID
    State      FulfillmentSagaState
    
    // Step tracking
    CurrentStep    int
    CompletedSteps jsonstore.Field[[]string]  // DB-agnostic JSON storage
    FailedStep     *string
    
    // Step results
    PaymentID     *uuidv7.UUID
    ReservedItems jsonstore.Field[[]ReservedItem]  // DB-agnostic JSON storage
    ShipmentID    *uuidv7.UUID
    TrackingNumber *string
    
    // Metadata
    StartedAt   time.Time
    CompletedAt *time.Time
    CancelledAt *time.Time
    FailureReason *string
    
    // Audit
    CreatedAt time.Time
    UpdatedAt time.Time
}

type ReservedItem struct {
    ProductID       uuidv7.UUID
    Quantity        int
    ReservationID   uuidv7.UUID
}

// Factory method
func NewFulfillmentSaga(orderID, customerID uuidv7.UUID) *FulfillmentSaga {
    now := time.Now()
    saga := &FulfillmentSaga{
        ID:             uuidv7.New(),
        OrderID:        orderID,
        CustomerID:     customerID,
        State:          FulfillmentSagaStatePending,
        CurrentStep:    0,
        StartedAt:      now,
        CreatedAt:      now,
        UpdatedAt:      now,
    }
    saga.CompletedSteps.Set([]string{})
    saga.ReservedItems.Set([]ReservedItem{})
    return saga
}

// State transitions
func (s *FulfillmentSaga) StartPaymentProcessing() {
    s.State = FulfillmentSagaStatePaymentProcessing
    s.CurrentStep = 1
    s.UpdatedAt = time.Now()
}

func (s *FulfillmentSaga) CompletePayment(paymentID uuidv7.UUID) {
    s.PaymentID = &paymentID
    steps := s.CompletedSteps.Get()
    steps = append(steps, "payment")
    s.CompletedSteps.Set(steps)
    s.State = FulfillmentSagaStateInventoryProcessing
    s.CurrentStep = 2
    s.UpdatedAt = time.Now()
}

func (s *FulfillmentSaga) CompleteInventory(items []ReservedItem) {
    s.ReservedItems.Set(items)
    steps := s.CompletedSteps.Get()
    steps = append(steps, "inventory")
    s.CompletedSteps.Set(steps)
    s.State = FulfillmentSagaStateShippingProcessing
    s.CurrentStep = 3
    s.UpdatedAt = time.Now()
}

func (s *FulfillmentSaga) CompleteShipping(shipmentID uuidv7.UUID, tracking string) {
    s.ShipmentID = &shipmentID
    s.TrackingNumber = &tracking
    steps := s.CompletedSteps.Get()
    steps = append(steps, "shipping")
    s.CompletedSteps.Set(steps)
    s.State = FulfillmentSagaStateCompleted
    now := time.Now()
    s.CompletedAt = &now
    s.UpdatedAt = now
}

func (s *FulfillmentSaga) StartCompensation(step string, reason string) {
    s.State = FulfillmentSagaStateCompensating
    s.FailedStep = &step
    s.FailureReason = &reason
    s.UpdatedAt = time.Now()
}

func (s *FulfillmentSaga) Cancel(reason string) {
    s.State = FulfillmentSagaStateCancelled
    s.FailureReason = &reason
    now := time.Now()
    s.CancelledAt = &now
    s.UpdatedAt = now
}
```

### 2. SagaOrchestrator

**Location**: `internal/contexts/order-mgmt/fulfillment/saga/orchestrator.go`

```go
package saga

import (
    "context"
    "fmt"
)

type StepResult struct {
    Success bool
    Data    interface{}
    Error   error
}

type SagaStep interface {
    Execute(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error)
    Compensate(ctx context.Context, saga *FulfillmentSaga) error
    Name() string
}

type Orchestrator struct {
    steps      []SagaStep
    repository ISagaRepository
}

func NewOrchestrator(repo ISagaRepository) *Orchestrator {
    return &Orchestrator{
        steps:      []SagaStep{},
        repository: repo,
    }
}

func (o *Orchestrator) AddStep(step SagaStep) {
    o.steps = append(o.steps, step)
}

func (o *Orchestrator) Execute(ctx context.Context, saga *FulfillmentSaga) error {
    // Execute steps sequentially
    for i := saga.CurrentStep; i < len(o.steps); i++ {
        step := o.steps[i]
        
        // Execute step
        result, err := step.Execute(ctx, saga)
        if err != nil || !result.Success {
            // Step failed - start compensation
            reason := "unknown error"
            if err != nil {
                reason = err.Error()
            } else if result.Error != nil {
                reason = result.Error.Error()
            }
            
            saga.StartCompensation(step.Name(), reason)
            if err := o.repository.Update(ctx, saga); err != nil {
                return fmt.Errorf("failed to update saga: %w", err)
            }
            
            // Compensate completed steps in reverse order
            return o.compensate(ctx, saga, i)
        }
        
        // Save progress
        if err := o.repository.Update(ctx, saga); err != nil {
            return fmt.Errorf("failed to update saga: %w", err)
        }
    }
    
    return nil
}

func (o *Orchestrator) compensate(ctx context.Context, saga *FulfillmentSaga, failedStepIndex int) error {
    // Compensate in reverse order (LIFO)
    for i := failedStepIndex - 1; i >= 0; i-- {
        step := o.steps[i]
        if err := step.Compensate(ctx, saga); err != nil {
            // Log compensation failure but continue
            // In production, would retry or alert
            fmt.Printf("Compensation failed for step %s: %v\n", step.Name(), err)
        }
    }
    
    // Mark saga as cancelled
    reason := "saga compensation completed"
    if saga.FailureReason != nil {
        reason = *saga.FailureReason
    }
    saga.Cancel(reason)
    
    return o.repository.Update(ctx, saga)
}
```

### 3. Step Handlers

#### Payment Step

**Location**: `internal/contexts/order-mgmt/fulfillment/payment_step.go`

```go
package fulfillment

import (
    "context"
    "fmt"
    "github.com/basilex/promenade/internal/contexts/order-mgmt/fulfillment/saga"
    "github.com/basilex/promenade/internal/contexts/billing/payment"
)

type PaymentStep struct {
    paymentUC payment.IUseCase
}

func NewPaymentStep(paymentUC payment.IUseCase) *PaymentStep {
    return &PaymentStep{paymentUC: paymentUC}
}

func (s *PaymentStep) Name() string {
    return "payment"
}

func (s *PaymentStep) Execute(ctx context.Context, saga *saga.FulfillmentSaga) (*saga.StepResult, error) {
    saga.StartPaymentProcessing()
    
    // Get order details to calculate amount
    // ... (fetch order from repository)
    
    // Process payment
    payment, err := s.paymentUC.ProcessPayment(ctx, saga.CustomerID, orderAmount, "USD")
    if err != nil {
        return &saga.StepResult{
            Success: false,
            Error:   fmt.Errorf("payment processing failed: %w", err),
        }, nil
    }
    
    // Update saga with payment ID
    saga.CompletePayment(payment.ID)
    
    return &saga.StepResult{
        Success: true,
        Data:    payment,
    }, nil
}

func (s *PaymentStep) Compensate(ctx context.Context, saga *saga.FulfillmentSaga) error {
    if saga.PaymentID == nil {
        return nil // Nothing to compensate
    }
    
    // Refund payment
    return s.paymentUC.RefundPayment(ctx, *saga.PaymentID, "order_fulfillment_failed")
}
```

#### Inventory Step

**Location**: `internal/contexts/order-mgmt/fulfillment/inventory_step.go`

```go
package fulfillment

import (
    "context"
    "fmt"
    "github.com/basilex/promenade/internal/contexts/order-mgmt/fulfillment/saga"
    "github.com/basilex/promenade/internal/contexts/warehouse/inventory"
)

type InventoryStep struct {
    inventoryUC inventory.IUseCase
}

func NewInventoryStep(inventoryUC inventory.IUseCase) *InventoryStep {
    return &InventoryStep{inventoryUC: inventoryUC}
}

func (s *InventoryStep) Name() string {
    return "inventory"
}

func (s *InventoryStep) Execute(ctx context.Context, saga *saga.FulfillmentSaga) (*saga.StepResult, error) {
    // Get order line items
    // ... (fetch from order repository)
    
    var reservedItems []saga.ReservedItem
    
    // Reserve each item
    for _, item := range orderItems {
        reservation, err := s.inventoryUC.ReserveStock(ctx, item.ProductID, item.Quantity, saga.OrderID)
        if err != nil {
            // Out of stock - cannot fulfill
            return &saga.StepResult{
                Success: false,
                Error:   fmt.Errorf("inventory reservation failed: %w", err),
            }, nil
        }
        
        reservedItems = append(reservedItems, saga.ReservedItem{
            ProductID:     item.ProductID,
            Quantity:      item.Quantity,
            ReservationID: reservation.ID,
        })
    }
    
    // Update saga with reservations
    saga.CompleteInventory(reservedItems)
    
    return &saga.StepResult{
        Success: true,
        Data:    reservedItems,
    }, nil
}

func (s *InventoryStep) Compensate(ctx context.Context, saga *saga.FulfillmentSaga) error {
    if len(saga.ReservedItems) == 0 {
        return nil
    }
    
    // Release all reservations
    for _, item := range saga.ReservedItems {
        if err := s.inventoryUC.ReleaseReservation(ctx, item.ReservationID); err != nil {
            // Log but continue - don't want to stop compensation
            fmt.Printf("Failed to release reservation %s: %v\n", item.ReservationID, err)
        }
    }
    
    return nil
}
```

#### Shipping Step

**Location**: `internal/contexts/order-mgmt/fulfillment/shipping_step.go`

```go
package fulfillment

import (
    "context"
    "fmt"
    "github.com/basilex/promenade/internal/contexts/order-mgmt/fulfillment/saga"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type ShippingStep struct {
    // Future: shipping service client
}

func NewShippingStep() *ShippingStep {
    return &ShippingStep{}
}

func (s *ShippingStep) Name() string {
    return "shipping"
}

func (s *ShippingStep) Execute(ctx context.Context, saga *saga.FulfillmentSaga) (*saga.StepResult, error) {
    // Create shipment (mock for now)
    shipmentID := uuidv7.New()
    trackingNumber := fmt.Sprintf("TRACK-%s", shipmentID.String()[:8])
    
    // Future: integrate with shipping provider API
    // - Create shipment
    // - Generate label
    // - Notify carrier
    
    saga.CompleteShipping(shipmentID, trackingNumber)
    
    return &saga.StepResult{
        Success: true,
        Data: map[string]interface{}{
            "shipment_id":      shipmentID,
            "tracking_number":  trackingNumber,
        },
    }, nil
}

func (s *ShippingStep) Compensate(ctx context.Context, saga *saga.FulfillmentSaga) error {
    if saga.ShipmentID == nil {
        return nil
    }
    
    // Cancel shipment
    // Future: call shipping provider API to cancel
    fmt.Printf("Cancelling shipment %s\n", *saga.ShipmentID)
    
    return nil
}
```

---

## Database Schema

### fulfillment_sagas Table

```sql
CREATE TABLE order_fulfillment_sagas (
    -- Identity (UUID generated by application via uuidv7.New())
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES order_orders(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL,
    
    -- State
    state VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_step INTEGER NOT NULL DEFAULT 0,
    completed_steps TEXT,  -- JSON array serialized as TEXT (db-agnostic)
    failed_step VARCHAR(50),
    
    -- Step results (JSON serialized as TEXT for db-agnostic support)
    payment_id UUID,
    reserved_items TEXT,  -- JSON array serialized as TEXT
    shipment_id UUID,
    tracking_number VARCHAR(100),
    
    -- Metadata
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    failure_reason TEXT,
    
    -- Audit
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT chk_fulfillment_saga_state CHECK (
        state IN ('pending', 'payment_processing', 'inventory_processing', 
                  'shipping_processing', 'compensating', 'completed', 'cancelled')
    )
);

-- Indexes
CREATE INDEX idx_fulfillment_sagas_order_id ON order_fulfillment_sagas(order_id);
CREATE INDEX idx_fulfillment_sagas_state ON order_fulfillment_sagas(state) WHERE cancelled_at IS NULL;
CREATE INDEX idx_fulfillment_sagas_customer_id ON order_fulfillment_sagas(customer_id);
```

---

## Integration with Order Context

### HTTP Endpoint

```
POST /api/v1/orders/:id/fulfill
```

**Request**: Empty (order ID in URL)

**Response**:
```json
{
  "status": "success",
  "data": {
    "saga_id": "01JH...",
    "order_id": "01JH...",
    "state": "completed",
    "payment_id": "01JH...",
    "shipment_id": "01JH...",
    "tracking_number": "TRACK-12345678"
  }
}
```

**Error Response** (compensation):
```json
{
  "status": "error",
  "error": {
    "code": "FULFILLMENT_FAILED",
    "message": "Inventory reservation failed: Product out of stock",
    "saga_id": "01JH...",
    "failed_step": "inventory"
  }
}
```

### Order Handler

```go
func (h *OrderHandler) FulfillOrder(c *gin.Context) {
    orderID, _ := uuidv7.Parse(c.Param("id"))
    
    // Create and execute saga
    saga, err := h.fulfillmentUC.FulfillOrder(c.Request.Context(), orderID)
    if err != nil {
        // Saga failed with compensation
        response.Error(c, 500, "FULFILLMENT_FAILED", err.Error())
        return
    }
    
    response.Success(c, ToFulfillmentSagaResponse(saga))
}
```

---

## Testing Strategy

### Unit Tests (Saga Logic)

```go
func TestFulfillmentSaga_HappyPath(t *testing.T) {
    saga := NewFulfillmentSaga(orderID, customerID)
    
    saga.StartPaymentProcessing()
    assert.Equal(t, FulfillmentSagaStatePaymentProcessing, saga.State)
    
    saga.CompletePayment(paymentID)
    assert.Equal(t, FulfillmentSagaStateInventoryProcessing, saga.State)
    assert.Contains(t, saga.CompletedSteps, "payment")
}
```

### Integration Tests (End-to-End)

```go
func TestOrchestrator_FullFlow_Success(t *testing.T) {
    // Setup: real DB, mock external services
    orchestrator := setupOrchestrator(t)
    saga := NewFulfillmentSaga(orderID, customerID)
    
    // Execute
    err := orchestrator.Execute(ctx, saga)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, FulfillmentSagaStateCompleted, saga.State)
    assert.Len(t, saga.CompletedSteps, 3)
}

func TestOrchestrator_InventoryFails_Compensation(t *testing.T) {
    // Setup: inventory step will fail
    orchestrator := setupOrchestrator(t)
    saga := NewFulfillmentSaga(orderID, customerID)
    
    // Execute
    err := orchestrator.Execute(ctx, saga)
    
    // Verify compensation
    assert.NoError(t, err) // Compensation succeeds
    assert.Equal(t, FulfillmentSagaStateCancelled, saga.State)
    assert.NotNil(t, saga.FailedStep)
    assert.Equal(t, "inventory", *saga.FailedStep)
    
    // Verify payment was refunded (compensation)
    payment := getPayment(t, *saga.PaymentID)
    assert.Equal(t, "refunded", payment.Status)
}
```

### Test Scenarios (50+ tests target)

1. **Happy Path** (10 tests)
   - All steps succeed
   - State transitions correct
   - Data persisted correctly

2. **Payment Failures** (10 tests)
   - Card declined
   - Insufficient funds
   - Payment service unavailable

3. **Inventory Failures** (10 tests)
   - Out of stock
   - Partial stock
   - Reservation timeout

4. **Shipping Failures** (5 tests)
   - Carrier API error
   - Invalid address
   - Service unavailable

5. **Compensation** (10 tests)
   - Payment refund works
   - Inventory released
   - Shipment cancelled
   - Multiple compensations

6. **Edge Cases** (5 tests)
   - Saga already completed
   - Duplicate execution
   - Timeout handling

---

## Performance Considerations

### Saga Execution Time

- **Target**: < 5 seconds for happy path
- **Components**:
  - Payment: ~500ms (API call)
  - Inventory: ~200ms (DB query + reservation)
  - Shipping: ~300ms (API call)
  - Total: ~1 second (plus coordination overhead)

### Retry Strategy

- **Transient failures**: Retry with exponential backoff
- **Permanent failures**: Immediate compensation
- **Timeout**: 30 seconds per step

### Idempotency

- All saga steps must be idempotent
- Use saga_id as idempotency key
- Prevent duplicate execution

---

## Monitoring & Observability

### Metrics

- Saga success rate (%)
- Average execution time
- Compensation rate by step
- Failed step distribution

### Logging

```go
logger.Info("Saga started",
    "saga_id", saga.ID,
    "order_id", saga.OrderID,
)

logger.Info("Step completed",
    "saga_id", saga.ID,
    "step", step.Name(),
    "duration_ms", duration,
)

logger.Error("Saga failed, compensating",
    "saga_id", saga.ID,
    "failed_step", *saga.FailedStep,
    "reason", *saga.FailureReason,
)
```

### Alerts

- Compensation rate > 5%
- Average execution time > 10s
- Any step timing out

---

## Future Enhancements

1. **Event Sourcing**: Store all saga events for replay
2. **Parallel Steps**: Execute independent steps concurrently
3. **Manual Intervention**: UI for manual compensation
4. **Saga Recovery**: Restart failed sagas
5. **A/B Testing**: Test different fulfillment strategies

---

## References

- [Saga Pattern](https://microservices.io/patterns/data/saga.html)
- [Orchestration vs Choreography](https://microservices.io/patterns/data/saga.html)
- [Compensating Transactions](https://docs.microsoft.com/en-us/azure/architecture/patterns/compensating-transaction)

---

**Status**: Design Complete  
**Next**: Implementation (Phase 7.2)  
**Last Updated**: January 8, 2026
