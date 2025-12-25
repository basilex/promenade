package entity

import (
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusUnpaid   SubscriptionStatus = "unpaid"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
)

// IsValid checks if the subscription status is valid
func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case SubscriptionStatusTrialing, SubscriptionStatusActive, SubscriptionStatusPastDue,
		SubscriptionStatusUnpaid, SubscriptionStatusCanceled, SubscriptionStatusExpired:
		return true
	default:
		return false
	}
}

// Subscription represents a user's subscription to a plan
type Subscription struct {
	ID                 uuidv7.UUID        `db:"id" json:"id"`
	UserID             uuidv7.UUID        `db:"user_id" json:"user_id" validate:"required"`
	PlanID             uuidv7.UUID        `db:"plan_id" json:"plan_id" validate:"required"`
	Status             SubscriptionStatus `db:"status" json:"status" validate:"required"`
	CurrentPeriodStart time.Time          `db:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `db:"current_period_end" json:"current_period_end"`
	TrialStart         *time.Time         `db:"trial_start" json:"trial_start"`
	TrialEnd           *time.Time         `db:"trial_end" json:"trial_end"`
	CancelAtPeriodEnd  bool               `db:"cancel_at_period_end" json:"cancel_at_period_end"`
	CanceledAt         *time.Time         `db:"canceled_at" json:"canceled_at"`
	Metadata           map[string]string  `db:"metadata" json:"metadata"` // JSONB
	CreatedAt          time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `db:"updated_at" json:"updated_at"`
	DeletedAt          *time.Time         `db:"deleted_at" json:"deleted_at,omitempty"`
}

// NewSubscription creates a new subscription
func NewSubscription(userID, planID uuidv7.UUID, trialDays int) (*Subscription, error) {
	now := time.Now()
	
	subscription := &Subscription{
		ID:                 uuidv7.New(),
		UserID:             userID,
		PlanID:             planID,
		Status:             SubscriptionStatusActive,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0), // 1 month by default
		CancelAtPeriodEnd:  false,
		Metadata:           make(map[string]string),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	
	// Apply trial period if specified
	if trialDays > 0 {
		trialStart := now
		trialEnd := now.AddDate(0, 0, trialDays)
		subscription.TrialStart = &trialStart
		subscription.TrialEnd = &trialEnd
		subscription.Status = SubscriptionStatusTrialing
	}
	
	if err := subscription.Validate(); err != nil {
		return nil, err
	}
	
	return subscription, nil
}

// Validate validates subscription data
func (s *Subscription) Validate() error {
	if s.UserID == uuidv7.Nil {
		return errors.New("user ID is required")
	}
	
	if s.PlanID == uuidv7.Nil {
		return errors.New("plan ID is required")
	}
	
	if s.CurrentPeriodEnd.Before(s.CurrentPeriodStart) {
		return errors.New("current period end must be after start")
	}
	
	return nil
}

// IsActive checks if subscription is active
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrialing
}

// IsTrialing checks if subscription is in trial period
func (s *Subscription) IsTrialing() bool {
	if s.Status != SubscriptionStatusTrialing || s.TrialEnd == nil {
		return false
	}
	
	return time.Now().Before(*s.TrialEnd)
}

// IsPastDue checks if subscription is past due
func (s *Subscription) IsPastDue() bool {
	return s.Status == SubscriptionStatusPastDue
}

// IsCanceled checks if subscription is canceled
func (s *Subscription) IsCanceled() bool {
	return s.Status == SubscriptionStatusCanceled
}

// IsExpired checks if subscription is expired
func (s *Subscription) IsExpired() bool {
	return s.Status == SubscriptionStatusExpired || time.Now().After(s.CurrentPeriodEnd)
}

// Activate activates the subscription
func (s *Subscription) Activate() {
	s.Status = SubscriptionStatusActive
	s.UpdatedAt = time.Now()
}

// MarkPastDue marks subscription as past due
func (s *Subscription) MarkPastDue() {
	s.Status = SubscriptionStatusPastDue
	s.UpdatedAt = time.Now()
}

// Cancel cancels the subscription
func (s *Subscription) Cancel(immediately bool) {
	now := time.Now()
	s.CanceledAt = &now
	
	if immediately {
		s.Status = SubscriptionStatusCanceled
	} else {
		s.CancelAtPeriodEnd = true
	}
	
	s.UpdatedAt = now
}

// Renew renews the subscription for next period
func (s *Subscription) Renew(interval PlanInterval) error {
	if !s.IsActive() {
		return errors.New("cannot renew inactive subscription")
	}
	
	now := time.Now()
	s.CurrentPeriodStart = s.CurrentPeriodEnd
	
	switch interval {
	case PlanIntervalMonthly:
		s.CurrentPeriodEnd = s.CurrentPeriodStart.AddDate(0, 1, 0)
	case PlanIntervalYearly:
		s.CurrentPeriodEnd = s.CurrentPeriodStart.AddDate(1, 0, 0)
	default:
		return errors.New("invalid plan interval")
	}
	
	s.UpdatedAt = now
	return nil
}

// RenewPeriod renews the subscription for next period using monthly interval
func (s *Subscription) RenewPeriod() {
	s.CurrentPeriodStart = s.CurrentPeriodEnd
	s.CurrentPeriodEnd = s.CurrentPeriodStart.AddDate(0, 1, 0)
	s.UpdatedAt = time.Now()
}

// EndTrial ends the trial period
func (s *Subscription) EndTrial() {
	if s.IsTrialing() {
		s.Status = SubscriptionStatusActive
	}
	s.UpdatedAt = time.Now()
}
