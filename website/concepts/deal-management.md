# Deal Management

**Complete sales pipeline management** with deal lifecycle tracking, stage-based probability, and revenue analytics for B2B and B2C sales processes.

## Overview

Deal Management is an **aggregate within Customer Management context** that handles all sales opportunity tracking functionality. This is where sales pipeline, deal stages, and revenue forecasting are managed.

**Status**:  Production-ready (December 2025)  
**Aggregate**: Deal (within Customer Management context)  
**Endpoints**: 12 HTTP routes  
**Database**: 1 table (`customer_deals`) with soft delete  
**Tests**: 24 tests (100% passing)

---

## Key Features

###  Deal Lifecycle Management
Track deals through sales pipeline stages with automatic probability calculation.

**Stages**:
- **Lead**: Initial opportunity, early contact (10% probability)
- **Qualified**: Qualified prospect with budget/authority (25%)
- **Proposal**: Proposal submitted to customer (50%)
- **Negotiation**: Active negotiation with customer (75%)
- **Closed Won**: Deal successfully closed (100%)
- **Closed Lost**: Deal lost to competitor or abandoned (0%)

**Stage Transitions**:
```
                           
 Lead  >  Qualified  >  Proposal  >  Negotiation 
                           
                                                                   
                                                                   
   >
                              Skip stages allowed                  
                                                                   
   v                                                                v
                                          
 Closed Lost                                             Closed Won   
                                          
```

###  Money Handling
Deal values stored as **cents** in database for precision, displayed as dollars in API.

**Example**:
```json
{
  "value": 50000000,     // $500,000.00 (stored as cents)
  "currency": "USD"
}
```

###  Pipeline Analytics
Real-time statistics for sales pipeline management.

**Available Statistics**:
1. **Pipeline Stats**: Count of deals per stage
2. **Won Deals**: Total count and revenue of won deals
3. **Stage Filtering**: List deals by specific stage

---

## Domain Model

### Deal Aggregate

```go
type Deal struct {
    aggregate.BaseAggregate

    ID           uuidv7.UUID         // Deal ID
    CustomerID   uuidv7.UUID         // Foreign key to Customer
    AssignedTo   uuidv7.UUID         // Sales rep (foreign key to User)
    
    // Deal Info
    Name         string              // Deal name/description
    Value        valueobject.Money   // Deal value (cents)
    Stage        DealStage           // Current pipeline stage
    Probability  int                 // Win probability (0-100%)
    
    // Dates
    ExpectedCloseDate *time.Time     // Forecast close date
    ActualCloseDate   *time.Time     // Actual close date (if closed)
    
    // Metadata
    Source       DealSource          // inbound, outbound, referral, partner
    LossReason   string              // Reason for loss (if closed_lost)
    Tags         []string            // JSONB array
    
    // Audit
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time          // Soft delete
}
```

### Deal Stage Enum

```go
type DealStage string

const (
    DealStageLead        DealStage = "lead"         // 10% probability
    DealStageQualified   DealStage = "qualified"    // 25% probability
    DealStageProposal    DealStage = "proposal"     // 50% probability
    DealStageNegotiation DealStage = "negotiation"  // 75% probability
    DealStageClosedWon   DealStage = "closed_won"   // 100% probability
    DealStageClosedLost  DealStage = "closed_lost"  // 0% probability
)
```

### Deal Source Enum

```go
type DealSource string

const (
    DealSourceInbound  DealSource = "inbound"   // Customer initiated
    DealSourceOutbound DealSource = "outbound"  // Sales outreach
    DealSourceReferral DealSource = "referral"  // Referral partner
    DealSourcePartner  DealSource = "partner"   // Channel partner
)
```

---

## Business Rules

### 1. Deal Creation Rules

**Required Fields**:
- `customer_id`: Must reference existing customer
- `assigned_to`: Must reference existing user (sales rep)
- `name`: Deal name (max 200 chars)
- `value`: Deal amount in cents (must be positive)
- `currency`: ISO 4217 currency code
- `expected_close_date`: Forecast close date (must be future)

**Default Values**:
- `stage`: Always starts as `lead`
- `probability`: Automatically set based on stage (10% for lead)
- `source`: Defaults to `inbound` if not specified

**Validation**:
```go
func (d *Deal) Validate() error {
    if d.CustomerID == uuid.Nil {
        return errors.New("customer_id is required")
    }
    if d.AssignedTo == uuid.Nil {
        return errors.New("assigned_to is required")
    }
    if d.Name == "" || len(d.Name) > 200 {
        return errors.New("name must be 1-200 characters")
    }
    if d.Value.Amount() <= 0 {
        return errors.New("value must be positive")
    }
    return nil
}
```

### 2. Stage Transition Rules

**Valid Transitions**:
- Can move forward through pipeline: `lead → qualified → proposal → negotiation`
- Can skip stages (e.g., `lead → proposal` directly)
- Can move to `closed_won` or `closed_lost` from any stage
- **Cannot** move from closed stages to other stages (terminal states)

**Probability Auto-Update**:
```go
func (d *Deal) MoveToStage(newStage DealStage) error {
    // Validate not moving from terminal stage
    if d.IsTerminalStage() {
        return fmt.Errorf("cannot move from terminal stage %s", d.Stage)
    }
    
    // Update stage and probability
    d.Stage = newStage
    d.Probability = d.StageProbability()
    
    return nil
}

func (d *Deal) StageProbability() int {
    switch d.Stage {
    case DealStageLead:
        return 10
    case DealStageQualified:
        return 25
    case DealStageProposal:
        return 50
    case DealStageNegotiation:
        return 75
    case DealStageClosedWon:
        return 100
    case DealStageClosedLost:
        return 0
    default:
        return 0
    }
}
```

### 3. Win/Loss Rules

**Mark as Won**:
- Sets `stage` to `closed_won`
- Sets `probability` to 100%
- Sets `actual_close_date` to provided date
- **Cannot** win already lost deal

**Mark as Lost**:
- Sets `stage` to `closed_lost`
- Sets `probability` to 0%
- Sets `actual_close_date` to provided date
- Optionally records `loss_reason`
- **Cannot** lose already won deal

---

## API Reference

### Base URL
```
/api/v1/customer-mgmt/deals
```

### Endpoints

#### 1. Create Deal
```http
POST /api/v1/customer-mgmt/deals
Content-Type: application/json

{
  "customer_id": "019b72b7-0636-79b5-b8fe-28b439ec801d",
  "assigned_to": "019b72b6-e679-7635-a86d-ac29a7923f3d",
  "name": "Enterprise Software License",
  "value": 50000000,
  "currency": "USD",
  "expected_close_date": "2026-03-31"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "019b72b7-339c-73ca-850c-8459913f2cde",
    "customer_id": "019b72b7-0636-79b5-b8fe-28b439ec801d",
    "assigned_to": "019b72b6-e679-7635-a86d-ac29a7923f3d",
    "name": "Enterprise Software License",
    "value": 50000000,
    "currency": "USD",
    "stage": "lead",
    "probability": 10,
    "source": "inbound",
    "expected_close_date": "2026-03-31",
    "created_at": "2025-12-31T06:42:47+02:00",
    "updated_at": "2025-12-31T06:42:47+02:00"
  }
}
```

#### 2. Get Deal by ID
```http
GET /api/v1/customer-mgmt/deals/:id
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "019b72b7-339c-73ca-850c-8459913f2cde",
    "customer_id": "019b72b7-0636-79b5-b8fe-28b439ec801d",
    "assigned_to": "019b72b6-e679-7635-a86d-ac29a7923f3d",
    "name": "Enterprise Software License",
    "value": 50000000,
    "currency": "USD",
    "stage": "lead",
    "probability": 10,
    "source": "inbound",
    "expected_close_date": "2026-03-31",
    "actual_close_date": null,
    "loss_reason": null,
    "tags": [],
    "created_at": "2025-12-31T06:42:47+02:00",
    "updated_at": "2025-12-31T06:42:47+02:00"
  }
}
```

#### 3. List Deals (with Pagination)
```http
GET /api/v1/customer-mgmt/deals?page=1&page_size=20
```

**Query Parameters**:
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "019b72b7-339c-73ca-850c-8459913f2cde",
      "name": "Enterprise Software License",
      "stage": "lead",
      "value": 50000000
    }
  ],
  "pagination": {
    "total": 45,
    "page": 1,
    "page_size": 20,
    "total_pages": 3
  }
}
```

#### 4. Update Deal Basic Info
```http
PUT /api/v1/customer-mgmt/deals/:id/basic-info
Content-Type: application/json

{
  "name": "Enterprise License + Support",
  "expected_close_date": "2026-04-30"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "019b72b7-339c-73ca-850c-8459913f2cde",
    "name": "Enterprise License + Support",
    "expected_close_date": "2026-04-30",
    "updated_at": "2025-12-31T06:45:00+02:00"
  }
}
```

#### 5. Update Deal Value
```http
PUT /api/v1/customer-mgmt/deals/:id/value
Content-Type: application/json

{
  "value": 75000000,
  "currency": "USD"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "019b72b7-339c-73ca-850c-8459913f2cde",
    "value": 75000000,
    "currency": "USD",
    "updated_at": "2025-12-31T06:46:00+02:00"
  }
}
```

#### 6. Move Deal to Stage
```http
PUT /api/v1/customer-mgmt/deals/:id/stage
Content-Type: application/json

{
  "stage": "qualified"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "019b72b7-339c-73ca-850c-8459913f2cde",
    "stage": "qualified",
    "probability": 25,
    "updated_at": "2025-12-31T06:47:00+02:00"
  }
}
```

**Error Response** (400 Bad Request - terminal stage):
```json
{
  "status": "error",
  "error": {
    "code": "MOVE_STAGE_FAILED",
    "message": "cannot move from terminal stage closed_won"
  }
}
```

#### 7. Mark Deal as Won
```http
POST /api/v1/customer-mgmt/deals/:id/win
Content-Type: application/json

{
  "close_date": "2025-12-31"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "019b72b7-339c-73ca-850c-8459913f2cde",
    "stage": "closed_won",
    "probability": 100,
    "actual_close_date": "2025-12-31",
    "updated_at": "2025-12-31T06:48:00+02:00"
  }
}
```

#### 8. Mark Deal as Lost
```http
POST /api/v1/customer-mgmt/deals/:id/lose
Content-Type: application/json

{
  "close_date": "2025-12-31",
  "reason": "Lost to competitor - pricing"
}
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "id": "019b72b7-339c-73ca-850c-8459913f2cde",
    "stage": "closed_lost",
    "probability": 0,
    "actual_close_date": "2025-12-31",
    "loss_reason": "Lost to competitor - pricing",
    "updated_at": "2025-12-31T06:49:00+02:00"
  }
}
```

#### 9. Delete Deal (Soft Delete)
```http
DELETE /api/v1/customer-mgmt/deals/:id
```

**Response** (204 No Content)

#### 10. List Deals by Stage
```http
GET /api/v1/customer-mgmt/deals/stage/:stage
```

**Example**:
```http
GET /api/v1/customer-mgmt/deals/stage/lead
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    {
      "id": "019b72b7-339c-73ca-850c-8459913f2cde",
      "name": "Enterprise Software License",
      "value": 50000000,
      "stage": "lead"
    }
  ]
}
```

#### 11. Get Pipeline Statistics
```http
GET /api/v1/customer-mgmt/deals/stats/pipeline
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "stats": {
      "lead": 12,
      "qualified": 8,
      "proposal": 5,
      "negotiation": 3,
      "closed_won": 45,
      "closed_lost": 23
    }
  }
}
```

#### 12. Get Won Deals Statistics
```http
GET /api/v1/customer-mgmt/deals/stats/won
```

**Response** (200 OK):
```json
{
  "status": "success",
  "data": {
    "count": 45,
    "total_value": 2250000000
  }
}
```

**Note**: `total_value` is in cents ($22,500,000.00 in this example)

---

## Usage Examples

### Example 1: Create and Close Deal

```bash
# 1. Create deal
curl -X POST http://localhost:8081/api/v1/customer-mgmt/deals \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "019b72b7-0636-79b5-b8fe-28b439ec801d",
    "assigned_to": "019b72b6-e679-7635-a86d-ac29a7923f3d",
    "name": "Enterprise License",
    "value": 50000000,
    "currency": "USD",
    "expected_close_date": "2026-03-31"
  }'

# Response: deal_id = 019b72b7-339c-73ca-850c-8459913f2cde

# 2. Move through stages
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/019b72b7-339c-73ca-850c-8459913f2cde/stage \
  -H "Content-Type: application/json" \
  -d '{"stage": "qualified"}'

curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/019b72b7-339c-73ca-850c-8459913f2cde/stage \
  -H "Content-Type: application/json" \
  -d '{"stage": "proposal"}'

curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/019b72b7-339c-73ca-850c-8459913f2cde/stage \
  -H "Content-Type: application/json" \
  -d '{"stage": "negotiation"}'

# 3. Close as won
curl -X POST http://localhost:8081/api/v1/customer-mgmt/deals/019b72b7-339c-73ca-850c-8459913f2cde/win \
  -H "Content-Type: application/json" \
  -d '{"close_date": "2025-12-31"}'
```

**Result**: Deal progressed through pipeline and closed as won with 100% probability.

### Example 2: Update Deal Value During Negotiation

```bash
# Deal currently at negotiation stage
DEAL_ID="019b72b7-339c-73ca-850c-8459913f2cde"

# Customer negotiated higher price
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/value \
  -H "Content-Type: application/json" \
  -d '{
    "value": 60000000,
    "currency": "USD"
  }'

# Update expected close date
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/basic-info \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Enterprise License + Premium Support",
    "expected_close_date": "2026-02-28"
  }'
```

### Example 3: Track Lost Deal

```bash
DEAL_ID="019b72b8-1234-5678-abcd-123456789abc"

# Mark as lost with reason
curl -X POST http://localhost:8081/api/v1/customer-mgmt/deals/$DEAL_ID/lose \
  -H "Content-Type: application/json" \
  -d '{
    "close_date": "2025-12-31",
    "reason": "Customer chose competitor due to pricing"
  }'

# Get pipeline stats to see impact
curl http://localhost:8081/api/v1/customer-mgmt/deals/stats/pipeline
```

### Example 4: Sales Pipeline Dashboard

```bash
# Get all pipeline statistics
curl http://localhost:8081/api/v1/customer-mgmt/deals/stats/pipeline | jq .

# Get deals in proposal stage
curl http://localhost:8081/api/v1/customer-mgmt/deals/stage/proposal | jq .

# Get won deals revenue
curl http://localhost:8081/api/v1/customer-mgmt/deals/stats/won | jq .
```

**Example Output**:
```json
{
  "status": "success",
  "data": {
    "stats": {
      "lead": 12,
      "qualified": 8,
      "proposal": 5,
      "negotiation": 3,
      "closed_won": 45,
      "closed_lost": 23
    }
  }
}
```

---

## Database Schema

### Table: `customer_deals`

```sql
CREATE TABLE IF NOT EXISTS customer_deals (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    
    -- Foreign Keys
    customer_id UUID NOT NULL REFERENCES customer_mgmt_customers(id),
    assigned_to UUID NOT NULL REFERENCES identity_users(id),
    
    -- Deal Information
    name VARCHAR(200) NOT NULL,
    value BIGINT NOT NULL CHECK (value >= 0),  -- Cents
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    
    -- Pipeline Stage
    stage VARCHAR(20) NOT NULL DEFAULT 'lead' 
        CHECK (stage IN ('lead', 'qualified', 'proposal', 'negotiation', 'closed_won', 'closed_lost')),
    probability INTEGER NOT NULL DEFAULT 10 CHECK (probability >= 0 AND probability <= 100),
    
    -- Deal Source
    source VARCHAR(20) NOT NULL DEFAULT 'inbound'
        CHECK (source IN ('inbound', 'outbound', 'referral', 'partner')),
    
    -- Dates
    expected_close_date DATE,
    actual_close_date DATE,
    
    -- Metadata
    loss_reason TEXT,
    tags JSONB DEFAULT '[]'::jsonb,
    
    -- Audit Fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Indexes
    CONSTRAINT fk_deal_customer FOREIGN KEY (customer_id) REFERENCES customer_mgmt_customers(id),
    CONSTRAINT fk_deal_assigned_to FOREIGN KEY (assigned_to) REFERENCES identity_users(id)
);

-- Indexes for performance
CREATE INDEX idx_deals_customer_id ON customer_deals(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_assigned_to ON customer_deals(assigned_to) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_stage ON customer_deals(stage) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_expected_close ON customer_deals(expected_close_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_actual_close ON customer_deals(actual_close_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_deals_tags ON customer_deals USING GIN (tags) WHERE deleted_at IS NULL;

-- Trigger for updated_at
CREATE TRIGGER trg_deals_updated_at
    BEFORE UPDATE ON customer_deals
    FOR EACH ROW
    EXECUTE FUNCTION tfn_entity_updated_at();
```

### Indexes Strategy

**Performance Optimizations**:
1. **Foreign Keys**: Fast lookup by customer and sales rep
2. **Stage Index**: Quick filtering by pipeline stage
3. **Date Indexes**: Fast date range queries for forecasting
4. **GIN Index**: Efficient JSONB tag searches
5. **Partial Indexes**: Exclude soft-deleted records for better performance

---

## Testing

### Test Coverage

**24 Tests** (100% passing)

**Entity Tests** (`entity_test.go`):
-  Create deal with valid data
-  Validate required fields (customer_id, assigned_to, name, value)
-  Stage transition validation
-  Probability auto-calculation per stage
-  Win/Loss state transitions
-  Terminal stage protection

**Use Case Tests** (`usecase_test.go`):
-  Create deal business logic
-  Update basic info
-  Update value
-  Move to stage (with validation)
-  Mark as won/lost
-  List with pagination
-  Filter by stage
-  Pipeline statistics
-  Won deals statistics

**Repository Tests** (`repository_test.go`):
-  Create and retrieve
-  Update operations
-  Soft delete
-  List with filters
-  Statistics queries

### Running Tests

```bash
# All Deal tests
go test ./internal/contexts/customer-mgmt/deal/... -v

# Entity tests only
go test ./internal/contexts/customer-mgmt/deal -run TestDeal -v

# Use case tests
go test ./internal/contexts/customer-mgmt/deal -run TestUseCase -v

# Integration tests
go test ./test/integration/contexts/customer-mgmt/deal/... -v
```

### Test Examples

**Entity Test - Stage Transition**:
```go
func TestDeal_MoveToStage(t *testing.T) {
    customerID := uuidv7.New()
    assignedTo := uuidv7.New()
    deal, _ := NewDeal(customerID, assignedTo, "Test Deal", 1000000, "USD", time.Now().AddDate(0, 1, 0))

    // Move from lead to qualified
    err := deal.MoveToStage(DealStageQualified)
    assert.NoError(t, err)
    assert.Equal(t, DealStageQualified, deal.Stage)
    assert.Equal(t, 25, deal.Probability)

    // Close as won
    err = deal.MarkWon(time.Now())
    assert.NoError(t, err)
    assert.Equal(t, DealStageClosedWon, deal.Stage)
    assert.Equal(t, 100, deal.Probability)

    // Cannot move from closed stage
    err = deal.MoveToStage(DealStageNegotiation)
    assert.Error(t, err)
}
```

---

## Integration with Other Contexts

### Customer Management Context
Deal is part of Customer Management context, integrated with Customer aggregate.

**Relationships**:
- **Deal → Customer**: Many-to-one (one customer can have multiple deals)
- **Deal → User (Sales Rep)**: Many-to-one (one sales rep manages multiple deals)

**Event Publishing**:
```go
// On deal creation
eventBus.Publish(ctx, bus.TopicDealCreated, bus.NewEvent(
    "customer.deal.created",
    deal.ID,
    map[string]interface{}{
        "customer_id": deal.CustomerID,
        "assigned_to": deal.AssignedTo,
        "value":       deal.Value.Amount(),
    },
))

// On deal won
eventBus.Publish(ctx, bus.TopicDealWon, bus.NewEvent(
    "customer.deal.won",
    deal.ID,
    map[string]interface{}{
        "customer_id": deal.CustomerID,
        "value":       deal.Value.Amount(),
        "close_date":  deal.ActualCloseDate,
    },
))
```

### Identity Context
Deal references User aggregate via `assigned_to` for sales rep assignment.

### Order Management Context (Future)
When deal is won, can trigger order creation:
```go
// Subscribe to deal won events
eventBus.Subscribe(bus.TopicDealWon, func(ctx context.Context, event bus.Event) error {
    // Create order from deal
    order := NewOrderFromDeal(deal)
    return orderRepo.Create(ctx, order)
})
```

---

## Best Practices

### 1. Deal Naming Convention
Use descriptive names that identify the opportunity:
```
 Good: "Enterprise License + Support", "Cloud Migration Project"
 Bad: "Deal 1", "Customer ABC"
```

### 2. Stage Progression
Progress deals through stages sequentially when possible:
```
 Recommended: lead → qualified → proposal → negotiation → closed_won
  Allowed: lead → proposal (skip stages)
 Avoid: closed_won → negotiation (cannot move from terminal)
```

### 3. Value Updates
Update deal value as negotiation progresses:
```go
// Initial deal
value: 50000000 // $500k

// After negotiation
value: 60000000 // $600k
```

### 4. Loss Tracking
Always provide reason when marking deal as lost:
```json
{
  "close_date": "2025-12-31",
  "reason": "Lost to competitor due to pricing"
}
```

### 5. Expected Close Date
Keep expected_close_date current:
```go
// Update as deal progresses
PUT /deals/:id/basic-info
{
  "expected_close_date": "2026-02-28"  // Moved out 1 month
}
```

---

## Common Patterns

### Pattern 1: Quick Win
Deal closed immediately without going through all stages:
```bash
# Create deal
curl -X POST /api/v1/customer-mgmt/deals -d '{...}'

# Close as won immediately
curl -X POST /api/v1/customer-mgmt/deals/:id/win -d '{"close_date": "2025-12-31"}'
```

### Pattern 2: Full Pipeline
Deal progresses through all stages:
```bash
# lead → qualified → proposal → negotiation → closed_won
for stage in qualified proposal negotiation; do
  curl -X PUT /api/v1/customer-mgmt/deals/:id/stage -d "{\"stage\": \"$stage\"}"
done
curl -X POST /api/v1/customer-mgmt/deals/:id/win -d '{"close_date": "2025-12-31"}'
```

### Pattern 3: Price Negotiation
Update value during negotiation:
```bash
# Increase value after negotiation
curl -X PUT /api/v1/customer-mgmt/deals/:id/value \
  -d '{"value": 75000000, "currency": "USD"}'
```

---

## Roadmap

### Current Features ( Implemented)
-  Deal CRUD operations
-  Stage-based pipeline management
-  Automatic probability calculation
-  Win/Loss tracking
-  Pipeline statistics
-  Money handling (cents precision)
-  Soft delete support
-  24 tests (100% passing)

### Planned Features ( Q1 2026)
- [ ] Deal activities timeline (calls, emails, meetings)
- [ ] Deal notes and attachments
- [ ] Deal collaborators (multiple sales reps)
- [ ] Custom deal stages per organization
- [ ] Revenue forecasting by period
- [ ] Deal conversion analytics
- [ ] Integration with email/calendar

### Future Enhancements ( Q2 2026)
- [ ] Deal templates for common scenarios
- [ ] Automated lead scoring
- [ ] AI-powered close date prediction
- [ ] Deal health indicators
- [ ] Competitor tracking

---

## Related Documentation

- [Customer Management](customer-management.md) - Parent context
- [Order Management](order-management.md) - Integration for won deals
- [Main README](../../README.md) - Project overview
- [API Reference](../reference/api-reference.md) - Complete HTTP API
- [Testing Guide](../../test/README.md) - Testing patterns

---

**Version**: 1.0.0  
**Status**: Production-ready  
**Last Updated**: December 31, 2025  
**Maintainer**: Promenade Team
