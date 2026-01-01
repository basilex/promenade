# Interaction Management

**Complete customer interaction tracking** for calls, emails, meetings, and notes with outcome tracking and follow-up management.

---

## Overview

**Interaction aggregate** tracks all customer touchpoints - phone calls, emails, meetings, and notes - with comprehensive metadata, participant tracking, and follow-up management.

**Status**: ✅ Production-ready (December 2025)  
**Endpoints**: 14 HTTP routes  
**Database**: 1 table with JSONB attendees

### Key Features

- **Interaction Types**: Call, Email, Meeting, Note, SMS, Chat
- **Directions**: Inbound (customer-initiated), Outbound (company-initiated)
- **Outcomes**: Successful, Failed, No Answer, Voicemail, Busy, Scheduled, Not Interested
- **JSONB Attendees**: Multi-participant tracking with flexible arrays
- **Follow-up Management**: Flag and schedule follow-up actions
- **Duration Tracking**: Automatic calculation for ended interactions
- **N+1 Optimization**: LEFT JOIN queries prevent performance issues

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
    Type      InteractionType      // call, email, meeting, note, sms, chat
    Direction InteractionDirection // inbound, outbound
    Outcome   *InteractionOutcome  // successful, failed, no_answer, etc.

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

---

## Business Rules

### Required Fields
- **CustomerID**: Every interaction must belong to a customer
- **Type**: call, email, meeting, note, sms, or chat
- **Direction**: inbound or outbound
- **Subject**: Brief description (max 255 chars)
- **CreatedBy**: User who logged the interaction
- **StartedAt**: When interaction occurred

### Duration Calculation
- If `EndedAt` is set, `DurationSec` calculated automatically
- Formula: `DurationSec = EndedAt.Unix() - StartedAt.Unix()`
- Cannot manually set duration (auto-calculated)

### Attendees Management
- Stored as JSONB array (flexible, no FK constraints)
- Can add/remove attendees dynamically
- Empty array `[]` if no attendees
- Duplicate attendees prevented by entity logic

### Follow-up Rules
- If `FollowUpRequired = true`, follow-up date recommended
- Follow-up notes optional but helpful for context
- Query pending follow-ups via `ListPendingFollowUps()`

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| **CRUD** |
| POST | `/api/v1/customer-mgmt/interactions` | Create interaction |
| GET | `/api/v1/customer-mgmt/interactions/:id` | Get by ID |
| PUT | `/api/v1/customer-mgmt/interactions/:id/content` | Update content |
| DELETE | `/api/v1/customer-mgmt/interactions/:id` | Soft delete |
| **Business Operations** |
| PUT | `/api/v1/customer-mgmt/interactions/:id/outcome` | Set outcome |
| PUT | `/api/v1/customer-mgmt/interactions/:id/end` | End interaction |
| PUT | `/api/v1/customer-mgmt/interactions/:id/follow-up` | Set follow-up |
| POST | `/api/v1/customer-mgmt/interactions/:id/attendees` | Add attendee |
| DELETE | `/api/v1/customer-mgmt/interactions/:id/attendees/:attendee_id` | Remove attendee |
| **List/Query** |
| GET | `/api/v1/customer-mgmt/interactions/customer/:id` | List by customer |
| GET | `/api/v1/customer-mgmt/interactions/company/:id` | List by company |
| GET | `/api/v1/customer-mgmt/interactions/type/:type` | List by type |
| GET | `/api/v1/customer-mgmt/interactions/created-by/:id` | List by creator |
| GET | `/api/v1/customer-mgmt/interactions/follow-ups` | Pending follow-ups |

---

## Examples

### Create Phone Call

```bash
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01JGEF1234567890ABCDEFGHIJ",
    "type": "call",
    "direction": "outbound",
    "subject": "Follow-up call about product demo",
    "description": "Called customer to discuss demo feedback",
    "started_at": "2025-12-28T10:30:00Z"
  }'
```

### End Interaction

```bash
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/end \
  -H "Content-Type: application/json" \
  -d '{
    "ended_at": "2025-12-28T10:45:00Z"
  }'

# Response includes auto-calculated duration_sec: 900
```

### Set Follow-up

```bash
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/follow-up \
  -H "Content-Type: application/json" \
  -d '{
    "follow_up_required": true,
    "follow_up_date": "2025-12-30T09:00:00Z",
    "follow_up_notes": "Customer requested pricing information"
  }'
```

### Add Meeting Attendees

```bash
# Add attendee 1
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/attendees \
  -H "Content-Type: application/json" \
  -d '{"attendee_id": "01JGEF7777777777777777777"}'

# Add attendee 2
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/attendees \
  -H "Content-Type: application/json" \
  -d '{"attendee_id": "01JGEF8888888888888888888"}'
```

### List Customer Interactions

```bash
curl http://localhost:8081/api/v1/customer-mgmt/interactions/customer/{customer_id}?page=1&page_size=20
```

---

## Performance Optimization

### N+1 Prevention

All List queries use **LEFT JOIN** to batch-load related data:

```sql
SELECT 
    i.*,
    c.name AS customer_name,
    comp.name AS company_name,
    u.email AS created_by_name
FROM customer_interactions i
LEFT JOIN customer_mgmt_customers c ON i.customer_id = c.id
LEFT JOIN customer_mgmt_companies comp ON i.company_id = comp.id
LEFT JOIN identity_users u ON i.created_by = u.id
WHERE i.customer_id = $1 AND i.deleted_at IS NULL
ORDER BY i.started_at DESC
LIMIT $2 OFFSET $3
```

**Benefits**:
- Single query instead of N+1 queries
- Customer/company names available without extra lookup
- 100x performance improvement for listing interactions

### Benchmark Results

**Hardware**: Apple M4 Max, PostgreSQL 16

```
BenchmarkCreateInteraction-16      9619    341426 ns/op    5270 B/op    99 allocs/op
BenchmarkListByCustomer-16         5134    602662 ns/op   58040 B/op   897 allocs/op
```

- **Create**: ~341μs per interaction
- **List**: ~603μs for 20 interactions with LEFT JOIN
- Memory efficient: 5.3KB create, 58KB list with 20 rows

---

## Use Case Scenarios

### Scenario 1: Logging a Sales Call

1. Sales rep calls customer to discuss product interest
2. Create interaction (type=call, direction=outbound)
3. During call, set outcome=successful
4. End interaction (auto-calculates duration)
5. Set follow-up for demo scheduling

### Scenario 2: Email Thread Tracking

1. Customer sends inquiry email (direction=inbound)
2. Create interaction (type=email)
3. Add sales rep as attendee who will respond
4. Log outcome=successful after sending reply
5. Flag for follow-up if needed

### Scenario 3: Meeting with Multiple Participants

1. Schedule meeting with customer and internal team
2. Create interaction (type=meeting)
3. Add all internal attendees (sales, solutions architect, manager)
4. After meeting, update outcome and notes
5. Set follow-up for proposal delivery

---

## Testing

**Test Coverage**: 74 tests, 100% passing

- **Unit Tests** (69): Entity validation, business logic
- **Integration Tests** (5): CRUD operations, List queries with real database
- **Benchmark Tests** (2): Create and List performance measurement

```bash
# Run interaction tests
go test ./internal/contexts/customer-mgmt/interaction/... -v

# Integration tests
go test ./test/integration/contexts/customer-mgmt/interaction -v

# Benchmarks
go test -bench=. ./test/benchmark/contexts/customer-mgmt/interaction
```

---

## Related Documentation

- [Customer Management](/concepts/customer-management) - Parent context
- [Company Management](/concepts/company-management) - B2B interactions
- [Deal Management](/concepts/deal-management) - Deal-related interactions
- [Analytics](/concepts/analytics) - Interaction insights
- [API Reference](/reference/api-reference) - Complete API docs

---

**Version**: 1.0.0  
**Status**: Production-ready  
**Test Coverage**: 74 tests, 100% passing  
**Maintainer**: Promenade Team
