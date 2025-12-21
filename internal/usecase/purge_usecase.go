package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/event"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
)

var (
	ErrPurgeDisabled    = errors.New("purge is disabled")
	ErrInvalidBatchSize = errors.New("batch size must be greater than 0")
	ErrInvalidRetention = errors.New("retention days must be greater than 0")
	ErrPolicyNotFound   = errors.New("retention policy not found")
	ErrPolicyDisabled   = errors.New("retention policy is disabled")
)

// PurgeUseCase defines business logic for purging soft-deleted records
type PurgeUseCase interface {
	// PurgeEntity purges a specific entity based on its retention policy
	PurgeEntity(ctx context.Context, entityName string, dryRun bool) (*entity.PurgeResult, error)

	// PurgeAll purges all entities based on their retention policies
	PurgeAll(ctx context.Context, dryRun bool) (*entity.PurgeSummary, error)

	// GetRetentionPolicies returns all configured retention policies
	GetRetentionPolicies(ctx context.Context) ([]entity.RetentionPolicy, error)

	// GetRetentionPolicy returns a specific retention policy
	GetRetentionPolicy(ctx context.Context, entityName string) (*entity.RetentionPolicy, error)

	// PreviewPurge returns counts of records that would be purged without actually deleting them
	PreviewPurge(ctx context.Context, entityName string) (int64, error)
}

type purgeUseCase struct {
	purgeRepo repository.PurgeRepository
	policies  map[string]entity.RetentionPolicy
	batchSize int
	eventBus  bus.Bus
}

// NewPurgeUseCase creates a new purge use case
func NewPurgeUseCase(
	purgeRepo repository.PurgeRepository,
	policies []entity.RetentionPolicy,
	batchSize int,
	eventBus bus.Bus,
) PurgeUseCase {
	policyMap := make(map[string]entity.RetentionPolicy)
	for _, policy := range policies {
		policyMap[policy.EntityName] = policy
	}

	return &purgeUseCase{
		purgeRepo: purgeRepo,
		policies:  policyMap,
		batchSize: batchSize,
		eventBus:  eventBus,
	}
}

func (uc *purgeUseCase) PurgeEntity(ctx context.Context, entityName string, dryRun bool) (*entity.PurgeResult, error) {
	log := logger.FromContext(ctx)

	policy, exists := uc.policies[entityName]
	if !exists {
		log.Error("Retention policy not found", "entity", entityName)
		return nil, ErrPolicyNotFound
	}

	if !policy.Enabled {
		log.Warn("Retention policy is disabled", "entity", entityName)
		return nil, ErrPolicyDisabled
	}

	if policy.RetentionDays <= 0 {
		log.Error("Invalid retention days", "entity", entityName, "retention_days", policy.RetentionDays)
		return nil, ErrInvalidRetention
	}

	cutoffDate := policy.GetCutoffDate()
	startTime := time.Now()

	log.Info("Starting purge operation",
		"entity", entityName,
		"cutoff_date", cutoffDate,
		"retention_days", policy.RetentionDays,
		"dry_run", dryRun,
	)

	var recordsPurged int64
	var err error

	switch entityName {
	case "user_posts":
		recordsPurged, err = uc.purgeRepo.PurgeUserPosts(ctx, cutoffDate, uc.batchSize, dryRun)
	case "post_comments":
		recordsPurged, err = uc.purgeRepo.PurgePostComments(ctx, cutoffDate, uc.batchSize, dryRun)
	default:
		log.Error("Unknown entity name", "entity", entityName)
		return nil, ErrPolicyNotFound
	}

	duration := time.Since(startTime)

	result := &entity.PurgeResult{
		EntityName:    entityName,
		RecordsPurged: recordsPurged,
		Duration:      duration,
		DryRun:        dryRun,
		Timestamp:     time.Now(),
	}

	if err != nil {
		log.Error("Purge operation failed",
			"entity", entityName,
			"error", err,
			"duration", duration,
		)
		result.Error = err.Error()

		// Publish failure event
		failedEvent := event.NewPurgeFailedEvent(entityName, err)
		if publishErr := uc.eventBus.Publish(ctx, bus.TopicPurgeFailed, failedEvent); publishErr != nil {
			log.Error("Failed to publish purge failed event", "error", publishErr)
		}

		return result, err
	}

	log.Info("Purge operation completed",
		"entity", entityName,
		"records_purged", recordsPurged,
		"duration", duration,
		"dry_run", dryRun,
	)

	return result, nil
}

func (uc *purgeUseCase) PurgeAll(ctx context.Context, dryRun bool) (*entity.PurgeSummary, error) {
	log := logger.FromContext(ctx)

	log.Info("Starting purge all operation", "dry_run", dryRun, "entities_count", len(uc.policies))

	startTime := time.Now()
	results := make([]entity.PurgeResult, 0, len(uc.policies))
	var totalPurged int64

	for entityName, policy := range uc.policies {
		if !policy.Enabled {
			log.Info("Skipping disabled policy", "entity", entityName)
			continue
		}

		result, err := uc.PurgeEntity(ctx, entityName, dryRun)
		if err != nil {
			// Continue with other entities even if one fails
			log.Error("Failed to purge entity", "entity", entityName, "error", err)
			if result != nil {
				results = append(results, *result)
			}
			continue
		}

		results = append(results, *result)
		totalPurged += result.RecordsPurged
	}

	duration := time.Since(startTime)

	summary := &entity.PurgeSummary{
		TotalRecordsPurged: totalPurged,
		Results:            results,
		Duration:           duration,
		DryRun:             dryRun,
		Timestamp:          time.Now(),
	}

	log.Info("Purge all operation completed",
		"total_records_purged", totalPurged,
		"duration", duration,
		"dry_run", dryRun,
	)

	// Publish completion event
	completedEvent := event.NewPurgeCompletedEvent(*summary)
	if err := uc.eventBus.Publish(ctx, bus.TopicPurgeCompleted, completedEvent); err != nil {
		log.Error("Failed to publish purge completed event", "error", err)
	}

	return summary, nil
}

func (uc *purgeUseCase) GetRetentionPolicies(ctx context.Context) ([]entity.RetentionPolicy, error) {
	policies := make([]entity.RetentionPolicy, 0, len(uc.policies))
	for _, policy := range uc.policies {
		policies = append(policies, policy)
	}
	return policies, nil
}

func (uc *purgeUseCase) GetRetentionPolicy(ctx context.Context, entityName string) (*entity.RetentionPolicy, error) {
	policy, exists := uc.policies[entityName]
	if !exists {
		return nil, ErrPolicyNotFound
	}
	return &policy, nil
}

func (uc *purgeUseCase) PreviewPurge(ctx context.Context, entityName string) (int64, error) {
	log := logger.FromContext(ctx)

	policy, exists := uc.policies[entityName]
	if !exists {
		return 0, ErrPolicyNotFound
	}

	cutoffDate := policy.GetCutoffDate()

	var count int64
	var err error

	switch entityName {
	case "user_posts":
		count, err = uc.purgeRepo.CountDeletableUserPosts(ctx, cutoffDate)
	case "post_comments":
		count, err = uc.purgeRepo.CountDeletablePostComments(ctx, cutoffDate)
	default:
		return 0, ErrPolicyNotFound
	}

	if err != nil {
		log.Error("Failed to preview purge", "entity", entityName, "error", err)
		return 0, err
	}

	return count, nil
}
