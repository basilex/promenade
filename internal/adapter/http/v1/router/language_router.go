package router

import (
	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

type LanguageRouter struct {
	languageHandler *handler.LanguageHandler
	authMiddleware  *middleware.AuthMiddleware
	authzMiddleware *middleware.AuthorizationMiddleware
}

func NewLanguageRouter(
	languageHandler *handler.LanguageHandler,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *LanguageRouter {
	return &LanguageRouter{
		languageHandler: languageHandler,
		authMiddleware:  authMiddleware,
		authzMiddleware: authzMiddleware,
	}
}

func (r *LanguageRouter) Setup(rg *gin.RouterGroup) {
	languages := rg.Group("/languages")
	{
		// Public routes - anyone can list/view languages
		languages.GET("", r.languageHandler.List)
		languages.GET("/active", r.languageHandler.ListActive)
		languages.GET("/:id", r.languageHandler.GetByID)
		languages.GET("/code/:code", r.languageHandler.GetByCode)

		// Protected routes - require admin permissions
		languages.POST("",
			r.authMiddleware.RequireAuth(),
			r.authzMiddleware.RequirePermission("languages:create"),
			r.languageHandler.Create,
		)
		languages.PUT("/:id",
			r.authMiddleware.RequireAuth(),
			r.authzMiddleware.RequirePermission("languages:update"),
			r.languageHandler.Update,
		)
		languages.DELETE("/:id",
			r.authMiddleware.RequireAuth(),
			r.authzMiddleware.RequirePermission("languages:delete"),
			r.languageHandler.Delete,
		)
	}
}
