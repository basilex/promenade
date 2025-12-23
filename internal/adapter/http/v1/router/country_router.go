package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

type CountryRouter struct {
	countryHandler *handler.CountryHandler
	regionHandler  *handler.RegionHandler
	cityHandler    *handler.CityHandler
}

func NewCountryRouter(countryHandler *handler.CountryHandler) *CountryRouter {
	return &CountryRouter{
		countryHandler: countryHandler,
	}
}

// SetRegionHandler sets the region handler for country-related region routes
func (r *CountryRouter) SetRegionHandler(regionHandler *handler.RegionHandler) {
	r.regionHandler = regionHandler
}

// SetCityHandler sets the city handler for country-related city routes
func (r *CountryRouter) SetCityHandler(cityHandler *handler.CityHandler) {
	r.cityHandler = cityHandler
}

func (r *CountryRouter) Setup(rg *gin.RouterGroup) {
	countries := rg.Group("/countries")
	{
		countries.POST("", r.countryHandler.Create)
		countries.GET("", r.countryHandler.List)
		countries.GET("/:id", r.countryHandler.GetByID)
		countries.GET("/code/:code", r.countryHandler.GetByCode)
		countries.PUT("/:id", r.countryHandler.Update)
		countries.DELETE("/:id", r.countryHandler.Delete)

		// Currency relationships
		countries.GET("/:id/currencies", r.countryHandler.GetCurrencies)
		countries.POST("/:id/currencies", r.countryHandler.AddCurrency)
		countries.DELETE("/:id/currencies/:currency_id", r.countryHandler.RemoveCurrency)

		// Region relationships (if handler is set)
		if r.regionHandler != nil {
			countries.GET("/:id/regions", r.regionHandler.ListByCountry)
		}

		// City relationships (if handler is set)
		if r.cityHandler != nil {
			countries.GET("/:id/cities", r.cityHandler.ListByCountry)
		}
	}
}
