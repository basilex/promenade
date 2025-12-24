package usecase

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/modules/audit/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/modules/audit/domain/entity"
	"github.com/basilex/promenade/internal/modules/audit/domain/repository"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// AuditEventUseCase defines the interface for audit event business logic
type AuditEventUseCase interface {
	CreateAuditEvent(ctx context.Context, data *entity.AuditEventCreate) (*entity.AuditEvent, error)
	GetAuditEvent(ctx context.Context, id uuidv7.UUID) (*entity.AuditEvent, error)
	ListAuditEvents(ctx context.Context, filters repository.AuditEventFilters, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error)
	GetAuditEventsByEntity(ctx context.Context, entityType string, entityID uuidv7.UUID) ([]*entity.AuditEvent, error)
	GetAuditEventsByUser(ctx context.Context, userID uuidv7.UUID, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error)
	GetAuditEventsByAction(ctx context.Context, action string, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error)
	VerifyAuditEvent(ctx context.Context, id uuidv7.UUID) (bool, error)
	PurgeOldAuditEvents(ctx context.Context, retentionDays int) (int64, error)
}

type auditEventUseCase struct {
	repo   repository.AuditEventRepository
	secret string
}

// NewAuditEventUseCase creates a new audit event use case
func NewAuditEventUseCase(repo repository.AuditEventRepository, secret string) AuditEventUseCase {
	return &auditEventUseCase{
		repo:   repo,
		secret: secret,
	}
}

// CreateAuditEvent creates a new audit event with signature
func (uc *auditEventUseCase) CreateAuditEvent(ctx context.Context, data *entity.AuditEventCreate) (*entity.AuditEvent, error) {
	if err := data.Validate(); err != nil {
		return nil, err
	}

	event := &entity.AuditEvent{
		ID:         uuidv7.New(),
		UserID:     data.UserID,
		Action:     data.Action,
		EntityType: data.EntityType,
		EntityID:   data.EntityID,
		OldData:    data.OldData,
		NewData:    data.NewData,
		IPAddress:  data.IPAddress,
		UserAgent:  data.UserAgent,
		RequestID:  data.RequestID,
		Metadata:   data.Metadata,
		CreatedAt:  time.Now(),
	}

	event.Signature = postgres.GenerateSignature(event, uc.secret)

	if err := uc.repo.Create(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// GetAuditEvent retrieves an audit event by ID
func (uc *auditEventUseCase) GetAuditEvent(ctx context.Context, id uuidv7.UUID) (*entity.AuditEvent, error) {
	return uc.repo.GetByID(ctx, id)
}

// ListAuditEvents retrieves audit events with filters and pagination
func (uc *auditEventUseCase) ListAuditEvents(ctx context.Context, filters repository.AuditEventFilters, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	return uc.repo.List(ctx, filters, params)
}

// GetAuditEventsByEntity retrieves all audit events for a specific entity
func (uc *auditEventUseCase) GetAuditEventsByEntity(ctx context.Context, entityType string, entityID uuidv7.UUID) ([]*entity.AuditEvent, error) {
	return uc.repo.GetByEntityID(ctx, entityType, entityID)
}

// GetAuditEventsByUser retrieves audit events by user with pagination
func (uc *auditEventUseCase) GetAuditEventsByUser(ctx context.Context, userID uuidv7.UUID, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	return uc.repo.GetByUserID(ctx, userID, params)
}

// GetAuditEventsByAction retrieves audit events by action type with pagination
func (uc *auditEventUseCase) GetAuditEventsByAction(ctx context.Context, action string, params pagination.Params) ([]*entity.AuditEvent, *pagination.Metadata, error) {
	return uc.repo.GetByAction(ctx, action, params)
}

// VerifyAuditEvent verifies the signature of an audit event
func (uc *auditEventUseCase) VerifyAuditEvent(ctx context.Context, id uuidv7.UUID) (bool, error) {
	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return uc.repo.VerifySignature(ctx, event, uc.secret)
}

// PurgeOldAuditEvents deletes audit events older than retention period
func (uc *auditEventUseCase) PurgeOldAuditEvents(ctx context.Context, retentionDays int) (int64, error) {
	before := time.Now().AddDate(0, 0, -retentionDays)
	return uc.repo.DeleteOlderThan(ctx, before)
}
