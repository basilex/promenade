package subscription

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock Repository
type mockSubscriptionRepository struct {
	CreateFunc             func(ctx context.Context, subscription *Subscription) error
	GetByIDFunc            func(ctx context.Context, id uuidv7.UUID) (*Subscription, error)
	UpdateFunc             func(ctx context.Context, subscription *Subscription) error
	DeleteFunc             func(ctx context.Context, id uuidv7.UUID) error
	ListSubscriptionsFunc  func(ctx context.Context, page, pageSize int) ([]*Subscription, error)
	ListByCustomerFunc     func(ctx context.Context, customerID uuidv7.UUID) ([]*Subscription, error)
	ListByStatusFunc       func(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error)
	CountByStatusFunc      func(ctx context.Context, status SubscriptionStatus) (int64, error)
	GetTotalRevenueFunc    func(ctx context.Context) (int64, error)
}

func (m *mockSubscriptionRepository) Create(ctx context.Context, subscription *Subscription) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, subscription)
	}
	return nil
}

func (m *mockSubscriptionRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, ErrSubscriptionNotFound
}

func (m *mockSubscriptionRepository) Update(ctx context.Context, subscription *Subscription) error {
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

func (m *mockSubscriptionRepository) ListSubscriptions(ctx context.Context, page, pageSize int) ([]*Subscription, error) {
	if m.ListSubscriptionsFunc != nil {
		return m.ListSubscriptionsFunc(ctx, page, pageSize)
	}
	return nil, nil
}

func (m *mockSubscriptionRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*Subscription, error) {
	if m.ListByCustomerFunc != nil {
		return m.ListByCustomerFunc(ctx, customerID)
	}
	return nil, nil
}

func (m *mockSubscriptionRepository) ListByStatus(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, status)
	}
	return nil, nil
}

func (m *mockSubscriptionRepository) CountByStatus(ctx context.Context, status SubscriptionStatus) (int64, error) {
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
			CreateFunc: func(ctx context.Context, s *Subscription) error {
				assert.Equal(t, customerID, s.CustomerID)
				assert.Equal(t, planID, s.PlanID)
				assert.Equal(t, SubscriptionStatusTrial, s.Status)
				assert.Equal(t, BillingPeriodMonthly, s.BillingPeriod)
				assert.Equal(t, "USD", s.Amount.Currency)
				assert.Equal(t, amount, s.Amount.Amount)
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		subscription, err := uc.CreateSubscription(ctx, customerID, planID, BillingPeriodMonthly, "USD", amount, 14)

		require.NoError(t, err)
		assert.NotNil(t, subscription)
		assert.Equal(t, customerID, subscription.CustomerID)
		assert.Equal(t, SubscriptionStatusTrial, subscription.Status)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			CreateFunc: func(ctx context.Context, s *Subscription) error {
				return errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
		subscription, err := uc.CreateSubscription(ctx, customerID, planID, BillingPeriodMonthly, "USD", amount, 14)

		assert.Error(t, err)
		assert.Nil(t, subscription)
		assert.Contains(t, err.Error(), "failed to save subscription")
	})
}

// Test GetSubscription
func TestUseCase_GetSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		expected, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				assert.Equal(t, subscriptionID, id)
				return expected, nil
			},
		}

		uc := NewUseCase(mockRepo)
		subscription, err := uc.GetSubscription(ctx, subscriptionID)

		require.NoError(t, err)
		assert.Equal(t, expected, subscription)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		subscription, err := uc.GetSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
		assert.Nil(t, subscription)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
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
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				assert.Equal(t, newPlanID, s.PlanID)
				assert.Equal(t, newAmount, s.Amount.Amount)
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.UpdateSubscription(ctx, subscriptionID, newPlanID, newAmount)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.UpdateSubscription(ctx, subscriptionID, newPlanID, newAmount)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})

	t.Run("repository_update_error", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				return errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.UpdateSubscription(ctx, subscriptionID, newPlanID, newAmount)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update subscription")
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

		uc := NewUseCase(mockRepo)
		err := uc.DeleteSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.DeleteSubscription(ctx, subscriptionID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete subscription")
	})
}

// Test ListSubscriptions
func TestUseCase_ListSubscriptions(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []*Subscription{
			{CustomerID: uuidv7.New(), Status: SubscriptionStatusActive},
			{CustomerID: uuidv7.New(), Status: SubscriptionStatusTrial},
		}

		mockRepo := &mockSubscriptionRepository{
			ListSubscriptionsFunc: func(ctx context.Context, page, pageSize int) ([]*Subscription, error) {
				assert.Equal(t, 1, page)
				assert.Equal(t, 20, pageSize)
				return expected, nil
			},
		}

		uc := NewUseCase(mockRepo)
		subscriptions, err := uc.ListSubscriptions(ctx, 1, 20)

		require.NoError(t, err)
		assert.Equal(t, expected, subscriptions)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			ListSubscriptionsFunc: func(ctx context.Context, page, pageSize int) ([]*Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
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
		expected := []*Subscription{
			{CustomerID: customerID, Status: SubscriptionStatusActive},
			{CustomerID: customerID, Status: SubscriptionStatusPaused},
		}

		mockRepo := &mockSubscriptionRepository{
			ListByCustomerFunc: func(ctx context.Context, id uuidv7.UUID) ([]*Subscription, error) {
				assert.Equal(t, customerID, id)
				return expected, nil
			},
		}

		uc := NewUseCase(mockRepo)
		subscriptions, err := uc.ListByCustomer(ctx, customerID)

		require.NoError(t, err)
		assert.Equal(t, expected, subscriptions)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			ListByCustomerFunc: func(ctx context.Context, id uuidv7.UUID) ([]*Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
		subscriptions, err := uc.ListByCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Nil(t, subscriptions)
	})
}

// Test ListByStatus
func TestUseCase_ListByStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		expected := []*Subscription{
			{Status: SubscriptionStatusActive},
			{Status: SubscriptionStatusActive},
		}

		mockRepo := &mockSubscriptionRepository{
			ListByStatusFunc: func(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error) {
				assert.Equal(t, SubscriptionStatusActive, status)
				return expected, nil
			},
		}

		uc := NewUseCase(mockRepo)
		subscriptions, err := uc.ListByStatus(ctx, SubscriptionStatusActive)

		require.NoError(t, err)
		assert.Equal(t, expected, subscriptions)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			ListByStatusFunc: func(ctx context.Context, status SubscriptionStatus) ([]*Subscription, error) {
				return nil, errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
		subscriptions, err := uc.ListByStatus(ctx, SubscriptionStatusActive)

		assert.Error(t, err)
		assert.Nil(t, subscriptions)
	})
}

// Test ActivateSubscription
func TestUseCase_ActivateSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success_from_trial", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 14)
		existing.ID = subscriptionID
		existing.Status = SubscriptionStatusTrial

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				assert.Equal(t, SubscriptionStatusActive, s.Status)
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.ActivateSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.ActivateSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})

	t.Run("activation_error", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.Status = SubscriptionStatusCancelled // Cannot activate cancelled

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.ActivateSubscription(ctx, subscriptionID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot activate")
	})
}

// Test PauseSubscription
func TestUseCase_PauseSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				assert.Equal(t, SubscriptionStatusPaused, s.Status)
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.PauseSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.PauseSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})
}

// Test ResumeSubscription
func TestUseCase_ResumeSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = SubscriptionStatusPaused

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				assert.Equal(t, SubscriptionStatusActive, s.Status)
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.ResumeSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.ResumeSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})
}

// Test CancelSubscription
func TestUseCase_CancelSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()
	reason := "User requested cancellation"
	effectiveDate := time.Now().AddDate(0, 1, 0) // 1 month from now

	t.Run("success", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				assert.Equal(t, SubscriptionStatusCancelled, s.Status)
				assert.Equal(t, reason, s.CancelReason)
				assert.NotNil(t, s.CancelledAt)
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.CancelSubscription(ctx, subscriptionID, reason, effectiveDate)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.CancelSubscription(ctx, subscriptionID, reason, effectiveDate)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})
}

// Test RenewSubscription
func TestUseCase_RenewSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.RenewSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.RenewSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})
}

// Test ExpireSubscription
func TestUseCase_ExpireSubscription(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	t.Run("success", func(t *testing.T) {
		existing, _ := NewSubscription(uuidv7.New(), "plan_basic", BillingPeriodMonthly, "USD", 2999, time.Now(), 0)
		existing.ID = subscriptionID
		existing.Status = SubscriptionStatusActive

		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return existing, nil
			},
			UpdateFunc: func(ctx context.Context, s *Subscription) error {
				assert.Equal(t, SubscriptionStatusExpired, s.Status)
				return nil
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.ExpireSubscription(ctx, subscriptionID)

		require.NoError(t, err)
	})

	t.Run("not_found", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*Subscription, error) {
				return nil, ErrSubscriptionNotFound
			},
		}

		uc := NewUseCase(mockRepo)
		err := uc.ExpireSubscription(ctx, subscriptionID)

		assert.ErrorIs(t, err, ErrSubscriptionNotFound)
	})
}

// Test CountByStatus
func TestUseCase_CountByStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			CountByStatusFunc: func(ctx context.Context, status SubscriptionStatus) (int64, error) {
				assert.Equal(t, SubscriptionStatusActive, status)
				return 42, nil
			},
		}

		uc := NewUseCase(mockRepo)
		count, err := uc.CountByStatus(ctx, SubscriptionStatusActive)

		require.NoError(t, err)
		assert.Equal(t, int64(42), count)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := &mockSubscriptionRepository{
			CountByStatusFunc: func(ctx context.Context, status SubscriptionStatus) (int64, error) {
				return 0, errors.New("database error")
			},
		}

		uc := NewUseCase(mockRepo)
		count, err := uc.CountByStatus(ctx, SubscriptionStatusActive)

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

		uc := NewUseCase(mockRepo)
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

		uc := NewUseCase(mockRepo)
		revenue, err := uc.GetTotalRevenue(ctx)

		assert.Error(t, err)
		assert.Equal(t, int64(0), revenue)
	})
}
