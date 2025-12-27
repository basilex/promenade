package country

import (
	"context"

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
	repo IRepository
}

// NewUseCase creates a new country use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Country, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *useCase) GetByCode(ctx context.Context, code string) (*Country, error) {
	return uc.repo.GetByCode(ctx, code)
}

func (uc *useCase) List(ctx context.Context) ([]*Country, error) {
	return uc.repo.List(ctx)
}

func (uc *useCase) Create(ctx context.Context, country *Country) error {
	if err := country.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, country)
}

func (uc *useCase) Update(ctx context.Context, country *Country) error {
	if err := country.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, country)
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.repo.Delete(ctx, id)
}
