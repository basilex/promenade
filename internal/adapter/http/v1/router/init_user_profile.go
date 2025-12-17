package router

import (
	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/jmoiron/sqlx"
)

func InitUserProfileModule(db *sqlx.DB, authMw *middleware.AuthMiddleware) *UserProfileRouter {
	// Initialize repository
	profileRepo := postgres.NewUserProfileRepository(db)

	// Initialize use case
	profileUC := usecase.NewUserProfileUseCase(profileRepo)

	// Initialize handler
	profileHandler := handler.NewUserProfileHandler(profileUC)

	// Return router
	return NewUserProfileRouter(profileHandler, authMw)
}
