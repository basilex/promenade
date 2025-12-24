package middleware

import (
	"net/http"

	"github.com/basilex/promenade/internal/adapter/http/shared/response"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/gin-gonic/gin"
)

// AuthorizationMiddleware handles RBAC permission checks
type AuthorizationMiddleware struct {
	roleUseCase usecase.IRoleUseCase
}

func NewAuthorizationMiddleware(roleUseCase usecase.IRoleUseCase) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		roleUseCase: roleUseCase,
	}
}

// RequirePermission checks if the authenticated user has the specified permission
// Must be used after RequireAuth middleware
func (m *AuthorizationMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
			c.Abort()
			return
		}

		hasPermission, err := m.roleUseCase.HasPermission(c.Request.Context(), userID, permission)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to check permissions", err)
			c.Abort()
			return
		}

		if !hasPermission {
			response.Error(c, http.StatusForbidden, "insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission checks if the authenticated user has at least one of the specified permissions
// Must be used after RequireAuth middleware
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
			c.Abort()
			return
		}

		hasAny, err := m.roleUseCase.HasAnyPermission(c.Request.Context(), userID, permissions)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to check permissions", err)
			c.Abort()
			return
		}

		if !hasAny {
			response.Error(c, http.StatusForbidden, "insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions checks if the authenticated user has all of the specified permissions
// Must be used after RequireAuth middleware
func (m *AuthorizationMiddleware) RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
			c.Abort()
			return
		}

		hasAll, err := m.roleUseCase.HasAllPermissions(c.Request.Context(), userID, permissions)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to check permissions", err)
			c.Abort()
			return
		}

		if !hasAll {
			response.Error(c, http.StatusForbidden, "insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if the authenticated user has the specified role
// Must be used after RequireAuth middleware
func (m *AuthorizationMiddleware) RequireRole(roleName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
			c.Abort()
			return
		}

		roles, err := m.roleUseCase.GetUserActiveRoles(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to check roles", err)
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range roles {
			if role.Name == roleName {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.Error(c, http.StatusForbidden, "insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole checks if the authenticated user has at least one of the specified roles
// Must be used after RequireAuth middleware
func (m *AuthorizationMiddleware) RequireAnyRole(roleNames ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "user not authenticated", nil)
			c.Abort()
			return
		}

		roles, err := m.roleUseCase.GetUserActiveRoles(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "failed to check roles", err)
			c.Abort()
			return
		}

		userRoleNames := make(map[string]bool)
		for _, role := range roles {
			userRoleNames[role.Name] = true
		}

		hasAnyRole := false
		for _, requiredRole := range roleNames {
			if userRoleNames[requiredRole] {
				hasAnyRole = true
				break
			}
		}

		if !hasAnyRole {
			response.Error(c, http.StatusForbidden, "insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
