package interaction

import "errors"

var (
	// ErrInteractionNotFound is returned when interaction is not found
	ErrInteractionNotFound = errors.New("interaction not found")

	// ErrInteractionAlreadyExists is returned when interaction already exists
	ErrInteractionAlreadyExists = errors.New("interaction already exists")

	// ErrInvalidInteractionType is returned when interaction type is invalid
	ErrInvalidInteractionType = errors.New("invalid interaction type")

	// ErrInvalidDirection is returned when direction is invalid
	ErrInvalidDirection = errors.New("invalid direction")

	// ErrInvalidOutcome is returned when outcome is invalid
	ErrInvalidOutcome = errors.New("invalid outcome")

	// ErrInteractionAlreadyEnded is returned when trying to end an already ended interaction
	ErrInteractionAlreadyEnded = errors.New("interaction already ended")
)
