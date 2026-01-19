package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/contexts/shared/language/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/language/repository"
	"github.com/basilex/promenade/pkg/cache"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

const (
	// cacheTTL is the cache TTL for language data (1 hour for reference data)
	cacheTTL = 1 * time.Hour
)

// ILanguageUseCase defines business logic for language operations
type ILanguageUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Language, error)
	GetByCode(ctx context.Context, code string) (*aggregate.Language, error)
	List(ctx context.Context) ([]*aggregate.Language, error)
	Create(ctx context.Context, language *aggregate.Language) error
	Update(ctx context.Context, language *aggregate.Language) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type LanguageUseCase struct {
	repo  repository.ILanguageRepository
	cache cache.ICache
}

// NewUseCase creates a new Language use case
func NewLanguageUseCase(repo repository.ILanguageRepository, cacheClient cache.ICache) ILanguageUseCase {
	return &LanguageUseCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (u *LanguageUseCase) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Language, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("language:id:%s", id.String())
	var language aggregate.Language
	err := u.cache.Get(ctx, cacheKey, &language)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return &language, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", "key", cacheKey)
	language_, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, language_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache language", "error", err)
	}

	return language_, nil
}

func (u *LanguageUseCase) GetByCode(ctx context.Context, code string) (*aggregate.Language, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("language:code:%s", code)
	var language aggregate.Language
	err := u.cache.Get(ctx, cacheKey, &language)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return &language, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", "key", cacheKey)
	language_, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, language_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache language", "error", err)
	}

	return language_, nil
}

func (u *LanguageUseCase) List(ctx context.Context) ([]*aggregate.Language, error) {
	// Try cache first
	cacheKey := "language:list:all"
	var languages []*aggregate.Language
	err := u.cache.Get(ctx, cacheKey, &languages)

	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", "key", cacheKey)
		return languages, nil
	}

	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", "error", err)
	}

	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", "key", cacheKey)
	languages, err = u.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := u.cache.Set(ctx, cacheKey, languages, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache languages", "error", err)
	}

	return languages, nil
}

func (u *LanguageUseCase) Create(ctx context.Context, language *aggregate.Language) error {
	if err := language.Validate(); err != nil {
		return err
	}

	if err := u.repo.Create(ctx, language); err != nil {
		return err
	}

	// Invalidate list cache
	if err := u.cache.Delete(ctx, "language:list:all"); err != nil {
		logger.FromContext(ctx).Error("Failed to invalidate cache", "error", err)
	}

	return nil
}

func (u *LanguageUseCase) Update(ctx context.Context, language *aggregate.Language) error {
	if err := language.Validate(); err != nil {
		return err
	}
	// Update timestamp
	language.Touch()
	if err := u.repo.Update(ctx, language); err != nil {
		return err
	}

	// Invalidate cache for this language
	_ = u.cache.Delete(ctx, fmt.Sprintf("language:id:%s", language.ID.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("language:code:%s", language.Code))
	_ = u.cache.Delete(ctx, "language:list:all")

	return nil
}

func (u *LanguageUseCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch language first to get code for cache invalidation
	language, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	_ = u.cache.Delete(ctx, fmt.Sprintf("language:id:%s", id.String()))
	_ = u.cache.Delete(ctx, fmt.Sprintf("language:code:%s", language.Code))
	_ = u.cache.Delete(ctx, "language:list:all")

	return nil
}
