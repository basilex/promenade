package reconciliation

// ReconciliationStatus represents the status of a reconciliation
type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusApproved   Status = "approved"
)

// TransactionType represents the type of transaction in reconciliation items
type TransactionType string

const (
	TransactionTypeBankTransaction TransactionType = "bank_transaction"
	TransactionTypeJournalEntry    TransactionType = "journal_entry"
	TransactionTypeOutstanding     TransactionType = "outstanding"
)
