package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowSchema_Validate_Success(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "processing", Type: "task"},
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "processing", Event: "start"},
			{From: "processing", To: "completed", Event: "complete"},
		},
	}

	err := schema.Validate()

	assert.NoError(t, err)
}

func TestWorkflowSchema_Validate_NoStates(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States:       []WorkflowState{},
		Transitions:  []WorkflowTransition{},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one state is required")
}

func TestWorkflowSchema_Validate_NoInitialState(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
		},
		Transitions: []WorkflowTransition{},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "initial state is required")
}

func TestWorkflowSchema_Validate_InvalidInitialState(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "non_existent",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "completed", Event: "complete"},
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "initial state 'non_existent' not found in states")
}

func TestWorkflowSchema_Validate_TransitionToInvalidState(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "non_existent", Event: "process"},
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'to' state 'non_existent' not found in states")
}

func TestWorkflowSchema_Validate_NoTerminalState(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "state1",
		States: []WorkflowState{
			{Name: "state1", Type: "task"},
			{Name: "state2", Type: "task"},
		},
		Transitions: []WorkflowTransition{
			{From: "state1", To: "state2", Event: "next"},
			{From: "state2", To: "state1", Event: "back"}, // Infinite loop
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one terminal state is required")
}

func TestWorkflowSchema_Validate_UnreachableState(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "completed", Type: "end", IsFinal: true},
			{Name: "orphaned", Type: "task"}, // This state is unreachable
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "completed", Event: "complete"},
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state 'orphaned' is unreachable")
}

func TestWorkflowSchema_Validate_DuplicateStateName(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "pending", Type: "task"}, // Duplicate
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "completed", Event: "complete"},
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate state name: pending")
}

func TestWorkflowSchema_Validate_EmptyStateName(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "", Type: "task"}, // Empty name
		},
		Transitions: []WorkflowTransition{},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state name cannot be empty")
}

func TestWorkflowSchema_Validate_EmptyTransitionFrom(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "", To: "completed", Event: "complete"},
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'from' state cannot be empty")
}

func TestWorkflowSchema_Validate_EmptyTransitionTo(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "", Event: "complete"},
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'to' state cannot be empty")
}

func TestWorkflowSchema_Validate_EmptyEventName(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "pending",
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "completed", Event: ""},
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "event name is required")
}

func TestWorkflowSchema_Validate_ComplexWorkflow(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "start",
		States: []WorkflowState{
			{Name: "start", Type: "start"},
			{Name: "review", Type: "task"},
			{Name: "approved", Type: "gateway"},
			{Name: "rejected", Type: "gateway"},
			{Name: "processing", Type: "task"},
			{Name: "completed", Type: "end", IsFinal: true},
			{Name: "cancelled", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "start", To: "review", Event: "submit"},
			{From: "review", To: "approved", Event: "approve"},
			{From: "review", To: "rejected", Event: "reject"},
			{From: "approved", To: "processing", Event: "process"},
			{From: "processing", To: "completed", Event: "complete"},
			{From: "rejected", To: "cancelled", Event: "cancel"},
		},
	}

	err := schema.Validate()

	assert.NoError(t, err)
}

func TestWorkflowSchema_DetectCycles_NoCycles(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "start",
		States: []WorkflowState{
			{Name: "start", Type: "start"},
			{Name: "middle", Type: "task"},
			{Name: "end", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "start", To: "middle", Event: "next"},
			{From: "middle", To: "end", Event: "complete"},
		},
	}

	cycles := schema.detectCycles()

	assert.Empty(t, cycles, "Expected no cycles in linear workflow")
}

func TestWorkflowSchema_DetectCycles_SimpleCycle(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "state1",
		States: []WorkflowState{
			{Name: "state1", Type: "task"},
			{Name: "state2", Type: "task"},
			{Name: "end", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "state1", To: "state2", Event: "next"},
			{From: "state2", To: "state1", Event: "back"}, // Cycle
			{From: "state2", To: "end", Event: "complete"}, // Exit from cycle
		},
	}

	cycles := schema.detectCycles()

	assert.NotEmpty(t, cycles, "Expected to detect cycle")
	assert.Len(t, cycles, 1, "Expected exactly one cycle")
}

func TestWorkflowSchema_Validate_CycleWithExit(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "start",
		States: []WorkflowState{
			{Name: "start", Type: "start"},
			{Name: "processing", Type: "task"},
			{Name: "retry", Type: "task"},
			{Name: "completed", Type: "end", IsFinal: true},
		},
		Transitions: []WorkflowTransition{
			{From: "start", To: "processing", Event: "begin"},
			{From: "processing", To: "retry", Event: "error"},
			{From: "retry", To: "processing", Event: "retry"}, // Cycle
			{From: "processing", To: "completed", Event: "success"}, // Exit
		},
	}

	err := schema.Validate()

	assert.NoError(t, err, "Cycle with exit path should be valid")
}

func TestWorkflowSchema_Validate_DeadlockCycle(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "state1",
		States: []WorkflowState{
			{Name: "state1", Type: "task"},
			{Name: "state2", Type: "task"},
			{Name: "state3", Type: "task"},
			{Name: "unreachable_end", Type: "end", IsFinal: true}, // Terminal exists but unreachable from cycle
		},
		Transitions: []WorkflowTransition{
			{From: "state1", To: "state2", Event: "next"},
			{From: "state2", To: "state3", Event: "next"},
			{From: "state3", To: "state1", Event: "back"}, // Infinite loop - no exit!
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	// Should fail on unreachable state check first
	assert.Contains(t, err.Error(), "unreachable")
}

func TestWorkflowSchema_Validate_DeadlockInMiddle(t *testing.T) {
	schema := WorkflowSchema{
		InitialState: "start",
		States: []WorkflowState{
			{Name: "start", Type: "start"},
			{Name: "loop1", Type: "task"},
			{Name: "loop2", Type: "task"},
		},
		Transitions: []WorkflowTransition{
			{From: "start", To: "loop1", Event: "begin"},
			{From: "loop1", To: "loop2", Event: "next"},
			{From: "loop2", To: "loop1", Event: "back"}, // Deadlock - no way out of this loop!
		},
	}

	err := schema.Validate()

	assert.Error(t, err)
	// Either deadlock or no terminal state - both are valid errors for this scenario
	assert.True(t, 
		err.Error() == "at least one terminal state is required (state with no outgoing transitions or marked as final)" ||
		err.Error() == "deadlock detected: state 'loop1' cannot reach any terminal state (infinite loop)" ||
		err.Error() == "deadlock detected: state 'loop2' cannot reach any terminal state (infinite loop)",
		"Expected deadlock or no terminal state error, got: %s", err.Error())
}
