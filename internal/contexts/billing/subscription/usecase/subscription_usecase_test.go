package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	subscriptionerrors "github.com/basilex/promenade/internal/contexts/billing/subscription"
	"github.com/basilex/promenade/internal/contexts/billing/subscription/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock Repository
type mockSubscriptionRepository struct {
	CreateFunc            func(ctx context.Context, subscription *aggregate.Subscription) error
	GetByIDFunc           func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error)
	UpdateFunc            func(ctx context.Context, subscription *aggregate.Subscription) error
	DeleteFunc            func(ctx context.Context, id uuidv7.UUID) error
	ListSubscriptionsFunc func(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error)
	ListByCustomerFunc    func(ctx context.Context, customerID uuidv7.UUID) ([]*aggregate.Subscription, error)
	ListByStatusFunc      func(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error)
	CountByStatusFunc     func(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error)
	GetTotalRevenueFunc   func(ctx context.Context) (int64, error)
}

func (m *mockSubscriptionRepository) Create(ctx context.Context, subscription *aggregate.Subscription) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, subscription)
	}
	return nil
}

func (m *mockSubscriptionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, subscriptionerrors.ErrSubscriptionNotFound
}

func (m *mockSubscriptionRepository) Update(ctx context.Context, subscription *aggregate.Subscription) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, subscription)
	}
	return nil
}

func (m *mockSubscriptionRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *mockSubscriptionRepository) ListSubscriptions(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error) {
	if m.ListSubscriptionsFunc != nil {
		return m.ListSubscriptionsFunc(ctx, page, pageSize)
	}
	return nil, nil
}

func (m *mockSubscriptionRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*aggregate.Subscription, error) {
	if m.ListByCustomerFunc != nil {
		return m.ListByCustomerFunc(ctx, customerID)
	}
	return nil, nil
}

func (m *mockSubscriptionRepository) ListByStatus(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, status)
	}
	return nil, nil
}

func (m *mockSubscriptionRepository) CountByStatus(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error) {
	if m.CountByStatusFunc != nil {
		return m.CountByStatusFunc(ctx, status)
	}
	return 0, nil
}

func (m *mockSubscriptionRepository) GetTotalRevenue(ctx context.Context) (int64, error) {
	if m.GetTotalRevenueFunc != nil {
		return m.GetTotalRevenueFunc(ctx)
	}
	return 0, nil
}

// Test CreateSubscription
func TestUseCase_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()
	planID := "plan_basic"
	amount := int64(2999) // $29.99

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			CreateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				assert.Equal(t, customerID, s.CustomerID)
				assert.Equal(t, planID, s.PlanID)
				assert.Equal(t, aggregate.SubscriptionStatusTrial, s.Status)
				assert.Equal(t, aggregate.BillingPeriodMonthly, s.BillingPeriod)
				assert.Equal(t, "USD", s.Amount.Currency)
				assert.Equal(t, amount, s.Amount.Amount)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscription, err := uc.CreateSubscription(ctx, customerID, planID, aggregate.BillingPeriodMonthly, "USD", amount, 14)

		require.NoError(t, err)
		assert.NotNil(t, subscription)
		assert.Equal(t, customerID, subscription.CustomerID)
		assert.Equal(t, aggregate.SubscriptionStatusTrial, subscription.Status)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			CreateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				return errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscription, err := uc.CreateSubscription(ctx, customerID, planID, aggregate.BillingPeriodMonthly, "USD", amount, 14)

		assert.Error(t, err)
		assert.Nil(t, subscription)
		// Now returns repository error directly, not wrapped
	})
}

// Test GetSubscription
func TestUseCase_GetSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		expected, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				assert.Equal(t, subscriptionID, id)
				return expected, nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscription, err := uc.GetSubscription(ctx, subscriptionID)

		require.NoError(t, err)
		assert.Equal(t, expected, subscription)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscription, err := uc.GetSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
		assert.Nil(t, subscription)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscription, err := uc.GetSubscription(ctx, subscriptionID)

		assert.Error(t, err)
		assert.Nil(t, subscription)
	})
}

// Test UpdateSubscription
func TestUseCase_UpdateSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()
	newPlanID := "plan_premium"
	newAmount := int64(4999) // $49.99

	t.Run("success", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = aggregate.SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				assert.Equal(t, newPlanID, s.PlanID)
				assert.Equal(t, newAmount, s.Amount.Amount)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.UpdateSubscription(ctx, subscriptionID, newPlanID, newAmount)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.UpdateSubscription(ctx, subscriptionID, newPlanID, newAmount)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
	})

	t.Run("repository_update_error", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				return errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.UpdateSubscription(ctx, subscriptionID, newPlanID, newAmount)

		assert.Error(t, err)
		// Now returns repository error directly
	})
}

// Test DeleteSubscription
func TestUseCase_DeleteSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
				assert.Equal(t, subscriptionID, id)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.DeleteSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.DeleteSubscription(ctx, subscriptionID)

		assert.Error(t, err)
		// Now returns repository error directly
	})
}

// Test ListSubscriptions
func TestUseCase_ListSubscriptions(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []*aggregate.Subscription{
			{CustomerID: uuidv7.New(), Status: aggregate.SubscriptionStatusActive},
			{CustomerID: uuidv7.New(), Status: aggregate.SubscriptionStatusTrial},
		}

		mockRepo := &mockSubscriptionRepository{
			ListSubscriptionsFunc: func(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error) {
				assert.Equal(t, 1, page)
				assert.Equal(t, 20, pageSize)
				return expected, nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscriptions, err := uc.ListSubscriptions(ctx, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, expected, subscriptions)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			ListSubscriptionsFunc: func(ctx context.Context, page, pageSize int) ([]*aggregate.Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscriptions, err := uc.ListSubscriptions(ctx, 1, 20)

		assert.Error(t, err)
		assert.Nil(t, subscriptions)
	})
}

// Test ListByCustomer
func TestUseCase_ListByCustomer(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		expected := []*aggregate.Subscription{
			{CustomerID: customerID, Status: aggregate.SubscriptionStatusActive},
			{CustomerID: customerID, Status: aggregate.SubscriptionStatusPaused},
		}

		mockRepo := &mockSubscriptionRepository{
			ListByCustomerFunc: func(ctx context.Context, id uuidv7.UUID) ([]*aggregate.Subscription, error) {
				assert.Equal(t, customerID, id)
				return expected, nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscriptions, err := uc.ListByCustomer(ctx, customerID)

		require.NoError(t, err)
		assert.Equal(t, expected, subscriptions)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			ListByCustomerFunc: func(ctx context.Context, id uuidv7.UUID) ([]*aggregate.Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscriptions, err := uc.ListByCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Nil(t, subscriptions)
	})
}

// Test ListByStatus
func TestUseCase_ListByStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []*aggregate.Subscription{
			{Status: aggregate.SubscriptionStatusActive},
			{Status: aggregate.SubscriptionStatusActive},
		}

		mockRepo := &mockSubscriptionRepository{
			ListByStatusFunc: func(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error) {
				assert.Equal(t, aggregate.SubscriptionStatusActive, status)
				return expected, nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscriptions, err := uc.ListByStatus(ctx, aggregate.SubscriptionStatusActive)

		require.NoError(t, err)
		assert.Equal(t, expected, subscriptions)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			ListByStatusFunc: func(ctx context.Context, status aggregate.SubscriptionStatus) ([]*aggregate.Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		subscriptions, err := uc.ListByStatus(ctx, aggregate.SubscriptionStatusActive)

		assert.Error(t, err)
		assert.Nil(t, subscriptions)
	})
}

// Test ActivateSubscription
func TestUseCase_ActivateSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success_from_trial", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 14)
		existing.ID = subscriptionID
		existing.Status = aggregate.SubscriptionStatusTrial

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				assert.Equal(t, aggregate.SubscriptionStatusActive, s.Status)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.ActivateSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.ActivateSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
	})

	t.Run("activation_error", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.Status = aggregate.SubscriptionStatusCancelled // Cannot activate cancelled

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.ActivateSubscription(ctx, subscriptionID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, subscriptionerrors.ErrCannotActivate)
	})
}

// Test PauseSubscription
func TestUseCase_PauseSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = aggregate.SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				assert.Equal(t, aggregate.SubscriptionStatusPaused, s.Status)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.PauseSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.PauseSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
	})
}

// Test ResumeSubscription
func TestUseCase_ResumeSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = aggregate.SubscriptionStatusPaused

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				assert.Equal(t, aggregate.SubscriptionStatusActive, s.Status)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.ResumeSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.ResumeSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
	})
}

// Test CancelSubscription
func TestUseCase_CancelSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()
	reason := "User requested cancellation"
	effectiveDate := time.Now().AddDate(0, 1, 0) // 1 month from now

	t.Run("success", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = aggregate.SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				assert.Equal(t, aggregate.SubscriptionStatusCancelled, s.Status)
				assert.Equal(t, reason, s.CancelReason)
				assert.NotNil(t, s.CancelledAt)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.CancelSubscription(ctx, subscriptionID, reason, effectiveDate)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.CancelSubscription(ctx, subscriptionID, reason, effectiveDate)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
	})
}

// Test RenewSubscription
func TestUseCase_RenewSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = aggregate.SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.RenewSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.RenewSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
	})
}

// Test ExpireSubscription
func TestUseCase_ExpireSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := aggregate.NewSubscription(uuidv7.New(), "plan_basic", aggregate.BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = aggregate.SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *aggregate.Subscription) error {
				assert.Equal(t, aggregate.SubscriptionStatusExpired, s.Status)
				return nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.ExpireSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.Subscription, error) {
				return nil, subscriptionerrors.ErrSubscriptionNotFound
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		err := uc.ExpireSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, subscriptionerrors.ErrSubscriptionNotFound)
	})
}

// Test CountByStatus
func TestUseCase_CountByStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			CountByStatusFunc: func(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error) {
				assert.Equal(t, aggregate.SubscriptionStatusActive, status)
				return 42, nil
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		count, err := uc.CountByStatus(ctx, aggregate.SubscriptionStatusActive)

		require.NoError(t, err)
		assert.Equal(t, int64(42), count)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			CountByStatusFunc: func(ctx context.Context, status aggregate.SubscriptionStatus) (int64, error) {
				return 0, errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		count, err := uc.CountByStatus(ctx, aggregate.SubscriptionStatusActive)

		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// Test GetTotalRevenue
func TestUseCase_GetTotalRevenue(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetTotalRevenueFunc: func(ctx context.Context) (int64, error) {
				return 999900, nil // $9,999.00
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		revenue, err := uc.GetTotalRevenue(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(999900), revenue)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetTotalRevenueFunc: func(ctx context.Context) (int64, error) {
				return 0, errors.New("database error")
			},
		}

		uc := NewSubscriptionUseCase(mockRepo)
		revenue, err := uc.GetTotalRevenue(ctx)

		assert.Error(t, err)
		assert.Equal(t, int64(0), revenue)
	})
}
