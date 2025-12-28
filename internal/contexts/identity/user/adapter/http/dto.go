package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/user"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Status          string     `json:"status"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// AuthResponse represents authentication response with JWT tokens
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

// ToUserResponse converts a User entity to UserResponse DTO
func ToUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:              u.ID.String(),
		Email:           u.Email.Value(),
		Status:          string(u.Status),
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		LastLoginAt:     u.LastLoginAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// ToUserListResponse converts a slice of users to response DTOs
func ToUserListResponse(users []*user.User) []UserResponse {
	result := make([]UserResponse, len(users))
	for i, u := range users {
		result[i] = ToUserResponse(u)
	}
	return result
}
