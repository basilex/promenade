package aggregate

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Item represents an individual transaction in the reconciliation
type Item struct {
	ID              uuidv7.UUID
	TransactionType reconciliation.TransactionType
	TransactionID   *uuidv7.UUID
	TransactionDate time.Time
	Description     string
	AmountCents     int64
	IsMatched       bool
	MatchedAt       *time.Time
	Notes           string
	CreatedAt       time.Time
}

// Reconciliation represents a bank statement reconciliation
type Reconciliation struct {
	aggregate.BaseAggregate

	OrganizationID            uuidv7.UUID
	BankAccountID             uuidv7.UUID // Reference to banking context
	AccountID                 uuidv7.UUID // 311 account in chart of accounts
	ReconciliationDate        time.Time
	StatementDate             time.Time
	BankStatementBalanceCents int64
	BookBalanceCents          int64
	Status                    reconciliation.Status
	OutstandingDepositsCents  int64
	OutstandingChecksCents    int64
	BankFeesCents             int64
	InterestEarnedCents       int64
	CurrencyCode              string
	ReconciledBy              *uuidv7.UUID
	ReconciledAt              *time.Time
	ApprovedBy                *uuidv7.UUID
	ApprovedAt                *time.Time
	Notes                     string
	LastUpdatedBy             uuidv7.UUID
	Items                     []Item
}

// NewReconciliation creates a new bank reconciliation
func NewReconciliation(
	organizationID uuidv7.UUID,
	bankAccountID uuidv7.UUID,
	accountID uuidv7.UUID,
	reconciliationDate time.Time,
	statementDate time.Time,
	bankStatementBalanceCents int64,
	bookBalanceCents int64,
	currencyCode string,
	createdBy uuidv7.UUID,
) (*Reconciliation, error) {
	if currencyCode == "" {
		currencyCode = "UAH"
	}

	now := time.Now()
	br := &Reconciliation{
		BaseAggregate: aggregate.BaseAggregate{
			ID:        uuidv7.New(),
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		OrganizationID:            organizationID,
		BankAccountID:             bankAccountID,
		AccountID:                 accountID,
		ReconciliationDate:        reconciliationDate,
		StatementDate:             statementDate,
		BankStatementBalanceCents: bankStatementBalanceCents,
		BookBalanceCents:          bookBalanceCents,
		Status:                    reconciliation.StatusInProgress,
		CurrencyCode:              currencyCode,
		LastUpdatedBy:             createdBy,
		Items:                     []Item{},
	}

	return br, nil
}

// AddItem adds a reconciliation item
func (br *Reconciliation) AddItem(
	transactionType reconciliation.TransactionType,
	transactionID *uuidv7.UUID,
	transactionDate time.Time,
	description string,
	amountCents int64,
	notes string,
) error {
	if br.Status != reconciliation.StatusInProgress {
		return reconciliation.ErrCannotModifyCompleted
	}
	if description == "" {
		return reconciliation.ErrDescriptionRequired
	}

	item := Item{
		ID:              uuidv7.New(),
		TransactionType: transactionType,
		TransactionID:   transactionID,
		TransactionDate: transactionDate,
		Description:     description,
		AmountCents:     amountCents,
		IsMatched:       false,
		Notes:           notes,
		CreatedAt:       time.Now(),
	}

	br.Items = append(br.Items, item)
	br.Touch()

	return nil
}

// RemoveItem removes a reconciliation item by ID
func (br *Reconciliation) RemoveItem(itemID uuidv7.UUID) error {
	if br.Status != reconciliation.StatusInProgress {
		return reconciliation.ErrCannotModifyCompleted
	}

	for i, item := range br.Items {
		if item.ID == itemID {
			br.Items = append(br.Items[:i], br.Items[i+1:]...)
			br.Touch()
			return nil
		}
	}

	return reconciliation.ErrItemNotFound
}

// MarkItemMatched marks an item as matched
func (br *Reconciliation) MarkItemMatched(itemID uuidv7.UUID) error {
	if br.Status != reconciliation.StatusInProgress {
		return reconciliation.ErrCannotModifyCompleted
	}

	for i := range br.Items {
		if br.Items[i].ID == itemID {
			now := time.Now()
			br.Items[i].IsMatched = true
			br.Items[i].MatchedAt = &now
			br.Touch()
			return nil
		}
	}

	return reconciliation.ErrItemNotFound
}

// Complete marks the reconciliation as completed
func (br *Reconciliation) Complete(reconciledBy uuidv7.UUID) error {
	if br.Status != reconciliation.StatusInProgress {
		return reconciliation.ErrAlreadyCompleted
	}

	// Verify all items are matched
	for _, item := range br.Items {
		if !item.IsMatched {
			return reconciliation.ErrHasUnmatchedItems
		}
	}

	now := time.Now()
	br.Status = reconciliation.StatusCompleted
	br.ReconciledBy = &reconciledBy
	br.ReconciledAt = &now
	br.Touch()

	return nil
}

// Approve approves the reconciliation
func (br *Reconciliation) Approve(approvedBy uuidv7.UUID) error {
	if br.Status != reconciliation.StatusCompleted {
		return reconciliation.ErrNotCompleted
	}

	now := time.Now()
	br.Status = reconciliation.StatusApproved
	br.ApprovedBy = &approvedBy
	br.ApprovedAt = &now
	br.Touch()

	return nil
}

// Reopen reopens a completed reconciliation for modifications
func (br *Reconciliation) Reopen() error {
	if br.Status == reconciliation.StatusApproved {
		return reconciliation.ErrCannotReopenApproved
	}
	if br.Status == reconciliation.StatusInProgress {
		return reconciliation.ErrAlreadyInProgress
	}

	br.Status = reconciliation.StatusInProgress
	br.ReconciledBy = nil
	br.ReconciledAt = nil
	br.Touch()

	return nil
}

// UpdateBalances updates the reconciliation balances
func (br *Reconciliation) UpdateBalances(
	bankStatementBalanceCents int64,
	bookBalanceCents int64,
	updatedBy uuidv7.UUID,
) error {
	if br.Status != reconciliation.StatusInProgress {
		return reconciliation.ErrCannotModifyCompleted
	}

	br.BankStatementBalanceCents = bankStatementBalanceCents
	br.BookBalanceCents = bookBalanceCents
	br.LastUpdatedBy = updatedBy
	br.Touch()

	return nil
}

// CalculateDifference calculates the difference between bank and book balances
func (br *Reconciliation) CalculateDifference() int64 {
	return br.BankStatementBalanceCents - br.BookBalanceCents
}

// IsReconciled checks if the reconciliation balances match
func (br *Reconciliation) IsReconciled() bool {
	return br.CalculateDifference() == 0
}
