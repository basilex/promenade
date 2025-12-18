package helpers

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// UserFixture creates a test user with unique email
func UserFixture(overrides ...func(*entity.User)) *entity.User {
	id := uuidv7.New()
	// Use full UUID for email to ensure uniqueness even with rapid creation
	user := &entity.User{
		ID:        id,
		Email:     fmt.Sprintf("test-%s@example.com", id.String()),
		Name:      "Test User",
		Password:  "$2a$10$test.hashed.password.dummy", // Pre-hashed dummy
		Status:    entity.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(user)
	}

	return user
}

// UnverifiedUserFixture creates an unverified test user
func UnverifiedUserFixture() *entity.User {
	return UserFixture(func(u *entity.User) {
		u.Email = "unverified@example.com"
		u.Status = entity.UserStatusUnverified
		u.EmailVerifiedAt = nil
	})
}

// SuspendedUserFixture creates a suspended test user
func SuspendedUserFixture() *entity.User {
	reason := "Test suspension"
	until := time.Now().Add(24 * time.Hour)
	return UserFixture(func(u *entity.User) {
		u.Email = "suspended@example.com"
		u.Status = entity.UserStatusSuspended
		u.SuspendedReason = &reason
		u.SuspendedUntil = &until
	})
}

// BannedUserFixture creates a banned test user
func BannedUserFixture() *entity.User {
	reason := "Test ban"
	return UserFixture(func(u *entity.User) {
		u.Email = "banned@example.com"
		u.Status = entity.UserStatusBanned
		u.SuspendedReason = &reason
	})
}

// SessionFixture creates a test session with unique refresh token
func SessionFixture(userID uuidv7.UUID, overrides ...func(*entity.Session)) *entity.Session {
	userAgent := "Test User Agent"
	ipAddress := "127.0.0.1"
	id := uuidv7.New()

	session := &entity.Session{
		ID:           id,
		UserID:       userID,
		RefreshToken: fmt.Sprintf("hashed_test_token_%s", id.String()),
		UserAgent:    &userAgent,
		IPAddress:    &ipAddress,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(session)
	}

	return session
}

// ExpiredSessionFixture creates an expired session
func ExpiredSessionFixture(userID uuidv7.UUID) *entity.Session {
	return SessionFixture(userID, func(s *entity.Session) {
		s.ExpiresAt = time.Now().Add(-1 * time.Hour)
	})
}

// UserContactFixture creates a test user contact
func UserContactFixture(userID uuidv7.UUID, contactType entity.ContactType, overrides ...func(*entity.UserContact)) *entity.UserContact {
	id := uuidv7.New()
	contactValue := fmt.Sprintf("test-%s@example.com", id.String())
	if contactType == entity.ContactTypePhone {
		contactValue = "+1234567890"
	}

	contact := &entity.UserContact{
		ID:           id,
		UserID:       userID,
		ContactType:  contactType,
		ContactValue: contactValue,
		IsVerified:   false,
		IsPrimary:    false,
		IsActive:     true,
		IsPublic:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(contact)
	}

	return contact
}

// UserProfileFixture creates a test user profile
func UserProfileFixture(userID uuidv7.UUID, nickname string, overrides ...func(*entity.UserProfile)) *entity.UserProfile {
	displayName := "Test User"
	bio := "Test bio"

	profile := &entity.UserProfile{
		ID:                uuidv7.New(),
		UserID:            userID,
		DisplayName:       &displayName,
		Nickname:          &nickname,
		Bio:               &bio,
		Timezone:          "UTC",
		Locale:            "en",
		IsPublic:          true,
		IsVerified:        false,
		ShowEmail:         false,
		ShowLocation:      false,
		ShowBirthday:      false,
		ProfileViewsCount: 0,
		FollowersCount:    0,
		FollowingCount:    0,
		IsBanned:          false,
		SocialLinks:       entity.SocialLinks{},
		Preferences:       entity.Preferences{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(profile)
	}

	return profile
}

// PermissionFixture creates a test permission
func PermissionFixture(resource, action string, overrides ...func(*entity.Permission)) *entity.Permission {
	desc := fmt.Sprintf("Permission to %s %s", action, resource)
	permission := &entity.Permission{
		ID:          uuidv7.New(),
		Resource:    resource,
		Action:      action,
		Description: &desc,
		CreatedAt:   time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(permission)
	}

	return permission
}

// RoleFixture creates a test role
func RoleFixture(name string, overrides ...func(*entity.Role)) *entity.Role {
	desc := fmt.Sprintf("Test role: %s", name)
	role := &entity.Role{
		ID:          uuidv7.New(),
		Name:        name,
		DisplayName: name,
		Description: &desc,
		IsSystem:    false,
		Permissions: []*entity.Permission{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Apply overrides
	for _, override := range overrides {
		override(role)
	}

	return role
}

// UserRoleFixture creates a test user role assignment
func UserRoleFixture(userID, roleID uuidv7.UUID, overrides ...func(*entity.UserRole)) *entity.UserRole {
	userRole := &entity.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedAt: time.Now(),
		AssignedBy: nil,
		ExpiresAt:  nil,
	}

	// Apply overrides
	for _, override := range overrides {
		override(userRole)
	}

	return userRole
}
