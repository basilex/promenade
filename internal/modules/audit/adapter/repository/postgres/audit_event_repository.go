package postgres

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/internal/modules/audit/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/jmoiron/sqlx"
)

type auditEventRepository struct {
	*BaseRepository
}

// NewAuditEventRepository creates a new audit event repository
func NewAuditEventRepository(db *sqlx.DB) repository.AuditEventRepository {
	return &auditEventRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new audit event
func (r *auditEventRepository) Create(ctx context.Context, event *entity.AuditEvent) error {
	query := `
		INSERT INTO audit_events (
			id, user_id, action, entity_type, entity_id,
			old_data, new_data, ip_address, user_agent, request_id,
			metadata, signature, created_at
		) VALUES (
			:id, :user_id, :action, :entity_type, :entity_id,
			:old_data, :new_data, :ip_address, :user_agent, :request_id,
			:metadata, :signature, :created_at
		)
	`
	_, err := r.NamedExec(ctx, query, event)
	return err
}

// GetByID retrieves an audit event by ID
func (r *auditEventRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.AuditEvent, error) {
	var event entity.AuditEvent
	query := `
		SELECT id, user_id, action, entity_type, entity_id,
			   old_data, new_data, ip_address, user_agent, request_id,
			   metadata, signature, created_at
		FROM audit_events
		WHERE id = $1
	`
	if err := r.Get(ctx, &event, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrAuditEventNotFound
		}
		return nil, err
	}
	return &event, nil
}

// List retrieves audit events with pagination and filters
func (r *auditEventRepository) List(ctx context.Context, filters repository.AuditEventFilters, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	query := `
		SELECT id, user_id, action, entity_type, entity_id,
			   old_data, new_data, ip_address, user_agent, request_id,
			   metadata, signature, created_at
		FROM audit_events
		WHERE 1=1
	`
	args := make(map[string]interface{})

	if filters.UserID != nil {
		query += " AND user_id = :user_id"
		args["user_id"] = *filters.UserID
	}
	if filters.Action != nil {
		query += " AND action = :action"
		args["action"] = *filters.Action
	}
	if filters.EntityType != nil {
		query += " AND entity_type = :entity_type"
		args["entity_type"] = *filters.EntityType
	}
	if filters.EntityID != nil {
		query += " AND entity_id = :entity_id"
		args["entity_id"] = *filters.EntityID
	}
	if filters.DateFrom != nil {
		query += " AND created_at >= :date_from"
		args["date_from"] = *filters.DateFrom
	}
	if filters.DateTo != nil {
		query += " AND created_at <= :date_to"
		args["date_to"] = *filters.DateTo
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS filtered"
	var total int
	namedCountQuery, countArgs, err := sqlx.Named(countQuery, args)
	if err != nil {
		return nil, nil, err
	}
	namedCountQuery = sqlx.Rebind(sqlx.DOLLAR, namedCountQuery)
	if err := r.Get(ctx, &total, namedCountQuery, countArgs...); err != nil {
		return nil, nil, err
	}

	// Add sorting and pagination
	query += " ORDER BY created_at DESC"
	limit := params.Limit
	if limit == 0 {
		limit = 20
	}
	offset := params.GetOffset()

	query += " LIMIT :limit OFFSET :offset"
	args["limit"] = limit
	args["offset"] = offset

	// Execute query
	var events []*entity.AuditEvent
	namedQuery, queryArgs, err := sqlx.Named(query, args)
	if err != nil {
		return nil, nil, err
	}
	namedQuery = sqlx.Rebind(sqlx.DOLLAR, namedQuery)
	if err := r.Select(ctx, &events, namedQuery, queryArgs...); err != nil {
		return nil, nil, err
	}

	metadata := pagination.NewMetadata(total, limit, offset)
	return events, metadata, nil
}

// GetByEntityID retrieves all audit events for a specific entity
func (r *auditEventRepository) GetByEntityID(ctx context.Context, entityType string, entityID uuidv7.UUID) ([]*entity.AuditEvent, error) {
	var events []*entity.AuditEvent
	query := `
		SELECT id, user_id, action, entity_type, entity_id,
			   old_data, new_data, ip_address, user_agent, request_id,
			   metadata, signature, created_at
		FROM audit_events
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
	`
	if err := r.Select(ctx, &events, query, entityType, entityID); err != nil {
		return nil, err
	}
	return events, nil
}

// GetByUserID retrieves all audit events by a specific user
func (r *auditEventRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM audit_events WHERE user_id = $1`
	if err := r.Get(ctx, &total, countQuery, userID); err != nil {
		return nil, nil, err
	}

	// Get events
	limit := params.Limit
	if limit == 0 {
		limit = 20
	}
	offset := params.GetOffset()

	var events []*entity.AuditEvent
	query := `
		SELECT id, user_id, action, entity_type, entity_id,
			   old_data, new_data, ip_address, user_agent, request_id,
			   metadata, signature, created_at
		FROM audit_events
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	if err := r.Select(ctx, &events, query, userID, limit, offset); err != nil {
		return nil, nil, err
	}

	metadata := pagination.NewMetadata(total, limit, offset)
	return events, metadata, nil
}

// GetByAction retrieves audit events by action type
func (r *auditEventRepository) GetByAction(ctx context.Context, action string, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM audit_events WHERE action = $1`
	if err := r.Get(ctx, &total, countQuery, action); err != nil {
		return nil, nil, err
	}

	// Get events
	limit := params.Limit
	if limit == 0 {
		limit = 20
	}
	offset := params.GetOffset()

	var events []*entity.AuditEvent
	query := `
		SELECT id, user_id, action, entity_type, entity_id,
			   old_data, new_data, ip_address, user_agent, request_id,
			   metadata, signature, created_at
		FROM audit_events
		WHERE action = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	if err := r.Select(ctx, &events, query, action, limit, offset); err != nil {
		return nil, nil, err
	}

	metadata := pagination.NewMetadata(total, limit, offset)
	return events, metadata, nil
}

// DeleteOlderThan deletes audit events older than the specified date
func (r *auditEventRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM audit_events WHERE created_at < $1`
	result, err := r.Exec(ctx, query, before)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// VerifySignature verifies the HMAC-SHA256 signature of an audit event
func (r *auditEventRepository) VerifySignature(ctx context.Context, event *entity.AuditEvent, secret string) (bool, error) {
	expected := generateSignature(event, secret)
	return hmac.Equal([]byte(event.Signature), []byte(expected)), nil
}

// generateSignature creates HMAC-SHA256 signature for audit event
func generateSignature(event *entity.AuditEvent, secret string) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		event.ID,
		event.UserID,
		event.Action,
		event.EntityType,
		event.EntityID,
		event.CreatedAt.Format(time.RFC3339Nano),
	)

	if event.OldData != nil {
		data += ":" + *event.OldData
	}
	if event.NewData != nil {
		data += ":" + *event.NewData
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSignature generates HMAC-SHA256 signature for an audit event (exported helper)
func GenerateSignature(event *entity.AuditEvent, secret string) string {
	return generateSignature(event, secret)
}
