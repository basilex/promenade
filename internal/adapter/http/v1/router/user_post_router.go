package router

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
)

type UserPostRouter struct {
	handler *handler.UserPostHandler
	authMw  *middleware.AuthMiddleware
}

func NewUserPostRouter(handler *handler.UserPostHandler, authMw *middleware.AuthMiddleware) *UserPostRouter {
	return &UserPostRouter{
		handler: handler,
		authMw:  authMw,
	}
}

func (r *UserPostRouter) Setup(rg *gin.RouterGroup) {
	posts := rg.Group("/posts")
	{
		// Public routes
		posts.GET("/published", r.handler.GetPublishedPosts) // GET /v1/posts/published
		posts.GET("/featured", r.handler.GetFeaturedPosts)   // GET /v1/posts/featured
		posts.GET("/search", r.handler.SearchPosts)          // GET /v1/posts/search?q=query
		posts.GET("/tag/:tag", r.handler.GetPostsByTag)      // GET /v1/posts/tag/:tag
		posts.GET("/:id", r.handler.GetPost)                 // GET /v1/posts/:id
		posts.POST("/:id/view", r.handler.ViewPost)          // POST /v1/posts/:id/view
		posts.POST("/:id/like", r.handler.LikePost)          // POST /v1/posts/:id/like
		posts.DELETE("/:id/like", r.handler.UnlikePost)      // DELETE /v1/posts/:id/like
		posts.GET("/list", r.handler.ListPosts)              // GET /v1/posts/list

		// Protected routes (require authentication)
		posts.POST("", r.authMw.RequireAuth(), r.handler.CreatePost)                  // POST /v1/posts
		posts.PUT("/:id", r.authMw.RequireAuth(), r.handler.UpdatePost)               // PUT /v1/posts/:id
		posts.DELETE("/:id", r.authMw.RequireAuth(), r.handler.DeletePost)            // DELETE /v1/posts/:id
		posts.POST("/:id/publish", r.authMw.RequireAuth(), r.handler.PublishPost)     // POST /v1/posts/:id/publish
		posts.POST("/:id/unpublish", r.authMw.RequireAuth(), r.handler.UnpublishPost) // POST /v1/posts/:id/unpublish
		posts.POST("/:id/featured", r.authMw.RequireAuth(), r.handler.ToggleFeatured) // POST /v1/posts/:id/featured
		posts.POST("/:id/comments", r.authMw.RequireAuth(), r.handler.ToggleComments) // POST /v1/posts/:id/comments
	}

	// User-specific posts
	users := rg.Group("/users")
	{
		users.GET("/:user_id/posts", r.handler.GetUserPosts) // GET /v1/users/:user_id/posts
	}
}
