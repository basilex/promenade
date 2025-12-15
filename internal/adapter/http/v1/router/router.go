package router

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
)

type V1Router struct {
	authHandler    *handler.AuthHandler
	productHandler *handler.ProductHandler
	roleHandler    *handler.RoleHandler
	authMiddleware *middleware.AuthMiddleware
}

func NewV1Router(
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	roleHandler *handler.RoleHandler,
	authMiddleware *middleware.AuthMiddleware,
) *V1Router {
	return &V1Router{
		authHandler:    authHandler,
		productHandler: productHandler,
		roleHandler:    roleHandler,
		authMiddleware: authMiddleware,
	}
}

func (r *V1Router) Setup(rg *gin.RouterGroup) {
	v1 := rg.Group("/v1")

	// ============================================
	// PUBLIC ROUTES (пока без аутентификации)
	// ============================================

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.RefreshToken)
	}

	// ============================================
	// PROTECTED ROUTES (требуют аутентификацию)
	// ============================================

	// Auth protected routes
	authProtected := v1.Group("/auth")
	authProtected.Use(r.authMiddleware.RequireAuth())
	{
		authProtected.GET("/me", r.authHandler.GetMe)
		authProtected.POST("/logout", r.authHandler.Logout)
	}

	// Product routes (защищенные)
	products := v1.Group("/products")
	products.Use(r.authMiddleware.RequireAuth())
	{
		products.POST("", r.productHandler.Create)
		products.GET("/:id", r.productHandler.GetByID)
		products.GET("", r.productHandler.List)
		products.PUT("/:id", r.productHandler.Update)
		products.DELETE("/:id", r.productHandler.Delete)
	}

	// Role routes (защищенные)
	roles := v1.Group("/roles")
	roles.Use(r.authMiddleware.RequireAuth())
	{
		roles.POST("", r.roleHandler.Create)
		roles.GET("/:id", r.roleHandler.GetByID)
		roles.GET("", r.roleHandler.List)
		roles.PUT("/:id", r.roleHandler.Update)
		roles.DELETE("/:id", r.roleHandler.Delete)
	}
}
