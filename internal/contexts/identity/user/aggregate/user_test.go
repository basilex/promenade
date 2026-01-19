package aggregate

import (
	"errors"
	"testing"
	"time"

	usererrors "github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	t.Run("valid user", func(t *testing.T) {
		email := "test@example.com"
		password := "Password123"

		user, err := NewUser(email, password)

		require.NoError(t, err)
		assert.NotEqual(t, uuidv7.Nil, user.ID)
		assert.Equal(t, email, user.Email.Value())
		assert.NotEmpty(t, user.PasswordHash)
		assert.NotEqual(t, password, user.PasswordHash) // Password should be hashed
		assert.Equal(t, UserStatusActive, user.Status)
		assert.False(t, user.EmailVerified)
		assert.Nil(t, user.EmailVerifiedAt)
		assert.Nil(t, user.LastLoginAt)
		assert.Equal(t, 0, user.FailedLoginCount)
		assert.Nil(t, user.LockedUntil)
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := NewUser("", "Password123")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrInvalidEmailFormat))
	})

	t.Run("invalid password", func(t *testing.T) {
		_, err := NewUser("test@example.com", "weak")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrPasswordTooShort))
	})

	t.Run("password without digit", func(t *testing.T) {
		_, err := NewUser("test@example.com", "Password")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrPasswordRequiresDigit))
	})

	t.Run("password without letter", func(t *testing.T) {
		_, err := NewUser("test@example.com", "12345678")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrPasswordRequiresLetter))
	})

	t.Run("password too long", func(t *testing.T) {
		longPassword := make([]byte, 73)
		for i := range longPassword {
			if i%2 == 0 {
				longPassword[i] = 'A'
			} else {
				longPassword[i] = '1'
			}
		}
		_, err := NewUser("test@example.com", string(longPassword))
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrPasswordTooLong))
	})
}

func TestUser_CheckPassword(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("correct password", func(t *testing.T) {
		err := user.CheckPassword("Password123")
		assert.NoError(t, err)
	})

	t.Run("incorrect password", func(t *testing.T) {
		err := user.CheckPassword("WrongPassword")
		assert.Error(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		err := user.CheckPassword("")
		assert.Error(t, err)
	})
}

func TestUser_ChangePassword(t *testing.T) {
	user, _ := NewUser("test@example.com", "OldPassword123")
	oldHash := user.PasswordHash

	t.Run("valid password change", func(t *testing.T) {
		err := user.ChangePassword("NewPassword456")
		require.NoError(t, err)
		assert.NotEqual(t, oldHash, user.PasswordHash)
		assert.NoError(t, user.CheckPassword("NewPassword456"))
		assert.Error(t, user.CheckPassword("OldPassword123"))
	})

	t.Run("invalid new password", func(t *testing.T) {
		err := user.ChangePassword("weak")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrPasswordTooShort))
	})
}

func TestUser_VerifyEmail(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("verify email", func(t *testing.T) {
		assert.False(t, user.EmailVerified)
		assert.Nil(t, user.EmailVerifiedAt)

		user.VerifyEmail()

		assert.True(t, user.EmailVerified)
		assert.NotNil(t, user.EmailVerifiedAt)
		assert.WithinDuration(t, time.Now(), *user.EmailVerifiedAt, 1*time.Second)
	})

	t.Run("verify already verified email", func(t *testing.T) {
		user, _ := NewUser("test@example.com", "Password123")
		user.VerifyEmail()
		firstVerifiedAt := *user.EmailVerifiedAt

		time.Sleep(10 * time.Millisecond)
		user.VerifyEmail()

		// VerifyEmail should update timestamp each time
		assert.True(t, user.EmailVerified)
		assert.NotNil(t, user.EmailVerifiedAt)
		// Timestamp may change on subsequent calls
		assert.True(t, user.EmailVerifiedAt.Equal(firstVerifiedAt) || user.EmailVerifiedAt.After(firstVerifiedAt))
	})
}

func TestUser_RecordLogin(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("record first login", func(t *testing.T) {
		assert.Nil(t, user.LastLoginAt)
		assert.Equal(t, 0, user.FailedLoginCount)

		user.RecordLogin()

		assert.NotNil(t, user.LastLoginAt)
		assert.WithinDuration(t, time.Now(), *user.LastLoginAt, 1*time.Second)
		assert.Equal(t, 0, user.FailedLoginCount)
	})

	t.Run("record login resets failed count", func(t *testing.T) {
		user, _ := NewUser("test@example.com", "Password123")
		user.FailedLoginCount = 3

		user.RecordLogin()

		assert.NotNil(t, user.LastLoginAt)
		assert.Equal(t, 0, user.FailedLoginCount)
	})
}

func TestUser_RecordFailedLogin(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("record failed login", func(t *testing.T) {
		assert.Equal(t, 0, user.FailedLoginCount)
		assert.False(t, user.IsLocked())

		user.RecordFailedLogin()
		assert.Equal(t, 1, user.FailedLoginCount)
		assert.False(t, user.IsLocked())
	})

	t.Run("lock after max attempts", func(t *testing.T) {
		user, _ := NewUser("test@example.com", "Password123")

		for i := 0; i < maxLoginAttempts; i++ {
			user.RecordFailedLogin()
		}

		assert.Equal(t, maxLoginAttempts, user.FailedLoginCount)
		assert.True(t, user.IsLocked())
		assert.NotNil(t, user.LockedUntil)
		assert.True(t, user.LockedUntil.After(time.Now()))
	})
}

func TestUser_IsLocked(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("not locked by default", func(t *testing.T) {
		assert.False(t, user.IsLocked())
	})

	t.Run("locked with future date", func(t *testing.T) {
		future := time.Now().Add(10 * time.Minute)
		user.LockedUntil = &future

		assert.True(t, user.IsLocked())
	})

	t.Run("not locked with past date", func(t *testing.T) {
		past := time.Now().Add(-10 * time.Minute)
		user.LockedUntil = &past

		assert.False(t, user.IsLocked())
	})
}

func TestUser_UnlockAccount(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("unlock locked account", func(t *testing.T) {
		// Lock account
		for i := 0; i < maxLoginAttempts; i++ {
			user.RecordFailedLogin()
		}
		assert.True(t, user.IsLocked())

		// Unlock
		user.UnlockAccount()

		assert.False(t, user.IsLocked())
		assert.Nil(t, user.LockedUntil)
		assert.Equal(t, 0, user.FailedLoginCount)
	})
}

func TestUser_StatusTransitions(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("activate user", func(t *testing.T) {
		user.Status = UserStatusInactive
		user.Activate()
		assert.Equal(t, UserStatusActive, user.Status)
	})

	t.Run("deactivate user", func(t *testing.T) {
		user.Status = UserStatusActive
		user.Deactivate()
		assert.Equal(t, UserStatusInactive, user.Status)
	})

	t.Run("suspend user", func(t *testing.T) {
		user.Status = UserStatusActive
		user.Suspend()
		assert.Equal(t, UserStatusSuspended, user.Status)
	})

	t.Run("ban user", func(t *testing.T) {
		user.Status = UserStatusActive
		user.Ban()
		assert.Equal(t, UserStatusBanned, user.Status)
	})
}

func TestUser_IsActive(t *testing.T) {
	user, _ := NewUser("test@example.com", "Password123")

	t.Run("active status", func(t *testing.T) {
		user.Status = UserStatusActive
		assert.True(t, user.IsActive())
	})

	t.Run("inactive status", func(t *testing.T) {
		user.Status = UserStatusInactive
		assert.False(t, user.IsActive())
	})

	t.Run("suspended status", func(t *testing.T) {
		user.Status = UserStatusSuspended
		assert.False(t, user.IsActive())
	})

	t.Run("banned status", func(t *testing.T) {
		user.Status = UserStatusBanned
		assert.False(t, user.IsActive())
	})
}

func TestUser_Validate(t *testing.T) {
	t.Run("valid user", func(t *testing.T) {
		user, _ := NewUser("test@example.com", "Password123")
		err := user.Validate()
		assert.NoError(t, err)
	})

	t.Run("missing email", func(t *testing.T) {
		user := &User{
			Email:        valueobject.Email{}, // Empty value object
			PasswordHash: "hash",
			Status:       UserStatusActive,
		}
		user.ID = uuidv7.New()
		err := user.Validate()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrEmailRequired))
	})

	t.Run("missing password hash", func(t *testing.T) {
		email, _ := valueobject.NewEmail("test@example.com")
		user := &User{
			Email:        email,
			PasswordHash: "",
			Status:       UserStatusActive,
		}
		user.ID = uuidv7.New()
		err := user.Validate()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrPasswordHashRequired))
	})

	t.Run("invalid status", func(t *testing.T) {
		email, _ := valueobject.NewEmail("test@example.com")
		user := &User{
			Email:        email,
			PasswordHash: "hash",
			Status:       "invalid",
		}
		user.ID = uuidv7.New()
		err := user.Validate()
		assert.Error(t, err)
		assert.True(t, errors.Is(err, usererrors.ErrInvalidUserStatus))
	})
}
