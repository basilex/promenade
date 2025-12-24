package repository

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// AuditEventFilters defines filtering options for audit event queries
type AuditEventFilters struct {
	UserID     *uuidv7.UUID
	Action     *string
	EntityType *string
	EntityID   *uuidv7.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
}

// AuditEventRepository defines audit event data access operations
type AuditEventRepository interface {
	// Create creates a new audit event (immutable after creation)
	Create(ctx context.Context, event *entity.AuditEvent) error

	// GetByID retrieves audit event by ID
	GetByID(ctx context.Context, id uuidv7.UUID) (*entity.AuditEvent, error)

	// List retrieves audit events with pagination and filters
	List(ctx context.Context, filters AuditEventFilters, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error)

	// GetByEntityID retrieves all audit events for specific entity
	GetByEntityID(ctx context.Context, entityType string, entityID uuidv7.UUID) ([]*entity.AuditEvent, error)

	// GetByUserID retrieves all audit events by specific user
	GetByUserID(ctx context.Context, userID uuidv7.UUID, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error)

	// GetByAction retrieves audit events by action type
	GetByAction(ctx context.Context, action string, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error)

	// DeleteOlderThan deletes audit events older than specified date (for retention policies)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)

	// VerifySignature verifies audit event signature
	VerifySignature(ctx context.Context, event *entity.AuditEvent, secret string) (bool, error)
}
