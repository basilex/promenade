package http

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/shared/country"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestToCountryResponse(t *testing.T) {
	id := uuidv7.New()
	c := &country.Country{
		ID:          id,
		Code:        "UA",
		Code3:       "UKR",
		NumericCode: "804",
		Name:        "Ukraine",
		NameLocal:   "Україна",
		PhoneCode:   "+380",
		IsActive:    true,
	}

	resp := ToCountryResponse(c)

	assert.Equal(t, id.String(), resp.ID)
	assert.Equal(t, "UA", resp.Code)
	assert.Equal(t, "UKR", resp.Code3)
	assert.Equal(t, "804", resp.NumericCode)
	assert.Equal(t, "Ukraine", resp.Name)
	assert.Equal(t, "Україна", resp.NameLocal)
	assert.Equal(t, "+380", resp.PhoneCode)
	assert.True(t, resp.IsActive)
}

func TestToCountryResponses(t *testing.T) {
	countries := []*country.Country{
		{ID: uuidv7.New(), Code: "UA", Name: "Ukraine", PhoneCode: "+380", IsActive: true},
		{ID: uuidv7.New(), Code: "US", Name: "United States", PhoneCode: "+1", IsActive: true},
	}

	responses := ToCountryResponses(countries)

	assert.Len(t, responses, 2)
	assert.Equal(t, "UA", responses[0].Code)
	assert.Equal(t, "US", responses[1].Code)
}

func TestCreateCountryRequest_JSONMarshal(t *testing.T) {
	req := CreateCountryRequest{
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

	var req CreateCountryRequest
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
	req := UpdateCountryRequest{
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

	var req UpdateCountryRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "UKR", req.Code3)
	assert.Equal(t, "Ukraine Updated", req.Name)
	assert.Equal(t, "+380", req.PhoneCode)
	assert.False(t, req.IsActive)
}

func TestCountryResponse_JSONMarshal(t *testing.T) {
	resp := CountryResponse{
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
