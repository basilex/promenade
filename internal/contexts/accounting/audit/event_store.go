package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// EventType represents the type of journal entry event
type EventType string

const (
	EventEntryCreated       EventType = "entry_created"
	EventLineAdded          EventType = "line_added"
	EventLineRemoved        EventType = "line_removed"
	EventLineUpdated        EventType = "line_updated"
	EventEntryPosted        EventType = "entry_posted"
	EventEntryReversed      EventType = "entry_reversed"
	EventDescriptionUpdated EventType = "description_updated"
)

// JournalEntryEvent represents an event in the journal entry event store
type JournalEntryEvent struct {
	ID             uuidv7.UUID
	JournalEntryID uuidv7.UUID
	EventType      EventType
	EventVersion   int
	EventData      map[string]interface{}
	CausationID    *uuidv7.UUID
	CorrelationID  *uuidv7.UUID
	OrganizationID uuidv7.UUID
	UserID         uuidv7.UUID
	Timestamp      time.Time
	Metadata       map[string]interface{}
}

// EventStore provides event sourcing for journal entries
type EventStore struct {
	db *sqlx.DB
}

// NewEventStore creates a new event store
func NewEventStore(db *sqlx.DB) *EventStore {
	return &EventStore{db: db}
}

// AppendEvent appends an event to the event store
func (s *EventStore) AppendEvent(ctx context.Context, event JournalEntryEvent) error {
	if event.ID == uuidv7.Nil {
		event.ID = uuidv7.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.EventVersion == 0 {
		// Get next version
		version, err := s.getNextVersion(ctx, event.JournalEntryID)
		if err != nil {
			return err
		}
		event.EventVersion = version
	}

	eventDataJSON, err := json.Marshal(event.EventData)
	if err != nil {
		return err
	}

	var metadataJSON *string
	if event.Metadata != nil {
		data, err := json.Marshal(event.Metadata)
		if err != nil {
			return err
		}
		str := string(data)
		metadataJSON = &str
	}

	var causationIDStr, correlationIDStr *string
	if event.CausationID != nil {
		str := event.CausationID.String()
		causationIDStr = &str
	}
	if event.CorrelationID != nil {
		str := event.CorrelationID.String()
		correlationIDStr = &str
	}

	query := `INSERT INTO accounting_journal_entry_events 
		(id, journal_entry_id, event_type, event_version, event_data, 
		 causation_id, correlation_id, organization_id, user_id, timestamp, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	executor := s.getExecutor(ctx)
	_, err = executor.ExecContext(ctx, query,
		event.ID.String(),
		event.JournalEntryID.String(),
		string(event.EventType),
		event.EventVersion,
		string(eventDataJSON),
		causationIDStr,
		correlationIDStr,
		event.OrganizationID.String(),
		event.UserID.String(),
		event.Timestamp,
		metadataJSON,
	)

	return err
}

// GetEventHistory retrieves all events for a journal entry
func (s *EventStore) GetEventHistory(ctx context.Context, journalEntryID uuidv7.UUID) ([]JournalEntryEvent, error) {
	query := `SELECT id, journal_entry_id, event_type, event_version, event_data,
		causation_id, correlation_id, organization_id, user_id, timestamp, metadata
		FROM accounting_journal_entry_events
		WHERE journal_entry_id = $1
		ORDER BY event_version ASC`

	rows, err := s.getExecutor(ctx).QueryContext(ctx, query, journalEntryID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []JournalEntryEvent
	for rows.Next() {
		var (
			id, jeID, orgID, userID          string
			eventTypeStr, eventDataJSON      string
			causationIDStr, correlationIDStr *string
			metadataJSON                     *string
			eventVersion                     int
			timestamp                        time.Time
		)

		err := rows.Scan(&id, &jeID, &eventTypeStr, &eventVersion, &eventDataJSON,
			&causationIDStr, &correlationIDStr, &orgID, &userID, &timestamp, &metadataJSON)
		if err != nil {
			return nil, err
		}

		event := JournalEntryEvent{
			EventType:    EventType(eventTypeStr),
			EventVersion: eventVersion,
			Timestamp:    timestamp,
		}

		event.ID, _ = uuidv7.Parse(id)
		event.JournalEntryID, _ = uuidv7.Parse(jeID)
		event.OrganizationID, _ = uuidv7.Parse(orgID)
		event.UserID, _ = uuidv7.Parse(userID)

		if causationIDStr != nil {
			cid, _ := uuidv7.Parse(*causationIDStr)
			event.CausationID = &cid
		}
		if correlationIDStr != nil {
			cid, _ := uuidv7.Parse(*correlationIDStr)
			event.CorrelationID = &cid
		}

		_ = json.Unmarshal([]byte(eventDataJSON), &event.EventData)
		if metadataJSON != nil {
			_ = json.Unmarshal([]byte(*metadataJSON), &event.Metadata)
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// GetEventsByCorrelation retrieves events by correlation ID (for distributed tracing)
func (s *EventStore) GetEventsByCorrelation(ctx context.Context, correlationID uuidv7.UUID) ([]JournalEntryEvent, error) {
	query := `SELECT id, journal_entry_id, event_type, event_version, event_data,
		causation_id, correlation_id, organization_id, user_id, timestamp, metadata
		FROM accounting_journal_entry_events
		WHERE correlation_id = $1
		ORDER BY timestamp ASC`

	rows, err := s.getExecutor(ctx).QueryContext(ctx, query, correlationID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []JournalEntryEvent
	for rows.Next() {
		var (
			id, jeID, orgID, userID          string
			eventTypeStr, eventDataJSON      string
			causationIDStr, correlationIDStr *string
			metadataJSON                     *string
			eventVersion                     int
			timestamp                        time.Time
		)

		err := rows.Scan(&id, &jeID, &eventTypeStr, &eventVersion, &eventDataJSON,
			&causationIDStr, &correlationIDStr, &orgID, &userID, &timestamp, &metadataJSON)
		if err != nil {
			return nil, err
		}

		event := JournalEntryEvent{
			EventType:    EventType(eventTypeStr),
			EventVersion: eventVersion,
			Timestamp:    timestamp,
		}

		event.ID, _ = uuidv7.Parse(id)
		event.JournalEntryID, _ = uuidv7.Parse(jeID)
		event.OrganizationID, _ = uuidv7.Parse(orgID)
		event.UserID, _ = uuidv7.Parse(userID)

		if causationIDStr != nil {
			cid, _ := uuidv7.Parse(*causationIDStr)
			event.CausationID = &cid
		}
		if correlationIDStr != nil {
			cid, _ := uuidv7.Parse(*correlationIDStr)
			event.CorrelationID = &cid
		}

		_ = json.Unmarshal([]byte(eventDataJSON), &event.EventData)
		if metadataJSON != nil {
			_ = json.Unmarshal([]byte(*metadataJSON), &event.Metadata)
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

func (s *EventStore) getNextVersion(ctx context.Context, journalEntryID uuidv7.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(event_version), 0) + 1 
		FROM accounting_journal_entry_events 
		WHERE journal_entry_id = $1`

	var version int
	executor := s.getExecutor(ctx)

	// Use QueryxContext for sqlx.ExtContext compatibility
	rows, err := executor.QueryxContext(ctx, query, journalEntryID.String())
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		if err := rows.Scan(&version); err != nil {
			return 0, err
		}
	}

	return version, rows.Err()
}

func (s *EventStore) getExecutor(ctx context.Context) sqlx.ExtContext {
	// Check if context contains a transaction
	if tx := ctx.Value("tx"); tx != nil {
		if sqlxTx, ok := tx.(*sqlx.Tx); ok {
			return sqlxTx
		}
	}
	return s.db
}
