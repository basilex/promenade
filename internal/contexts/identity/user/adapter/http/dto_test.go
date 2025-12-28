package http

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToUserResponse(t *testing.T) {
	u, err := user.NewUser("test@example.com", "hashedPassword123")
	require.NoError(t, err)

	u.VerifyEmail()
	loginTime := time.Now()
	u.LastLoginAt = &loginTime

	resp := ToUserResponse(u)

	assert.Equal(t, u.ID.String(), resp.ID)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "active", resp.Status)
	assert.True(t, resp.EmailVerified)
	assert.NotNil(t, resp.EmailVerifiedAt)
	assert.NotNil(t, resp.LastLoginAt)
}

func TestToUserListResponse(t *testing.T) {
	u1, err := user.NewUser("user1@example.com", "hashedPassword1")
	require.NoError(t, err)
	u2, err := user.NewUser("user2@example.com", "hashedPassword2")
	require.NoError(t, err)

	responses := ToUserListResponse([]*user.User{u1, u2})

	assert.Len(t, responses, 2)
	assert.Equal(t, "user1@example.com", responses[0].Email)
	assert.Equal(t, "user2@example.com", responses[1].Email)
}

func TestUserDTOJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		validate func(*testing.T, []byte, error)
	}{
		{
			name:     "RegisterRequest",
			jsonData: `{"email":"test@example.com","name":"John Doe","password":"securepassword123"}`,
			validate: func(t *testing.T, data []byte, err error) {
				require.NoError(t, err)
				var req RegisterRequest
				err = json.Unmarshal(data, &req)
				require.NoError(t, err)
				assert.Equal(t, "test@example.com", req.Email)
				assert.Equal(t, "John Doe", req.Name)
				assert.Equal(t, "securepassword123", req.Password)
			},
		},
		{
			name:     "LoginRequest",
			jsonData: `{"email":"test@example.com","password":"mypassword"}`,
			validate: func(t *testing.T, data []byte, err error) {
				require.NoError(t, err)
				var req LoginRequest
				err = json.Unmarshal(data, &req)
				require.NoError(t, err)
				assert.Equal(t, "test@example.com", req.Email)
				assert.Equal(t, "mypassword", req.Password)
			},
		},
		{
			name:     "ChangePasswordRequest",
			jsonData: `{"old_password":"oldpassword123","new_password":"newpassword456"}`,
			validate: func(t *testing.T, data []byte, err error) {
				require.NoError(t, err)
				var req ChangePasswordRequest
				err = json.Unmarshal(data, &req)
				require.NoError(t, err)
				assert.Equal(t, "oldpassword123", req.OldPassword)
				assert.Equal(t, "newpassword456", req.NewPassword)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, []byte(tt.jsonData), nil)
		})
	}
}

func TestUserResponse_JSONMarshaling(t *testing.T) {
	u, err := user.NewUser("test@example.com", "hashedPassword123")
	require.NoError(t, err)
	u.VerifyEmail()

	resp := ToUserResponse(u)
	jsonData, err := json.Marshal(resp)

	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "test@example.com")
	assert.Contains(t, string(jsonData), "active")
	assert.Contains(t, string(jsonData), `"email_verified":true`)
}

func TestUserResponse_JSONUnmarshaling(t *testing.T) {
	jsonData := `{
		"id":"019b64f7-24f6-7e22-a85e-c6614c03b11f",
		"email":"test@example.com",
		"status":"active",
		"email_verified":true,
		"email_verified_at":"2024-01-01T00:00:00Z",
		"created_at":"2024-01-01T00:00:00Z",
		"updated_at":"2024-01-01T00:00:00Z"
	}`

	var resp UserResponse
	err := json.Unmarshal([]byte(jsonData), &resp)

	require.NoError(t, err)
	assert.Equal(t, "019b64f7-24f6-7e22-a85e-c6614c03b11f", resp.ID)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, "active", resp.Status)
	assert.True(t, resp.EmailVerified)
}

func TestAuthResponse_JSON(t *testing.T) {
	t.Run("marshaling", func(t *testing.T) {
		u, err := user.NewUser("test@example.com", "hashedPassword123")
		require.NoError(t, err)

		authResp := AuthResponse{
			Token: "jwt.token.here",
			User:  ToUserResponse(u),
		}

		jsonData, err := json.Marshal(authResp)

		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "jwt.token.here")
		assert.Contains(t, string(jsonData), "test@example.com")
	})

	t.Run("unmarshaling", func(t *testing.T) {
		jsonData := `{
			"token":"jwt.token.here",
			"user":{
				"id":"019b64f7-24f6-7e22-a85e-c6614c03b11f",
				"email":"test@example.com",
				"status":"active",
				"email_verified":false,
				"created_at":"2024-01-01T00:00:00Z",
				"updated_at":"2024-01-01T00:00:00Z"
			}
		}`

		var authResp AuthResponse
		err := json.Unmarshal([]byte(jsonData), &authResp)

		require.NoError(t, err)
		assert.Equal(t, "jwt.token.here", authResp.Token)
		assert.Equal(t, "test@example.com", authResp.User.Email)
	})
}

func TestRegisterRequest_Marshaling(t *testing.T) {
	req := RegisterRequest{
		Email:    "test@example.com",
		Name:     "John Doe",
		Password: "securepassword123",
	}

	jsonData, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "test@example.com")
	assert.Contains(t, string(jsonData), "John Doe")
}

func TestLoginRequest_Marshaling(t *testing.T) {
	req := LoginRequest{
		Email:    "test@example.com",
		Password: "mypassword",
	}

	jsonData, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "test@example.com")
	assert.Contains(t, string(jsonData), "mypassword")
}
