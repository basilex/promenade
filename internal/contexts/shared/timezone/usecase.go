package timezone

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

const (
	// cacheTTL is the cache TTL for timezone data (1 hour for reference data)
	cacheTTL = 1 * time.Hour
)

// IUseCase defines business logic for timezone operations
type IUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Timezone, error)
	GetByName(ctx context.Context, name string) (*Timezone, error)
	List(ctx context.Context) ([]*Timezone, error)
	Create(ctx context.Context, timezone *Timezone) error
	Update(ctx context.Context, timezone *Timezone) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
	repo  IRepository
	cache cache.Cache
}

// NewUseCase creates a new timezone use case with cache support
func NewUseCase(repo IRepository, cacheClient cache.Cache) IUseCase {
	return &useCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Timezone, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("timezone:id:%s", id.String())
	var timezone Timezone
	err := uc.cache.Get(ctx, cacheKey, &timezone)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &timezone, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	timezone_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, timezone_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache timezone", slog.Any("error", err))
	}
	
	return timezone_, nil
}

func (uc *useCase) GetByName(ctx context.Context, name string) (*Timezone, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("timezone:name:%s", name)
	var timezone Timezone
	err := uc.cache.Get(ctx, cacheKey, &timezone)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &timezone, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	timezone_, err := uc.repo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, timezone_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache timezone", slog.Any("error", err))
	}
	
	return timezone_, nil
}

func (uc *useCase) List(ctx context.Context) ([]*Timezone, error) {
	// Try cache first
	cacheKey := "timezone:list:all"
	var timezones []*Timezone
	err := uc.cache.Get(ctx, cacheKey, &timezones)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return timezones, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	timezones, err = uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, timezones, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache timezones", slog.Any("error", err))
	}
	
	return timezones, nil
}

func (uc *useCase) Create(ctx context.Context, timezone *Timezone) error {
	if err := timezone.Validate(); err != nil {
		return err
	}
	
	if err := uc.repo.Create(ctx, timezone); err != nil {
		return err
	}
	
	// Invalidate list cache
	_ = uc.cache.Delete(ctx, "timezone:list:all")
	
	return nil
}

func (uc *useCase) Update(ctx context.Context, timezone *Timezone) error {
	if err := timezone.Validate(); err != nil {
		return err
	}
	
	if err := uc.repo.Update(ctx, timezone); err != nil {
		return err
	}
	
	// Invalidate cache for this timezone
	_ = uc.cache.Delete(ctx, fmt.Sprintf("timezone:id:%s", timezone.ID.String()))
	_ = uc.cache.Delete(ctx, fmt.Sprintf("timezone:name:%s", timezone.Name))
	_ = uc.cache.Delete(ctx, "timezone:list:all")
	
	return nil
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch timezone first to get name for cache invalidation
	timezone, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	
	// Invalidate cache
	_ = uc.cache.Delete(ctx, fmt.Sprintf("timezone:id:%s", id.String()))
	_ = uc.cache.Delete(ctx, fmt.Sprintf("timezone:name:%s", timezone.Name))
	_ = uc.cache.Delete(ctx, "timezone:list:all")
	
	return nil
}
