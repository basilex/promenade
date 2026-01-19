package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/basilex/promenade/internal/contexts/shared/language/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/language/dto"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestToLanguageResponse(t *testing.T) {
	l, _ := aggregate.NewLanguage("en", "English", "English")
	l.Code3 = "eng"
	l.IsActive = true

	resp := dto.ToLanguageResponse(l)

	assert.Equal(t, l.GetID().String(), resp.ID)
	assert.Equal(t, "en", resp.Code)
	assert.Equal(t, "eng", resp.Code3)
	assert.Equal(t, "English", resp.Name)
	assert.Equal(t, "English", resp.NativeName)
	assert.True(t, resp.IsActive)
}

func TestToLanguageResponses(t *testing.T) {
	l1, _ := aggregate.NewLanguage("en", "English", "English")
	l1.IsActive = true
	l2, _ := aggregate.NewLanguage("uk", "Ukrainian", "Українська")
	l2.IsActive = true
	languages := []*aggregate.Language{l1, l2}

	responses := dto.ToLanguageResponses(languages)

	assert.Len(t, responses, 2)
	assert.Equal(t, "en", responses[0].Code)
	assert.Equal(t, "uk", responses[1].Code)
}

func TestCreateLanguageRequest_JSONMarshal(t *testing.T) {
	req := dto.CreateLanguageRequest{
		Code:       "en",
		Code3:      "eng",
		Name:       "English",
		NativeName: "English",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"code":"en"`)
	assert.Contains(t, string(data), `"name":"English"`)
	assert.Contains(t, string(data), `"native_name":"English"`)
}

func TestCreateLanguageRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"code": "uk",
		"code3": "ukr",
		"name": "Ukrainian",
		"native_name": "Українська"
	}`

	var req dto.CreateLanguageRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "uk", req.Code)
	assert.Equal(t, "ukr", req.Code3)
	assert.Equal(t, "Ukrainian", req.Name)
	assert.Equal(t, "Українська", req.NativeName)
}

func TestUpdateLanguageRequest_JSONMarshal(t *testing.T) {
	req := dto.UpdateLanguageRequest{
		Code3:      "eng",
		Name:       "English",
		NativeName: "English",
		IsActive:   true,
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"name":"English"`)
	assert.Contains(t, string(data), `"is_active":true`)
}

func TestUpdateLanguageRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"code3": "eng",
		"name": "English Updated",
		"native_name": "English",
		"is_active": false
	}`

	var req dto.UpdateLanguageRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "eng", req.Code3)
	assert.Equal(t, "English Updated", req.Name)
	assert.Equal(t, "English", req.NativeName)
	assert.False(t, req.IsActive)
}

func TestLanguageResponse_JSONMarshal(t *testing.T) {
	resp := dto.LanguageResponse{
		ID:         uuidv7.New().String(),
		Code:       "uk",
		Code3:      "ukr",
		Name:       "Ukrainian",
		NativeName: "Українська",
		IsActive:   true,
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"code":"uk"`)
	assert.Contains(t, string(data), `"name":"Ukrainian"`)
	assert.Contains(t, string(data), `"native_name":"Українська"`)
	assert.Contains(t, string(data), `"is_active":true`)
}

func TestCreateLanguageRequest_CodeLowercase(t *testing.T) {
	req := dto.CreateLanguageRequest{
		Code:       "EN",
		Name:       "English",
		NativeName: "English",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"code":"EN"`)
}

func TestLanguageResponse_WithOptionalFields(t *testing.T) {
	resp := dto.LanguageResponse{
		ID:         uuidv7.New().String(),
		Code:       "en",
		Name:       "English",
		NativeName: "English",
		IsActive:   true,
	}

	data, err := json.Marshal(resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var unmarshaled dto.LanguageResponse
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, resp.Code, unmarshaled.Code)
}
