package entity_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/workflows/domain/entity"
)

// ============================================================================
// Stress Tests for Cycle Detection Performance
// ============================================================================
// These tests verify that cycle detection algorithms scale properly with
// graph size and don't have performance issues with large workflows.

// generateLinearWorkflow creates a linear workflow (no cycles)
// States: state_0 -> state_1 -> ... -> state_N -> final
func generateLinearWorkflow(numStates int) entity.WorkflowSchema {
	states := make([]entity.WorkflowState, numStates+1) // +1 for final state
	transitions := make([]entity.WorkflowTransition, numStates)

	// Generate states
	for i := 0; i < numStates; i++ {
		states[i] = entity.WorkflowState{
			Name:    fmt.Sprintf("state_%d", i),
			Type:    "activity",
			IsFinal: false,
		}
	}
	states[numStates] = entity.WorkflowState{
		Name:    "final",
		Type:    "final",
		IsFinal: true,
	}

	// Generate linear transitions
	for i := 0; i < numStates; i++ {
		if i == numStates-1 {
			transitions[i] = entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    "final",
				Event: "finish",
			}
		} else {
			transitions[i] = entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    fmt.Sprintf("state_%d", i+1),
				Event: "next",
			}
		}
	}

	return entity.WorkflowSchema{
		InitialState: "state_0",
		States:       states,
		Transitions:  transitions,
	}
}

// generateComplexWorkflow creates a complex workflow with multiple paths
// Each state has 2-3 outgoing transitions, creating a dense graph
func generateComplexWorkflow(numStates int) entity.WorkflowSchema {
	states := make([]entity.WorkflowState, numStates+1) // +1 for final state
	transitions := make([]entity.WorkflowTransition, 0, numStates*3)

	// Generate states
	for i := 0; i < numStates; i++ {
		states[i] = entity.WorkflowState{
			Name:    fmt.Sprintf("state_%d", i),
			Type:    "activity",
			IsFinal: false,
		}
	}
	states[numStates] = entity.WorkflowState{
		Name:    "final",
		Type:    "final",
		IsFinal: true,
	}

	// Generate complex transitions (multiple paths, no cycles)
	for i := 0; i < numStates; i++ {
		// Main path forward
		if i < numStates-1 {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    fmt.Sprintf("state_%d", i+1),
				Event: "next",
			})
		} else {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    "final",
				Event: "finish",
			})
		}

		// Alternative path (skip one state)
		if i < numStates-2 {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    fmt.Sprintf("state_%d", i+2),
				Event: "skip",
			})
		}

		// Direct path to final (from middle states)
		if i > numStates/3 && i < numStates*2/3 {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    "final",
				Event: "abort",
			})
		}
	}

	return entity.WorkflowSchema{
		InitialState: "state_0",
		States:       states,
		Transitions:  transitions,
	}
}

// generateCyclicWorkflow creates a workflow with valid retry loops
// Each loop has exit to final state (valid pattern)
func generateCyclicWorkflow(numStates int, loopSize int) entity.WorkflowSchema {
	states := make([]entity.WorkflowState, numStates+1)
	transitions := make([]entity.WorkflowTransition, 0, numStates*2)

	// Generate states
	for i := 0; i < numStates; i++ {
		states[i] = entity.WorkflowState{
			Name:    fmt.Sprintf("state_%d", i),
			Type:    "activity",
			IsFinal: false,
		}
	}
	states[numStates] = entity.WorkflowState{
		Name:    "final",
		Type:    "final",
		IsFinal: true,
	}

	// Generate transitions with valid cycles
	for i := 0; i < numStates; i++ {
		// Forward path
		if i < numStates-1 {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    fmt.Sprintf("state_%d", i+1),
				Event: "next",
			})
		} else {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    "final",
				Event: "finish",
			})
		}

		// Retry loop (cycle back N states)
		if i >= loopSize {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    fmt.Sprintf("state_%d", i-loopSize),
				Event: "retry",
			})
		}

		// Exit from loop to final (every 5th state)
		if i%5 == 0 && i > 0 {
			transitions = append(transitions, entity.WorkflowTransition{
				From:  fmt.Sprintf("state_%d", i),
				To:    "final",
				Event: "exit",
			})
		}
	}

	return entity.WorkflowSchema{
		InitialState: "state_0",
		States:       states,
		Transitions:  transitions,
	}
}

// measureMemory returns current memory allocation in MB
func measureMemory() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

// ============================================================================
// Small Graph Tests (Baseline)
// ============================================================================

func TestStress_SmallGraph_LinearWorkflow(t *testing.T) {
	const numStates = 10

	schema := generateLinearWorkflow(numStates)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	assert.Equal(t, numStates+1, len(schema.States))
	assert.Equal(t, numStates, len(schema.Transitions))

	t.Logf("Small Linear Graph: %d states, %d transitions", numStates, len(schema.Transitions))
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	// Performance assertions (baseline)
	assert.Less(t, duration.Milliseconds(), int64(10), "Should validate in < 10ms")
}

func TestStress_SmallGraph_ComplexWorkflow(t *testing.T) {
	const numStates = 10

	schema := generateComplexWorkflow(numStates)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	t.Logf("Small Complex Graph: %d states, %d transitions", numStates, len(schema.Transitions))
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	assert.Less(t, duration.Milliseconds(), int64(10), "Should validate in < 10ms")
}

// ============================================================================
// Medium Graph Tests
// ============================================================================

func TestStress_MediumGraph_LinearWorkflow(t *testing.T) {
	const numStates = 50

	schema := generateLinearWorkflow(numStates)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	assert.Equal(t, numStates+1, len(schema.States))
	assert.Equal(t, numStates, len(schema.Transitions))

	t.Logf("Medium Linear Graph: %d states, %d transitions", numStates, len(schema.Transitions))
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	// Should still be fast
	assert.Less(t, duration.Milliseconds(), int64(50), "Should validate in < 50ms")
}

func TestStress_MediumGraph_ComplexWorkflow(t *testing.T) {
	const numStates = 50

	schema := generateComplexWorkflow(numStates)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	t.Logf("Medium Complex Graph: %d states, %d transitions", numStates, len(schema.Transitions))
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	assert.Less(t, duration.Milliseconds(), int64(100), "Should validate in < 100ms")
}

func TestStress_MediumGraph_WithRetryLoops(t *testing.T) {
	const numStates = 50
	const loopSize = 3

	schema := generateCyclicWorkflow(numStates, loopSize)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	t.Logf("Medium Cyclic Graph: %d states, %d transitions (loop size: %d)", 
		numStates, len(schema.Transitions), loopSize)
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	assert.Less(t, duration.Milliseconds(), int64(150), "Should validate in < 150ms")
}

// ============================================================================
// Large Graph Tests (Performance Critical)
// ============================================================================

func TestStress_LargeGraph_LinearWorkflow(t *testing.T) {
	const numStates = 200

	schema := generateLinearWorkflow(numStates)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	assert.Equal(t, numStates+1, len(schema.States))
	assert.Equal(t, numStates, len(schema.Transitions))

	t.Logf("Large Linear Graph: %d states, %d transitions", numStates, len(schema.Transitions))
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	// Linear graph should still be reasonably fast
	assert.Less(t, duration.Milliseconds(), int64(200), "Should validate in < 200ms")
}

func TestStress_LargeGraph_ComplexWorkflow(t *testing.T) {
	const numStates = 200

	schema := generateComplexWorkflow(numStates)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	t.Logf("Large Complex Graph: %d states, %d transitions", numStates, len(schema.Transitions))
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	// Complex graph with multiple paths - most expensive test
	assert.Less(t, duration.Milliseconds(), int64(500), "Should validate in < 500ms")
	assert.Less(t, memUsed, 10.0, "Should use < 10MB memory")
}

func TestStress_LargeGraph_WithRetryLoops(t *testing.T) {
	const numStates = 200
	const loopSize = 5

	schema := generateCyclicWorkflow(numStates, loopSize)

	memBefore := measureMemory()
	start := time.Now()

	err := schema.Validate()

	duration := time.Since(start)
	memAfter := measureMemory()
	memUsed := memAfter - memBefore

	require.NoError(t, err)
	t.Logf("Large Cyclic Graph: %d states, %d transitions (loop size: %d)", 
		numStates, len(schema.Transitions), loopSize)
	t.Logf("  Validation time: %v", duration)
	t.Logf("  Memory used: %.2f MB", memUsed)

	// Cyclic graphs are most expensive due to cycle detection
	assert.Less(t, duration.Milliseconds(), int64(600), "Should validate in < 600ms")
	assert.Less(t, memUsed, 15.0, "Should use < 15MB memory")
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkValidate_Small_Linear(b *testing.B) {
	schema := generateLinearWorkflow(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate()
	}
}

func BenchmarkValidate_Medium_Linear(b *testing.B) {
	schema := generateLinearWorkflow(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate()
	}
}

func BenchmarkValidate_Large_Linear(b *testing.B) {
	schema := generateLinearWorkflow(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate()
	}
}

func BenchmarkValidate_Medium_Complex(b *testing.B) {
	schema := generateComplexWorkflow(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate()
	}
}

func BenchmarkValidate_Large_Complex(b *testing.B) {
	schema := generateComplexWorkflow(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate()
	}
}

func BenchmarkValidate_Large_Cyclic(b *testing.B) {
	schema := generateCyclicWorkflow(200, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.Validate()
	}
}
