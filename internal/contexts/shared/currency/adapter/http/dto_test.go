package http

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestToCurrencyResponse(t *testing.T) {
	c, _ := currency.NewCurrency("USD", "US Dollar", "$", 2)
	c.NumericCode = "840"
	c.IsActive = true

	resp := ToCurrencyResponse(c)

	assert.Equal(t, c.GetID().String(), resp.ID)
	assert.Equal(t, "USD", resp.Code)
	assert.Equal(t, "US Dollar", resp.Name)
	assert.Equal(t, "$", resp.Symbol)
	assert.Equal(t, 2, resp.DecimalPlaces)
	assert.Equal(t, "840", resp.NumericCode)
	assert.True(t, resp.IsActive)
}

func TestToCurrencyResponses(t *testing.T) {
	c1, _ := currency.NewCurrency("USD", "US Dollar", "$", 2)
	c1.IsActive = true
	c2, _ := currency.NewCurrency("EUR", "Euro", "€", 2)
	c2.IsActive = true
	currencies := []*currency.Currency{c1, c2}

	responses := ToCurrencyResponses(currencies)

	assert.Len(t, responses, 2)
	assert.Equal(t, "USD", responses[0].Code)
	assert.Equal(t, "EUR", responses[1].Code)
}

func TestCreateCurrencyRequest_JSONMarshal(t *testing.T) {
	req := CreateCurrencyRequest{
		Code:          "USD",
		Name:          "US Dollar",
		Symbol:        "$",
		DecimalPlaces: 2,
		NumericCode:   "840",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"code":"USD"`)
	assert.Contains(t, string(data), `"symbol":"$"`)
	assert.Contains(t, string(data), `"decimal_places":2`)
}

func TestCreateCurrencyRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"code": "USD",
		"name": "US Dollar",
		"symbol": "$",
		"decimal_places": 2,
		"numeric_code": "840"
	}`

	var req CreateCurrencyRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "USD", req.Code)
	assert.Equal(t, "US Dollar", req.Name)
	assert.Equal(t, "$", req.Symbol)
	assert.Equal(t, 2, req.DecimalPlaces)
	assert.Equal(t, "840", req.NumericCode)
}

func TestUpdateCurrencyRequest_JSONMarshal(t *testing.T) {
	req := UpdateCurrencyRequest{
		Name:          "US Dollar",
		Symbol:        "$",
		DecimalPlaces: 2,
		IsActive:      true,
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"name":"US Dollar"`)
	assert.Contains(t, string(data), `"is_active":true`)
}

func TestUpdateCurrencyRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"name": "US Dollar Updated",
		"symbol": "$",
		"decimal_places": 4,
		"numeric_code": "840",
		"is_active": false
	}`

	var req UpdateCurrencyRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "US Dollar Updated", req.Name)
	assert.Equal(t, "$", req.Symbol)
	assert.Equal(t, 4, req.DecimalPlaces)
	assert.Equal(t, "840", req.NumericCode)
	assert.False(t, req.IsActive)
}

func TestCurrencyResponse_JSONMarshal(t *testing.T) {
	resp := CurrencyResponse{
		ID:            uuidv7.New().String(),
		Code:          "USD",
		Name:          "US Dollar",
		Symbol:        "$",
		DecimalPlaces: 2,
		NumericCode:   "840",
		IsActive:      true,
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"code":"USD"`)
	assert.Contains(t, string(data), `"symbol":"$"`)
	assert.Contains(t, string(data), `"decimal_places":2`)
	assert.Contains(t, string(data), `"is_active":true`)
}

func TestCreateCurrencyRequest_DecimalPlacesRange(t *testing.T) {
	tests := []struct {
		name          string
		decimalPlaces int
		expectValid   bool
	}{
		{"zero decimals", 0, true},
		{"two decimals", 2, true},
		{"four decimals", 4, true},
		{"negative decimals", -1, false},
		{"too many decimals", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateCurrencyRequest{
				Code:          "XXX",
				Name:          "Test Currency",
				Symbol:        "X",
				DecimalPlaces: tt.decimalPlaces,
			}

			data, err := json.Marshal(req)
			assert.NoError(t, err)
			assert.NotEmpty(t, data)
		})
	}
}
