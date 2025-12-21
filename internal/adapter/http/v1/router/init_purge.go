package router

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/infrastructure/config"
	"github.com/basilex/promenade/internal/infrastructure/scheduler"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/logger"
)

// InitPurgeModule initializes the purge system with all dependencies and starts the scheduler
// Returns the use case and scheduler for admin module and graceful shutdown management
func InitPurgeModule(
	db *sqlx.DB,
	purgeConfig config.PurgeConfig,
	eventBus bus.Bus,
) (usecase.PurgeUseCase, *scheduler.Scheduler, error) {
	// Repository layer
	purgeRepo := postgres.NewPurgeRepository(db)

	// Define retention policies based on configuration
	purgePolicies := []entity.RetentionPolicy{
		{
			EntityName:    "user_posts",
			RetentionDays: purgeConfig.RetentionDaysUserPosts,
			Enabled:       purgeConfig.Enabled,
		},
		{
			EntityName:    "post_comments",
			RetentionDays: purgeConfig.RetentionDaysPostComments,
			Enabled:       purgeConfig.Enabled,
		},
	}

	// Use case layer
	purgeUseCase := usecase.NewPurgeUseCase(purgeRepo, purgePolicies, purgeConfig.BatchSize, eventBus)

	// Scheduler layer
	purgeScheduler := scheduler.NewScheduler(purgeUseCase, purgeConfig.Schedule, purgeConfig.DryRun, purgeConfig.Enabled)

	// Start scheduler
	if err := purgeScheduler.Start(context.Background()); err != nil {
		logger.Error("Failed to start purge scheduler", "error", err)
		return nil, nil, err
	}

	return purgeUseCase, purgeScheduler, nil
}
