package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/basilex/promenade/internal/contexts/shared/timezone/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/timezone/repository"
	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

const (
	// cacheTTL is the cache TTL for timezone data (1 hour for reference data)
	cacheTTL = 1 * time.Hour
)

// ITimezoneUseCase defines business logic for timezone operations
type ITimezoneUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Timezone, error)
	GetByName(ctx context.Context, name string) (*aggregate.Timezone, error)
	List(ctx context.Context) ([]*aggregate.Timezone, error)
	Create(ctx context.Context, timezone *aggregate.Timezone) error
	Update(ctx context.Context, timezone *aggregate.Timezone) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type TimezoneUseCase struct {
	repo  repository.ITimezoneRepository
	cache cache.ICache
}

// NewTimezoneUseCase creates a new Timezone use case
func NewTimezoneUseCase(repo repository.ITimezoneRepository, cacheClient cache.ICache) ITimezoneUseCase {
	return &TimezoneUseCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (u *TimezoneUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Timezone, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("timezone:id:%s", id.String())
	var timezone aggregate.Timezone
	err := u.cache.Get(ctx, cacheKey, &timezone)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &timezone, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	timezone_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, timezone_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache timezone", slog.Any("error", err))
	}

	return timezone_, nil
}

func (u *TimezoneUseCase) GetByName(ctx context.Context, name string) (*aggregate.Timezone, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("timezone:name:%s", name)
	var timezone aggregate.Timezone
	err := u.cache.Get(ctx, cacheKey, &timezone)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &timezone, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	timezone_, err := u.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, timezone_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache timezone", slog.Any("error", err))
	}

	return timezone_, nil
}

func (u *TimezoneUseCase) List(ctx context.Context) ([]*aggregate.Timezone, error) {
	// Try cache first
	cacheKey := "timezone:list:all"
	var timezones []*aggregate.Timezone
	err := u.cache.Get(ctx, cacheKey, &timezones)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return timezones, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	timezones, err = u.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, timezones, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache timezones", slog.Any("error", err))
	}

	return timezones, nil
}

func (u *TimezoneUseCase) Create(ctx context.Context, timezone *aggregate.Timezone) error {
	if err := timezone.Validate(); err != nil {
		return err
	}

	if err := u.repo.Create(ctx, timezone); err != nil {
		return err
	}

	// Invalidate list cache
	_ = u.cache.Delete(ctx, "timezone:list:all")

	return nil
}

func (u *TimezoneUseCase) Update(ctx context.Context, timezone *aggregate.Timezone) error {
	if err := timezone.Validate(); err != nil {
		return err
	}
	// Update timestamp
	timezone.Touch()
	if err := u.repo.Update(ctx, timezone); err != nil {
		return err
	}

	// Invalidate cache for this timezone
	_ = u.cache.Delete(ctx, fmt.Sprintf("timezone:id:%s", timezone.ID.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("timezone:name:%s", timezone.Name))
	_ = u.cache.Delete(ctx, "timezone:list:all")

	return nil
}

func (u *TimezoneUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch timezone first to get name for cache invalidation
	timezone, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	_ = u.cache.Delete(ctx, fmt.Sprintf("timezone:id:%s", id.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("timezone:name:%s", timezone.Name))
	_ = u.cache.Delete(ctx, "timezone:list:all")

	return nil
}
