package ui

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/ui/metadata/form"
	formHTTP "github.com/basilex/promenade/internal/contexts/ui/metadata/form/adapter/http"
	formRepo "github.com/basilex/promenade/internal/contexts/ui/metadata/form/adapter/repository/postgres"
)

// Router handles routing for the UI context.
type Router struct {
	formHandler *formHTTP.FormHandler
}

// NewRouter creates a new UI router.
func NewRouter(db *sqlx.DB) *Router {
	formRepository := formRepo.NewFormRepository(db)
	formUseCase := form.NewUseCase(formRepository)
	formHandler := formHTTP.NewFormHandler(formUseCase)

	return &Router{formHandler: formHandler}
}

// RegisterRoutes registers UI context routes.
func (r *Router) RegisterRoutes(api *gin.RouterGroup) {
	ui := api.Group("/ui")
	{
		forms := ui.Group("/forms")
		{
			forms.POST("", r.formHandler.Create)
			forms.GET("", r.formHandler.List)
			forms.GET("/:id", r.formHandler.GetByID)
			forms.PUT("/:id", r.formHandler.Update)
			forms.DELETE("/:id", r.formHandler.Delete)
		}
	}
}
