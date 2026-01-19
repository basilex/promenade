package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	interactionerrors "github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// MockRepository is a mock implementation of IRepository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, interaction *aggregate.Interaction) error {
	args := m.Called(ctx, interaction)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Interaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregate.Interaction), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, interaction *aggregate.Interaction) error {
	args := m.Called(ctx, interaction)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListByCustomer(ctx context.Context, customerID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	args := m.Called(ctx, customerID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Interaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) ListByCompany(ctx context.Context, companyID uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	args := m.Called(ctx, companyID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Interaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) ListByType(ctx context.Context, interactionType string, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	args := m.Called(ctx, interactionType, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Interaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) ListByCreatedBy(ctx context.Context, createdBy uuidv7.UUID, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	args := m.Called(ctx, createdBy, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Interaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) ListPendingFollowUps(ctx context.Context, page, pageSize int) ([]*aggregate.Interaction, int64, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*aggregate.Interaction), args.Get(1).(int64), args.Error(2)
}

// Test Helpers
func createTestUseCase() (*InteractionUseCase, *MockRepository) {
	mockRepo := new(MockRepository)
	uc := NewInteractionUseCase(mockRepo).(*InteractionUseCase)
	return uc, mockRepo
}

// Tests
func TestUseCase_CreateInteraction(t *testing.T) {
	ctx := context.Background()
	customerID := uuidv7.New()
	createdBy := uuidv7.New()
	startedAt := time.Now()

	t.Run("successful creation", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()

		mockRepo.On("Create", ctx, mock.AnythingOfType("*aggregate.Interaction")).
			Return(nil)

		interaction, err := uc.CreateInteraction(
			ctx,
			customerID,
			nil,
			string(aggregate.InteractionTypeCall),
			string(aggregate.InteractionDirectionOutbound),
			"Test call",
			"Test description",
			createdBy,
			startedAt,
		)

		require.NoError(t, err)
		assert.NotNil(t, interaction)
		assert.Equal(t, customerID, interaction.CustomerID)
		assert.Equal(t, aggregate.InteractionTypeCall, interaction.Type)
		mockRepo.AssertExpectations(t)
	})

	t.Run("with company ID", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		companyID := uuidv7.New()

		mockRepo.On("Create", ctx, mock.AnythingOfType("*aggregate.Interaction")).
			Return(nil)

		interaction, err := uc.CreateInteraction(
			ctx,
			customerID,
			&companyID,
			string(aggregate.InteractionTypeMeeting),
			string(aggregate.InteractionDirectionInbound),
			"Client meeting",
			"Discussed requirements",
			createdBy,
			startedAt,
		)

		require.NoError(t, err)
		assert.Equal(t, companyID, *interaction.CompanyID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid type", func(t *testing.T) {
		uc, _ := createTestUseCase()

		_, err := uc.CreateInteraction(
			ctx,
			customerID,
			nil,
			"invalid",
			string(aggregate.InteractionDirectionOutbound),
			"Test",
			"Test",
			createdBy,
			startedAt,
		)

		assert.Error(t, err)
		assert.Equal(t, errors.New("failed to create interaction"), err)
	})

	t.Run("repository error", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()

		mockRepo.On("Create", ctx, mock.AnythingOfType("*aggregate.Interaction")).
			Return(errors.New("database error"))

		_, err := uc.CreateInteraction(
			ctx,
			customerID,
			nil,
			string(aggregate.InteractionTypeEmail),
			string(aggregate.InteractionDirectionOutbound),
			"Test",
			"Test",
			createdBy,
			startedAt,
		)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_GetInteraction(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()

	t.Run("successful retrieval", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		expectedInteraction := &aggregate.Interaction{}
		expectedInteraction.ID = interactionID
		expectedInteraction.Type = aggregate.InteractionTypeCall

		mockRepo.On("GetByID", ctx, interactionID).
			Return(expectedInteraction, nil)

		interaction, err := uc.GetInteraction(ctx, interactionID)

		require.NoError(t, err)
		assert.Equal(t, interactionID, interaction.GetID())
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()

		mockRepo.On("GetByID", ctx, interactionID).
			Return(nil, interactionerrors.ErrInteractionNotFound)

		_, err := uc.GetInteraction(ctx, interactionID)

		assert.Equal(t, interactionerrors.ErrInteractionNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_UpdateContent(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()

	t.Run("successful update", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		existingInteraction := &aggregate.Interaction{
			Subject:     "Old subject",
			Description: "Old description",
		}
		existingInteraction.ID = interactionID

		mockRepo.On("GetByID", ctx, interactionID).
			Return(existingInteraction, nil)
		mockRepo.On("Update", ctx, existingInteraction).
			Return(nil)

		updated, err := uc.UpdateContent(ctx, interactionID, "New subject", "New description")

		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "New subject", existingInteraction.Subject)
		assert.Equal(t, "New description", existingInteraction.Description)
		mockRepo.AssertExpectations(t)
	})

	t.Run("interaction not found", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()

		mockRepo.On("GetByID", ctx, interactionID).
			Return(nil, interactionerrors.ErrInteractionNotFound)

		_, err := uc.UpdateContent(ctx, interactionID, "Test", "Test")

		assert.Equal(t, interactionerrors.ErrInteractionNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_SetOutcome(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()

	t.Run("successful outcome set", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		existingInteraction := &aggregate.Interaction{}
		existingInteraction.ID = interactionID

		mockRepo.On("GetByID", ctx, interactionID).
			Return(existingInteraction, nil)
		mockRepo.On("Update", ctx, existingInteraction).
			Return(nil)

		updated, err := uc.SetOutcome(ctx, interactionID, string(aggregate.InteractionOutcomeSuccessful))

		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, aggregate.InteractionOutcomeSuccessful, *updated.Outcome)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_EndInteraction(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()

	t.Run("successful end", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		startedAt := time.Now().Add(-30 * time.Minute)
		existingInteraction := &aggregate.Interaction{
			StartedAt: startedAt,
		}
		existingInteraction.ID = interactionID
		endedAt := time.Now()

		mockRepo.On("GetByID", ctx, interactionID).
			Return(existingInteraction, nil)
		mockRepo.On("Update", ctx, existingInteraction).
			Return(nil)

		updated, err := uc.EndInteraction(ctx, interactionID, endedAt)

		require.NoError(t, err)
		assert.NotNil(t, updated.EndedAt)
		assert.NotNil(t, updated.DurationSec)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_SetFollowUp(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()

	t.Run("enable follow-up", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		existingInteraction := &aggregate.Interaction{}
		existingInteraction.ID = interactionID
		followUpDate := time.Now().Add(24 * time.Hour)

		mockRepo.On("GetByID", ctx, interactionID).
			Return(existingInteraction, nil)
		mockRepo.On("Update", ctx, existingInteraction).
			Return(nil)

		updated, err := uc.SetFollowUp(ctx, interactionID, true, &followUpDate, "Follow up notes")

		require.NoError(t, err)
		assert.True(t, updated.FollowUpRequired)
		assert.NotNil(t, updated.FollowUpDate)
		mockRepo.AssertExpectations(t)
	})

	t.Run("disable follow-up", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		followUpDate := time.Now().Add(24 * time.Hour)
		existingInteraction := &aggregate.Interaction{
			FollowUpRequired: true,
			FollowUpDate:     &followUpDate,
		}
		existingInteraction.ID = interactionID

		mockRepo.On("GetByID", ctx, interactionID).
			Return(existingInteraction, nil)
		mockRepo.On("Update", ctx, existingInteraction).
			Return(nil)

		updated, err := uc.SetFollowUp(ctx, interactionID, false, nil, "")

		require.NoError(t, err)
		assert.False(t, updated.FollowUpRequired)
		assert.Nil(t, updated.FollowUpDate)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_AddAttendee(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()
	attendeeID := uuidv7.New()

	t.Run("successful add", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		existingInteraction := &aggregate.Interaction{
			Attendees: []uuidv7.UUID{},
		}
		existingInteraction.ID = interactionID

		mockRepo.On("GetByID", ctx, interactionID).
			Return(existingInteraction, nil)
		mockRepo.On("Update", ctx, existingInteraction).
			Return(nil)

		updated, err := uc.AddAttendee(ctx, interactionID, attendeeID)

		require.NoError(t, err)
		assert.Len(t, updated.Attendees, 1)
		assert.Contains(t, updated.Attendees, attendeeID)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_RemoveAttendee(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()
	attendeeID := uuidv7.New()

	t.Run("successful removal", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		existingInteraction := &aggregate.Interaction{
			Attendees: []uuidv7.UUID{attendeeID},
		}
		existingInteraction.ID = interactionID

		mockRepo.On("GetByID", ctx, interactionID).
			Return(existingInteraction, nil)
		mockRepo.On("Update", ctx, existingInteraction).
			Return(nil)

		updated, err := uc.RemoveAttendee(ctx, interactionID, attendeeID)

		require.NoError(t, err)
		assert.Len(t, updated.Attendees, 0)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_DeleteInteraction(t *testing.T) {
	ctx := context.Background()
	interactionID := uuidv7.New()

	t.Run("successful delete", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()

		mockRepo.On("Delete", ctx, interactionID).
			Return(nil)

		err := uc.DeleteInteraction(ctx, interactionID)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()

		mockRepo.On("Delete", ctx, interactionID).
			Return(errors.New("database error"))

		err := uc.DeleteInteraction(ctx, interactionID)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestUseCase_ListByType(t *testing.T) {
	ctx := context.Background()

	t.Run("valid type", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		inter := &aggregate.Interaction{Type: aggregate.InteractionTypeCall}
		inter.ID = uuidv7.New()
		expectedInteractions := []*aggregate.Interaction{inter}

		mockRepo.On("ListByType", ctx, string(aggregate.InteractionTypeCall), 1, 20).
			Return(expectedInteractions, int64(1), nil)

		interactions, total, err := uc.ListByType(ctx, string(aggregate.InteractionTypeCall), 1, 20)

		require.NoError(t, err)
		assert.Len(t, interactions, 1)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid type", func(t *testing.T) {
		uc, _ := createTestUseCase()

		_, _, err := uc.ListByType(ctx, "invalid", 1, 20)

		assert.Error(t, err)
		assert.ErrorIs(t, err, interactionerrors.ErrInvalidInteractionType)
	})
}

func TestUseCase_ListPendingFollowUps(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		uc, mockRepo := createTestUseCase()
		followUpDate := time.Now().Add(24 * time.Hour)
		inter := &aggregate.Interaction{
			FollowUpRequired: true,
			FollowUpDate:     &followUpDate,
		}
		inter.ID = uuidv7.New()
		expectedInteractions := []*aggregate.Interaction{inter}

		mockRepo.On("ListPendingFollowUps", ctx, 1, 20).
			Return(expectedInteractions, int64(1), nil)

		interactions, total, err := uc.ListPendingFollowUps(ctx, 1, 20)

		require.NoError(t, err)
		assert.Len(t, interactions, 1)
		assert.Equal(t, int64(1), total)
		mockRepo.AssertExpectations(t)
	})
}
