# Customer Management Context

**Customer Management Context** - CRM core functionality for managing customers, companies, and sales pipeline.

## Overview

🚧 **Status**: In Development

The Customer Management Context will provide:

- **Customer Lifecycle** - Lead → Prospect → Customer → Churned
- **Company Management** - B2B customer organizations
- **Deal Pipeline** - Sales opportunities and tracking
- **Interaction History** - Calls, emails, meetings

## Planned Aggregates

### Customer Aggregate

Customer lifecycle and segmentation.

**Entity**:
```go
type Customer struct {
    ID          uuid.UUID
    Email       valueobject.Email
    Name        string
    Status      CustomerStatus  // lead, prospect, customer, churned
    Tier        CustomerTier    // bronze, silver, gold, platinum
    Source      string
    CreatedAt   time.Time
}
```

**Lifecycle**:
```
Lead → Prospect → Customer → Churned
```

### Company Aggregate (Planned)

B2B customer organizations.

### Deal Aggregate (Planned)

Sales pipeline and opportunity tracking.

### Interaction Aggregate (Planned)

Customer interaction history.

## Next Steps

- [Identity Context](/contexts/identity) - User management
- [Shared Context](/contexts/shared) - Reference data
- [Architecture Guide](/guide/architecture) - DDD concepts
