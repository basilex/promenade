package dto

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/contexts/banking/bankaccount/aggregate"
)

type CreateManualAccountRequest struct {
	OrganizationID string `json:"organization_id" binding:"required"`
	Name           string `json:"name" binding:"required"`
	BankName       string `json:"bank_name" binding:"required"`
	CurrencyCode   string `json:"currency_code"`
	IBAN           string `json:"iban,omitempty"`
	AccountNumber  string `json:"account_number,omitempty"`
}

type ConnectProviderAccountRequest struct {
	OrganizationID    string `json:"organization_id" binding:"required"`
	Name              string `json:"name" binding:"required"`
	BankName          string `json:"bank_name" binding:"required"`
	Provider          string `json:"provider" binding:"required"`
	ProviderAccountID string `json:"provider_account_id" binding:"required"`
}

type UpdateAccountDetailsRequest struct {
	Name          string `json:"name,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	IBAN          string `json:"iban,omitempty"`
	AccountNumber string `json:"account_number,omitempty"`
}

type UpdateBalanceRequest struct {
	BalanceCents int64 `json:"balance_cents" binding:"required"`
}

type BankAccountResponse struct {
	ID                string     `json:"id"`
	OrganizationID    string     `json:"organization_id"`
	Name              string     `json:"name"`
	BankName          string     `json:"bank_name"`
	IBAN              string     `json:"iban,omitempty"`
	AccountNumber     string     `json:"account_number,omitempty"`
	CurrencyCode      string     `json:"currency_code"`
	Provider          string     `json:"provider"`
	ProviderAccountID string     `json:"provider_account_id,omitempty"`
	Status            string     `json:"status"`
	BalanceCents      int64      `json:"balance_cents"`
	BalanceFormatted  string     `json:"balance_formatted"`
	LastSyncAt        *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type BankAccountListResponse struct {
	Accounts   []BankAccountResponse `json:"accounts"`
	Total      int                   `json:"total"`
	Limit      int                   `json:"limit"`
	Offset     int                   `json:"offset"`
	TotalPages int                   `json:"total_pages"`
}

func ToBankAccountResponse(acc *aggregate.BankAccount) BankAccountResponse {
	return BankAccountResponse{
		ID:                acc.GetID().String(),
		OrganizationID:    acc.OrganizationID.String(),
		Name:              acc.Name,
		BankName:          acc.BankName,
		IBAN:              acc.IBAN,
		AccountNumber:     acc.AccountNumber,
		CurrencyCode:      acc.CurrencyCode,
		Provider:          string(acc.Provider),
		ProviderAccountID: acc.ProviderAccountID,
		Status:            string(acc.Status),
		BalanceCents:      acc.BalanceCents,
		BalanceFormatted:  fmt.Sprintf("%.2f %s", float64(acc.BalanceCents)/100, acc.CurrencyCode),
		LastSyncAt:        acc.LastSyncAt,
		CreatedAt:         acc.CreatedAt,
		UpdatedAt:         acc.UpdatedAt,
	}
}

func ToBankAccountListResponse(accounts []*aggregate.BankAccount, total, limit, offset int) BankAccountListResponse {
	responses := make([]BankAccountResponse, len(accounts))
	for i, acc := range accounts {
		responses[i] = ToBankAccountResponse(acc)
	}
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	return BankAccountListResponse{
		Accounts:   responses,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPages,
	}
}
