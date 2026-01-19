package identity

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"golang.org/x/time/rate"

	contactHTTP "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/http"
	contactRepo "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/identity/contact/usecase"
	permissionHTTP "github.com/basilex/promenade/internal/contexts/identity/permission/adapter/http"
	permissionRepo "github.com/basilex/promenade/internal/contexts/identity/permission/adapter/repository/postgres"
	permissionusecase "github.com/basilex/promenade/internal/contexts/identity/permission/usecase"
	profileHTTP "github.com/basilex/promenade/internal/contexts/identity/profile/adapter/http"
	profileRepo "github.com/basilex/promenade/internal/contexts/identity/profile/adapter/repository/postgres"
	profileusecase "github.com/basilex/promenade/internal/contexts/identity/profile/usecase"
	roleHTTP "github.com/basilex/promenade/internal/contexts/identity/role/adapter/http"
	roleRepo "github.com/basilex/promenade/internal/contexts/identity/role/adapter/repository/postgres"
	roleusecase "github.com/basilex/promenade/internal/contexts/identity/role/usecase"
	userHTTP "github.com/basilex/promenade/internal/contexts/identity/user/adapter/http"
	userRepo "github.com/basilex/promenade/internal/contexts/identity/user/adapter/repository/postgres"
	userusecase "github.com/basilex/promenade/internal/contexts/identity/user/usecase"
	"github.com/basilex/promenade/pkg/jwt"
	"github.com/basilex/promenade/pkg/middleware"
)

// Router manages routes for Identity context
type Router struct {
	contactHandler    *contactHTTP.ContactHandler
	profileHandler    *profileHTTP.ProfileHandler
	userHandler       *userHTTP.UserHandler
	roleHandler       *roleHTTP.RoleHandler
	permissionHandler *permissionHTTP.PermissionHandler
	tokenRevoker      *jwt.TokenRevoker
	jwtManager        *jwt.Manager
}

// NewRouter creates a new Identity router with all dependencies
func NewRouter(db *sqlx.DB, jwtManager *jwt.Manager, tokenRevoker *jwt.TokenRevoker) *Router {
	// Initialize Contact aggregate
	contactRepository := contactRepo.NewContactRepository(db)
	contactUseCase := usecase.NewContactUseCase(contactRepository)
	contactHandler := contactHTTP.NewContactHandler(contactUseCase)

	// Initialize Profile aggregate
	profileRepository := profileRepo.NewProfileRepository(db)
	profileUseCase := profileusecase.NewProfileUseCase(profileRepository)
	profileHandler := profileHTTP.NewProfileHandler(profileUseCase)

	// Initialize Role aggregate
	roleRepository := roleRepo.NewRoleRepository(db)
	roleUseCase := roleusecase.NewRoleUseCase(roleRepository)
	roleHandler := roleHTTP.NewRoleHandler(roleUseCase)

	// Initialize Permission aggregate
	permissionRepository := permissionRepo.NewPermissionRepository(db)
	permissionUseCase := permissionusecase.NewPermissionUseCase(permissionRepository)
	permissionHandler := permissionHTTP.NewPermissionHandler(permissionUseCase)

	// Initialize User aggregate (with role repository and token revoker)
	userRepository := userRepo.NewUserRepository(db)
	userUseCase := userusecase.NewUserUseCase(userRepository, roleRepository)
	userHandler := userHTTP.NewUserHandler(userUseCase, jwtManager, tokenRevoker)

	return &Router{
		contactHandler:    contactHandler,
		profileHandler:    profileHandler,
		userHandler:       userHandler,
		roleHandler:       roleHandler,
		permissionHandler: permissionHandler,
		jwtManager:        jwtManager,
		tokenRevoker:      tokenRevoker,
	}
}

// RegisterRoutes registers all Identity context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	// Identity group
	identity := api.Group("/identity")
	{
		// Contact routes (protected with JWT + token revocation check)
		contacts := identity.Group("/contacts")
		contacts.Use(jwt.AuthMiddleware(r.jwtManager, r.tokenRevoker))
		{
			contacts.POST("", r.contactHandler.Create)                // Create new contact
			contacts.GET("", r.contactHandler.List)                   // List contacts (filter by user_id, type)
			contacts.GET("/:id", r.contactHandler.GetByID)            // Get contact by ID
			contacts.PUT("/:id", r.contactHandler.Update)             // Update contact
			contacts.DELETE("/:id", r.contactHandler.Delete)          // Delete contact
			contacts.PUT("/:id/verify", r.contactHandler.Verify)      // Verify contact
			contacts.PUT("/:id/primary", r.contactHandler.SetPrimary) // Set as primary
		}

		// Profile routes (mixed: public + protected)
		profiles := identity.Group("/profiles")
		{
			// Public routes (no JWT required)
			profiles.GET("", r.profileHandler.ListPublic)                // List public profiles
			profiles.GET("/:id", r.profileHandler.GetByID)               // Get profile by ID (public profiles)
			profiles.GET("/user/:user_id", r.profileHandler.GetByUserID) // Get profile by user ID

			// Protected routes (JWT required + token revocation check)
			protected := profiles.Group("")
			protected.Use(jwt.AuthMiddleware(r.jwtManager, r.tokenRevoker))
			{
				protected.POST("", r.profileHandler.Create)                              // Create new profile
				protected.DELETE("/:id", r.profileHandler.Delete)                        // Delete profile
				protected.PUT("/:id/display-name", r.profileHandler.UpdateDisplayName)   // Update display name
				protected.PUT("/:id/bio", r.profileHandler.UpdateBio)                    // Update bio
				protected.PUT("/:id/avatar", r.profileHandler.UpdateAvatar)              // Update avatar
				protected.PUT("/:id/personal-info", r.profileHandler.UpdatePersonalInfo) // Update personal info
				protected.PUT("/:id/gender", r.profileHandler.UpdateGender)              // Update gender
				protected.PUT("/:id/date-of-birth", r.profileHandler.UpdateDateOfBirth)  // Update date of birth
				protected.PUT("/:id/localization", r.profileHandler.UpdateLocalization)  // Update localization
				protected.PUT("/:id/social-links", r.profileHandler.UpdateSocialLinks)   // Update social links
				protected.PUT("/:id/public", r.profileHandler.SetPublic)                 // Set profile public
				protected.PUT("/:id/private", r.profileHandler.SetPrivate)               // Set profile private
			}
		}

		// User routes
		users := identity.Group("/users")
		{
			// Rate limiters
			loginLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/5), 1)    // 5 attempts per minute
			registerLimiter := middleware.NewRateLimiter(rate.Every(time.Minute/3), 1) // 3 attempts per minute

			// Public routes with rate limiting
			users.POST("/register", registerLimiter.Limit(), r.userHandler.Register) // Register new user (rate limited)
			users.POST("/login", loginLimiter.Limit(), r.userHandler.Login)          // Authenticate user (rate limited)

			// Protected routes (authentication required + token revocation check)
			protected := users.Group("")
			protected.Use(jwt.AuthMiddleware(r.jwtManager, r.tokenRevoker))
			{
				protected.GET("/:id", r.userHandler.GetByID)                         // Get user by ID
				protected.GET("/email/:email", r.userHandler.GetByEmail)             // Get user by email
				protected.POST("/:id/verify-email", r.userHandler.VerifyEmail)       // Verify user email
				protected.POST("/:id/change-password", r.userHandler.ChangePassword) // Change password
				protected.POST("/:id/suspend", r.userHandler.Suspend)                // Suspend user
				protected.POST("/:id/ban", r.userHandler.Ban)                        // Ban user
				protected.POST("/:id/activate", r.userHandler.Activate)              // Activate user
				protected.POST("/:id/unlock", r.userHandler.Unlock)                  // Unlock user account
				protected.GET("", r.userHandler.List)                                // List users with pagination
			}
		}

		// Auth routes
		auth := identity.Group("/auth")
		{
			auth.POST("/refresh", r.userHandler.RefreshToken) // Refresh access token

			// Protected auth routes (token revocation check)
			authProtected := auth.Group("")
			authProtected.Use(jwt.AuthMiddleware(r.jwtManager, r.tokenRevoker))
			{
				authProtected.POST("/revoke", r.userHandler.RevokeToken) // Revoke current token (logout)
			}
		}

		// Role routes (admin only + token revocation check)
		roles := identity.Group("/roles")
		roles.Use(jwt.AuthMiddleware(r.jwtManager, r.tokenRevoker))
		// TODO: Add admin authorization middleware here
		{
			roles.POST("", r.roleHandler.Create)                    // Create new role
			roles.GET("", r.roleHandler.List)                       // List roles with pagination
			roles.GET("/:id", r.roleHandler.GetByID)                // Get role by ID
			roles.GET("/name/:name", r.roleHandler.GetByName)       // Get role by name
			roles.PUT("/:id", r.roleHandler.Update)                 // Update role
			roles.DELETE("/:id", r.roleHandler.Delete)              // Delete role
			roles.GET("/user/:user_id", r.roleHandler.GetUserRoles) // Get user's roles
		}

		// Permission routes (admin only + token revocation check)
		permissions := identity.Group("/permissions")
		permissions.Use(jwt.AuthMiddleware(r.jwtManager, r.tokenRevoker))
		// TODO: Add admin authorization middleware here
		{
			permissions.POST("", r.permissionHandler.Create)                          // Create new permission
			permissions.GET("", r.permissionHandler.List)                             // List permissions with pagination
			permissions.GET("/:id", r.permissionHandler.GetByID)                      // Get permission by ID
			permissions.GET("/name/:name", r.permissionHandler.GetByName)             // Get permission by name
			permissions.PUT("/:id", r.permissionHandler.Update)                       // Update permission
			permissions.DELETE("/:id", r.permissionHandler.Delete)                    // Delete permission
			permissions.GET("/role/:role_id", r.permissionHandler.GetRolePermissions) // Get role's permissions
		}
	}
}
