package router

import (
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/repository"
	"github.com/basilex/promenade/internal/usecase"
)

func InitPostCommentModule(
	db *sqlx.DB,
	authMw *middleware.AuthMiddleware,
	userPostRepo repository.UserPostRepository,
) *PostCommentRouter {
	// Repository layer
	commentRepo := postgres.NewPostCommentRepository(db)

	// Use case layer
	commentUC := usecase.NewPostCommentUseCase(commentRepo, userPostRepo)

	// Handler layer
	commentHandler := handler.NewPostCommentHandler(commentUC)

	// Router layer
	return NewPostCommentRouter(commentHandler, authMw)
}
