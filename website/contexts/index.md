# Contexts Overview

**Bounded Contexts** are autonomous business domains in Promenade Platform. Each context has its own domain model, database schema, and communicates via Event Bus.

## Available Contexts

### Identity Context ✅

**Status**: Production  
**Aggregates**: User, Contact, Profile  
**Documentation**: [Identity Context](/contexts/identity)

Manages user authentication, authorization, and contact information.

**Key Features**:
- JWT-based authentication
- Role-Based Access Control (RBAC)
- Email, phone, address management
- User profiles with social links

### Shared Context ✅

**Status**: Production  
**Aggregates**: Country, Currency, Language, Timezone  
**Documentation**: [Shared Context](/contexts/shared)

Provides reference data for all contexts (read-only).

**Key Features**:
- ISO country codes (250+ countries)
- Currency codes with symbols
- Language codes (ISO 639-1)
- IANA timezone database

### Customer Management ✅

**Status**: Production  
**Aggregates**: Customer, Company, Deal, Interaction, Analytics  
**Documentation**: [Customer Management](/concepts/customer-management)

Complete CRM functionality for managing customers, companies, deals, and sales pipeline.

**Key Features**:
- Customer lifecycle management (Lead → Prospect → Customer → Churned)
- B2B company management with hierarchies
- Deal pipeline tracking with stages and probability
- Interaction tracking (calls, emails, meetings, notes)
- Real-time analytics and reporting (CQRS read models)

### Order Management 📅

**Status**: Planned (Q2 2026)  
**Aggregates**: Order, OrderItem, Fulfillment

Order processing and tracking system.

**Planned Features**:
- Order creation and lifecycle
- Order items with inventory
- Fulfillment saga (payment → inventory → shipping)
- Integration with Customer context

### Billing 📅

**Status**: Planned (Q2 2026)  
**Aggregates**: Invoice, Payment, Subscription

Billing and payment processing.

**Planned Features**:
- Invoice generation
- Payment processing
- Subscription management
- Integration with Order context

### Warehouse 📅

**Status**: Planned (Q3 2026)  
**Aggregates**: Product, Inventory, Warehouse

Inventory and warehouse management.

**Planned Features**:
- Product catalog
- Inventory tracking
- Warehouse locations
- Stock movements

## Context Structure

Each context follows the same pattern:

```
internal/contexts/{context}/
├── {aggregate}/
│   ├── entity.go              # Domain entity
│   ├── entity_test.go         # Tests
│   ├── repository.go          # Repository interface
│   ├── usecase.go            # Business logic
│   ├── usecase_test.go       # Tests
│   └── adapter/
│       ├── http/handler/      # HTTP handlers
│       └── repository/postgres/ # PostgreSQL implementation
├── router.go                  # Context router
└── README.md                  # Context documentation
```

## Context Communication

**Rule**: Contexts communicate ONLY via Event Bus (no direct dependencies).

```go
// ❌ DON'T - Direct import
import "github.com/basilex/promenade/internal/contexts/customer-mgmt/customer"

// ✅ DO - Event Bus
event := bus.NewBaseEvent("user.registered", userID)
eventBus.Publish(ctx, bus.TopicUserRegistered, event)
```

## Context Isolation

Each context is autonomous:

- **Own domain model** - Aggregates, entities, value objects
- **Own database schema** - Tables prefixed with context name
- **Own migrations** - `migrations/{context}/`
- **Own HTTP routes** - Registered via router
- **Own tests** - Unit, smoke, integration

## Adding New Context

1. **Create directory structure**:
   ```bash
   mkdir -p internal/contexts/new-context/{aggregate}
   ```

2. **Create aggregate**:
   - Entity with domain logic
   - Repository interface
   - Use case implementation
   - HTTP handlers

3. **Create migrations**:
   ```bash
   make migrate-new CONTEXT=new-context NAME=create_tables
   ```

4. **Create router**:
   ```go
   // internal/contexts/new-context/router.go
   func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
       // Register routes
   }
   ```

5. **Register in main.go**:
   ```go
   newContextRouter := newcontext.NewRouter(db)
   newContextRouter.RegisterRoutes(api)
   ```

## Next Steps

- [Identity Context](/contexts/identity) - Explore User, Contact, Profile
- [Shared Context](/contexts/shared) - Reference data
- [Architecture Guide](/guide/architecture) - Bounded Contexts deep dive
- [Event Bus](/packages/bus) - Context communication
