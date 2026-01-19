package usecase

import (
	"context"
	"log/slog"

	formerrors "github.com/basilex/promenade/internal/contexts/ui/metadata/form"
	"github.com/basilex/promenade/internal/contexts/ui/metadata/form/aggregate"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IFormRepository defines persistence operations for form definitions
type IFormRepository interface {
	Create(ctx context.Context, form *aggregate.FormDefinition) error
	GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error)
	GetByFormID(ctx context.Context, formID string) (*aggregate.FormDefinition, error)
	Update(ctx context.Context, form *aggregate.FormDefinition) error
	Delete(ctx context.Context, id uuidv7.UUID) error
	List(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error)
}

// IFormUseCase defines business operations for UI form definitions.
type IFormUseCase interface {
	CreateForm(ctx context.Context, form *aggregate.FormDefinition) (*aggregate.FormDefinition, error)
	GetForm(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error)
	GetFormByFormID(ctx context.Context, formID string) (*aggregate.FormDefinition, error)
	UpdateForm(ctx context.Context, id uuidv7.UUID, update *aggregate.FormDefinition) (*aggregate.FormDefinition, error)
	DeleteForm(ctx context.Context, id uuidv7.UUID) error
	ListForms(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error)
	ListAllForms(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error)
}

type FormUseCase struct {
	repo IFormRepository
}

// NewFormUseCase creates a new form use case.
func NewFormUseCase(repo IFormRepository) IFormUseCase {
	return &FormUseCase{repo: repo}
}

// CreateForm creates a new form definition.
func (uc *FormUseCase) CreateForm(ctx context.Context, form *aggregate.FormDefinition) (*aggregate.FormDefinition, error) {
	log := logger.FromContext(ctx)

	if form == nil {
		return nil, formerrors.ErrFormCreateFailed
	}

	if err := form.Validate(); err != nil {
		return nil, err
	}

	// Ensure unique form_id
	existing, err := uc.repo.GetByFormID(ctx, form.FormID)
	if err == nil && existing != nil {
		return nil, formerrors.ErrFormIDExists
	}

	if err := uc.repo.Create(ctx, form); err != nil {
		log.Error("Failed to create form definition",
			slog.String("form_id", form.FormID),
			slog.Any("error", err),
		)
		return nil, formerrors.ErrFormCreateFailed
	}

	return form, nil
}

// GetForm retrieves a form definition by ID.
func (uc *FormUseCase) GetForm(ctx context.Context, id uuidv7.UUID) (*aggregate.FormDefinition, error) {
	form, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, formerrors.ErrFormNotFound
	}
	return form, nil
}

// GetFormByFormID retrieves a form definition by form_id.
func (uc *FormUseCase) GetFormByFormID(ctx context.Context, formID string) (*aggregate.FormDefinition, error) {
	form, err := uc.repo.GetByFormID(ctx, formID)
	if err != nil {
		return nil, formerrors.ErrFormNotFound
	}
	return form, nil
}

// UpdateForm updates a form definition.
func (uc *FormUseCase) UpdateForm(ctx context.Context, id uuidv7.UUID, update *aggregate.FormDefinition) (*aggregate.FormDefinition, error) {
	log := logger.FromContext(ctx)

	form, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, formerrors.ErrFormNotFound
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
		return nil, formerrors.ErrFormUpdateFailed
	}

	return form, nil
}

// DeleteForm soft-deletes a form definition.
func (uc *FormUseCase) DeleteForm(ctx context.Context, id uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	if err := uc.repo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete form definition",
			slog.String("form_id", id.String()),
			slog.Any("error", err),
		)
		return formerrors.ErrFormDeleteFailed
	}
	return nil
}

// ListForms lists forms filtered by entity type.
func (uc *FormUseCase) ListForms(ctx context.Context, entityType string, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
	forms, total, err := uc.repo.List(ctx, entityType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return forms, total, nil
}

// ListAllForms lists all forms with pagination.
func (uc *FormUseCase) ListAllForms(ctx context.Context, limit, offset int) ([]*aggregate.FormDefinition, int, error) {
	forms, total, err := uc.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return forms, total, nil
}
