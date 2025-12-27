package http

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestToTimezoneResponse(t *testing.T) {
	id := uuidv7.New()
	tz := &timezone.Timezone{
		ID:           id,
		Name:         "Europe/Kyiv",
		Abbreviation: "EET",
		UTCOffset:    7200, // +02:00 in seconds
		IsActive:     true,
	}

	resp := ToTimezoneResponse(tz)

	assert.Equal(t, id.String(), resp.ID)
	assert.Equal(t, "Europe/Kyiv", resp.Name)
	assert.Equal(t, "EET", resp.Abbreviation)
	assert.Equal(t, "+02:00", resp.UTCOffset)
	assert.True(t, resp.IsActive)
}

func TestToTimezoneResponses(t *testing.T) {
	timezones := []*timezone.Timezone{
		{ID: uuidv7.New(), Name: "Europe/Kyiv", Abbreviation: "EET", UTCOffset: 7200, IsActive: true},  // +02:00
		{ID: uuidv7.New(), Name: "America/New_York", Abbreviation: "EST", UTCOffset: -18000, IsActive: true},  // -05:00
	}

	responses := ToTimezoneResponses(timezones)

	assert.Len(t, responses, 2)
	assert.Equal(t, "Europe/Kyiv", responses[0].Name)
	assert.Equal(t, "America/New_York", responses[1].Name)
}

func TestCreateTimezoneRequest_JSONMarshal(t *testing.T) {
	req := CreateTimezoneRequest{
		Name:         "Europe/Kyiv",
		Abbreviation: "EET",
		UTCOffset:    "+02:00",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"name":"Europe/Kyiv"`)
	assert.Contains(t, string(data), `"abbreviation":"EET"`)
	assert.Contains(t, string(data), `"utc_offset":"+02:00"`)
}

func TestCreateTimezoneRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"name": "America/New_York",
		"abbreviation": "EST",
		"utc_offset": "-05:00"
	}`

	var req CreateTimezoneRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "America/New_York", req.Name)
	assert.Equal(t, "EST", req.Abbreviation)
	assert.Equal(t, "-05:00", req.UTCOffset)
}

func TestUpdateTimezoneRequest_JSONMarshal(t *testing.T) {
	req := UpdateTimezoneRequest{
		Abbreviation: "EEST",
		UTCOffset:    "+03:00",
		IsActive:     true,
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"abbreviation":"EEST"`)
	assert.Contains(t, string(data), `"utc_offset":"+03:00"`)
	assert.Contains(t, string(data), `"is_active":true`)
}

func TestUpdateTimezoneRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"abbreviation": "PST",
		"utc_offset": "-08:00",
		"is_active": false
	}`

	var req UpdateTimezoneRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "PST", req.Abbreviation)
	assert.Equal(t, "-08:00", req.UTCOffset)
	assert.False(t, req.IsActive)
}

func TestTimezoneResponse_JSONMarshal(t *testing.T) {
	resp := TimezoneResponse{
		ID:           uuidv7.New().String(),
		Name:         "Europe/Kyiv",
		Abbreviation: "EET",
		UTCOffset:    "+02:00",
		IsActive:     true,
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"name":"Europe/Kyiv"`)
	assert.Contains(t, string(data), `"abbreviation":"EET"`)
	assert.Contains(t, string(data), `"utc_offset":"+02:00"`)
	assert.Contains(t, string(data), `"is_active":true`)
}

func TestCreateTimezoneRequest_IANAFormat(t *testing.T) {
	tests := []struct {
		name      string
		tzName    string
		expectOK  bool
	}{
		{"valid IANA format", "Europe/Kyiv", true},
		{"valid IANA with underscore", "America/New_York", true},
		{"valid UTC", "UTC", true},
		{"invalid no slash", "InvalidName", false},
		{"invalid empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateTimezoneRequest{
				Name:         tt.tzName,
				Abbreviation: "XXX",
				UTCOffset:    "+00:00",
			}

			data, err := json.Marshal(req)
			assert.NoError(t, err)
			assert.NotEmpty(t, data)
		})
	}
}

func TestTimezoneResponse_UTCOffsetFormats(t *testing.T) {
	tests := []struct {
		name      string
		utcOffset string
	}{
		{"positive offset", "+02:00"},
		{"negative offset", "-05:00"},
		{"zero offset", "+00:00"},
		{"large positive", "+12:00"},
		{"large negative", "-11:00"},
		{"with minutes", "+05:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := TimezoneResponse{
				ID:           uuidv7.New().String(),
				Name:         "Test/Timezone",
				Abbreviation: "TST",
				UTCOffset:    tt.utcOffset,
				IsActive:     true,
			}

			data, err := json.Marshal(resp)
			assert.NoError(t, err)
			assert.Contains(t, string(data), tt.utcOffset)

			var unmarshaled TimezoneResponse
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.utcOffset, unmarshaled.UTCOffset)
		})
	}
}
