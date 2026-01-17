package integration

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/scheduler"
	"github.com/basilex/promenade/pkg/uuidv7"
)

const (
	defaultReceiptRetryCron = "*/5 * * * *"
	receiptRetryJobName     = "fiscal_receipt_retry"
)

// ReceiptRetryExecutor retries printing pending fiscal receipts.
type ReceiptRetryExecutor struct {
	base      *scheduler.BaseExecutor
	receiptUC receipt.IUseCase
}

// NewReceiptRetryExecutor creates a new receipt retry executor.
func NewReceiptRetryExecutor(receiptUC receipt.IUseCase) *ReceiptRetryExecutor {
	return &ReceiptRetryExecutor{
		base:      scheduler.NewBaseExecutor(scheduler.JobTypeCustom),
		receiptUC: receiptUC,
	}
}

// Type returns job type.
func (e *ReceiptRetryExecutor) Type() scheduler.JobType {
	return scheduler.JobTypeCustom
}

// Execute retries printing all pending receipts.
func (e *ReceiptRetryExecutor) Execute(ctx context.Context, job *scheduler.Job) error {
	log := logger.FromContext(ctx)

	status := receipt.ReceiptStatusPending
	receipts, err := e.receiptUC.ListReceipts(ctx, &receipt.ListFilters{Status: &status})
	if err != nil {
		log.Error("Failed to list pending receipts for retry", slog.Any("error", err))
		return err
	}

	if len(receipts) == 0 {
		log.Debug("No pending receipts for retry", slog.String("job", job.Name))
		return nil
	}

	log.Info("Retrying pending receipts", slog.Int("count", len(receipts)))

	for _, rec := range receipts {
		if rec == nil {
			continue
		}

		_, err := e.receiptUC.PrintReceipt(ctx, rec.GetID(), rec.LastUpdatedBy)
		if err != nil {
			log.Warn("Receipt retry failed",
				slog.String("receipt_id", rec.GetID().String()),
				slog.Any("error", err),
			)
			continue
		}

		log.Info("Receipt printed on retry", slog.String("receipt_id", rec.GetID().String()))
	}

	return nil
}

// RegisterReceiptRetryJob registers fiscal receipt retry job with scheduler.
func RegisterReceiptRetryJob(engine *scheduler.Engine, receiptUC receipt.IUseCase, cronExpr string) error {
	if engine == nil || receiptUC == nil {
		return errors.New("scheduler engine and receipt use case are required")
	}

	if cronExpr == "" {
		cronExpr = defaultReceiptRetryCron
	}

	executor := NewReceiptRetryExecutor(receiptUC)
	if err := engine.RegisterExecutor(executor.Type(), executor); err != nil && !errors.Is(err, scheduler.ErrExecutorAlreadyRegistered) {
		return err
	}

	now := time.Now()
	job := &scheduler.Job{
		ID:             uuidv7.New(),
		Name:           receiptRetryJobName,
		Type:           executor.Type(),
		CronExpression: cronExpr,
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return engine.AddJob(job)
}
