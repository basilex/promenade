package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/gin-gonic/gin"
)

type CurrencyRouter struct {
	currencyHandler *handler.CurrencyHandler
}

func NewCurrencyRouter(currencyHandler *handler.CurrencyHandler) *CurrencyRouter {
	return &CurrencyRouter{
		currencyHandler: currencyHandler,
	}
}

func (r *CurrencyRouter) Setup(rg *gin.RouterGroup) {
	currencies := rg.Group("/currencies")
	{
		currencies.POST("", r.currencyHandler.Create)
		currencies.GET("", r.currencyHandler.List)
		currencies.GET("/:id", r.currencyHandler.GetByID)
		currencies.GET("/code/:code", r.currencyHandler.GetByCode)
		currencies.PUT("/:id", r.currencyHandler.Update)
		currencies.DELETE("/:id", r.currencyHandler.Delete)

		// Country relationships
		currencies.GET("/:id/countries", r.currencyHandler.GetCountries)
		currencies.POST("/:id/countries", r.currencyHandler.AddCountry)
		currencies.DELETE("/:id/countries/:country_id", r.currencyHandler.RemoveCountry)
	}
}
