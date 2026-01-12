package subscription

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// Subscription is the aggregate root for recurring billing subscriptions.
// It manages customer subscriptions with auto-renewal, trial periods, and lifecycle management.
type Subscription struct {
	aggregate.BaseAggregate

	SubscriptionNo string          // Auto-generated: SUB-YYYY-NNNNNN
	CustomerID     uuidv7.UUID     // Required: reference to customer
	PlanID             string          // Subscription plan identifier (e.g., "basic", "pro", "enterprise")
	Status             SubscriptionStatus
	BillingPeriod      BillingPeriod   // monthly, quarterly, yearly

	// Financial details
	Currency string                     // ISO 4217 (USD, EUR, UAH)
	Amount   valueobject.Money          // Subscription price per billing period

	// Dates
	StartDate   time.Time  // Subscription start date
	EndDate     *time.Time // Subscription end date (nullable for active subscriptions)
	RenewalDate time.Time  // Next renewal date
	TrialEndDate *time.Time // Trial period end date (nullable)

	// Cancellation
	CancelledAt               *time.Time // When subscription was cancelled (nullable)
	CancelReason              string     // Why subscription was cancelled
	CancellationEffectiveDate *time.Time // When cancellation takes effect (nullable)

	// Metadata (stored as TEXT JSON in database for cross-database compatibility)
	Metadata jsonstore.Field[map[string]string] // Additional subscription metadata
}

// SubscriptionStatus represents the lifecycle state of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusTrial     SubscriptionStatus = "trial"     // In trial period
	SubscriptionStatusActive    SubscriptionStatus = "active"    // Active and billing
	SubscriptionStatusPaused    SubscriptionStatus = "paused"    // Temporarily suspended
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled" // Cancelled by user
	SubscriptionStatusExpired   SubscriptionStatus = "expired"   // Ended naturally
)

// BillingPeriod represents how often the subscription renews
type BillingPeriod string

const (
	BillingPeriodMonthly   BillingPeriod = "monthly"
	BillingPeriodQuarterly BillingPeriod = "quarterly"
	BillingPeriodYearly    BillingPeriod = "yearly"
)

// NewSubscription creates a new subscription aggregate with validation
func NewSubscription(
	customerID uuidv7.UUID,
	planID string,
	billingPeriod BillingPeriod,
	currency string,
	amount int64,
	startDate time.Time,
	trialDays int,
) (*Subscription, error) {
	if customerID == uuidv7.Nil {
		return nil, ErrCustomerIDRequired
	}
	if planID == "" {
		return nil, ErrPlanIDRequired
	}
	if currency == "" {
		return nil, ErrCurrencyRequired
	}
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	money, err := valueobject.NewMoney(amount, currency)
	if err != nil {
		return nil, ErrInvalidMoney
	}

	subscription := &Subscription{
		BaseAggregate: aggregate.NewBaseAggregate(),
		SubscriptionNo: generateSubscriptionNumber(),
		CustomerID:    customerID,
		PlanID:        planID,
		Status:        SubscriptionStatusActive,
		BillingPeriod: billingPeriod,
		Currency:      currency,
		Amount:        money,
		StartDate:     startDate,
		RenewalDate:   calculateRenewalDate(startDate, billingPeriod),
	}
	// Initialize metadata as empty jsonstore Field
	subscription.Metadata.Set(make(map[string]string))

	// Set trial period if specified
	if trialDays > 0 {
		subscription.Status = SubscriptionStatusTrial
		trialEnd := startDate.AddDate(0, 0, trialDays)
		subscription.TrialEndDate = &trialEnd
	}

	return subscription, nil
}

// generateSubscriptionNumber generates a unique subscription number
// Format: SUB-YYYY-NNNNNN (where NNNNNN is random 6-digit number for uniqueness)
func generateSubscriptionNumber() string {
	// Use random number for uniqueness (important for concurrent tests)
	now := time.Now()
	randomSuffix := rand.Intn(1000000) // 0-999999
	return fmt.Sprintf("SUB-%d-%06d", now.Year(), randomSuffix)
}

// calculateRenewalDate calculates next renewal date based on billing period
func calculateRenewalDate(startDate time.Time, period BillingPeriod) time.Time {
	switch period {
	case BillingPeriodMonthly:
		return startDate.AddDate(0, 1, 0)
	case BillingPeriodQuarterly:
		return startDate.AddDate(0, 3, 0)
	case BillingPeriodYearly:
		return startDate.AddDate(1, 0, 0)
	default:
		return startDate.AddDate(0, 1, 0) // Default to monthly
	}
}

// Activate transitions subscription from trial to active status
func (s *Subscription) Activate() error {
	if s.Status != SubscriptionStatusTrial && s.Status != SubscriptionStatusPaused {
		return ErrCannotActivate
	}

	s.Status = SubscriptionStatusActive
	s.Touch()
	return nil
}

// Pause temporarily suspends the subscription
func (s *Subscription) Pause() error {
	if s.Status != SubscriptionStatusActive {
		return ErrCanOnlyPauseActive
	}

	s.Status = SubscriptionStatusPaused
	s.Touch()
	return nil
}

// Resume reactivates a paused subscription
func (s *Subscription) Resume() error {
	if s.Status != SubscriptionStatusPaused {
		return ErrCanOnlyResumePaused
	}

	s.Status = SubscriptionStatusActive
	s.Touch()
	return nil
}

// Cancel cancels the subscription
func (s *Subscription) Cancel(reason string, effectiveDate time.Time) error {
	if s.Status == SubscriptionStatusCancelled || s.Status == SubscriptionStatusExpired {
		return ErrAlreadyInTerminalStatus
	}

	now := time.Now()
	s.Status = SubscriptionStatusCancelled
	s.CancelledAt = &now
	s.CancelReason = reason
	s.CancellationEffectiveDate = &effectiveDate
	s.Touch()
	return nil
}

// Renew extends the subscription for another billing period
func (s *Subscription) Renew() error {
	if s.Status != SubscriptionStatusActive {
		return ErrCanOnlyRenewActive
	}

	s.RenewalDate = calculateRenewalDate(s.RenewalDate, s.BillingPeriod)
	s.Touch()
	return nil
}

// Expire marks the subscription as expired
func (s *Subscription) Expire() error {
	if s.Status == SubscriptionStatusExpired {
		return ErrAlreadyExpired
	}

	s.Status = SubscriptionStatusExpired
	now := time.Now()
	s.EndDate = &now
	s.Touch()
	return nil
}

// IsActive checks if subscription is currently active or in trial
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrial
}

// IsInTrial checks if subscription is in trial period
func (s *Subscription) IsInTrial() bool {
	if s.Status != SubscriptionStatusTrial || s.TrialEndDate == nil {
		return false
	}
	return time.Now().Before(*s.TrialEndDate)
}

// DaysUntilRenewal calculates days until next renewal
func (s *Subscription) DaysUntilRenewal() int {
	duration := time.Until(s.RenewalDate)
	return int(duration.Hours() / 24)
}

// IsPastDue checks if subscription is past its renewal date
func (s *Subscription) IsPastDue() bool {
	return time.Now().After(s.RenewalDate) && s.IsActive()
}

// Validate performs business rule validation
func (s *Subscription) Validate() error {
	if s.CustomerID == uuidv7.Nil {
		return ErrCustomerIDRequired
	}
	if s.PlanID == "" {
		return ErrPlanIDRequired
	}
	if s.Amount.Amount <= 0 {
		return ErrAmountMustBePositive
	}
	if s.StartDate.IsZero() {
		return ErrStartDateRequired
	}
	if s.RenewalDate.Before(s.StartDate) {
		return ErrRenewalDateInvalid
	}
	return nil
}
