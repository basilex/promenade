# Bounded Contexts Strategy

**Bounded Contexts** are the cornerstone of Promenade's architecture - autonomous business domains with clear boundaries and independent evolution.

---

## Core Concept

**Definition**: A Bounded Context is a logical boundary within which a particular domain model is defined and applicable. Each context has its own ubiquitous language, database schema, and business logic.

**Key Principle**: **Context Isolation** - Contexts never directly depend on each other. They communicate only via Event Bus (asynchronous, decoupled).

---

## Why Bounded Contexts?

### Problems They Solve

**Without Bounded Contexts**:
- God models (User table with 50+ columns)
- Tight coupling between features
- Cascade failures
- Hard to scale teams
- Difficult to refactor

**With Bounded Contexts**:
-  Clear boundaries - each context owns its domain
-  Independent deployment - update one without affecting others
-  Team autonomy - different teams work on different contexts
-  Technology diversity - choose best tool per context
-  Resilience - one context failure doesn't crash system

---

## Promenade Contexts

### Production Contexts

#### 1. Shared Context

**Purpose**: Reference data shared across all contexts (read-only)

**Aggregates**:
- Country (ISO 3166-1, 249 countries)
- Currency (ISO 4217, 179 currencies)
- Language (ISO 639-1, 184 languages)
- Timezone (IANA, 594 timezones)

**Database**: `shared_*` tables  
**Communication**: Rarely publishes events (reference data changes)  
**Documentation**: [Shared Context README](../../internal/contexts/shared/README.md)

---

#### 2. Identity Context

**Purpose**: User authentication, authorization, profile management

**Aggregates**:
- User (authentication, password, status)
- Contact (email, phone, address with verification)
- Profile (display name, bio, avatar, localization)
- Role (RBAC roles)
- Permission (RBAC permissions)

**Database**: `identity_*` tables  
**Communication**: Publishes events (user.registered, contact.verified, user.activated)  
**Documentation**: [Identity Context README](../../internal/contexts/identity/README.md)

**Key Features**:
- JWT authentication (15-min access, 7-day refresh)
- RBAC with 5 system roles (superadmin, admin, manager, user, guest)
- 29+ permissions across all resources
- Rate limiting (Login: 5/min, Register: 3/min)
- Token revocation with Redis

---

#### 3. Customer Management Context

**Purpose**: CRM functionality - customers, companies, deals, interactions

**Aggregates**:
- Customer (lifecycle, segmentation, contact info)
- Company (planned)
- Deal (planned)
- Interaction (planned)

**Database**: `customer_*` tables  
**Communication**: Publishes events (customer.created, deal.won)  
**Documentation**: [Customer Management README](../../internal/contexts/customer-mgmt/README.md)

---

### Planned Contexts

#### 4. Order Management Context (Q2 2026)

**Purpose**: Order processing, fulfillment, tracking

**Aggregates**:
- Order (creation, status, items)
- OrderItem (products, quantities, prices)
- Fulfillment (shipping, delivery tracking)

**Database**: `order_*` tables  
**Communication**: Publishes events (order.created, order.confirmed, order.shipped)

---

#### 5. Billing Context (Q2 2026)

**Purpose**: Invoicing, payments, subscriptions

**Aggregates**:
- Invoice (generation, items, taxes)
- Payment (processing, status, reconciliation)
- Subscription (recurring billing, plans)

**Database**: `billing_*` tables  
**Communication**: Publishes events (invoice.generated, payment.received, payment.failed)

---

#### 6. Warehouse Context (Q3 2026)

**Purpose**: Inventory management, stock tracking

**Aggregates**:
- Inventory (stock levels, locations)
- Stock (movements, adjustments)

**Database**: `warehouse_*` tables  
**Communication**: Publishes events (stock.low, inventory.adjusted)

---

## Context Communication

### Event Bus Pattern

**Rule**: Contexts communicate ONLY via Event Bus (no direct dependencies)

```
                    
  Identity      user.registered     Customer   
   Context     >  Management  
                                    Context    
                    
                                         
        contact.verified                 customer.created
                                         

             Event Bus (Memory/Redis)             
                                                  
  Topics: identity.*, customer.*, order.*        
  Retry: 3 attempts, exponential backoff         
  Performance: 377K events/sec (Memory)          

                                         
        order.created                    payment.received
                                         
                    
    Order                           Billing    
   Context                          Context    
                                               
                    
```

**Benefits**:
-  Asynchronous - no blocking calls
-  Decoupled - contexts don't know about each other
-  Resilient - retry policy handles transient failures
-  Scalable - can add new contexts without changing existing ones

**Event Bus Documentation**: [pkg/bus/README.md](../../pkg/bus/README.md)

---

## Context Structure

### Directory Layout

Each context follows consistent structure:

```
internal/contexts/{context-name}/
 README.md                 # Context documentation
 router.go                 # HTTP routes registration
 {aggregate}/             # One directory per aggregate
     entity.go            # Domain entity (aggregate root)
     entity_test.go       # Entity tests
     repository.go        # Repository interface
     usecase.go           # Business logic (use cases)
     usecase_test.go      # Use case tests
     adapter/
         http/handler/
            {aggregate}_handler.go    # HTTP handlers
            dto/
                {aggregate}_dto.go    # Data transfer objects
         repository/postgres/
             base_repository.go        # BaseRepository (per context)
             {aggregate}_repository.go # PostgreSQL implementation
```

**Example** (Identity Contact Aggregate):

```
internal/contexts/identity/
 README.md
 router.go
 contact/
    entity.go              # Contact aggregate root
    entity_test.go         # 35 tests
    repository.go          # IRepository interface
    usecase.go             # IUseCase interface + implementation
    usecase_test.go        # 52 tests
    adapter/
        http/handler/
            contact_handler.go
            dto/
                contact_dto.go
        repository/postgres/
            base_repository.go
            contact_repository.go
```

---

## Database Schema

### Namespace-Based Tables

Each context owns its tables with namespace prefix:

**Shared Context**: `shared_countries`, `shared_currencies`, `shared_languages`, `shared_timezones`  
**Identity Context**: `identity_users`, `identity_contacts`, `identity_profiles`, `identity_roles`, `identity_permissions`  
**Customer Management**: `customer_customers`, `customer_companies`, `customer_deals`

**Benefits**:
- Clear ownership
- No table name conflicts
- Easy to see which context owns what
- Migration namespaces match table prefixes

### Migration Namespaces

```
migrations/
 core/                    # Core infrastructure (UUID v7, extensions)
 shared/                  # Shared context migrations
 identity/                # Identity context migrations
 customer-mgmt/           # Customer Management migrations
```

**Run migrations**:

```bash
make migrate              # All namespaces (core → shared → identity → customer-mgmt)
make migrate-identity     # Identity context only
make migrate-customer-mgmt # Customer Management only
```

**Migration Documentation**: [migrations/README.md](../../migrations/README.md)

---

## Adding New Context

### Step-by-Step Guide

1. **Create directory structure**:

   ```bash
   mkdir -p internal/contexts/billing/{invoice,payment,subscription}
   mkdir -p internal/contexts/billing/invoice/adapter/{http/handler/dto,repository/postgres}
   ```

2. **Create router** (`internal/contexts/billing/router.go`):

   ```go
   package billing
   
   import (
       "github.com/gin-gonic/gin"
       "github.com/jmoiron/sqlx"
   )
   
   type Router struct {
       invoiceHandler *invoiceHTTP.InvoiceHandler
   }
   
   func NewRouter(db *sqlx.DB) *Router {
       // Initialize Invoice aggregate
       invoiceRepo := invoiceRepo.NewInvoiceRepository(db)
       invoiceUseCase := invoice.NewUseCase(invoiceRepo)
       invoiceHandler := invoiceHTTP.NewInvoiceHandler(invoiceUseCase)
       
       return &Router{
           invoiceHandler: invoiceHandler,
       }
   }
   
   func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
       billing := api.Group("/billing")
       {
           invoices := billing.Group("/invoices")
           {
               invoices.POST("", r.invoiceHandler.Create)
               invoices.GET("/:id", r.invoiceHandler.GetByID)
           }
       }
   }
   ```

3. **Register in main.go**:

   ```go
   // cmd/api/main.go
   
   import "github.com/basilex/promenade/internal/contexts/billing"
   
   func main() {
       // ... existing setup
       
       // Initialize Billing Context
       billingRouter := billing.NewRouter(db)
       
       // Register routes
       api := r.Group("/api")
       {
           v1 := api.Group("/v1")
           {
               billingRouter.RegisterRoutes(v1)  // /api/v1/billing/*
           }
       }
   }
   ```

4. **Create migrations**:

   ```bash
   make migrate-new CONTEXT=billing NAME=add_invoices_table
   ```

5. **Publish domain events**:

   ```go
   // In Invoice UseCase
   event := bus.NewBaseEvent("billing.invoice.generated", invoice.ID)
   eventBus.Publish(ctx, bus.TopicInvoiceGenerated, event)
   ```

6. **Subscribe to events** (if needed):

   ```go
   // Listen to order.created from Order Management Context
   eventBus.Subscribe(bus.TopicOrderCreated, h.HandleOrderCreated)
   ```

---

## Context Boundaries

### What Belongs in a Context?

** Include**:

- Aggregates with strong cohesion
- Related use cases
- Domain events for the context
- Value objects used only in this context

** Exclude**:

- Shared value objects (put in `pkg/valueobject/`)
- Infrastructure concerns (put in `internal/infrastructure/`)
- Cross-context queries (use Event Bus)

### Example: User vs Customer

**Identity Context** (User aggregate):

- Authentication (login, password)
- Authorization (roles, permissions)
- User status (active, suspended, banned)
- Basic contact info (for login)

**Customer Management Context** (Customer aggregate):

- Customer lifecycle (lead → prospect → customer)
- CRM data (deals, interactions, notes)
- Customer segmentation
- Business relationships

**Why separate?**:

- User can exist without being a customer
- Customer can be managed by multiple users (sales team)
- Different teams work on authentication vs CRM
- Different update frequencies

---

## Best Practices

### DO 

- **Keep contexts small** - 3-5 aggregates per context is ideal
- **Use Event Bus** for all cross-context communication
- **Version events** if schema changes (user.registered.v1, user.registered.v2)
- **Document context README** with aggregates, events, and integration points
- **Test in isolation** - each context should have its own test suite

### DON'T 

- **DON'T share database tables** between contexts
- **DON'T import other context packages** (use Event Bus)
- **DON'T create "Shared" aggregate** (use Shared Context or pkg/)
- **DON'T put everything in one context** (God Context antipattern)
- **DON'T skip event versioning** (breaking changes will break subscribers)

---

## Real-World Example: User Registration Flow

### Cross-Context Orchestration

1. **Identity Context** receives registration request:

   ```go
   // Identity UseCase
   user, _ := user.NewUser(email, name, password)
   userRepo.Create(ctx, user)
   
   // Publish event
   event := bus.NewBaseEvent("identity.user.registered", user.ID)
   eventBus.Publish(ctx, bus.TopicUserRegistered, event)
   ```

2. **Customer Management Context** listens to user.registered:

   ```go
   // Customer Management Subscriber
   func (h *CustomerHandler) HandleUserRegistered(ctx context.Context, e bus.Event) error {
       userID := e.AggregateID()
       
       // Create customer record
       customer := customer.NewCustomer(userID)
       return customerRepo.Create(ctx, customer)
   }
   ```

3. **Billing Context** also listens:

   ```go
   // Billing Subscriber
   func (h *BillingHandler) HandleUserRegistered(ctx context.Context, e bus.Event) error {
       userID := e.AggregateID()
       
       // Create billing account
       account := account.NewAccount(userID)
       return accountRepo.Create(ctx, account)
   }
   ```

**Result**: One registration triggers actions across 3 contexts (Identity, Customer, Billing) without direct dependencies.

---

## Related Documentation

- [Clean Architecture with DDD](clean-architecture.md)
- [Event-Driven Architecture](event-driven.md)
- [Identity Context README](../../internal/contexts/identity/README.md)
- [Shared Context README](../../internal/contexts/shared/README.md)
- [Customer Management README](../../internal/contexts/customer-mgmt/README.md)
- [Event Bus Guide](../../pkg/bus/README.md)

---

**Last Updated**: December 29, 2025  
**Status**: Production pattern (3 contexts live)  
**Maintainer**: Promenade Team
