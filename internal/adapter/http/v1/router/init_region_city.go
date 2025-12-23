package router

import (
	"github.com/basilex/promenade/internal/adapter/http/v1/handler"
	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/jmoiron/sqlx"
)

// InitRegionCityModules initializes both region and city modules and wires them together
func InitRegionCityModules(db *sqlx.DB, countryRouter *CountryRouter) (*RegionRouter, *CityRouter) {
	// Region module
	regionRepo := postgres.NewRegionRepository(db)
	regionUseCase := usecase.NewRegionUseCase(regionRepo)
	regionHandler := handler.NewRegionHandler(regionUseCase)
	regionRouter := NewRegionRouter(regionHandler)

	// City module
	cityRepo := postgres.NewCityRepository(db)
	cityUseCase := usecase.NewCityUseCase(cityRepo)
	cityHandler := handler.NewCityHandler(cityUseCase)
	cityRouter := NewCityRouter(cityHandler)

	// Wire cross-module relationships
	countryRouter.SetRegionHandler(regionHandler)
	countryRouter.SetCityHandler(cityHandler)
	regionRouter.SetCityHandler(cityHandler)

	return regionRouter, cityRouter
}
