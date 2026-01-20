package dto

import (
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
)

type CreateTaxCodeRequest struct {
	Code    string `json:"code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	TaxType string `json:"tax_type" binding:"required,oneof=vat income_tax payroll_tax withholding excise customs property other"`
	Rate    int    `json:"rate" binding:"required,min=0,max=10000"`
}

type UpdateTaxCodeRequest struct {
	Rate int `json:"rate" binding:"required,min=0,max=10000"`
}

type SetTaxAccountRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

type TaxCodeResponse struct {
	ID                     string  `json:"id"`
	OrganizationID         string  `json:"organization_id"`
	Code                   string  `json:"code"`
	Name                   string  `json:"name"`
	TaxType                string  `json:"tax_type"`
	Rate                   int     `json:"rate"`
	TaxPayableAccountID    *string `json:"tax_payable_account_id,omitempty"`
	TaxReceivableAccountID *string `json:"tax_receivable_account_id,omitempty"`
	IsActive               bool    `json:"is_active"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

func ToTaxCodeResponse(tc *aggregate.TaxCode) TaxCodeResponse {
	resp := TaxCodeResponse{
		ID:             tc.ID.String(),
		OrganizationID: tc.OrganizationID.String(),
		Code:           tc.Code,
		Name:           tc.Name,
		TaxType:        string(tc.TaxType),
		Rate:           tc.Rate,
		IsActive:       tc.IsActive,
		CreatedAt:      tc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      tc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if tc.TaxPayableAccountID != nil {
		accountID := tc.TaxPayableAccountID.String()
		resp.TaxPayableAccountID = &accountID
	}

	if tc.TaxReceivableAccountID != nil {
		accountID := tc.TaxReceivableAccountID.String()
		resp.TaxReceivableAccountID = &accountID
	}

	return resp
}

func ToTaxCodeResponseList(taxCodes []*aggregate.TaxCode) []TaxCodeResponse {
	responses := make([]TaxCodeResponse, 0, len(taxCodes))
	for _, tc := range taxCodes {
		responses = append(responses, ToTaxCodeResponse(tc))
	}
	return responses
}
