package interaction

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewInteraction(t *testing.T) {
	customerID := uuidv7.New()
	createdBy := uuidv7.New()
	startedAt := time.Now()

	t.Run("valid call interaction", func(t *testing.T) {
		inter, err := NewInteraction(
			customerID,
			nil,
			InteractionTypeCall,
			InteractionDirectionOutbound,
			"Follow-up call",
			"Discussed project requirements",
			createdBy,
			startedAt,
		)

		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.UUID{}, inter.ID)
		assert.Equal(t, customerID, inter.CustomerID)
		assert.Nil(t, inter.CompanyID)
		assert.Equal(t, InteractionTypeCall, inter.Type)
		assert.Equal(t, InteractionDirectionOutbound, inter.Direction)
		assert.Equal(t, "Follow-up call", inter.Subject)
		assert.Equal(t, "Discussed project requirements", inter.Description)
		assert.Equal(t, createdBy, inter.CreatedBy)
		assert.Equal(t, startedAt, inter.StartedAt)
		assert.Nil(t, inter.Outcome)
		assert.Nil(t, inter.EndedAt)
		assert.Nil(t, inter.DurationSec)
		assert.False(t, inter.FollowUpRequired)
		assert.Empty(t, inter.Attendees)
	})

	t.Run("with company ID", func(t *testing.T) {
		companyID := uuidv7.New()
		inter, err := NewInteraction(
			customerID,
			&companyID,
			InteractionTypeMeeting,
			InteractionDirectionInbound,
			"Client meeting",
			"Discussed contract terms",
			createdBy,
			startedAt,
		)

		require.NoError(t, err)
		assert.Equal(t, companyID, *inter.CompanyID)
	})

	t.Run("invalid type", func(t *testing.T) {
		_, err := NewInteraction(
			customerID,
			nil,
			"invalid",
			InteractionDirectionOutbound,
			"Test",
			"Test",
			createdBy,
			startedAt,
		)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidInteractionType))
	})

	t.Run("invalid direction", func(t *testing.T) {
		_, err := NewInteraction(
			customerID,
			nil,
			InteractionTypeCall,
			"invalid",
			"Test",
			"Test",
			createdBy,
			startedAt,
		)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidDirection))
	})

	t.Run("empty subject", func(t *testing.T) {
		_, err := NewInteraction(
			customerID,
			nil,
			InteractionTypeCall,
			InteractionDirectionOutbound,
			"",
			"Test description",
			createdBy,
			startedAt,
		)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrSubjectEmpty))
	})

	t.Run("empty description", func(t *testing.T) {
		_, err := NewInteraction(
			customerID,
			nil,
			InteractionTypeCall,
			InteractionDirectionOutbound,
			"Test subject",
			"",
			createdBy,
			startedAt,
		)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrDescriptionEmpty))
	})
}

func TestInteraction_UpdateContent(t *testing.T) {
	inter := createTestInteraction(t)

	t.Run("successful update", func(t *testing.T) {
		err := inter.UpdateContent("Updated subject", "Updated description")

		require.NoError(t, err)
		assert.Equal(t, "Updated subject", inter.Subject)
		assert.Equal(t, "Updated description", inter.Description)
	})

	t.Run("empty subject", func(t *testing.T) {
		err := inter.UpdateContent("", "Test")

		assert.Error(t, err)
	})

	t.Run("empty description", func(t *testing.T) {
		err := inter.UpdateContent("Test", "")

		assert.Error(t, err)
	})
}

func TestInteraction_SetCompany(t *testing.T) {
	inter := createTestInteraction(t)
	companyID := uuidv7.New()

	inter.SetCompany(&companyID)
	assert.Equal(t, companyID, *inter.CompanyID)

	inter.SetCompany(nil)
	assert.Nil(t, inter.CompanyID)
}

func TestInteraction_SetOutcome(t *testing.T) {
	inter := createTestInteraction(t)

	t.Run("valid outcome", func(t *testing.T) {
		err := inter.SetOutcome(InteractionOutcomeSuccessful)

		require.NoError(t, err)
		assert.Equal(t, InteractionOutcomeSuccessful, *inter.Outcome)
	})

	t.Run("invalid outcome", func(t *testing.T) {
		err := inter.SetOutcome("invalid")

		assert.Error(t, err)
	})
}

func TestInteraction_EndInteraction(t *testing.T) {
	inter := createTestInteraction(t)

	t.Run("successful end", func(t *testing.T) {
		endedAt := inter.StartedAt.Add(30 * time.Minute)
		err := inter.EndInteraction(endedAt)

		require.NoError(t, err)
		assert.Equal(t, endedAt, *inter.EndedAt)
		assert.Equal(t, 1800, *inter.DurationSec) // 30 minutes = 1800 seconds
	})

	t.Run("end before start", func(t *testing.T) {
		inter2 := createTestInteraction(t)
		endedAt := inter2.StartedAt.Add(-10 * time.Minute)
		err := inter2.EndInteraction(endedAt)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrEndedAtBeforeStartedAt))
	})

	t.Run("already ended", func(t *testing.T) {
		inter2 := createTestInteraction(t)
		endedAt := inter2.StartedAt.Add(10 * time.Minute)
		_ = inter2.EndInteraction(endedAt)

		err := inter2.EndInteraction(endedAt.Add(5 * time.Minute))

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInteractionAlreadyEnded)
	})
}

func TestInteraction_SetFollowUp(t *testing.T) {
	inter := createTestInteraction(t)

	t.Run("require follow-up with date", func(t *testing.T) {
		followUpDate := time.Now().Add(24 * time.Hour)
		err := inter.SetFollowUp(true, &followUpDate, "Call again tomorrow")

		require.NoError(t, err)
		assert.True(t, inter.FollowUpRequired)
		assert.Equal(t, followUpDate, *inter.FollowUpDate)
		assert.Equal(t, "Call again tomorrow", inter.FollowUpNotes)
	})

	t.Run("require follow-up without date", func(t *testing.T) {
		inter2 := createTestInteraction(t)
		err := inter2.SetFollowUp(true, nil, "Follow up soon")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrFollowUpDateRequired))
	})

	t.Run("disable follow-up", func(t *testing.T) {
		followUpDate := time.Now().Add(24 * time.Hour)
		_ = inter.SetFollowUp(true, &followUpDate, "Test")

		err := inter.SetFollowUp(false, nil, "")

		require.NoError(t, err)
		assert.False(t, inter.FollowUpRequired)
		assert.Nil(t, inter.FollowUpDate)
		assert.Empty(t, inter.FollowUpNotes)
	})
}

func TestInteraction_AddRemoveAttendee(t *testing.T) {
	inter := createTestInteraction(t)

	t.Run("add attendee", func(t *testing.T) {
		attendeeID := uuidv7.New()
		inter.AddAttendee(attendeeID)

		assert.Len(t, inter.Attendees, 1)
		assert.Contains(t, inter.Attendees, attendeeID)
	})

	t.Run("add duplicate attendee", func(t *testing.T) {
		inter2 := createTestInteraction(t)
		attendeeID := uuidv7.New()
		inter2.AddAttendee(attendeeID)
		inter2.AddAttendee(attendeeID) // Duplicate

		assert.Len(t, inter2.Attendees, 2) // Allows duplicates
	})

	t.Run("remove attendee", func(t *testing.T) {
		attendeeID := uuidv7.New()
		inter.AddAttendee(attendeeID)
		initialLen := len(inter.Attendees)

		inter.RemoveAttendee(attendeeID)

		assert.Len(t, inter.Attendees, initialLen-1)
		assert.NotContains(t, inter.Attendees, attendeeID)
	})

	t.Run("remove non-existent attendee", func(t *testing.T) {
		attendeeID := uuidv7.New()
		initialLen := len(inter.Attendees)

		inter.RemoveAttendee(attendeeID)

		assert.Len(t, inter.Attendees, initialLen) // No change
	})
}

func TestInteraction_Delete(t *testing.T) {
	inter := createTestInteraction(t)
	assert.Nil(t, inter.DeletedAt)

	inter.Delete()

	assert.NotNil(t, inter.DeletedAt)
}

func TestInteraction_Validate(t *testing.T) {
	t.Run("valid interaction", func(t *testing.T) {
		inter := createTestInteraction(t)
		err := inter.Validate()

		assert.NoError(t, err)
	})

	t.Run("invalid type", func(t *testing.T) {
		inter := createTestInteraction(t)
		inter.Type = "invalid"
		err := inter.Validate()

		assert.Error(t, err)
	})

	t.Run("invalid direction", func(t *testing.T) {
		inter := createTestInteraction(t)
		inter.Direction = "invalid"
		err := inter.Validate()

		assert.Error(t, err)
	})

	t.Run("invalid outcome", func(t *testing.T) {
		inter := createTestInteraction(t)
		outcome := InteractionOutcome("invalid")
		inter.Outcome = &outcome
		err := inter.Validate()

		assert.Error(t, err)
	})

	t.Run("empty subject", func(t *testing.T) {
		inter := createTestInteraction(t)
		inter.Subject = ""
		err := inter.Validate()

		assert.Error(t, err)
	})

	t.Run("empty description", func(t *testing.T) {
		inter := createTestInteraction(t)
		inter.Description = ""
		err := inter.Validate()

		assert.Error(t, err)
	})

	t.Run("negative duration", func(t *testing.T) {
		inter := createTestInteraction(t)
		duration := -10
		inter.DurationSec = &duration
		err := inter.Validate()

		assert.Error(t, err)
	})

	t.Run("ended before started", func(t *testing.T) {
		inter := createTestInteraction(t)
		endedAt := inter.StartedAt.Add(-10 * time.Minute)
		inter.EndedAt = &endedAt
		err := inter.Validate()

		assert.Error(t, err)
	})

	t.Run("follow-up required without date", func(t *testing.T) {
		inter := createTestInteraction(t)
		inter.FollowUpRequired = true
		inter.FollowUpDate = nil
		err := inter.Validate()

		assert.Error(t, err)
	})
}

// Helper function to create a test interaction
func createTestInteraction(t *testing.T) *Interaction {
	t.Helper()

	customerID := uuidv7.New()
	createdBy := uuidv7.New()
	startedAt := time.Now()

	inter, err := NewInteraction(
		customerID,
		nil,
		InteractionTypeCall,
		InteractionDirectionOutbound,
		"Test call",
		"Test description",
		createdBy,
		startedAt,
	)

	require.NoError(t, err)
	return inter
}
