package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewWorkflowInstance(t *testing.T) {
	definitionID := uuidv7.New()
	version := 1
	initialState := "pending"
	input := jsonb.Map{"order_id": "12345"}
	startedBy := uuidv7.New()

	instance := NewWorkflowInstance(definitionID, version, initialState, input, startedBy)

	require.NotNil(t, instance)
	assert.NotEqual(t, uuidv7.Nil, instance.ID)
	assert.Equal(t, definitionID, instance.DefinitionID)
	assert.Equal(t, version, instance.DefinitionVersion)
	assert.Equal(t, WorkflowInstanceStatusPending, instance.Status)
	assert.Equal(t, initialState, instance.CurrentState)
	assert.Nil(t, instance.PreviousState)
	assert.Equal(t, input, instance.Input)
	assert.Equal(t, startedBy, instance.StartedBy)
	assert.Equal(t, 5, instance.Priority)
	assert.NotNil(t, instance.CreatedAt)
	assert.NotNil(t, instance.UpdatedAt)
}

func TestWorkflowInstance_Start(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "initial", jsonb.Map{}, uuidv7.New(),
	)

	err := instance.Start()

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusRunning, instance.Status)
	assert.NotNil(t, instance.StartedAt)
}

func TestWorkflowInstance_Start_NotPending(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "initial", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()

	err := instance.Start()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot start workflow in status")
}

func TestWorkflowInstance_TransitionTo(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "state1", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()

	err := instance.TransitionTo("state2", "event_occurred")

	assert.NoError(t, err)
	assert.Equal(t, "state2", instance.CurrentState)
	assert.NotNil(t, instance.PreviousState)
	assert.Equal(t, "state1", *instance.PreviousState)
}

func TestWorkflowInstance_Complete(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "final", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()
	output := jsonb.Map{"result": "success"}

	err := instance.Complete(output)

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusCompleted, instance.Status)
	assert.Equal(t, output, instance.Output)
	assert.NotNil(t, instance.CompletedAt)
}

func TestWorkflowInstance_Fail(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "processing", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()
	errorMsg := "Processing failed"
	errorDetails := jsonb.Map{"code": "ERR_001"}

	err := instance.Fail(errorMsg, errorDetails)

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusFailed, instance.Status)
	assert.NotNil(t, instance.ErrorMessage)
	assert.Equal(t, errorMsg, *instance.ErrorMessage)
	assert.Equal(t, errorDetails, instance.ErrorDetails)
	assert.NotNil(t, instance.CompletedAt)
}

func TestWorkflowInstance_Cancel(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "running", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()

	err := instance.Cancel()

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusCancelled, instance.Status)
	assert.NotNil(t, instance.CompletedAt)
}

func TestWorkflowInstance_Pause(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "running", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()

	err := instance.Pause()

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusPaused, instance.Status)
}

func TestWorkflowInstance_Resume(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "paused", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()
	instance.Pause()

	err := instance.Resume()

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusRunning, instance.Status)
}

func TestWorkflowInstance_WaitForEvent(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "waiting", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()

	err := instance.WaitForEvent()

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusWaiting, instance.Status)
}

func TestWorkflowInstance_Timeout(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "long_task", jsonb.Map{}, uuidv7.New(),
	)
	instance.Start()

	err := instance.Timeout()

	assert.NoError(t, err)
	assert.Equal(t, WorkflowInstanceStatusTimedOut, instance.Status)
	assert.NotNil(t, instance.ErrorMessage)
	assert.Contains(t, *instance.ErrorMessage, "timed out")
	assert.NotNil(t, instance.CompletedAt)
}

func TestWorkflowInstance_IncrementRetry(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "retry_task", jsonb.Map{}, uuidv7.New(),
	)

	assert.Equal(t, 0, instance.RetryCount)
	instance.IncrementRetry()
	assert.Equal(t, 1, instance.RetryCount)
	instance.IncrementRetry()
	assert.Equal(t, 2, instance.RetryCount)
}

func TestWorkflowInstance_UpdateContext(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "state", jsonb.Map{}, uuidv7.New(),
	)
	newContext := jsonb.Map{"step": 2, "data": "updated"}

	instance.UpdateContext(newContext)

	assert.Equal(t, newContext, instance.Context)
}

func TestWorkflowInstance_Assign(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "manual_task", jsonb.Map{}, uuidv7.New(),
	)
	assigneeID := uuidv7.New()

	instance.Assign(assigneeID)

	require.NotNil(t, instance.AssignedTo)
	assert.Equal(t, assigneeID, *instance.AssignedTo)
}

func TestWorkflowInstance_Unassign(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "manual_task", jsonb.Map{}, uuidv7.New(),
	)
	instance.Assign(uuidv7.New())

	instance.Unassign()

	assert.Nil(t, instance.AssignedTo)
}

func TestWorkflowInstance_SetPriority(t *testing.T) {
	instance := NewWorkflowInstance(
		uuidv7.New(), 1, "task", jsonb.Map{}, uuidv7.New(),
	)

	err := instance.SetPriority(8)
	assert.NoError(t, err)
	assert.Equal(t, 8, instance.Priority)

	err = instance.SetPriority(0)
	assert.Error(t, err)

	err = instance.SetPriority(11)
	assert.Error(t, err)
}
