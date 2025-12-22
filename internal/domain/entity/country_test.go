package entity

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
)

func TestCountry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		country *Country
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid country",
			country: &Country{
				ID:        uuidv7.New(),
				Name:      "United States",
				Code:      "USA",
				ISO2:      "US",
				ISO3:      "USA",
				Region:    "north_america",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			country: &Country{
				ID:     uuidv7.New(),
				Code:   "USA",
				ISO2:   "US",
				ISO3:   "USA",
				Region: "north_america",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too short",
			country: &Country{
				ID:     uuidv7.New(),
				Name:   "A",
				Code:   "USA",
				ISO2:   "US",
				ISO3:   "USA",
				Region: "north_america",
			},
			wantErr: true,
			errMsg:  "name must be between 2 and 100 characters",
		},
		{
			name: "missing code",
			country: &Country{
				ID:     uuidv7.New(),
				Name:   "United States",
				ISO2:   "US",
				ISO3:   "USA",
				Region: "north_america",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "invalid ISO2 length",
			country: &Country{
				ID:     uuidv7.New(),
				Name:   "United States",
				Code:   "USA",
				ISO2:   "USA",
				ISO3:   "USA",
				Region: "north_america",
			},
			wantErr: true,
			errMsg:  "iso2 must be exactly 2 characters",
		},
		{
			name: "invalid ISO3 length",
			country: &Country{
				ID:     uuidv7.New(),
				Name:   "United States",
				Code:   "USA",
				ISO2:   "US",
				ISO3:   "US",
				Region: "north_america",
			},
			wantErr: true,
			errMsg:  "iso3 must be exactly 3 characters",
		},
		{
			name: "invalid region",
			country: &Country{
				ID:     uuidv7.New(),
				Name:   "United States",
				Code:   "USA",
				ISO2:   "US",
				ISO3:   "USA",
				Region: "invalid_region",
			},
			wantErr: true,
			errMsg:  "region must be one of",
		},
		{
			name: "auto-uppercase code and ISO codes",
			country: &Country{
				ID:     uuidv7.New(),
				Name:   "United States",
				Code:   "usa",
				ISO2:   "us",
				ISO3:   "usa",
				Region: "north_america",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.country.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				if tt.name == "auto-uppercase code and ISO codes" {
					assert.Equal(t, "USA", tt.country.Code)
					assert.Equal(t, "US", tt.country.ISO2)
					assert.Equal(t, "USA", tt.country.ISO3)
				}
			}
		})
	}
}

func TestCurrency_Validate(t *testing.T) {
	tests := []struct {
		name     string
		currency *Currency
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid currency",
			currency: &Currency{
				ID:        uuidv7.New(),
				Name:      "US Dollar",
				Code:      "USD",
				Symbol:    "$",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing name",
			currency: &Currency{
				ID:     uuidv7.New(),
				Code:   "USD",
				Symbol: "$",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "missing code",
			currency: &Currency{
				ID:     uuidv7.New(),
				Name:   "US Dollar",
				Symbol: "$",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "code too short",
			currency: &Currency{
				ID:     uuidv7.New(),
				Name:   "US Dollar",
				Code:   "US",
				Symbol: "$",
			},
			wantErr: true,
			errMsg:  "code must be between 3 and 10 characters",
		},
		{
			name: "symbol too long",
			currency: &Currency{
				ID:     uuidv7.New(),
				Name:   "US Dollar",
				Code:   "USD",
				Symbol: "verylongsymbol",
			},
			wantErr: true,
			errMsg:  "symbol must not exceed 10 characters",
		},
		{
			name: "auto-uppercase code",
			currency: &Currency{
				ID:     uuidv7.New(),
				Name:   "US Dollar",
				Code:   "usd",
				Symbol: "$",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currency.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				if tt.name == "auto-uppercase code" {
					assert.Equal(t, "USD", tt.currency.Code)
				}
			}
		})
	}
}

func TestCountryCurrency_Validate(t *testing.T) {
	tests := []struct {
		name            string
		countryCurrency *CountryCurrency
		wantErr         bool
		errMsg          string
	}{
		{
			name: "valid country-currency",
			countryCurrency: &CountryCurrency{
				CountryID:  uuidv7.New(),
				CurrencyID: uuidv7.New(),
				IsPrimary:  true,
				CreatedAt:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing country_id",
			countryCurrency: &CountryCurrency{
				CountryID:  uuidv7.Nil,
				CurrencyID: uuidv7.New(),
				IsPrimary:  true,
			},
			wantErr: true,
			errMsg:  "country_id is required",
		},
		{
			name: "missing currency_id",
			countryCurrency: &CountryCurrency{
				CountryID:  uuidv7.New(),
				CurrencyID: uuidv7.Nil,
				IsPrimary:  true,
			},
			wantErr: true,
			errMsg:  "currency_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.countryCurrency.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCountry_RegionValidation(t *testing.T) {
	validRegions := []string{
		"north_america",
		"south_america",
		"western_europe",
		"eastern_europe",
		"asia",
		"middle_east",
		"africa",
		"oceania",
	}

	for _, region := range validRegions {
		t.Run(region, func(t *testing.T) {
			country := &Country{
				ID:     uuidv7.New(),
				Name:   "Test Country",
				Code:   "TST",
				ISO2:   "TS",
				ISO3:   "TST",
				Region: region,
			}
			err := country.Validate()
			assert.NoError(t, err)
			assert.Equal(t, region, country.Region)
		})
	}
}

func TestCurrency_CommonCurrencies(t *testing.T) {
	currencies := []struct {
		name   string
		code   string
		symbol string
	}{
		{"US Dollar", "USD", "$"},
		{"Euro", "EUR", "€"},
		{"British Pound", "GBP", "£"},
		{"Japanese Yen", "JPY", "¥"},
		{"Russian Ruble", "RUB", "₽"},
	}

	for _, curr := range currencies {
		t.Run(curr.name, func(t *testing.T) {
			currency := &Currency{
				ID:        uuidv7.New(),
				Name:      curr.name,
				Code:      curr.code,
				Symbol:    curr.symbol,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			err := currency.Validate()
			assert.NoError(t, err)
			assert.Equal(t, curr.code, currency.Code)
			assert.Equal(t, curr.symbol, currency.Symbol)
		})
	}
}
