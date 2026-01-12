package interaction

import (
	"context"
	"time"

	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the interface for interaction use cases
type IUseCase interface {
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
	) (*Interaction, error)

	// GetInteraction retrieves an interaction by ID
	GetInteraction(ctx context.Context, id uuidv7.UUID) (*Interaction, error)

	// UpdateContent updates interaction content
	UpdateContent(ctx context.Context, id uuidv7.UUID, subject, description string) (*Interaction, error)

	// SetOutcome sets the outcome of an interaction
	SetOutcome(ctx context.Context, id uuidv7.UUID, outcome string) (*Interaction, error)

	// EndInteraction marks an interaction as ended
	EndInteraction(ctx context.Context, id uuidv7.UUID, endedAt time.Time) (*Interaction, error)

	// SetFollowUp configures follow-up for an interaction
	SetFollowUp(ctx context.Context, id uuidv7.UUID, required bool, followUpDate *time.Time, notes string) (*Interaction, error)

	// AddAttendee adds a participant to an interaction
	AddAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*Interaction, error)

	// RemoveAttendee removes a participant from an interaction
	RemoveAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*Interaction, error)

	// ListByCustomer lists all interactions for a customer
	ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error)

	// ListByCompany lists all interactions for a company
	ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error)

	// ListByType lists interactions by type
	ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*Interaction, int64, error)

	// ListByCreatedBy lists interactions created by a user
	ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error)

	// ListPendingFollowUps lists interactions with pending follow-ups
	ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*Interaction, int64, error)

	// DeleteInteraction soft deletes an interaction
	DeleteInteraction(ctx context.Context, id uuidv7.UUID) error
}

// useCase implements IUseCase
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new interaction use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateInteraction creates a new customer interaction
func (uc *useCase) CreateInteraction(
	ctx context.Context,
	customerID uuidv7.UUID,
	companyID *uuidv7.UUID,
	interactionType string,
	direction string,
	subject string,
	description string,
	createdBy uuidv7.UUID,
	startedAt time.Time,
) (*Interaction, error) {
	log := logger.FromContext(ctx)

	// Create interaction entity
	interaction, err := NewInteraction(
		customerID,
		companyID,
		InteractionType(interactionType),
		InteractionDirection(direction),
		subject,
		description,
		createdBy,
		startedAt,
	)
	if err != nil {
		log.Error("Failed to create interaction entity", "error", err)
		return nil, ErrInteractionCreateFailed
	}

	// Validate
	if err := interaction.Validate(); err != nil {
		log.Error("Interaction validation failed", "error", err)
		return nil, ErrInteractionValidationFailed
	}

	// Save to repository
	if err := uc.repo.Create(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "error", err)
		return nil, ErrInteractionCreateFailed
	}

	log.Info("Interaction created", "id", interaction.ID, "type", interaction.Type, "customer_id", customerID)

	return interaction, nil
}

// GetInteraction retrieves an interaction by ID
func (uc *useCase) GetInteraction(ctx context.Context, id uuidv7.UUID) (*Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err // Return repository error (ErrInteractionNotFound)
	}

	return interaction, nil
}

// UpdateContent updates interaction content
func (uc *useCase) UpdateContent(ctx context.Context, id uuidv7.UUID, subject, description string) (*Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if err := interaction.UpdateContent(subject, description); err != nil {
		log.Error("Failed to update interaction content", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	log.Info("Interaction content updated", "id", id)

	return interaction, nil
}

// SetOutcome sets the outcome of an interaction
func (uc *useCase) SetOutcome(ctx context.Context, id uuidv7.UUID, outcome string) (*Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if err := interaction.SetOutcome(InteractionOutcome(outcome)); err != nil {
		log.Error("Failed to set interaction outcome", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	log.Info("Interaction outcome set", "id", id, "outcome", outcome)

	return interaction, nil
}

// EndInteraction marks an interaction as ended
func (uc *useCase) EndInteraction(ctx context.Context, id uuidv7.UUID, endedAt time.Time) (*Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if interaction.EndedAt != nil {
		return nil, ErrInteractionAlreadyEnded
	}

	if err := interaction.EndInteraction(endedAt); err != nil {
		log.Error("Failed to end interaction", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	log.Info("Interaction ended", "id", id, "duration_sec", interaction.DurationSec)

	return interaction, nil
}

// SetFollowUp configures follow-up for an interaction
func (uc *useCase) SetFollowUp(ctx context.Context, id uuidv7.UUID, required bool, followUpDate *time.Time, notes string) (*Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	if err := interaction.SetFollowUp(required, followUpDate, notes); err != nil {
		log.Error("Failed to set follow-up", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	log.Info("Interaction follow-up set", "id", id, "required", required)

	return interaction, nil
}

// AddAttendee adds a participant to an interaction
func (uc *useCase) AddAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	interaction.AddAttendee(attendeeID)

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	log.Info("Attendee added to interaction", "id", id, "attendee_id", attendeeID)

	return interaction, nil
}

// RemoveAttendee removes a participant from an interaction
func (uc *useCase) RemoveAttendee(ctx context.Context, id uuidv7.UUID, attendeeID uuidv7.UUID) (*Interaction, error) {
	log := logger.FromContext(ctx)

	interaction, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		log.Error("Failed to get interaction", "id", id, "error", err)
		return nil, err
	}

	interaction.RemoveAttendee(attendeeID)

	if err := uc.repo.Update(ctx, interaction); err != nil {
		log.Error("Failed to save interaction", "id", id, "error", err)
		return nil, ErrInteractionUpdateFailed
	}

	log.Info("Attendee removed from interaction", "id", id, "attendee_id", attendeeID)

	return interaction, nil
}

// ListByCustomer lists all interactions for a customer
func (uc *useCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListByCustomer(ctx, customerID, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by customer", "customer_id", customerID, "error", err)
		return nil, 0, ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListByCompany lists all interactions for a company
func (uc *useCase) ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListByCompany(ctx, companyID, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by company", "company_id", companyID, "error", err)
		return nil, 0, ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListByType lists interactions by type
func (uc *useCase) ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*Interaction, int64, error) {
	log := logger.FromContext(ctx)

	iType := InteractionType(interactionType)
	if !isValidInteractionType(iType) {
		return nil, 0, ErrInvalidInteractionType
	}

interactions, total, err := uc.repo.ListByType(ctx, interactionType, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by type", "type", interactionType, "error", err)
		return nil, 0, ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListByCreatedBy lists interactions created by a user
func (uc *useCase) ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListByCreatedBy(ctx, createdBy, page, pageSize)
	if err != nil {
		log.Error("Failed to list interactions by created_by", "created_by", createdBy, "error", err)
		return nil, 0, ErrInteractionListFailed
	}

	return interactions, total, nil
}

// ListPendingFollowUps lists interactions with pending follow-ups
func (uc *useCase) ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*Interaction, int64, error) {
	log := logger.FromContext(ctx)

	interactions, total, err := uc.repo.ListPendingFollowUps(ctx, page, pageSize)
	if err != nil {
		log.Error("Failed to list pending follow-ups", "error", err)
		return nil, 0, ErrInteractionListFailed
	}

	return interactions, total, nil
}

// DeleteInteraction soft deletes an interaction
func (uc *useCase) DeleteInteraction(ctx context.Context, id uuidv7.UUID) error {
	log := logger.FromContext(ctx)

	if err := uc.repo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete interaction", "id", id, "error", err)
		return err // Return repository error (ErrInteractionNotFound)
	}

	log.Info("Interaction deleted", "id", id)

	return nil
}
