# Workflow Design Best Practices

**Level:** Intermediate → Advanced  
**Reading Time:** 20 minutes  
**Applies to:** Workflow schema design

---

## Table of Contents

1. [State Design Principles](#state-design-principles)
2. [Transition Patterns](#transition-patterns)
3. [Anti-Patterns to Avoid](#anti-patterns-to-avoid)
4. [Real-World Examples](#real-world-examples)
5. [Validation Checklist](#validation-checklist)
6. [Performance Considerations](#performance-considerations)
7. [Testing Strategies](#testing-strategies)

---

## State Design Principles

### 1. Single Responsibility

**Each state should represent ONE clear business stage.**

✅ **GOOD:**

```json
{
  "states": ["draft", "awaiting_manager_approval", "awaiting_finance_approval", "approved", "rejected"]
}
```

❌ **BAD:**

```json
{
  "states": [
    "draft",
    "in_approval", // ← Ambiguous: which approver?
    "done" // ← Not specific enough
  ]
}
```

**Why it matters:**

- Clear audit trail
- Easy to understand state history
- Simpler debugging

---

### 2. Descriptive Naming

**State names should explain themselves without documentation.**

✅ **GOOD:**

```json
{
  "states": [
    "pending_customer_payment",
    "payment_received",
    "order_fulfilled",
    "shipment_in_transit",
    "delivered_to_customer"
  ]
}
```

❌ **BAD:**

```json
{
  "states": [
    "state_1", // ← What does this mean?
    "state_2",
    "processing", // ← Processing what?
    "done" // ← Too generic
  ]
}
```

**Naming Convention:**

- Use `lowercase_with_underscores`
- Start with present participle for active states: `processing_payment`
- Past participle for completed states: `payment_received`
- Avoid abbreviations: `awaiting_approval` not `await_appr`

---

### 3. Terminal States

**Always have clear end states where workflow completes.**

✅ **GOOD:**

```json
{
  "states": [
    "pending",
    "processing",
    "completed", // ← Terminal (success)
    "cancelled", // ← Terminal (user action)
    "failed" // ← Terminal (system error)
  ]
}
```

**Terminal State Characteristics:**

- No outgoing transitions
- Represent final business outcome
- Multiple terminals OK (success, failure, cancellation)

**Common Terminal States:**

- Success: `completed`, `approved`, `fulfilled`, `delivered`
- Failure: `failed`, `rejected`, `expired`, `invalid`
- Cancellation: `cancelled`, `withdrawn`, `aborted`

---

### 4. State Granularity

**Find the right level of detail for your use case.**

**Too Fine-Grained (❌):**

```json
{
  "states": [
    "form_opened",
    "field_1_filled",
    "field_2_filled",
    "field_3_filled",
    "validation_started",
    "validation_step_1",
    "validation_step_2",
    "form_submitted"
  ]
}
```

**Too Coarse-Grained (❌):**

```json
{
  "states": ["pending", "done"]
}
```

**Just Right (✅):**

```json
{
  "states": [
    "draft", // User editing
    "validation_pending", // System validating
    "awaiting_approval", // Manager reviewing
    "approved", // Final state
    "rejected" // Final state
  ]
}
```

**Rule of Thumb:**

- 3-10 states for simple workflows
- 10-30 states for complex workflows
- 30+ states: Consider breaking into sub-workflows

---

## Transition Patterns

### 1. Linear Flow (Simplest)

**Use case:** Sequential process with no branches

```json
{
  "states": ["draft", "submitted", "processing", "completed"],
  "initial_state": "draft",
  "transitions": [
    { "from": "draft", "to": "submitted", "event": "submit" },
    { "from": "submitted", "to": "processing", "event": "begin_processing" },
    { "from": "processing", "to": "completed", "event": "finish" }
  ]
}
```

**Visual:**

```
draft → submitted → processing → completed
```

**When to use:**

- Simple approvals
- Document processing
- Sequential operations

---

### 2. Branching (Decision Points)

**Use case:** Different outcomes based on conditions

```json
{
  "states": ["pending", "manager_review", "approved", "rejected"],
  "initial_state": "pending",
  "transitions": [
    { "from": "pending", "to": "manager_review", "event": "submit" },
    { "from": "manager_review", "to": "approved", "event": "approve" },
    { "from": "manager_review", "to": "rejected", "event": "reject" }
  ]
}
```

**Visual:**

```
                    manager_review
                    /            \
              approve            reject
               /                    \
          approved                rejected
```

**When to use:**

- Approval workflows
- Quality checks
- Risk assessment

---

### 3. Multi-Stage Approval

**Use case:** Multiple sequential approvers

```json
{
  "states": ["draft", "manager_approval", "director_approval", "finance_approval", "approved", "rejected"],
  "initial_state": "draft",
  "transitions": [
    { "from": "draft", "to": "manager_approval", "event": "submit" },
    { "from": "manager_approval", "to": "director_approval", "event": "manager_approve" },
    { "from": "manager_approval", "to": "rejected", "event": "manager_reject" },
    { "from": "director_approval", "to": "finance_approval", "event": "director_approve" },
    { "from": "director_approval", "to": "rejected", "event": "director_reject" },
    { "from": "finance_approval", "to": "approved", "event": "finance_approve" },
    { "from": "finance_approval", "to": "rejected", "event": "finance_reject" }
  ]
}
```

**Visual:**

```
draft → manager_approval → director_approval → finance_approval → approved
            │                    │                    │
            └─────────────┬──────┴────────────────────┘
                          ▼
                      rejected
```

**Best Practice:** Each stage can reject independently.

---

### 4. Retry Loop (Error Recovery)

**Use case:** Automatic retry on temporary failures

```json
{
  "states": ["pending", "processing", "retry", "completed", "failed"],
  "initial_state": "pending",
  "transitions": [
    { "from": "pending", "to": "processing", "event": "start" },
    { "from": "processing", "to": "completed", "event": "success" },
    { "from": "processing", "to": "retry", "event": "temporary_error" },
    { "from": "retry", "to": "processing", "event": "retry_attempt" },
    { "from": "retry", "to": "failed", "event": "max_retries_reached" }
  ]
}
```

**Visual:**

```
              ┌──────────────────┐
              │                  │
pending → processing ─────> completed
              │     │
              │     └──> retry ──┘
              │           │
              │           └──> failed
              │
              └──> (direct failure possible)
```

**Key Points:**

- Cycle is valid because it has exit: `retry → failed`
- Prevents infinite loops
- Use for network calls, external API, file operations

---

### 5. Parallel Paths (Future Implementation)

**Use case:** Multiple independent tasks

```json
{
  "states": ["start", "legal_review", "technical_review", "both_approved", "rejected"],
  "initial_state": "start",
  "transitions": [
    { "from": "start", "to": "legal_review", "event": "start_legal" },
    { "from": "start", "to": "technical_review", "event": "start_technical" },
    { "from": "legal_review", "to": "both_approved", "event": "legal_ok" },
    { "from": "technical_review", "to": "both_approved", "event": "technical_ok" },
    { "from": "legal_review", "to": "rejected", "event": "legal_reject" },
    { "from": "technical_review", "to": "rejected", "event": "technical_reject" }
  ]
}
```

**Visual:**

```
            start
           /     \
     legal_review  technical_review
           \     /
        both_approved
```

**Note:** Currently requires application logic to manage parallel execution.

---

### 6. Conditional Branching

**Use case:** Different paths based on business rules

```json
{
  "states": [
    "pending",
    "auto_approved", // Amount < $1000
    "manager_review", // $1000-$10000
    "director_review", // > $10000
    "approved",
    "rejected"
  ],
  "initial_state": "pending",
  "transitions": [
    { "from": "pending", "to": "auto_approved", "event": "low_amount" },
    { "from": "pending", "to": "manager_review", "event": "medium_amount" },
    { "from": "pending", "to": "director_review", "event": "high_amount" },
    { "from": "auto_approved", "to": "approved", "event": "finalize" },
    { "from": "manager_review", "to": "approved", "event": "approve" },
    { "from": "manager_review", "to": "rejected", "event": "reject" },
    { "from": "director_review", "to": "approved", "event": "approve" },
    { "from": "director_review", "to": "rejected", "event": "reject" }
  ]
}
```

**Visual:**

```
            pending
           /   |    \
     low  /    |mid  \ high
         /     |      \
auto_approved manager director
    |         /  \    /  \
  finalize   /    \  /    \
    |       /      \/      \
    └─> approved   rejected
```

**Implementation:** Application logic determines which event to trigger based on amount.

---

## Anti-Patterns to Avoid

### 1. Unreachable States ❌

**Problem:** States with no path from initial state

```json
{
  "states": ["start", "middle", "orphaned", "end"],
  "initial_state": "start",
  "transitions": [
    { "from": "start", "to": "middle", "event": "next" },
    { "from": "middle", "to": "end", "event": "finish" }
    // "orphaned" has no incoming transitions!
  ]
}
```

**Validation Error:**

```
state 'orphaned' is unreachable from initial state 'start'
```

**Fix:**

- Add transition to orphaned state
- Remove orphaned state if not needed

---

### 2. Dead-End States (Except Terminals) ❌

**Problem:** Non-terminal states with no exit

```json
{
  "states": ["start", "processing", "stuck"],
  "initial_state": "start",
  "transitions": [
    { "from": "start", "to": "processing", "event": "begin" },
    { "from": "processing", "to": "stuck", "event": "error" }
    // "stuck" has no outgoing transitions!
  ]
}
```

**Issue:** Workflow can't complete, instances get stuck

**Fix:**

- Add transitions out of "stuck" (retry, cancel, etc.)
- Or make it a terminal state if that's the intent

---

### 3. Cycles Without Exits ❌

**Problem:** Infinite loops with no escape

```json
{
  "states": ["start", "state_a", "state_b"],
  "initial_state": "start",
  "transitions": [
    { "from": "start", "to": "state_a", "event": "begin" },
    { "from": "state_a", "to": "state_b", "event": "forward" },
    { "from": "state_b", "to": "state_a", "event": "back" }
    // Cycle A ↔ B with no exit!
  ]
}
```

**Validation Error:**

```
cycle detected without exit: state_a -> state_b -> state_a
```

**Fix:** Add at least one exit transition

```json
{
  "transitions": [
    ...
    {"from": "state_a", "to": "completed", "event": "finish"}  // Exit
  ]
}
```

---

### 4. Ambiguous State Names ❌

**Problem:** States that don't clearly indicate status

```json
{
  "states": ["pending", "in_progress", "processing", "working"]
  // All sound similar!
}
```

**Better:**

```json
{
  "states": [
    "draft", // User editing
    "validation_in_progress", // System validating
    "awaiting_approval", // Manager reviewing
    "executing_payment" // Payment processing
  ]
}
```

---

### 5. Too Many States ❌

**Problem:** Overly complex workflows

```json
{
  "states": [
    "state_1", "state_2", "state_3", ..., "state_97", "state_98"
  ]
}
```

**Issues:**

- Hard to understand
- Hard to maintain
- Performance degradation
- Validation slowdown

**Solution:** Break into sub-workflows

```
Main Workflow (5 states)
  ├─> Sub-Workflow A (10 states)
  ├─> Sub-Workflow B (15 states)
  └─> Sub-Workflow C (8 states)
```

---

### 6. Generic Event Names ❌

**Problem:** Events that don't describe the action

```json
{
  "transitions": [
    { "from": "pending", "to": "approved", "event": "next" },
    { "from": "approved", "to": "completed", "event": "do" },
    { "from": "draft", "to": "submitted", "event": "action" }
  ]
}
```

**Better:**

```json
{
  "transitions": [
    { "from": "pending", "to": "approved", "event": "manager_approve" },
    { "from": "approved", "to": "completed", "event": "finalize_approval" },
    { "from": "draft", "to": "submitted", "event": "submit_for_review" }
  ]
}
```

---

## Real-World Examples

### Example 1: E-commerce Order Fulfillment

```json
{
  "name": "order_fulfillment",
  "display_name": "E-commerce Order Fulfillment",
  "category": "logistics",
  "schema": {
    "states": [
      "payment_pending",
      "payment_confirmed",
      "warehouse_picking",
      "quality_check",
      "packaging",
      "ready_for_shipment",
      "in_transit",
      "delivered",
      "cancelled",
      "refunded"
    ],
    "initial_state": "payment_pending",
    "transitions": [
      { "from": "payment_pending", "to": "payment_confirmed", "event": "payment_received" },
      { "from": "payment_pending", "to": "cancelled", "event": "payment_timeout" },
      { "from": "payment_confirmed", "to": "warehouse_picking", "event": "start_picking" },
      { "from": "warehouse_picking", "to": "quality_check", "event": "picking_complete" },
      { "from": "quality_check", "to": "packaging", "event": "quality_approved" },
      { "from": "quality_check", "to": "warehouse_picking", "event": "quality_failed" },
      { "from": "packaging", "to": "ready_for_shipment", "event": "package_ready" },
      { "from": "ready_for_shipment", "to": "in_transit", "event": "shipped" },
      { "from": "in_transit", "to": "delivered", "event": "delivery_confirmed" },
      { "from": "delivered", "to": "refunded", "event": "customer_return" },
      { "from": "payment_confirmed", "to": "cancelled", "event": "customer_cancel" },
      { "from": "warehouse_picking", "to": "cancelled", "event": "out_of_stock" }
    ]
  }
}
```

**Key Features:**

- Linear main flow with quality check retry
- Multiple cancellation points
- Refund as post-delivery state

---

### Example 2: Employee Onboarding

```json
{
  "name": "employee_onboarding",
  "display_name": "New Employee Onboarding Process",
  "category": "hr",
  "schema": {
    "states": [
      "offer_accepted",
      "background_check",
      "equipment_ordered",
      "workspace_prepared",
      "it_account_created",
      "training_scheduled",
      "first_day_completed",
      "onboarding_complete",
      "offer_withdrawn"
    ],
    "initial_state": "offer_accepted",
    "transitions": [
      { "from": "offer_accepted", "to": "background_check", "event": "start_verification" },
      { "from": "background_check", "to": "equipment_ordered", "event": "check_passed" },
      { "from": "background_check", "to": "offer_withdrawn", "event": "check_failed" },
      { "from": "equipment_ordered", "to": "workspace_prepared", "event": "equipment_arrived" },
      { "from": "workspace_prepared", "to": "it_account_created", "event": "workspace_ready" },
      { "from": "it_account_created", "to": "training_scheduled", "event": "accounts_ready" },
      { "from": "training_scheduled", "to": "first_day_completed", "event": "first_day_done" },
      { "from": "first_day_completed", "to": "onboarding_complete", "event": "training_complete" },
      { "from": "offer_accepted", "to": "offer_withdrawn", "event": "candidate_declined" }
    ]
  }
}
```

**Key Features:**

- Linear checklist-style flow
- Early exit on background check failure
- Clear completion state

---

### Example 3: Incident Management

```json
{
  "name": "incident_management",
  "display_name": "IT Incident Management",
  "category": "support",
  "schema": {
    "states": [
      "reported",
      "triaged",
      "assigned",
      "investigating",
      "fix_identified",
      "fix_applied",
      "testing",
      "resolved",
      "closed",
      "reopened"
    ],
    "initial_state": "reported",
    "transitions": [
      { "from": "reported", "to": "triaged", "event": "triage" },
      { "from": "triaged", "to": "assigned", "event": "assign_engineer" },
      { "from": "assigned", "to": "investigating", "event": "start_investigation" },
      { "from": "investigating", "to": "fix_identified", "event": "solution_found" },
      { "from": "investigating", "to": "triaged", "event": "escalate" },
      { "from": "fix_identified", "to": "fix_applied", "event": "apply_fix" },
      { "from": "fix_applied", "to": "testing", "event": "begin_testing" },
      { "from": "testing", "to": "resolved", "event": "test_passed" },
      { "from": "testing", "to": "investigating", "event": "test_failed" },
      { "from": "resolved", "to": "closed", "event": "customer_confirmed" },
      { "from": "closed", "to": "reopened", "event": "issue_recurred" },
      { "from": "reopened", "to": "investigating", "event": "reinvestigate" }
    ]
  }
}
```

**Key Features:**

- Escalation path (investigating → triaged)
- Test failure retry (testing → investigating)
- Reopen capability (closed → reopened → investigating)

---

## Validation Checklist

Before activating a workflow, verify:

### ✅ Structure

- [ ] Has at least 2 states (initial + terminal)
- [ ] Initial state is defined
- [ ] Initial state exists in states array
- [ ] All transitions reference valid states
- [ ] No duplicate state names
- [ ] State names follow naming convention

### ✅ Graph Integrity

- [ ] All states reachable from initial state
- [ ] At least one terminal state exists
- [ ] All cycles have exit transitions
- [ ] No dead-end non-terminal states
- [ ] Clear path to completion

### ✅ Business Logic

- [ ] States represent meaningful business stages
- [ ] Transitions have descriptive event names
- [ ] Error states have recovery paths
- [ ] Cancellation paths exist where needed
- [ ] Approvers clearly identified in state names

### ✅ Performance

- [ ] States count: < 200 (recommended)
- [ ] Transitions count: < 1000 (recommended)
- [ ] No deeply nested cycles (> 10 levels)
- [ ] Tested with validation endpoint

### ✅ Testing

- [ ] Happy path tested (initial → terminal)
- [ ] Error paths tested (initial → error → recovery)
- [ ] All branching points tested
- [ ] Cycle exit tested (if applicable)
- [ ] State history verified

---

## Performance Considerations

### Validation Performance

| Graph Size | States  | Transitions | Typical Time |
| ---------- | ------- | ----------- | ------------ |
| Small      | < 20    | < 50        | < 100µs      |
| Medium     | 20-100  | 50-500      | < 500µs      |
| Large      | 100-200 | 500-1000    | < 2ms        |
| Very Large | > 200   | > 1000      | 2-10ms       |

**Optimization Tips:**

1. **Reduce State Count**

   - Merge similar states
   - Remove unnecessary intermediate states
   - Break into sub-workflows if > 200 states

2. **Simplify Transitions**

   - Remove redundant paths
   - Combine similar transitions
   - Avoid deeply nested cycles

3. **Database Queries**
   - Use pagination for instance lists
   - Index external_reference for lookups
   - Archive completed instances periodically

---

## Testing Strategies

### 1. Unit Test Schema Validation

```bash
# Test valid schema
curl -X POST http://localhost:8081/api/v1/workflows/definitions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @valid_schema.json

# Test invalid schema (should fail)
curl -X POST http://localhost:8081/api/v1/workflows/definitions \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d @invalid_schema.json
```

### 2. Integration Test Transitions

```bash
# Start instance
INSTANCE_ID=$(curl -X POST .../instances -d '...' | jq -r '.data.id')

# Test each transition
curl -X POST .../instances/$INSTANCE_ID/transition -d '{"event": "submit"}'
curl -X POST .../instances/$INSTANCE_ID/transition -d '{"event": "approve"}'

# Verify final state
curl -X GET .../instances/$INSTANCE_ID | jq '.data.current_state'
```

### 3. Load Test

```bash
# Create 100 instances concurrently
for i in {1..100}; do
  curl -X POST .../instances -d '...' &
done
wait

# Measure performance
time curl -X GET .../instances?definition_id=$WORKFLOW_ID
```

---

## Summary

### Key Takeaways

1. **State Design:**

   - Descriptive names
   - Single responsibility
   - Clear terminals

2. **Transitions:**

   - Action verbs for events
   - Exit paths from cycles
   - Multiple branches OK

3. **Avoid:**

   - Unreachable states
   - Cycles without exits
   - Generic names
   - Too many states

4. **Test:**
   - Validate schema
   - Test all paths
   - Load test production scenarios

---

**Next Steps:**

- [Read Validation Architecture](VALIDATION_ARCHITECTURE.md)
- [Review Test Status](TEST_STATUS.md)
- [Try Quick Start](QUICK_START.md)

---

**Version:** 1.0.0  
**Last Updated:** December 26, 2025  
**Maintained by:** Promenade Workflows Team
