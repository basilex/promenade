package aggregate

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/journalentry"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// EntryStatus represents the status of a journal entry
type EntryStatus string

const (
	EntryStatusDraft    EntryStatus = "draft"    // Can be edited
	EntryStatusPosted   EntryStatus = "posted"   // Posted to ledger, immutable
	EntryStatusReversed EntryStatus = "reversed" // Cancelled via reverse entry
)

// SourceType represents the type of source event that created this entry
type SourceType string

const (
	SourceTypeManual          SourceType = "manual"
	SourceTypeBankTransaction SourceType = "bank_transaction"
	SourceTypeInvoice         SourceType = "invoice"
	SourceTypePayment         SourceType = "payment"
	SourceTypeReceipt         SourceType = "receipt"
)

// JournalEntry is an aggregate root representing a double-entry bookkeeping transaction
type JournalEntry struct {
	aggregate.BaseAggregate

	OrganizationID uuidv7.UUID // Organization that owns this entry

	// Entry Details
	EntryDate   time.Time   // Transaction date
	Description string      // Entry description
	Status      EntryStatus // Entry status

	// Lines
	Lines []*JournalEntryLine // Debit/Credit lines

	// Source Reference
	SourceType SourceType   // Type of source event
	SourceID   *uuidv7.UUID // Source entity ID (optional)

	// Audit Trail
	LastUpdatedBy uuidv7.UUID  // Last user who modified entry
	PostedBy      *uuidv7.UUID // User who posted entry
	PostedAt      *time.Time   // Posting timestamp
	ReversedBy    *uuidv7.UUID // User who reversed entry
	ReversedAt    *time.Time   // Reversal timestamp
}

// JournalEntryLine represents a single line in a journal entry (debit or credit)
type JournalEntryLine struct {
	ID           uuidv7.UUID
	AccountID    uuidv7.UUID // Reference to Account aggregate
	DebitCents   int64       // Debit amount in cents (> 0 for debit side)
	CreditCents  int64       // Credit amount in cents (> 0 for credit side)
	CurrencyCode string      // Currency
	Description  string      // Line description
	LineOrder    int         // Display order
}

// NewJournalEntry creates a new journal entry in draft status
func NewJournalEntry(organizationID uuidv7.UUID, entryDate time.Time, description string, sourceType SourceType, sourceID *uuidv7.UUID, createdBy uuidv7.UUID) (*JournalEntry, error) {
	if description == "" {
		return nil, journalentry.ErrJournalEntryDescriptionEmpty
	}
	if entryDate.IsZero() {
		return nil, journalentry.ErrJournalEntryInvalidDate
	}

	return &JournalEntry{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		OrganizationID: organizationID,
		EntryDate:      entryDate,
		Description:    description,
		Status:         EntryStatusDraft,
		Lines:          make([]*JournalEntryLine, 0),
		SourceType:     sourceType,
		LastUpdatedBy:  createdBy,
		SourceID:       sourceID,
	}, nil
}

// AddLine adds a new line to the journal entry
func (e *JournalEntry) AddLine(accountID uuidv7.UUID, debitCents, creditCents int64, currencyCode, description string) error {
	if e.Status != EntryStatusDraft {
		return journalentry.ErrJournalEntryAlreadyPosted
	}
	if accountID == uuidv7.Nil {
		return journalentry.ErrJournalEntryLineInvalidAccount
	}
	if debitCents < 0 || creditCents < 0 {
		return journalentry.ErrJournalEntryLineInvalidAmount
	}
	if debitCents > 0 && creditCents > 0 {
		return journalentry.ErrJournalEntryLineBothSides
	}
	if debitCents == 0 && creditCents == 0 {
		return journalentry.ErrJournalEntryLineInvalidAmount
	}
	if currencyCode == "" {
		currencyCode = "UAH"
	}

	line := &JournalEntryLine{
		ID:           uuidv7.New(),
		AccountID:    accountID,
		DebitCents:   debitCents,
		CreditCents:  creditCents,
		CurrencyCode: currencyCode,
		Description:  description,
		LineOrder:    len(e.Lines) + 1,
	}

	e.Lines = append(e.Lines, line)
	e.Touch()
	return nil
}

// RemoveLine removes a line from the journal entry
func (e *JournalEntry) RemoveLine(lineID uuidv7.UUID) error {
	if e.Status != EntryStatusDraft {
		return journalentry.ErrJournalEntryAlreadyPosted
	}

	for i, line := range e.Lines {
		if line.ID == lineID {
			e.Lines = append(e.Lines[:i], e.Lines[i+1:]...)
			e.reorderLines()
			e.Touch()
			return nil
		}
	}

	return nil // Line not found, no error
}

// UpdateDescription updates the entry description
func (e *JournalEntry) UpdateDescription(description string) error {
	if e.Status != EntryStatusDraft {
		return journalentry.ErrJournalEntryAlreadyPosted
	}
	if description == "" {
		return journalentry.ErrJournalEntryDescriptionEmpty
	}

	e.Description = description
	e.Touch()
	return nil
}

// Validate checks if the journal entry is valid for posting
func (e *JournalEntry) Validate() error {
	if len(e.Lines) == 0 {
		return journalentry.ErrJournalEntryNoLines
	}

	// Calculate totals
	var totalDebit, totalCredit int64
	for _, line := range e.Lines {
		totalDebit += line.DebitCents
		totalCredit += line.CreditCents
	}

	// Double-entry rule: Debits must equal Credits
	if totalDebit != totalCredit {
		return journalentry.ErrJournalEntryUnbalanced
	}

	return nil
}

// Post posts the journal entry to the ledger
func (e *JournalEntry) Post(postedBy uuidv7.UUID) error {
	if e.Status != EntryStatusDraft {
		return journalentry.ErrJournalEntryAlreadyPosted
	}

	// Validate before posting
	if err := e.Validate(); err != nil {
		return journalentry.ErrJournalEntryCannotPostDraft
	}

	now := time.Now()
	e.Status = EntryStatusPosted
	e.PostedBy = &postedBy
	e.PostedAt = &now
	e.Touch()
	return nil
}

// Reverse reverses the journal entry by creating a reverse entry
func (e *JournalEntry) Reverse(reversedBy uuidv7.UUID) error {
	if e.Status != EntryStatusPosted {
		return journalentry.ErrJournalEntryNotPosted
	}
	if e.Status == EntryStatusReversed {
		return journalentry.ErrJournalEntryAlreadyReversed
	}

	now := time.Now()
	e.Status = EntryStatusReversed
	e.ReversedBy = &reversedBy
	e.ReversedAt = &now
	e.Touch()
	return nil
}

// GetTotalDebit returns total debit amount
func (e *JournalEntry) GetTotalDebit() int64 {
	var total int64
	for _, line := range e.Lines {
		total += line.DebitCents
	}
	return total
}

// GetTotalCredit returns total credit amount
func (e *JournalEntry) GetTotalCredit() int64 {
	var total int64
	for _, line := range e.Lines {
		total += line.CreditCents
	}
	return total
}

// IsBalanced returns true if debits equal credits
func (e *JournalEntry) IsBalanced() bool {
	return e.GetTotalDebit() == e.GetTotalCredit()
}

// reorderLines updates line order after removal
func (e *JournalEntry) reorderLines() {
	for i, line := range e.Lines {
		line.LineOrder = i + 1
	}
}
