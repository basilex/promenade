package router

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
)

type UserContactRouter struct {
	handler *handler.UserContactHandler
	authMw  *middleware.AuthMiddleware
}

func NewUserContactRouter(handler *handler.UserContactHandler, authMw *middleware.AuthMiddleware) *UserContactRouter {
	return &UserContactRouter{
		handler: handler,
		authMw:  authMw,
	}
}

func (r *UserContactRouter) Setup(rg *gin.RouterGroup) {
	contacts := rg.Group("/users/contacts")
	contacts.Use(r.authMw.RequireAuth())
	{
		contacts.POST("", r.handler.CreateContact)
		contacts.GET("", r.handler.GetUserContacts)
		contacts.GET("/:id", r.handler.GetContact)
		contacts.PUT("/:id", r.handler.UpdateContact)
		contacts.DELETE("/:id", r.handler.DeleteContact)
		contacts.POST("/:id/primary", r.handler.SetPrimaryContact)
		contacts.POST("/:id/toggle", r.handler.ToggleContactActive)

		// Type-based endpoints
		contacts.GET("/type/:type", r.handler.GetContactsByType)
		contacts.GET("/primary/:type", r.handler.GetPrimaryContact)
	}
}
