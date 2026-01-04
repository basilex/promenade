package http

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

func TestToContactResponse(t *testing.T) {
	userID := uuidv7.New()
	contactID := uuidv7.New()

	t.Run("email contact", func(t *testing.T) {
		email, _ := valueobject.NewEmail("test@example.com")
		c := &contact.Contact{
			BaseAggregate: aggregate.BaseAggregate{ID: contactID},
			UserID:        userID,
			Type:          contact.ContactTypeEmail,
			Label:         "Work",
			Email:         &email,
			IsPrimary:     true,
			IsVerified:    false,
			IsPublic:      true,
		}

		resp := ToContactResponse(c)

		assert.Equal(t, contactID.String(), resp.ID)
		assert.Equal(t, userID.String(), resp.UserID)
		assert.Equal(t, "email", resp.Type)
		assert.Equal(t, "Work", resp.Label)
		assert.True(t, resp.IsPrimary)
		assert.False(t, resp.IsVerified)
		assert.True(t, resp.IsPublic)
		require.NotNil(t, resp.Email)
		assert.Equal(t, "test@example.com", *resp.Email)
		assert.Nil(t, resp.Phone)
		assert.Nil(t, resp.Address)
	})

	t.Run("phone contact", func(t *testing.T) {
		phone, _ := valueobject.NewPhone("+380501234567")
		c := &contact.Contact{
			BaseAggregate: aggregate.BaseAggregate{ID: contactID},
			UserID:        userID,
			Type:          contact.ContactTypePhone,
			Label:         "Mobile",
			Phone:         &phone,
			IsPrimary:     false,
			IsVerified:    true,
			IsPublic:      false,
		}

		resp := ToContactResponse(c)

		assert.Equal(t, "phone", resp.Type)
		assert.Equal(t, "Mobile", resp.Label)
		assert.Nil(t, resp.Email)
		require.NotNil(t, resp.Phone)
		assert.Equal(t, "380", resp.Phone.CountryCode)
		assert.Equal(t, "+380501234567", resp.Phone.Number)
		assert.Nil(t, resp.Address)
	})

	t.Run("address contact", func(t *testing.T) {
		addr, _ := valueobject.NewAddress("123 Main St", "Kyiv", "01001", "UA")
		c := &contact.Contact{
			BaseAggregate: aggregate.BaseAggregate{ID: contactID},
			UserID:        userID,
			Type:          contact.ContactTypeAddress,
			Label:         "Home",
			Address:       &addr,
			IsPrimary:     false,
			IsVerified:    false,
			IsPublic:      true,
		}

		resp := ToContactResponse(c)

		assert.Equal(t, "address", resp.Type)
		assert.Equal(t, "Home", resp.Label)
		assert.Nil(t, resp.Email)
		assert.Nil(t, resp.Phone)
		require.NotNil(t, resp.Address)
		assert.Equal(t, "123 Main St", resp.Address.Street)
		assert.Equal(t, "Kyiv", resp.Address.City)
		assert.Equal(t, "UA", resp.Address.Country)
		assert.Equal(t, "01001", resp.Address.PostalCode)
	})
}

func TestToContactListResponse(t *testing.T) {
	userID := uuidv7.New()
	email1, _ := valueobject.NewEmail("work@example.com")
	email2, _ := valueobject.NewEmail("personal@example.com")

	contacts := []*contact.Contact{
		{
			BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
			UserID:        userID,
			Type:          contact.ContactTypeEmail,
			Label:         "Work",
			Email:         &email1,
		},
		{
			BaseAggregate: aggregate.BaseAggregate{ID: uuidv7.New()},
			UserID:        userID,
			Type:          contact.ContactTypeEmail,
			Label:         "Personal",
			Email:         &email2,
		},
	}

	responses := ToContactListResponse(contacts)

	assert.Len(t, responses, 2)
	assert.Equal(t, "Work", responses[0].Label)
	assert.Equal(t, "Personal", responses[1].Label)
	require.NotNil(t, responses[0].Email)
	require.NotNil(t, responses[1].Email)
	assert.Equal(t, "work@example.com", *responses[0].Email)
	assert.Equal(t, "personal@example.com", *responses[1].Email)
}

func TestCreateContactRequest_JSONMarshal(t *testing.T) {
	t.Run("email contact", func(t *testing.T) {
		email := "test@example.com"
		req := CreateContactRequest{
			Type:  "email",
			Label: "Work",
			Email: &email,
		}

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"type":"email"`)
		assert.Contains(t, string(data), `"label":"Work"`)
		assert.Contains(t, string(data), `"email":"test@example.com"`)
	})

	t.Run("phone contact", func(t *testing.T) {
		req := CreateContactRequest{
			Type:  "phone",
			Label: "Mobile",
			Phone: &CreatePhoneRequest{
				CountryCode: "+380",
				Number:      "501234567",
			},
		}

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"type":"phone"`)
		assert.Contains(t, string(data), `"country_code":"+380"`)
		assert.Contains(t, string(data), `"number":"501234567"`)
	})

	t.Run("address contact", func(t *testing.T) {
		req := CreateContactRequest{
			Type:  "address",
			Label: "Home",
			Address: &CreateAddressRequest{
				Street:     "123 Main St",
				City:       "Kyiv",
				PostalCode: "01001",
				Country:    "Ukraine",
			},
		}

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"type":"address"`)
		assert.Contains(t, string(data), `"street":"123 Main St"`)
		assert.Contains(t, string(data), `"city":"Kyiv"`)
	})
}

func TestCreateContactRequest_JSONUnmarshal(t *testing.T) {
	t.Run("email contact", func(t *testing.T) {
		jsonData := `{
			"type": "email",
			"label": "Work",
			"email": "test@example.com"
		}`

		var req CreateContactRequest
		err := json.Unmarshal([]byte(jsonData), &req)

		assert.NoError(t, err)
		assert.Equal(t, "email", req.Type)
		assert.Equal(t, "Work", req.Label)
		require.NotNil(t, req.Email)
		assert.Equal(t, "test@example.com", *req.Email)
	})

	t.Run("phone contact", func(t *testing.T) {
		jsonData := `{
			"type": "phone",
			"label": "Mobile",
			"phone": {
				"country_code": "+380",
				"number": "501234567"
			}
		}`

		var req CreateContactRequest
		err := json.Unmarshal([]byte(jsonData), &req)

		assert.NoError(t, err)
		assert.Equal(t, "phone", req.Type)
		require.NotNil(t, req.Phone)
		assert.Equal(t, "+380", req.Phone.CountryCode)
		assert.Equal(t, "501234567", req.Phone.Number)
	})

	t.Run("address contact", func(t *testing.T) {
		jsonData := `{
			"type": "address",
			"label": "Home",
			"address": {
				"street": "123 Main St",
				"city": "Kyiv",
				"postal_code": "01001",
				"country": "Ukraine"
			}
		}`

		var req CreateContactRequest
		err := json.Unmarshal([]byte(jsonData), &req)

		assert.NoError(t, err)
		assert.Equal(t, "address", req.Type)
		require.NotNil(t, req.Address)
		assert.Equal(t, "123 Main St", req.Address.Street)
		assert.Equal(t, "Kyiv", req.Address.City)
		assert.Equal(t, "01001", req.Address.PostalCode)
		assert.Equal(t, "Ukraine", req.Address.Country)
	})
}

func TestCreateContactFromRequest(t *testing.T) {
	userID := uuidv7.New().String()

	t.Run("email contact success", func(t *testing.T) {
		email := "test@example.com"
		req := CreateContactRequest{
			Type:  "email",
			Label: "Work",
			Email: &email,
		}

		c, err := CreateContactFromRequest(userID, req)

		assert.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, contact.ContactTypeEmail, c.Type)
		assert.Equal(t, "Work", c.Label)
		require.NotNil(t, c.Email)
		assert.Equal(t, "test@example.com", c.Email.Value())
	})

	t.Run("phone contact success", func(t *testing.T) {
		req := CreateContactRequest{
			Type:  "phone",
			Label: "Mobile",
			Phone: &CreatePhoneRequest{
				CountryCode: "+380",
				Number:      "501234567",
			},
		}

		c, err := CreateContactFromRequest(userID, req)

		assert.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, contact.ContactTypePhone, c.Type)
		assert.Equal(t, "Mobile", c.Label)
		require.NotNil(t, c.Phone)
	})

	t.Run("address contact success", func(t *testing.T) {
		req := CreateContactRequest{
			Type:  "address",
			Label: "Home",
			Address: &CreateAddressRequest{
				Street:     "123 Main St",
				City:       "Kyiv",
				PostalCode: "01001",
				Country:    "UA",
			},
		}

		c, err := CreateContactFromRequest(userID, req)

		assert.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, contact.ContactTypeAddress, c.Type)
		assert.Equal(t, "Home", c.Label)
		require.NotNil(t, c.Address)
		assert.Equal(t, "123 Main St", c.Address.Street)
	})

	t.Run("invalid user id", func(t *testing.T) {
		email := "test@example.com"
		req := CreateContactRequest{
			Type:  "email",
			Label: "Work",
			Email: &email,
		}

		c, err := CreateContactFromRequest("invalid-uuid", req)

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "invalid user_id")
	})

	t.Run("email contact missing email", func(t *testing.T) {
		req := CreateContactRequest{
			Type:  "email",
			Label: "Work",
			Email: nil,
		}

		c, err := CreateContactFromRequest(userID, req)

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("phone contact missing phone", func(t *testing.T) {
		req := CreateContactRequest{
			Type:  "phone",
			Label: "Mobile",
			Phone: nil,
		}

		c, err := CreateContactFromRequest(userID, req)

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "phone is required")
	})

	t.Run("address contact missing address", func(t *testing.T) {
		req := CreateContactRequest{
			Type:    "address",
			Label:   "Home",
			Address: nil,
		}

		c, err := CreateContactFromRequest(userID, req)

		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "address is required")
	})
}

func TestUpdateContactRequest_JSONMarshal(t *testing.T) {
	label := "Updated Label"
	isPublic := true
	req := UpdateContactRequest{
		Label:    &label,
		IsPublic: &isPublic,
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"label":"Updated Label"`)
	assert.Contains(t, string(data), `"is_public":true`)
}

func TestUpdateContactRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"label": "Updated Label",
		"is_public": false
	}`

	var req UpdateContactRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	require.NotNil(t, req.Label)
	assert.Equal(t, "Updated Label", *req.Label)
	require.NotNil(t, req.IsPublic)
	assert.False(t, *req.IsPublic)
}
