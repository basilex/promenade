package identity

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	contactHTTP "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/http"
	contactRepo "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/identity/profile"
	profileHTTP "github.com/basilex/promenade/internal/contexts/identity/profile/adapter/http"
	profileRepo "github.com/basilex/promenade/internal/contexts/identity/profile/adapter/repository/postgres"
)

// Router manages routes for Identity context
type Router struct {
	contactHandler *contactHTTP.ContactHandler
	profileHandler *profileHTTP.ProfileHandler
}

// NewRouter creates a new Identity router with all dependencies
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Contact aggregate
	contactRepository := contactRepo.NewContactRepository(db)
	contactUseCase := contact.NewUseCase(contactRepository)
	contactHandler := contactHTTP.NewContactHandler(contactUseCase)

	// Initialize Profile aggregate
	profileRepository := profileRepo.NewProfileRepository(db)
	profileUseCase := profile.NewUseCase(profileRepository)
	profileHandler := profileHTTP.NewProfileHandler(profileUseCase)

	return &Router{
		contactHandler: contactHandler,
		profileHandler: profileHandler,
	}
}

// RegisterRoutes registers all Identity context routes
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	// Identity group
	identity := api.Group("/identity")
	{
		// Contact routes
		contacts := identity.Group("/contacts")
		{
			contacts.POST("", r.contactHandler.Create)                // Create new contact
			contacts.GET("", r.contactHandler.List)                   // List contacts (filter by user_id, type)
			contacts.GET("/:id", r.contactHandler.GetByID)            // Get contact by ID
			contacts.PUT("/:id", r.contactHandler.Update)             // Update contact
			contacts.DELETE("/:id", r.contactHandler.Delete)          // Delete contact
			contacts.PUT("/:id/verify", r.contactHandler.Verify)      // Verify contact
			contacts.PUT("/:id/primary", r.contactHandler.SetPrimary) // Set as primary
		}

		// Profile routes
		profiles := identity.Group("/profiles")
		{
			profiles.POST("", r.profileHandler.Create)                              // Create new profile
			profiles.GET("/:id", r.profileHandler.GetByID)                          // Get profile by ID
			profiles.GET("/user/:user_id", r.profileHandler.GetByUserID)            // Get profile by user ID
			profiles.DELETE("/:id", r.profileHandler.Delete)                        // Delete profile
			profiles.GET("", r.profileHandler.ListPublic)                           // List public profiles
			profiles.PUT("/:id/display-name", r.profileHandler.UpdateDisplayName)   // Update display name
			profiles.PUT("/:id/bio", r.profileHandler.UpdateBio)                    // Update bio
			profiles.PUT("/:id/avatar", r.profileHandler.UpdateAvatar)              // Update avatar
			profiles.PUT("/:id/personal-info", r.profileHandler.UpdatePersonalInfo) // Update personal info
			profiles.PUT("/:id/gender", r.profileHandler.UpdateGender)              // Update gender
			profiles.PUT("/:id/date-of-birth", r.profileHandler.UpdateDateOfBirth)  // Update date of birth
			profiles.PUT("/:id/localization", r.profileHandler.UpdateLocalization)  // Update localization
			profiles.PUT("/:id/social-links", r.profileHandler.UpdateSocialLinks)   // Update social links
			profiles.PUT("/:id/public", r.profileHandler.SetPublic)                 // Set profile public
			profiles.PUT("/:id/private", r.profileHandler.SetPrivate)               // Set profile private
		}

		// TODO: User routes will be added here
		// users := identity.Group("/users")
		// {
		//     users.POST("", userHandler.Create)
		//     users.GET("/:id", userHandler.GetByID)
		//     ...
		// }
	}
}
