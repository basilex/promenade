package router

import (
	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/adapter/http/shared/middleware"
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
)

// RBACRouter handles all RBAC-related routes (roles, permissions, assignments)
type RBACRouter struct {
	roleHandler       *handler.RoleHandler
	permissionHandler *handler.PermissionHandler
	authMiddleware    *middleware.AuthMiddleware
	authzMiddleware   *middleware.AuthorizationMiddleware
}

// NewRBACRouter creates a new RBAC router instance
func NewRBACRouter(
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
	authMiddleware *middleware.AuthMiddleware,
	authzMiddleware *middleware.AuthorizationMiddleware,
) *RBACRouter {
	return &RBACRouter{
		roleHandler:       roleHandler,
		permissionHandler: permissionHandler,
		authMiddleware:    authMiddleware,
		authzMiddleware:   authzMiddleware,
	}
}

// Setup registers all RBAC routes
func (r *RBACRouter) Setup(rg *gin.RouterGroup) {
	// All RBAC routes require authentication
	rbac := rg.Group("")
	rbac.Use(r.authMiddleware.RequireAuth())

	// ============================================
	// ROLE MANAGEMENT ROUTES
	// ============================================
	roles := rbac.Group("/roles")
	{
		// List roles - any authenticated user can view roles
		roles.GET("", r.roleHandler.ListRoles)
		roles.GET("/:id", r.roleHandler.GetRole)

		// Role management - requires roles:manage permission
		roles.POST("", r.authzMiddleware.RequirePermission("roles:create"), r.roleHandler.CreateRole)
		roles.PUT("/:id", r.authzMiddleware.RequirePermission("roles:update"), r.roleHandler.UpdateRole)
		roles.DELETE("/:id", r.authzMiddleware.RequirePermission("roles:delete"), r.roleHandler.DeleteRole)
		roles.PUT("/:id/permissions", r.authzMiddleware.RequirePermission("roles:update"), r.roleHandler.SyncRolePermissions)

		// Role assignment - requires roles:assign permission
		roles.POST("/assign", r.authzMiddleware.RequirePermission("roles:assign"), r.roleHandler.AssignRoleToUser)
		roles.POST("/remove", r.authzMiddleware.RequirePermission("roles:assign"), r.roleHandler.RemoveRoleFromUser)
	}

	// ============================================
	// PERMISSION MANAGEMENT ROUTES
	// ============================================
	permissions := rbac.Group("/permissions")
	{
		// List permissions - any authenticated user can view permissions
		permissions.GET("", r.permissionHandler.ListPermissions)
		permissions.GET("/:id", r.permissionHandler.GetPermission)
		permissions.GET("/search", r.permissionHandler.FindPermissionsByResource)

		// Permission management - requires permissions:manage permission
		permissions.POST("", r.authzMiddleware.RequirePermission("permissions:create"), r.permissionHandler.CreatePermission)
		permissions.PUT("/:id", r.authzMiddleware.RequirePermission("permissions:update"), r.permissionHandler.UpdatePermission)
		permissions.DELETE("/:id", r.authzMiddleware.RequirePermission("permissions:delete"), r.permissionHandler.DeletePermission)
	}

	// ============================================
	// USER ROLE/PERMISSION QUERY ROUTES
	// ============================================
	users := rbac.Group("/users")
	{
		// Get user roles and permissions
		// Users can view their own, admins can view anyone's
		users.GET("/:user_id/roles", r.roleHandler.GetUserRoles)
		users.GET("/:user_id/permissions", r.roleHandler.GetUserPermissions)
		users.POST("/:user_id/check-permission", r.roleHandler.CheckPermission)
	}
}
