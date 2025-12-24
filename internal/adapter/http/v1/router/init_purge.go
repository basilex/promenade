package router

import (
	"context"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/scheduler"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/purge"
)

// InitPurgeModule initializes the purge system with dynamic handler and policy registries
// Modules register their own purge handlers and retention policies via pkg/purge registries
// Core only manages the orchestration (scheduler, use case, infrastructure)
// Returns the use case and scheduler for admin module and graceful shutdown management
func InitPurgeModule(
	purgeConfig config.PurgeSection,
	eventBus bus.IBus,
) (usecase.IPurgeUseCase, *scheduler.Scheduler, error) {
	// Use the global purge handler registry
	// Modules will register their handlers during initialization
	handlerRegistry := purge.DefaultRegistry
	
	// Use the global policy registry
	// Modules will register their retention policies during initialization
	policyRegistry := purge.DefaultPolicyRegistry

	// Convert purge.RetentionPolicy to entity.RetentionPolicy
	// This is necessary because usecase works with domain entities
	purgeRetentionPolicies := policyRegistry.GetAllPolicies()
	purgePolicies := make([]entity.RetentionPolicy, 0, len(purgeRetentionPolicies))
	for _, policy := range purgeRetentionPolicies {
		purgePolicies = append(purgePolicies, entity.RetentionPolicy{
			EntityName:    policy.EntityName,
			RetentionDays: policy.RetentionDays,
			Enabled:       policy.Enabled && purgeConfig.Enabled, // Respect global enable flag
		})
	}

	// Use case layer
	purgeUseCase := usecase.NewPurgeUseCase(handlerRegistry, purgePolicies, purgeConfig.BatchSize, eventBus)

	// Scheduler layer
	purgeScheduler := scheduler.NewScheduler(purgeUseCase, purgeConfig.Schedule, purgeConfig.DryRun, purgeConfig.Enabled)

	// Start scheduler
	if err := purgeScheduler.Start(context.Background()); err != nil {
		logger.Error("Failed to start purge scheduler", "error", err)
		return nil, nil, err
	}

	return purgeUseCase, purgeScheduler, nil
}
