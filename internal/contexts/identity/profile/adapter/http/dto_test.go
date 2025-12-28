package http

import (
	"encoding/json"
	"testing"

	"github.com/basilex/promenade/internal/contexts/identity/profile"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToProfileResponse(t *testing.T) {
	userID := uuidv7.New()
	p, err := profile.NewProfile(userID, "John Doe")
	require.NoError(t, err)

	p.Bio = "Software Engineer"
	p.Gender = profile.GenderMale

	resp := ToProfileResponse(p)

	assert.Equal(t, p.ID.String(), resp.ID)
	assert.Equal(t, userID.String(), resp.UserID)
	assert.Equal(t, "John Doe", resp.DisplayName)
	assert.Equal(t, "Software Engineer", resp.Bio)
	assert.Equal(t, "male", resp.Gender)
	assert.True(t, resp.IsActive)
}

func TestToProfileResponseList(t *testing.T) {
	userID1 := uuidv7.New()
	userID2 := uuidv7.New()

	p1, err := profile.NewProfile(userID1, "John Doe")
	require.NoError(t, err)
	p2, err := profile.NewProfile(userID2, "Jane Smith")
	require.NoError(t, err)

	responses := ToProfileResponseList([]*profile.Profile{p1, p2})

	assert.Len(t, responses, 2)
	assert.Equal(t, "John Doe", responses[0].DisplayName)
	assert.Equal(t, "Jane Smith", responses[1].DisplayName)
}

func TestProfileDTOJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected interface{}
	}{
		{
			name:     "CreateProfileRequest",
			jsonData: `{"display_name":"Test User"}`,
			expected: &CreateProfileRequest{DisplayName: "Test User"},
		},
		{
			name:     "UpdateBioRequest",
			jsonData: `{"bio":"New bio"}`,
			expected: &UpdateBioRequest{Bio: "New bio"},
		},
		{
			name:     "UpdateGenderRequest",
			jsonData: `{"gender":"male"}`,
			expected: &UpdateGenderRequest{Gender: "male"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.expected.(type) {
			case *CreateProfileRequest:
				var req CreateProfileRequest
				err := json.Unmarshal([]byte(tt.jsonData), &req)
				require.NoError(t, err)
				assert.Equal(t, v.DisplayName, req.DisplayName)
			case *UpdateBioRequest:
				var req UpdateBioRequest
				err := json.Unmarshal([]byte(tt.jsonData), &req)
				require.NoError(t, err)
				assert.Equal(t, v.Bio, req.Bio)
			case *UpdateGenderRequest:
				var req UpdateGenderRequest
				err := json.Unmarshal([]byte(tt.jsonData), &req)
				require.NoError(t, err)
				assert.Equal(t, v.Gender, req.Gender)
			}
		})
	}
}

func TestProfileResponse_JSONMarshaling(t *testing.T) {
	userID := uuidv7.New()
	p, err := profile.NewProfile(userID, "John Doe")
	require.NoError(t, err)
	p.Bio = "Engineer"

	resp := ToProfileResponse(p)
	jsonData, err := json.Marshal(resp)

	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "John Doe")
	assert.Contains(t, string(jsonData), "Engineer")
}

func TestUpdateDateOfBirthRequest_JSON(t *testing.T) {
	t.Run("valid date", func(t *testing.T) {
		jsonData := `{"date_of_birth":"1990-05-15T00:00:00Z"}`
		var req UpdateDateOfBirthRequest
		err := json.Unmarshal([]byte(jsonData), &req)

		require.NoError(t, err)
		require.NotNil(t, req.DateOfBirth)
		assert.Equal(t, 1990, req.DateOfBirth.Year())
	})

	t.Run("null date", func(t *testing.T) {
		jsonData := `{"date_of_birth":null}`
		var req UpdateDateOfBirthRequest
		err := json.Unmarshal([]byte(jsonData), &req)

		require.NoError(t, err)
		assert.Nil(t, req.DateOfBirth)
	})
}

func TestUpdateLocalizationRequest_JSON(t *testing.T) {
	jsonData := `{"timezone":"Europe/Kyiv","language":"uk","country":"UA"}`
	var req UpdateLocalizationRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	require.NoError(t, err)
	assert.Equal(t, "Europe/Kyiv", req.Timezone)
	assert.Equal(t, "uk", req.Language)
	assert.Equal(t, "UA", req.Country)
}

func TestUpdateSocialLinksRequest_JSON(t *testing.T) {
	jsonData := `{
		"website":"https://example.com",
		"github":"https://github.com/johndoe"
	}`
	var req UpdateSocialLinksRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	require.NoError(t, err)
	assert.Equal(t, "https://example.com", req.Website)
	assert.Equal(t, "https://github.com/johndoe", req.GitHub)
}

func TestUpdatePersonalInfoRequest_JSON(t *testing.T) {
	jsonData := `{"first_name":"John","last_name":"Doe"}`
	var req UpdatePersonalInfoRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	require.NoError(t, err)
	assert.Equal(t, "John", req.FirstName)
	assert.Equal(t, "Doe", req.LastName)
}
