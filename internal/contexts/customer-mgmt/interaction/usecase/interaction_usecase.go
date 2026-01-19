package usecase

import (
	"context"
	"time"

	interactionerrors "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/aggregate"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/repository"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IInteractionUseCase defines the interface for interaction use cases
type IInteractionUseCase interface {
	// CreateInteraction creates a new customer interaction
	CreateInteraction(
		ctx context.Context,
		customerID uuidv7.UUID,
		companyID *uuidv7.UUID,
		interactionType string,
		direction string,
		subject string,
		description string,
		createdBy uuidv7.UUID,
		startedAt time.Time,
	) (*aggregate.Interaction, error)

	// GetInteraction retrieves an interaction by ID
	GetInteraction(ctx context.Context, id uuidv7.UUID) (*aggregate.Interaction, error)

	// UpdateContent updates interaction content
	UpdateContent(ctx context.Context, id uuidv7.UUID, subject, description string) (*aggregate.Interaction, error)

	// SetOutcome sets the outcome of an interaction
	SetOutcome(ctx context.Context, id uuidv7.UUID, outcome string) (*aggregate.Interaction, error)

	// EndInteraction marks an interaction as ended
	EndInteraction(ctx context.Context, id uuidv7.UUID, endedAt time.Time) (*aggregate.Interaction, error)

	// SetFollowUp configures follow-up for an interaction
	SetFollowUp(ctx context.Context, id uuidv7.UUID, required bool, followUpDate *time.Time, notes string) (*aggregate.Interaction, error)

	// AddAttendee adds a participant to an interaction
	AddAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*aggregate.Interaction, error)

	// RemoveAttendee removes a participant from an interaction
	RemoveAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*aggregate.Interaction, error)

	// ListByCustomer lists all interactions for a customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListByCompany lists all interactions for a company
	ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListByType lists interactions by type
	ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListByCreatedBy lists interactions created by a user
	ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// ListPendingFollowUps lists interactions with pending follow-ups
	ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*aggregate.Interaction, int64, error)

	// DeleteInteraction soft deletes an interaction
	DeleteInteraction(ctx context.Context, id uuidv7.UUID) error
}

// InteractionUseCase implements IInteractionUseCase
type InteractionUseCase struct {
	repo repository.IInteractionRepository
}

// NewUseCase creates a new interaction use case
func NewInteractionUseCase(repo repository.IInteractionRepository) IInteractionUseCase {
	return &InteractionUseCase{
		repo: repo,
	}
}

// CreateInteraction creates a new customer interaction
func (uc *InteractionUseCase) CreateInteraction(
	ctx context.Context,
	customerID uuidv7.UUID,
	companyID *uuidv7.UUID,
	interactionType string,
	direction string,
	subject string,
	description string,
	createdBy uuidv7.UUID,
	startedAt time.Time,
) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	// Create interaction entity
	interaction, err := aggregate.NewInteraction(
		customerID,
		companyID,
		aggregate.InteractionType(interactionType),
		aggregate.InteractionDirection(direction),
		subject,
		description,
		createdBy,
		startedAt,
	)
	if err != nil {
		log.Error("Failed to create interaction entity", "error", err)
		return nil, interactionerrors.ErrInteractionCreateFailed
	}

	// Validate
	if err := interaction.Validate(); err != nil {
		log.Error("Interaction validation failed", "error", err)
		return nil, interactionerrors.ErrInteractionValidationFailed
	}

	// Save to repository
	if err := uc.repo.Create(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "error", err)
		return nil, interactionerrors.ErrInteractionCreateFailed
	}

	log.Info("Interaction created", "id", interaction.ID, "type", interaction.Type, "customer_id", customerID)

	return interaction, nil
}

// GetInteraction retrieves an interaction by ID
func (uc *InteractionUseCase) GetInteraction(ctx context.Context, id uuidv7.UUID) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err // Return repository error (interactionerrors.ErrInteractionNotFound)
	}

	return interaction, nil
}

// UpdateContent updates interaction content
func (uc *InteractionUseCase) UpdateContent(ctx context.Context, id uuidv7.UUID, subject, description string) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if err := interaction.UpdateContent(subject, description); err != nil {
		log.Error("Failed to update interaction content", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	log.Info("Interaction content updated", "id", id)

	return interaction, nil
}

// SetOutcome sets the outcome of an interaction
func (uc *InteractionUseCase) SetOutcome(ctx context.Context, id uuidv7.UUID, outcome string) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if err := interaction.SetOutcome(aggregate.InteractionOutcome(outcome)); err != nil {
		log.Error("Failed to set interaction outcome", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	log.Info("Interaction outcome set", "id", id, "outcome", outcome)

	return interaction, nil
}

// EndInteraction marks an interaction as ended
func (uc *InteractionUseCase) EndInteraction(ctx context.Context, id uuidv7.UUID, endedAt time.Time) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if interaction.EndedAt != nil {
		return nil, interactionerrors.ErrInteractionAlreadyEnded
	}

	if err := interaction.EndInteraction(endedAt); err != nil {
		log.Error("Failed to end interaction", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	log.Info("Interaction ended", "id", id, "duration_sec", interaction.DurationSec)

	return interaction, nil
}

// SetFollowUp configures follow-up for an interaction
func (uc *InteractionUseCase) SetFollowUp(ctx context.Context, id uuidv7.UUID, required bool, followUpDate *time.Time, notes string) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if err := interaction.SetFollowUp(required, followUpDate, notes); err != nil {
		log.Error("Failed to set follow-up", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	log.Info("Interaction follow-up set", "id", id, "required", required)

	return interaction, nil
}

// AddAttendee adds a participant to an interaction
func (uc *InteractionUseCase) AddAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	interaction.AddAttendee(attendeeID)

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	log.Info("Attendee added to interaction", "id", id, "attendee_id", attendeeID)

	return interaction, nil
}

// RemoveAttendee removes a participant from an interaction
func (uc *InteractionUseCase) RemoveAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*aggregate.Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	interaction.RemoveAttendee(attendeeID)

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, interactionerrors.ErrInteractionUpdateFailed
	}

	log.Info("Attendee removed from interaction", "id", id, "attendee_id", attendeeID)

	return interaction, nil
}

// ListByCustomer lists all interactions for a customer
func (uc *InteractionUseCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListByCustomer(ctx, customerID, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by customer", "customer_id", customerID, "error", err)
		return nil, 0, interactionerrors.ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListByCompany lists all interactions for a company
func (uc *InteractionUseCase) ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListByCompany(ctx, companyID, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by company", "company_id", companyID, "error", err)
		return nil, 0, interactionerrors.ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListByType lists interactions by type
func (uc *InteractionUseCase) ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	log := logger.FromContext(ctx)

	// Validate interaction type
	if interactionType == "" {
		return nil, 0, interactionerrors.ErrInvalidInteractionType
	}

	// Validate interaction type
	if err := validateInteractionType(interactionType); err != nil {
		return nil, 0, err
	}

	interactions, total, err := uc.repo.ListByType(ctx, interactionType, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by type", "type", interactionType, "error", err)
		return nil, 0, interactionerrors.ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListByCreatedBy lists interactions created by a user
func (uc *InteractionUseCase) ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListByCreatedBy(ctx, createdBy, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by created_by", "created_by", createdBy, "error", err)
		return nil, 0, interactionerrors.ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListPendingFollowUps lists interactions with pending follow-ups
func (uc *InteractionUseCase) ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListPendingFollowUps(ctx, page, pageSize)
	if err != nil {
		log.Error("Failed to list pending follow-ups", "error", err)
		return nil, 0, interactionerrors.ErrInteractionListFailed
	}

	return interactions, total, nil
}

// validateInteractionType checks if the interaction type is valid
func validateInteractionType(iType string) error {
	validTypes := []string{
		string(aggregate.InteractionTypeCall),
		string(aggregate.InteractionTypeEmail),
		string(aggregate.InteractionTypeMeeting),
		string(aggregate.InteractionTypeNote),
		string(aggregate.InteractionTypeSMS),
		string(aggregate.InteractionTypeChat),
	}
	for _, valid := range validTypes {
		if iType == valid {
			return nil
		}
	}
	return interactionerrors.ErrInvalidInteractionType
}

// DeleteInteraction soft deletes an interaction
func (uc *InteractionUseCase) DeleteInteraction(ctx context.Context, id uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	if err := uc.repo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete interaction", "id", id, "error", err)
		return err // Return repository error (interactionerrors.ErrInteractionNotFound)
	}

	log.Info("Interaction deleted", "id", id)

	return nil
}
