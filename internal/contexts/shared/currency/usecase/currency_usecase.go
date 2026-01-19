package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/basilex/promenade/internal/contexts/shared/currency/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/currency/repository"
	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

const (
	// cacheTTL is the cache TTL for currency data (1 hour for reference data)
	cacheTTL = 1 * time.Hour
)

// ICurrencyUseCase defines business logic for currency operations
type ICurrencyUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Currency, error)
	GetByCode(ctx context.Context, code string) (*aggregate.Currency, error)
	List(ctx context.Context) ([]*aggregate.Currency, error)
	Create(ctx context.Context, currency *aggregate.Currency) error
	Update(ctx context.Context, currency *aggregate.Currency) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type CurrencyUseCase struct {
	repo  repository.ICurrencyRepository
	cache cache.ICache
}

// NewCurrencyUseCase creates a new Currency use case
func NewCurrencyUseCase(repo repository.ICurrencyRepository, cacheClient cache.ICache) ICurrencyUseCase {
	return &CurrencyUseCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (u *CurrencyUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Currency, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("currency:id:%s", id.String())
	var currency aggregate.Currency
	err := u.cache.Get(ctx, cacheKey, &currency)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &currency, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	currency_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, currency_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache currency", slog.Any("error", err))
	}

	return currency_, nil
}

func (u *CurrencyUseCase) GetByCode(ctx context.Context, code string) (*aggregate.Currency, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("currency:code:%s", code)
	var currency aggregate.Currency
	err := u.cache.Get(ctx, cacheKey, &currency)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &currency, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	currency_, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, currency_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache currency", slog.Any("error", err))
	}

	return currency_, nil
}

func (u *CurrencyUseCase) List(ctx context.Context) ([]*aggregate.Currency, error) {
	// Try cache first
	cacheKey := "currency:list:all"
	var currencies []*aggregate.Currency
	err := u.cache.Get(ctx, cacheKey, &currencies)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return currencies, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	currencies, err = u.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, currencies, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache currencies", slog.Any("error", err))
	}

	return currencies, nil
}

func (u *CurrencyUseCase) Create(ctx context.Context, currency *aggregate.Currency) error {
	if err := currency.Validate(); err != nil {
		return err
	}

	if err := u.repo.Create(ctx, currency); err != nil {
		return err
	}

	// Invalidate list cache
	_ = u.cache.Delete(ctx, "currency:list:all")

	return nil
}

func (u *CurrencyUseCase) Update(ctx context.Context, currency *aggregate.Currency) error {
	if err := currency.Validate(); err != nil {
		return err
	}

	// Update timestamp
	currency.Touch()

	if err := u.repo.Update(ctx, currency); err != nil {
		return err
	}

	// Invalidate cache for this currency
	_ = u.cache.Delete(ctx, fmt.Sprintf("currency:id:%s", currency.ID.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("currency:code:%s", currency.Code))
	_ = u.cache.Delete(ctx, "currency:list:all")

	return nil
}

func (u *CurrencyUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch currency first to get code for cache invalidation
	currency, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache (graceful degradation if cache unavailable)
	_ = u.cache.Delete(ctx, fmt.Sprintf("currency:id:%s", id.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("currency:code:%s", currency.Code))
	_ = u.cache.Delete(ctx, "currency:list:all")

	return nil
}
