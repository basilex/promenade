package timezone

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
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
	repo IRepository
}

// NewUseCase creates a new timezone use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Timezone, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *useCase) GetByName(ctx context.Context, name string) (*Timezone, error) {
	return uc.repo.GetByName(ctx, name)
}

func (uc *useCase) List(ctx context.Context) ([]*Timezone, error) {
	return uc.repo.List(ctx)
}

func (uc *useCase) Create(ctx context.Context, timezone *Timezone) error {
	if err := timezone.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, timezone)
}

func (uc *useCase) Update(ctx context.Context, timezone *Timezone) error {
	if err := timezone.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, timezone)
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.repo.Delete(ctx, id)
}
