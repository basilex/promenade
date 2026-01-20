package usecase

import (
	"context"
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/audit"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/cache"
	"github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IFiscalPeriodUseCase defines the interface for fiscal period business logic
type IFiscalPeriodUseCase interface {
	CreateFiscalPeriod(ctx context.Context, organizationID uuidv7.UUID, code, name string, periodType aggregate.PeriodType, startDate, endDate time.Time, createdBy uuidv7.UUID) (*aggregate.FiscalPeriod, error)
	GetFiscalPeriodByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error)
	GetCurrentPeriod(ctx context.Context, organizationID uuidv7.UUID, date string) (*aggregate.FiscalPeriod, error)
	ClosePeriod(ctx context.Context, id, closedBy uuidv7.UUID) (*aggregate.FiscalPeriod, error)
	ReopenPeriod(ctx context.Context, id, reopenedBy uuidv7.UUID) (*aggregate.FiscalPeriod, error)
	LockPeriod(ctx context.Context, id, lockedBy uuidv7.UUID) (*aggregate.FiscalPeriod, error)
	DeleteFiscalPeriod(ctx context.Context, id uuidv7.UUID) error
	ListFiscalPeriodsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error)
	ListFiscalPeriodsByYear(ctx context.Context, organizationID uuidv7.UUID, fiscalYear int) ([]*aggregate.FiscalPeriod, error)
	ListOpenPeriods(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error)
}

type fiscalPeriodUseCase struct {
	fiscalPeriodRepo  repository.IFiscalPeriodRepository
	fiscalPeriodCache *cache.FiscalPeriodCache
	auditLogger       *audit.AuditLogger
}

// NewFiscalPeriodUseCase creates a new fiscal period use case
func NewFiscalPeriodUseCase(
	fiscalPeriodRepo repository.IFiscalPeriodRepository,
	fiscalPeriodCache *cache.FiscalPeriodCache,
	auditLogger *audit.AuditLogger,
) IFiscalPeriodUseCase {
	return &fiscalPeriodUseCase{
		fiscalPeriodRepo:  fiscalPeriodRepo,
		fiscalPeriodCache: fiscalPeriodCache,
		auditLogger:       auditLogger,
	}
}

func (uc *fiscalPeriodUseCase) CreateFiscalPeriod(
	ctx context.Context,
	organizationID uuidv7.UUID,
	code, name string,
	periodType aggregate.PeriodType,
	startDate, endDate time.Time,
	createdBy uuidv7.UUID,
) (*aggregate.FiscalPeriod, error) {
	// Create new fiscal period
	period, err := aggregate.NewFiscalPeriod(organizationID, code, name, periodType, startDate, endDate, createdBy)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.fiscalPeriodRepo.Create(ctx, period); err != nil {
		return nil, err
	}

	// Log audit
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogCreate(ctx, audit.AuditRecord{
			EntityType:     "fiscal_period",
			EntityID:       period.ID,
			Action:         "create",
			OrganizationID: organizationID,
			UserID:         createdBy,
			Details:        map[string]interface{}{"code": code, "name": name, "type": periodType, "start": startDate, "end": endDate},
		})
	}

	// Invalidate cache
	uc.fiscalPeriodCache.Invalidate(organizationID)

	return period, nil
}

func (uc *fiscalPeriodUseCase) GetFiscalPeriodByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	// Try cache first
	if period, err := uc.fiscalPeriodCache.GetByID(ctx, id); err == nil && period != nil {
		return period, nil
	}

	// Fallback to repository
	return uc.fiscalPeriodRepo.GetByID(ctx, id)
}

func (uc *fiscalPeriodUseCase) GetCurrentPeriod(ctx context.Context, organizationID uuidv7.UUID, date string) (*aggregate.FiscalPeriod, error) {
	// Parse date string
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return uc.fiscalPeriodRepo.GetByOrganizationAndDate(ctx, organizationID, date)
	}

	// Try cache first
	if period, err := uc.fiscalPeriodCache.GetByOrganizationAndDate(ctx, organizationID, parsedDate); err == nil && period != nil {
		return period, nil
	}

	// Fallback to repository
	return uc.fiscalPeriodRepo.GetByOrganizationAndDate(ctx, organizationID, date)
}

func (uc *fiscalPeriodUseCase) ClosePeriod(ctx context.Context, id, closedBy uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	// Get fiscal period
	period, err := uc.fiscalPeriodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Close period
	if err := period.Close(closedBy); err != nil {
		return nil, err
	}

	period.LastUpdatedBy = closedBy
	period.Touch()

	// Save
	if err := uc.fiscalPeriodRepo.Update(ctx, period); err != nil {
		return nil, err
	}

	// Log audit
	if uc.auditLogger != nil {
		_ = uc.auditLogger.LogPeriodClose(ctx, audit.AuditRecord{
			EntityType:     "fiscal_period",
			EntityID:       period.ID,
			Action:         "close",
			OrganizationID: period.OrganizationID,
			UserID:         closedBy,
			Details:        map[string]interface{}{"closed_at": period.ClosedAt, "status": "closed"},
		})
	}

	// Invalidate cache
	uc.fiscalPeriodCache.Invalidate(period.OrganizationID)

	return period, nil
}

func (uc *fiscalPeriodUseCase) ReopenPeriod(ctx context.Context, id, reopenedBy uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	// Get fiscal period
	period, err := uc.fiscalPeriodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Reopen period
	if err := period.Reopen(reopenedBy); err != nil {
		return nil, err
	}

	period.LastUpdatedBy = reopenedBy
	period.Touch()

	// Save
	if err := uc.fiscalPeriodRepo.Update(ctx, period); err != nil {
		return nil, err
	}

	// Invalidate cache
	uc.fiscalPeriodCache.Invalidate(period.OrganizationID)

	return period, nil
}

func (uc *fiscalPeriodUseCase) LockPeriod(ctx context.Context, id, lockedBy uuidv7.UUID) (*aggregate.FiscalPeriod, error) {
	// Get fiscal period
	period, err := uc.fiscalPeriodRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Lock period
	if err := period.Lock(lockedBy); err != nil {
		return nil, err
	}

	period.LastUpdatedBy = lockedBy
	period.Touch()

	// Save
	if err := uc.fiscalPeriodRepo.Update(ctx, period); err != nil {
		return nil, err
	}

	// Invalidate cache
	uc.fiscalPeriodCache.Invalidate(period.OrganizationID)

	return period, nil
}

func (uc *fiscalPeriodUseCase) DeleteFiscalPeriod(ctx context.Context, id uuidv7.UUID) error {
	// Get period to verify it can be deleted
	period, err := uc.fiscalPeriodRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Can only delete open periods (not closed or locked)
	if period.Status != aggregate.PeriodStatusOpen {
		return fiscalperiod.ErrPeriodAlreadyClosed
	}

	return uc.fiscalPeriodRepo.Delete(ctx, id)
}

func (uc *fiscalPeriodUseCase) ListFiscalPeriodsByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.FiscalPeriod, error) {
	return uc.fiscalPeriodRepo.ListByOrganization(ctx, organizationID, limit, offset)
}

func (uc *fiscalPeriodUseCase) ListFiscalPeriodsByYear(ctx context.Context, organizationID uuidv7.UUID, fiscalYear int) ([]*aggregate.FiscalPeriod, error) {
	return uc.fiscalPeriodRepo.ListByYear(ctx, organizationID, fiscalYear)
}

func (uc *fiscalPeriodUseCase) ListOpenPeriods(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.FiscalPeriod, error) {
	// Try cache first
	periods, err := uc.fiscalPeriodCache.ListOpen(ctx, organizationID)
	if err == nil && len(periods) > 0 {
		return periods, nil
	}

	// Fallback to repository
	return uc.fiscalPeriodRepo.ListOpen(ctx, organizationID)
}
