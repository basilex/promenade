package dto

import (
	"github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
)

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	AccountType  string `json:"account_type" binding:"required,oneof=asset liability equity revenue expense"`
	CurrencyCode string `json:"currency_code" binding:"required,len=3"`
	ParentID     string `json:"parent_id,omitempty"`
}

// UpdateAccountRequest represents the request to update an account
type UpdateAccountRequest struct {
	Code         string `json:"code,omitempty"`
	Name         string `json:"name,omitempty"`
	CurrencyCode string `json:"currency_code,omitempty"`
}

// SetParentRequest represents the request to set the parent account
type SetParentRequest struct {
	ParentID string `json:"parent_id" binding:"required"`
}

// AccountResponse represents an account in the response
type AccountResponse struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	AccountType    string  `json:"account_type"`
	CurrencyCode   string  `json:"currency_code"`
	ParentID       *string `json:"parent_id,omitempty"`
	Level          int     `json:"level"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ToAccountResponse converts an account aggregate to a response DTO
func ToAccountResponse(acc *aggregate.Account) AccountResponse {
	resp := AccountResponse{
		ID:             acc.ID.String(),
		OrganizationID: acc.OrganizationID.String(),
		Code:           acc.Code,
		Name:           acc.Name,
		AccountType:    string(acc.Type),
		CurrencyCode:   acc.CurrencyCode,
		Level:          acc.Level,
		IsActive:       acc.IsActive,
		CreatedAt:      acc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      acc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if acc.ParentID != nil {
		parentIDStr := acc.ParentID.String()
		resp.ParentID = &parentIDStr
	}

	return resp
}

// ToAccountResponseList converts a list of account aggregates to response DTOs
func ToAccountResponseList(accounts []*aggregate.Account) []AccountResponse {
	responses := make([]AccountResponse, 0, len(accounts))
	for _, acc := range accounts {
		responses = append(responses, ToAccountResponse(acc))
	}
	return responses
}
