# Workflows Module

**Status**: Commercial (requires license)  
**Version**: 1.0.0  
**License Tier**: BASIC, PRO, ENTERPRISE

Production-ready workflow management system for orchestrating business processes, approval workflows, and complex multi-step operations.

## Overview

The Workflows module provides a powerful, flexible system for defining and executing business processes. It combines the best practices of modern workflow engines with a clean, maintainable architecture.

### Key Features

- **State Machine** - Clear states and transitions (BPMN-inspired)
- **Visual Workflow Designer** - JSON-based workflow definitions
- **Event-Driven** - Integrated with Promenade event bus
- **Parallel & Sequential** - Support for complex execution patterns
- **Timers & Delays** - Scheduled actions and timeouts
- **Retry & Compensation** - Automatic retries and error handling
- **Complete Audit Trail** - Full execution history
- **Extensible** - Custom activities via plugins
- **Human Tasks** - Manual approval and decision points
- **Priority & SLA** - Priority management and due dates

## Architecture

### Core Concepts

```
WorkflowDefinition (Template)
    ↓
WorkflowInstance (Execution)
    ↓
WorkflowSteps (Audit Trail)
    ↓
WorkflowVariables (Context)
    ↓
WorkflowEvents (Triggers)
```

### Entities

#### 1. WorkflowDefinition

The blueprint for a workflow. Defines:

- **States**: All possible states in the workflow
- **Transitions**: Valid state changes and their conditions
- **Activities**: Actions to perform in each state
- **Schema**: BPMN-inspired JSON structure

```go
type WorkflowDefinition struct {
    Name        string
    Version     int
    Status      string  // draft, active, deprecated, archived
    Definition  WorkflowSchema
    InputSchema json.RawMessage  // JSON Schema for validation
}

type WorkflowSchema struct {
    States       []WorkflowState
    Transitions  []WorkflowTransition
    InitialState string
    FinalStates  []string
}
```

#### 2. WorkflowInstance

A running instance of a workflow. Tracks:

- Current state and status
- Execution context (variables)
- Input/output data
- Assignment and priority
- Timing and SLA

```go
type WorkflowInstance struct {
    DefinitionID      uuidv7.UUID
    Status            string  // pending, running, waiting, completed, failed, etc.
    CurrentState      string
    Context           json.RawMessage  // Workflow variables
    Priority          int              // 1-10
    DueDate           *time.Time
    AssignedTo        *uuidv7.UUID
}
```

#### 3. WorkflowStep

Audit trail of execution. Records:

- Each action performed
- Input/output for each step
- Timing and duration
- Errors and retries

```go
type WorkflowStep struct {
    InstanceID   uuidv7.UUID
    StepNumber   int
    Type         string  // activity, transition, gateway, event, timer
    Status       string  // pending, running, completed, failed, skipped
    Duration     *int64  // milliseconds
}
```

#### 4. WorkflowVariable

Context storage during execution:

- Global, state, or local scope
- Typed values (JSON)
- Set by users or system

#### 5. WorkflowEvent

External events that trigger workflows:

- Signals (webhooks, API calls)
- Messages (from other systems)
- Timers (scheduled events)
- Manual triggers (user actions)

## Use Cases

### 1. Order Fulfillment

```
[New Order] → [Payment Check] → [Inventory Check]
    → [Prepare Shipment] → [Ship] → [Delivered]
```

- Automatic state transitions
- External events (payment confirmed, shipped)
- Timeout handling (cancel if not paid in 24h)

### 2. Approval Workflow

```
[Draft] → [Pending Approval] → [Approved] / [Rejected]
    → [Published] / [Revision]
```

- Manual approval steps
- Assignment to reviewers
- Escalation on timeout
- Comments and feedback

### 3. Employee Onboarding

```
[New Hire] → [Create Accounts] → [Assign Equipment]
    → [Training] → [Active Employee]
```

- Multiple parallel tasks
- Checklist completion
- Integration with HR systems
- Notifications

### 4. Incident Management

```
[Reported] → [Triage] → [In Progress]
    → [Resolved] → [Closed]
```

- Priority management
- SLA tracking
- Escalation rules
- Status updates

## Workflow Definition Example

### Simple Approval Workflow

```json
{
  "states": [
    {
      "name": "draft",
      "display_name": "Draft",
      "type": "task",
      "activities": [
        {
          "name": "validate_input",
          "type": "script",
          "config": {
            "script": "validateDocument(context.document)"
          }
        }
      ]
    },
    {
      "name": "pending_approval",
      "display_name": "Pending Approval",
      "type": "task",
      "timeout": "48h",
      "activities": [
        {
          "name": "assign_approver",
          "type": "script",
          "config": {
            "script": "assignToManager(context.department)"
          }
        },
        {
          "name": "send_notification",
          "type": "email",
          "config": {
            "template": "approval_request",
            "to": "{{context.approver_email}}"
          }
        }
      ]
    },
    {
      "name": "approved",
      "display_name": "Approved",
      "type": "task",
      "activities": [
        {
          "name": "publish",
          "type": "http",
          "config": {
            "url": "https://api.example.com/publish",
            "method": "POST",
            "body": "{{context.document}}"
          }
        }
      ]
    },
    {
      "name": "rejected",
      "display_name": "Rejected",
      "type": "end"
    }
  ],
  "transitions": [
    {
      "from": "draft",
      "to": "pending_approval",
      "event": "submit"
    },
    {
      "from": "pending_approval",
      "to": "approved",
      "event": "approve"
    },
    {
      "from": "pending_approval",
      "to": "rejected",
      "event": "reject"
    },
    {
      "from": "pending_approval",
      "to": "draft",
      "event": "request_changes"
    }
  ],
  "initial_state": "draft",
  "final_states": ["approved", "rejected"]
}
```

## Activity Types

### Built-in Activities

1. **HTTP** - Call external APIs
2. **Email** - Send notifications
3. **Script** - Execute JavaScript/expressions
4. **Manual** - Wait for user action
5. **Timer** - Delay execution
6. **Gateway** - Conditional branching
7. **Subprocess** - Start child workflow
8. **Event** - Wait for external event

### Custom Activities

Extend with custom activity handlers:

```go
type ActivityHandler interface {
    Execute(ctx context.Context, config map[string]interface{}, variables map[string]interface{}) (interface{}, error)
    Validate(config map[string]interface{}) error
}
```

## API Endpoints

### Workflow Definitions

```
GET    /api/v1/workflows/definitions       - List definitions
POST   /api/v1/workflows/definitions       - Create definition
GET    /api/v1/workflows/definitions/:id   - Get definition
PUT    /api/v1/workflows/definitions/:id   - Update definition
DELETE /api/v1/workflows/definitions/:id   - Delete definition
POST   /api/v1/workflows/definitions/:id/activate  - Activate
POST   /api/v1/workflows/definitions/:id/versions  - Create new version
```

### Workflow Instances

```
GET    /api/v1/workflows/instances         - List instances
POST   /api/v1/workflows/instances         - Start workflow
GET    /api/v1/workflows/instances/:id     - Get instance details
POST   /api/v1/workflows/instances/:id/events  - Send event to workflow
POST   /api/v1/workflows/instances/:id/pause   - Pause workflow
POST   /api/v1/workflows/instances/:id/resume  - Resume workflow
POST   /api/v1/workflows/instances/:id/cancel  - Cancel workflow
GET    /api/v1/workflows/instances/:id/steps   - Get execution history
GET    /api/v1/workflows/instances/:id/variables - Get variables
PUT    /api/v1/workflows/instances/:id/variables/:name - Set variable
```

### Dashboard & Monitoring

```
GET    /api/v1/workflows/stats             - Overall statistics
GET    /api/v1/workflows/metrics           - Performance metrics
GET    /api/v1/workflows/tasks/my          - My pending tasks
POST   /api/v1/workflows/tasks/:id/complete - Complete task
```

## Configuration

### Module Config (`config/config.{env}.yaml`)

```yaml
workflows:
  enabled: true
  version: "1.0.0"

  # License (commercial)
  license_key: "${WORKFLOWS_LICENSE_KEY}"

  # Execution settings
  execution:
    max_concurrent_instances: 1000
    step_timeout_default: "5m"
    instance_timeout_default: "24h"

  # Retry policy defaults
  retry:
    max_attempts: 3
    initial_interval: "1s"
    multiplier: 2.0
    max_interval: "1m"

  # Cleanup
  purge:
    completed_instances_days: 90
    failed_instances_days: 180
    steps_retention_days: 90
```

## License Tiers

### BASIC

- Up to 100 active workflows
- Basic activities (HTTP, Email, Script)
- 90 days retention

### PRO

- Up to 1000 active workflows
- All activities + Custom activities
- Advanced monitoring
- 180 days retention

### ENTERPRISE

- Unlimited workflows
- Priority support
- Custom integrations
- Unlimited retention
- High availability

## Migration

### Create Workflows Tables

```bash
make migrate-create MODULE=workflows NAME=init
```

Tables created:

- `workflows_definitions` - Workflow templates
- `workflows_instances` - Running workflows
- `workflows_steps` - Execution audit trail
- `workflows_variables` - Context storage
- `workflows_events` - Event triggers

## Testing

```bash
make test-module-workflows
```

Test coverage:

- Entity validation
- State transitions
- Activity execution
- Event handling
- Retry logic
- Timeout handling

## Integration Example

### Start a Workflow

```go
// Create definition
definition := entity.NewWorkflowDefinition(
    "order_fulfillment",
    "Order Fulfillment Process",
    "Handles order from payment to delivery",
    "sales",
    userID,
)

// Define schema
definition.Definition = entity.WorkflowSchema{
    States: []entity.WorkflowState{
        {Name: "pending_payment", DisplayName: "Pending Payment"},
        {Name: "processing", DisplayName: "Processing Order"},
        {Name: "shipped", DisplayName: "Shipped"},
        {Name: "delivered", DisplayName: "Delivered"},
    },
    Transitions: []entity.WorkflowTransition{
        {From: "pending_payment", To: "processing", Event: "payment_received"},
        {From: "processing", To: "shipped", Event: "order_shipped"},
        {From: "shipped", To: "delivered", Event: "delivery_confirmed"},
    },
    InitialState: "pending_payment",
    FinalStates: []string{"delivered"},
}

// Activate
definition.Activate()
workflowDefRepo.Create(ctx, definition)

// Start instance
input := json.RawMessage(`{"order_id": "12345", "amount": 99.99}`)
instance := entity.NewWorkflowInstance(
    definition.ID,
    definition.Version,
    definition.Definition.InitialState,
    input,
    userID,
)

instance.Start()
workflowInstanceRepo.Create(ctx, instance)

// Send event to progress workflow
event := entity.NewWorkflowEvent(
    instance.ID,
    entity.WorkflowEventTypeSignal,
    "payment_received",
    json.RawMessage(`{"payment_id": "67890"}`),
)
workflowEventRepo.Create(ctx, event)
```

## Best Practices

1. **Keep States Simple** - One responsibility per state
2. **Use Events** - Prefer event-driven transitions over polling
3. **Set Timeouts** - Always define timeout behavior
4. **Version Workflows** - Create new versions instead of modifying active ones
5. **Log Everything** - Leverage WorkflowSteps for debugging
6. **Test Thoroughly** - Test all state transitions and error paths
7. **Monitor Performance** - Track execution times and bottlenecks
8. **Handle Failures** - Define compensation logic for critical operations

## Roadmap

- [ ] Visual workflow designer UI
- [ ] BPMN 2.0 import/export
- [ ] Advanced analytics dashboard
- [ ] Workflow templates marketplace
- [ ] Mobile app for task management
- [ ] Integration with popular tools (Zapier, Slack, etc.)

## Support

For issues, questions, or feature requests:

- GitHub Issues: https://github.com/basilex/promenade/issues
- Email: support@promenade.com
- Documentation: https://promenade.com/docs/workflows

---

**License**: Commercial - Requires valid license key  
**Status**: Production Ready  
**Maintained by**: Promenade Team
