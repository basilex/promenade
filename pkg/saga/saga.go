package saga

import (
	"context"
	"fmt"
)

// StepFunc is a function that executes a saga step.
type StepFunc func(ctx context.Context) error

// CompensateFunc is a function that compensates (rolls back) a saga step.
type CompensateFunc func(ctx context.Context) error

// Step represents a single step in a saga.
type Step struct {
	Name       string
	Execute    StepFunc
	Compensate CompensateFunc
}

// Saga represents a sequence of steps with compensation logic.
type Saga struct {
	name  string
	steps []Step
}

// New creates a new Saga.
func New(name string) *Saga {
	return &Saga{
		name:  name,
		steps: make([]Step, 0),
	}
}

// AddStep adds a step to the saga.
func (s *Saga) AddStep(step Step) *Saga {
	s.steps = append(s.steps, step)
	return s
}

// Name returns the saga name.
func (s *Saga) Name() string {
	return s.name
}

// Steps returns all steps in the saga.
func (s *Saga) Steps() []Step {
	return s.steps
}

// StepCount returns the number of steps.
func (s *Saga) StepCount() int {
	return len(s.steps)
}

// Execute runs all steps in the saga.
func (s *Saga) Execute(ctx context.Context) error {
	executedSteps := make([]Step, 0, len(s.steps))

	for i, step := range s.steps {
		if err := step.Execute(ctx); err != nil {
			compensateErr := s.compensate(ctx, executedSteps)
			if compensateErr != nil {
				return fmt.Errorf("saga %s: step %d (%s) failed: %w, compensation also failed: %v",
					s.name, i+1, step.Name, err, compensateErr)
			}
			return fmt.Errorf("saga %s: step %d (%s) failed: %w",
				s.name, i+1, step.Name, err)
		}
		executedSteps = append(executedSteps, step)
	}

	return nil
}

// compensate executes compensation functions in reverse order.
func (s *Saga) compensate(ctx context.Context, steps []Step) error {
	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		if step.Compensate == nil {
			continue
		}

		if err := step.Compensate(ctx); err != nil {
			return fmt.Errorf("compensation failed for step %s: %w", step.Name, err)
		}
	}
	return nil
}

// Builder provides a fluent interface for building sagas.
type Builder struct {
	saga *Saga
}

// NewBuilder creates a new saga builder.
func NewBuilder(name string) *Builder {
	return &Builder{
		saga: New(name),
	}
}

// Step adds a step with both execute and compensate functions.
func (b *Builder) Step(name string, execute StepFunc, compensate CompensateFunc) *Builder {
	b.saga.AddStep(Step{
		Name:       name,
		Execute:    execute,
		Compensate: compensate,
	})
	return b
}

// StepWithoutCompensation adds a step without compensation function.
func (b *Builder) StepWithoutCompensation(name string, execute StepFunc) *Builder {
	b.saga.AddStep(Step{
		Name:    name,
		Execute: execute,
	})
	return b
}

// Build returns the constructed saga.
func (b *Builder) Build() *Saga {
	return b.saga
}

// ExecutionResult represents the result of saga execution.
type ExecutionResult struct {
	Success          bool
	FailedStep       string
	CompletedSteps   int
	CompensatedSteps int
	Error            error
}

// ExecuteWithResult runs the saga and returns detailed execution result.
func (s *Saga) ExecuteWithResult(ctx context.Context) ExecutionResult {
	result := ExecutionResult{
		Success: true,
	}

	executedSteps := make([]Step, 0, len(s.steps))

	for i, step := range s.steps {
		if err := step.Execute(ctx); err != nil {
			result.Success = false
			result.FailedStep = step.Name
			result.CompletedSteps = i
			result.Error = err

			compensateErr := s.compensate(ctx, executedSteps)
			if compensateErr != nil {
				result.Error = fmt.Errorf("step failed: %w, compensation failed: %v", err, compensateErr)
			}
			result.CompensatedSteps = len(executedSteps)

			return result
		}
		executedSteps = append(executedSteps, step)
	}

	result.CompletedSteps = len(s.steps)
	return result
}
