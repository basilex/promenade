package dto

import (
	"fmt"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/deal/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateDealRequest is the request body for creating a deal
type CreateDealRequest struct {
	Name              string  `json:"name" binding:"required,min=2,max=255"`
	CustomerID        string  `json:"customer_id" binding:"required"`
	CompanyID         *string `json:"company_id,omitempty"`
	AssignedTo        *string `json:"assigned_to,omitempty"`
	Value             int64   `json:"value" binding:"required,min=0"`
	Currency          string  `json:"currency" binding:"required,len=3"` // ISO 4217
	ExpectedCloseDate string  `json:"expected_close_date" binding:"required"`
	Source            *string `json:"source,omitempty" binding:"omitempty,oneof=website referral cold_call email social_media event"`
}

// UpdateDealBasicInfoRequest is the request body for updating basic deal info
type UpdateDealBasicInfoRequest struct {
	Name              *string `json:"name,omitempty" binding:"omitempty,min=2,max=255"`
	ExpectedCloseDate *string `json:"expected_close_date,omitempty"`
}

// UpdateDealValueRequest is the request body for updating deal value
type UpdateDealValueRequest struct {
	Value    int64  `json:"value" binding:"required,min=0"`
	Currency string `json:"currency" binding:"required,len=3"`
}

// MoveDealToStageRequest is the request body for moving deal to stage
type MoveDealToStageRequest struct {
	Stage string `json:"stage" binding:"required,oneof=lead qualified proposal negotiation closed_won closed_lost"`
}

// MarkDealAsWonRequest is the request body for marking deal as won
type MarkDealAsWonRequest struct {
	CloseDate string `json:"close_date" binding:"required"`
}

// MarkDealAsLostRequest is the request body for marking deal as lost
type MarkDealAsLostRequest struct {
	CloseDate  string  `json:"close_date" binding:"required"`
	LostReason *string `json:"lost_reason,omitempty" binding:"omitempty,max=500"`
}

// DealResponse is the response body for deal queries
type DealResponse struct {
	ID                string  `json:"id"`
	CustomerID        string  `json:"customer_id"`
	CompanyID         *string `json:"company_id,omitempty"`
	AssignedTo        *string `json:"assigned_to,omitempty"`
	Name              string  `json:"name"`
	Value             int64   `json:"value"`
	Currency          string  `json:"currency"`
	Stage             string  `json:"stage"`
	Probability       int     `json:"probability"`
	Source            string  `json:"source"`
	ExpectedCloseDate string  `json:"expected_close_date"`
	ActualCloseDate   *string `json:"actual_close_date,omitempty"`
	LostReason        *string `json:"lost_reason,omitempty"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	DeletedAt         *string `json:"deleted_at,omitempty"`
}

// PipelineStatsResponse is the response body for pipeline statistics
type PipelineStatsResponse struct {
	Stats map[string]int64 `json:"stats"`
}

// WonDealsResponse is the response body for won deals statistics
type WonDealsResponse struct {
	Count      int64 `json:"count"`
	TotalValue int64 `json:"total_value"`
}

// ToDealResponse converts deal entity to response DTO
func ToDealResponse(d *aggregate.Deal) DealResponse {
	resp := DealResponse{
		ID:                d.ID.String(),
		CustomerID:        d.CustomerID.String(),
		Name:              d.Name,
		Value:             d.Value.Amount,
		Currency:          d.Currency,
		Stage:             string(d.Stage),
		Probability:       d.Probability,
		Source:            string(d.Source),
		ExpectedCloseDate: d.ExpectedCloseDate.Format("2006-01-02"),
		CreatedAt:         d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Optional fields
	if d.CompanyID != nil {
		companyID := d.CompanyID.String()
		resp.CompanyID = &companyID
	}
	if d.AssignedTo != uuidv7.Nil {
		assignedTo := d.AssignedTo.String()
		resp.AssignedTo = &assignedTo
	}
	if d.ActualCloseDate != nil {
		closeDate := d.ActualCloseDate.Format("2006-01-02")
		resp.ActualCloseDate = &closeDate
	}
	if d.CloseReason != "" {
		resp.LostReason = &d.CloseReason
	}
	if d.DeletedAt != nil {
		deletedAt := d.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.DeletedAt = &deletedAt
	}

	return resp
}

// ToDealListResponse converts a list of deals to list response DTO
func ToDealListResponse(deals []*aggregate.Deal) []*DealResponse {
	dealResponses := make([]*DealResponse, 0, len(deals))
	for _, d := range deals {
		resp := ToDealResponse(d)
		dealResponses = append(dealResponses, &resp)
	}
	return dealResponses
}

// ToPipelineStatsResponse converts pipeline stats to response DTO
func ToPipelineStatsResponse(stats map[aggregate.DealStage]int64) PipelineStatsResponse {
	statsMap := make(map[string]int64)
	for stage, count := range stats {
		statsMap[string(stage)] = count
	}
	return PipelineStatsResponse{Stats: statsMap}
}

// Helper functions for parsing fields

// parseCustomerID parses customer ID from request
func parseCustomerID(customerID string) (uuidv7.UUID, error) {
	id, err := uuidv7.Parse(customerID)
	if err != nil {
		return uuidv7.UUID{}, fmt.Errorf("invalid customer ID format: %w", err)
	}
	return id, nil
}


// parseAssignedTo parses assigned to user ID from request
func parseAssignedTo(assignedTo *string) (*uuidv7.UUID, error) {
	if assignedTo == nil || *assignedTo == "" {
		return nil, nil
	}
	id, err := uuidv7.Parse(*assignedTo)
	if err != nil {
		return nil, fmt.Errorf("invalid assigned to ID format: %w", err)
	}
	return &id, nil
}
