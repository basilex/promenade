package event

import (
	"time"

	"github.com/basilex/promenade/pkg/bus"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserRegisteredEvent is published when a new user registers.
type UserRegisteredEvent struct {
	*bus.BaseEvent
	UserID uuidv7.UUID `json:"user_id"`
	Email  string      `json:"email"`
	Name   string      `json:"name"`
}

// NewUserRegisteredEvent creates a user registration event.
func NewUserRegisteredEvent(userID uuidv7.UUID, email, name string) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		BaseEvent: bus.NewBaseEvent(bus.TopicUserRegistered, userID),
		UserID:    userID,
		Email:     email,
		Name:      name,
	}
}

// UserActivatedEvent is published when a user account is activated.
type UserActivatedEvent struct {
	*bus.BaseEvent
	UserID uuidv7.UUID `json:"user_id"`
	Email  string      `json:"email"`
}

// NewUserActivatedEvent creates a user activation event.
func NewUserActivatedEvent(userID uuidv7.UUID, email string) *UserActivatedEvent {
	return &UserActivatedEvent{
		BaseEvent: bus.NewBaseEvent(bus.TopicUserActivated, userID),
		UserID:    userID,
		Email:     email,
	}
}

// UserSuspendedEvent is published when a user is suspended.
type UserSuspendedEvent struct {
	*bus.BaseEvent
	UserID    uuidv7.UUID `json:"user_id"`
	Email     string      `json:"email"`
	Reason    string      `json:"reason"`
	ExpiresAt *time.Time  `json:"expires_at,omitempty"`
}

// NewUserSuspendedEvent creates a user suspension event.
func NewUserSuspendedEvent(userID uuidv7.UUID, email, reason string, expiresAt *time.Time) *UserSuspendedEvent {
	return &UserSuspendedEvent{
		BaseEvent: bus.NewBaseEvent(bus.TopicUserSuspended, userID),
		UserID:    userID,
		Email:     email,
		Reason:    reason,
		ExpiresAt: expiresAt,
	}
}

// UserBannedEvent is published when a user is permanently banned.
type UserBannedEvent struct {
	*bus.BaseEvent
	UserID uuidv7.UUID `json:"user_id"`
	Email  string      `json:"email"`
	Reason string      `json:"reason"`
}

// NewUserBannedEvent creates a user ban event.
func NewUserBannedEvent(userID uuidv7.UUID, email, reason string) *UserBannedEvent {
	return &UserBannedEvent{
		BaseEvent: bus.NewBaseEvent(bus.TopicUserBanned, userID),
		UserID:    userID,
		Email:     email,
		Reason:    reason,
	}
}

// UserEmailVerifiedEvent is published when a user verifies their email.
type UserEmailVerifiedEvent struct {
	*bus.BaseEvent
	UserID uuidv7.UUID `json:"user_id"`
	Email  string      `json:"email"`
}

// NewUserEmailVerifiedEvent creates an email verification event.
func NewUserEmailVerifiedEvent(userID uuidv7.UUID, email string) *UserEmailVerifiedEvent {
	return &UserEmailVerifiedEvent{
		BaseEvent: bus.NewBaseEvent(bus.TopicUserEmailVerified, userID),
		UserID:    userID,
		Email:     email,
	}
}

// UserPasswordChangedEvent is published when a user changes password.
type UserPasswordChangedEvent struct {
	*bus.BaseEvent
	UserID uuidv7.UUID `json:"user_id"`
	Email  string      `json:"email"`
}

// NewUserPasswordChangedEvent creates a password change event.
func NewUserPasswordChangedEvent(userID uuidv7.UUID, email string) *UserPasswordChangedEvent {
	return &UserPasswordChangedEvent{
		BaseEvent: bus.NewBaseEvent(bus.TopicUserPasswordChanged, userID),
		UserID:    userID,
		Email:     email,
	}
}
