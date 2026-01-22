package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	rolerepository "github.com/basilex/promenade/internal/contexts/identity/role/repository"
	usererrors "github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/basilex/promenade/internal/contexts/identity/user/aggregate"
	"github.com/basilex/promenade/internal/contexts/identity/user/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUserUseCase defines the interface for User use cases
type IUserUseCase interface {
	// Register creates a new user account
	Register(ctx context.Context, email, name, password string) (*aggregate.User, error)

	// Authenticate authenticates a user with email and password
	Authenticate(ctx context.Context, email, password string) (*aggregate.User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, userID uuidv7.UUID) (*aggregate.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*aggregate.User, error)

	// VerifyEmail marks user's email as verified
	VerifyEmail(ctx context.Context, userID uuidv7.UUID) error

	// ChangePassword changes user's password
	ChangePassword(ctx context.Context, userID uuidv7.UUID, oldPassword, newPassword string) error

	// SuspendUser suspends a user account
	SuspendUser(ctx context.Context, userID uuidv7.UUID) error

	// BanUser bans a user account
	BanUser(ctx context.Context, userID uuidv7.UUID) error

	// ActivateUser activates a user account
	ActivateUser(ctx context.Context, userID uuidv7.UUID) error

	// UnlockUser manually unlocks a locked user account
	UnlockUser(ctx context.Context, userID uuidv7.UUID) error

	// ListUsers lists users with pagination
	ListUsers(ctx context.Context, limit, offset int) ([]*aggregate.User, int, error)
}

// UserUseCase implements IUserUseCase interface
type UserUseCase struct {
	userRepo repository.IUserRepository
	roleRepo rolerepository.IRoleRepository
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(userRepo repository.IUserRepository, roleRepo rolerepository.IRoleRepository) IUserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Register creates a new user account
func (u *UserUseCase) Register(ctx context.Context, email, name, password string) (*aggregate.User, error) {
	// Check if email already exists
	exists, err := u.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, usererrors.ErrEmailCheckFailed
	}
	if exists {
		return nil, usererrors.ErrEmailAlreadyExists
	}

	// Create new user
	user, err := aggregate.NewUser(email, password)
	if err != nil {
		return nil, err
	}

	// Save to repository
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, usererrors.ErrUserSaveFailed
	}

	// Assign default "user" role
	defaultRole, err := u.roleRepo.GetByName(ctx, "user")
	if err != nil {
		// Log warning but don't fail registration if role not found
		// (migrations might not have run yet in tests)
		slog.Warn("failed to get default user role",
			slog.String("error", err.Error()))
		return user, nil
	}

	if err := u.roleRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID, nil); err != nil {
		// Log warning but don't fail registration
		slog.Warn("failed to assign user role",
			slog.String("user_id", user.ID.String()),
			slog.String("error", err.Error()))
		return user, nil
	}

	slog.Info("assigned default user role",
		slog.String("user_id", user.ID.String()))

	// Reload user to get roles
	user, err = u.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		// Return user without roles if reload fails
		slog.Warn("failed to reload user after role assignment",
			slog.String("user_id", user.ID.String()),
			slog.String("error", err.Error()))
		return user, nil
	}

	return user, nil
}

// Authenticate authenticates a user with email and password
func (u *UserUseCase) Authenticate(ctx context.Context, email, password string) (*aggregate.User, error) {
	// Get user by email
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, usererrors.ErrUserNotFound) {
			return nil, usererrors.ErrInvalidCredentials
		}
		return nil, usererrors.ErrUserGetFailed
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, usererrors.ErrAccountLocked
	}

	// Check if account is active
	if !user.IsActive() {
		return nil, usererrors.ErrAccountNotActive
	}

	// Verify password
	if err := user.CheckPassword(password); err != nil {
		// Record failed login
		user.RecordFailedLogin()
		if updateErr := u.userRepo.Update(ctx, user); updateErr != nil {
			return nil, usererrors.ErrUserUpdateFailed
		}
		return nil, usererrors.ErrInvalidCredentials
	}

	// Record successful login
	user.RecordLogin()
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, usererrors.ErrUserUpdateFailed
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (u *UserUseCase) GetUser(ctx context.Context, userID uuidv7.UUID) (*aggregate.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, usererrors.ErrUserNotFound) {
			return nil, usererrors.ErrUserNotFound
		}
		return nil, usererrors.ErrUserGetFailed
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (u *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*aggregate.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, usererrors.ErrUserNotFound) {
			return nil, usererrors.ErrUserNotFound
		}
		return nil, usererrors.ErrUserGetFailed
	}
	return user, nil
}

// VerifyEmail marks user's email as verified
func (u *UserUseCase) VerifyEmail(ctx context.Context, userID uuidv7.UUID) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("verify email: %w", err)
	}

	if user.EmailVerified {
		return nil // Already verified
	}

	user.VerifyEmail()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return usererrors.ErrUserUpdateFailed
	}

	return nil
}

// ChangePassword changes user's password
func (u *UserUseCase) ChangePassword(ctx context.Context, userID uuidv7.UUID, oldPassword, newPassword string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}

	// Verify old password
	if err := user.CheckPassword(oldPassword); err != nil {
		return usererrors.ErrInvalidCredentials
	}

	// Change password
	if err := user.ChangePassword(newPassword); err != nil {
		return err
	}

	// Save
	if err := u.userRepo.Update(ctx, user); err != nil {
		return usererrors.ErrUserUpdateFailed
	}

	return nil
}

// SuspendUser suspends a user account
func (u *UserUseCase) SuspendUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("suspend user: %w", err)
	}

	user.Suspend()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return usererrors.ErrUserUpdateFailed
	}

	return nil
}

// BanUser bans a user account
func (u *UserUseCase) BanUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("ban user: %w", err)
	}

	user.Ban()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return usererrors.ErrUserUpdateFailed
	}

	return nil
}

// ActivateUser activates a user account
func (u *UserUseCase) ActivateUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("activate user: %w", err)
	}

	user.Activate()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return usererrors.ErrUserUpdateFailed
	}

	return nil
}

// UnlockUser manually unlocks a locked user account
func (u *UserUseCase) UnlockUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("unlock user: %w", err)
	}

	user.UnlockAccount()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return usererrors.ErrUserUpdateFailed
	}

	return nil
}

// ListUsers lists users with pagination
func (u *UserUseCase) ListUsers(ctx context.Context, limit, offset int) ([]*aggregate.User, int, error) {
	users, total, err := u.userRepo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, usererrors.ErrListUsersFailed
	}
	return users, total, nil
}
