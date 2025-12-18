// Package ref provides helper functions for working with references to values.
// It simplifies the creation and dereferencing of reference values,
// making it easier to work with nullable fields in database entities.
package ref

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// String returns a pointer to the given string value.
// This is useful when you need to pass a string literal as a pointer.
//
// Example:
//
//	user.Nickname = ref.String("john_doe")
func String(s string) *string {
	return &s
}

// StringValue returns the dereferenced string value or an empty string if nil.
// This provides safe access to nullable string fields.
//
// Example:
//
//	nickname := ref.StringValue(user.Nickname) // "" if nil
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringOr returns the dereferenced string value or the provided default if nil.
//
// Example:
//
//	bio := ref.StringOr(user.Bio, "No bio provided")
func StringOr(s *string, defaultValue string) string {
	if s == nil {
		return defaultValue
	}
	return *s
}

// Time returns a pointer to the given time.Time value.
//
// Example:
//
//	user.LastLoginAt = ref.Time(time.Now())
func Time(t time.Time) *time.Time {
	return &t
}

// TimeValue returns the dereferenced time.Time value or zero time if nil.
//
// Example:
//
//	lastLogin := ref.TimeValue(user.LastLoginAt)
func TimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

// TimeOr returns the dereferenced time.Time value or the provided default if nil.
//
// Example:
//
//	lastSeen := ref.TimeOr(user.LastSeenAt, time.Now())
func TimeOr(t *time.Time, defaultValue time.Time) time.Time {
	if t == nil {
		return defaultValue
	}
	return *t
}

// UUID returns a pointer to the given UUID value.
//
// Example:
//
//	profile.CountryID = ref.UUID(countryUUID)
func UUID(u uuidv7.UUID) *uuidv7.UUID {
	return &u
}

// UUIDValue returns the dereferenced UUID value or a nil UUID if nil.
//
// Example:
//
//	countryID := ref.UUIDValue(profile.CountryID)
func UUIDValue(u *uuidv7.UUID) uuidv7.UUID {
	if u == nil {
		return uuidv7.UUID{}
	}
	return *u
}

// Int returns a pointer to the given int value.
//
// Example:
//
//	config.MaxRetries = ref.Int(3)
func Int(i int) *int {
	return &i
}

// IntValue returns the dereferenced int value or 0 if nil.
//
// Example:
//
//	maxRetries := ref.IntValue(config.MaxRetries)
func IntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// IntOr returns the dereferenced int value or the provided default if nil.
//
// Example:
//
//	maxRetries := ref.IntOr(config.MaxRetries, 3)
func IntOr(i *int, defaultValue int) int {
	if i == nil {
		return defaultValue
	}
	return *i
}

// Bool returns a pointer to the given bool value.
//
// Example:
//
//	feature.Enabled = ref.Bool(true)
func Bool(b bool) *bool {
	return &b
}

// BoolValue returns the dereferenced bool value or false if nil.
//
// Example:
//
//	enabled := ref.BoolValue(feature.Enabled)
func BoolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// BoolOr returns the dereferenced bool value or the provided default if nil.
//
// Example:
//
//	enabled := ref.BoolOr(feature.Enabled, true)
func BoolOr(b *bool, defaultValue bool) bool {
	if b == nil {
		return defaultValue
	}
	return *b
}

// IsNil checks if a pointer is nil. This is a convenience function
// that can be used with any pointer type.
//
// Example:
//
//	if !ref.IsNil(user.Nickname) {
//	    // nickname is set
//	}
func IsNil[T any](p *T) bool {
	return p == nil
}

// IsSet checks if a pointer is not nil and not empty/zero.
// For strings, it checks if the dereferenced value is not empty.
//
// Example:
//
//	if ref.IsSet(user.Nickname) {
//	    // nickname is set and not empty
//	}
func IsSet(s *string) bool {
	return s != nil && *s != ""
}
