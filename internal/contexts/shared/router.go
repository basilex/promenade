package shared

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/shared/country"
	countryHTTP "github.com/basilex/promenade/internal/contexts/shared/country/adapter/http"
	countryPostgres "github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository/postgres"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	currencyHTTP "github.com/basilex/promenade/internal/contexts/shared/currency/adapter/http"
	currencyPostgres "github.com/basilex/promenade/internal/contexts/shared/currency/adapter/repository/postgres"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	languageHTTP "github.com/basilex/promenade/internal/contexts/shared/language/adapter/http"
	languagePostgres "github.com/basilex/promenade/internal/contexts/shared/language/adapter/repository/postgres"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	timezoneHTTP "github.com/basilex/promenade/internal/contexts/shared/timezone/adapter/http"
	timezonePostgres "github.com/basilex/promenade/internal/contexts/shared/timezone/adapter/repository/postgres"

	"github.com/basilex/promenade/pkg/cache"
)

// Router handles HTTP routing for Shared Context (Reference Data)
type Router struct {
	countryHandler  *countryHTTP.Handler
	currencyHandler *currencyHTTP.Handler
	languageHandler *languageHTTP.Handler
	timezoneHandler *timezoneHTTP.Handler
}

// NewRouter creates a new Shared Context router with Clean Architecture layers
func NewRouter(db *sqlx.DB, cacheClient cache.Cache) *Router {
	// Country aggregate
	countryRepo := countryPostgres.NewRepository(db)
	countryUC := country.NewUseCase(countryRepo, cacheClient)
	countryHandler := countryHTTP.NewHandler(countryUC)

	// Currency aggregate
	currencyRepo := currencyPostgres.NewRepository(db)
	currencyUC := currency.NewUseCase(currencyRepo, cacheClient)
	currencyHandler := currencyHTTP.NewHandler(currencyUC)

	// Language aggregate
	languageRepo := languagePostgres.NewRepository(db)
	languageUC := language.NewUseCase(languageRepo, cacheClient)
	languageHandler := languageHTTP.NewHandler(languageUC)

	// Timezone aggregate
	timezoneRepo := timezonePostgres.NewRepository(db)
	timezoneUC := timezone.NewUseCase(timezoneRepo, cacheClient)
	timezoneHandler := timezoneHTTP.NewHandler(timezoneUC)

	return &Router{
		countryHandler:  countryHandler,
		currencyHandler: currencyHandler,
		languageHandler: languageHandler,
		timezoneHandler: timezoneHandler,
	}
}

// RegisterRoutes registers all Shared Context routes
func (r *Router) RegisterRoutes(rg *gin.RouterGroup) {
	// Countries
	countries := rg.Group("/countries")
	{
		countries.GET("", r.countryHandler.ListCountries)
		countries.POST("", r.countryHandler.CreateCountry)
		countries.GET("/:code", r.countryHandler.GetCountryByCode)
		countries.PUT("/:id", r.countryHandler.UpdateCountry)
		countries.DELETE("/:id", r.countryHandler.DeleteCountry)
	}

	// Currencies
	currencies := rg.Group("/currencies")
	{
		currencies.GET("", r.currencyHandler.ListCurrencies)
		currencies.POST("", r.currencyHandler.CreateCurrency)
		currencies.GET("/:code", r.currencyHandler.GetCurrencyByCode)
		currencies.PUT("/:id", r.currencyHandler.UpdateCurrency)
		currencies.DELETE("/:id", r.currencyHandler.DeleteCurrency)
	}

	// Languages
	languages := rg.Group("/languages")
	{
		languages.GET("", r.languageHandler.ListLanguages)
		languages.POST("", r.languageHandler.CreateLanguage)
		languages.GET("/:code", r.languageHandler.GetLanguageByCode)
		languages.PUT("/:id", r.languageHandler.UpdateLanguage)
		languages.DELETE("/:id", r.languageHandler.DeleteLanguage)
	}

	// Timezones
	timezones := rg.Group("/timezones")
	{
		timezones.GET("", r.timezoneHandler.ListTimezones)
		timezones.POST("", r.timezoneHandler.CreateTimezone)
		timezones.GET("/*name", r.timezoneHandler.GetTimezoneByName) // * to capture full path (e.g., Europe/Kyiv)
		timezones.PUT("/:id", r.timezoneHandler.UpdateTimezone)
		timezones.DELETE("/:id", r.timezoneHandler.DeleteTimezone)
	}
}

