package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

// RegionRouter handles region-related routes
type RegionRouter struct {
	handler     *handler.RegionHandler
	cityHandler *handler.CityHandler
}

// NewRegionRouter creates a new region router
func NewRegionRouter(handler *handler.RegionHandler) *RegionRouter {
	return &RegionRouter{
		handler: handler,
	}
}

// SetCityHandler sets the city handler for region-related city routes
func (r *RegionRouter) SetCityHandler(cityHandler *handler.CityHandler) {
	r.cityHandler = cityHandler
}

// RegisterRoutes registers all region routes
func (r *RegionRouter) RegisterRoutes(rg *gin.RouterGroup) {
	regions := rg.Group("/regions")
	{
		regions.POST("", r.handler.Create)
		regions.GET("/:id", r.handler.GetByID)
		regions.GET("", r.handler.List)
		regions.GET("/active", r.handler.ListActive)
		regions.PUT("/:id", r.handler.Update)
		regions.DELETE("/:id", r.handler.Delete)

		// City relationships (if handler is set)
		if r.cityHandler != nil {
			regions.GET("/:id/cities", r.cityHandler.ListByRegion)
		}
	}
}
