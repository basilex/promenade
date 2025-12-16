package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/jmoiron/sqlx"
)

// InitCurrencyModule initializes the currency module with all dependencies
func InitCurrencyModule(db *sqlx.DB) *CurrencyRouter {
	// Repository layer
	currencyRepo := postgres.NewCurrencyRepository(db)
	countryRepo := postgres.NewCountryRepository(db)

	// Use case layer
	currencyUseCase := usecase.NewCurrencyUseCase(currencyRepo, countryRepo)

	// Handler layer
	currencyHandler := handler.NewCurrencyHandler(currencyUseCase)

	// Router
	return NewCurrencyRouter(currencyHandler)
}
