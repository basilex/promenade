package dto

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
)

// CreateCashRegisterRequest represents cash register creation request
type CreateCashRegisterRequest struct {
	OrganizationID string `json:"organization_id" binding:"required"`
	FiscalNumber   string `json:"fiscal_number" binding:"required"`
	Model          string `json:"model" binding:"required"`
	CreatedBy      string `json:"created_by" binding:"required"`
}

// UpdateCashRegisterRequest represents cash register update request
type UpdateCashRegisterRequest struct {
	Model         *string `json:"model,omitempty"`
	LicenseKey    *string `json:"license_key,omitempty"`
	LastUpdatedBy string  `json:"last_updated_by" binding:"required"`
}

// ActivateCashRegisterRequest represents cash register activation request
type ActivateCashRegisterRequest struct {
	LicenseKey  string `json:"license_key" binding:"required"`
	ActivatedBy string `json:"activated_by" binding:"required"`
}

// dto.CashRegisterResponse represents cash register response
type CashRegisterResponse struct {
	ID                     string     `json:"id"`
	OrganizationID         string     `json:"organization_id"`
	FiscalNumber           string     `json:"fiscal_number"`
	Model                  string     `json:"model"`
	Status                 string     `json:"status"`
	LicenseKey             string     `json:"license_key,omitempty"`
	LastSyncAt             *time.Time `json:"last_sync_at,omitempty"`
	ProviderCashRegisterID string     `json:"provider_cash_register_id,omitempty"`
	ActiveShiftID          string     `json:"active_shift_id,omitempty"`
	ShiftOpenedAt          *time.Time `json:"shift_opened_at,omitempty"`
	ShiftClosedAt          *time.Time `json:"shift_closed_at,omitempty"`
	LastZReportID          string     `json:"last_z_report_id,omitempty"`
	LastZReportAt          *time.Time `json:"last_z_report_at,omitempty"`
	LastUpdatedBy          string     `json:"last_updated_by"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// ToCashRegisterResponse converts entity to response DTO
func ToCashRegisterResponse(cr *aggregate.CashRegister) *CashRegisterResponse {
	return &CashRegisterResponse{
		ID:                     cr.GetID().String(),
		OrganizationID:         cr.OrganizationID.String(),
		FiscalNumber:           cr.FiscalNumber,
		Model:                  cr.Model,
		Status:                 string(cr.Status),
		LicenseKey:             cr.LicenseKey,
		LastSyncAt:             cr.LastSyncAt,
		ProviderCashRegisterID: cr.ProviderCashRegisterID,
		ActiveShiftID:          cr.ActiveShiftID,
		ShiftOpenedAt:          cr.ShiftOpenedAt,
		ShiftClosedAt:          cr.ShiftClosedAt,
		LastZReportID:          cr.LastZReportID,
		LastZReportAt:          cr.LastZReportAt,
		LastUpdatedBy:          cr.LastUpdatedBy.String(),
		CreatedAt:              cr.CreatedAt,
		UpdatedAt:              cr.UpdatedAt,
	}
}

// ToCashRegisterListResponse converts slice of entities to response DTOs
func ToCashRegisterListResponse(cashRegisters []*aggregate.CashRegister) []*CashRegisterResponse {
	responses := make([]*CashRegisterResponse, len(cashRegisters))
	for i, cr := range cashRegisters {
		responses[i] = ToCashRegisterResponse(cr)
	}
	return responses
}
