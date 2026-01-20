package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// AuditRecord represents an audit log entry
type AuditRecord struct {
	EntityType     string
	EntityID       uuidv7.UUID
	Action         string
	OrganizationID uuidv7.UUID
	UserID         uuidv7.UUID
	ChangesBefore  map[string]interface{}
	ChangesAfter   map[string]interface{}
	Details        map[string]interface{}
	IPAddress      string
	UserAgent      string
}

// AuditLogger provides audit logging functionality
type AuditLogger struct {
	db *sqlx.DB
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(db *sqlx.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// LogCreate logs entity creation
func (l *AuditLogger) LogCreate(ctx context.Context, record AuditRecord) error {
	return l.log(ctx, record, "create")
}

// LogUpdate logs entity updates with before/after state
func (l *AuditLogger) LogUpdate(ctx context.Context, record AuditRecord) error {
	return l.log(ctx, record, "update")
}

// LogPost logs journal entry posting
func (l *AuditLogger) LogPost(ctx context.Context, record AuditRecord) error {
	return l.log(ctx, record, "post")
}

// LogReverse logs journal entry reversal
func (l *AuditLogger) LogReverse(ctx context.Context, record AuditRecord) error {
	return l.log(ctx, record, "reverse")
}

// LogPeriodClose logs fiscal period closure
func (l *AuditLogger) LogPeriodClose(ctx context.Context, record AuditRecord) error {
	return l.log(ctx, record, "period_close")
}

func (l *AuditLogger) log(ctx context.Context, record AuditRecord, changeType string) error {
	id := uuidv7.New()
	timestamp := time.Now()

	var beforeStateJSON, afterStateJSON, detailsJSON *string

	if record.ChangesBefore != nil {
		data, err := json.Marshal(record.ChangesBefore)
		if err != nil {
			return err
		}
		str := string(data)
		beforeStateJSON = &str
	}

	if record.ChangesAfter != nil {
		data, err := json.Marshal(record.ChangesAfter)
		if err != nil {
			return err
		}
		str := string(data)
		afterStateJSON = &str
	}

	if record.Details != nil {
		data, err := json.Marshal(record.Details)
		if err != nil {
			return err
		}
		str := string(data)
		detailsJSON = &str
	}

	var ipAddress, userAgent *string
	if record.IPAddress != "" {
		ipAddress = &record.IPAddress
	}
	if record.UserAgent != "" {
		userAgent = &record.UserAgent
	}

	query := `INSERT INTO accounting_audit_log 
		(id, entity_type, entity_id, action, change_type, organization_id, user_id, 
		 before_state, after_state, details, ip_address, user_agent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	executor := l.getExecutor(ctx)
	_, err := executor.ExecContext(ctx, query,
		id.String(),
		record.EntityType,
		record.EntityID.String(),
		record.Action,
		changeType,
		record.OrganizationID.String(),
		record.UserID.String(),
		beforeStateJSON,
		afterStateJSON,
		detailsJSON,
		ipAddress,
		userAgent,
		timestamp,
	)

	return err
}

// GetAuditTrail retrieves audit history for an entity
func (l *AuditLogger) GetAuditTrail(ctx context.Context, entityType string, entityID uuidv7.UUID) ([]AuditEntry, error) {
	query := `SELECT id, entity_type, entity_id, action, change_type, organization_id, user_id,
		before_state, after_state, details, ip_address, user_agent, timestamp
		FROM accounting_audit_log
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY timestamp DESC`

	rows, err := l.getExecutor(ctx).QueryContext(ctx, query, entityType, entityID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []AuditEntry
	for rows.Next() {
		var (
			id, entType, entID, action, changeType, orgID, userID string
			beforeStateJSON, afterStateJSON, detailsJSON           *string
			ipAddress, userAgent                                    *string
			timestamp                                               time.Time
		)

		err := rows.Scan(&id, &entType, &entID, &action, &changeType, &orgID, &userID,
			&beforeStateJSON, &afterStateJSON, &detailsJSON, &ipAddress, &userAgent, &timestamp)
		if err != nil {
			return nil, err
		}

		entry := AuditEntry{
			EntityType: entType,
			Action:     action,
			ChangeType: changeType,
			Timestamp:  timestamp,
		}

		entry.ID, _ = uuidv7.Parse(id)
		entry.EntityID, _ = uuidv7.Parse(entID)
		entry.OrganizationID, _ = uuidv7.Parse(orgID)
		entry.UserID, _ = uuidv7.Parse(userID)

		if ipAddress != nil {
			entry.IPAddress = *ipAddress
		}
		if userAgent != nil {
			entry.UserAgent = *userAgent
		}

		if beforeStateJSON != nil {
			_ = json.Unmarshal([]byte(*beforeStateJSON), &entry.BeforeState)
		}
		if afterStateJSON != nil {
			_ = json.Unmarshal([]byte(*afterStateJSON), &entry.AfterState)
		}
		if detailsJSON != nil {
			_ = json.Unmarshal([]byte(*detailsJSON), &entry.Details)
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// AuditEntry represents a retrieved audit log entry
type AuditEntry struct {
	ID             uuidv7.UUID
	EntityType     string
	EntityID       uuidv7.UUID
	Action         string
	ChangeType     string
	OrganizationID uuidv7.UUID
	UserID         uuidv7.UUID
	BeforeState    map[string]interface{}
	AfterState     map[string]interface{}
	Details        map[string]interface{}
	IPAddress      string
	UserAgent      string
	Timestamp      time.Time
}

func (l *AuditLogger) getExecutor(ctx context.Context) sqlx.ExtContext {
	// Check if context contains a transaction
	if tx := ctx.Value("tx"); tx != nil {
		if sqlxTx, ok := tx.(*sqlx.Tx); ok {
			return sqlxTx
		}
	}
	return l.db
}
