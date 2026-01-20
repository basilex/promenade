package dto

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/aggregate"
)

type RecordTransactionRequest struct {
	AccountID        string    `json:"account_id" binding:"required"`
	Direction        string    `json:"direction" binding:"required,oneof=debit credit"`
	AmountCents      int64     `json:"amount_cents" binding:"required,gt=0"`
	CurrencyCode     string    `json:"currency_code"`
	TransactionAt    time.Time `json:"transaction_at" binding:"required"`
	Description      string    `json:"description" binding:"required"`
	ExternalID       string    `json:"external_id,omitempty"`
	CounterpartyName string    `json:"counterparty_name,omitempty"`
	CounterpartyIBAN string    `json:"counterparty_iban,omitempty"`
}

type MatchTransactionRequest struct {
	EntityType string `json:"entity_type" binding:"required,oneof=invoice order payment"`
	EntityID   string `json:"entity_id" binding:"required"`
}

type SetCounterpartyRequest struct {
	Name string `json:"name" binding:"required"`
	IBAN string `json:"iban,omitempty"`
}

type BankTransactionResponse struct {
	ID                string     `json:"id"`
	BankAccountID     string     `json:"bank_account_id"`
	ExternalID        string     `json:"external_id,omitempty"`
	Direction         string     `json:"direction"`
	AmountCents       int64      `json:"amount_cents"`
	AmountFormatted   string     `json:"amount_formatted"`
	CurrencyCode      string     `json:"currency_code"`
	CounterpartyName  string     `json:"counterparty_name,omitempty"`
	CounterpartyIBAN  string     `json:"counterparty_iban,omitempty"`
	Description       string     `json:"description"`
	TransactionAt     time.Time  `json:"transaction_at"`
	BookedAt          *time.Time `json:"booked_at,omitempty"`
	Status            string     `json:"status"`
	MatchedEntityType *string    `json:"matched_entity_type,omitempty"`
	MatchedEntityID   *string    `json:"matched_entity_id,omitempty"`
	MatchedAt         *time.Time `json:"matched_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type BankTransactionListResponse struct {
	Transactions []BankTransactionResponse `json:"transactions"`
	Total        int                       `json:"total"`
	Limit        int                       `json:"limit"`
	Offset       int                       `json:"offset"`
	TotalPages   int                       `json:"total_pages"`
}

func ToBankTransactionResponse(tx *aggregate.BankTransaction) BankTransactionResponse {
	var matchedType, matchedID *string
	if tx.MatchedEntityType != nil {
		t := string(*tx.MatchedEntityType)
		matchedType = &t
	}
	if tx.MatchedEntityID != nil {
		id := tx.MatchedEntityID.String()
		matchedID = &id
	}

	return BankTransactionResponse{
		ID:                tx.GetID().String(),
		BankAccountID:     tx.BankAccountID.String(),
		ExternalID:        tx.ExternalID,
		Direction:         string(tx.Direction),
		AmountCents:       tx.AmountCents,
		AmountFormatted:   fmt.Sprintf("%.2f %s", float64(tx.AmountCents)/100, tx.CurrencyCode),
		CurrencyCode:      tx.CurrencyCode,
		CounterpartyName:  tx.CounterpartyName,
		CounterpartyIBAN:  tx.CounterpartyIBAN,
		Description:       tx.Description,
		TransactionAt:     tx.TransactionAt,
		BookedAt:          tx.BookedAt,
		Status:            string(tx.Status),
		MatchedEntityType: matchedType,
		MatchedEntityID:   matchedID,
		MatchedAt:         tx.MatchedAt,
		CreatedAt:         tx.CreatedAt,
		UpdatedAt:         tx.UpdatedAt,
	}
}

func ToBankTransactionListResponse(transactions []*aggregate.BankTransaction, total, limit, offset int) BankTransactionListResponse {
	responses := make([]BankTransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = ToBankTransactionResponse(tx)
	}
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	return BankTransactionListResponse{
		Transactions: responses,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
		TotalPages:   totalPages,
	}
}
