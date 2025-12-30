package language

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
	// cacheTTL is the cache TTL for language data (1 hour for reference data)
	cacheTTL = 1 * time.Hour
)

// IUseCase defines business logic for language operations
type IUseCase interface {
	GetByID(ctx context.Context, id uuidv7.UUID) (*Language, error)
	GetByCode(ctx context.Context, code string) (*Language, error)
	List(ctx context.Context) ([]*Language, error)
	Create(ctx context.Context, language *Language) error
	Update(ctx context.Context, language *Language) error
	Delete(ctx context.Context, id uuidv7.UUID) error
}

type useCase struct {
	repo  IRepository
	cache cache.ICache
}

// NewUseCase creates a new Language use case
func NewUseCase(repo IRepository, cacheClient cache.ICache) IUseCase {
	return &useCase{
		repo:  repo,
		cache: cacheClient,
	}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Language, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("language:id:%s", id.String())
	var language Language
	err := uc.cache.Get(ctx, cacheKey, &language)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &language, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	language_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, language_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache language", slog.Any("error", err))
	}
	
	return language_, nil
}

func (uc *useCase) GetByCode(ctx context.Context, code string) (*Language, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("language:code:%s", code)
	var language Language
	err := uc.cache.Get(ctx, cacheKey, &language)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return &language, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	language_, err := uc.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, language_, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache language", slog.Any("error", err))
	}
	
	return language_, nil
}

func (uc *useCase) List(ctx context.Context) ([]*Language, error) {
	// Try cache first
	cacheKey := "language:list:all"
	var languages []*Language
	err := uc.cache.Get(ctx, cacheKey, &languages)
	
	if err == nil {
		logger.FromContext(ctx).Debug("Cache hit", slog.String("key", cacheKey))
		return languages, nil
	}
	
	if !cache.IsCacheMiss(err) {
		logger.FromContext(ctx).Error("Cache error", slog.Any("error", err))
	}
	
	// Cache miss - fetch from database
	logger.FromContext(ctx).Debug("Cache miss", slog.String("key", cacheKey))
	languages, err = uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	
	// Store in cache
	if err := uc.cache.Set(ctx, cacheKey, languages, cacheTTL); err != nil {
		logger.FromContext(ctx).Error("Failed to cache languages", slog.Any("error", err))
	}
	
	return languages, nil
}

func (uc *useCase) Create(ctx context.Context, language *Language) error {
	if err := language.Validate(); err != nil {
		return err
	}
	
	if err := uc.repo.Create(ctx, language); err != nil {
		return err
	}
	
	// Invalidate list cache
	if err := uc.cache.Delete(ctx, "language:list:all"); err != nil {
		logger.FromContext(ctx).Error("Failed to invalidate cache", slog.Any("error", err))
	}
	
	return nil
}

func (uc *useCase) Update(ctx context.Context, language *Language) error {
	if err := language.Validate(); err != nil {
		return err
	}
	
	if err := uc.repo.Update(ctx, language); err != nil {
		return err
	}
	
	// Invalidate cache for this language
	_ = uc.cache.Delete(ctx, fmt.Sprintf("language:id:%s", language.ID.String()))
	_ = uc.cache.Delete(ctx, fmt.Sprintf("language:code:%s", language.Code))
	_ = uc.cache.Delete(ctx, "language:list:all")
	
	return nil
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	// Fetch language first to get code for cache invalidation
	language, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	
	// Invalidate cache
	_ = uc.cache.Delete(ctx, fmt.Sprintf("language:id:%s", id.String()))
	_ = uc.cache.Delete(ctx, fmt.Sprintf("language:code:%s", language.Code))
	_ = uc.cache.Delete(ctx, "language:list:all")
	
	return nil
}
