package dto

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/reconciliation/aggregate"
)

// CreateReconciliationRequest represents a request to create a bank reconciliation
type CreateReconciliationRequest struct {
	BankAccountID             string    `json:"bank_account_id" binding:"required"`
	AccountID                 string    `json:"account_id" binding:"required"`
	ReconciliationDate        time.Time `json:"reconciliation_date" binding:"required"`
	StatementDate             time.Time `json:"statement_date" binding:"required"`
	BankStatementBalanceCents int64     `json:"bank_statement_balance_cents" binding:"required"`
	BookBalanceCents          int64     `json:"book_balance_cents" binding:"required"`
	CurrencyCode              string    `json:"currency_code"`
}

// AddReconciliationItemRequest represents a request to add an item to reconciliation
type AddReconciliationItemRequest struct {
	TransactionType string    `json:"transaction_type" binding:"required"`
	TransactionID   *string   `json:"transaction_id,omitempty"`
	TransactionDate time.Time `json:"transaction_date" binding:"required"`
	Description     string    `json:"description" binding:"required"`
	AmountCents     int64     `json:"amount_cents" binding:"required"`
	Notes           string    `json:"notes"`
}

// ReconciliationItemResponse represents a reconciliation item response
type ReconciliationItemResponse struct {
	ID              string     `json:"id"`
	TransactionType string     `json:"transaction_type"`
	TransactionID   *string    `json:"transaction_id,omitempty"`
	TransactionDate time.Time  `json:"transaction_date"`
	Description     string     `json:"description"`
	AmountCents     int64      `json:"amount_cents"`
	IsMatched       bool       `json:"is_matched"`
	MatchedAt       *time.Time `json:"matched_at,omitempty"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ReconciliationResponse represents a bank reconciliation response
type ReconciliationResponse struct {
	ID                        string                       `json:"id"`
	OrganizationID            string                       `json:"organization_id"`
	BankAccountID             string                       `json:"bank_account_id"`
	AccountID                 string                       `json:"account_id"`
	ReconciliationDate        time.Time                    `json:"reconciliation_date"`
	StatementDate             time.Time                    `json:"statement_date"`
	BankStatementBalanceCents int64                        `json:"bank_statement_balance_cents"`
	BookBalanceCents          int64                        `json:"book_balance_cents"`
	Status                    string                       `json:"status"`
	OutstandingDepositsCents  int64                        `json:"outstanding_deposits_cents"`
	OutstandingChecksCents    int64                        `json:"outstanding_checks_cents"`
	BankFeesCents             int64                        `json:"bank_fees_cents"`
	InterestEarnedCents       int64                        `json:"interest_earned_cents"`
	CurrencyCode              string                       `json:"currency_code"`
	ReconciledBy              *string                      `json:"reconciled_by,omitempty"`
	ReconciledAt              *time.Time                   `json:"reconciled_at,omitempty"`
	ApprovedBy                *string                      `json:"approved_by,omitempty"`
	ApprovedAt                *time.Time                   `json:"approved_at,omitempty"`
	Notes                     string                       `json:"notes"`
	Items                     []ReconciliationItemResponse `json:"items"`
	CreatedAt                 time.Time                    `json:"created_at"`
	UpdatedAt                 time.Time                    `json:"updated_at"`
}

// ToReconciliationResponse converts a bank reconciliation aggregate to response DTO
func ToReconciliationResponse(br *aggregate.Reconciliation) ReconciliationResponse {
	resp := ReconciliationResponse{
		ID:                        br.ID.String(),
		OrganizationID:            br.OrganizationID.String(),
		BankAccountID:             br.BankAccountID.String(),
		AccountID:                 br.AccountID.String(),
		ReconciliationDate:        br.ReconciliationDate,
		StatementDate:             br.StatementDate,
		BankStatementBalanceCents: br.BankStatementBalanceCents,
		BookBalanceCents:          br.BookBalanceCents,
		Status:                    string(br.Status),
		OutstandingDepositsCents:  br.OutstandingDepositsCents,
		OutstandingChecksCents:    br.OutstandingChecksCents,
		BankFeesCents:             br.BankFeesCents,
		InterestEarnedCents:       br.InterestEarnedCents,
		CurrencyCode:              br.CurrencyCode,
		ReconciledAt:              br.ReconciledAt,
		ApprovedAt:                br.ApprovedAt,
		Notes:                     br.Notes,
		CreatedAt:                 br.CreatedAt,
		UpdatedAt:                 br.UpdatedAt,
		Items:                     make([]ReconciliationItemResponse, len(br.Items)),
	}

	if br.ReconciledBy != nil {
		reconciledByStr := br.ReconciledBy.String()
		resp.ReconciledBy = &reconciledByStr
	}

	if br.ApprovedBy != nil {
		approvedByStr := br.ApprovedBy.String()
		resp.ApprovedBy = &approvedByStr
	}

	for i, item := range br.Items {
		itemResp := ReconciliationItemResponse{
			ID:              item.ID.String(),
			TransactionType: string(item.TransactionType),
			TransactionDate: item.TransactionDate,
			Description:     item.Description,
			AmountCents:     item.AmountCents,
			IsMatched:       item.IsMatched,
			MatchedAt:       item.MatchedAt,
			Notes:           item.Notes,
			CreatedAt:       item.CreatedAt,
		}

		if item.TransactionID != nil {
			transactionIDStr := item.TransactionID.String()
			itemResp.TransactionID = &transactionIDStr
		}

		resp.Items[i] = itemResp
	}

	return resp
}

// ToReconciliationResponseList converts a list of bank reconciliations to response DTOs
func ToReconciliationResponseList(reconciliations []*aggregate.Reconciliation) []ReconciliationResponse {
	responses := make([]ReconciliationResponse, len(reconciliations))
	for i, br := range reconciliations {
		responses[i] = ToReconciliationResponse(br)
	}
	return responses
}
