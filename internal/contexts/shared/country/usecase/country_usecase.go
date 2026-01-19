package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/contexts/shared/country/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/country/repository"
	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines business logic for country operations
type ICountryUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Country, error)
	GetByCode(ctx context.Context, code string) (*aggregate.Country, error)
	List(ctx context.Context) ([]*aggregate.Country, error)
	Create(ctx context.Context, country *aggregate.Country) error
	Update(ctx context.Context, country *aggregate.Country) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type CountryUseCase struct {
	repo  repository.ICountryRepository
	cache cache.ICache
}

// NewUseCase creates a new Country use case
func NewCountryUseCase(repo repository.ICountryRepository, cacheClient cache.ICache) ICountryUseCase {
	return &CountryUseCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (u *CountryUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Country, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("country:id:%s", id.String())
	var country aggregate.Country
	err := u.cache.Get(ctx, cacheKey, &country)
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return &country, nil
	}

	// Cache miss - fetch from DB
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	entity, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache with 1-hour TTL
	if err := u.cache.Set(ctx, cacheKey, entity, 1*time.Hour); err != nil {
		logger.FromContext(ctx).Error("Failed to set cache", "error", err)
		// Continue even if cache fails
	}

	return entity, nil
}

func (u *CountryUseCase) GetByCode(ctx context.Context, code string) (*aggregate.Country, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("country:code:%s", code)
	var country aggregate.Country
	err := u.cache.Get(ctx, cacheKey, &country)
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return &country, nil
	}

	// Cache miss - fetch from DB
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	entity, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Store in cache with 1-hour TTL
	if err := u.cache.Set(ctx, cacheKey, entity, 1*time.Hour); err != nil {
		logger.FromContext(ctx).Error("Failed to set cache", "error", err)
	}

	return entity, nil
}

func (u *CountryUseCase) List(ctx context.Context) ([]*aggregate.Country, error) {
	// Try cache first
	cacheKey := "country:list:all"
	var countries []*aggregate.Country
	err := u.cache.Get(ctx, cacheKey, &countries)
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return countries, nil
	}

	// Cache miss - fetch from DB
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	countries, err = u.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache with 1-hour TTL
	if err := u.cache.Set(ctx, cacheKey, countries, 1*time.Hour); err != nil {
		logger.FromContext(ctx).Error("Failed to set cache", "error", err)
	}

	return countries, nil
}

func (u *CountryUseCase) Create(ctx context.Context, country *aggregate.Country) error {
	if err := country.Validate(); err != nil {
		return err
	}

	if err := u.repo.Create(ctx, country); err != nil {
		return err
	}

	// Invalidate list cache
	if err := u.cache.Delete(ctx, "country:list:all"); err != nil {
		logger.FromContext(ctx).Error("Failed to invalidate cache", "error", err)
	}

	return nil
}

func (u *CountryUseCase) Update(ctx context.Context, country *aggregate.Country) error {
	if err := country.Validate(); err != nil {
		return err
	}

	// Update timestamp
	country.Touch()

	if err := u.repo.Update(ctx, country); err != nil {
		return err
	}

	// Invalidate caches
	_ = u.cache.Delete(ctx, fmt.Sprintf("country:id:%s", country.ID.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("country:code:%s", country.Code))
	_ = u.cache.Delete(ctx, "country:list:all")

	return nil
}

func (u *CountryUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch country first to get code for cache invalidation
	country, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	_ = u.cache.Delete(ctx, fmt.Sprintf("country:id:%s", id.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("country:code:%s", country.Code))
	_ = u.cache.Delete(ctx, "country:list:all")

	return nil
}
