package purge

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/modules/audit/usecase"
	"github.com/basilex/promenade/pkg/purge"
)

// AuditEventPurgeHandler handles purging of old audit events
type AuditEventPurgeHandler struct {
	useCase       usecase.AuditEventUseCase
	retentionDays int
}

// NewAuditEventPurgeHandler creates a new audit event purge handler
func NewAuditEventPurgeHandler(useCase usecase.AuditEventUseCase, retentionDays int) *AuditEventPurgeHandler {
	return &AuditEventPurgeHandler{
		useCase:       useCase,
		retentionDays: retentionDays,
	}
}

// Name returns handler name
func (h *AuditEventPurgeHandler) Name() string {
	return "audit_events"
}

// EntityName returns entity name
func (h *AuditEventPurgeHandler) EntityName() string {
	return "audit_events"
}

// Purge deletes audit events older than retention period
func (h *AuditEventPurgeHandler) Purge(ctx context.Context, cutoffDate time.Time, batchSize int, dryRun bool) (int64, error) {
	if dryRun {
		return 0, nil
	}
	return h.useCase.PurgeOldAuditEvents(ctx, h.retentionDays)
}

// RegisterPurgeHandlers registers audit module purge handlers
func RegisterPurgeHandlers(useCase usecase.AuditEventUseCase, retentionDays int) error {
	handler := NewAuditEventPurgeHandler(useCase, retentionDays)
	
	// Register handler
	if err := purge.DefaultRegistry.Register(handler); err != nil {
		return err
	}
	
	// Register policy
	policy := purge.RetentionPolicy{
		EntityName:    "audit_events",
		RetentionDays: retentionDays,
		Enabled:       retentionDays > 0,
	}
	
	return purge.DefaultPolicyRegistry.RegisterPolicy(policy)
}
