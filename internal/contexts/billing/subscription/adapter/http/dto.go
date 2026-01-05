package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/billing/subscription"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateSubscriptionRequest for subscription creation
type CreateSubscriptionRequest struct {
	CustomerID    string `json:"customer_id" binding:"required"`
	PlanID        string `json:"plan_id" binding:"required"`
	BillingPeriod string `json:"billing_period" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	Amount        int64  `json:"amount" binding:"required"`
	TrialDays     int    `json:"trial_days"`
}

// UpdateSubscriptionRequest for subscription updates
type UpdateSubscriptionRequest struct {
	PlanID *string `json:"plan_id"`
	Amount *int64  `json:"amount"`
}

// SubscriptionResponse represents subscription in API response
type SubscriptionResponse struct {
	ID             string  `json:"id"`
	SubscriptionNo string  `json:"subscription_no"`
	CustomerID     string  `json:"customer_id"`
	PlanID         string  `json:"plan_id"`
	Status         string  `json:"status"`
	BillingPeriod  string  `json:"billing_period"`
	Currency       string  `json:"currency"`
	Amount         int64   `json:"amount"`
	StartDate      string  `json:"start_date"`
	RenewalDate    string  `json:"renewal_date"`
	TrialEndDate   *string `json:"trial_end_date,omitempty"`
	CancelledAt    *string `json:"cancelled_at,omitempty"`
	CancelReason   *string `json:"cancel_reason,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// ListSubscriptionsRequest for list queries
type ListSubscriptionsRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	CustomerID string `form:"customer_id"`
	Status     string `form:"status"`
}

// ToSubscriptionResponse converts entity to DTO
func ToSubscriptionResponse(s *subscription.Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:             s.ID.String(),
		SubscriptionNo: s.SubscriptionNo,
		CustomerID:     s.CustomerID.String(),
		PlanID:         s.PlanID,
		Status:         string(s.Status),
		BillingPeriod:  string(s.BillingPeriod),
		Currency:       s.Amount.Currency,
		Amount:         s.Amount.Amount,
		StartDate:      s.StartDate.Format(time.RFC3339),
		RenewalDate:    s.RenewalDate.Format(time.RFC3339),
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      s.UpdatedAt.Format(time.RFC3339),
	}

	if s.TrialEndDate != nil {
		t := s.TrialEndDate.Format(time.RFC3339)
		resp.TrialEndDate = &t
	}

	if s.CancelledAt != nil {
		t := s.CancelledAt.Format(time.RFC3339)
		resp.CancelledAt = &t
	}

	if s.CancelReason != "" {
		resp.CancelReason = &s.CancelReason
	}

	return resp
}

// ParseBillingPeriod converts string to BillingPeriod
func ParseBillingPeriod(s string) subscription.BillingPeriod {
	switch s {
	case "monthly":
		return subscription.BillingPeriodMonthly
	case "quarterly":
		return subscription.BillingPeriodQuarterly
	case "yearly":
		return subscription.BillingPeriodYearly
	default:
		return subscription.BillingPeriodMonthly
	}
}

// ParseCustomerID parses UUID string
func ParseCustomerID(s string) (uuidv7.UUID, error) {
	return uuidv7.Parse(s)
}
