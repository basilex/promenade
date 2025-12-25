package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestPlanUseCase_CreatePlan(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		planName      string
		slug          string
		description   string
		currency      string
		amount        int64
		interval      entity.PlanInterval
		features      []string
		mockSetup     func(*mockPlanRepository)
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful plan creation - monthly",
			planName:    "Basic Plan",
			slug:        "basic-monthly",
			description: "Basic features for small teams",
			currency:    "USD",
			amount:      999,
			interval:    entity.PlanIntervalMonthly,
			features:    []string{"Feature 1", "Feature 2"},
			mockSetup: func(m *mockPlanRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Plan")).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "successful plan creation - yearly",
			planName:    "Pro Plan",
			slug:        "pro-yearly",
			description: "Professional features",
			currency:    "EUR",
			amount:      9999,
			interval:    entity.PlanIntervalYearly,
			features:    []string{"All features", "Priority support"},
			mockSetup: func(m *mockPlanRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Plan")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "invalid plan name - empty",
			planName:      "",
			slug:          "basic-monthly",
			description:   "Description",
			currency:      "USD",
			amount:        999,
			interval:      entity.PlanIntervalMonthly,
			features:      []string{},
			mockSetup:     func(m *mockPlanRepository) {},
			expectError:   true,
			errorContains: "name",
		},
		{
			name:        "valid free plan - zero amount",
			planName:    "Free Plan",
			slug:        "free",
			description: "Free tier",
			currency:    "USD",
			amount:      0,
			interval:    entity.PlanIntervalMonthly,
			features:    []string{"Basic feature"},
			mockSetup: func(m *mockPlanRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Plan")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "invalid amount - negative",
			planName:      "Invalid Plan",
			slug:          "invalid",
			description:   "Invalid",
			currency:      "USD",
			amount:        -100,
			interval:      entity.PlanIntervalMonthly,
			features:      []string{},
			mockSetup:     func(m *mockPlanRepository) {},
			expectError:   true,
			errorContains: "amount",
		},
		{
			name:        "repository error",
			planName:    "Basic Plan",
			slug:        "basic-monthly",
			description: "Description",
			currency:    "USD",
			amount:      999,
			interval:    entity.PlanIntervalMonthly,
			features:    []string{},
			mockSetup: func(m *mockPlanRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Plan")).
					Return(errors.New("database error"))
			},
			expectError:   true,
			errorContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			plan, err := uc.CreatePlan(ctx, tt.planName, tt.slug, tt.description, tt.currency, tt.amount, tt.interval, tt.features)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, plan)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, plan)
				assert.Equal(t, tt.planName, plan.Name)
				assert.Equal(t, tt.slug, plan.Slug)
				assert.Equal(t, tt.description, plan.Description)
				assert.Equal(t, tt.currency, plan.Currency)
				assert.Equal(t, tt.amount, plan.Amount)
				assert.Equal(t, tt.interval, plan.Interval)
				assert.Equal(t, tt.features, plan.Features)
				assert.Equal(t, entity.PlanStatusActive, plan.Status)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_GetPlan(t *testing.T) {
	ctx := context.Background()
	planID := uuidv7.New()

	tests := []struct {
		name        string
		planID      uuidv7.UUID
		mockSetup   func(*mockPlanRepository)
		expectError bool
		expectNil   bool
	}{
		{
			name:   "successful retrieval",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:       planID,
					Name:     "Basic Plan",
					Slug:     "basic",
					Currency: "USD",
					Amount:   999,
					Interval: entity.PlanIntervalMonthly,
					Status:   entity.PlanStatusActive,
				}
				m.On("GetByID", ctx, planID).Return(plan, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name:   "plan not found",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				m.On("GetByID", ctx, planID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
		{
			name:   "repository error",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				m.On("GetByID", ctx, planID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			plan, err := uc.GetPlan(ctx, tt.planID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, plan)
			} else {
				assert.NotNil(t, plan)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_GetPlanBySlug(t *testing.T) {
	ctx := context.Background()
	slug := "basic-monthly"

	tests := []struct {
		name        string
		slug        string
		mockSetup   func(*mockPlanRepository)
		expectError bool
		expectNil   bool
	}{
		{
			name: "successful retrieval",
			slug: slug,
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:       uuidv7.New(),
					Name:     "Basic Plan",
					Slug:     slug,
					Currency: "USD",
					Amount:   999,
					Interval: entity.PlanIntervalMonthly,
					Status:   entity.PlanStatusActive,
				}
				m.On("GetBySlug", ctx, slug).Return(plan, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name: "plan not found",
			slug: "non-existent",
			mockSetup: func(m *mockPlanRepository) {
				m.On("GetBySlug", ctx, "non-existent").Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			plan, err := uc.GetPlanBySlug(ctx, tt.slug)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, plan)
			} else {
				assert.NotNil(t, plan)
				assert.Equal(t, tt.slug, plan.Slug)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_UpdatePlan(t *testing.T) {
	ctx := context.Background()
	planID := uuidv7.New()

	tests := []struct {
		name        string
		planID      uuidv7.UUID
		updates     map[string]any
		mockSetup   func(*mockPlanRepository)
		expectError bool
	}{
		{
			name:    "successful update",
			planID:  planID,
			updates: map[string]any{"name": "Updated Plan"},
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:       planID,
					Name:     "Original Plan",
					Slug:     "original",
					Currency: "USD",
					Amount:   999,
					Interval: entity.PlanIntervalMonthly,
					Status:   entity.PlanStatusActive,
				}
				m.On("GetByID", ctx, planID).Return(plan, nil)
				m.On("Update", ctx, plan).Return(nil)
			},
			expectError: false,
		},
		{
			name:    "plan not found",
			planID:  planID,
			updates: map[string]any{"name": "Updated"},
			mockSetup: func(m *mockPlanRepository) {
				m.On("GetByID", ctx, planID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:    "update error",
			planID:  planID,
			updates: map[string]any{"name": "Updated"},
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:     planID,
					Name:   "Original",
					Status: entity.PlanStatusActive,
				}
				m.On("GetByID", ctx, planID).Return(plan, nil)
				m.On("Update", ctx, plan).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			plan, err := uc.UpdatePlan(ctx, tt.planID, tt.updates)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, plan)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plan)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_DeletePlan(t *testing.T) {
	ctx := context.Background()
	planID := uuidv7.New()

	tests := []struct {
		name        string
		planID      uuidv7.UUID
		mockSetup   func(*mockPlanRepository)
		expectError bool
	}{
		{
			name:   "successful deletion",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				m.On("Delete", ctx, planID).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "plan not found",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				m.On("Delete", ctx, planID).Return(entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:   "deletion error",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				m.On("Delete", ctx, planID).Return(errors.New("deletion failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			err := uc.DeletePlan(ctx, tt.planID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_ListPlans(t *testing.T) {
	ctx := context.Background()
	activeStatus := entity.PlanStatusActive
	inactiveStatus := entity.PlanStatusInactive

	tests := []struct {
		name        string
		status      *entity.PlanStatus
		limit       int
		offset      int
		mockSetup   func(*mockPlanRepository)
		expectError bool
		expectCount int
	}{
		{
			name:   "list all plans",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mockPlanRepository) {
				plans := []*entity.Plan{
					{ID: uuidv7.New(), Name: "Plan 1", Status: entity.PlanStatusActive},
					{ID: uuidv7.New(), Name: "Plan 2", Status: entity.PlanStatusInactive},
				}
				m.On("List", ctx, (*entity.PlanStatus)(nil), 10, 0).Return(plans, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:   "list active plans only",
			status: &activeStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mockPlanRepository) {
				plans := []*entity.Plan{
					{ID: uuidv7.New(), Name: "Active Plan", Status: entity.PlanStatusActive},
				}
				m.On("List", ctx, &activeStatus, 10, 0).Return(plans, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "list with pagination",
			status: nil,
			limit:  5,
			offset: 10,
			mockSetup: func(m *mockPlanRepository) {
				plans := []*entity.Plan{
					{ID: uuidv7.New(), Name: "Plan 11", Status: entity.PlanStatusActive},
				}
				m.On("List", ctx, (*entity.PlanStatus)(nil), 5, 10).Return(plans, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "empty result",
			status: &inactiveStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mockPlanRepository) {
				m.On("List", ctx, &inactiveStatus, 10, 0).Return([]*entity.Plan{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mockPlanRepository) {
				m.On("List", ctx, (*entity.PlanStatus)(nil), 10, 0).
					Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			plans, err := uc.ListPlans(ctx, tt.status, tt.limit, tt.offset)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, plans)
			} else {
				assert.NoError(t, err)
				assert.Len(t, plans, tt.expectCount)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_CountPlans(t *testing.T) {
	ctx := context.Background()
	activeStatus := entity.PlanStatusActive

	tests := []struct {
		name          string
		status        *entity.PlanStatus
		mockSetup     func(*mockPlanRepository)
		expectError   bool
		expectedCount int
	}{
		{
			name:   "count all plans",
			status: nil,
			mockSetup: func(m *mockPlanRepository) {
				m.On("Count", ctx, (*entity.PlanStatus)(nil)).Return(42, nil)
			},
			expectError:   false,
			expectedCount: 42,
		},
		{
			name:   "count active plans",
			status: &activeStatus,
			mockSetup: func(m *mockPlanRepository) {
				m.On("Count", ctx, &activeStatus).Return(10, nil)
			},
			expectError:   false,
			expectedCount: 10,
		},
		{
			name:   "zero count",
			status: nil,
			mockSetup: func(m *mockPlanRepository) {
				m.On("Count", ctx, (*entity.PlanStatus)(nil)).Return(0, nil)
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			mockSetup: func(m *mockPlanRepository) {
				m.On("Count", ctx, (*entity.PlanStatus)(nil)).
					Return(0, errors.New("db error"))
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			count, err := uc.CountPlans(ctx, tt.status)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_ActivatePlan(t *testing.T) {
	ctx := context.Background()
	planID := uuidv7.New()

	tests := []struct {
		name        string
		planID      uuidv7.UUID
		mockSetup   func(*mockPlanRepository)
		expectError bool
	}{
		{
			name:   "successful activation",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:     planID,
					Name:   "Test Plan",
					Status: entity.PlanStatusInactive,
				}
				m.On("GetByID", ctx, planID).Return(plan, nil)
				m.On("Update", ctx, mock.MatchedBy(func(p *entity.Plan) bool {
					return p.Status == entity.PlanStatusActive
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "plan not found",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				m.On("GetByID", ctx, planID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:   "update error",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:     planID,
					Name:   "Test Plan",
					Status: entity.PlanStatusInactive,
				}
				m.On("GetByID", ctx, planID).Return(plan, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			err := uc.ActivatePlan(ctx, tt.planID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}

func TestPlanUseCase_DeactivatePlan(t *testing.T) {
	ctx := context.Background()
	planID := uuidv7.New()

	tests := []struct {
		name        string
		planID      uuidv7.UUID
		mockSetup   func(*mockPlanRepository)
		expectError bool
	}{
		{
			name:   "successful deactivation",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:     planID,
					Name:   "Test Plan",
					Status: entity.PlanStatusActive,
				}
				m.On("GetByID", ctx, planID).Return(plan, nil)
				m.On("Update", ctx, mock.MatchedBy(func(p *entity.Plan) bool {
					return p.Status == entity.PlanStatusInactive
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "plan not found",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				m.On("GetByID", ctx, planID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:   "update error",
			planID: planID,
			mockSetup: func(m *mockPlanRepository) {
				plan := &entity.Plan{
					ID:     planID,
					Name:   "Test Plan",
					Status: entity.PlanStatusActive,
				}
				m.On("GetByID", ctx, planID).Return(plan, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlanRepo := new(mockPlanRepository)
			mockSubRepo := new(mockSubscriptionRepository)
			tt.mockSetup(mockPlanRepo)

			uc := NewPlanUseCase(mockPlanRepo, mockSubRepo, nil)

			err := uc.DeactivatePlan(ctx, tt.planID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPlanRepo.AssertExpectations(t)
		})
	}
}
