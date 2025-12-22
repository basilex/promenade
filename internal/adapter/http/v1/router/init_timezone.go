package router

import (
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
)

// InitTimezoneModule initializes the complete timezone module with all dependencies
// This encapsulates the entire dependency chain: repo → usecase → handler → router
func InitTimezoneModule(
	db *sqlx.DB,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *TimezoneRouter {
	// Repository layer
	timezoneRepo := postgres.NewTimezoneRepository(db)

	// Use case layer
	timezoneUseCase := usecase.NewTimezoneUseCase(timezoneRepo)

	// Handler layer
	timezoneHandler := handler.NewTimezoneHandler(timezoneUseCase)

	// Router layer
	return NewTimezoneRouter(timezoneHandler, authMiddleware, authzMiddleware)
}
