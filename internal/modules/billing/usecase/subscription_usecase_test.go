package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestSubscriptionUseCase_CreateSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	planID := uuidv7.New()

	tests := []struct {
		name        string
		userID      uuidv7.UUID
		planID      uuidv7.UUID
		mockSetup   func(*mocks.MockISubscriptionRepository, *mocks.MockIPlanRepository)
		expectError bool
	}{
		{
			name:   "successful subscription creation - no trial",
			userID: userID,
			planID: planID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository, planRepo *mocks.MockIPlanRepository) {
				plan := &entity.Plan{
					ID:        planID,
					Name:      "Basic Plan",
					Amount:    999,
					Currency:  "USD",
					Interval:  entity.PlanIntervalMonthly,
					TrialDays: 0,
				}
				planRepo.On("GetByID", ctx, planID).Return(plan, nil)
				subRepo.On("Create", ctx, mock.AnythingOfType("*entity.Subscription")).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "successful subscription creation - with trial",
			userID: userID,
			planID: planID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository, planRepo *mocks.MockIPlanRepository) {
				plan := &entity.Plan{
					ID:        planID,
					Name:      "Pro Plan",
					Amount:    1999,
					Currency:  "USD",
					Interval:  entity.PlanIntervalMonthly,
					TrialDays: 14,
				}
				planRepo.On("GetByID", ctx, planID).Return(plan, nil)
				subRepo.On("Create", ctx, mock.AnythingOfType("*entity.Subscription")).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "plan not found",
			userID: userID,
			planID: planID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository, planRepo *mocks.MockIPlanRepository) {
				planRepo.On("GetByID", ctx, planID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:   "repository create error",
			userID: userID,
			planID: planID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository, planRepo *mocks.MockIPlanRepository) {
				plan := &entity.Plan{
					ID:        planID,
					Name:      "Basic Plan",
					TrialDays: 0,
				}
				planRepo.On("GetByID", ctx, planID).Return(plan, nil)
				subRepo.On("Create", ctx, mock.AnythingOfType("*entity.Subscription")).
					Return(errors.New("db error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo, mockPlanRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			subscription, err := uc.CreateSubscription(ctx, tt.userID, tt.planID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, subscription)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, subscription)
				assert.Equal(t, tt.userID, subscription.UserID)
				assert.Equal(t, tt.planID, subscription.PlanID)
			}

			mockSubRepo.AssertExpectations(t)
			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUseCase_GetSubscription(t *testing.T) {
	ctx := context.Background()
	subID := uuidv7.New()

	tests := []struct {
		name        string
		subID       uuidv7.UUID
		mockSetup   func(*mocks.MockISubscriptionRepository)
		expectError bool
		expectNil   bool
	}{
		{
			name:  "successful retrieval",
			subID: subID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				sub := &entity.Subscription{
					ID:     subID,
					UserID: uuidv7.New(),
					PlanID: uuidv7.New(),
					Status: entity.SubscriptionStatusActive,
				}
				subRepo.On("GetByID", ctx, subID).Return(sub, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name:  "subscription not found",
			subID: subID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("GetByID", ctx, subID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
		{
			name:  "repository error",
			subID: subID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("GetByID", ctx, subID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			subscription, err := uc.GetSubscription(ctx, tt.subID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, subscription)
			} else {
				assert.NotNil(t, subscription)
			}

			mockSubRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUseCase_GetUserSubscriptions(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	tests := []struct {
		name        string
		userID      uuidv7.UUID
		mockSetup   func(*mocks.MockISubscriptionRepository)
		expectError bool
		expectCount int
	}{
		{
			name:   "successful retrieval - multiple subscriptions",
			userID: userID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subs := []*entity.Subscription{
					{ID: uuidv7.New(), UserID: userID, Status: entity.SubscriptionStatusActive},
					{ID: uuidv7.New(), UserID: userID, Status: entity.SubscriptionStatusCanceled},
				}
				subRepo.On("GetByUserID", ctx, userID).Return(subs, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:   "no subscriptions found",
			userID: userID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("GetByUserID", ctx, userID).Return([]*entity.Subscription{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "repository error",
			userID: userID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			subscriptions, err := uc.GetUserSubscriptions(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, subscriptions)
			} else {
				assert.NoError(t, err)
				assert.Len(t, subscriptions, tt.expectCount)
			}

			mockSubRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUseCase_GetActiveSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	tests := []struct {
		name        string
		userID      uuidv7.UUID
		mockSetup   func(*mocks.MockISubscriptionRepository)
		expectError bool
		expectNil   bool
	}{
		{
			name:   "successful retrieval",
			userID: userID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				sub := &entity.Subscription{
					ID:     uuidv7.New(),
					UserID: userID,
					Status: entity.SubscriptionStatusActive,
				}
				subRepo.On("GetActiveByUserID", ctx, userID).Return(sub, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name:   "no active subscription",
			userID: userID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("GetActiveByUserID", ctx, userID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			subscription, err := uc.GetActiveSubscription(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, subscription)
			} else {
				assert.NotNil(t, subscription)
				assert.Equal(t, entity.SubscriptionStatusActive, subscription.Status)
			}

			mockSubRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUseCase_CancelSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	subID := uuidv7.New()

	tests := []struct {
		name        string
		userID      uuidv7.UUID
		subID       uuidv7.UUID
		mockSetup   func(*mocks.MockISubscriptionRepository)
		expectError bool
	}{
		{
			name:   "successful cancellation",
			userID: userID,
			subID:  subID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				sub := &entity.Subscription{
					ID:     subID,
					UserID: userID,
					Status: entity.SubscriptionStatusActive,
				}
				subRepo.On("GetByID", ctx, subID).Return(sub, nil)
				subRepo.On("Update", ctx, mock.MatchedBy(func(s *entity.Subscription) bool {
					// Cancel(false) sets CancelAtPeriodEnd=true but keeps status active
					return s.CancelAtPeriodEnd == true && s.CanceledAt != nil
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "subscription not found",
			userID: userID,
			subID:  subID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("GetByID", ctx, subID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:   "update error",
			userID: userID,
			subID:  subID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				sub := &entity.Subscription{
					ID:     subID,
					UserID: userID,
					Status: entity.SubscriptionStatusActive,
				}
				subRepo.On("GetByID", ctx, subID).Return(sub, nil)
				subRepo.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			err := uc.CancelSubscription(ctx, tt.userID, tt.subID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockSubRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUseCase_UpgradeSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	subID := uuidv7.New()
	oldPlanID := uuidv7.New()
	newPlanID := uuidv7.New()

	tests := []struct {
		name        string
		userID      uuidv7.UUID
		subID       uuidv7.UUID
		newPlanID   uuidv7.UUID
		mockSetup   func(*mocks.MockISubscriptionRepository)
		expectError bool
	}{
		{
			name:      "successful upgrade",
			userID:    userID,
			subID:     subID,
			newPlanID: newPlanID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				sub := &entity.Subscription{
					ID:     subID,
					UserID: userID,
					PlanID: oldPlanID,
					Status: entity.SubscriptionStatusActive,
				}
				subRepo.On("GetByID", ctx, subID).Return(sub, nil)
				subRepo.On("Update", ctx, mock.MatchedBy(func(s *entity.Subscription) bool {
					return s.PlanID == newPlanID
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "subscription not found",
			userID:    userID,
			subID:     subID,
			newPlanID: newPlanID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("GetByID", ctx, subID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:      "update error",
			userID:    userID,
			subID:     subID,
			newPlanID: newPlanID,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				sub := &entity.Subscription{
					ID:     subID,
					UserID: userID,
					PlanID: oldPlanID,
					Status: entity.SubscriptionStatusActive,
				}
				subRepo.On("GetByID", ctx, subID).Return(sub, nil)
				subRepo.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			subscription, err := uc.UpgradeSubscription(ctx, tt.userID, tt.subID, tt.newPlanID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, subscription)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, subscription)
				assert.Equal(t, tt.newPlanID, subscription.PlanID)
			}

			mockSubRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUseCase_ListSubscriptions(t *testing.T) {
	ctx := context.Background()
	activeStatus := entity.SubscriptionStatusActive
	cancelledStatus := entity.SubscriptionStatusCanceled

	tests := []struct {
		name        string
		status      *entity.SubscriptionStatus
		limit       int
		offset      int
		mockSetup   func(*mocks.MockISubscriptionRepository)
		expectError bool
		expectCount int
	}{
		{
			name:   "list all subscriptions",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subs := []*entity.Subscription{
					{ID: uuidv7.New(), Status: entity.SubscriptionStatusActive},
					{ID: uuidv7.New(), Status: entity.SubscriptionStatusCanceled},
				}
				subRepo.On("List", ctx, (*entity.SubscriptionStatus)(nil), 10, 0).Return(subs, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:   "list active subscriptions only",
			status: &activeStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subs := []*entity.Subscription{
					{ID: uuidv7.New(), Status: entity.SubscriptionStatusActive},
				}
				subRepo.On("List", ctx, &activeStatus, 10, 0).Return(subs, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "list with pagination",
			status: nil,
			limit:  5,
			offset: 10,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subs := []*entity.Subscription{
					{ID: uuidv7.New(), Status: entity.SubscriptionStatusActive},
				}
				subRepo.On("List", ctx, (*entity.SubscriptionStatus)(nil), 5, 10).Return(subs, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "empty result",
			status: &cancelledStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("List", ctx, &cancelledStatus, 10, 0).Return([]*entity.Subscription{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("List", ctx, (*entity.SubscriptionStatus)(nil), 10, 0).
					Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			subscriptions, err := uc.ListSubscriptions(ctx, tt.status, tt.limit, tt.offset)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, subscriptions)
			} else {
				assert.NoError(t, err)
				assert.Len(t, subscriptions, tt.expectCount)
			}

			mockSubRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriptionUseCase_CountSubscriptions(t *testing.T) {
	ctx := context.Background()
	activeStatus := entity.SubscriptionStatusActive

	tests := []struct {
		name          string
		status        *entity.SubscriptionStatus
		mockSetup     func(*mocks.MockISubscriptionRepository)
		expectError   bool
		expectedCount int
	}{
		{
			name:   "count all subscriptions",
			status: nil,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("Count", ctx, (*entity.SubscriptionStatus)(nil)).Return(42, nil)
			},
			expectError:   false,
			expectedCount: 42,
		},
		{
			name:   "count active subscriptions",
			status: &activeStatus,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("Count", ctx, &activeStatus).Return(10, nil)
			},
			expectError:   false,
			expectedCount: 10,
		},
		{
			name:   "zero count",
			status: nil,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("Count", ctx, (*entity.SubscriptionStatus)(nil)).Return(0, nil)
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			mockSetup: func(subRepo *mocks.MockISubscriptionRepository) {
				subRepo.On("Count", ctx, (*entity.SubscriptionStatus)(nil)).
					Return(0, errors.New("db error"))
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			mockPlanRepo := new(mocks.MockIPlanRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockSubRepo)

			uc := NewSubscriptionUseCase(mockSubRepo, mockPlanRepo, mockInvoiceRepo, nil)

			count, err := uc.CountSubscriptions(ctx, tt.status)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockSubRepo.AssertExpectations(t)
		})
	}
}
