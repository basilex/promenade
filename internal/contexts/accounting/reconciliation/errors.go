package reconciliation

import "errors"

var (
	// Validation errors
	ErrDescriptionRequired = errors.New("description is required")

	// Business logic errors
	ErrReconciliationNotFound = errors.New("reconciliation not found")
	ErrCannotModifyCompleted  = errors.New("cannot modify completed reconciliation")
	ErrAlreadyCompleted       = errors.New("reconciliation is already completed")
	ErrHasUnmatchedItems      = errors.New("reconciliation has unmatched items")
	ErrNotCompleted           = errors.New("reconciliation must be completed before approval")
	ErrCannotReopenApproved   = errors.New("cannot reopen approved reconciliation")
	ErrAlreadyInProgress      = errors.New("reconciliation is already in progress")
	ErrItemNotFound           = errors.New("reconciliation item not found")
)
