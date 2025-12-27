package language

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
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
	repo IRepository
}

// NewUseCase creates a new language use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) GetByID(ctx context.Context, id uuidv7.UUID) (*Language, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *useCase) GetByCode(ctx context.Context, code string) (*Language, error) {
	return uc.repo.GetByCode(ctx, code)
}

func (uc *useCase) List(ctx context.Context) ([]*Language, error) {
	return uc.repo.List(ctx)
}

func (uc *useCase) Create(ctx context.Context, language *Language) error {
	if err := language.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, language)
}

func (uc *useCase) Update(ctx context.Context, language *Language) error {
	if err := language.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, language)
}

func (uc *useCase) Delete(ctx context.Context, id uuidv7.UUID) error {
	return uc.repo.Delete(ctx, id)
}
