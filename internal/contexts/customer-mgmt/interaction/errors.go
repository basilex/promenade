package interaction

import "errors"

var (
	// Repository Errors
	
	// ErrInteractionNotFound is returned when interaction is not found
	ErrInteractionNotFound = errors.New("interaction not found")

	// ErrInteractionAlreadyExists is returned when interaction already exists
	ErrInteractionAlreadyExists = errors.New("interaction already exists")

	// Business Logic Errors
	
	// ErrInvalidInteractionType is returned when interaction type is invalid
	ErrInvalidInteractionType = errors.New("invalid interaction type")

	// ErrInvalidDirection is returned when direction is invalid
	ErrInvalidDirection = errors.New("invalid direction")

	// ErrInvalidOutcome is returned when outcome is invalid
	ErrInvalidOutcome = errors.New("invalid outcome")

	// ErrInteractionAlreadyEnded is returned when trying to end an already ended interaction
	ErrInteractionAlreadyEnded = errors.New("interaction already ended")

	// ErrSubjectEmpty is returned when subject is empty
	ErrSubjectEmpty = errors.New("subject cannot be empty")

	// ErrDescriptionEmpty is returned when description is empty
	ErrDescriptionEmpty = errors.New("description cannot be empty")

	// ErrEndedAtBeforeStartedAt is returned when ended_at is before started_at
	ErrEndedAtBeforeStartedAt = errors.New("ended_at cannot be before started_at")

	// ErrFollowUpDateRequired is returned when follow-up date is required but not provided
	ErrFollowUpDateRequired = errors.New("follow_up_date required when follow_up_required is true")

	// Technical Operation Errors
	
	// ErrInteractionCreateFailed is returned when interaction creation fails
	ErrInteractionCreateFailed = errors.New("failed to create interaction")

	// ErrInteractionGetFailed is returned when getting interaction fails
	ErrInteractionGetFailed = errors.New("failed to get interaction")

	// ErrInteractionUpdateFailed is returned when updating interaction fails
	ErrInteractionUpdateFailed = errors.New("failed to update interaction")

	// ErrInteractionDeleteFailed is returned when deleting interaction fails
	ErrInteractionDeleteFailed = errors.New("failed to delete interaction")

	// ErrInteractionListFailed is returned when listing interactions fails
	ErrInteractionListFailed = errors.New("failed to list interactions")

	// ErrInteractionValidationFailed is returned when interaction validation fails
	ErrInteractionValidationFailed = errors.New("validation failed")
)
