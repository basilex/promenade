package router

import (
	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

type UserProfileRouter struct {
	handler *handler.UserProfileHandler
	authMw  *middleware.AuthMiddleware
}

func NewUserProfileRouter(handler *handler.UserProfileHandler, authMw *middleware.AuthMiddleware) *UserProfileRouter {
	return &UserProfileRouter{
		handler: handler,
		authMw:  authMw,
	}
}

func (r *UserProfileRouter) Setup(rg *gin.RouterGroup) {
	profiles := rg.Group("/profiles")
	{
		// Public routes
		profiles.GET("", r.handler.ListProfiles)                            // GET /v1/profiles
		profiles.GET("/search", r.handler.SearchProfiles)                   // GET /v1/profiles/search?q=query
		profiles.GET("/:id", r.handler.GetProfile)                          // GET /v1/profiles/:id
		profiles.GET("/nickname/:nickname", r.handler.GetProfileByNickname) // GET /v1/profiles/nickname/:nickname

		// Protected routes (require authentication)
		profiles.POST("", r.authMw.RequireAuth(), r.handler.CreateProfile)       // POST /v1/profiles
		profiles.GET("/me", r.authMw.RequireAuth(), r.handler.GetMyProfile)      // GET /v1/profiles/me
		profiles.PUT("/:id", r.authMw.RequireAuth(), r.handler.UpdateProfile)    // PUT /v1/profiles/:id
		profiles.DELETE("/:id", r.authMw.RequireAuth(), r.handler.DeleteProfile) // DELETE /v1/profiles/:id

		// Admin routes (TODO: add admin middleware)
		profiles.POST("/:id/ban", r.authMw.RequireAuth(), r.handler.BanProfile)           // POST /v1/profiles/:id/ban
		profiles.POST("/:id/unban", r.authMw.RequireAuth(), r.handler.UnbanProfile)       // POST /v1/profiles/:id/unban
		profiles.POST("/:id/verify", r.authMw.RequireAuth(), r.handler.VerifyProfile)     // POST /v1/profiles/:id/verify
		profiles.POST("/:id/unverify", r.authMw.RequireAuth(), r.handler.UnverifyProfile) // POST /v1/profiles/:id/unverify
	}
}
