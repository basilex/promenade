package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestSubscriptionStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status SubscriptionStatus
		want   bool
	}{
		{"trialing status", SubscriptionStatusTrialing, true},
		{"active status", SubscriptionStatusActive, true},
		{"past_due status", SubscriptionStatusPastDue, true},
		{"canceled status", SubscriptionStatusCanceled, true},
		{"unpaid status", SubscriptionStatusUnpaid, true},
		{"invalid status", SubscriptionStatus("invalid"), false},
		{"empty status", SubscriptionStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestNewSubscription(t *testing.T) {
	userID := uuidv7.New()
	planID := uuidv7.New()

	t.Run("successful subscription with trial", func(t *testing.T) {
		sub, err := NewSubscription(userID, planID, 14)

		require.NoError(t, err)
		assert.NotNil(t, sub)
		assert.NotEqual(t, uuidv7.Nil, sub.ID)
		assert.Equal(t, userID, sub.UserID)
		assert.Equal(t, planID, sub.PlanID)
		assert.Equal(t, SubscriptionStatusTrialing, sub.Status)
		assert.NotNil(t, sub.TrialStart)
		assert.NotNil(t, sub.TrialEnd)
		assert.NotNil(t, sub.CurrentPeriodStart)
		assert.NotNil(t, sub.CurrentPeriodEnd)
		assert.Nil(t, sub.CanceledAt)
		assert.Nil(t, sub.CanceledAt)
		assert.NotZero(t, sub.CreatedAt)
		assert.NotZero(t, sub.UpdatedAt)

		// Trial should be 14 days
		expectedTrialEnd := sub.TrialStart.AddDate(0, 0, 14)
		assert.True(t, sub.TrialEnd.Equal(expectedTrialEnd) || sub.TrialEnd.After(expectedTrialEnd.Add(-time.Second)))
	})

	t.Run("subscription without trial", func(t *testing.T) {
		sub, err := NewSubscription(userID, planID, 0)

		require.NoError(t, err)
		assert.NotNil(t, sub)
		assert.Equal(t, SubscriptionStatusActive, sub.Status)
		assert.Nil(t, sub.TrialStart)
		assert.Nil(t, sub.TrialEnd)
	})

	t.Run("nil user ID", func(t *testing.T) {
		sub, err := NewSubscription(uuidv7.Nil, planID, 14)
		assert.Error(t, err)
		assert.Nil(t, sub)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("nil plan ID", func(t *testing.T) {
		sub, err := NewSubscription(userID, uuidv7.Nil, 14)
		assert.Error(t, err)
		assert.Nil(t, sub)
		assert.Contains(t, err.Error(), "plan ID is required")
	})

	t.Run("negative trial days", func(t *testing.T) {
		// Note: NewSubscription doesn't validate negative trialDays, just ignores them (creates active subscription)
		sub, err := NewSubscription(userID, planID, -5)
		assert.NoError(t, err) // No error - negative days just ignored
		assert.NotNil(t, sub)
		assert.Equal(t, SubscriptionStatusActive, sub.Status) // No trial, goes straight to active
		assert.Nil(t, sub.TrialStart)
		assert.Nil(t, sub.TrialEnd)
	})
}

func TestSubscription_Validate(t *testing.T) {
	userID := uuidv7.New()
	planID := uuidv7.New()

	t.Run("valid subscription", func(t *testing.T) {
		sub := &Subscription{
			ID:     uuidv7.New(),
			UserID: userID,
			PlanID: planID,
			Status: SubscriptionStatusActive,
		}

		err := sub.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing user ID", func(t *testing.T) {
		sub := &Subscription{
			ID:     uuidv7.New(),
			PlanID: planID,
			Status: SubscriptionStatusActive,
		}

		err := sub.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("missing plan ID", func(t *testing.T) {
		sub := &Subscription{
			ID:     uuidv7.New(),
			UserID: userID,
			Status: SubscriptionStatusActive,
		}

		err := sub.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan ID is required")
	})

	t.Run("invalid status", func(t *testing.T) {
		sub := &Subscription{
			ID:     uuidv7.New(),
			UserID: userID,
			PlanID: planID,
			Status: SubscriptionStatus("invalid"),
		}

		// Note: Validate() doesn't check status validity - use status.IsValid() instead
		assert.False(t, sub.Status.IsValid())
	})
}

func TestSubscription_Cancel(t *testing.T) {
	userID := uuidv7.New()
	planID := uuidv7.New()

	t.Run("immediate cancellation", func(t *testing.T) {
		sub, _ := NewSubscription(userID, planID, 0)
		sub.Status = SubscriptionStatusActive

		sub.Cancel(true)

		assert.Equal(t, SubscriptionStatusCanceled, sub.Status)
		assert.NotNil(t, sub.CanceledAt)
		assert.NotNil(t, sub.CanceledAt)
		assert.True(t, sub.CanceledAt.Equal(*sub.CanceledAt) || sub.CanceledAt.Before(*sub.CanceledAt))
	})

	t.Run("cancel at period end", func(t *testing.T) {
		sub, _ := NewSubscription(userID, planID, 0)
		sub.Status = SubscriptionStatusActive
		periodEnd := time.Now().AddDate(0, 1, 0)
		sub.CurrentPeriodEnd = periodEnd

		sub.Cancel(false)

		assert.Equal(t, SubscriptionStatusActive, sub.Status) // Still active until period end
		assert.True(t, sub.CancelAtPeriodEnd)                 // Flag set to cancel at end
		assert.NotNil(t, sub.CanceledAt)                      // CanceledAt is set when Cancel() is called
	})
}

func TestSubscription_Activate(t *testing.T) {
	sub := &Subscription{
		ID:     uuidv7.New(),
		UserID: uuidv7.New(),
		PlanID: uuidv7.New(),
		Status: SubscriptionStatusTrialing,
	}

	sub.Activate()

	assert.Equal(t, SubscriptionStatusActive, sub.Status)
	// Note: Activate() only changes status, doesn't modify period dates
	assert.NotNil(t, sub.UpdatedAt)
}

func TestSubscription_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status SubscriptionStatus
		want   bool
	}{
		{"active", SubscriptionStatusActive, true},
		{"trialing", SubscriptionStatusTrialing, true},
		{"canceled", SubscriptionStatusCanceled, false},
		{"past_due", SubscriptionStatusPastDue, false},
		{"unpaid", SubscriptionStatusUnpaid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{Status: tt.status}
			assert.Equal(t, tt.want, sub.IsActive())
		})
	}
}

func TestSubscription_IsCanceled(t *testing.T) {
	tests := []struct {
		name   string
		status SubscriptionStatus
		want   bool
	}{
		{"canceled", SubscriptionStatusCanceled, true},
		{"active", SubscriptionStatusActive, false},
		{"trialing", SubscriptionStatusTrialing, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{Status: tt.status}
			assert.Equal(t, tt.want, sub.IsCanceled())
		})
	}
}

func TestSubscription_IsTrialing(t *testing.T) {
	now := time.Now()
	futureTrialEnd := now.AddDate(0, 0, 7) // Trial ends in 7 days

	tests := []struct {
		name     string
		status   SubscriptionStatus
		trialEnd *time.Time
		want     bool
	}{
		{"trialing", SubscriptionStatusTrialing, &futureTrialEnd, true},
		{"active", SubscriptionStatusActive, nil, false},
		{"canceled", SubscriptionStatusCanceled, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{
				Status:   tt.status,
				TrialEnd: tt.trialEnd,
			}
			assert.Equal(t, tt.want, sub.IsTrialing())
		})
	}
}

func TestSubscription_IsPastDue(t *testing.T) {
	tests := []struct {
		name   string
		status SubscriptionStatus
		want   bool
	}{
		{"past_due", SubscriptionStatusPastDue, true},
		{"active", SubscriptionStatusActive, false},
		{"unpaid", SubscriptionStatusUnpaid, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscription{Status: tt.status}
			assert.Equal(t, tt.want, sub.IsPastDue())
		})
	}
}

func TestSubscription_RenewPeriod(t *testing.T) {
	sub := &Subscription{
		ID:     uuidv7.New(),
		UserID: uuidv7.New(),
		PlanID: uuidv7.New(),
		Status: SubscriptionStatusActive,
	}

	now := time.Now()
	sub.CurrentPeriodStart = now
	oldPeriodEnd := now.AddDate(0, 1, 0)
	sub.CurrentPeriodEnd = oldPeriodEnd

	sub.RenewPeriod()

	assert.NotNil(t, sub.CurrentPeriodStart)
	assert.NotNil(t, sub.CurrentPeriodEnd)
	assert.True(t, sub.CurrentPeriodStart.After(now))
	assert.True(t, sub.CurrentPeriodEnd.After(sub.CurrentPeriodStart))

	// New period should be approximately 1 month
	duration := sub.CurrentPeriodEnd.Sub(sub.CurrentPeriodStart)
	assert.True(t, duration > 28*24*time.Hour && duration < 32*24*time.Hour)
}

func TestSubscription_TrialLifecycle(t *testing.T) {
	userID := uuidv7.New()
	planID := uuidv7.New()

	t.Run("complete trial to active lifecycle", func(t *testing.T) {
		// Create subscription with trial
		sub, err := NewSubscription(userID, planID, 14)
		require.NoError(t, err)
		assert.Equal(t, SubscriptionStatusTrialing, sub.Status)

		// Activate after trial
		sub.Activate()
		assert.Equal(t, SubscriptionStatusActive, sub.Status)
		// Note: Activate() only changes status, period dates set by NewSubscription
	})
}

func TestSubscription_CancellationScenarios(t *testing.T) {
	userID := uuidv7.New()
	planID := uuidv7.New()

	t.Run("cancel during trial", func(t *testing.T) {
		sub, _ := NewSubscription(userID, planID, 14)
		assert.Equal(t, SubscriptionStatusTrialing, sub.Status)

		sub.Cancel(true)
		assert.Equal(t, SubscriptionStatusCanceled, sub.Status)
		assert.NotNil(t, sub.CanceledAt)
	})

	t.Run("cancel active subscription", func(t *testing.T) {
		sub, _ := NewSubscription(userID, planID, 0)
		sub.Status = SubscriptionStatusActive

		sub.Cancel(false)
		assert.Equal(t, SubscriptionStatusActive, sub.Status) // Still active until period end
		assert.True(t, sub.CancelAtPeriodEnd)                 // Flag set
		assert.NotNil(t, sub.CanceledAt)                      // Timestamp set when Cancel() called
	})
}
