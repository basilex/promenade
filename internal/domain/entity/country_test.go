package entity

import (
	"testing"

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
				Name:   "United States",
				Code:   "US",
				ISO2:   "US",
				ISO3:   "USA",
				Region: "north_america",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			country: &Country{
				Name: "",
				Code: "US",
				ISO2: "US",
				ISO3: "USA",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too short",
			country: &Country{
				Name: "A",
				Code: "US",
				ISO2: "US",
				ISO3: "USA",
			},
			wantErr: true,
			errMsg:  "name must be between 2 and 100 characters",
		},
		{
			name: "name too long",
			country: &Country{
				Name: string(make([]byte, 101)),
				Code: "US",
				ISO2: "US",
				ISO3: "USA",
			},
			wantErr: true,
			errMsg:  "name must be between 2 and 100 characters",
		},
		{
			name: "empty code",
			country: &Country{
				Name: "United States",
				Code: "",
				ISO2: "US",
				ISO3: "USA",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "invalid ISO2 length",
			country: &Country{
				Name: "United States",
				Code: "US",
				ISO2: "USA",
				ISO3: "USA",
			},
			wantErr: true,
			errMsg:  "iso2 must be exactly 2 characters",
		},
		{
			name: "invalid ISO2 non-alpha",
			country: &Country{
				Name: "United States",
				Code: "US",
				ISO2: "U1",
				ISO3: "USA",
			},
			wantErr: true,
			errMsg:  "iso2 must contain only letters",
		},
		{
			name: "invalid ISO3 length",
			country: &Country{
				Name: "United States",
				Code: "US",
				ISO2: "US",
				ISO3: "US",
			},
			wantErr: true,
			errMsg:  "iso3 must be exactly 3 characters",
		},
		{
			name: "invalid region",
			country: &Country{
				Name:   "United States",
				Code:   "US",
				ISO2:   "US",
				ISO3:   "USA",
				Region: "invalid_region",
			},
			wantErr: true,
			errMsg:  "region must be one of",
		},
		{
			name: "lowercase gets uppercased",
			country: &Country{
				Name:   "United States",
				Code:   "us",
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
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Check if lowercase was uppercased
				if tt.name == "lowercase gets uppercased" {
					assert.Equal(t, "US", tt.country.Code)
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
				Name:   "US Dollar",
				Code:   "USD",
				Symbol: "$",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			currency: &Currency{
				Name:   "",
				Code:   "USD",
				Symbol: "$",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too short",
			currency: &Currency{
				Name:   "A",
				Code:   "USD",
				Symbol: "$",
			},
			wantErr: true,
			errMsg:  "name must be between 2 and 100 characters",
		},
		{
			name: "empty code",
			currency: &Currency{
				Name:   "US Dollar",
				Code:   "",
				Symbol: "$",
			},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "code too short",
			currency: &Currency{
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
				Name:   "US Dollar",
				Code:   "USD",
				Symbol: "123456789012",
			},
			wantErr: true,
			errMsg:  "symbol must not exceed 10 characters",
		},
		{
			name: "code with special characters",
			currency: &Currency{
				Name:   "US Dollar",
				Code:   "US$",
				Symbol: "$",
			},
			wantErr: true,
			errMsg:  "code must contain only letters and numbers",
		},
		{
			name: "lowercase gets uppercased",
			currency: &Currency{
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
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				// Check if lowercase was uppercased
				if tt.name == "lowercase gets uppercased" {
					assert.Equal(t, "USD", tt.currency.Code)
				}
			}
		})
	}
}
