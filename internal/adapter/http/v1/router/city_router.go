package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

// CityRouter handles city-related routes
type CityRouter struct {
	handler *handler.CityHandler
}

// NewCityRouter creates a new city router
func NewCityRouter(handler *handler.CityHandler) *CityRouter {
	return &CityRouter{
		handler: handler,
	}
}

// RegisterRoutes registers all city routes
func (r *CityRouter) RegisterRoutes(rg *gin.RouterGroup) {
	cities := rg.Group("/cities")
	{
		cities.POST("", r.handler.Create)
		cities.GET("/:id", r.handler.GetByID)
		cities.GET("", r.handler.List)
		cities.GET("/capitals", r.handler.ListCapitals)
		cities.GET("/search", r.handler.Search)
		cities.PUT("/:id", r.handler.Update)
		cities.DELETE("/:id", r.handler.Delete)
	}
}
