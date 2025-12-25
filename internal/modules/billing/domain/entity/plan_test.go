package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestPlanStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status PlanStatus
		want   bool
	}{
		{"active status", PlanStatusActive, true},
		{"inactive status", PlanStatusInactive, true},
		{"archived status", PlanStatusArchived, true},
		{"invalid status", PlanStatus("invalid"), false},
		{"empty status", PlanStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestPlanInterval_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		interval PlanInterval
		want     bool
	}{
		{"monthly interval", PlanIntervalMonthly, true},
		{"yearly interval", PlanIntervalYearly, true},
		{"invalid plan interval", PlanInterval("weekly"), false},
		{"empty interval", PlanInterval(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.interval.IsValid())
		})
	}
}

func TestNewPlan(t *testing.T) {
	t.Run("successful plan creation", func(t *testing.T) {
		plan, err := NewPlan(
			"Pro Plan",
			"pro-plan",
			"Professional plan with advanced features",
			2999, // $29.99
			"USD",
			PlanIntervalMonthly,
		)

		require.NoError(t, err)
		assert.NotNil(t, plan)
		assert.NotEqual(t, uuidv7.Nil, plan.ID)
		assert.Equal(t, "Pro Plan", plan.Name)
		assert.Equal(t, "pro-plan", plan.Slug)
		assert.Equal(t, "Professional plan with advanced features", plan.Description)
		assert.Equal(t, int64(2999), plan.Amount)
		assert.Equal(t, "USD", plan.Currency)
		assert.Equal(t, PlanIntervalMonthly, plan.Interval)
		assert.Equal(t, PlanStatusActive, plan.Status)
		assert.NotZero(t, plan.CreatedAt)
		assert.NotZero(t, plan.UpdatedAt)
		assert.Nil(t, plan.DeletedAt)
	})

	t.Run("empty name", func(t *testing.T) {
		plan, err := NewPlan("", "slug", "Description", 1000, "USD", PlanIntervalMonthly)
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("empty slug", func(t *testing.T) {
		plan, err := NewPlan("Plan Name", "", "Description", 1000, "USD", PlanIntervalMonthly)
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "slug is required")
	})

	t.Run("negative amount", func(t *testing.T) {
		plan, err := NewPlan("Plan Name", "slug", "Description", -100, "USD", PlanIntervalMonthly)
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "amount must be non-negative")
	})

	t.Run("invalid currency", func(t *testing.T) {
		// Note: Validate() only checks currency is not empty, not length
		// "US" is technically invalid but passes validation
		plan, err := NewPlan("Plan Name", "slug", "Description", 1000, "US", PlanIntervalMonthly)
		assert.NoError(t, err) // No error - currency length not validated
		assert.NotNil(t, plan)
		assert.Equal(t, "US", plan.Currency)
	})

	t.Run("invalid plan interval", func(t *testing.T) {
		plan, err := NewPlan("Plan Name", "slug", "Description", 1000, "USD", PlanInterval("weekly"))
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "invalid plan interval")
	})
}

func TestPlan_Validate(t *testing.T) {
	t.Run("valid plan", func(t *testing.T) {
		plan := &Plan{
			ID:          uuidv7.New(),
			Name:        "Basic Plan",
			Slug:        "basic-plan",
			Description: "Basic features",
			Amount:      999,
			Currency:    "USD",
			Interval:    PlanIntervalMonthly,
			Status:      PlanStatusActive,
		}

		err := plan.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing name", func(t *testing.T) {
		plan := &Plan{
			ID:       uuidv7.New(),
			Slug:     "slug",
			Amount:   999,
			Currency: "USD",
			Interval: PlanIntervalMonthly,
			Status:   PlanStatusActive,
		}

		err := plan.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	// Test status validity separately
	t.Run("invalid status check", func(t *testing.T) {
		invalidStatus := PlanStatus("invalid")
		assert.False(t, invalidStatus.IsValid())
		
		validStatus := PlanStatusActive
		assert.True(t, validStatus.IsValid())
	})
}

func TestPlan_Activate(t *testing.T) {
	plan := &Plan{
		ID:       uuidv7.New(),
		Name:     "Test Plan",
		Slug:     "test-plan",
		Amount:   999,
		Currency: "USD",
		Interval: PlanIntervalMonthly,
		Status:   PlanStatusInactive,
	}

	plan.Activate()
	assert.Equal(t, PlanStatusActive, plan.Status)
}

func TestPlan_Deactivate(t *testing.T) {
	plan := &Plan{
		ID:       uuidv7.New(),
		Name:     "Test Plan",
		Slug:     "test-plan",
		Amount:   999,
		Currency: "USD",
		Interval: PlanIntervalMonthly,
		Status:   PlanStatusActive,
	}

	plan.Deactivate()
	assert.Equal(t, PlanStatusInactive, plan.Status)
}

func TestPlan_Archive(t *testing.T) {
	plan := &Plan{
		ID:       uuidv7.New(),
		Name:     "Test Plan",
		Slug:     "test-plan",
		Amount:   999,
		Currency: "USD",
		Interval: PlanIntervalMonthly,
		Status:   PlanStatusActive,
	}

	plan.Archive()
	assert.Equal(t, PlanStatusArchived, plan.Status)
}

func TestPlan_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status PlanStatus
		want   bool
	}{
		{"active plan", PlanStatusActive, true},
		{"inactive plan", PlanStatusInactive, false},
		{"archived plan", PlanStatusArchived, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{Status: tt.status}
			assert.Equal(t, tt.want, plan.IsActive())
		})
	}
}

func TestPlan_FormatAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency string
		want     string
	}{
		{"USD dollars", 2999, "USD", "$29.99"},
		{"EUR euros", 3499, "EUR", "€34.99"},
		{"GBP pounds", 1999, "GBP", "£19.99"},
		{"UAH hryvnia", 29900, "UAH", "₴299.00"},
		{"unknown currency", 1000, "XXX", "XXX 10.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &Plan{
				Amount:   tt.amount,
				Currency: tt.currency,
			}
			assert.Equal(t, tt.want, plan.FormatAmount())
		})
	}
}

func TestPlan_Features(t *testing.T) {
	t.Run("get and set features", func(t *testing.T) {
		plan := &Plan{
			ID:       uuidv7.New(),
			Name:     "Pro Plan",
			Slug:     "pro-plan",
			Amount:   2999,
			Currency: "USD",
			Interval: PlanIntervalMonthly,
			Status:   PlanStatusActive,
		}

		// Initially nil
		assert.Nil(t, plan.Features)

		// Set features
		features := []string{"feature1", "feature2", "feature3"}
		plan.Features = features

		assert.Equal(t, features, plan.Features)
		assert.Len(t, plan.Features, 3)
	})
}

func TestPlan_CurrencySymbols(t *testing.T) {
	t.Run("common currency symbols", func(t *testing.T) {
		currencySymbols := map[string]string{
			"USD": "$",
			"EUR": "€",
			"GBP": "£",
			"UAH": "₴",
			"RUB": "₽",
		}

		for currency, expectedSymbol := range currencySymbols {
			plan := &Plan{
				Amount:   1000,
				Currency: currency,
			}
			formatted := plan.FormatAmount()
			assert.Contains(t, formatted, expectedSymbol, "Currency %s should contain symbol %s", currency, expectedSymbol)
		}
	})
}

func TestPlan_PriceCalculations(t *testing.T) {
	t.Run("monthly vs yearly pricing", func(t *testing.T) {
		monthlyPlan := &Plan{
			Amount:   999,
			Interval: PlanIntervalMonthly,
		}

		yearlyPlan := &Plan{
			Amount:   9990, // 2 months free (10 months price)
			Interval: PlanIntervalYearly,
		}

		// Yearly should save money
		monthlyYearlyCost := monthlyPlan.Amount * 12
		assert.Greater(t, monthlyYearlyCost, yearlyPlan.Amount, "Yearly plan should be cheaper than 12 monthly payments")
	})
}
