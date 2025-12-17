package router

import (
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
)

func InitUserPostModule(db *sqlx.DB, authMw *middleware.AuthMiddleware) *UserPostRouter {
	// Repository
	postRepo := postgres.NewUserPostRepository(db)

	// Use case
	postUC := usecase.NewUserPostUseCase(postRepo)

	// Handler
	postHandler := handler.NewUserPostHandler(postUC)

	// Router
	return NewUserPostRouter(postHandler, authMw)
}
