package router

import (
	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/infrastructure/scheduler"
	"github.com/basilex/promenade/internal/usecase"
)

// InitAdminModule initializes the admin module with all dependencies
func InitAdminModule(
	purgeUseCase usecase.IPurgeUseCase,
	scheduler *scheduler.Scheduler,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *AdminRouter {
	// Initialize handler
	purgeHandler := handler.NewAdminPurgeHandler(purgeUseCase, scheduler)

	// Initialize router
	return NewAdminRouter(purgeHandler, authMiddleware, authzMiddleware)
}
