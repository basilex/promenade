package router

import (
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
)

// InitLanguageModule initializes the complete language module with all dependencies
// This encapsulates the entire dependency chain: repo → usecase → handler → router
func InitLanguageModule(
	db *sqlx.DB,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *LanguageRouter {
	// Repository layer
	languageRepo := postgres.NewLanguageRepository(db)

	// Use case layer
	languageUseCase := usecase.NewLanguageUseCase(languageRepo)

	// Handler layer
	languageHandler := handler.NewLanguageHandler(languageUseCase)

	// Router layer
	return NewLanguageRouter(languageHandler, authMiddleware, authzMiddleware)
}
