package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/shared/country/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/country/dto"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestToCountryResponse(t *testing.T) {
	c, _ := aggregate.NewCountry("UA", "Ukraine", "+380")
	c.Code3 = "UKR"
	c.NumericCode = "804"
	c.NameLocal = "Україна"
	c.IsActive = true

	resp := dto.ToCountryResponse(c)

	assert.Equal(t, c.GetID().String(), resp.ID)
	assert.Equal(t, "UA", resp.Code)
	assert.Equal(t, "UKR", resp.Code3)
	assert.Equal(t, "804", resp.NumericCode)
	assert.Equal(t, "Ukraine", resp.Name)
	assert.Equal(t, "Україна", resp.NameLocal)
	assert.Equal(t, "+380", resp.PhoneCode)
	assert.True(t, resp.IsActive)
}

func TestToCountryResponses(t *testing.T) {
	c1, _ := aggregate.NewCountry("UA", "Ukraine", "+380")
	c1.IsActive = true
	c2, _ := aggregate.NewCountry("US", "United States", "+1")
	c2.IsActive = true
	countries := []*aggregate.Country{c1, c2}

	responses := dto.ToCountryResponses(countries)

	assert.Len(t, responses, 2)
	assert.Equal(t, "UA", responses[0].Code)
	assert.Equal(t, "US", responses[1].Code)
}

func TestCreateCountryRequest_JSONMarshal(t *testing.T) {
	req := dto.CreateCountryRequest{
		Code:      "UA",
		Code3:     "UKR",
		Name:      "Ukraine",
		NameLocal: "Україна",
		PhoneCode: "+380",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"code":"UA"`)
	assert.Contains(t, string(data), `"name":"Ukraine"`)
}

func TestCreateCountryRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"code": "UA",
		"code3": "UKR",
		"numeric_code": "804",
		"name": "Ukraine",
		"name_local": "Україна",
		"phone_code": "+380"
	}`

	var req dto.CreateCountryRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "UA", req.Code)
	assert.Equal(t, "UKR", req.Code3)
	assert.Equal(t, "804", req.NumericCode)
	assert.Equal(t, "Ukraine", req.Name)
	assert.Equal(t, "Україна", req.NameLocal)
	assert.Equal(t, "+380", req.PhoneCode)
}

func TestUpdateCountryRequest_JSONMarshal(t *testing.T) {
	req := dto.UpdateCountryRequest{
		Code3:     "UKR",
		Name:      "Ukraine",
		PhoneCode: "+380",
		IsActive:  true,
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"name":"Ukraine"`)
	assert.Contains(t, string(data), `"is_active":true`)
}

func TestUpdateCountryRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"code3": "UKR",
		"name": "Ukraine Updated",
		"phone_code": "+380",
		"is_active": false
	}`

	var req dto.UpdateCountryRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "UKR", req.Code3)
	assert.Equal(t, "Ukraine Updated", req.Name)
	assert.Equal(t, "+380", req.PhoneCode)
	assert.False(t, req.IsActive)
}

func TestCountryResponse_JSONMarshal(t *testing.T) {
	resp := dto.CountryResponse{
		ID:          uuidv7.New().String(),
		Code:        "UA",
		Code3:       "UKR",
		NumericCode: "804",
		Name:        "Ukraine",
		NameLocal:   "Україна",
		PhoneCode:   "+380",
		IsActive:    true,
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"code":"UA"`)
	assert.Contains(t, string(data), `"name":"Ukraine"`)
	assert.Contains(t, string(data), `"is_active":true`)
}
