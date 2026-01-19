package aggregate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewSubscription(t *testing.T) {
	customerID := uuidv7.New()
	startDate := time.Now()

	tests := []struct {
		name           string
		customerID     uuidv7.UUID
		planID         string
		billingPeriod  BillingPeriod
		currency       string
		amount         int64
		trialDays      int
		wantErr        bool
		expectedStatus SubscriptionStatus
	}{
		{
			name:           "valid monthly subscription",
			customerID:     customerID,
			planID:         "basic",
			billingPeriod:  BillingPeriodMonthly,
			currency:       "USD",
			amount:         9900,
			trialDays:      0,
			wantErr:        false,
			expectedStatus: SubscriptionStatusActive,
		},
		{
			name:           "valid subscription with trial",
			customerID:     customerID,
			planID:         "pro",
			billingPeriod:  BillingPeriodMonthly,
			currency:       "USD",
			amount:         19900,
			trialDays:      14,
			wantErr:        false,
			expectedStatus: SubscriptionStatusTrial,
		},
		{
			name:           "valid quarterly subscription",
			customerID:     customerID,
			planID:         "enterprise",
			billingPeriod:  BillingPeriodQuarterly,
			currency:       "EUR",
			amount:         49900,
			trialDays:      0,
			wantErr:        false,
			expectedStatus: SubscriptionStatusActive,
		},
		{
			name:           "valid yearly subscription",
			customerID:     customerID,
			planID:         "premium",
			billingPeriod:  BillingPeriodYearly,
			currency:       "USD",
			amount:         99900,
			trialDays:      0,
			wantErr:        false,
			expectedStatus: SubscriptionStatusActive,
		},
		{
			name:          "missing customer ID",
			customerID:    uuidv7.Nil,
			planID:        "basic",
			billingPeriod: BillingPeriodMonthly,
			currency:      "USD",
			amount:        9900,
			trialDays:     0,
			wantErr:       true,
		},
		{
			name:          "missing plan ID",
			customerID:    customerID,
			planID:        "",
			billingPeriod: BillingPeriodMonthly,
			currency:      "USD",
			amount:        9900,
			trialDays:     0,
			wantErr:       true,
		},
		{
			name:          "missing currency",
			customerID:    customerID,
			planID:        "basic",
			billingPeriod: BillingPeriodMonthly,
			currency:      "",
			amount:        9900,
			trialDays:     0,
			wantErr:       true,
		},
		{
			name:          "zero amount",
			customerID:    customerID,
			planID:        "basic",
			billingPeriod: BillingPeriodMonthly,
			currency:      "USD",
			amount:        0,
			trialDays:     0,
			wantErr:       true,
		},
		{
			name:          "negative amount",
			customerID:    customerID,
			planID:        "basic",
			billingPeriod: BillingPeriodMonthly,
			currency:      "USD",
			amount:        -9900,
			trialDays:     0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := NewSubscription(
				tt.customerID,
				tt.planID,
				tt.billingPeriod,
				tt.currency,
				tt.amount,
				startDate,
				tt.trialDays,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, sub)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, sub)
			assert.NotEqual(t, uuidv7.Nil, sub.GetID())
			assert.Equal(t, tt.customerID, sub.CustomerID)
			assert.Equal(t, tt.planID, sub.PlanID)
			assert.Equal(t, tt.expectedStatus, sub.Status)
			assert.Equal(t, tt.billingPeriod, sub.BillingPeriod)
			assert.Equal(t, tt.currency, sub.Currency)
			assert.Equal(t, tt.amount, sub.Amount.Amount)
			assert.NotEmpty(t, sub.SubscriptionNo)
			assert.False(t, sub.StartDate.IsZero())
			assert.False(t, sub.RenewalDate.IsZero())

			if tt.trialDays > 0 {
				assert.NotNil(t, sub.TrialEndDate)
			} else {
				assert.Nil(t, sub.TrialEndDate)
			}
		})
	}
}

func TestSubscription_Activate(t *testing.T) {
	tests := []struct {
		name    string
		status  SubscriptionStatus
		wantErr bool
	}{
		{"from trial", SubscriptionStatusTrial, false},
		{"from paused", SubscriptionStatusPaused, false},
		{"from active", SubscriptionStatusActive, true},
		{"from cancelled", SubscriptionStatusCancelled, true},
		{"from expired", SubscriptionStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status

			err := sub.Activate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, sub.Status) // Status unchanged
			} else {
				assert.NoError(t, err)
				assert.Equal(t, SubscriptionStatusActive, sub.Status)
			}
		})
	}
}

func TestSubscription_Pause(t *testing.T) {
	tests := []struct {
		name    string
		status  SubscriptionStatus
		wantErr bool
	}{
		{"from active", SubscriptionStatusActive, false},
		{"from trial", SubscriptionStatusTrial, true},
		{"from paused", SubscriptionStatusPaused, true},
		{"from cancelled", SubscriptionStatusCancelled, true},
		{"from expired", SubscriptionStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status

			err := sub.Pause()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, sub.Status) // Status unchanged
			} else {
				assert.NoError(t, err)
				assert.Equal(t, SubscriptionStatusPaused, sub.Status)
			}
		})
	}
}

func TestSubscription_Resume(t *testing.T) {
	tests := []struct {
		name    string
		status  SubscriptionStatus
		wantErr bool
	}{
		{"from paused", SubscriptionStatusPaused, false},
		{"from trial", SubscriptionStatusTrial, true},
		{"from active", SubscriptionStatusActive, true},
		{"from cancelled", SubscriptionStatusCancelled, true},
		{"from expired", SubscriptionStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status

			err := sub.Resume()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, sub.Status) // Status unchanged
			} else {
				assert.NoError(t, err)
				assert.Equal(t, SubscriptionStatusActive, sub.Status)
			}
		})
	}
}

func TestSubscription_Cancel(t *testing.T) {
	reason := "No longer needed"
	effectiveDate := time.Now().AddDate(0, 0, 30)

	tests := []struct {
		name    string
		status  SubscriptionStatus
		wantErr bool
	}{
		{"from active", SubscriptionStatusActive, false},
		{"from trial", SubscriptionStatusTrial, false},
		{"from paused", SubscriptionStatusPaused, false},
		{"from cancelled", SubscriptionStatusCancelled, true},
		{"from expired", SubscriptionStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status

			err := sub.Cancel(reason, effectiveDate)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, sub.Status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, SubscriptionStatusCancelled, sub.Status)
				assert.NotNil(t, sub.CancelledAt)
				assert.Equal(t, reason, sub.CancelReason)
				assert.NotNil(t, sub.CancellationEffectiveDate)
				assert.Equal(t, effectiveDate, *sub.CancellationEffectiveDate)
			}
		})
	}
}

func TestSubscription_Renew(t *testing.T) {
	tests := []struct {
		name    string
		status  SubscriptionStatus
		wantErr bool
	}{
		{"from active", SubscriptionStatusActive, false},
		{"from trial", SubscriptionStatusTrial, true},
		{"from paused", SubscriptionStatusPaused, true},
		{"from cancelled", SubscriptionStatusCancelled, true},
		{"from expired", SubscriptionStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status
			originalRenewalDate := sub.RenewalDate

			err := sub.Renew()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, originalRenewalDate, sub.RenewalDate)
			} else {
				assert.NoError(t, err)
				assert.True(t, sub.RenewalDate.After(originalRenewalDate))
			}
		})
	}
}

func TestSubscription_Expire(t *testing.T) {
	tests := []struct {
		name    string
		status  SubscriptionStatus
		wantErr bool
	}{
		{"from active", SubscriptionStatusActive, false},
		{"from trial", SubscriptionStatusTrial, false},
		{"from paused", SubscriptionStatusPaused, false},
		{"from cancelled", SubscriptionStatusCancelled, false},
		{"from expired", SubscriptionStatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status

			err := sub.Expire()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.status, sub.Status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, SubscriptionStatusExpired, sub.Status)
				assert.NotNil(t, sub.EndDate)
			}
		})
	}
}

func TestSubscription_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   SubscriptionStatus
		expected bool
	}{
		{"active", SubscriptionStatusActive, true},
		{"trial", SubscriptionStatusTrial, true},
		{"paused", SubscriptionStatusPaused, false},
		{"cancelled", SubscriptionStatusCancelled, false},
		{"expired", SubscriptionStatusExpired, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status

			result := sub.IsActive()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscription_IsInTrial(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		status       SubscriptionStatus
		trialEndDate *time.Time
		expected     bool
	}{
		{
			name:         "in trial period",
			status:       SubscriptionStatusTrial,
			trialEndDate: ptrTime(now.AddDate(0, 0, 7)),
			expected:     true,
		},
		{
			name:         "trial expired",
			status:       SubscriptionStatusTrial,
			trialEndDate: ptrTime(now.AddDate(0, 0, -1)),
			expected:     false,
		},
		{
			name:         "active status",
			status:       SubscriptionStatusActive,
			trialEndDate: ptrTime(now.AddDate(0, 0, 7)),
			expected:     false,
		},
		{
			name:         "no trial end date",
			status:       SubscriptionStatusTrial,
			trialEndDate: nil,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status
			sub.TrialEndDate = tt.trialEndDate

			result := sub.IsInTrial()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscription_DaysUntilRenewal(t *testing.T) {
	sub := createTestSubscription()
	sub.RenewalDate = time.Now().AddDate(0, 0, 10)

	days := sub.DaysUntilRenewal()

	assert.GreaterOrEqual(t, days, 9) // Allow for timing differences
	assert.LessOrEqual(t, days, 10)
}

func TestSubscription_IsPastDue(t *testing.T) {
	tests := []struct {
		name        string
		status      SubscriptionStatus
		renewalDate time.Time
		expected    bool
	}{
		{
			name:        "active and past due",
			status:      SubscriptionStatusActive,
			renewalDate: time.Now().AddDate(0, 0, -1),
			expected:    true,
		},
		{
			name:        "active and not due",
			status:      SubscriptionStatusActive,
			renewalDate: time.Now().AddDate(0, 0, 10),
			expected:    false,
		},
		{
			name:        "trial and past renewal",
			status:      SubscriptionStatusTrial,
			renewalDate: time.Now().AddDate(0, 0, -1),
			expected:    true,
		},
		{
			name:        "paused and past due",
			status:      SubscriptionStatusPaused,
			renewalDate: time.Now().AddDate(0, 0, -1),
			expected:    false,
		},
		{
			name:        "cancelled and past due",
			status:      SubscriptionStatusCancelled,
			renewalDate: time.Now().AddDate(0, 0, -1),
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createTestSubscription()
			sub.Status = tt.status
			sub.RenewalDate = tt.renewalDate

			result := sub.IsPastDue()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscription_Validate(t *testing.T) {
	validSub := createTestSubscription()

	tests := []struct {
		name     string
		modifier func(*Subscription)
		wantErr  bool
	}{
		{
			name:     "valid subscription",
			modifier: func(s *Subscription) {},
			wantErr:  false,
		},
		{
			name: "missing customer ID",
			modifier: func(s *Subscription) {
				s.CustomerID = uuidv7.Nil
			},
			wantErr: true,
		},
		{
			name: "missing plan ID",
			modifier: func(s *Subscription) {
				s.PlanID = ""
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			modifier: func(s *Subscription) {
				s.Amount.Amount = 0
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			modifier: func(s *Subscription) {
				s.Amount.Amount = -100
			},
			wantErr: true,
		},
		{
			name: "zero start date",
			modifier: func(s *Subscription) {
				s.StartDate = time.Time{}
			},
			wantErr: true,
		},
		{
			name: "renewal before start",
			modifier: func(s *Subscription) {
				s.RenewalDate = s.StartDate.AddDate(0, 0, -1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := *validSub // Copy
			tt.modifier(&sub)

			err := sub.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalculateRenewalDate(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		billingPeriod BillingPeriod
		expected      time.Time
	}{
		{
			name:          "monthly",
			billingPeriod: BillingPeriodMonthly,
			expected:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "quarterly",
			billingPeriod: BillingPeriodQuarterly,
			expected:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "yearly",
			billingPeriod: BillingPeriodYearly,
			expected:      time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRenewalDate(startDate, tt.billingPeriod)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSubscriptionNumber(t *testing.T) {
	number := generateSubscriptionNumber()

	assert.NotEmpty(t, number)
	assert.Contains(t, number, "SUB-")
	assert.Contains(t, number, "2026")
	assert.Len(t, number, 15) // SUB-YYYY-NNNNNN = 15 chars
}

// Helper functions
func createTestSubscription() *Subscription {
	sub, _ := NewSubscription(
		uuidv7.New(),
		"basic",
		BillingPeriodMonthly,
		"USD",
		9900,
		time.Now(),
		0,
	)
	return sub
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
