# Saga Package

**Purpose:** Distributed transaction orchestration for long-running workflows  
**Status:** Production-ready  
**Tests:** 28 tests, 100% coverage

---

## Overview

The `saga` package provides **Saga pattern implementation** for coordinating distributed transactions across multiple services or bounded contexts. Sagas ensure data consistency without distributed ACID transactions by using compensating actions to rollback completed steps when failures occur.

## Features

- **Distributed Transactions:** Coordinate multi-step workflows across services
- **Automatic Compensation:** Rollback completed steps on failure
- **Builder Pattern:** Fluent API for saga construction
- **Execution Tracking:** Detailed execution results
- **Context Support:** Cancellation and timeout propagation
- **Type-Safe:** Strongly typed step functions
- **Zero Dependencies:** Pure Go implementation

---

## Installation

```go
import "github.com/basilex/promenade/pkg/saga"
```

---

## Quick Start

```go
package main

import (
    "context"
    "github.com/basilex/promenade/pkg/saga"
)

func main() {
    // Create saga with builder
    orderSaga := saga.NewBuilder("OrderFulfillment").
        Step("ReserveInventory", reserveInventory, releaseInventory).
        Step("ChargePayment", chargePayment, refundPayment).
        Step("CreateShipment", createShipment, cancelShipment).
        Build()
    
    // Execute saga
    if err := orderSaga.Execute(context.Background()); err != nil {
        log.Fatal("Saga failed:", err)
    }
}
```

---

## Core Concepts

### What is a Saga?

A **Saga** is a sequence of local transactions where each transaction updates data within a single service. If any transaction fails, the saga executes compensating transactions to undo the changes made by preceding transactions.

### Saga Pattern

```
Step 1: Reserve Inventory 
                              Success → Continue
Step 2: Charge Payment 
                              Success → Continue
Step 3: Create Shipment 
                              Failure → Compensate
                             ↓
Compensation (Reverse Order):
  3. Cancel Shipment
  2. Refund Payment
  1. Release Inventory
```

### When to Use Sagas

- **Distributed transactions** across multiple services
- **Long-running workflows** (hours/days duration)
- **Cross-context operations** (Order → Payment → Inventory)
- **Event-driven architectures** requiring eventual consistency

### When NOT to Use Sagas

- **Single database transactions** - use ACID instead
- **Real-time requirements** - sagas are eventually consistent
- **Simple workflows** - overhead not justified

---

## Usage Examples

### 1. Basic Saga with Compensation

```go
package order

import (
    "context"
    "github.com/basilex/promenade/pkg/saga"
)

func CreateOrderSaga(orderID string) *saga.Saga {
    s := saga.New("CreateOrder")
    
    // Step 1: Reserve inventory
    s.AddStep(saga.Step{
        Name: "ReserveInventory",
        Execute: func(ctx context.Context) error {
            return inventoryService.Reserve(orderID)
        },
        Compensate: func(ctx context.Context) error {
            return inventoryService.Release(orderID)
        },
    })
    
    // Step 2: Charge payment
    s.AddStep(saga.Step{
        Name: "ChargePayment",
        Execute: func(ctx context.Context) error {
            return paymentService.Charge(orderID)
        },
        Compensate: func(ctx context.Context) error {
            return paymentService.Refund(orderID)
        },
    })
    
    // Step 3: Create shipment (no compensation needed)
    s.AddStep(saga.Step{
        Name: "CreateShipment",
        Execute: func(ctx context.Context) error {
            return shippingService.CreateShipment(orderID)
        },
    })
    
    return s
}

// Execute saga
func ProcessOrder(ctx context.Context, orderID string) error {
    orderSaga := CreateOrderSaga(orderID)
    return orderSaga.Execute(ctx)
}
```

### 2. Builder Pattern (Fluent API)

```go
package customer

import (
    "context"
    "github.com/basilex/promenade/pkg/saga"
)

func CustomerRegistrationSaga(email, name string) *saga.Saga {
    return saga.NewBuilder("CustomerRegistration").
        Step("CreateUser", 
            func(ctx context.Context) error {
                return userService.Create(email, name)
            },
            func(ctx context.Context) error {
                return userService.Delete(email)
            },
        ).
        Step("CreateProfile",
            func(ctx context.Context) error {
                return profileService.Create(email)
            },
            func(ctx context.Context) error {
                return profileService.Delete(email)
            },
        ).
        Step("SendWelcomeEmail",
            func(ctx context.Context) error {
                return emailService.SendWelcome(email)
            },
            nil, // No compensation for email
        ).
        Build()
}
```

### 3. Execution with Detailed Results

```go
package main

import (
    "context"
    "log"
    "github.com/basilex/promenade/pkg/saga"
)

func main() {
    orderSaga := CreateOrderSaga("order-123")
    
    // Execute with detailed result
    result := orderSaga.ExecuteWithResult(context.Background())
    
    if !result.Success {
        log.Printf("Saga failed at step: %s\n", result.FailedStep)
        log.Printf("Completed steps: %d\n", result.CompletedSteps)
        log.Printf("Compensated steps: %d\n", result.CompensatedSteps)
        log.Printf("Error: %v\n", result.Error)
        return
    }
    
    log.Printf("Saga completed successfully: %d steps\n", result.CompletedSteps)
}
```

### 4. Context Cancellation

```go
package main

import (
    "context"
    "time"
    "github.com/basilex/promenade/pkg/saga"
)

func ProcessWithTimeout(orderID string) error {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    orderSaga := CreateOrderSaga(orderID)
    
    // Execute with cancellation support
    if err := orderSaga.Execute(ctx); err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("saga timeout: %w", err)
        }
        return err
    }
    
    return nil
}
```

### 5. Step Without Compensation

```go
// Some steps don't need compensation (e.g., notifications)
saga := saga.NewBuilder("Notification").
    Step("CreateNotification",
        func(ctx context.Context) error {
            return notificationService.Create(userID, message)
        },
        func(ctx context.Context) error {
            return notificationService.Delete(notificationID)
        },
    ).
    StepWithoutCompensation("SendEmail",
        func(ctx context.Context) error {
            return emailService.Send(email, subject, body)
        },
    ).
    Build()
```

### 6. Complex Workflow Example

```go
package billing

import (
    "context"
    "github.com/basilex/promenade/pkg/saga"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type SubscriptionSaga struct {
    CustomerID     uuidv7.UUID
    PlanID         string
    PaymentMethod  string
}

func (s *SubscriptionSaga) Create() *saga.Saga {
    return saga.NewBuilder("CreateSubscription").
        // Step 1: Validate payment method
        Step("ValidatePaymentMethod",
            func(ctx context.Context) error {
                return paymentService.ValidateMethod(s.CustomerID, s.PaymentMethod)
            },
            nil, // Validation has no side effects
        ).
        // Step 2: Create subscription record
        Step("CreateSubscription",
            func(ctx context.Context) error {
                return subscriptionService.Create(s.CustomerID, s.PlanID)
            },
            func(ctx context.Context) error {
                return subscriptionService.Delete(s.CustomerID)
            },
        ).
        // Step 3: Charge initial payment
        Step("ChargeInitialPayment",
            func(ctx context.Context) error {
                return paymentService.Charge(s.CustomerID, s.PaymentMethod)
            },
            func(ctx context.Context) error {
                return paymentService.Refund(s.CustomerID)
            },
        ).
        // Step 4: Activate subscription
        Step("ActivateSubscription",
            func(ctx context.Context) error {
                return subscriptionService.Activate(s.CustomerID)
            },
            func(ctx context.Context) error {
                return subscriptionService.Deactivate(s.CustomerID)
            },
        ).
        // Step 5: Send confirmation email (no compensation)
        StepWithoutCompensation("SendConfirmation",
            func(ctx context.Context) error {
                return emailService.SendSubscriptionConfirmation(s.CustomerID)
            },
        ).
        Build()
}

// Execute subscription saga
func CreateSubscription(ctx context.Context, customerID uuidv7.UUID, planID, paymentMethod string) error {
    saga := &SubscriptionSaga{
        CustomerID:    customerID,
        PlanID:        planID,
        PaymentMethod: paymentMethod,
    }
    
    return saga.Create().Execute(ctx)
}
```

### 7. Error Handling

```go
package main

import (
    "context"
    "errors"
    "github.com/basilex/promenade/pkg/saga"
)

var (
    ErrInsufficientInventory = errors.New("insufficient inventory")
    ErrPaymentFailed         = errors.New("payment failed")
    ErrShippingUnavailable   = errors.New("shipping unavailable")
)

func CreateOrderWithErrorHandling(ctx context.Context, orderID string) error {
    orderSaga := saga.NewBuilder("OrderWithErrorHandling").
        Step("ReserveInventory",
            func(ctx context.Context) error {
                if available, err := inventoryService.CheckAvailability(orderID); err != nil {
                    return err
                } else if !available {
                    return ErrInsufficientInventory
                }
                return inventoryService.Reserve(orderID)
            },
            func(ctx context.Context) error {
                return inventoryService.Release(orderID)
            },
        ).
        Step("ChargePayment",
            func(ctx context.Context) error {
                if err := paymentService.Validate(orderID); err != nil {
                    return fmt.Errorf("%w: %v", ErrPaymentFailed, err)
                }
                return paymentService.Charge(orderID)
            },
            func(ctx context.Context) error {
                return paymentService.Refund(orderID)
            },
        ).
        Build()
    
    if err := orderSaga.Execute(ctx); err != nil {
        // Check specific error types
        if errors.Is(err, ErrInsufficientInventory) {
            return fmt.Errorf("cannot fulfill order: %w", err)
        }
        if errors.Is(err, ErrPaymentFailed) {
            return fmt.Errorf("payment processing failed: %w", err)
        }
        return fmt.Errorf("order creation failed: %w", err)
    }
    
    return nil
}
```

---

## API Reference

### Saga

```go
// New creates a new Saga
func New(name string) *Saga

// AddStep adds a step to the saga
func (s *Saga) AddStep(step Step) *Saga

// Execute runs all steps in the saga
// Automatically compensates on failure
func (s *Saga) Execute(ctx context.Context) error

// ExecuteWithResult runs the saga and returns detailed execution result
func (s *Saga) ExecuteWithResult(ctx context.Context) ExecutionResult

// Name returns the saga name
func (s *Saga) Name() string

// Steps returns all steps in the saga
func (s *Saga) Steps() []Step

// StepCount returns the number of steps
func (s *Saga) StepCount() int
```

### Builder

```go
// NewBuilder creates a new saga builder
func NewBuilder(name string) *Builder

// Step adds a step with both execute and compensate functions
func (b *Builder) Step(name string, execute StepFunc, compensate CompensateFunc) *Builder

// StepWithoutCompensation adds a step without compensation function
func (b *Builder) StepWithoutCompensation(name string, execute StepFunc) *Builder

// Build returns the constructed saga
func (b *Builder) Build() *Saga
```

### Types

```go
// Step represents a single step in a saga
type Step struct {
    Name       string
    Execute    StepFunc
    Compensate CompensateFunc // Optional
}

// StepFunc is a function that executes a saga step
type StepFunc func(ctx context.Context) error

// CompensateFunc is a function that compensates (rolls back) a saga step
type CompensateFunc func(ctx context.Context) error

// ExecutionResult represents the result of saga execution
type ExecutionResult struct {
    Success          bool
    FailedStep       string
    CompletedSteps   int
    CompensatedSteps int
    Error            error
}
```

---

## Best Practices

### DO

- **Design idempotent steps** - Steps may be retried
  ```go
  // Good: Check if already executed
  func reserveInventory(ctx context.Context) error {
      if alreadyReserved(orderID) {
          return nil // Idempotent
      }
      return inventory.Reserve(orderID)
  }
  ```

- **Order steps by risk** - Place risky steps early
  ```go
  // Good: Risky payment before low-risk notification
  saga.NewBuilder("Order").
      Step("ChargePayment", ...).    // High risk first
      Step("SendEmail", ...).         // Low risk later
      Build()
  ```

- **Keep compensation simple** - Avoid complex rollback logic
  ```go
  // Good: Simple compensation
  Compensate: func(ctx context.Context) error {
      return inventory.Release(orderID)
  }
  
  // Bad: Complex compensation with nested logic
  Compensate: func(ctx context.Context) error {
      // Too complex - split into multiple steps instead
  }
  ```

- **Use context for timeouts** - Prevent hanging sagas
  ```go
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
  defer cancel()
  saga.Execute(ctx)
  ```

### DON'T

- **Don't nest sagas** - Keep flat structure
  ```go
  // Bad: Nested sagas
  Step("OuterStep", func(ctx context.Context) error {
      innerSaga := saga.New("Inner")
      return innerSaga.Execute(ctx) // Don't do this
  }, nil)
  
  // Good: Flat structure with all steps
  saga.NewBuilder("Combined").
      Step("Step1", ...).
      Step("Step2", ...).
      Build()
  ```

- **Don't ignore compensation errors** - Log and alert
  ```go
  // Bad: Silent failure
  Compensate: func(ctx context.Context) error {
      inventory.Release(orderID) // Ignoring error
      return nil
  }
  
  // Good: Proper error handling
  Compensate: func(ctx context.Context) error {
      if err := inventory.Release(orderID); err != nil {
          logger.Error("Compensation failed", slog.Any("error", err))
          return err
      }
      return nil
  }
  ```

- **Don't share state between steps** - Use context values
  ```go
  // Bad: Shared variable
  var reservationID string
  Step("Reserve", func(ctx context.Context) error {
      reservationID = inventory.Reserve() // Shared state
      return nil
  }, nil)
  
  // Good: Context values
  Step("Reserve", func(ctx context.Context) error {
      id := inventory.Reserve()
      ctx = context.WithValue(ctx, "reservationID", id)
      return nil
  }, nil)
  ```

- **Don't put UI logic in sagas** - Keep pure business logic
  ```go
  // Bad: UI concerns in saga
  Step("ShowNotification", func(ctx context.Context) error {
      ui.ShowSuccess("Order created") // Wrong layer
      return nil
  }, nil)
  
  // Good: Business logic only
  Step("CreateOrder", func(ctx context.Context) error {
      return orderService.Create(orderID)
  }, nil)
  ```

---

## Saga Patterns

### 1. Orchestration Pattern

**Centralized coordinator** controls saga execution:

```go
type OrderOrchestrator struct {
    saga *saga.Saga
}

func (o *OrderOrchestrator) ProcessOrder(ctx context.Context, order Order) error {
    return o.saga.Execute(ctx)
}
```

**Pros**: Centralized control, easier to understand  
**Cons**: Single point of failure, tight coupling

### 2. Choreography Pattern

**Distributed coordination** via events:

```go
// Each service reacts to events
eventBus.Subscribe("order.created", func(ctx context.Context, e Event) error {
    inventoryService.Reserve(e.OrderID)
    eventBus.Publish("inventory.reserved", ...)
})

eventBus.Subscribe("inventory.reserved", func(ctx context.Context, e Event) error {
    paymentService.Charge(e.OrderID)
    eventBus.Publish("payment.charged", ...)
})
```

**Pros**: Loose coupling, no single point of failure  
**Cons**: Harder to understand, distributed tracing needed

### 3. Hybrid Pattern

**Combine both approaches**:

```go
// Orchestrator for critical path
orderSaga := saga.NewBuilder("OrderProcessing").
    Step("Reserve", ...).
    Step("Charge", ...).
    Build()

// Choreography for notifications
eventBus.Publish("order.completed", orderID)
```

---

## Testing

```go
package saga_test

import (
    "context"
    "errors"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/basilex/promenade/pkg/saga"
)

func TestOrderSaga_Success(t *testing.T) {
    executed := []string{}
    
    orderSaga := saga.NewBuilder("Order").
        Step("Reserve",
            func(ctx context.Context) error {
                executed = append(executed, "reserve")
                return nil
            },
            nil,
        ).
        Step("Charge",
            func(ctx context.Context) error {
                executed = append(executed, "charge")
                return nil
            },
            nil,
        ).
        Build()
    
    err := orderSaga.Execute(context.Background())
    
    assert.NoError(t, err)
    assert.Equal(t, []string{"reserve", "charge"}, executed)
}

func TestOrderSaga_Compensation(t *testing.T) {
    executed := []string{}
    compensated := []string{}
    
    orderSaga := saga.NewBuilder("Order").
        Step("Reserve",
            func(ctx context.Context) error {
                executed = append(executed, "reserve")
                return nil
            },
            func(ctx context.Context) error {
                compensated = append(compensated, "unreserve")
                return nil
            },
        ).
        Step("Charge",
            func(ctx context.Context) error {
                executed = append(executed, "charge")
                return errors.New("payment failed")
            },
            func(ctx context.Context) error {
                compensated = append(compensated, "refund")
                return nil
            },
        ).
        Build()
    
    err := orderSaga.Execute(context.Background())
    
    assert.Error(t, err)
    assert.Equal(t, []string{"reserve", "charge"}, executed)
    assert.Equal(t, []string{"unreserve"}, compensated) // Reverse order
}

func TestOrderSaga_DetailedResult(t *testing.T) {
    orderSaga := saga.NewBuilder("Order").
        Step("Step1", func(ctx context.Context) error { return nil }, nil).
        Step("Step2", func(ctx context.Context) error { return errors.New("fail") }, nil).
        Build()
    
    result := orderSaga.ExecuteWithResult(context.Background())
    
    assert.False(t, result.Success)
    assert.Equal(t, "Step2", result.FailedStep)
    assert.Equal(t, 1, result.CompletedSteps)
}
```

---

## Performance Considerations

### Memory Usage

- Each saga instance: ~200 bytes + step closures
- Closure capture: Avoid capturing large objects
- Use context values for large data

### Execution Time

- Saga overhead: ~10μs per step (negligible)
- Total time = sum of step execution times
- Use timeouts to prevent long-running sagas

### Scalability

- Sagas are stateless (no shared state)
- Concurrent saga execution: Safe
- Database bottleneck: Step implementations, not saga framework

---

## Related Packages

- `pkg/bus` - Event Bus for choreography pattern
- `pkg/aggregate` - Domain aggregates coordinated by sagas
- `internal/contexts/order-mgmt` - Order fulfillment sagas (planned)
- `internal/contexts/billing` - Payment sagas (planned)

---

## References

- [Saga Pattern](https://microservices.io/patterns/data/saga.html) - Microservices.io
- [CQRS and Event Sourcing](https://martinfowler.com/eaaDev/EventSourcing.html) - Martin Fowler
- [Eventual Consistency](https://www.allthingsdistributed.com/2008/12/eventually_consistent.html) - Werner Vogels

---

**Last Updated:** 2025-12-28  
**Status:** Production-ready  
**Maintainer:** Promenade Team
