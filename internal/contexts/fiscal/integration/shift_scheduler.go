package integration

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/pkg/fiscal/checkbox"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/scheduler"
	"github.com/basilex/promenade/pkg/uuidv7"
)

const (
	defaultShiftOpenCron  = "0 9 * * *"
	defaultShiftCloseCron = "0 23 * * *"
	shiftOpenJobName      = "fiscal_shift_open"
	shiftCloseJobName     = "fiscal_shift_close"
)

// ShiftClient defines the Checkbox client methods needed for shift automation.
type ShiftClient interface {
	OpenShift(ctx context.Context, cashRegisterID string) (*checkbox.ShiftResponse, error)
	CloseShift(ctx context.Context, shiftID string) (*checkbox.ZReport, error)
}

// RegisterShiftJobs registers shift open/close jobs with scheduler.
func RegisterShiftJobs(engine *scheduler.Engine, repo cashregister.IRepository, client ShiftClient, openCron, closeCron string) error {
	if engine == nil || repo == nil || client == nil {
		return errors.New("scheduler engine, repository, and checkbox client are required")
	}

	if openCron == "" {
		openCron = defaultShiftOpenCron
	}
	if closeCron == "" {
		closeCron = defaultShiftCloseCron
	}

	openExecutor := NewShiftOpenExecutor(repo, client)
	closeExecutor := NewShiftCloseExecutor(repo, client)

	if err := engine.RegisterExecutor(openExecutor.Type(), openExecutor); err != nil && !errors.Is(err, scheduler.ErrExecutorAlreadyRegistered) {
		return err
	}
	if err := engine.RegisterExecutor(closeExecutor.Type(), closeExecutor); err != nil && !errors.Is(err, scheduler.ErrExecutorAlreadyRegistered) {
		return err
	}

	now := time.Now()
	openJob := &scheduler.Job{
		ID:             uuidv7.New(),
		Name:           shiftOpenJobName,
		Type:           openExecutor.Type(),
		CronExpression: openCron,
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := engine.AddJob(openJob); err != nil {
		return err
	}

	closeJob := &scheduler.Job{
		ID:             uuidv7.New(),
		Name:           shiftCloseJobName,
		Type:           closeExecutor.Type(),
		CronExpression: closeCron,
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return engine.AddJob(closeJob)
}

// ShiftOpenExecutor opens shifts for active cash registers.
type ShiftOpenExecutor struct {
	base   *scheduler.BaseExecutor
	repo   cashregister.IRepository
	client ShiftClient
}

// NewShiftOpenExecutor creates a new shift open executor.
func NewShiftOpenExecutor(repo cashregister.IRepository, client ShiftClient) *ShiftOpenExecutor {
	return &ShiftOpenExecutor{base: scheduler.NewBaseExecutor(scheduler.JobTypeCustom), repo: repo, client: client}
}

// Type returns job type.
func (e *ShiftOpenExecutor) Type() scheduler.JobType {
	return scheduler.JobTypeCustom
}

// Execute opens shifts for all active registers without an active shift.
func (e *ShiftOpenExecutor) Execute(ctx context.Context, job *scheduler.Job) error {
	log := logger.FromContext(ctx)

	active := true
	registers, err := e.repo.List(ctx, &cashregister.ListFilters{IsActive: &active})
	if err != nil {
		log.Error("Failed to list active cash registers", slog.Any("error", err))
		return err
	}

	for _, cr := range registers {
		if cr == nil {
			continue
		}
		if cr.ActiveShiftID != "" {
			continue
		}

		providerID := cr.ProviderCashRegisterID
		if providerID == "" {
			providerID = cr.FiscalNumber
		}
		if providerID == "" {
			log.Warn("Skipping shift open: missing provider cash register id",
				slog.String("cash_register_id", cr.GetID().String()),
			)
			continue
		}

		shift, err := e.client.OpenShift(ctx, providerID)
		if err != nil {
			log.Warn("Failed to open shift",
				slog.String("cash_register_id", cr.GetID().String()),
				slog.Any("error", err),
			)
			continue
		}

		actorID := uuidv7.New()
		if err := cr.OpenShift(shift.ID, actorID); err != nil {
			log.Warn("Failed to persist shift open",
				slog.String("cash_register_id", cr.GetID().String()),
				slog.Any("error", err),
			)
			continue
		}

		if err := e.repo.Update(ctx, cr); err != nil {
			log.Warn("Failed to update cash register after shift open",
				slog.String("cash_register_id", cr.GetID().String()),
				slog.Any("error", err),
			)
			continue
		}
	}

	return nil
}

// ShiftCloseExecutor closes active shifts and stores Z-report references.
type ShiftCloseExecutor struct {
	base   *scheduler.BaseExecutor
	repo   cashregister.IRepository
	client ShiftClient
}

// NewShiftCloseExecutor creates a new shift close executor.
func NewShiftCloseExecutor(repo cashregister.IRepository, client ShiftClient) *ShiftCloseExecutor {
	return &ShiftCloseExecutor{base: scheduler.NewBaseExecutor(scheduler.JobTypeCustom), repo: repo, client: client}
}

// Type returns job type.
func (e *ShiftCloseExecutor) Type() scheduler.JobType {
	return scheduler.JobTypeCustom
}

// Execute closes shifts for all active registers with an open shift.
func (e *ShiftCloseExecutor) Execute(ctx context.Context, job *scheduler.Job) error {
	log := logger.FromContext(ctx)

	active := true
	registers, err := e.repo.List(ctx, &cashregister.ListFilters{IsActive: &active})
	if err != nil {
		log.Error("Failed to list active cash registers", slog.Any("error", err))
		return err
	}

	for _, cr := range registers {
		if cr == nil {
			continue
		}
		if cr.ActiveShiftID == "" {
			continue
		}

		report, err := e.client.CloseShift(ctx, cr.ActiveShiftID)
		if err != nil {
			log.Warn("Failed to close shift",
				slog.String("cash_register_id", cr.GetID().String()),
				slog.Any("error", err),
			)
			continue
		}

		actorID := uuidv7.New()
		zReportID := ""
		if report != nil {
			zReportID = report.ID
		}
		if err := cr.CloseShift(zReportID, actorID); err != nil {
			log.Warn("Failed to persist shift close",
				slog.String("cash_register_id", cr.GetID().String()),
				slog.Any("error", err),
			)
			continue
		}

		if err := e.repo.Update(ctx, cr); err != nil {
			log.Warn("Failed to update cash register after shift close",
				slog.String("cash_register_id", cr.GetID().String()),
				slog.Any("error", err),
			)
			continue
		}
	}

	return nil
}
