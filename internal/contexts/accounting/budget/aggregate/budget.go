package aggregate

import (
    "github.com/basilex/promenade/internal/contexts/accounting/budget"
    "github.com/basilex/promenade/pkg/aggregate"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type BudgetStatus string

const (
    BudgetStatusDraft    BudgetStatus = "draft"
    BudgetStatusApproved BudgetStatus = "approved"
    BudgetStatusActive   BudgetStatus = "active"
    BudgetStatusClosed   BudgetStatus = "closed"
)

type Budget struct {
    aggregate.BaseAggregate

    OrganizationID uuidv7.UUID
    Name           string
    FiscalYear     int
    Status         BudgetStatus
    Lines          []*BudgetLine
    TotalBudget    int64
    TotalActual    int64
    ApprovedBy     *uuidv7.UUID
    ApprovedAt     *string
    LastUpdatedBy  uuidv7.UUID
}

type BudgetLine struct {
    ID               uuidv7.UUID
    AccountID        uuidv7.UUID
    BudgetAmount     int64
    ActualAmount     int64
    VarianceAmount   int64
    VariancePercent  int
    Description      string
}

func NewBudget(organizationID uuidv7.UUID, name string, fiscalYear int, createdBy uuidv7.UUID) (*Budget, error) {
    if name == "" {
        return nil, budget.ErrBudgetNameEmpty
    }
    if fiscalYear < 2000 || fiscalYear > 2100 {
        return nil, budget.ErrInvalidFiscalYear
    }

    return &Budget{
        BaseAggregate:  aggregate.NewBaseAggregate(),
        OrganizationID: organizationID,
        Name:           name,
        FiscalYear:     fiscalYear,
        Status:         BudgetStatusDraft,
        Lines:          make([]*BudgetLine, 0),
        TotalBudget:    0,
        TotalActual:    0,
        LastUpdatedBy:  createdBy,
    }, nil
}

func (b *Budget) AddLine(accountID uuidv7.UUID, budgetAmount int64, description string) error {
    if b.Status != BudgetStatusDraft {
        return budget.ErrCannotModifyApprovedBudget
    }
    if accountID == uuidv7.Nil {
        return budget.ErrAccountRequired
    }
    if budgetAmount < 0 {
        return budget.ErrInvalidBudgetAmount
    }

    for _, line := range b.Lines {
        if line.AccountID == accountID {
            return budget.ErrAccountAlreadyExists
        }
    }

    line := &BudgetLine{
        ID:           uuidv7.New(),
        AccountID:    accountID,
        BudgetAmount: budgetAmount,
        ActualAmount: 0,
        Description:  description,
    }

    b.Lines = append(b.Lines, line)
    b.recalculateTotals()
    b.Touch()
    return nil
}

func (b *Budget) UpdateLine(lineID uuidv7.UUID, budgetAmount int64) error {
    if b.Status == BudgetStatusActive || b.Status == BudgetStatusClosed {
        return budget.ErrCannotModifyActiveBudget
    }
    if budgetAmount < 0 {
        return budget.ErrInvalidBudgetAmount
    }

    for _, line := range b.Lines {
        if line.ID == lineID {
            line.BudgetAmount = budgetAmount
            line.VarianceAmount = line.ActualAmount - line.BudgetAmount
            if line.BudgetAmount > 0 {
                line.VariancePercent = int((line.VarianceAmount * 10000) / line.BudgetAmount)
            }
            b.recalculateTotals()
            b.Touch()
            return nil
        }
    }

    return budget.ErrBudgetLineNotFound
}

func (b *Budget) RemoveLine(lineID uuidv7.UUID) error {
    if b.Status != BudgetStatusDraft {
        return budget.ErrCannotModifyApprovedBudget
    }

    for i, line := range b.Lines {
        if line.ID == lineID {
            b.Lines = append(b.Lines[:i], b.Lines[i+1:]...)
            b.recalculateTotals()
            b.Touch()
            return nil
        }
    }

    return budget.ErrBudgetLineNotFound
}

func (b *Budget) UpdateActuals(lineID uuidv7.UUID, actualAmount int64) error {
    if b.Status != BudgetStatusActive {
        return budget.ErrBudgetNotActive
    }

    for _, line := range b.Lines {
        if line.ID == lineID {
            line.ActualAmount = actualAmount
            line.VarianceAmount = line.ActualAmount - line.BudgetAmount
            if line.BudgetAmount > 0 {
                line.VariancePercent = int((line.VarianceAmount * 10000) / line.BudgetAmount)
            }
            b.recalculateTotals()
            b.Touch()
            return nil
        }
    }

    return budget.ErrBudgetLineNotFound
}

func (b *Budget) Approve(userID uuidv7.UUID) error {
    if b.Status != BudgetStatusDraft {
        return budget.ErrBudgetAlreadyApproved
    }
    if len(b.Lines) == 0 {
        return budget.ErrBudgetHasNoLines
    }

    approvedAt := "approved"
    b.Status = BudgetStatusApproved
    b.ApprovedBy = &userID
    b.ApprovedAt = &approvedAt
    b.Touch()
    return nil
}

func (b *Budget) Activate() error {
    if b.Status != BudgetStatusApproved {
        return budget.ErrBudgetNotApproved
    }

    b.Status = BudgetStatusActive
    b.Touch()
    return nil
}

func (b *Budget) Close() error {
    if b.Status != BudgetStatusActive {
        return budget.ErrBudgetNotActive
    }

    b.Status = BudgetStatusClosed
    b.Touch()
    return nil
}

func (b *Budget) recalculateTotals() {
    var totalBudget, totalActual int64
    for _, line := range b.Lines {
        totalBudget += line.BudgetAmount
        totalActual += line.ActualAmount
    }
    b.TotalBudget = totalBudget
    b.TotalActual = totalActual
}