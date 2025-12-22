package router

import (
	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

type TimezoneRouter struct {
	timezoneHandler *handler.TimezoneHandler
	authMiddleware  *middleware.AuthMiddleware
	authzMiddleware *middleware.AuthorizationMiddleware
}

func NewTimezoneRouter(
	timezoneHandler *handler.TimezoneHandler,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *TimezoneRouter {
	return &TimezoneRouter{
		timezoneHandler: timezoneHandler,
		authMiddleware:  authMiddleware,
		authzMiddleware: authzMiddleware,
	}
}

func (r *TimezoneRouter) Setup(rg *gin.RouterGroup) {
	timezones := rg.Group("/timezones")
	{
		// Public routes - anyone can list/view timezones
		timezones.GET("", r.timezoneHandler.List)
		timezones.GET("/active", r.timezoneHandler.ListActive)
		timezones.GET("/:id", r.timezoneHandler.GetByID)
		timezones.GET("/name/:name", r.timezoneHandler.GetByName)

		// Protected routes - require admin permissions
		timezones.POST("",
			r.authMiddleware.RequireAuth(),
			r.authzMiddleware.RequirePermission("timezones:create"),
			r.timezoneHandler.Create,
		)
		timezones.PUT("/:id",
			r.authMiddleware.RequireAuth(),
			r.authzMiddleware.RequirePermission("timezones:update"),
			r.timezoneHandler.Update,
		)
		timezones.DELETE("/:id",
			r.authMiddleware.RequireAuth(),
			r.authzMiddleware.RequirePermission("timezones:delete"),
			r.timezoneHandler.Delete,
		)
	}
}
