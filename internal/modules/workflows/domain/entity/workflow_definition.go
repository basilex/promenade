package entity

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/jsonb"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/lib/pq"
)

// WorkflowDefinitionStatus represents workflow definition status
type WorkflowDefinitionStatus string

const (
	WorkflowDefinitionStatusDraft      WorkflowDefinitionStatus = "draft"
	WorkflowDefinitionStatusActive     WorkflowDefinitionStatus = "active"
	WorkflowDefinitionStatusDeprecated WorkflowDefinitionStatus = "deprecated"
	WorkflowDefinitionStatusArchived   WorkflowDefinitionStatus = "archived"
)

// WorkflowDefinition represents a workflow template/blueprint
type WorkflowDefinition struct {
	ID          uuidv7.UUID              `db:"id" json:"id"`
	Name        string                   `db:"name" json:"name"`
	DisplayName string                   `db:"display_name" json:"display_name"`
	Description string                   `db:"description" json:"description"`
	Version     int                      `db:"version" json:"version"`
	Status      WorkflowDefinitionStatus `db:"status" json:"status"`
	Category    string                        `db:"category" json:"category"`
	Tags        pq.StringArray                `db:"tags" json:"tags"`
	Definition  jsonb.JSON[WorkflowSchema]    `db:"definition" json:"definition"`
	InputSchema *jsonb.JSON[WorkflowSchema]   `db:"input_schema" json:"input_schema,omitempty"`
	CreatedBy   uuidv7.UUID                   `db:"created_by" json:"created_by"`
	CreatedAt   time.Time                `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time                `db:"updated_at" json:"updated_at"`
	DeletedAt   *time.Time               `db:"deleted_at" json:"deleted_at,omitempty"`
}

// WorkflowSchema represents BPMN-inspired workflow definition
type WorkflowSchema struct {
	InitialState string                 `json:"initial_state"`
	States       []WorkflowState        `json:"states"`
	Transitions  []WorkflowTransition   `json:"transitions"`
	Activities   map[string]interface{} `json:"activities,omitempty"`
	RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty"`
}

// WorkflowState represents a state in the workflow
type WorkflowState struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // activity, gateway, event, final
	Activities  []string               `json:"activities,omitempty"`
	Timeout     *int64                 `json:"timeout,omitempty"`
	IsFinal     bool                   `json:"is_final"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WorkflowTransition represents a transition between states
type WorkflowTransition struct {
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Event     string                 `json:"event"`
	Condition *string                `json:"condition,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts    int   `json:"max_attempts"`
	InitialDelay   int64 `json:"initial_delay"`   // milliseconds
	MaxDelay       int64 `json:"max_delay"`       // milliseconds
	BackoffFactor  float64 `json:"backoff_factor"`
}

// NewWorkflowDefinition creates a new workflow definition
func NewWorkflowDefinition(name, displayName, description string, definition WorkflowSchema, createdBy uuidv7.UUID) *WorkflowDefinition {
	now := time.Now()
	wd := &WorkflowDefinition{
		ID:          uuidv7.New(),
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Version:     1,
		Status:      WorkflowDefinitionStatusDraft,
		Category:    "general",
		Tags:        pq.StringArray{},
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	wd.Definition.Set(definition)
	return wd
}

// Validate validates the workflow definition
func (wd *WorkflowDefinition) Validate() error {
	if wd.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if !wd.Definition.Valid {
		return fmt.Errorf("definition is required")
	}
	
	// Validate the schema structure
	if err := wd.Definition.Data.Validate(); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}
	
	return nil
}

// Activate activates the workflow definition
func (wd *WorkflowDefinition) Activate() error {
	if wd.Status != WorkflowDefinitionStatusDraft {
		return fmt.Errorf("only draft definitions can be activated")
	}
	wd.Status = WorkflowDefinitionStatusActive
	wd.UpdatedAt = time.Now()
	return nil
}

// Deprecate deprecates the workflow definition
func (wd *WorkflowDefinition) Deprecate() error {
	if wd.Status != WorkflowDefinitionStatusActive {
		return fmt.Errorf("only active definitions can be deprecated")
	}
	wd.Status = WorkflowDefinitionStatusDeprecated
	wd.UpdatedAt = time.Now()
	return nil
}

// Archive archives the workflow definition
func (wd *WorkflowDefinition) Archive() error {
	wd.Status = WorkflowDefinitionStatusArchived
	wd.UpdatedAt = time.Now()
	return nil
}

// Validate validates the workflow schema structure
func (ws *WorkflowSchema) Validate() error {
	// 1. Basic checks
	if len(ws.States) == 0 {
		return fmt.Errorf("at least one state is required")
	}
	if ws.InitialState == "" {
		return fmt.Errorf("initial state is required")
	}

	// 2. Build state name map for quick lookup
	stateMap := make(map[string]bool)
	for _, state := range ws.States {
		if state.Name == "" {
			return fmt.Errorf("state name cannot be empty")
		}
		if stateMap[state.Name] {
			return fmt.Errorf("duplicate state name: %s", state.Name)
		}
		stateMap[state.Name] = true
	}

	// 3. Initial state must exist in states
	if !stateMap[ws.InitialState] {
		return fmt.Errorf("initial state '%s' not found in states", ws.InitialState)
	}

	// 4. All transitions must reference valid states
	for i, transition := range ws.Transitions {
		if transition.From == "" {
			return fmt.Errorf("transition[%d]: 'from' state cannot be empty", i)
		}
		if transition.To == "" {
			return fmt.Errorf("transition[%d]: 'to' state cannot be empty", i)
		}
		if !stateMap[transition.From] {
			return fmt.Errorf("transition[%d]: 'from' state '%s' not found in states", i, transition.From)
		}
		if !stateMap[transition.To] {
			return fmt.Errorf("transition[%d]: 'to' state '%s' not found in states", i, transition.To)
		}
		if transition.Event == "" {
			return fmt.Errorf("transition[%d]: event name is required", i)
		}
	}

	// 5. At least one terminal state (state with no outgoing transitions or marked as final)
	outgoingTransitions := make(map[string]bool)
	for _, transition := range ws.Transitions {
		outgoingTransitions[transition.From] = true
	}

	hasTerminalState := false
	for _, state := range ws.States {
		if state.IsFinal || !outgoingTransitions[state.Name] {
			hasTerminalState = true
			break
		}
	}
	if !hasTerminalState {
		return fmt.Errorf("at least one terminal state is required (state with no outgoing transitions or marked as final)")
	}

	// 6. All states must be reachable from initial state
	reachable := ws.getReachableStates()
	for _, state := range ws.States {
		if !reachable[state.Name] {
			return fmt.Errorf("state '%s' is unreachable from initial state '%s'", state.Name, ws.InitialState)
		}
	}

	// 7. Detect deadlock cycles (cycles with no exit path to terminal state)
	if err := ws.validateNoCyclesWithoutExit(); err != nil {
		return err
	}

	return nil
}

// getReachableStates performs BFS to find all states reachable from initial state
func (ws *WorkflowSchema) getReachableStates() map[string]bool {
	reachable := make(map[string]bool)
	queue := []string{ws.InitialState}
	reachable[ws.InitialState] = true

	// Build adjacency list
	adjacency := make(map[string][]string)
	for _, transition := range ws.Transitions {
		adjacency[transition.From] = append(adjacency[transition.From], transition.To)
	}

	// BFS traversal
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, next := range adjacency[current] {
			if !reachable[next] {
				reachable[next] = true
				queue = append(queue, next)
			}
		}
	}

	return reachable
}

// validateNoCyclesWithoutExit checks for deadlock cycles (cycles with no path to terminal state)
func (ws *WorkflowSchema) validateNoCyclesWithoutExit() error {
	// Build adjacency list
	adjacency := make(map[string][]string)
	for _, transition := range ws.Transitions {
		adjacency[transition.From] = append(adjacency[transition.From], transition.To)
	}

	// Identify terminal states
	terminalStates := make(map[string]bool)
	for _, state := range ws.States {
		if state.IsFinal || len(adjacency[state.Name]) == 0 {
			terminalStates[state.Name] = true
		}
	}

	// For each state, check if it can reach a terminal state
	for _, state := range ws.States {
		if terminalStates[state.Name] {
			continue // Terminal states are always valid
		}

		// BFS to find if this state can reach any terminal state
		visited := make(map[string]bool)
		queue := []string{state.Name}
		visited[state.Name] = true
		canReachTerminal := false

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			// Check if current is terminal
			if terminalStates[current] {
				canReachTerminal = true
				break
			}

			// Explore neighbors
			for _, next := range adjacency[current] {
				if !visited[next] {
					visited[next] = true
					queue = append(queue, next)
				}
			}
		}

		if !canReachTerminal {
			return fmt.Errorf("deadlock detected: state '%s' cannot reach any terminal state (infinite loop)", state.Name)
		}
	}

	return nil
}

// detectCycles finds all cycles in the workflow graph using DFS
func (ws *WorkflowSchema) detectCycles() [][]string {
	// Build adjacency list
	adjacency := make(map[string][]string)
	for _, transition := range ws.Transitions {
		adjacency[transition.From] = append(adjacency[transition.From], transition.To)
	}

	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range adjacency[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			} else if recStack[neighbor] {
				// Found a cycle - extract it from path
				cycleStart := -1
				for i, state := range path {
					if state == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	// Run DFS from each unvisited state
	for _, state := range ws.States {
		if !visited[state.Name] {
			dfs(state.Name)
		}
	}

	return cycles
}
