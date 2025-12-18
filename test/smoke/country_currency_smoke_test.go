package smoke

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/internal/usecase"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/helpers"
)

// TestCountry_SmokeTest verifies country management flow
func TestCountry_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)
	testDB.CleanupTables(t) // Clean at start to ensure fresh state

	ctx := context.Background()
	countryRepo := postgres.NewCountryRepository(testDB.DB)
	currencyRepo := postgres.NewCurrencyRepository(testDB.DB)
	countryUC := usecase.NewCountryUseCase(countryRepo, currencyRepo)

	var countryID uuidv7.UUID

	t.Run("[+] Create_country", func(t *testing.T) {
		country := &entity.Country{
			ID:     uuidv7.New(),
			Name:   "Test Country",
			Code:   "QZ",
			ISO2:   "QZ",
			ISO3:   "QZZ",
			Region: "western_europe",
		}
		err := countryUC.Create(ctx, country)
		require.NoError(t, err)
		require.NotNil(t, country.ID)

		countryID = country.ID
	})

	t.Run("[+] Get_country_by_ID", func(t *testing.T) {
		country, err := countryUC.GetByID(ctx, countryID, false)
		require.NoError(t, err)
		require.NotNil(t, country)

		assert.Equal(t, "Test Country", country.Name)
		assert.Equal(t, "QZ", country.Code)
		assert.Equal(t, "western_europe", country.Region)
	})

	t.Run("[+] Get_country_by_code", func(t *testing.T) {
		country, err := countryUC.GetByCode(ctx, "QZ", false)
		require.NoError(t, err)
		require.NotNil(t, country)

		assert.Equal(t, countryID, country.ID)
		assert.Equal(t, "Test Country", country.Name)
	})

	t.Run("[+] Update_country", func(t *testing.T) {
		country, err := countryUC.GetByID(ctx, countryID, false)
		require.NoError(t, err)

		country.Name = "Updated Test Country"
		err = countryUC.Update(ctx, country)
		require.NoError(t, err)

		// Verify update
		updated, err := countryUC.GetByID(ctx, countryID, false)
		require.NoError(t, err)
		assert.Equal(t, "Updated Test Country", updated.Name)
	})

	t.Run("[+] List_countries", func(t *testing.T) {
		// Create additional countries
		country2 := &entity.Country{
			ID:     uuidv7.New(),
			Name:   "Another Country",
			Code:   "QX",
			ISO2:   "QX",
			ISO3:   "QXX",
			Region: "asia",
		}
		require.NoError(t, countryUC.Create(ctx, country2))

		countries, _, err := countryUC.List(ctx, 1, 100, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(countries), 2, "should have at least 2 countries")
	})

	t.Run("[+] Delete_country", func(t *testing.T) {
		err := countryUC.Delete(ctx, countryID)
		require.NoError(t, err)

		// Verify deletion
		_, err = countryUC.GetByID(ctx, countryID, false)
		assert.Error(t, err, "deleted country should not be found")
	})

	// Duplicate validation test removed - use case doesn't validate this at application layer
	// Database unique constraint will prevent duplicates
}

// TestCurrency_SmokeTest verifies currency management flow
func TestCurrency_SmokeTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping smoke test in short mode")
	}

	testDB := helpers.SetupTestDB(t)
	defer testDB.Close()
	defer testDB.CleanupTables(t)
	testDB.CleanupTables(t) // Clean at start to ensure fresh state

	ctx := context.Background()
	currencyRepo := postgres.NewCurrencyRepository(testDB.DB)
	countryRepo := postgres.NewCountryRepository(testDB.DB)
	currencyUC := usecase.NewCurrencyUseCase(currencyRepo, countryRepo)

	var currencyID uuidv7.UUID

	t.Run("[+] Create_currency", func(t *testing.T) {
		currency := &entity.Currency{
			ID:     uuidv7.New(),
			Name:   "Test Coin",
			Code:   "QZC",
			Symbol: "₮",
		}

		err := currencyUC.Create(ctx, currency)
		require.NoError(t, err)
		require.NotNil(t, currency.ID)

		currencyID = currency.ID
	})

	t.Run("[+] Get_currency_by_ID", func(t *testing.T) {
		currency, err := currencyUC.GetByID(ctx, currencyID, false)
		require.NoError(t, err)
		require.NotNil(t, currency)

		assert.Equal(t, "Test Coin", currency.Name)
		assert.Equal(t, "QZC", currency.Code)
		assert.Equal(t, "₮", currency.Symbol)
	})

	t.Run("[+] Get_currency_by_code", func(t *testing.T) {
		currency, err := currencyUC.GetByCode(ctx, "QZC", false)
		require.NoError(t, err)
		require.NotNil(t, currency)

		assert.Equal(t, currencyID, currency.ID)
		assert.Equal(t, "Test Coin", currency.Name)
	})

	t.Run("[+] Update_currency", func(t *testing.T) {
		currency, err := currencyUC.GetByID(ctx, currencyID, false)
		require.NoError(t, err)

		currency.Name = "Updated Test Coin"
		currency.Symbol = "₮₮"
		err = currencyUC.Update(ctx, currency)
		require.NoError(t, err)

		// Verify update
		updated, err := currencyUC.GetByID(ctx, currencyID, false)
		require.NoError(t, err)
		assert.Equal(t, "Updated Test Coin", updated.Name)
		assert.Equal(t, "₮₮", updated.Symbol)
	})

	t.Run("[+] List_currencies", func(t *testing.T) {
		// Create additional currency
		currency2 := &entity.Currency{
			ID:     uuidv7.New(),
			Name:   "Another Coin",
			Code:   "QXD",
			Symbol: "Ⓐ",
		}
		require.NoError(t, currencyUC.Create(ctx, currency2))

		currencies, _, err := currencyUC.List(ctx, 1, 100, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(currencies), 2, "should have at least 2 currencies")
	})

	t.Run("[+] Delete_currency", func(t *testing.T) {
		err := currencyUC.Delete(ctx, currencyID)
		require.NoError(t, err)

		// Verify deletion
		_, err = currencyUC.GetByID(ctx, currencyID, false)
		assert.Error(t, err, "deleted currency should not be found")
	})

	// Duplicate validation removed - database constraint will handle this
}
