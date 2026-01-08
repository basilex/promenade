package saga

import "context"

// StepResult represents the result of a saga step execution
type StepResult struct {
	Success bool
	Data    interface{}
	Error   error
}

// SagaStep defines the interface for saga steps
// Each step must implement Execute and Compensate logic
type SagaStep interface {
	// Execute performs the forward action of the step
	Execute(ctx context.Context, saga *FulfillmentSaga) (*StepResult, error)

	// Compensate performs the rollback/compensation action
	Compensate(ctx context.Context, saga *FulfillmentSaga) error

	// Name returns the step name for logging and tracking
	Name() string
}
