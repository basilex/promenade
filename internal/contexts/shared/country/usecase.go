package country

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines business logic for country operations
type IUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Country, error)
	GetByCode(ctx context.Context, code string) (*Country, error)
	List(ctx context.Context) ([]*Country, error)
	Create(ctx context.Context, country *Country) error
	Update(ctx context.Context, country *Country) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
	repo  IRepository
	cache cache.Cache
}

// NewUseCase creates a new country use case with cache support
func NewUseCase(repo IRepository, cacheClient cache.Cache) IUseCase {
	return &useCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Country, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("country:id:%s", id.String())
	var country Country
	err := uc.cache.Get(ctx, cacheKey, &country)
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return &country, nil
	}

	// Cache miss - fetch from DB
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	entity, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache with 1-hour TTL
	if err := uc.cache.Set(ctx, cacheKey, entity, 1*time.Hour); err != nil {
		logger.FromContext(ctx).Error("Failed to set cache", "error", err)
		// Continue even if cache fails
	}

	return entity, nil
}

func (uc *useCase) GetByCode(ctx context.Context, code string) (*Country, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("country:code:%s", code)
	var country Country
	err := uc.cache.Get(ctx, cacheKey, &country)
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return &country, nil
	}

	// Cache miss - fetch from DB
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	entity, err := uc.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Store in cache with 1-hour TTL
	if err := uc.cache.Set(ctx, cacheKey, entity, 1*time.Hour); err != nil {
		logger.FromContext(ctx).Error("Failed to set cache", "error", err)
	}

	return entity, nil
}

func (uc *useCase) List(ctx context.Context) ([]*Country, error) {
	// Try cache first
	cacheKey := "country:list:all"
	var countries []*Country
	err := uc.cache.Get(ctx, cacheKey, &countries)
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return countries, nil
	}

	// Cache miss - fetch from DB
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	countries, err = uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache with 1-hour TTL
	if err := uc.cache.Set(ctx, cacheKey, countries, 1*time.Hour); err != nil {
		logger.FromContext(ctx).Error("Failed to set cache", "error", err)
	}

	return countries, nil
}

func (uc *useCase) Create(ctx context.Context, country *Country) error {
	if err := country.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Create(ctx, country); err != nil {
		return err
	}

	// Invalidate list cache
	if err := uc.cache.Delete(ctx, "country:list:all"); err != nil {
		logger.FromContext(ctx).Error("Failed to invalidate cache", "error", err)
	}

	return nil
}

func (uc *useCase) Update(ctx context.Context, country *Country) error {
	if err := country.Validate(); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, country); err != nil {
		return err
	}

	// Invalidate caches
	uc.cache.Delete(ctx, fmt.Sprintf("country:id:%s", country.ID.String()))
	uc.cache.Delete(ctx, fmt.Sprintf("country:code:%s", country.Code))
	uc.cache.Delete(ctx, "country:list:all")

	return nil
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch country first to get code for cache invalidation
	country, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	uc.cache.Delete(ctx, fmt.Sprintf("country:id:%s", id.String()))
	uc.cache.Delete(ctx, fmt.Sprintf("country:code:%s", country.Code))
	uc.cache.Delete(ctx, "country:list:all")

	return nil
}