package eventstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
	"github.com/basilex/promenade/internal/infrastructure/database"
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

// JournalEntryEventStore provides event sourcing for journal entries
type JournalEntryEventStore struct {
	db *sqlx.DB
}

// NewJournalEntryEventStore creates a new journal entry event store
func NewJournalEntryEventStore(db *sqlx.DB) *JournalEntryEventStore {
	return &JournalEntryEventStore{db: db}
}

// getExecutor returns the appropriate executor based on context
func (s *JournalEntryEventStore) getExecutor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := database.GetTx(ctx); ok {
		return tx
	}
	return s.db
}

// AppendEvent appends an event to the event store
func (s *JournalEntryEventStore) AppendEvent(ctx context.Context, event JournalEntryEvent) error {
	if event.ID == uuidv7.Nil {
		event.ID = uuidv7.New()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.EventVersion == 0 {
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

// getNextVersion gets the next version number for a journal entry
func (s *JournalEntryEventStore) getNextVersion(ctx context.Context, journalEntryID uuidv7.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(event_version), 0) + 1 FROM accounting_journal_entry_events WHERE journal_entry_id = $1`

	executor := s.getExecutor(ctx)
	var version int
	err := sqlx.GetContext(ctx, executor, &version, query, journalEntryID.String())
	if err != nil {
		return 0, err
	}

	return version, nil
}

// GetEvents retrieves all events for a journal entry
func (s *JournalEntryEventStore) GetEvents(ctx context.Context, journalEntryID uuidv7.UUID) ([]JournalEntryEvent, error) {
	query := `SELECT id, journal_entry_id, event_type, event_version, event_data, 
		causation_id, correlation_id, organization_id, user_id, timestamp, metadata
		FROM accounting_journal_entry_events 
		WHERE journal_entry_id = $1 
		ORDER BY event_version ASC`

	executor := s.getExecutor(ctx)
	rows, err := executor.QueryxContext(ctx, query, journalEntryID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []JournalEntryEvent
	for rows.Next() {
		var event JournalEntryEvent
		var eventDataStr string
		var metadataStr *string
		var causationIDStr, correlationIDStr *string
		var idStr, journalEntryIDStr, organizationIDStr, userIDStr string

		err := rows.Scan(
			&idStr,
			&journalEntryIDStr,
			&event.EventType,
			&event.EventVersion,
			&eventDataStr,
			&causationIDStr,
			&correlationIDStr,
			&organizationIDStr,
			&userIDStr,
			&event.Timestamp,
			&metadataStr,
		)
		if err != nil {
			return nil, err
		}

		event.ID, _ = uuidv7.Parse(idStr)
		event.JournalEntryID, _ = uuidv7.Parse(journalEntryIDStr)
		event.OrganizationID, _ = uuidv7.Parse(organizationIDStr)
		event.UserID, _ = uuidv7.Parse(userIDStr)

		if causationIDStr != nil {
			id, _ := uuidv7.Parse(*causationIDStr)
			event.CausationID = &id
		}
		if correlationIDStr != nil {
			id, _ := uuidv7.Parse(*correlationIDStr)
			event.CorrelationID = &id
		}

		if err := json.Unmarshal([]byte(eventDataStr), &event.EventData); err != nil {
			return nil, err
		}

		if metadataStr != nil {
			if err := json.Unmarshal([]byte(*metadataStr), &event.Metadata); err != nil {
				return nil, err
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// GetEventsByOrganization retrieves events for an organization
func (s *JournalEntryEventStore) GetEventsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit int) ([]JournalEntryEvent, error) {
	query := `SELECT id, journal_entry_id, event_type, event_version, event_data, 
		causation_id, correlation_id, organization_id, user_id, timestamp, metadata
		FROM accounting_journal_entry_events 
		WHERE organization_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2`

	executor := s.getExecutor(ctx)
	rows, err := executor.QueryxContext(ctx, query, organizationID.String(), limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []JournalEntryEvent
	for rows.Next() {
		var event JournalEntryEvent
		var eventDataStr string
		var metadataStr *string
		var causationIDStr, correlationIDStr *string
		var idStr, journalEntryIDStr, organizationIDStr, userIDStr string

		err := rows.Scan(
			&idStr,
			&journalEntryIDStr,
			&event.EventType,
			&event.EventVersion,
			&eventDataStr,
			&causationIDStr,
			&correlationIDStr,
			&organizationIDStr,
			&userIDStr,
			&event.Timestamp,
			&metadataStr,
		)
		if err != nil {
			return nil, err
		}

		event.ID, _ = uuidv7.Parse(idStr)
		event.JournalEntryID, _ = uuidv7.Parse(journalEntryIDStr)
		event.OrganizationID, _ = uuidv7.Parse(organizationIDStr)
		event.UserID, _ = uuidv7.Parse(userIDStr)

		if causationIDStr != nil {
			id, _ := uuidv7.Parse(*causationIDStr)
			event.CausationID = &id
		}
		if correlationIDStr != nil {
			id, _ := uuidv7.Parse(*correlationIDStr)
			event.CorrelationID = &id
		}

		if err := json.Unmarshal([]byte(eventDataStr), &event.EventData); err != nil {
			return nil, err
		}

		if metadataStr != nil {
			if err := json.Unmarshal([]byte(*metadataStr), &event.Metadata); err != nil {
				return nil, err
			}
		}

		events = append(events, event)
	}

	return events, rows.Err()
}

// RebuildAggregate rebuilds a journal entry aggregate from events
func (s *JournalEntryEventStore) RebuildAggregate(ctx context.Context, journalEntryID uuidv7.UUID) (*aggregate.JournalEntry, error) {
	events, err := s.GetEvents(ctx, journalEntryID)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	firstEvent := events[0]
	if firstEvent.EventType != EventEntryCreated {
		return nil, nil
	}

	entryData := firstEvent.EventData
	organizationID := firstEvent.OrganizationID
	entryDate, _ := time.Parse(time.RFC3339, entryData["entry_date"].(string))
	description := entryData["description"].(string)

	entry := &aggregate.JournalEntry{}
	entry.ID = journalEntryID
	entry.OrganizationID = organizationID
	entry.EntryDate = entryDate
	entry.Description = description
	entry.Status = aggregate.EntryStatusDraft
	entry.CreatedAt = firstEvent.Timestamp
	entry.UpdatedAt = firstEvent.Timestamp

	for i := 1; i < len(events); i++ {
		event := events[i]
		switch event.EventType {
		case EventEntryPosted:
			entry.Status = aggregate.EntryStatusPosted
		case EventEntryReversed:
			entry.Status = aggregate.EntryStatusReversed
		case EventDescriptionUpdated:
			entry.Description = event.EventData["new_description"].(string)
		}
		entry.UpdatedAt = event.Timestamp
	}

	return entry, nil
}
