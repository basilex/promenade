package currency

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
	// cacheTTL is the cache TTL for currency data (1 hour for reference data)
	cacheTTL = 1 * time.Hour
)

// IUseCase defines business logic for currency operations
type IUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Currency, error)
	GetByCode(ctx context.Context, code string) (*Currency, error)
	List(ctx context.Context) ([]*Currency, error)
	Create(ctx context.Context, currency *Currency) error
	Update(ctx context.Context, currency *Currency) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
	repo  IRepository
	cache cache.ICache
}

// NewUseCase creates a new Currency use case
func NewUseCase(repo IRepository, cacheClient cache.ICache) IUseCase {
	return &useCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Currency, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("currency:id:%s", id.String())
	var currency Currency
	err := uc.cache.Get(ctx, cacheKey, &currency)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &currency, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	currency_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, currency_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache currency", slog.Any("error", err))
	}
	
	return currency_, nil
}

func (uc *useCase) GetByCode(ctx context.Context, code string) (*Currency, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("currency:code:%s", code)
	var currency Currency
	err := uc.cache.Get(ctx, cacheKey, &currency)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &currency, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	currency_, err := uc.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, currency_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache currency", slog.Any("error", err))
	}
	
	return currency_, nil
}

func (uc *useCase) List(ctx context.Context) ([]*Currency, error) {
	// Try cache first
	cacheKey := "currency:list:all"
	var currencies []*Currency
	err := uc.cache.Get(ctx, cacheKey, &currencies)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return currencies, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	currencies, err = uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, currencies, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache currencies", slog.Any("error", err))
	}
	
	return currencies, nil
}

func (uc *useCase) Create(ctx context.Context, currency *Currency) error {
	if err := currency.Validate(); err != nil {
		return err
	}
	
	if err := uc.repo.Create(ctx, currency); err != nil {
		return err
	}
	
	// Invalidate list cache
	_ = uc.cache.Delete(ctx, "currency:list:all")
	
	return nil
}

func (uc *useCase) Update(ctx context.Context, currency *Currency) error {
	if err := currency.Validate(); err != nil {
		return err
	}
	
	// Update timestamp
	currency.Touch()
	
	if err := uc.repo.Update(ctx, currency); err != nil {
		return err
	}
	
	// Invalidate cache for this currency
	_ = uc.cache.Delete(ctx, fmt.Sprintf("currency:id:%s", currency.ID.String()))
	_ = uc.cache.Delete(ctx, fmt.Sprintf("currency:code:%s", currency.Code))
	_ = uc.cache.Delete(ctx, "currency:list:all")
	
	return nil
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch currency first to get code for cache invalidation
	currency, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	
	// Invalidate cache (graceful degradation if cache unavailable)
	_ = uc.cache.Delete(ctx, fmt.Sprintf("currency:id:%s", id.String()))
	_ = uc.cache.Delete(ctx, fmt.Sprintf("currency:code:%s", currency.Code))
	_ = uc.cache.Delete(ctx, "currency:list:all")
	
	return nil
}
