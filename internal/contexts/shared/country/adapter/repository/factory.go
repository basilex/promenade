package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository/mssql"
	"github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository/postgres"
	countryRepository "github.com/basilex/promenade/internal/contexts/shared/country/repository"
)

// NewCountryRepository creates a country repository based on the database driver.
// Supported drivers: postgres, postgresql, mssql, sqlserver
func NewCountryRepository(db *sqlx.DB, driver string) (countryRepository.ICountryRepository, error) {
	switch driver {
	case "postgres", "postgresql":
		return postgres.NewRepository(db), nil
	case "mssql", "sqlserver":
		return mssql.NewRepository(db), nil
	default:
		return nil, fmt.Errorf("unsupported database driver for country repository: %s (supported: postgres, mssql)", driver)
	}
}
