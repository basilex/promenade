package dto

import (
    "github.com/basilex/promenade/internal/contexts/accounting/fiscalperiod/aggregate"
)

type CreateFiscalPeriodRequest struct {
    Code       string `json:"code" binding:"required"`
    Name       string `json:"name" binding:"required"`
    PeriodType string `json:"period_type" binding:"required,oneof=month quarter year"`
    StartDate  string `json:"start_date" binding:"required"`
    EndDate    string `json:"end_date" binding:"required"`
}

type FiscalPeriodResponse struct {
    ID             string  `json:"id"`
    OrganizationID string  `json:"organization_id"`
    Code           string  `json:"code"`
    Name           string  `json:"name"`
    PeriodType     string  `json:"period_type"`
    StartDate      string  `json:"start_date"`
    EndDate        string  `json:"end_date"`
    Status         string  `json:"status"`
    LockDate       *string `json:"lock_date,omitempty"`
    ClosedBy       *string `json:"closed_by,omitempty"`
    ClosedAt       *string `json:"closed_at,omitempty"`
    ReopenedBy     *string `json:"reopened_by,omitempty"`
    ReopenedAt     *string `json:"reopened_at,omitempty"`
    LockedBy       *string `json:"locked_by,omitempty"`
    LockedAt       *string `json:"locked_at,omitempty"`
    CreatedAt      string  `json:"created_at"`
    UpdatedAt      string  `json:"updated_at"`
}

func ToFiscalPeriodResponse(fp *aggregate.FiscalPeriod) FiscalPeriodResponse {
    resp := FiscalPeriodResponse{
        ID:             fp.ID.String(),
        OrganizationID: fp.OrganizationID.String(),
        Code:           fp.Code,
        Name:           fp.Name,
        PeriodType:     string(fp.PeriodType),
        StartDate:      fp.StartDate.Format("2006-01-02"),
        EndDate:        fp.EndDate.Format("2006-01-02"),
        Status:         string(fp.Status),
        CreatedAt:      fp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
        UpdatedAt:      fp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
    }

    if fp.LockDate != nil {
        lockDate := fp.LockDate.Format("2006-01-02")
        resp.LockDate = &lockDate
    }
    if fp.ClosedBy != nil {
        closedBy := fp.ClosedBy.String()
        resp.ClosedBy = &closedBy
    }
    if fp.ClosedAt != nil {
        closedAt := fp.ClosedAt.Format("2006-01-02T15:04:05Z07:00")
        resp.ClosedAt = &closedAt
    }
    if fp.ReopenedBy != nil {
        reopenedBy := fp.ReopenedBy.String()
        resp.ReopenedBy = &reopenedBy
    }
    if fp.ReopenedAt != nil {
        reopenedAt := fp.ReopenedAt.Format("2006-01-02T15:04:05Z07:00")
        resp.ReopenedAt = &reopenedAt
    }
    if fp.LockedBy != nil {
        lockedBy := fp.LockedBy.String()
        resp.LockedBy = &lockedBy
    }
    if fp.LockedAt != nil {
        lockedAt := fp.LockedAt.Format("2006-01-02T15:04:05Z07:00")
        resp.LockedAt = &lockedAt
    }

    return resp
}

func ToFiscalPeriodResponseList(periods []*aggregate.FiscalPeriod) []FiscalPeriodResponse {
    responses := make([]FiscalPeriodResponse, 0, len(periods))
    for _, fp := range periods {
        responses = append(responses, ToFiscalPeriodResponse(fp))
    }
    return responses
}