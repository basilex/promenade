package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/domain/repository"
	jwtpkg "github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/uuidv7"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrUserSuspended      = errors.New("user account is suspended")
	ErrUserBanned         = errors.New("user account is banned")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrEmailNotVerified   = errors.New("email not verified")
)

type AuthUseCase interface {
	Register(ctx context.Context, email, name, password string) (*entity.User, error)
	Login(ctx context.Context, email, password, userAgent, ipAddress string) (accessToken, refreshToken string, user *entity.User, err error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error)
	GetMe(ctx context.Context, userID uuidv7.UUID) (*entity.User, error)
	GetUserSessions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Session, error)

	// Admin operations
	SuspendUser(ctx context.Context, userID uuidv7.UUID, reason string, until *time.Time) error
	BanUser(ctx context.Context, userID uuidv7.UUID, reason string) error
	ReactivateUser(ctx context.Context, userID uuidv7.UUID) error
}

type authUseCase struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	jwtManager  *jwtpkg.JWTManager
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	jwtManager *jwtpkg.JWTManager,
) AuthUseCase {
	return &authUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
	}
}

func (uc *authUseCase) Register(ctx context.Context, email, name, password string) (*entity.User, error) {
	// Check if user already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Create new user
	user := &entity.User{
		ID:        uuidv7.New(),
		Email:     email,
		Name:      name,
		Status:    entity.UserStatusUnverified,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	if err := user.HashPassword(password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Save user
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// TODO: Send verification email

	return user, nil
}

func (uc *authUseCase) Login(ctx context.Context, email, password, userAgent, ipAddress string) (string, string, *entity.User, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return "", "", nil, ErrInvalidCredentials
		}
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check password
	if !user.CheckPassword(password) {
		return "", "", nil, ErrInvalidCredentials
	}

	// Check user status
	if !user.CanLogin() {
		switch user.Status {
		case entity.UserStatusSuspended:
			return "", "", nil, ErrUserSuspended
		case entity.UserStatusBanned:
			return "", "", nil, ErrUserBanned
		case entity.UserStatusInactive:
			return "", "", nil, ErrUserNotActive
		default:
			return "", "", nil, ErrUserNotActive
		}
	}

	// Generate tokens
	accessToken, _, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := uc.generateRefreshToken()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Hash refresh token for storage
	hashedToken := uc.hashToken(refreshToken)

	// Create session
	session := &entity.Session{
		ID:           uuidv7.New(),
		UserID:       user.ID,
		RefreshToken: hashedToken,
		UserAgent:    &userAgent,
		IPAddress:    &ipAddress,
		ExpiresAt:    time.Now().Add(uc.jwtManager.GetRefreshTokenTTL()),
		CreatedAt:    time.Now(),
	}

	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return "", "", nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	if err := uc.userRepo.UpdateLastLogin(ctx, user.ID, time.Now()); err != nil {
		// Log error but don't fail the login
		fmt.Printf("failed to update last login: %v\n", err)
	}

	return accessToken, refreshToken, user, nil
}

func (uc *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	// Hash token to find session
	hashedToken := uc.hashToken(refreshToken)

	// Get session
	session, err := uc.sessionRepo.GetByRefreshToken(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return ErrInvalidToken
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Delete session
	if err := uc.sessionRepo.Delete(ctx, session.ID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// Hash token to find session
	hashedToken := uc.hashToken(refreshToken)

	// Get session
	session, err := uc.sessionRepo.GetByRefreshToken(ctx, hashedToken)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return "", "", ErrInvalidToken
		}
		return "", "", fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session is expired
	if session.IsExpired() {
		_ = uc.sessionRepo.Delete(ctx, session.ID)
		return "", "", ErrInvalidToken
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user can still login
	if !user.CanLogin() {
		return "", "", ErrUserNotActive
	}

	// Generate new tokens
	newAccessToken, _, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := uc.generateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with new refresh token
	newHashedToken := uc.hashToken(newRefreshToken)
	session.RefreshToken = newHashedToken
	session.ExpiresAt = time.Now().Add(uc.jwtManager.GetRefreshTokenTTL())

	// Delete old session and create new one
	if err := uc.sessionRepo.Delete(ctx, session.ID); err != nil {
		return "", "", fmt.Errorf("failed to delete old session: %w", err)
	}

	session.ID = uuidv7.New()
	session.CreatedAt = time.Now()
	if err := uc.sessionRepo.Create(ctx, session); err != nil {
		return "", "", fmt.Errorf("failed to create new session: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (uc *authUseCase) GetMe(ctx context.Context, userID uuidv7.UUID) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (uc *authUseCase) GetUserSessions(ctx context.Context, userID uuidv7.UUID) ([]*entity.Session, error) {
	sessions, err := uc.sessionRepo.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	return sessions, nil
}

func (uc *authUseCase) SuspendUser(ctx context.Context, userID uuidv7.UUID, reason string, until *time.Time) error {
	if err := uc.userRepo.Suspend(ctx, userID, reason, until); err != nil {
		return fmt.Errorf("failed to suspend user: %w", err)
	}

	// Invalidate all user sessions
	if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

func (uc *authUseCase) BanUser(ctx context.Context, userID uuidv7.UUID, reason string) error {
	if err := uc.userRepo.Ban(ctx, userID, reason); err != nil {
		return fmt.Errorf("failed to ban user: %w", err)
	}

	// Invalidate all user sessions
	if err := uc.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

func (uc *authUseCase) ReactivateUser(ctx context.Context, userID uuidv7.UUID) error {
	if err := uc.userRepo.Reactivate(ctx, userID); err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}
	return nil
}

// generateRefreshToken generates a cryptographically secure random refresh token
func (uc *authUseCase) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// hashToken hashes a token using SHA256
func (uc *authUseCase) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
