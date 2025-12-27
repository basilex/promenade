package currency

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
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
	repo IRepository
}

// NewUseCase creates a new currency use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Currency, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *useCase) GetByCode(ctx context.Context, code string) (*Currency, error) {
	return uc.repo.GetByCode(ctx, code)
}

func (uc *useCase) List(ctx context.Context) ([]*Currency, error) {
	return uc.repo.List(ctx)
}

func (uc *useCase) Create(ctx context.Context, currency *Currency) error {
	if err := currency.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, currency)
}

func (uc *useCase) Update(ctx context.Context, currency *Currency) error {
	if err := currency.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, currency)
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.repo.Delete(ctx, id)
}
