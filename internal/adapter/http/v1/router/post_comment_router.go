package router

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
)

type PostCommentRouter struct {
	handler *handler.PostCommentHandler
	authMw  *middleware.AuthMiddleware
}

func NewPostCommentRouter(handler *handler.PostCommentHandler, authMw *middleware.AuthMiddleware) *PostCommentRouter {
	return &PostCommentRouter{
		handler: handler,
		authMw:  authMw,
	}
}

func (r *PostCommentRouter) Setup(rg *gin.RouterGroup) {
	comments := rg.Group("/comments")
	{
		// Public routes
		comments.GET("/:id", r.handler.GetComment)                // GET /v1/comments/:id
		comments.GET("/:id/replies", r.handler.GetCommentReplies) // GET /v1/comments/:id/replies

		// Protected routes (require authentication)
		comments.POST("", r.authMw.RequireAuth(), r.handler.CreateComment)            // POST /v1/comments
		comments.PUT("/:id", r.authMw.RequireAuth(), r.handler.UpdateComment)         // PUT /v1/comments/:id
		comments.DELETE("/:id", r.authMw.RequireAuth(), r.handler.DeleteComment)      // DELETE /v1/comments/:id
		comments.POST("/:id/like", r.authMw.RequireAuth(), r.handler.LikeComment)     // POST /v1/comments/:id/like
		comments.DELETE("/:id/like", r.authMw.RequireAuth(), r.handler.UnlikeComment) // DELETE /v1/comments/:id/like
	}

	// Post-specific comments via query parameter (public)
	// Note: We use query parameter to avoid route conflict with /posts/:id from UserPostRouter
	comments.GET("", r.handler.GetPostComments) // GET /v1/comments?post_id=xxx

	// User-specific comments (public)
	users := rg.Group("/users")
	{
		users.GET("/:user_id/comments", r.handler.GetUserComments) // GET /v1/users/:user_id/comments
	}
}
