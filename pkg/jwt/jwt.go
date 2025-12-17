package jwt

import (
	"errors"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID uuidv7.UUID `json:"user_id"`
	Email  string      `json:"email"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type JWTManager struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTManager(secretKey string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateTokenPair generates access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID uuidv7.UUID, email string) (*TokenPair, error) {
	accessToken, expiresAt, err := m.generateToken(userID, email, m.accessTokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := m.generateToken(userID, email, m.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// GenerateAccessToken generates only access token
func (m *JWTManager) GenerateAccessToken(userID uuidv7.UUID, email string) (string, time.Time, error) {
	return m.generateToken(userID, email, m.accessTokenTTL)
}

func (m *JWTManager) generateToken(userID uuidv7.UUID, email string, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates and parses JWT token
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// RefreshAccessToken creates new access token from refresh token
func (m *JWTManager) RefreshAccessToken(refreshToken string) (string, time.Time, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", time.Time{}, err
	}

	return m.GenerateAccessToken(claims.UserID, claims.Email)
}

// GetRefreshTokenTTL returns refresh token TTL
func (m *JWTManager) GetRefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}

// GetAccessTokenTTL returns access token TTL
func (m *JWTManager) GetAccessTokenTTL() time.Duration {
	return m.accessTokenTTL
}
