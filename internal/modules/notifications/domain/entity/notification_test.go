package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestNewNotification(t *testing.T) {
	userID := uuidv7.New()
	notif := NewNotification(
		userID,
		NotificationTypeSystem,
		NotificationChannelEmail,
		"welcome",
		"Welcome!",
		"Welcome to our platform",
	)

	assert.NotEqual(t, uuidv7.Nil, notif.ID)
	assert.Equal(t, userID, notif.UserID)
	assert.Equal(t, NotificationTypeSystem, notif.Type)
	assert.Equal(t, NotificationChannelEmail, notif.Channel)
	assert.Equal(t, NotificationStatusPending, notif.Status)
	assert.Equal(t, "welcome", notif.Template)
	assert.Equal(t, "Welcome!", notif.Subject)
	assert.Equal(t, "Welcome to our platform", notif.Content)
	assert.NotNil(t, notif.Data)
	assert.NotZero(t, notif.CreatedAt)
	assert.NotZero(t, notif.UpdatedAt)
	assert.Nil(t, notif.SentAt)
	assert.Nil(t, notif.OpenedAt)
}

func TestNotification_Validate(t *testing.T) {
	tests := []struct {
		name    string
		notif   *Notification
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid notification",
			notif: &Notification{
				ID:      uuidv7.New(),
				UserID:  uuidv7.New(),
				Type:    NotificationTypeSystem,
				Channel: NotificationChannelEmail,
				Subject: "Test",
				Content: "Test content",
			},
			wantErr: false,
		},
		{
			name: "missing user_id",
			notif: &Notification{
				UserID:  uuidv7.Nil,
				Type:    NotificationTypeSystem,
				Channel: NotificationChannelEmail,
				Content: "Test",
			},
			wantErr: true,
			errMsg:  "user_id is required",
		},
		{
			name: "missing type",
			notif: &Notification{
				UserID:  uuidv7.New(),
				Channel: NotificationChannelEmail,
				Content: "Test",
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "missing channel",
			notif: &Notification{
				UserID:  uuidv7.New(),
				Type:    NotificationTypeSystem,
				Content: "Test",
			},
			wantErr: true,
			errMsg:  "channel is required",
		},
		{
			name: "email without subject",
			notif: &Notification{
				UserID:  uuidv7.New(),
				Type:    NotificationTypeSystem,
				Channel: NotificationChannelEmail,
				Content: "Test",
			},
			wantErr: true,
			errMsg:  "subject is required for email notifications",
		},
		{
			name: "missing content",
			notif: &Notification{
				UserID:  uuidv7.New(),
				Type:    NotificationTypeSystem,
				Channel: NotificationChannelSMS,
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "push notification without subject is valid",
			notif: &Notification{
				UserID:  uuidv7.New(),
				Type:    NotificationTypeSystem,
				Channel: NotificationChannelPush,
				Content: "Push message",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.notif.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotification_MarkAsSent(t *testing.T) {
	notif := NewNotification(
		uuidv7.New(),
		NotificationTypeSystem,
		NotificationChannelEmail,
		"test",
		"Test",
		"Content",
	)

	beforeTime := time.Now()
	notif.MarkAsSent()
	afterTime := time.Now()

	assert.Equal(t, NotificationStatusSent, notif.Status)
	assert.NotNil(t, notif.SentAt)
	assert.True(t, notif.SentAt.After(beforeTime) || notif.SentAt.Equal(beforeTime))
	assert.True(t, notif.SentAt.Before(afterTime) || notif.SentAt.Equal(afterTime))
	assert.True(t, notif.UpdatedAt.After(beforeTime))
}

func TestNotification_MarkAsDelivered(t *testing.T) {
	notif := NewNotification(
		uuidv7.New(),
		NotificationTypeSystem,
		NotificationChannelEmail,
		"test",
		"Test",
		"Content",
	)

	originalUpdatedAt := notif.UpdatedAt
	time.Sleep(1 * time.Millisecond)
	notif.MarkAsDelivered()

	assert.Equal(t, NotificationStatusDelivered, notif.Status)
	assert.True(t, notif.UpdatedAt.After(originalUpdatedAt))
}

func TestNotification_MarkAsFailed(t *testing.T) {
	notif := NewNotification(
		uuidv7.New(),
		NotificationTypeSystem,
		NotificationChannelEmail,
		"test",
		"Test",
		"Content",
	)

	errorMsg := "SMTP connection failed"
	beforeTime := time.Now()
	notif.MarkAsFailed(errorMsg)
	afterTime := time.Now()

	assert.Equal(t, NotificationStatusFailed, notif.Status)
	assert.NotNil(t, notif.FailedAt)
	assert.NotNil(t, notif.ErrorMsg)
	assert.Equal(t, errorMsg, *notif.ErrorMsg)
	assert.True(t, notif.FailedAt.After(beforeTime) || notif.FailedAt.Equal(beforeTime))
	assert.True(t, notif.FailedAt.Before(afterTime) || notif.FailedAt.Equal(afterTime))
	assert.True(t, notif.UpdatedAt.After(beforeTime))
}

func TestNotification_MarkAsOpened(t *testing.T) {
	notif := NewNotification(
		uuidv7.New(),
		NotificationTypeSystem,
		NotificationChannelEmail,
		"test",
		"Test",
		"Content",
	)

	beforeTime := time.Now()
	notif.MarkAsOpened()
	afterTime := time.Now()

	assert.Equal(t, NotificationStatusOpened, notif.Status)
	assert.NotNil(t, notif.OpenedAt)
	assert.True(t, notif.OpenedAt.After(beforeTime) || notif.OpenedAt.Equal(beforeTime))
	assert.True(t, notif.OpenedAt.Before(afterTime) || notif.OpenedAt.Equal(afterTime))
	assert.True(t, notif.UpdatedAt.After(beforeTime))
}

func TestNotification_MarkAsClicked(t *testing.T) {
	notif := NewNotification(
		uuidv7.New(),
		NotificationTypeSystem,
		NotificationChannelEmail,
		"test",
		"Test",
		"Content",
	)

	beforeTime := time.Now()
	notif.MarkAsClicked()
	afterTime := time.Now()

	assert.Equal(t, NotificationStatusClicked, notif.Status)
	assert.NotNil(t, notif.ClickedAt)
	assert.True(t, notif.ClickedAt.After(beforeTime) || notif.ClickedAt.Equal(beforeTime))
	assert.True(t, notif.ClickedAt.Before(afterTime) || notif.ClickedAt.Equal(afterTime))
	assert.True(t, notif.UpdatedAt.After(beforeTime))
}

func TestNotification_IsDelivered(t *testing.T) {
	notif := NewNotification(
		uuidv7.New(),
		NotificationTypeSystem,
		NotificationChannelEmail,
		"test",
		"Test",
		"Content",
	)

	// Initially not delivered
	assert.False(t, notif.IsDelivered())

	// After sending - not yet delivered
	notif.MarkAsSent()
	assert.False(t, notif.IsDelivered())

	// After marking as delivered
	notif.MarkAsDelivered()
	assert.True(t, notif.IsDelivered())
}

func TestNotification_IsFailed(t *testing.T) {
	notif := NewNotification(
		uuidv7.New(),
		NotificationTypeSystem,
		NotificationChannelEmail,
		"test",
		"Test",
		"Content",
	)

	// Initially not failed
	assert.False(t, notif.IsFailed())

	// After marking as failed
	notif.MarkAsFailed("error")
	assert.True(t, notif.IsFailed())
}

