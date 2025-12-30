package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func setupTestRouter() (*gin.Engine, *Manager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	manager := NewManager(Config{SecretKey: "test-secret"})
	return router, manager
}

func TestAuthMiddleware(t *testing.T) {
	router, manager := setupTestRouter()

	router.GET("/protected", AuthMiddleware(manager, nil), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	userID := uuidv7.New()
	email := "test@example.com"
	roles := []string{"user"}
	tokenPair, _ := manager.GenerateTokenPair(userID, email, roles)

	t.Run("valid token passes", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequireRole(t *testing.T) {
	router, manager := setupTestRouter()

	router.GET("/admin", AuthMiddleware(manager, nil), RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access"})
	})

	t.Run("user has required role", func(t *testing.T) {
		userID := uuidv7.New()
		tokenPair, _ := manager.GenerateTokenPair(userID, "admin@example.com", []string{"admin"})

		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("user lacks required role", func(t *testing.T) {
		userID := uuidv7.New()
		tokenPair, _ := manager.GenerateTokenPair(userID, "user@example.com", []string{"user"})

		req := httptest.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireAnyRole(t *testing.T) {
	router, manager := setupTestRouter()

	router.GET("/moderator", AuthMiddleware(manager, nil), RequireAnyRole("admin", "moderator"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "moderator access"})
	})

	t.Run("user has one of required roles", func(t *testing.T) {
		userID := uuidv7.New()
		tokenPair, _ := manager.GenerateTokenPair(userID, "mod@example.com", []string{"moderator"})

		req := httptest.NewRequest("GET", "/moderator", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("user lacks all required roles", func(t *testing.T) {
		userID := uuidv7.New()
		tokenPair, _ := manager.GenerateTokenPair(userID, "user@example.com", []string{"user"})

		req := httptest.NewRequest("GET", "/moderator", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireAllRoles(t *testing.T) {
	router, manager := setupTestRouter()

	router.GET("/superadmin", AuthMiddleware(manager, nil), RequireAllRoles("admin", "superadmin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "superadmin access"})
	})

	t.Run("user has all required roles", func(t *testing.T) {
		userID := uuidv7.New()
		tokenPair, _ := manager.GenerateTokenPair(userID, "super@example.com", []string{"admin", "superadmin"})

		req := httptest.NewRequest("GET", "/superadmin", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("user lacks some required roles", func(t *testing.T) {
		userID := uuidv7.New()
		tokenPair, _ := manager.GenerateTokenPair(userID, "admin@example.com", []string{"admin"})

		req := httptest.NewRequest("GET", "/superadmin", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestGetClaims(t *testing.T) {
	router, manager := setupTestRouter()

	router.GET("/test", AuthMiddleware(manager, nil), func(c *gin.Context) {
		claims := GetClaims(c)
		if claims != nil {
			c.JSON(http.StatusOK, gin.H{"email": claims.Email})
		}
	})

	t.Run("gets claims after auth middleware", func(t *testing.T) {
		userID := uuidv7.New()
		email := "test@example.com"
		tokenPair, _ := manager.GenerateTokenPair(userID, email, []string{"user"})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), email)
	})

	t.Run("returns nil without auth middleware", func(t *testing.T) {
		router2 := gin.New()
		router2.GET("/unprotected", func(c *gin.Context) {
			claims := GetClaims(c)
			assert.Nil(t, claims)
			c.JSON(http.StatusOK, gin.H{"message": "ok"})
		})

		req := httptest.NewRequest("GET", "/unprotected", nil)
		w := httptest.NewRecorder()

		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetUserID(t *testing.T) {
	router, manager := setupTestRouter()

	router.GET("/user-id", AuthMiddleware(manager, nil), func(c *gin.Context) {
		userID := GetUserID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	t.Run("gets user ID after auth middleware", func(t *testing.T) {
		userID := uuidv7.New()
		email := "test@example.com"
		tokenPair, _ := manager.GenerateTokenPair(userID, email, []string{"user"})

		req := httptest.NewRequest("GET", "/user-id", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), userID.String())
	})

	t.Run("returns empty string without auth middleware", func(t *testing.T) {
		router2 := gin.New()
		router2.GET("/unprotected", func(c *gin.Context) {
			userID := GetUserID(c)
			assert.Empty(t, userID)
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req := httptest.NewRequest("GET", "/unprotected", nil)
		w := httptest.NewRecorder()

		router2.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
