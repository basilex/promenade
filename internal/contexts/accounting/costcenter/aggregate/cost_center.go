package aggregate

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/accounting/costcenter"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CenterType represents the type of cost center
type CenterType string

const (
    CenterTypeCost       CenterType = "cost_center"
    CenterTypeProfit     CenterType = "profit_center"
    CenterTypeInvestment CenterType = "investment_center"
)

// CostCenter represents a cost or profit center for management accounting
type CostCenter struct {
    aggregate.BaseAggregate

    OrganizationID uuidv7.UUID
    Code           string
    Name           string
    CenterType     CenterType
    ParentID       *uuidv7.UUID // For hierarchical structure
    Level          int
    ManagerID      *uuidv7.UUID // Person responsible
    IsActive       bool
    Description    string
    LastUpdatedBy  uuidv7.UUID
}

// NewCostCenter creates a new cost center
func NewCostCenter(
    organizationID uuidv7.UUID,
    code string,
    name string,
    centerType CenterType,
    createdBy uuidv7.UUID,
) (*CostCenter, error) {
    if code == "" {
        return nil, costcenter.ErrCodeRequired
    }
    if name == "" {
        return nil, costcenter.ErrNameRequired
    }
    if centerType != CenterTypeCost && centerType != CenterTypeProfit && centerType != CenterTypeInvestment {
        return nil, costcenter.ErrInvalidCenterType
    }

    now := time.Now()
    cc := &CostCenter{
        BaseAggregate: aggregate.BaseAggregate{
            ID:        uuidv7.New(),
            Version:   1,
            CreatedAt: now,
            UpdatedAt: now,
        },
        OrganizationID: organizationID,
        Code:           code,
        Name:           name,
        CenterType:     centerType,
        Level:          1, // Root level by default
        IsActive:       true,
        LastUpdatedBy:  createdBy,
    }

    return cc, nil
}

// NewChildCostCenter creates a child cost center under a parent
func NewChildCostCenter(
    organizationID uuidv7.UUID,
    code string,
    name string,
    centerType CenterType,
    parentID uuidv7.UUID,
    parentLevel int,
    createdBy uuidv7.UUID,
) (*CostCenter, error) {
    cc, err := NewCostCenter(organizationID, code, name, centerType, createdBy)
    if err != nil {
        return nil, err
    }

    cc.ParentID = &parentID
    cc.Level = parentLevel + 1

    return cc, nil
}

// SetParent sets the parent cost center
func (cc *CostCenter) SetParent(parentID uuidv7.UUID, parentLevel int) error {
	if cc.ID == parentID {
		return costcenter.ErrCannotBeOwnParent
	}

	cc.ParentID = &parentID
	cc.Level = parentLevel + 1
	cc.Touch()

	return nil
}

// SetManager assigns a manager to the cost center
func (cc *CostCenter) SetManager(managerID uuidv7.UUID) {
	cc.ManagerID = &managerID
	cc.Touch()
}

// Activate activates the cost center
func (cc *CostCenter) Activate() {
	cc.IsActive = true
	cc.Touch()
}

// Deactivate deactivates the cost center
func (cc *CostCenter) Deactivate() {
	cc.IsActive = false
	cc.Touch()
}

// UpdateDetails updates cost center details
func (cc *CostCenter) UpdateDetails(name string, description string, updatedBy uuidv7.UUID) error {
	if name == "" {
		return costcenter.ErrNameRequired
	}

	cc.Name = name
	cc.Description = description
	cc.LastUpdatedBy = updatedBy
	cc.Touch()

	return nil
}