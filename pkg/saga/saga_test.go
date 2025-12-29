package saga_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/basilex/promenade/pkg/saga"
)

// TestNew verifies saga creation.
func TestNew(t *testing.T) {
	t.Run("creates saga with name", func(t *testing.T) {
		s := saga.New("test-saga")
		if s.Name() != "test-saga" {
			t.Errorf("expected name 'test-saga', got %s", s.Name())
		}
	})

	t.Run("initializes empty steps", func(t *testing.T) {
		s := saga.New("test")
		if s.StepCount() != 0 {
			t.Errorf("expected 0 steps, got %d", s.StepCount())
		}
	})
}

// TestSaga_AddStep verifies step addition.
func TestSaga_AddStep(t *testing.T) {
	t.Run("adds single step", func(t *testing.T) {
		s := saga.New("test")
		step := saga.Step{
			Name:    "step1",
			Execute: func(ctx context.Context) error { return nil },
		}
		s.AddStep(step)

		if s.StepCount() != 1 {
			t.Errorf("expected 1 step, got %d", s.StepCount())
		}
	})

	t.Run("adds multiple steps", func(t *testing.T) {
		s := saga.New("test")
		s.AddStep(saga.Step{Name: "step1", Execute: func(ctx context.Context) error { return nil }})
		s.AddStep(saga.Step{Name: "step2", Execute: func(ctx context.Context) error { return nil }})

		if s.StepCount() != 2 {
			t.Errorf("expected 2 steps, got %d", s.StepCount())
		}
	})

	t.Run("maintains step order", func(t *testing.T) {
		s := saga.New("test")
		s.AddStep(saga.Step{Name: "first", Execute: func(ctx context.Context) error { return nil }})
		s.AddStep(saga.Step{Name: "second", Execute: func(ctx context.Context) error { return nil }})

		steps := s.Steps()
		if steps[0].Name != "first" || steps[1].Name != "second" {
			t.Error("steps not in correct order")
		}
	})

	t.Run("supports method chaining", func(t *testing.T) {
		s := saga.New("test").
			AddStep(saga.Step{Name: "step1", Execute: func(ctx context.Context) error { return nil }}).
			AddStep(saga.Step{Name: "step2", Execute: func(ctx context.Context) error { return nil }})

		if s.StepCount() != 2 {
			t.Errorf("expected 2 steps, got %d", s.StepCount())
		}
	})
}

// TestSaga_Execute_Success verifies successful execution.
func TestSaga_Execute_Success(t *testing.T) {
	t.Run("executes all steps when all succeed", func(t *testing.T) {
		executed := []string{}
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:    "step1",
			Execute: func(ctx context.Context) error { executed = append(executed, "step1"); return nil },
		})
		s.AddStep(saga.Step{
			Name:    "step2",
			Execute: func(ctx context.Context) error { executed = append(executed, "step2"); return nil },
		})

		err := _ = s.Execute(context.Background())
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(executed) != 2 {
			t.Errorf("expected 2 executions, got %d", len(executed))
		}
	})

	t.Run("executes empty saga without error", func(t *testing.T) {
		s := saga.New("empty")
		err := _ = s.Execute(context.Background())
		if err != nil {
			t.Errorf("expected no error for empty saga, got %v", err)
		}
	})

	t.Run("executes steps in order", func(t *testing.T) {
		order := []int{}
		s := saga.New("test")
		s.AddStep(saga.Step{Name: "1", Execute: func(ctx context.Context) error { order = append(order, 1); return nil }})
		s.AddStep(saga.Step{Name: "2", Execute: func(ctx context.Context) error { order = append(order, 2); return nil }})
		s.AddStep(saga.Step{Name: "3", Execute: func(ctx context.Context) error { order = append(order, 3); return nil }})

		_ = s.Execute(context.Background())
		if len(order) != 3 || order[0] != 1 || order[1] != 2 || order[2] != 3 {
			t.Errorf("steps not executed in order: %v", order)
		}
	})
}

// TestSaga_Execute_Failure verifies error handling.
func TestSaga_Execute_Failure(t *testing.T) {
	t.Run("stops execution on first error", func(t *testing.T) {
		executed := []string{}
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:    "step1",
			Execute: func(ctx context.Context) error { executed = append(executed, "step1"); return nil },
		})
		s.AddStep(saga.Step{
			Name:    "step2",
			Execute: func(ctx context.Context) error { return errors.New("step2 failed") },
		})
		s.AddStep(saga.Step{
			Name:    "step3",
			Execute: func(ctx context.Context) error { executed = append(executed, "step3"); return nil },
		})

		err := _ = s.Execute(context.Background())
		if err == nil {
			t.Error("expected error, got nil")
		}
		if len(executed) != 1 {
			t.Errorf("expected only 1 execution, got %d", len(executed))
		}
	})

	t.Run("includes saga name in error", func(t *testing.T) {
		s := saga.New("order-saga")
		s.AddStep(saga.Step{
			Name:    "validate",
			Execute: func(ctx context.Context) error { return errors.New("validation failed") },
		})

		err := _ = s.Execute(context.Background())
		errMsg := err.Error()
		if errMsg == "" || !contains(errMsg, "order-saga") {
			t.Errorf("error should include saga name, got: %s", errMsg)
		}
	})

	t.Run("includes step name in error", func(t *testing.T) {
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:    "payment-processing",
			Execute: func(ctx context.Context) error { return errors.New("payment failed") },
		})

		err := _ = s.Execute(context.Background())
		errMsg := err.Error()
		if !contains(errMsg, "payment-processing") {
			t.Errorf("error should include step name, got: %s", errMsg)
		}
	})
}

// TestSaga_Compensation verifies compensation logic.
func TestSaga_Compensation(t *testing.T) {
	t.Run("compensates in reverse order", func(t *testing.T) {
		compensated := []string{}
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensated = append(compensated, "step1"); return nil },
		})
		s.AddStep(saga.Step{
			Name:       "step2",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensated = append(compensated, "step2"); return nil },
		})
		s.AddStep(saga.Step{
			Name:    "step3",
			Execute: func(ctx context.Context) error { return errors.New("failed") },
		})

		_ = s.Execute(context.Background())
		if len(compensated) != 2 {
			t.Errorf("expected 2 compensations, got %d", len(compensated))
		}
		if compensated[0] != "step2" || compensated[1] != "step1" {
			t.Errorf("compensation not in reverse order: %v", compensated)
		}
	})

	t.Run("skips nil compensation functions", func(t *testing.T) {
		compensated := []string{}
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensated = append(compensated, "step1"); return nil },
		})
		s.AddStep(saga.Step{
			Name:    "step2",
			Execute: func(ctx context.Context) error { return nil },
		})
		s.AddStep(saga.Step{
			Name:    "step3",
			Execute: func(ctx context.Context) error { return errors.New("failed") },
		})

		_ = s.Execute(context.Background())
		if len(compensated) != 1 || compensated[0] != "step1" {
			t.Errorf("expected only step1 compensation, got: %v", compensated)
		}
	})

	t.Run("reports compensation failure", func(t *testing.T) {
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { return errors.New("compensation failed") },
		})
		s.AddStep(saga.Step{
			Name:    "step2",
			Execute: func(ctx context.Context) error { return errors.New("execute failed") },
		})

		err := _ = s.Execute(context.Background())
		errMsg := err.Error()
		if !contains(errMsg, "compensation") || !contains(errMsg, "failed") {
			t.Errorf("expected compensation failure in error, got: %s", errMsg)
		}
	})

	t.Run("compensates only executed steps", func(t *testing.T) {
		compensated := []string{}
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensated = append(compensated, "step1"); return nil },
		})
		s.AddStep(saga.Step{
			Name:    "step2",
			Execute: func(ctx context.Context) error { return errors.New("failed") },
		})
		s.AddStep(saga.Step{
			Name:       "step3",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensated = append(compensated, "step3"); return nil },
		})

		_ = s.Execute(context.Background())
		if len(compensated) != 1 || compensated[0] != "step1" {
			t.Errorf("should only compensate step1, got: %v", compensated)
		}
	})

	t.Run("does not compensate when all steps succeed", func(t *testing.T) {
		compensated := false
		s := saga.New("test")
		s.AddStep(saga.Step{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensated = true; return nil },
		})

		_ = s.Execute(context.Background())
		if compensated {
			t.Error("should not compensate when all steps succeed")
		}
	})
}

// TestBuilder verifies the builder pattern.
func TestBuilder(t *testing.T) {
	t.Run("builds saga with steps", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, nil).
			Step("step2", func(ctx context.Context) error { return nil }, nil).
			Build()

		if s.StepCount() != 2 {
			t.Errorf("expected 2 steps, got %d", s.StepCount())
		}
	})

	t.Run("builds saga with and without compensation", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil }).
			StepWithoutCompensation("step2", func(ctx context.Context) error { return nil }).
			Build()

		steps := s.Steps()
		if steps[0].Compensate == nil {
			t.Error("step1 should have compensation")
		}
		if steps[1].Compensate != nil {
			t.Error("step2 should not have compensation")
		}
	})

	t.Run("supports method chaining", func(t *testing.T) {
		builder := saga.NewBuilder("test")
		result := builder.Step("step1", func(ctx context.Context) error { return nil }, nil)
		if result != builder {
			t.Error("Step should return builder for chaining")
		}
	})

	t.Run("creates named saga", func(t *testing.T) {
		s := saga.NewBuilder("order-saga").Build()
		if s.Name() != "order-saga" {
			t.Errorf("expected name 'order-saga', got %s", s.Name())
		}
	})
}

// TestSaga_ExecuteWithResult verifies detailed result tracking.
func TestSaga_ExecuteWithResult(t *testing.T) {
	t.Run("returns success result", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, nil).
			Step("step2", func(ctx context.Context) error { return nil }, nil).
			Build()

		result := s.ExecuteWithResult(context.Background())
		if !result.Success {
			t.Error("expected success")
		}
		if result.CompletedSteps != 2 {
			t.Errorf("expected 2 completed steps, got %d", result.CompletedSteps)
		}
	})

	t.Run("returns failure result", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, nil).
			Step("step2", func(ctx context.Context) error { return errors.New("failed") }, nil).
			Build()

		result := s.ExecuteWithResult(context.Background())
		if result.Success {
			t.Error("expected failure")
		}
		if result.FailedStep != "step2" {
			t.Errorf("expected failed step 'step2', got %s", result.FailedStep)
		}
	})

	t.Run("tracks compensation count", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil }).
			Step("step2", func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil }).
			Step("step3", func(ctx context.Context) error { return errors.New("failed") }, nil).
			Build()

		result := s.ExecuteWithResult(context.Background())
		if result.CompensatedSteps != 2 {
			t.Errorf("expected 2 compensated steps, got %d", result.CompensatedSteps)
		}
	})

	t.Run("tracks completed steps before failure", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, nil).
			Step("step2", func(ctx context.Context) error { return nil }, nil).
			Step("step3", func(ctx context.Context) error { return errors.New("failed") }, nil).
			Build()

		result := s.ExecuteWithResult(context.Background())
		if result.CompletedSteps != 2 {
			t.Errorf("expected 2 completed steps before failure, got %d", result.CompletedSteps)
		}
	})
}

// TestSaga_RealWorldScenarios tests realistic use cases.
func TestSaga_RealWorldScenarios(t *testing.T) {
	t.Run("order creation saga", func(t *testing.T) {
		customerValidated := false
		inventoryReserved := false
		paymentProcessed := false
		orderCreated := false

		s := saga.NewBuilder("create-order").
			Step("validate-customer",
				func(ctx context.Context) error { customerValidated = true; return nil },
				func(ctx context.Context) error { customerValidated = false; return nil }).
			Step("reserve-inventory",
				func(ctx context.Context) error { inventoryReserved = true; return nil },
				func(ctx context.Context) error { inventoryReserved = false; return nil }).
			Step("process-payment",
				func(ctx context.Context) error { return errors.New("payment declined") },
				func(ctx context.Context) error { paymentProcessed = false; return nil }).
			Step("create-order",
				func(ctx context.Context) error { orderCreated = true; return nil },
				nil).
			Build()

		_ = s.Execute(context.Background())

		if customerValidated {
			t.Error("customer validation should be compensated")
		}
		if inventoryReserved {
			t.Error("inventory reservation should be compensated")
		}
		if paymentProcessed {
			t.Error("payment should not be processed")
		}
		if orderCreated {
			t.Error("order should not be created")
		}
	})

	t.Run("multi-step workflow", func(t *testing.T) {
		steps := []string{}
		s := saga.NewBuilder("workflow").
			Step("initialize", func(ctx context.Context) error { steps = append(steps, "init"); return nil }, nil).
			Step("process", func(ctx context.Context) error { steps = append(steps, "process"); return nil }, nil).
			Step("finalize", func(ctx context.Context) error { steps = append(steps, "finalize"); return nil }, nil).
			Build()

		err := _ = s.Execute(context.Background())
		if err != nil {
			t.Errorf("workflow should succeed, got error: %v", err)
		}
		if len(steps) != 3 {
			t.Errorf("expected 3 steps executed, got %d", len(steps))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return nil
				}
			}, nil).
			Build()

		err := s.Execute(ctx)
		if err == nil {
			t.Error("expected context cancellation error")
		}
	})
}

// TestSaga_EdgeCases tests boundary conditions.
func TestSaga_EdgeCases(t *testing.T) {
	t.Run("single step saga", func(t *testing.T) {
		executed := false
		s := saga.NewBuilder("single").
			Step("only-step", func(ctx context.Context) error { executed = true; return nil }, nil).
			Build()

		err := _ = s.Execute(context.Background())
		if err != nil || !executed {
			t.Error("single step saga should execute successfully")
		}
	})

	t.Run("saga with all steps having compensation", func(t *testing.T) {
		compensated := 0
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, func(ctx context.Context) error { compensated++; return nil }).
			Step("step2", func(ctx context.Context) error { return nil }, func(ctx context.Context) error { compensated++; return nil }).
			Step("step3", func(ctx context.Context) error { return errors.New("failed") }, func(ctx context.Context) error { return nil }).
			Build()

		_ = s.Execute(context.Background())
		if compensated != 2 {
			t.Errorf("expected 2 compensations, got %d", compensated)
		}
	})

	t.Run("saga with no compensation functions", func(t *testing.T) {
		s := saga.NewBuilder("test").
			StepWithoutCompensation("step1", func(ctx context.Context) error { return nil }).
			StepWithoutCompensation("step2", func(ctx context.Context) error { return errors.New("failed") }).
			Build()

		err := _ = s.Execute(context.Background())
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("first step fails", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return errors.New("failed") }, nil).
			Build()

		result := s.ExecuteWithResult(context.Background())
		if result.CompletedSteps != 0 {
			t.Errorf("expected 0 completed steps, got %d", result.CompletedSteps)
		}
	})

	t.Run("last step fails", func(t *testing.T) {
		s := saga.NewBuilder("test").
			Step("step1", func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil }).
			Step("step2", func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil }).
			Step("step3", func(ctx context.Context) error { return errors.New("failed") }, nil).
			Build()

		result := s.ExecuteWithResult(context.Background())
		if result.CompensatedSteps != 2 {
			t.Errorf("expected 2 compensations, got %d", result.CompensatedSteps)
		}
	})

	t.Run("empty saga name", func(t *testing.T) {
		s := saga.New("")
		if s.Name() != "" {
			t.Error("empty name should be preserved")
		}
	})
}

// Helper function to check if string contains substring.
func contains(s, substr string) bool {
	return fmt.Sprintf("%s", s) != "" && fmt.Sprintf("%s", substr) != "" &&
		len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
