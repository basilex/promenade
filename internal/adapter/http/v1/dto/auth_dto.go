package dto

import (
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=8"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	User    *UserResponse `json:"user"`
	Message string        `json:"message"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int           `json:"expires_in"`
	User         *UserResponse `json:"user"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// SuspendUserRequest represents a suspend user request
type SuspendUserRequest struct {
	Reason string     `json:"reason" binding:"required"`
	Until  *time.Time `json:"until,omitempty"`
}

// BanUserRequest represents a ban user request
type BanUserRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID              uuidv7.UUID `json:"id"`
	Email           string      `json:"email"`
	Name            string      `json:"name"`
	Status          string      `json:"status"`
	EmailVerifiedAt *time.Time  `json:"email_verified_at"`
	SuspendedReason *string     `json:"suspended_reason,omitempty"`
	SuspendedUntil  *time.Time  `json:"suspended_until,omitempty"`
	LastLoginAt     *time.Time  `json:"last_login_at,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// SessionResponse represents a session response
type SessionResponse struct {
	ID        uuidv7.UUID `json:"id"`
	UserAgent *string     `json:"user_agent,omitempty"`
	IPAddress *string     `json:"ip_address,omitempty"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
}

// ToUserResponse converts entity.User to UserResponse
func ToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		Name:            user.Name,
		Status:          string(user.Status),
		EmailVerifiedAt: user.EmailVerifiedAt,
		SuspendedReason: user.SuspendedReason,
		SuspendedUntil:  user.SuspendedUntil,
		LastLoginAt:     user.LastLoginAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// ToSessionResponse converts entity.Session to SessionResponse
func ToSessionResponse(session *entity.Session) *SessionResponse {
	return &SessionResponse{
		ID:        session.ID,
		UserAgent: session.UserAgent,
		IPAddress: session.IPAddress,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
	}
}

// ToSessionResponses converts multiple sessions to responses
func ToSessionResponses(sessions []*entity.Session) []*SessionResponse {
	responses := make([]*SessionResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = ToSessionResponse(session)
	}
	return responses
}
