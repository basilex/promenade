package dto

import (
	"github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateCostCenterRequest represents a request to create a cost center
type CreateCostCenterRequest struct {
	Code        string       `json:"code" binding:"required"`
	Name        string       `json:"name" binding:"required"`
	Description string       `json:"description"`
	CenterType  string       `json:"center_type" binding:"required"`
	ParentID    *uuidv7.UUID `json:"parent_id,omitempty"`
	ManagerID   *uuidv7.UUID `json:"manager_id,omitempty"`
}

// SetParentRequest represents a request to set parent cost center
type SetParentRequest struct {
	ParentID uuidv7.UUID `json:"parent_id" binding:"required"`
}

// SetManagerRequest represents a request to set cost center manager
type SetManagerRequest struct {
	ManagerID uuidv7.UUID `json:"manager_id" binding:"required"`
}

// CostCenterResponse represents a cost center response
type CostCenterResponse struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	CenterType     string  `json:"center_type"`
	ParentID       *string `json:"parent_id,omitempty"`
	ManagerID      *string `json:"manager_id,omitempty"`
	Level          int     `json:"level"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ToCostCenterResponse converts a cost center aggregate to response DTO
func ToCostCenterResponse(cc *aggregate.CostCenter) CostCenterResponse {
	resp := CostCenterResponse{
		ID:             cc.ID.String(),
		OrganizationID: cc.OrganizationID.String(),
		Code:           cc.Code,
		Name:           cc.Name,
		Description:    cc.Description,
		CenterType:     string(cc.CenterType),
		Level:          cc.Level,
		IsActive:       cc.IsActive,
		CreatedAt:      cc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      cc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if cc.ParentID != nil {
		parentIDStr := cc.ParentID.String()
		resp.ParentID = &parentIDStr
	}

	if cc.ManagerID != nil {
		managerIDStr := cc.ManagerID.String()
		resp.ManagerID = &managerIDStr
	}

	return resp
}

// ToCostCenterResponseList converts a list of cost centers to response DTOs
func ToCostCenterResponseList(costCenters []*aggregate.CostCenter) []CostCenterResponse {
	responses := make([]CostCenterResponse, len(costCenters))
	for i, cc := range costCenters {
		responses[i] = ToCostCenterResponse(cc)
	}
	return responses
}
