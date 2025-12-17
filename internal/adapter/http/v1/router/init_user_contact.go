package router

import (
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
)

// InitUserContactModule initializes the user contact module
func InitUserContactModule(db *sqlx.DB, authMw *middleware.AuthMiddleware) *UserContactRouter {
	// Initialize repository
	repo := postgres.NewUserContactRepository(db)

	// Initialize use case
	uc := usecase.NewUserContactUseCase(repo)

	// Initialize handler
	h := handler.NewUserContactHandler(uc)

	// Initialize and return router
	return NewUserContactRouter(h, authMw)
}
