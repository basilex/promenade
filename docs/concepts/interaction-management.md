# Interaction Management

**Complete customer interaction tracking** for calls, emails, meetings, and notes with outcome tracking and follow-up management.

---

## Overview

**Interaction aggregate** tracks all customer touchpoints - phone calls, emails, meetings, and notes - with comprehensive metadata, participant tracking, and follow-up management.

### Key Concepts

- **Interaction Types**: Call, Email, Meeting, Note
- **Directions**: Inbound (customer-initiated), Outbound (company-initiated)
- **Outcomes**: Successful, Failed, No Answer, Scheduled, Cancelled
- **JSONB Attendees**: Flexible multi-participant tracking
- **Follow-up Management**: Flag and schedule follow-up actions
- **Duration Tracking**: Automatic calculation for ended interactions

---

## Domain Model

### Interaction Aggregate

```go
type Interaction struct {
    // Identity
    ID         uuidv7.UUID
    CustomerID uuidv7.UUID  // Required: customer this interaction belongs to
    CompanyID  *uuidv7.UUID // Optional: for B2B interactions

    // Classification
    Type      InteractionType      // call, email, meeting, note
    Direction InteractionDirection // inbound, outbound
    Outcome   *InteractionOutcome  // successful, failed, no_answer, scheduled, cancelled

    // Content
    Subject     string // Required: brief description (max 255 chars)
    Description string // Optional: detailed notes

    // Participants
    CreatedBy uuidv7.UUID   // User who created/logged this interaction
    Attendees []uuidv7.UUID // JSONB array of user IDs who participated

    // Timing
    StartedAt   time.Time  // When interaction started
    EndedAt     *time.Time // When interaction ended (nil if in progress)
    DurationSec *int       // Calculated duration in seconds

    // Follow-up
    FollowUpRequired bool
    FollowUpDate     *time.Time // When follow-up is needed
    FollowUpNotes    string     // Notes about follow-up action

    // Metadata
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}
```

### Enums

**Interaction Types**:
```go
const (
    InteractionTypeCall    = "call"    // Phone call
    InteractionTypeEmail   = "email"   // Email correspondence
    InteractionTypeMeeting = "meeting" // In-person or virtual meeting
    InteractionTypeNote    = "note"    // General note/memo
)
```

**Interaction Directions**:
```go
const (
    InteractionDirectionInbound  = "inbound"  // Customer-initiated
    InteractionDirectionOutbound = "outbound" // Company-initiated
)
```

**Interaction Outcomes**:
```go
const (
    InteractionOutcomeSuccessful = "successful" // Goal achieved
    InteractionOutcomeFailed     = "failed"     // Goal not achieved
    InteractionOutcomeNoAnswer   = "no_answer"  // Call not answered
    InteractionOutcomeScheduled  = "scheduled"  // Follow-up scheduled
    InteractionOutcomeCancelled  = "cancelled"  // Cancelled before completion
)
```

---

## Business Rules

### Entity Rules

1. **Required Fields**:
   - CustomerID (every interaction must belong to a customer)
   - Type (call, email, meeting, note)
   - Direction (inbound, outbound)
   - Subject (brief description, max 255 chars)
   - CreatedBy (user who logged the interaction)
   - StartedAt (when interaction occurred)

2. **Duration Calculation**:
   - If `EndedAt` is set, `DurationSec` calculated automatically
   - Formula: `DurationSec = EndedAt.Unix() - StartedAt.Unix()`
   - Cannot manually set duration (auto-calculated)

3. **Attendees Management**:
   - Stored as JSONB array (flexible, no FK constraints)
   - Can add/remove attendees dynamically
   - Empty array `[]` if no attendees
   - Duplicate attendees prevented by entity logic

4. **Follow-up Rules**:
   - If `FollowUpRequired = true`, follow-up date recommended (not enforced)
   - Follow-up notes optional but helpful for context
   - Query pending follow-ups via `ListPendingFollowUps()`

### Factory Methods

```go
// Create new interaction
func NewInteraction(
    customerID uuidv7.UUID,
    companyID *uuidv7.UUID,
    interactionType InteractionType,
    direction InteractionDirection,
    subject string,
    description string,
    createdBy uuidv7.UUID,
    startedAt time.Time,
) (*Interaction, error)
```

**Validation**:
- Subject cannot be empty
- CreatedBy must be valid UUID
- Type must be valid enum value
- Direction must be valid enum value
- StartedAt cannot be zero time

### Entity Methods

**Content Management**:
```go
func (i *Interaction) UpdateContent(subject, description string) error
```

**Outcome Tracking**:
```go
func (i *Interaction) SetOutcome(outcome InteractionOutcome) error
```

**Time Management**:
```go
func (i *Interaction) EndInteraction(endedAt time.Time) error
// Validates: endedAt must be after StartedAt
// Calculates: DurationSec automatically
```

**Follow-up Management**:
```go
func (i *Interaction) SetFollowUp(required bool, followUpDate *time.Time, notes string) error
// Validates: if required=true and date provided, date must be in future
```

**Attendee Management**:
```go
func (i *Interaction) AddAttendee(attendeeID uuidv7.UUID) error
// Prevents: duplicate attendees

func (i *Interaction) RemoveAttendee(attendeeID uuidv7.UUID) error
// Silent: no error if attendee not found
```

---

## Use Cases

### IUseCase Interface

```go
type IUseCase interface {
    // CRUD Operations
    CreateInteraction(ctx, customerID, companyID, type, direction, subject, description, createdBy, startedAt) (*Interaction, error)
    GetInteraction(ctx context.Context, id uuidv7.UUID) (*Interaction, error)
    UpdateContent(ctx context.Context, id uuidv7.UUID, subject, description string) (*Interaction, error)
    DeleteInteraction(ctx context.Context, id uuidv7.UUID) error

    // Business Operations
    SetOutcome(ctx context.Context, id uuidv7.UUID, outcome string) (*Interaction, error)
    EndInteraction(ctx context.Context, id uuidv7.UUID, endedAt time.Time) (*Interaction, error)
    SetFollowUp(ctx, id, required, followUpDate, notes) (*Interaction, error)
    AddAttendee(ctx context.Context, id, attendeeID uuidv7.UUID) (*Interaction, error)
    RemoveAttendee(ctx context.Context, id, attendeeID uuidv7.UUID) (*Interaction, error)

    // List/Query Operations (with N+1 optimization)
    ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error)
    ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error)
    ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*Interaction, int64, error)
    ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error)
    ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*Interaction, int64, error)
}
```

---

## Repository Pattern

### N+1 Optimization

All List queries use **LEFT JOIN** to batch-load related data and prevent N+1 problem:

```go
// interactionRowWithRelations - optimized row struct
type interactionRowWithRelations struct {
    interactionRow
    CustomerName  *string `db:"customer_name"`   // From customer_mgmt_customers
    CompanyName   *string `db:"company_name"`    // From customer_mgmt_companies (NULL if not available)
    CreatedByName *string `db:"created_by_name"` // From identity_users (email field)
}

// Example: ListByCustomer with LEFT JOIN
query := `
    SELECT 
        i.*,
        c.name AS customer_name,
        NULL AS company_name,  -- Company table not in all environments yet
        u.email AS created_by_name
    FROM customer_interactions i
    LEFT JOIN customer_mgmt_customers c ON i.customer_id = c.id
    LEFT JOIN identity_users u ON i.created_by = u.id
    WHERE i.customer_id = $1 AND i.deleted_at IS NULL
    ORDER BY i.started_at DESC
    LIMIT $2 OFFSET $3
`
```

**Benefits**:
- Single query instead of N+1 queries
- Customer name available without extra lookup
- Created by user email available without extra lookup
- Future-ready for displaying names in responses

### Query Patterns

**By Customer** (most common):
```sql
WHERE i.customer_id = $1 AND i.deleted_at IS NULL
ORDER BY i.started_at DESC
```

**By Type** (e.g., all calls):
```sql
WHERE i.type = 'call' AND i.deleted_at IS NULL
ORDER BY i.started_at DESC
```

**Pending Follow-ups**:
```sql
WHERE i.follow_up_required = true 
  AND (i.follow_up_date IS NULL OR i.follow_up_date <= NOW())
  AND i.deleted_at IS NULL
ORDER BY i.follow_up_date ASC NULLS FIRST, i.started_at DESC
```

---

## Database Schema

### Table: `customer_interactions`

```sql
CREATE TABLE customer_interactions (
  id                 TEXT PRIMARY KEY,
  customer_id        TEXT NOT NULL REFERENCES customer_mgmt_customers(id),
  company_id         TEXT REFERENCES customer_mgmt_companies(id),
    
    type               interaction_type NOT NULL,
    direction          interaction_direction NOT NULL,
    outcome            interaction_outcome,
    
    subject            VARCHAR(255) NOT NULL,
    description        TEXT NOT NULL DEFAULT '',
    
    created_by         TEXT NOT NULL REFERENCES identity_users(id),
    attendees          TEXT NOT NULL DEFAULT '[]',
    
    started_at         TIMESTAMPTZ NOT NULL,
    ended_at           TIMESTAMPTZ,
    duration_sec       INTEGER,
    
    follow_up_required BOOLEAN NOT NULL DEFAULT false,
    follow_up_date     TIMESTAMPTZ,
    follow_up_notes    TEXT,
    
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at         TIMESTAMPTZ
);
```

### Enums

```sql
CREATE TYPE interaction_type AS ENUM ('call', 'email', 'meeting', 'note');
CREATE TYPE interaction_direction AS ENUM ('inbound', 'outbound');
CREATE TYPE interaction_outcome AS ENUM ('successful', 'failed', 'no_answer', 'scheduled', 'cancelled');
```

### Indexes

```sql
-- Primary key
CREATE INDEX idx_interactions_pkey ON customer_interactions(id);

-- Foreign keys (partial - only non-deleted)
CREATE INDEX idx_interactions_customer_id ON customer_interactions(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_interactions_company_id ON customer_interactions(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_interactions_created_by ON customer_interactions(created_by) WHERE deleted_at IS NULL;

-- Query optimization
CREATE INDEX idx_interactions_customer_started ON customer_interactions(customer_id, started_at DESC) WHERE deleted_at IS NULL;

-- Follow-ups
CREATE INDEX idx_interactions_follow_up ON customer_interactions(follow_up_date) 
    WHERE follow_up_required = true AND deleted_at IS NULL;

-- JSONB attendees (GIN index for array containment)
CREATE INDEX idx_interactions_attendees ON customer_interactions USING GIN(attendees);
```

### Trigger

Auto-update `updated_at` timestamp:

```sql
CREATE TRIGGER trigger_interactions_updated_at
    BEFORE UPDATE ON customer_interactions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

---

## API Examples

### Create Interaction

**Request**:
```http
POST /api/v1/customer-mgmt/interactions
Content-Type: application/json

{
  "customer_id": "01JGEF1234567890ABCDEFGHIJ",
  "type": "call",
  "direction": "outbound",
  "subject": "Follow-up call about product demo",
  "description": "Called customer to discuss demo feedback and next steps",
  "started_at": "2025-12-28T10:30:00Z"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGEF9876543210ZYXWVUTSRQ",
    "customer_id": "01JGEF1234567890ABCDEFGHIJ",
    "type": "call",
    "direction": "outbound",
    "subject": "Follow-up call about product demo",
    "description": "Called customer to discuss demo feedback and next steps",
    "created_by": "01JGEF5555555555555555555",
    "attendees": [],
    "started_at": "2025-12-28T10:30:00Z",
    "created_at": "2025-12-28T10:30:05Z",
    "updated_at": "2025-12-28T10:30:05Z"
  }
}
```

### End Interaction

**Request**:
```http
PUT /api/v1/customer-mgmt/interactions/01JGEF9876543210ZYXWVUTSRQ/end
Content-Type: application/json

{
  "ended_at": "2025-12-28T10:45:00Z"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGEF9876543210ZYXWVUTSRQ",
    "started_at": "2025-12-28T10:30:00Z",
    "ended_at": "2025-12-28T10:45:00Z",
    "duration_sec": 900
  }
}
```

### Set Follow-up

**Request**:
```http
PUT /api/v1/customer-mgmt/interactions/01JGEF9876543210ZYXWVUTSRQ/follow-up
Content-Type: application/json

{
  "follow_up_required": true,
  "follow_up_date": "2025-12-30T09:00:00Z",
  "follow_up_notes": "Customer requested additional pricing information"
}
```

### Add Attendee

**Request**:
```http
POST /api/v1/customer-mgmt/interactions/01JGEF9876543210ZYXWVUTSRQ/attendees
Content-Type: application/json

{
  "attendee_id": "01JGEF7777777777777777777"
}
```

### List by Customer

**Request**:
```http
GET /api/v1/customer-mgmt/interactions/customer/01JGEF1234567890ABCDEFGHIJ?page=1&page_size=20
```

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGEF9876543210ZYXWVUTSRQ",
      "customer_id": "01JGEF1234567890ABCDEFGHIJ",
      "type": "call",
      "direction": "outbound",
      "outcome": "successful",
      "subject": "Follow-up call about product demo",
      "started_at": "2025-12-28T10:30:00Z",
      "ended_at": "2025-12-28T10:45:00Z",
      "duration_sec": 900
    }
  ],
  "pagination": {
    "total": 150,
    "page": 1,
    "page_size": 20,
    "total_pages": 8
  }
}
```

---

## Performance Optimization

### Benchmark Results

**Hardware**: Apple M4 Max  
**Test DB**: PostgreSQL 16 on port 5433

```
BenchmarkCreateInteraction-16      9619    341426 ns/op    5270 B/op    99 allocs/op
BenchmarkListByCustomer-16         5134    602662 ns/op   58040 B/op   897 allocs/op
```

**Analysis**:
- **Create**: ~341μs per interaction (single INSERT with JSONB)
- **List**: ~603μs for 20 interactions from 50 total (with LEFT JOIN, pagination)
- Memory efficient: 5.3KB for create, 58KB for list with 20 rows

### N+1 Prevention

Without optimization (N+1 problem):
```
1 query:  SELECT * FROM customer_interactions WHERE customer_id = X
50 queries: SELECT name FROM customer_mgmt_customers WHERE id = Y (for each interaction)
50 queries: SELECT email FROM identity_users WHERE id = Z (for each interaction)
= 101 queries total
```

With LEFT JOIN optimization:
```
1 query: SELECT i.*, c.name, u.email FROM customer_interactions i
         LEFT JOIN customer_mgmt_customers c ON i.customer_id = c.id
         LEFT JOIN identity_users u ON i.created_by = u.id
         WHERE i.customer_id = X
= 1 query total (100x improvement!)
```

---

## Testing

### Test Coverage

**Total**: 74 tests, 100% passing

**Unit Tests** (69 tests):
- `entity_test.go`: Factory methods, validation, business rules
- `usecase_test.go`: Business logic, error handling

**Integration Tests** (5 tests):
- `repository_test.go`: CRUD operations, List queries with real database

**Benchmark Tests** (2 benchmarks):
- `repository_bench_test.go`: Create and List performance measurement

### Running Tests

```bash
# All interaction tests
go test ./internal/contexts/customer-mgmt/interaction/... -v

# Integration tests only
go test ./test/integration/contexts/customer-mgmt/interaction -v

# Benchmarks
go test -bench=. -benchmem ./test/benchmark/contexts/customer-mgmt/interaction
```

---

## Use Case Scenarios

### Scenario 1: Logging a Sales Call

**Workflow**:
1. Sales rep calls customer to discuss product interest
2. Create interaction with type=call, direction=outbound
3. During call, add outcome=successful
4. End interaction with actual end time (auto-calculates duration)
5. Set follow-up for demo scheduling

**API Flow**:
```
POST /interactions → Create call interaction
PUT  /interactions/:id/outcome → Set outcome=successful
PUT  /interactions/:id/end → End call, calculate duration
PUT  /interactions/:id/follow-up → Schedule demo follow-up
```

### Scenario 2: Email Thread Tracking

**Workflow**:
1. Customer sends inquiry email (direction=inbound)
2. Create interaction with type=email
3. Add sales rep as attendee who will respond
4. Log outcome=successful after sending reply
5. If needs follow-up, flag and schedule

**API Flow**:
```
POST /interactions → Create email interaction
POST /interactions/:id/attendees → Add responding sales rep
PUT  /interactions/:id/outcome → Mark as successful reply
PUT  /interactions/:id/follow-up → Flag if needs follow-up
```

### Scenario 3: Meeting with Multiple Participants

**Workflow**:
1. Schedule meeting with customer and internal team
2. Create interaction with type=meeting
3. Add all internal attendees (sales, solutions architect, manager)
4. After meeting, update outcome and notes
5. Set follow-up for proposal delivery

**API Flow**:
```
POST /interactions → Create meeting interaction
POST /interactions/:id/attendees → Add attendee 1
POST /interactions/:id/attendees → Add attendee 2
POST /interactions/:id/attendees → Add attendee 3
PUT  /interactions/:id/end → End meeting
PUT  /interactions/:id/outcome → Set outcome=successful
PUT  /interactions/:id/follow-up → Schedule proposal follow-up
```

---

## Future Enhancements

### Planned Features

- [ ] **Email Integration**: Auto-create interactions from email system
- [ ] **Calendar Sync**: Import meetings from calendar systems
- [ ] **Call Recording**: Store call recordings in S3/cloud storage
- [ ] **Transcription**: Auto-transcribe calls and meetings
- [ ] **Sentiment Analysis**: Analyze customer sentiment from interactions
- [ ] **Activity Feed**: Real-time feed of all customer interactions
- [ ] **Interaction Templates**: Pre-filled templates for common interaction types
- [ ] **Reminders**: Auto-reminders for pending follow-ups

### Not Planned

- **Real-time Communication**: Use dedicated communication platform
- **Video Conferencing**: Integrate with existing tools (Zoom, Teams)
- **CRM Workflows**: Complex automation (use dedicated workflow engine)

---

## Related Documentation

- [Customer Management Guide](customer-management.md) - Customer aggregate (interactions belong to customers)
- [Company Management Guide](company-management.md) - Company aggregate (B2B interactions)
- [Deal Management Guide](deal-management.md) - Deal aggregate (interactions related to deals)
- [Documentation Index](../INDEX.md) - All documentation
- [Main README](../../README.md) - Project overview

---

**Version**: 0.1.0  
**Status**: Production-ready  
**Test Coverage**: 74 tests, 100% passing  
**Maintainer**: Promenade Team
