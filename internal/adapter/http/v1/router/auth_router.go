package router

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
)

// AuthRouter handles all authentication and user management routes
type AuthRouter struct {
	authHandler     *handler.AuthHandler
	authMiddleware  *middleware.AuthMiddleware
	authzMiddleware *middleware.AuthorizationMiddleware
}

// NewAuthRouter creates a new auth router instance
func NewAuthRouter(
	authHandler *handler.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *AuthRouter {
	return &AuthRouter{
		authHandler:     authHandler,
		authMiddleware:  authMiddleware,
		authzMiddleware: authzMiddleware,
	}
}

// Setup registers all auth routes under /auth prefix
func (r *AuthRouter) Setup(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")

	// ============================================
	// PUBLIC ROUTES (no authentication required)
	// ============================================
	r.setupPublicRoutes(auth)

	// ============================================
	// PROTECTED ROUTES (authentication required)
	// ============================================
	protected := auth.Group("")
	protected.Use(r.authMiddleware.RequireAuth())
	r.setupProtectedRoutes(protected)

	// ============================================
	// ADMIN ROUTES (require authentication + admin permissions)
	// ============================================
	admin := auth.Group("/users")
	admin.Use(r.authMiddleware.RequireAuth())
	admin.Use(r.authzMiddleware.RequirePermission("users:manage"))
	r.setupAdminRoutes(admin)
}

// setupPublicRoutes registers public authentication routes
func (r *AuthRouter) setupPublicRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", r.authHandler.Register)
	rg.POST("/login", r.authHandler.Login)
	rg.POST("/logout", r.authHandler.Logout)
	rg.POST("/refresh", r.authHandler.RefreshToken)
}

// setupProtectedRoutes registers protected user routes (requires authentication)
func (r *AuthRouter) setupProtectedRoutes(rg *gin.RouterGroup) {
	rg.GET("/me", r.authHandler.GetMe)
	rg.GET("/sessions", r.authHandler.GetUserSessions)
}

// setupAdminRoutes registers admin user management routes (requires admin role)
func (r *AuthRouter) setupAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("/:id/suspend", r.authHandler.SuspendUser)
	rg.POST("/:id/ban", r.authHandler.BanUser)
	rg.POST("/:id/reactivate", r.authHandler.ReactivateUser)
}
