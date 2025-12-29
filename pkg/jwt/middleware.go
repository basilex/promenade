package jwt

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

const (
	// ClaimsContextKey is the key for JWT claims in Gin context
	ClaimsContextKey = "jwt_claims"
	// UserIDContextKey is the key for user ID in Gin context
	UserIDContextKey = "user_id"
)

// AuthMiddleware validates JWT token from Authorization header.
// If TokenRevoker is provided, also checks if token is revoked.
func AuthMiddleware(manager *Manager, revoker *TokenRevoker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		// Check if token is revoked (if revoker provided)
		if revoker != nil {
			revoked, err := revoker.IsRevoked(c.Request.Context(), token)
			if err != nil {
				response.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check token revocation")
				c.Abort()
				return
			}
			if revoked {
				response.ErrorResponse(c, http.StatusUnauthorized, "TOKEN_REVOKED", "Token has been revoked")
				c.Abort()
				return
			}
		}

		claims, err := manager.ValidateAccessToken(token)
		if err != nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ClaimsContextKey, claims)
		c.Set(UserIDContextKey, claims.UserID)
		c.Next()
	}
}

// RequireRole checks if the user has a specific role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "No authentication claims found")
			c.Abort()
			return
		}

		if !claims.HasRole(role) {
			response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole checks if the user has any of the specified roles
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "No authentication claims found")
			c.Abort()
			return
		}

		if !claims.HasAnyRole(roles...) {
			response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllRoles checks if the user has all of the specified roles
func RequireAllRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			response.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "No authentication claims found")
			c.Abort()
			return
		}

		if !claims.HasAllRoles(roles...) {
			response.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetClaims extracts JWT claims from Gin context
func GetClaims(c *gin.Context) *Claims {
	value, exists := c.Get(ClaimsContextKey)
	if !exists {
		return nil
	}

	claims, ok := value.(*Claims)
	if !ok {
		return nil
	}

	return claims
}

// GetUserID extracts user ID from Gin context
func GetUserID(c *gin.Context) string {
	value, exists := c.Get(UserIDContextKey)
	if !exists {
		return ""
	}

	userID, ok := value.(string)
	if !ok {
		return ""
	}

	return userID
}

// MustGetClaims extracts JWT claims or panics
func MustGetClaims(c *gin.Context) *Claims {
	claims := GetClaims(c)
	if claims == nil {
		panic("JWT claims not found in context")
	}
	return claims
}

// MustGetUserID extracts user ID or panics
func MustGetUserID(c *gin.Context) uuidv7.UUID {
	userID := GetUserID(c)
	if userID == "" {
		panic("User ID not found in context")
	}

	uuid, err := uuidv7.Parse(userID)
	if err != nil {
		panic("Invalid user ID in context: " + err.Error())
	}

	return uuid
}
