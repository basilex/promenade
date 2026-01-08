package saga

import (
	"time"

	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// FulfillmentSagaState represents the state of the fulfillment saga
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

// FulfillmentSaga is the aggregate root for order fulfillment workflow
type FulfillmentSaga struct {
	ID         uuidv7.UUID
	OrderID    uuidv7.UUID
	CustomerID uuidv7.UUID

	// Step tracking
	State       FulfillmentSagaState
	CurrentStep int

	CompletedSteps jsonstore.Field[[]string] // DB-agnostic array storage
	FailedStep     *string

	// Step results
	PaymentID      *uuidv7.UUID
	ReservedItems  jsonstore.Field[[]ReservedItem] // DB-agnostic JSON storage
	ShipmentID     *uuidv7.UUID
	TrackingNumber *string

	// Metadata
	StartedAt     time.Time
	CompletedAt   *time.Time
	CancelledAt   *time.Time
	FailureReason *string

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReservedItem represents inventory reservation for fulfillment
type ReservedItem struct {
	ProductID     uuidv7.UUID `json:"product_id"`
	Quantity      int         `json:"quantity"`
	ReservationID uuidv7.UUID `json:"reservation_id"`
}

// NewFulfillmentSaga creates a new fulfillment saga
func NewFulfillmentSaga(orderID, customerID uuidv7.UUID) *FulfillmentSaga {
	now := time.Now().UTC() // Use UTC to avoid timezone issues
	saga := &FulfillmentSaga{
		ID:          uuidv7.New(),
		OrderID:     orderID,
		CustomerID:  customerID,
		State:       FulfillmentSagaStatePending,
		CurrentStep: 0,
		StartedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	saga.CompletedSteps.Set([]string{})
	saga.ReservedItems.Set([]ReservedItem{})

	return saga
}

// StartPaymentProcessing transitions to payment processing state
func (s *FulfillmentSaga) StartPaymentProcessing() {
	s.State = FulfillmentSagaStatePaymentProcessing
	s.CurrentStep = 1
	s.UpdatedAt = time.Now().UTC()
}

// CompletePayment marks payment step as complete and stores payment ID
func (s *FulfillmentSaga) CompletePayment(paymentID uuidv7.UUID) {
	s.PaymentID = &paymentID
	steps := s.CompletedSteps.Get()
	steps = append(steps, "payment")
	s.CompletedSteps.Set(steps)

	s.State = FulfillmentSagaStateInventoryProcessing
	s.CurrentStep = 2
	s.UpdatedAt = time.Now().UTC()
}

// CompleteInventory marks inventory step as complete and stores reserved items
func (s *FulfillmentSaga) CompleteInventory(items []ReservedItem) {
	s.ReservedItems.Set(items)

	steps := s.CompletedSteps.Get()
	steps = append(steps, "inventory")
	s.CompletedSteps.Set(steps)

	s.State = FulfillmentSagaStateShippingProcessing
	s.CurrentStep = 3
	s.UpdatedAt = time.Now().UTC()
}

// CompleteShipping marks shipping step as complete and stores shipment details
func (s *FulfillmentSaga) CompleteShipping(shipmentID uuidv7.UUID, tracking string) {
	s.ShipmentID = &shipmentID
	s.TrackingNumber = &tracking

	steps := s.CompletedSteps.Get()
	steps = append(steps, "shipping")
	s.CompletedSteps.Set(steps)

	now := time.Now()
	s.CompletedAt = &now
	s.State = FulfillmentSagaStateCompleted
	s.UpdatedAt = now
}

// StartCompensation transitions to compensating state when a step fails
func (s *FulfillmentSaga) StartCompensation(step string, reason string) {
	s.State = FulfillmentSagaStateCompensating
	s.FailedStep = &step
	s.FailureReason = &reason
	s.UpdatedAt = time.Now().UTC()
}

// Cancel cancels the saga with a reason
func (s *FulfillmentSaga) Cancel(reason string) {
	s.State = FulfillmentSagaStateCancelled
	s.FailureReason = &reason

	now := time.Now()
	s.CancelledAt = &now
	s.UpdatedAt = now
}

// IsCompleted returns true if saga has completed successfully
func (s *FulfillmentSaga) IsCompleted() bool {
	return s.State == FulfillmentSagaStateCompleted
}

// IsCancelled returns true if saga was cancelled
func (s *FulfillmentSaga) IsCancelled() bool {
	return s.State == FulfillmentSagaStateCancelled
}

// IsInProgress returns true if saga is still executing
func (s *FulfillmentSaga) IsInProgress() bool {
	return s.State == FulfillmentSagaStatePaymentProcessing ||
		s.State == FulfillmentSagaStateInventoryProcessing ||
		s.State == FulfillmentSagaStateShippingProcessing ||
		s.State == FulfillmentSagaStateCompensating
}

// CanCompensate returns true if saga can be compensated
func (s *FulfillmentSaga) CanCompensate() bool {
	return len(s.CompletedSteps.Get()) > 0 && s.State != FulfillmentSagaStateCompleted
}

// GetCompletedSteps returns list of completed steps
func (s *FulfillmentSaga) GetCompletedSteps() []string {
	return s.CompletedSteps.Get()
}

// GetReservedItems returns list of reserved inventory items
func (s *FulfillmentSaga) GetReservedItems() []ReservedItem {
	return s.ReservedItems.Get()
}
