package shared

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/shared/handler"
	"github.com/basilex/promenade/pkg/reference"
)

// Router handles HTTP routing for Shared Context (Reference Data)
type Router struct {
	countryHandler  *handler.CountryHandler
	currencyHandler *handler.CurrencyHandler
	languageHandler *handler.LanguageHandler
	timezoneHandler *handler.TimezoneHandler
}

// NewRouter creates a new Shared Context router
func NewRouter(db *sqlx.DB) *Router {
	// Initialize repository (read-only)
	repo := reference.NewRepository(db)

	return &Router{
		countryHandler:  handler.NewCountryHandler(repo),
		currencyHandler: handler.NewCurrencyHandler(repo),
		languageHandler: handler.NewLanguageHandler(repo),
		timezoneHandler: handler.NewTimezoneHandler(repo),
	}
}

// RegisterRoutes registers all Shared Context routes
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	// Countries
	countries := rg.Group("/countries")
	{
		countries.GET("", r.countryHandler.ListCountries)
		countries.GET("/:code", r.countryHandler.GetCountryByCode)
	}

	// Currencies
	currencies := rg.Group("/currencies")
	{
		currencies.GET("", r.currencyHandler.ListCurrencies)
		currencies.GET("/:code", r.currencyHandler.GetCurrencyByCode)
	}

	// Languages
	languages := rg.Group("/languages")
	{
		languages.GET("", r.languageHandler.ListLanguages)
		languages.GET("/:code", r.languageHandler.GetLanguageByCode)
	}

	// Timezones
	timezones := rg.Group("/timezones")
	{
		timezones.GET("", r.timezoneHandler.ListTimezones)
		timezones.GET("/:name", r.timezoneHandler.GetTimezoneByName)
	}
}
