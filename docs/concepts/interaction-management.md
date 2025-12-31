# Interaction Management Guide

**Complete customer interaction tracking** for Promenade Platform - log calls, emails, meetings, notes, and manage follow-ups.

---

## Overview

The **Interaction aggregate** provides comprehensive customer communication logging within the **Customer Management Context**. It enables sales teams and customer service reps to track all touchpoints with customers and companies, manage follow-ups, and maintain a complete interaction history.

**Status**: ✅ **Production-ready** (Phase 4)

---

## Key Features

### 1. Multiple Interaction Types

Support for 6 types of customer communications:

- **Call** - Phone calls (inbound/outbound)
- **Email** - Email correspondence  
- **Meeting** - In-person or virtual meetings
- **Note** - General notes and memos
- **SMS** - SMS/text messages
- **Chat** - Live chat or messaging

### 2. Direction Tracking

Track communication direction:

- **Inbound** - From customer to company
- **Outbound** - From company to customer

### 3. Outcome Recording

Capture interaction results:

- **successful** - Successful interaction
- **no_answer** - No answer (calls)
- **voicemail** - Left voicemail
- **busy** - Line busy
- **scheduled** - Meeting scheduled
- **not_interested** - Customer not interested

### 4. Duration Calculation

Automatic duration tracking:

- Record start time when interaction begins
- End interaction to calculate duration
- Duration stored in seconds for analytics

### 5. Follow-up Management

Built-in follow-up workflow:

- Mark interactions requiring follow-up
- Set follow-up date
- Add follow-up notes
- Query pending follow-ups

### 6. Attendee Tracking

For meetings, track participants:

- Add multiple attendees by user ID
- Remove attendees
- Store as JSONB array in database

### 7. B2B Support

Link interactions to companies:

- Optional company_id field
- Track interactions at company level
- Support for both B2C and B2B workflows

---

## Domain Model

### Interaction Aggregate

```go
type Interaction struct {
    aggregate.BaseAggregate

    // Identity
    ID         uuidv7.UUID
    CustomerID uuidv7.UUID  // Required: link to customer
    CompanyID  *uuidv7.UUID // Optional: link to company (B2B)

    // Classification
    Type      InteractionType      // call, email, meeting, note, sms, chat
    Direction InteractionDirection // inbound, outbound
    Outcome   *InteractionOutcome  // successful, no_answer, etc.

    // Content
    Subject     string // Subject/title
    Description string // Detailed notes

    // Participants
    CreatedBy UUID          // User who logged interaction
    Attendees []uuidv7.UUID // Meeting participants (JSONB)

    // Timing
    StartedAt   time.Time  // When interaction started
    EndedAt     *time.Time // When interaction ended
    DurationSec *int       // Duration in seconds

    // Follow-up
    FollowUpRequired bool
    FollowUpDate     *time.Time
    FollowUpNotes    string

    // Timestamps
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time // Soft delete
}
```

### Enumerations

**InteractionType**:
```go
const (
    InteractionTypeCall    = "call"
    InteractionTypeEmail   = "email"
    InteractionTypeMeeting = "meeting"
    InteractionTypeNote    = "note"
    InteractionTypeSMS     = "sms"
    InteractionTypeChat    = "chat"
)
```

**InteractionDirection**:
```go
const (
    InteractionDirectionInbound  = "inbound"
    InteractionDirectionOutbound = "outbound"
)
```

**InteractionOutcome**:
```go
const (
    InteractionOutcomeSuccessful    = "successful"
    InteractionOutcomeNoAnswer      = "no_answer"
    InteractionOutcomeVoicemail     = "voicemail"
    InteractionOutcomeBusy          = "busy"
    InteractionOutcomeScheduled     = "scheduled"
    InteractionOutcomeNotInterested = "not_interested"
)
```

---

## Business Rules

### 1. Interaction Creation

- **Customer ID required** - must link to existing customer
- **Company ID optional** - for B2B interactions
- **Type and direction required** - must be valid enum values
- **Subject and description required** - cannot be empty
- **Created by required** - must be valid user ID
- **Started at required** - when interaction began

### 2. Duration Calculation

- **Auto-calculated** - when `EndInteraction()` called
- **Formula**: `duration_sec = ended_at - started_at`
- **Cannot end before start** - validation enforced
- **Cannot end twice** - returns `ErrInteractionAlreadyEnded`

### 3. Follow-up Requirements

- If `follow_up_required = true`, `follow_up_date` must be set
- Follow-up date can be null if not required
- Follow-up notes are optional

### 4. Attendee Management

- Attendees stored as JSONB array of UUIDs
- Can add/remove attendees dynamically
- Primarily for meetings, but available for all types
- No duplicate check (implementation choice)

### 5. Soft Delete

- All deletions are soft deletes (`deleted_at IS NOT NULL`)
- Deleted interactions excluded from queries
- Preserves interaction history for audit

---

## API Endpoints

### Create Interaction

**POST** `/api/v1/customer-mgmt/interactions`

```json
{
  "customer_id": "01JGABC...",
  "company_id": "01JGXYZ...",
  "type": "call",
  "direction": "outbound",
  "subject": "Follow-up call regarding proposal",
  "description": "Discussed pricing options. Customer interested in enterprise plan.",
  "created_by": "01JGUSER...",
  "started_at": "2025-12-30T14:30:00Z"
}
```

**Response** (201 Created):
```json
{
  "status": "success",
  "data": {
    "id": "01JGINT123...",
    "customer_id": "01JGABC...",
    "company_id": "01JGXYZ...",
    "type": "call",
    "direction": "outbound",
    "outcome": null,
    "subject": "Follow-up call regarding proposal",
    "description": "Discussed pricing options...",
    "created_by": "01JGUSER...",
    "attendees": [],
    "started_at": "2025-12-30T14:30:00Z",
    "ended_at": null,
    "duration_sec": null,
    "follow_up_required": false,
    "follow_up_date": null,
    "follow_up_notes": "",
    "created_at": "2025-12-30T14:30:05Z",
    "updated_at": "2025-12-30T14:30:05Z"
  }
}
```

---

### Get Interaction by ID

**GET** `/api/v1/customer-mgmt/interactions/{id}`

**Response** (200 OK):
```json
{
  "status": "success",
  "data": { ...interaction }
}
```

**Error** (404 Not Found):
```json
{
  "status": "error",
  "error": {
    "code": "INTERACTION_NOT_FOUND",
    "message": "interaction not found"
  }
}
```

---

### List Interactions by Customer

**GET** `/api/v1/customer-mgmt/interactions/customer/{customer_id}?page=1&page_size=20`

**Response** (200 OK):
```json
{
  "status": "success",
  "data": [
    { ...interaction1 },
    { ...interaction2 }
  ],
  "pagination": {
    "total": 45,
    "page": 1,
    "page_size": 20,
    "total_pages": 3
  }
}
```

---

### List Interactions by Company

**GET** `/api/v1/customer-mgmt/interactions/company/{company_id}?page=1&page_size=20`

Returns all interactions for a specific company.

---

### List Interactions by Type

**GET** `/api/v1/customer-mgmt/interactions/type/{type}?page=1&page_size=20`

Filter interactions by type (call, email, meeting, note, sms, chat).

---

### List Pending Follow-ups

**GET** `/api/v1/customer-mgmt/interactions/follow-ups/pending?page=1&page_size=20`

Returns interactions with:
- `follow_up_required = true`
- `follow_up_date <= NOW()` or `follow_up_date IS NULL`

Ordered by `follow_up_date ASC NULLS FIRST`.

**Response**:
```json
{
  "status": "success",
  "data": [
    {
      "id": "01JGINT...",
      "customer_id": "01JGCUST...",
      "subject": "Proposal discussion",
      "follow_up_required": true,
      "follow_up_date": "2025-12-29T10:00:00Z",
      "follow_up_notes": "Call back to discuss contract terms",
      ...
    }
  ],
  "pagination": { ... }
}
```

---

### Update Content

**PUT** `/api/v1/customer-mgmt/interactions/{id}/content`

```json
{
  "subject": "Updated subject",
  "description": "Updated description with more details"
}
```

---

### Set Outcome

**PUT** `/api/v1/customer-mgmt/interactions/{id}/outcome`

```json
{
  "outcome": "successful"
}
```

Valid outcomes: `successful`, `no_answer`, `voicemail`, `busy`, `scheduled`, `not_interested`

---

### End Interaction

**POST** `/api/v1/customer-mgmt/interactions/{id}/end`

```json
{
  "ended_at": "2025-12-30T15:00:00Z"
}
```

**Response**:
```json
{
  "status": "success",
  "data": {
    "id": "01JGINT...",
    ...
    "started_at": "2025-12-30T14:30:00Z",
    "ended_at": "2025-12-30T15:00:00Z",
    "duration_sec": 1800
  }
}
```

**Business Logic**:
- Calculates `duration_sec = ended_at - started_at`
- Cannot end before start (validation)
- Cannot end twice (`ErrInteractionAlreadyEnded`)

---

### Set Follow-up

**PUT** `/api/v1/customer-mgmt/interactions/{id}/follow-up`

```json
{
  "required": true,
  "follow_up_date": "2025-12-31T10:00:00Z",
  "notes": "Call back to finalize contract"
}
```

To disable follow-up:
```json
{
  "required": false
}
```

---

### Add Attendee

**POST** `/api/v1/customer-mgmt/interactions/{id}/attendees`

```json
{
  "attendee_id": "01JGUSER..."
}
```

Adds a user to the attendees list (for meetings).

---

### Remove Attendee

**DELETE** `/api/v1/customer-mgmt/interactions/{id}/attendees/{attendee_id}`

Removes a user from the attendees list.

---

### Delete Interaction

**DELETE** `/api/v1/customer-mgmt/interactions/{id}`

**Response** (204 No Content)

Soft deletes the interaction.

---

## Use Cases

### 1. Log Sales Call

```bash
# Create outbound call
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01JGCUST...",
    "type": "call",
    "direction": "outbound",
    "subject": "Follow-up call",
    "description": "Discussed pricing and timeline",
    "created_by": "01JGUSER...",
    "started_at": "2025-12-30T10:00:00Z"
  }'

# End call
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/end \
  -H "Content-Type: application/json" \
  -d '{
    "ended_at": "2025-12-30T10:15:00Z"
  }'

# Set outcome
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/outcome \
  -H "Content-Type: application/json" \
  -d '{
    "outcome": "successful"
  }'
```

---

### 2. Schedule Meeting with Follow-up

```bash
# Create meeting
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01JGCUST...",
    "company_id": "01JGCOMP...",
    "type": "meeting",
    "direction": "inbound",
    "subject": "Product demo",
    "description": "Demonstrated key features",
    "created_by": "01JGUSER...",
    "started_at": "2025-12-30T14:00:00Z"
  }'

# Add attendees
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/attendees \
  -H "Content-Type: application/json" \
  -d '{"attendee_id": "01JGUSER1..."}'

curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/attendees \
  -H "Content-Type: application/json" \
  -d '{"attendee_id": "01JGUSER2..."}'

# Set follow-up
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/follow-up \
  -H "Content-Type: application/json" \
  -d '{
    "required": true,
    "follow_up_date": "2026-01-03T10:00:00Z",
    "notes": "Send contract proposal"
  }'
```

---

### 3. Track Email Thread

```bash
# Log initial email
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01JGCUST...",
    "type": "email",
    "direction": "outbound",
    "subject": "Proposal for enterprise plan",
    "description": "Sent detailed proposal with pricing",
    "created_by": "01JGUSER...",
    "started_at": "2025-12-29T09:00:00Z"
  }'

# Log customer reply
curl -X POST http://localhost:8081/api/v1/customer-mgmt/interactions \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "01JGCUST...",
    "type": "email",
    "direction": "inbound",
    "subject": "Re: Proposal for enterprise plan",
    "description": "Customer requested clarification on pricing",
    "created_by": "01JGUSER...",
    "started_at": "2025-12-29T14:30:00Z"
  }'
```

---

### 4. View Customer Interaction History

```bash
# Get all interactions for customer
curl http://localhost:8081/api/v1/customer-mgmt/interactions/customer/{customer_id}?page=1&page_size=50

# Filter by type
curl http://localhost:8081/api/v1/customer-mgmt/interactions/type/call?page=1&page_size=20
```

---

### 5. Manage Follow-ups

```bash
# Get pending follow-ups
curl http://localhost:8081/api/v1/customer-mgmt/interactions/follow-ups/pending

# After completing follow-up, disable it
curl -X PUT http://localhost:8081/api/v1/customer-mgmt/interactions/{id}/follow-up \
  -H "Content-Type: application/json" \
  -d '{
    "required": false
  }'
```

---

## Database Schema

### Table: customer_interactions

```sql
CREATE TABLE customer_interactions (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    customer_id UUID NOT NULL,
    company_id UUID,
    type interaction_type NOT NULL,
    direction interaction_direction NOT NULL,
    outcome interaction_outcome,
    subject VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    created_by UUID NOT NULL,
    attendees JSONB DEFAULT '[]',
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    duration_sec INTEGER,
    follow_up_required BOOLEAN NOT NULL DEFAULT false,
    follow_up_date TIMESTAMPTZ,
    follow_up_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_interactions_customer
        FOREIGN KEY (customer_id)
        REFERENCES customer_mgmt_customers(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_interactions_company
        FOREIGN KEY (company_id)
        REFERENCES customer_companies(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_interactions_created_by
        FOREIGN KEY (created_by)
        REFERENCES identity_users(id)
        ON DELETE RESTRICT
);
```

### Indexes

```sql
-- Customer lookups (most common)
CREATE INDEX idx_interactions_customer_id 
    ON customer_interactions(customer_id) 
    WHERE deleted_at IS NULL;

-- Company lookups
CREATE INDEX idx_interactions_company_id 
    ON customer_interactions(company_id) 
    WHERE deleted_at IS NULL;

-- Type filtering
CREATE INDEX idx_interactions_type 
    ON customer_interactions(type) 
    WHERE deleted_at IS NULL;

-- User activity
CREATE INDEX idx_interactions_created_by 
    ON customer_interactions(created_by) 
    WHERE deleted_at IS NULL;

-- Follow-up management (critical)
CREATE INDEX idx_interactions_follow_up 
    ON customer_interactions(follow_up_date) 
    WHERE follow_up_required = true AND deleted_at IS NULL;

-- Timeline queries
CREATE INDEX idx_interactions_started_at 
    ON customer_interactions(started_at DESC) 
    WHERE deleted_at IS NULL;

-- Composite: customer timeline
CREATE INDEX idx_interactions_customer_started 
    ON customer_interactions(customer_id, started_at DESC) 
    WHERE deleted_at IS NULL;

-- JSONB: attendee searches
CREATE INDEX idx_interactions_attendees 
    ON customer_interactions USING gin(attendees);
```

---

## Integration Points

### With Customer Aggregate

- **customer_id** (required) - every interaction linked to customer
- **Cascade delete** - deleting customer removes interactions
- **Customer timeline** - show interaction history on customer detail

### With Company Aggregate

- **company_id** (optional) - B2B interactions linked to company
- **Set null on delete** - deleting company removes link, keeps interaction
- **Company activity log** - show interactions per company

### With Identity Context

- **created_by** (required) - track who logged interaction
- **Attendees** - reference identity_users for meeting participants
- **Restrict delete** - cannot delete user if they created interactions
- **Activity tracking** - show user's interaction history

### With Deal Aggregate

- **Interaction → Deal** - track interactions related to deals
- **Deal notes** - interactions can provide context for deal progression
- **Follow-up workflow** - interactions trigger deal follow-ups

---

## Testing

### Unit Tests

**entity_test.go** (17 test cases):

```bash
go test ./internal/contexts/customer-mgmt/interaction -v
```

**Coverage**:
- NewInteraction constructor (6 tests)
- UpdateContent (3 tests)
- SetOutcome (2 tests)
- EndInteraction (3 tests)
- SetFollowUp (3 tests)
- AddAttendee/RemoveAttendee (4 tests)
- Delete (1 test)
- Validate (10 tests)

---

## Error Handling

### Domain Errors

```go
var (
    ErrInteractionNotFound      = errors.New("interaction not found")
    ErrInteractionAlreadyExists = errors.New("interaction already exists")
    ErrInvalidInteractionType   = errors.New("invalid interaction type")
    ErrInvalidDirection         = errors.New("invalid direction")
    ErrInvalidOutcome           = errors.New("invalid outcome")
    ErrInteractionAlreadyEnded  = errors.New("interaction already ended")
)
```

### HTTP Error Codes

| Code | Error                       | Reason                           |
|------|-----------------------------|----------------------------------|
| 400  | VALIDATION_ERROR            | Invalid request body             |
| 400  | INVALID_CUSTOMER_ID         | Invalid customer UUID            |
| 400  | INVALID_INTERACTION_TYPE    | Invalid type enum                |
| 400  | INVALID_DIRECTION           | Invalid direction enum           |
| 400  | INVALID_OUTCOME             | Invalid outcome enum             |
| 404  | INTERACTION_NOT_FOUND       | Interaction doesn't exist        |
| 409  | INTERACTION_ALREADY_ENDED   | Cannot end twice                 |
| 500  | INTERNAL_ERROR              | Database or system error         |

---

## Performance Considerations

### Query Optimization

1. **Customer Timeline** - Composite index `(customer_id, started_at DESC)`
2. **Follow-up Query** - Partial index on `follow_up_date WHERE follow_up_required = true`
3. **Type Filtering** - Index on `type` column
4. **Pagination** - All list endpoints support `page` and `page_size`

### Database Load

- **JSONB Attendees** - GIN index for attendee searches
- **Soft Delete Filter** - All indexes include `WHERE deleted_at IS NULL`
- **Cascading Deletes** - Customer deletion removes interactions (performance impact for large datasets)

---

## Future Enhancements

### Planned Features

- [ ] **Interaction Templates** - Pre-defined templates for common interactions
- [ ] **Email Integration** - Auto-import emails from mailbox
- [ ] **Calendar Integration** - Sync meetings with calendar
- [ ] **Attachments** - Upload files/documents to interactions
- [ ] **Tags** - Tag interactions for categorization
- [ ] **Sentiment Analysis** - AI-powered sentiment detection
- [ ] **Activity Reports** - Sales rep activity dashboards
- [ ] **Interaction Scoring** - Quality scoring for interactions
- [ ] **Bulk Import** - CSV import for historical data
- [ ] **Export/Archive** - Export interactions for compliance

---

## Related Documentation

- [Main README](../../README.md)
- [Documentation Index](../INDEX.md)
- [Customer Management Guide](customer-management.md)
- [Company Management Guide](company-management.md)
- [Deal Management Guide](deal-management.md)
- [Customer Management Context](../../internal/contexts/customer-mgmt/README.md)

---

**Version**: 0.1.0  
**Status**: Production-ready  
**Test Coverage**: 17 unit tests  
**Maintainer**: Promenade Team  
**Last Updated**: 2025-12-30
