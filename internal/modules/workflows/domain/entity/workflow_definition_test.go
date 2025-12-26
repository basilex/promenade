package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewWorkflowDefinition(t *testing.T) {
	name := "order_fulfillment"
	displayName := "Order Fulfillment"
	description := "Process for fulfilling customer orders"
	schema := WorkflowSchema{
		States: []WorkflowState{
			{Name: "pending", Type: "start"},
			{Name: "processing", Type: "task"},
			{Name: "completed", Type: "end"},
		},
		Transitions: []WorkflowTransition{
			{From: "pending", To: "processing", Event: "start_processing"},
			{From: "processing", To: "completed", Event: "complete"},
		},
		InitialState: "pending",
	}
	createdBy := uuidv7.New()

	def := NewWorkflowDefinition(name, displayName, description, schema, createdBy)

	require.NotNil(t, def)
	assert.NotEqual(t, uuidv7.Nil, def.ID)
	assert.Equal(t, name, def.Name)
	assert.Equal(t, displayName, def.DisplayName)
	assert.Equal(t, description, def.Description)
	assert.Equal(t, 1, def.Version)
	assert.Equal(t, WorkflowDefinitionStatusDraft, def.Status)
	assert.Equal(t, createdBy, def.CreatedBy)
	assert.Len(t, schema.States, 3)
	assert.Len(t, schema.Transitions, 2)
	assert.NotNil(t, def.CreatedAt)
	assert.NotNil(t, def.UpdatedAt)
}

func TestWorkflowDefinition_Activate(t *testing.T) {
	schema := WorkflowSchema{
		States: []WorkflowState{
			{Name: "start", Type: "start"},
			{Name: "end", Type: "end"},
		},
		Transitions:  []WorkflowTransition{{From: "start", To: "end", Event: "complete"}},
		InitialState: "start",
	}
	def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())

	err := def.Activate()

	assert.NoError(t, err)
	assert.Equal(t, WorkflowDefinitionStatusActive, def.Status)
}

func TestWorkflowDefinition_Validate(t *testing.T) {
	createdBy := uuidv7.New()
	schema := WorkflowSchema{
		States:       []WorkflowState{{Name: "start", Type: "start"}},
		InitialState: "start",
	}

	def := NewWorkflowDefinition("test_workflow", "Test Workflow", "Description", schema, createdBy)

	err := def.Validate()

	assert.NoError(t, err)
}

func TestWorkflowDefinition_Deprecate(t *testing.T) {
	schema := WorkflowSchema{
		States: []WorkflowState{
			{Name: "start", Type: "start"},
			{Name: "end", Type: "end"},
		},
		Transitions:  []WorkflowTransition{{From: "start", To: "end", Event: "complete"}},
		InitialState: "start",
	}

	t.Run("Success_ActiveToDeprecated", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		err := def.Activate()
		require.NoError(t, err)

		err = def.Deprecate()

		assert.NoError(t, err)
		assert.Equal(t, WorkflowDefinitionStatusDeprecated, def.Status)
	})

	t.Run("Error_DraftToDeprecated", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		assert.Equal(t, WorkflowDefinitionStatusDraft, def.Status)

		err := def.Deprecate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only active definitions can be deprecated")
		assert.Equal(t, WorkflowDefinitionStatusDraft, def.Status)
	})

	t.Run("Error_ArchivedToDeprecated", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		def.Archive()
		assert.Equal(t, WorkflowDefinitionStatusArchived, def.Status)

		err := def.Deprecate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only active definitions can be deprecated")
		assert.Equal(t, WorkflowDefinitionStatusArchived, def.Status)
	})

	t.Run("Error_AlreadyDeprecated", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		err := def.Activate()
		require.NoError(t, err)
		err = def.Deprecate()
		require.NoError(t, err)

		err = def.Deprecate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only active definitions can be deprecated")
		assert.Equal(t, WorkflowDefinitionStatusDeprecated, def.Status)
	})
}

func TestWorkflowDefinition_Archive(t *testing.T) {
	schema := WorkflowSchema{
		States: []WorkflowState{
			{Name: "start", Type: "start"},
			{Name: "end", Type: "end"},
		},
		Transitions:  []WorkflowTransition{{From: "start", To: "end", Event: "complete"}},
		InitialState: "start",
	}

	t.Run("Success_DraftToArchived", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		assert.Equal(t, WorkflowDefinitionStatusDraft, def.Status)

		def.Archive()

		assert.Equal(t, WorkflowDefinitionStatusArchived, def.Status)
	})

	t.Run("Success_ActiveToArchived", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		err := def.Activate()
		require.NoError(t, err)

		def.Archive()

		assert.Equal(t, WorkflowDefinitionStatusArchived, def.Status)
	})

	t.Run("Success_DeprecatedToArchived", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		err := def.Activate()
		require.NoError(t, err)
		err = def.Deprecate()
		require.NoError(t, err)

		def.Archive()

		assert.Equal(t, WorkflowDefinitionStatusArchived, def.Status)
	})

	t.Run("AlreadyArchived_NoError", func(t *testing.T) {
		def := NewWorkflowDefinition("test", "Test", "Description", schema, uuidv7.New())
		def.Archive()
		assert.Equal(t, WorkflowDefinitionStatusArchived, def.Status)

		// Calling Archive() on already archived should be idempotent (no error)
		def.Archive()

		assert.Equal(t, WorkflowDefinitionStatusArchived, def.Status)
	})
}
