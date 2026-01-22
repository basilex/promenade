package saga

import (
	"context"
	"fmt"
	"log/slog"
)

// Orchestrator coordinates the execution of saga steps
type Orchestrator struct {
	steps      []SagaStep
	repository ISagaRepository
}

// NewOrchestrator creates a new saga orchestrator
func NewOrchestrator(repo ISagaRepository) *Orchestrator {
	return &Orchestrator{
		steps:      []SagaStep{},
		repository: repo,
	}
}

// AddStep adds a step to the orchestrator
func (o *Orchestrator) AddStep(step SagaStep) {
	o.steps = append(o.steps, step)
}

// Execute runs the saga steps sequentially with automatic compensation on failure
func (o *Orchestrator) Execute(ctx context.Context, saga *FulfillmentSaga) error {
	// Execute steps sequentially starting from current step
	for i := saga.CurrentStep; i < len(o.steps); i++ {
		step := o.steps[i]

		// Execute step
		result, err := step.Execute(ctx, saga)
		if err != nil || (result != nil && !result.Success) {
			// Step failed - start compensation
			reason := "unknown error"
			originalErr := err
			if err != nil {
				reason = err.Error()
			} else if result != nil && result.Error != nil {
				reason = result.Error.Error()
				originalErr = result.Error
			}

			saga.StartCompensation(step.Name(), reason)
			if err := o.repository.Update(ctx, saga); err != nil {
				return fmt.Errorf("failed to update saga: %w", err)
			}

			// Compensate completed steps in reverse order
			if err := o.compensate(ctx, saga, i); err != nil {
				return fmt.Errorf("compensation failed: %w", err)
			}

			// Return original error after successful compensation
			if originalErr != nil {
				return originalErr
			}
			return fmt.Errorf("step %s failed: %s", step.Name(), reason)
		}

		// Save progress after each successful step
		if err := o.repository.Update(ctx, saga); err != nil {
			return fmt.Errorf("failed to update saga: %w", err)
		}
	}

	return nil
}

// compensate executes compensation logic for completed steps in reverse order
func (o *Orchestrator) compensate(ctx context.Context, saga *FulfillmentSaga, failedStepIndex int) error {
	// Compensate in reverse order (LIFO)
	for i := failedStepIndex - 1; i >= 0; i-- {
		step := o.steps[i]
		if err := step.Compensate(ctx, saga); err != nil {
			// Log compensation failure but continue
			// In production, would retry or alert operations team
			slog.Warn("compensation failed for saga step",
				slog.String("step", step.Name()),
				slog.String("saga_id", saga.ID.String()),
				slog.String("error", err.Error()))
			// Continue with other compensations
		}
	}

	// Note: Saga remains in "Compensating" state
	// Caller is responsible for cancelling the saga if needed
	return nil
}

// Resume attempts to resume a failed saga from where it left off
func (o *Orchestrator) Resume(ctx context.Context, saga *FulfillmentSaga) error {
	if !saga.IsInProgress() {
		return fmt.Errorf("saga is not in progress (state: %s)", saga.State)
	}

	// Continue execution from current step
	return o.Execute(ctx, saga)
}

// GetStepCount returns the total number of steps
func (o *Orchestrator) GetStepCount() int {
	return len(o.steps)
}
