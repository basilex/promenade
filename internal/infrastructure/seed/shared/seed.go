package shared

import (
	"context"
	"log/slog"

	"github.com/jmoiron/sqlx"

	countryRepo "github.com/basilex/promenade/internal/contexts/shared/country/adapter/repository/postgres"
	currencyRepo "github.com/basilex/promenade/internal/contexts/shared/currency/adapter/repository/postgres"
	languageRepo "github.com/basilex/promenade/internal/contexts/shared/language/adapter/repository/postgres"
	timezoneRepo "github.com/basilex/promenade/internal/contexts/shared/timezone/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/logger"
)

// SeedAll seeds all reference data for Shared context
func SeedAll(ctx context.Context, db *sqlx.DB) error {
	logger.FromContext(ctx).Info("Starting Shared context seeding...")

	// Create repositories
	countryRepository := countryRepo.NewRepository(db)
	currencyRepository := currencyRepo.NewRepository(db)
	languageRepository := languageRepo.NewRepository(db)
	timezoneRepository := timezoneRepo.NewRepository(db)

	// Seed countries
	if err := SeedCountries(ctx, countryRepository); err != nil {
		logger.FromContext(ctx).Error("Failed to seed countries", slog.Any("error", err))
		return err
	}

	// Seed currencies
	if err := SeedCurrencies(ctx, currencyRepository); err != nil {
		logger.FromContext(ctx).Error("Failed to seed currencies", slog.Any("error", err))
		return err
	}

	// Seed languages
	if err := SeedLanguages(ctx, languageRepository); err != nil {
		logger.FromContext(ctx).Error("Failed to seed languages", slog.Any("error", err))
		return err
	}

	// Seed timezones
	if err := SeedTimezones(ctx, timezoneRepository); err != nil {
		logger.FromContext(ctx).Error("Failed to seed timezones", slog.Any("error", err))
		return err
	}

	logger.FromContext(ctx).Info("Shared context seeding complete!")
	return nil
}
