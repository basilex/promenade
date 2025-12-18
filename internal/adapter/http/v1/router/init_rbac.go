package router

import (
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
)

// InitRBACModule initializes the complete RBAC module with all dependencies
// This encapsulates the entire dependency chain: repo → usecase → handler → router
func InitRBACModule(
	db *sqlx.DB,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *RBACRouter {
	// Repository layer
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)

	// Use case layer
	roleUseCase := usecase.NewRoleUseCase(roleRepo, permissionRepo)
	permissionUseCase := usecase.NewPermissionUseCase(permissionRepo)

	// Handler layer
	roleHandler := handler.NewRoleHandler(roleUseCase)
	permissionHandler := handler.NewPermissionHandler(permissionUseCase)

	// Router layer
	return NewRBACRouter(roleHandler, permissionHandler, authMiddleware, authzMiddleware)
}
