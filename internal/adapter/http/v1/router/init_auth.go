package router

import (
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/bus"
	jwtpkg "github.com/basilex/promenade/pkg/jwt"
)

// InitAuthModule initializes the complete auth module with all dependencies
// This encapsulates the entire dependency chain: repo → usecase → handler → router
func InitAuthModule(
	db *sqlx.DB,
	jwtManager *jwtpkg.JWTManager,
	authMiddleware *middleware.AuthMiddleware,
	eventBus bus.Bus,
) *AuthRouter {
	// Repository layer
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)

	// Use case layer (теперь с event bus для асинхронных нотификаций)
	authUseCase := usecase.NewAuthUseCase(userRepo, sessionRepo, jwtManager, eventBus)

	// Handler layer
	authHandler := handler.NewAuthHandler(authUseCase)

	// Router layer
	return NewAuthRouter(authHandler, authMiddleware)
}
