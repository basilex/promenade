package router

import (
	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

// AdminRouter handles admin routes
type AdminRouter struct {
	purgeHandler *handler.AdminPurgeHandler
	authMW       *middleware.AuthMiddleware
	authzMW      *middleware.AuthorizationMiddleware
}

// NewAdminRouter creates a new admin router
func NewAdminRouter(
	purgeHandler *handler.AdminPurgeHandler,
	authMW *middleware.AuthMiddleware,
	authzMW *middleware.AuthorizationMiddleware,
) *AdminRouter {
	return &AdminRouter{
		purgeHandler: purgeHandler,
		authMW:       authMW,
		authzMW:      authzMW,
	}
}

// Setup configures admin routes
func (r *AdminRouter) Setup(router *gin.RouterGroup) {
	admin := router.Group("/admin")
	admin.Use(r.authMW.RequireAuth())
	// Only superadmin and admin roles can access purge operations
	admin.Use(r.authzMW.RequirePermission("admin:purge"))

	// Purge endpoints
	purge := admin.Group("/purge")
	{
		purge.POST("/trigger", r.purgeHandler.TriggerPurge)
		purge.GET("/policies", r.purgeHandler.GetRetentionPolicies)
		purge.GET("/preview/:entity_name", r.purgeHandler.PreviewPurge)
		purge.GET("/scheduler/status", r.purgeHandler.GetSchedulerStatus)
	}
}
