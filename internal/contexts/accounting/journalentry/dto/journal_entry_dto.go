package dto

import (
	"github.com/basilex/promenade/internal/contexts/accounting/journalentry/aggregate"
)

type CreateJournalEntryRequest struct {
	EntryDate   string  `json:"entry_date" binding:"required"`
	Description string  `json:"description" binding:"required"`
	SourceType  string  `json:"source_type" binding:"required,oneof=manual bank_transaction invoice payment receipt"`
	SourceID    *string `json:"source_id,omitempty"`
}

type AddLineRequest struct {
	AccountID   string `json:"account_id" binding:"required"`
	DebitCents  int64  `json:"debit_cents"`
	CreditCents int64  `json:"credit_cents"`
	Description string `json:"description"`
}

type RemoveLineRequest struct {
	LineID string `json:"line_id" binding:"required"`
}

type UpdateDescriptionRequest struct {
	Description string `json:"description" binding:"required"`
}

type ReverseEntryRequest struct {
	ReverseDescription string `json:"reverse_description" binding:"required"`
}

type JournalEntryLineResponse struct {
	ID           string `json:"id"`
	AccountID    string `json:"account_id"`
	DebitCents   int64  `json:"debit_cents"`
	CreditCents  int64  `json:"credit_cents"`
	CurrencyCode string `json:"currency_code"`
	Description  string `json:"description"`
	LineOrder    int    `json:"line_order"`
}

type JournalEntryResponse struct {
	ID             string                     `json:"id"`
	OrganizationID string                     `json:"organization_id"`
	EntryDate      string                     `json:"entry_date"`
	Description    string                     `json:"description"`
	Status         string                     `json:"status"`
	Lines          []JournalEntryLineResponse `json:"lines"`
	SourceType     string                     `json:"source_type"`
	SourceID       *string                    `json:"source_id,omitempty"`
	PostedBy       *string                    `json:"posted_by,omitempty"`
	PostedAt       *string                    `json:"posted_at,omitempty"`
	ReversedBy     *string                    `json:"reversed_by,omitempty"`
	ReversedAt     *string                    `json:"reversed_at,omitempty"`
	CreatedAt      string                     `json:"created_at"`
	UpdatedAt      string                     `json:"updated_at"`
}

func ToJournalEntryLineResponse(line *aggregate.JournalEntryLine) JournalEntryLineResponse {
	return JournalEntryLineResponse{
		ID:           line.ID.String(),
		AccountID:    line.AccountID.String(),
		DebitCents:   line.DebitCents,
		CreditCents:  line.CreditCents,
		CurrencyCode: line.CurrencyCode,
		Description:  line.Description,
		LineOrder:    line.LineOrder,
	}
}

func ToJournalEntryResponse(je *aggregate.JournalEntry) JournalEntryResponse {
	lines := make([]JournalEntryLineResponse, 0, len(je.Lines))
	for _, line := range je.Lines {
		lines = append(lines, ToJournalEntryLineResponse(line))
	}

	resp := JournalEntryResponse{
		ID:             je.ID.String(),
		OrganizationID: je.OrganizationID.String(),
		EntryDate:      je.EntryDate.Format("2006-01-02"),
		Description:    je.Description,
		Status:         string(je.Status),
		Lines:          lines,
		SourceType:     string(je.SourceType),
		CreatedAt:      je.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      je.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if je.SourceID != nil {
		sourceID := je.SourceID.String()
		resp.SourceID = &sourceID
	}

	if je.PostedBy != nil {
		postedBy := je.PostedBy.String()
		resp.PostedBy = &postedBy
	}

	if je.PostedAt != nil {
		postedAt := je.PostedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PostedAt = &postedAt
	}

	if je.ReversedBy != nil {
		reversedBy := je.ReversedBy.String()
		resp.ReversedBy = &reversedBy
	}

	if je.ReversedAt != nil {
		reversedAt := je.ReversedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ReversedAt = &reversedAt
	}

	return resp
}

func ToJournalEntryResponseList(entries []*aggregate.JournalEntry) []JournalEntryResponse {
	responses := make([]JournalEntryResponse, 0, len(entries))
	for _, je := range entries {
		responses = append(responses, ToJournalEntryResponse(je))
	}
	return responses
}
