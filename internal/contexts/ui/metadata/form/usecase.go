package form

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines business operations for UI form definitions.
type IUseCase interface {
	CreateForm(ctx context.Context, form *FormDefinition) (*FormDefinition, error)
	GetForm(ctx context.Context, id uuidv7.UUID) (*FormDefinition, error)
	GetFormByFormID(ctx context.Context, formID string) (*FormDefinition, error)
	UpdateForm(ctx context.Context, id uuidv7.UUID, update *FormDefinition) (*FormDefinition, error)
	DeleteForm(ctx context.Context, id uuidv7.UUID) error
	ListForms(ctx context.Context, entityType string, limit, offset int) ([]*FormDefinition, int, error)
	ListAllForms(ctx context.Context, limit, offset int) ([]*FormDefinition, int, error)
}

type useCase struct {
	repo IRepository
}

// NewUseCase creates a new form use case.
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{repo: repo}
}

// CreateForm creates a new form definition.
func (uc *useCase) CreateForm(ctx context.Context, form *FormDefinition) (*FormDefinition, error) {
	log := logger.FromContext(ctx)

	if form == nil {
		return nil, ErrFormCreateFailed
	}

	if err := form.Validate(); err != nil {
		return nil, err
	}

	// Ensure unique form_id
	existing, err := uc.repo.GetByFormID(ctx, form.FormID)
	if err == nil && existing != nil {
		return nil, ErrFormIDExists
	}

	if err := uc.repo.Create(ctx, form); err != nil {
		log.Error("Failed to create form definition",
			slog.String("form_id", form.FormID),
			slog.Any("error", err),
		)
		return nil, ErrFormCreateFailed
	}

	return form, nil
}

// GetForm retrieves a form definition by ID.
func (uc *useCase) GetForm(ctx context.Context, id uuidv7.UUID) (*FormDefinition, error) {
	form, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFormNotFound
	}
	return form, nil
}

// GetFormByFormID retrieves a form definition by form_id.
func (uc *useCase) GetFormByFormID(ctx context.Context, formID string) (*FormDefinition, error) {
	form, err := uc.repo.GetByFormID(ctx, formID)
	if err != nil {
		return nil, ErrFormNotFound
	}
	return form, nil
}

// UpdateForm updates a form definition.
func (uc *useCase) UpdateForm(ctx context.Context, id uuidv7.UUID, update *FormDefinition) (*FormDefinition, error) {
	log := logger.FromContext(ctx)

	form, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrFormNotFound
	}

	err = form.UpdateMetadata(
		update.Name,
		update.Description,
		update.Layout.Get(),
		update.Fields.Get(),
		update.Validation.Get(),
		update.Events.Get(),
		update.Permissions.Get(),
		update.I18n.Get(),
		update.IsActive,
		update.EntityType,
		update.TenantID,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, form); err != nil {
		log.Error("Failed to update form definition",
			slog.String("form_id", form.FormID),
			slog.Any("error", err),
		)
		return nil, ErrFormUpdateFailed
	}

	return form, nil
}

// DeleteForm soft-deletes a form definition.
func (uc *useCase) DeleteForm(ctx context.Context, id uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	if err := uc.repo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete form definition",
			slog.String("form_id", id.String()),
			slog.Any("error", err),
		)
		return ErrFormDeleteFailed
	}
	return nil
}

// ListForms lists forms filtered by entity type.
func (uc *useCase) ListForms(ctx context.Context, entityType string, limit, offset int) ([]*FormDefinition, int, error) {
	forms, total, err := uc.repo.List(ctx, entityType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return forms, total, nil
}

// ListAllForms lists all forms with pagination.
func (uc *useCase) ListAllForms(ctx context.Context, limit, offset int) ([]*FormDefinition, int, error) {
	forms, total, err := uc.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return forms, total, nil
}
