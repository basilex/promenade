package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/basilex/promenade/pkg/logger"

	// Shared Context - Country
	countryRepoPostgres "github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository/postgres"
	countryRepoMssql "github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository/mssql"
	countryRepository "github.com/basilex/promenade/internal/contexts/shared/country/repository"
)

// newCountryRepository creates country repository based on database driver
func newCountryRepository(db *sqlx.DB, driver string) (countryRepository.ICountryRepository, error) {
	switch driver {
	case "postgres", "postgresql":
		logger.Debug("Creating PostgreSQL country repository")
		return countryRepoPostgres.NewRepository(db), nil
		
	case "mssql", "sqlserver":
		logger.Debug("Creating MS SQL Server country repository")
		return countryRepoMssql.NewRepository(db), nil
		
	default:
		return nil, fmt.Errorf("unsupported database driver for country repository: %s (supported: postgres, mssql)", driver)
	}
}

// TODO: Add more repository factories here as we implement multi-database support for other contexts
// Example pattern:
//
// func newCustomerRepository(db *sqlx.DB, driver string) (customerRepository.ICustomerRepository, error) {
//     switch driver {
//     case "postgres", "postgresql":
//         return customerRepoPostgres.NewRepository(db), nil
//     case "mssql", "sqlserver":
//         return customerRepoMssql.NewRepository(db), nil
//     default:
//         return nil, fmt.Errorf("unsupported database driver: %s", driver)
//     }
// }
