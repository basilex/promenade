package journalentry

import "errors"

// Validation Errors - Input validation failures
var (
	// ErrJournalEntryDescriptionEmpty is returned when description is empty
	ErrJournalEntryDescriptionEmpty = errors.New("journal entry description cannot be empty")

	// ErrJournalEntryInvalidDate is returned when entry date is invalid
	ErrJournalEntryInvalidDate = errors.New("invalid entry date")

	// ErrJournalEntryNoLines is returned when entry has no lines
	ErrJournalEntryNoLines = errors.New("journal entry must have at least one line")

	// ErrJournalEntryInvalidStatus is returned when status is invalid
	ErrJournalEntryInvalidStatus = errors.New("invalid journal entry status")

	// ErrJournalEntryLineInvalidAccount is returned when line has invalid account
	ErrJournalEntryLineInvalidAccount = errors.New("journal entry line must have valid account")

	// ErrJournalEntryLineInvalidAmount is returned when line has invalid amount
	ErrJournalEntryLineInvalidAmount = errors.New("journal entry line must have debit or credit amount")

	// ErrJournalEntryLineBothSides is returned when line has both debit and credit
	ErrJournalEntryLineBothSides = errors.New("journal entry line cannot have both debit and credit")
)

// Not Found Errors - Entity lookup failures
var (
	// ErrJournalEntryNotFound is returned when a journal entry cannot be found
	ErrJournalEntryNotFound = errors.New("journal entry not found")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrJournalEntryUnbalanced is returned when debits don't equal credits
	ErrJournalEntryUnbalanced = errors.New("journal entry is unbalanced: debits must equal credits")

	// ErrJournalEntryAlreadyPosted is returned when trying to modify posted entry
	ErrJournalEntryAlreadyPosted = errors.New("cannot modify posted journal entry")

	// ErrJournalEntryNotPosted is returned when trying to reverse non-posted entry
	ErrJournalEntryNotPosted = errors.New("cannot reverse non-posted journal entry")

	// ErrJournalEntryAlreadyReversed is returned when entry is already reversed
	ErrJournalEntryAlreadyReversed = errors.New("journal entry is already reversed")

	// ErrJournalEntryCannotPostDraft is returned when trying to post invalid draft
	ErrJournalEntryCannotPostDraft = errors.New("cannot post invalid draft entry")
)

// Technical Operation Errors - Infrastructure failures
var (
	// ErrJournalEntryCreateFailed is returned when journal entry creation fails
	ErrJournalEntryCreateFailed = errors.New("failed to create journal entry")

	// ErrJournalEntryGetFailed is returned when journal entry retrieval fails
	ErrJournalEntryGetFailed = errors.New("failed to retrieve journal entry")

	// ErrJournalEntryUpdateFailed is returned when journal entry update fails
	ErrJournalEntryUpdateFailed = errors.New("failed to update journal entry")

	// ErrJournalEntryPostFailed is returned when journal entry posting fails
	ErrJournalEntryPostFailed = errors.New("failed to post journal entry")

	// ErrJournalEntryReverseFailed is returned when journal entry reversal fails
	ErrJournalEntryReverseFailed = errors.New("failed to reverse journal entry")
)
