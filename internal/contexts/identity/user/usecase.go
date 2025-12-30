package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/basilex/promenade/internal/contexts/identity/role"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines the interface for User use cases
type IUseCase interface {
	// Register creates a new user account
	Register(ctx context.Context, email, name, password string) (*User, error)

	// Authenticate authenticates a user with email and password
	Authenticate(ctx context.Context, email, password string) (*User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, userID uuidv7.UUID) (*User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)

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
	ListUsers(ctx context.Context, limit, offset int) ([]*User, int, error)
}

// UseCase implements IUseCase interface
type UseCase struct {
	userRepo IRepository
	roleRepo role.IRepository
}

// NewUseCase creates a new user use case
func NewUseCase(userRepo IRepository, roleRepo role.IRepository) IUseCase {
	return &UseCase{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

// Register creates a new user account
func (uc *UseCase) Register(ctx context.Context, email, name, password string) (*User, error) {
	// Check if email already exists
	exists, err := uc.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Create new user
	user, err := NewUser(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Assign default "user" role
	defaultRole, err := uc.roleRepo.GetByName(ctx, "user")
	if err != nil {
		// Log warning but don't fail registration if role not found
		// (migrations might not have run yet in tests)
		fmt.Printf("[WARN] Failed to get default 'user' role: %v\n", err)
		return user, nil
	}

	if err := uc.roleRepo.AssignRoleToUser(ctx, user.ID, defaultRole.ID, nil); err != nil {
		// Log warning but don't fail registration
		fmt.Printf("[WARN] Failed to assign 'user' role to user %s: %v\n", user.ID.String(), err)
		return user, nil
	}

	fmt.Printf("[INFO] Assigned 'user' role to user %s\n", user.ID.String())

	// Reload user to get roles
	user, err = uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		// Return user without roles if reload fails
		fmt.Printf("[WARN] Failed to reload user after role assignment: %v\n", err)
		return user, nil
	}

	return user, nil
}

// Authenticate authenticates a user with email and password
func (uc *UseCase) Authenticate(ctx context.Context, email, password string) (*User, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, ErrAccountLocked
	}

	// Check if account is active
	if !user.IsActive() {
		return nil, ErrAccountNotActive
	}

	// Verify password
	if err := user.CheckPassword(password); err != nil {
		// Record failed login
		user.RecordFailedLogin()
		if updateErr := uc.userRepo.Update(ctx, user); updateErr != nil {
			return nil, fmt.Errorf("failed to update failed login count: %w", updateErr)
		}
		return nil, ErrInvalidCredentials
	}

	// Record successful login
	user.RecordLogin()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update login time: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (uc *UseCase) GetUser(ctx context.Context, userID uuidv7.UUID) (*User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (uc *UseCase) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// VerifyEmail marks user's email as verified
func (uc *UseCase) VerifyEmail(ctx context.Context, userID uuidv7.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.EmailVerified {
		return nil // Already verified
	}

	user.VerifyEmail()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ChangePassword changes user's password
func (uc *UseCase) ChangePassword(ctx context.Context, userID uuidv7.UUID, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if err := user.CheckPassword(oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Change password
	if err := user.ChangePassword(newPassword); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	// Save
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// SuspendUser suspends a user account
func (uc *UseCase) SuspendUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Suspend()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// BanUser bans a user account
func (uc *UseCase) BanUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Ban()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ActivateUser activates a user account
func (uc *UseCase) ActivateUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.Activate()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UnlockUser manually unlocks a locked user account
func (uc *UseCase) UnlockUser(ctx context.Context, userID uuidv7.UUID) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.UnlockAccount()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// ListUsers lists users with pagination
func (uc *UseCase) ListUsers(ctx context.Context, limit, offset int) ([]*User, int, error) {
	users, total, err := uc.userRepo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	return users, total, nil
}
