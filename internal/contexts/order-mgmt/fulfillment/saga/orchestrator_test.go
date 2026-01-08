package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock Step Implementation
type mockStep struct {
	name           string
	executeFunc    func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error)
	compensateFunc func(ctx context.Context, saga *FulfillmentSaga) error
}

func (m *mockStep) Execute(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, saga)
	}
	return &StepResult{Success: true}, nil
}

func (m *mockStep) Compensate(ctx context.Context, saga *FulfillmentSaga) error {
	if m.compensateFunc != nil {
		return m.compensateFunc(ctx, saga)
	}
	return nil
}

func (m *mockStep) Name() string {
	return m.name
}

// Mock Repository Implementation
type mockRepository struct {
	sagas       map[uuidv7.UUID]*FulfillmentSaga
	updateCalls int
	saveCalls   int
	saveError   error
	updateError error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		sagas: make(map[uuidv7.UUID]*FulfillmentSaga),
	}
}

func (m *mockRepository) Save(ctx context.Context, saga *FulfillmentSaga) error {
	m.saveCalls++
	if m.saveError != nil {
		return m.saveError
	}
	m.sagas[saga.ID] = saga
	return nil
}

func (m *mockRepository) Update(ctx context.Context, saga *FulfillmentSaga) error {
	m.updateCalls++
	if m.updateError != nil {
		return m.updateError
	}
	m.sagas[saga.ID] = saga
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uuidv7.UUID) (*FulfillmentSaga, error) {
	saga, ok := m.sagas[id]
	if !ok {
		return nil, errors.New("saga not found")
	}
	return saga, nil
}

func (m *mockRepository) FindByOrderID(ctx context.Context, orderID uuidv7.UUID) (*FulfillmentSaga, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepository) FindInProgressSagas(ctx context.Context) ([]*FulfillmentSaga, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	return errors.New("delete not supported for sagas")
}

// Test: NewOrchestrator
func TestNewOrchestrator(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	assert.NotNil(t, orch)
	assert.Equal(t, 0, orch.GetStepCount())
}

// Test: AddStep
func TestOrchestrator_AddStep(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step1 := &mockStep{name: "payment"}
	step2 := &mockStep{name: "inventory"}

	orch.AddStep(step1)
	assert.Equal(t, 1, orch.GetStepCount())

	orch.AddStep(step2)
	assert.Equal(t, 2, orch.GetStepCount())
}

// Test: Execute - Empty Steps
func TestOrchestrator_Execute_EmptySteps(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.NoError(t, err)
	assert.Equal(t, 0, repo.updateCalls)
}

// Test: Execute - Single Step Success
func TestOrchestrator_Execute_SingleStepSuccess(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.StartPaymentProcessing()
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
	}
	orch.AddStep(step)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.NoError(t, err)
	assert.Equal(t, 1, repo.updateCalls)
	assert.Len(t, saga.GetCompletedSteps(), 1)
	assert.Equal(t, "payment", saga.GetCompletedSteps()[0])
}

// Test: Execute - Multiple Steps Success
func TestOrchestrator_Execute_MultipleStepsSuccess(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step1 := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.StartPaymentProcessing()
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
	}
	step2 := &mockStep{
		name: "inventory",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			items := []ReservedItem{{ProductID: uuidv7.New(), ReservationID: uuidv7.New(), Quantity: 1}}
			saga.CompleteInventory(items)
			return &StepResult{Success: true}, nil
		},
	}
	step3 := &mockStep{
		name: "shipping",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompleteShipping(uuidv7.New(), "TRACK123")
			return &StepResult{Success: true}, nil
		},
	}

	orch.AddStep(step1)
	orch.AddStep(step2)
	orch.AddStep(step3)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.NoError(t, err)
	assert.Equal(t, 3, repo.updateCalls) // One per step
	assert.Len(t, saga.GetCompletedSteps(), 3)
	assert.Equal(t, FulfillmentSagaStateCompleted, saga.State)
}

// Test: Execute - Resume From Current Step
func TestOrchestrator_Execute_ResumeFromCurrentStep(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step1ExecuteCalls := 0
	step2ExecuteCalls := 0

	step1 := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			step1ExecuteCalls++
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
	}
	step2 := &mockStep{
		name: "inventory",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			step2ExecuteCalls++
			items := []ReservedItem{{ProductID: uuidv7.New(), ReservationID: uuidv7.New(), Quantity: 1}}
			saga.CompleteInventory(items)
			return &StepResult{Success: true}, nil
		},
	}

	orch.AddStep(step1)
	orch.AddStep(step2)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.CurrentStep = 1 // Already completed payment

	err := orch.Execute(context.Background(), saga)

	assert.NoError(t, err)
	assert.Equal(t, 0, step1ExecuteCalls) // Not called (already done)
	assert.Equal(t, 1, step2ExecuteCalls) // Called once
	assert.Equal(t, 1, repo.updateCalls)
}

// Test: Execute - First Step Fails (No Compensation)
func TestOrchestrator_Execute_FirstStepFails(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			return nil, errors.New("payment failed")
		},
	}
	orch.AddStep(step)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment failed")
	assert.Equal(t, FulfillmentSagaStateCompensating, saga.State)
	assert.Equal(t, "payment", *saga.FailedStep)
	assert.Equal(t, 1, repo.updateCalls) // One for compensation start
}

// Test: Execute - Second Step Fails (Compensate First)
func TestOrchestrator_Execute_SecondStepFails(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step1CompensateCalls := 0

	step1 := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
		compensateFunc: func(ctx context.Context, saga *FulfillmentSaga) error {
			step1CompensateCalls++
			return nil
		},
	}
	step2 := &mockStep{
		name: "inventory",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			return nil, errors.New("inventory failed")
		},
	}

	orch.AddStep(step1)
	orch.AddStep(step2)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inventory failed")
	assert.Equal(t, 1, step1CompensateCalls) // Compensated step 0
	assert.Equal(t, FulfillmentSagaStateCompensating, saga.State)
	assert.Equal(t, "inventory", *saga.FailedStep)
}

// Test: Execute - Third Step Fails (Compensate in LIFO Order)
func TestOrchestrator_Execute_ThirdStepFails(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	compensationOrder := []string{}

	step1 := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
		compensateFunc: func(ctx context.Context, saga *FulfillmentSaga) error {
			compensationOrder = append(compensationOrder, "payment")
			return nil
		},
	}
	step2 := &mockStep{
		name: "inventory",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			items := []ReservedItem{{ProductID: uuidv7.New(), ReservationID: uuidv7.New(), Quantity: 1}}
			saga.CompleteInventory(items)
			return &StepResult{Success: true}, nil
		},
		compensateFunc: func(ctx context.Context, saga *FulfillmentSaga) error {
			compensationOrder = append(compensationOrder, "inventory")
			return nil
		},
	}
	step3 := &mockStep{
		name: "shipping",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			return nil, errors.New("shipping failed")
		},
	}

	orch.AddStep(step1)
	orch.AddStep(step2)
	orch.AddStep(step3)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.Error(t, err)
	assert.Equal(t, []string{"inventory", "payment"}, compensationOrder) // LIFO order
	assert.Equal(t, FulfillmentSagaStateCompensating, saga.State)
}

// Test: Execute - Step Result False (Not Error)
func TestOrchestrator_Execute_StepResultFalse(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			return &StepResult{Success: false, Error: errors.New("validation failed")}, nil
		},
	}
	orch.AddStep(step)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Equal(t, FulfillmentSagaStateCompensating, saga.State)
}

// Test: Execute - Context Cancellation
func TestOrchestrator_Execute_ContextCancellation(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	step := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			// Check context
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
	}
	orch.AddStep(step)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(ctx, saga)

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

// Test: Execute - Repository Update Fails
func TestOrchestrator_Execute_RepositoryUpdateFails(t *testing.T) {
	repo := newMockRepository()
	repo.updateError = errors.New("database connection lost")
	orch := NewOrchestrator(repo)

	step := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
	}
	orch.AddStep(step)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection lost")
}

// Test: Resume - Successful Resume
func TestOrchestrator_Resume_SuccessfulResume(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step1 := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
	}
	step2 := &mockStep{
		name: "inventory",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			items := []ReservedItem{{ProductID: uuidv7.New(), ReservationID: uuidv7.New(), Quantity: 1}}
			saga.CompleteInventory(items)
			return &StepResult{Success: true}, nil
		},
	}

	orch.AddStep(step1)
	orch.AddStep(step2)

	// Saga already completed first step
	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.State = FulfillmentSagaStateInventoryProcessing
	saga.CurrentStep = 1
	steps := []string{"payment"}
	saga.CompletedSteps.Set(steps)

	err := orch.Resume(context.Background(), saga)

	assert.NoError(t, err)
	assert.Len(t, saga.GetCompletedSteps(), 2)
	assert.Equal(t, FulfillmentSagaStateShippingProcessing, saga.State)
}

// Test: Resume - Not In Progress
func TestOrchestrator_Resume_NotInProgress(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())
	saga.State = FulfillmentSagaStateCompleted // Not in progress

	err := orch.Resume(context.Background(), saga)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "saga is not in progress")
}

// Test: Compensation Failure Continues
func TestOrchestrator_Execute_CompensationFailureContinues(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	compensationCalls := []string{}

	step1 := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true}, nil
		},
		compensateFunc: func(ctx context.Context, saga *FulfillmentSaga) error {
			compensationCalls = append(compensationCalls, "payment")
			return errors.New("compensation failed for payment")
		},
	}
	step2 := &mockStep{
		name: "inventory",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			items := []ReservedItem{{ProductID: uuidv7.New(), ReservationID: uuidv7.New(), Quantity: 1}}
			saga.CompleteInventory(items)
			return &StepResult{Success: true}, nil
		},
		compensateFunc: func(ctx context.Context, saga *FulfillmentSaga) error {
			compensationCalls = append(compensationCalls, "inventory")
			return nil
		},
	}
	step3 := &mockStep{
		name: "shipping",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			return nil, errors.New("shipping failed")
		},
	}

	orch.AddStep(step1)
	orch.AddStep(step2)
	orch.AddStep(step3)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	assert.Error(t, err)
	// Both compensation functions were called despite first one failing
	assert.Equal(t, []string{"inventory", "payment"}, compensationCalls)
	assert.Equal(t, FulfillmentSagaStateCompensating, saga.State)
}

// Test: Full Workflow With Mock Repository
func TestOrchestrator_FullWorkflow(t *testing.T) {
	repo := newMockRepository()
	orch := NewOrchestrator(repo)

	step1 := &mockStep{
		name: "payment",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompletePayment(uuidv7.New())
			return &StepResult{Success: true, Data: "payment_id_123"}, nil
		},
	}
	step2 := &mockStep{
		name: "inventory",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			items := []ReservedItem{
				{ProductID: uuidv7.New(), ReservationID: uuidv7.New(), Quantity: 2},
				{ProductID: uuidv7.New(), ReservationID: uuidv7.New(), Quantity: 1},
			}
			saga.CompleteInventory(items)
			return &StepResult{Success: true}, nil
		},
	}
	step3 := &mockStep{
		name: "shipping",
		executeFunc: func(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error) {
			saga.CompleteShipping(uuidv7.New(), "TRACK456")
			return &StepResult{Success: true}, nil
		},
	}

	orch.AddStep(step1)
	orch.AddStep(step2)
	orch.AddStep(step3)

	saga := NewFulfillmentSaga(uuidv7.New(), uuidv7.New())

	err := orch.Execute(context.Background(), saga)

	require.NoError(t, err)
	assert.Equal(t, FulfillmentSagaStateCompleted, saga.State)
	assert.Len(t, saga.GetCompletedSteps(), 3)
	assert.Len(t, saga.GetReservedItems(), 2)
	assert.NotNil(t, saga.PaymentID)
	assert.NotNil(t, saga.ShipmentID)
	assert.Equal(t, "TRACK456", *saga.TrackingNumber)
	assert.NotNil(t, saga.CompletedAt)
	assert.Equal(t, 3, repo.updateCalls)
}
