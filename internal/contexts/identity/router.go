package identity

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	contactHandler "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/http/handler"
	contactRepo "github.com/basilex/promenade/internal/contexts/identity/contact/adapter/repository/postgres"
)

// Router manages routes for Identity context
type Router struct {
	contactHandler *contactHandler.ContactHandler
}

// NewRouter creates a new Identity router with all dependencies
func NewRouter(db *sqlx.DB) *Router {
	// Initialize Contact aggregate
	contactRepository := contactRepo.NewContactRepository(db)
	contactUseCase := contact.NewUseCase(contactRepository)
	handler := contactHandler.NewContactHandler(contactUseCase)

	return &Router{
		contactHandler: handler,
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

		// TODO: User routes will be added here
		// users := identity.Group("/users")
		// {
		//     users.POST("", userHandler.Create)
		//     users.GET("/:id", userHandler.GetByID)
		//     ...
		// }
	}
}
