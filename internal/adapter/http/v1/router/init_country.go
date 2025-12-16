package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/jmoiron/sqlx"
)

// InitCountryModule initializes the country module with all dependencies
func InitCountryModule(db *sqlx.DB) *CountryRouter {
	// Repository layer
	countryRepo := postgres.NewCountryRepository(db)
	currencyRepo := postgres.NewCurrencyRepository(db)

	// Use case layer
	countryUseCase := usecase.NewCountryUseCase(countryRepo, currencyRepo)

	// Handler layer
	countryHandler := handler.NewCountryHandler(countryUseCase)

	// Router
	return NewCountryRouter(countryHandler)
}
