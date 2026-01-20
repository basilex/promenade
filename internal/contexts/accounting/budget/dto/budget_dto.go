package dto

import (
    "github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
)

type CreateBudgetRequest struct {
    Name       string `json:"name" binding:"required"`
    FiscalYear int    `json:"fiscal_year" binding:"required,min=2000,max=2100"`
}

type AddLineRequest struct {
    AccountID    string `json:"account_id" binding:"required"`
    BudgetAmount int64  `json:"budget_amount" binding:"required,min=0"`
    Description  string `json:"description"`
}

type UpdateLineRequest struct {
    BudgetAmount int64 `json:"budget_amount" binding:"required,min=0"`
}

type BudgetLineResponse struct {
    ID              string `json:"id"`
    AccountID       string `json:"account_id"`
    BudgetAmount    int64  `json:"budget_amount"`
    ActualAmount    int64  `json:"actual_amount"`
    VarianceAmount  int64  `json:"variance_amount"`
    VariancePercent int    `json:"variance_percent"`
    Description     string `json:"description"`
}

type BudgetResponse struct {
    ID             string               `json:"id"`
    OrganizationID string               `json:"organization_id"`
    Name           string               `json:"name"`
    FiscalYear     int                  `json:"fiscal_year"`
    Status         string               `json:"status"`
    Lines          []BudgetLineResponse `json:"lines"`
    TotalBudget    int64                `json:"total_budget"`
    TotalActual    int64                `json:"total_actual"`
    ApprovedBy     *string              `json:"approved_by,omitempty"`
    ApprovedAt     *string              `json:"approved_at,omitempty"`
    CreatedAt      string               `json:"created_at"`
    UpdatedAt      string               `json:"updated_at"`
}

func ToBudgetLineResponse(line *aggregate.BudgetLine) BudgetLineResponse {
    return BudgetLineResponse{
        ID:              line.ID.String(),
        AccountID:       line.AccountID.String(),
        BudgetAmount:    line.BudgetAmount,
        ActualAmount:    line.ActualAmount,
        VarianceAmount:  line.VarianceAmount,
        VariancePercent: line.VariancePercent,
        Description:     line.Description,
    }
}

func ToBudgetResponse(b *aggregate.Budget) BudgetResponse {
    lines := make([]BudgetLineResponse, 0, len(b.Lines))
    for _, line := range b.Lines {
        lines = append(lines, ToBudgetLineResponse(line))
    }

    resp := BudgetResponse{
        ID:             b.ID.String(),
        OrganizationID: b.OrganizationID.String(),
        Name:           b.Name,
        FiscalYear:     b.FiscalYear,
        Status:         string(b.Status),
        Lines:          lines,
        TotalBudget:    b.TotalBudget,
        TotalActual:    b.TotalActual,
        CreatedAt:      b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        UpdatedAt:      b.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
    }

    if b.ApprovedBy != nil {
        approvedBy := b.ApprovedBy.String()
        resp.ApprovedBy = &approvedBy
    }

    if b.ApprovedAt != nil {
        resp.ApprovedAt = b.ApprovedAt
    }

    return resp
}

func ToBudgetResponseList(budgets []*aggregate.Budget) []BudgetResponse {
    responses := make([]BudgetResponse, 0, len(budgets))
    for _, b := range budgets {
        responses = append(responses, ToBudgetResponse(b))
    }
    return responses
}