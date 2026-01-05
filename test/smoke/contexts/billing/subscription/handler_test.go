package subscription_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/billing/subscription"
	subscriptionHTTP "github.com/basilex/promenade/internal/contexts/billing/subscription/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockSubscriptionUseCase implements subscription.IUseCase for testing
type MockSubscriptionUseCase struct {
	CreateSubscriptionFunc   func(ctx context.Context, customerID uuidv7.UUID, planID string, billingPeriod subscription.BillingPeriod, currency string, amount int64, trialDays int) (*subscription.Subscription, error)
	GetSubscriptionFunc      func(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error)
	UpdateSubscriptionFunc   func(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error
	DeleteSubscriptionFunc   func(ctx context.Context, id uuidv7.UUID) error
	ListSubscriptionsFunc    func(ctx context.Context, page, pageSize int) ([]*subscription.Subscription, error)
	ListByCustomerFunc       func(ctx context.Context, customerID uuidv7.UUID) ([]*subscription.Subscription, error)
	ListByStatusFunc         func(ctx context.Context, status subscription.SubscriptionStatus) ([]*subscription.Subscription, error)
	ActivateSubscriptionFunc func(ctx context.Context, id uuidv7.UUID) error
	PauseSubscriptionFunc    func(ctx context.Context, id uuidv7.UUID) error
	ResumeSubscriptionFunc   func(ctx context.Context, id uuidv7.UUID) error
	CancelSubscriptionFunc   func(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error
	RenewSubscriptionFunc    func(ctx context.Context, id uuidv7.UUID) error
	ExpireSubscriptionFunc   func(ctx context.Context, id uuidv7.UUID) error
	CountByStatusFunc        func(ctx context.Context, status subscription.SubscriptionStatus) (int64, error)
	GetTotalRevenueFunc      func(ctx context.Context) (int64, error)
}

// Implement all IUseCase methods
func (m *MockSubscriptionUseCase) CreateSubscription(ctx context.Context, customerID uuidv7.UUID, planID string, billingPeriod subscription.BillingPeriod, currency string, amount int64, trialDays int) (*subscription.Subscription, error) {
	if m.CreateSubscriptionFunc != nil {
		return m.CreateSubscriptionFunc(ctx, customerID, planID, billingPeriod, currency, amount, trialDays)
	}
	return nil, fmt.Errorf("CreateSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) GetSubscription(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error) {
	if m.GetSubscriptionFunc != nil {
		return m.GetSubscriptionFunc(ctx, id)
	}
	return nil, subscription.ErrSubscriptionNotFound
}

func (m *MockSubscriptionUseCase) UpdateSubscription(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error {
	if m.UpdateSubscriptionFunc != nil {
		return m.UpdateSubscriptionFunc(ctx, id, planID, amount)
	}
	return fmt.Errorf("UpdateSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) DeleteSubscription(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteSubscriptionFunc != nil {
		return m.DeleteSubscriptionFunc(ctx, id)
	}
	return fmt.Errorf("DeleteSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) ListSubscriptions(ctx context.Context, page, pageSize int) ([]*subscription.Subscription, error) {
	if m.ListSubscriptionsFunc != nil {
		return m.ListSubscriptionsFunc(ctx, page, pageSize)
	}
	return nil, fmt.Errorf("ListSubscriptionsFunc not implemented")
}

func (m *MockSubscriptionUseCase) ListByCustomer(ctx context.Context, customerID uuidv7.UUID) ([]*subscription.Subscription, error) {
	if m.ListByCustomerFunc != nil {
		return m.ListByCustomerFunc(ctx, customerID)
	}
	return nil, fmt.Errorf("ListByCustomerFunc not implemented")
}

func (m *MockSubscriptionUseCase) ListByStatus(ctx context.Context, status subscription.SubscriptionStatus) ([]*subscription.Subscription, error) {
	if m.ListByStatusFunc != nil {
		return m.ListByStatusFunc(ctx, status)
	}
	return nil, fmt.Errorf("ListByStatusFunc not implemented")
}

func (m *MockSubscriptionUseCase) ActivateSubscription(ctx context.Context, id uuidv7.UUID) error {
	if m.ActivateSubscriptionFunc != nil {
		return m.ActivateSubscriptionFunc(ctx, id)
	}
	return fmt.Errorf("ActivateSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) PauseSubscription(ctx context.Context, id uuidv7.UUID) error {
	if m.PauseSubscriptionFunc != nil {
		return m.PauseSubscriptionFunc(ctx, id)
	}
	return fmt.Errorf("PauseSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) ResumeSubscription(ctx context.Context, id uuidv7.UUID) error {
	if m.ResumeSubscriptionFunc != nil {
		return m.ResumeSubscriptionFunc(ctx, id)
	}
	return fmt.Errorf("ResumeSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) CancelSubscription(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error {
	if m.CancelSubscriptionFunc != nil {
		return m.CancelSubscriptionFunc(ctx, id, reason, effectiveDate)
	}
	return fmt.Errorf("CancelSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) RenewSubscription(ctx context.Context, id uuidv7.UUID) error {
	if m.RenewSubscriptionFunc != nil {
		return m.RenewSubscriptionFunc(ctx, id)
	}
	return fmt.Errorf("RenewSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) ExpireSubscription(ctx context.Context, id uuidv7.UUID) error {
	if m.ExpireSubscriptionFunc != nil {
		return m.ExpireSubscriptionFunc(ctx, id)
	}
	return fmt.Errorf("ExpireSubscriptionFunc not implemented")
}

func (m *MockSubscriptionUseCase) CountByStatus(ctx context.Context, status subscription.SubscriptionStatus) (int64, error) {
	if m.CountByStatusFunc != nil {
		return m.CountByStatusFunc(ctx, status)
	}
	return 0, fmt.Errorf("CountByStatusFunc not implemented")
}

func (m *MockSubscriptionUseCase) GetTotalRevenue(ctx context.Context) (int64, error) {
	if m.GetTotalRevenueFunc != nil {
		return m.GetTotalRevenueFunc(ctx)
	}
	return 0, fmt.Errorf("GetTotalRevenueFunc not implemented")
}

// fakeSubscription creates a fake subscription for testing
func fakeSubscription() *subscription.Subscription {
	sub, _ := subscription.NewSubscription(
		uuidv7.New(),
		"plan_premium",
		subscription.BillingPeriodMonthly,
		"USD",
		2999,
		time.Now(),
		14,
	)
	return sub
}

// TestSubscriptionHandler_Create tests subscription creation endpoint
func TestSubscriptionHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			CreateSubscriptionFunc: func(ctx context.Context, customerID uuidv7.UUID, planID string, billingPeriod subscription.BillingPeriod, currency string, amount int64, trialDays int) (*subscription.Subscription, error) {
				return fakeSubscription(), nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions", handler.Create)

		body := map[string]any{
			"customer_id":    smoke.FakeUUID(),
			"plan_id":        "plan_premium",
			"billing_period": "monthly",
			"currency":       "USD",
			"amount":         2999,
			"trial_days":     14,
		}

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions", body)
		smoke.AssertSuccessResponse(t, w, 201)
	})

	t.Run("validation_error", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{}
		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions", handler.Create)

		body := map[string]any{
			"customer_id": "invalid-uuid",
		}

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions", body)
		smoke.AssertErrorResponse(t, w, 400, "VALIDATION_ERROR")
	})
}

// TestSubscriptionHandler_GetByID tests subscription retrieval endpoint
func TestSubscriptionHandler_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			GetSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error) {
				return fakeSubscription(), nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.GET("/subscriptions/:id", handler.GetByID)

		w := smoke.MakeRequest(t, router, "GET", "/subscriptions/"+smoke.FakeUUID(), nil)
		smoke.AssertSuccessResponse(t, w, 200)
	})

	t.Run("not_found", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			GetSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error) {
				return nil, subscription.ErrSubscriptionNotFound
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.GET("/subscriptions/:id", handler.GetByID)

		w := smoke.MakeRequest(t, router, "GET", "/subscriptions/"+smoke.FakeUUID(), nil)
		smoke.AssertErrorResponse(t, w, 404, "SUBSCRIPTION_NOT_FOUND")
	})
}

// TestSubscriptionHandler_Update tests subscription update endpoint
func TestSubscriptionHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			UpdateSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error {
				return nil
			},
			GetSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error) {
				return fakeSubscription(), nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.PUT("/subscriptions/:id", handler.Update)

		body := map[string]any{
			"plan_id": "plan_enterprise",
			"amount":  4999,
		}

		w := smoke.MakeRequest(t, router, "PUT", "/subscriptions/"+smoke.FakeUUID(), body)
		smoke.AssertSuccessResponse(t, w, 200)
	})

	t.Run("not_found", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			UpdateSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID, planID string, amount int64) error {
				return subscription.ErrSubscriptionNotFound
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.PUT("/subscriptions/:id", handler.Update)

		body := map[string]any{
			"plan_id": "plan_enterprise",
			"amount":  4999,
		}

		w := smoke.MakeRequest(t, router, "PUT", "/subscriptions/"+smoke.FakeUUID(), body)
		smoke.AssertErrorResponse(t, w, 404, "SUBSCRIPTION_NOT_FOUND")
	})
}

// TestSubscriptionHandler_Delete tests subscription deletion endpoint
func TestSubscriptionHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			DeleteSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.DELETE("/subscriptions/:id", handler.Delete)

		w := smoke.MakeRequest(t, router, "DELETE", "/subscriptions/"+smoke.FakeUUID(), nil)
		smoke.AssertSuccessResponse(t, w, 200)
	})

	t.Run("not_found", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			DeleteSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return subscription.ErrSubscriptionNotFound
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.DELETE("/subscriptions/:id", handler.Delete)

		w := smoke.MakeRequest(t, router, "DELETE", "/subscriptions/"+smoke.FakeUUID(), nil)
		smoke.AssertErrorResponse(t, w, 404, "SUBSCRIPTION_NOT_FOUND")
	})
}

// TestSubscriptionHandler_List tests subscription list endpoint
func TestSubscriptionHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			ListSubscriptionsFunc: func(ctx context.Context, page, pageSize int) ([]*subscription.Subscription, error) {
				return []*subscription.Subscription{fakeSubscription()}, nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.GET("/subscriptions", handler.List)

		w := smoke.MakeRequest(t, router, "GET", "/subscriptions?page=1&page_size=10", nil)
		smoke.AssertSuccessResponse(t, w, 200)
	})

	t.Run("empty_result", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			ListSubscriptionsFunc: func(ctx context.Context, page, pageSize int) ([]*subscription.Subscription, error) {
				return []*subscription.Subscription{}, nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.GET("/subscriptions", handler.List)

		w := smoke.MakeRequest(t, router, "GET", "/subscriptions?page=1&page_size=10", nil)
		smoke.AssertSuccessResponse(t, w, 200)
	})
}

// TestSubscriptionHandler_Activate tests subscription activation endpoint
func TestSubscriptionHandler_Activate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			ActivateSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return nil
			},
			GetSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error) {
				return fakeSubscription(), nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions/:id/activate", handler.Activate)

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions/"+smoke.FakeUUID()+"/activate", nil)
		smoke.AssertSuccessResponse(t, w, 200)
	})

	t.Run("not_found", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			ActivateSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return subscription.ErrSubscriptionNotFound
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions/:id/activate", handler.Activate)

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions/"+smoke.FakeUUID()+"/activate", nil)
		smoke.AssertErrorResponse(t, w, 404, "SUBSCRIPTION_NOT_FOUND")
	})
}

// TestSubscriptionHandler_Pause tests subscription pause endpoint
func TestSubscriptionHandler_Pause(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			PauseSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return nil
			},
			GetSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error) {
				return fakeSubscription(), nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions/:id/pause", handler.Pause)

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions/"+smoke.FakeUUID()+"/pause", nil)
		smoke.AssertSuccessResponse(t, w, 200)
	})

	t.Run("not_found", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			PauseSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) error {
				return subscription.ErrSubscriptionNotFound
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions/:id/pause", handler.Pause)

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions/"+smoke.FakeUUID()+"/pause", nil)
		smoke.AssertErrorResponse(t, w, 404, "SUBSCRIPTION_NOT_FOUND")
	})
}

// TestSubscriptionHandler_Cancel tests subscription cancellation endpoint
func TestSubscriptionHandler_Cancel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			CancelSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error {
				return nil
			},
			GetSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID) (*subscription.Subscription, error) {
				return fakeSubscription(), nil
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions/:id/cancel", handler.Cancel)

		body := map[string]any{
			"reason":         "Customer request",
			"effective_date": time.Now().Format(time.RFC3339),
		}

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions/"+smoke.FakeUUID()+"/cancel", body)
		smoke.AssertSuccessResponse(t, w, 200)
	})

	t.Run("not_found", func(t *testing.T) {
		router := smoke.SetupRouter()

		mockUC := &MockSubscriptionUseCase{
			CancelSubscriptionFunc: func(ctx context.Context, id uuidv7.UUID, reason string, effectiveDate time.Time) error {
				return subscription.ErrSubscriptionNotFound
			},
		}

		handler := subscriptionHTTP.NewSubscriptionHandler(mockUC)
		router.POST("/subscriptions/:id/cancel", handler.Cancel)

		body := map[string]any{
			"reason":         "Customer request",
			"effective_date": time.Now().Format(time.RFC3339),
		}

		w := smoke.MakeRequest(t, router, "POST", "/subscriptions/"+smoke.FakeUUID()+"/cancel", body)
		smoke.AssertErrorResponse(t, w, 404, "SUBSCRIPTION_NOT_FOUND")
	})
}
