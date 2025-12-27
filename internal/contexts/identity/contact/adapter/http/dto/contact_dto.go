package dto

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/internal/contexts/identity/contact"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateContactRequest represents the request to create a new contact
type CreateContactRequest struct {
	Type    string                `json:"type" binding:"required,oneof=email phone address"`
	Label   string                `json:"label" binding:"required,min=1,max=50"`
	Email   *string               `json:"email,omitempty"`
	Phone   *CreatePhoneRequest   `json:"phone,omitempty"`
	Address *CreateAddressRequest `json:"address,omitempty"`
}

// CreatePhoneRequest represents phone data in create request
type CreatePhoneRequest struct {
	CountryCode string `json:"country_code" binding:"required"`
	Number      string `json:"number" binding:"required"`
}

// CreateAddressRequest represents address data in create request
type CreateAddressRequest struct {
	Street     string  `json:"street" binding:"required"`
	City       string  `json:"city" binding:"required"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country" binding:"required"`
	Apartment  *string `json:"apartment,omitempty"`
}

// UpdateContactRequest represents the request to update a contact
type UpdateContactRequest struct {
	Label    *string               `json:"label,omitempty" binding:"omitempty,min=1,max=50"`
	Email    *string               `json:"email,omitempty"`
	Phone    *CreatePhoneRequest   `json:"phone,omitempty"`
	Address  *CreateAddressRequest `json:"address,omitempty"`
	IsPublic *bool                 `json:"is_public,omitempty"`
}

// ContactResponse represents a contact in API responses
type ContactResponse struct {
	ID         string           `json:"id"`
	UserID     string           `json:"user_id"`
	Type       string           `json:"type"`
	Label      string           `json:"label"`
	Email      *string          `json:"email,omitempty"`
	Phone      *PhoneResponse   `json:"phone,omitempty"`
	Address    *AddressResponse `json:"address,omitempty"`
	IsPrimary  bool             `json:"is_primary"`
	IsVerified bool             `json:"is_verified"`
	IsPublic   bool             `json:"is_public"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// PhoneResponse represents phone data in responses
type PhoneResponse struct {
	CountryCode string `json:"country_code"`
	Number      string `json:"number"`
	Formatted   string `json:"formatted"`
}

// AddressResponse represents address data in responses
type AddressResponse struct {
	Street     string  `json:"street"`
	City       string  `json:"city"`
	State      string  `json:"state,omitempty"`
	PostalCode string  `json:"postal_code,omitempty"`
	Country    string  `json:"country"`
	Apartment  *string `json:"apartment,omitempty"`
}

// ToContactResponse converts a Contact entity to ContactResponse DTO
func ToContactResponse(c *contact.Contact) ContactResponse {
	resp := ContactResponse{
		ID:         c.ID.String(),
		UserID:     c.UserID.String(),
		Type:       string(c.Type),
		Label:      c.Label,
		IsPrimary:  c.IsPrimary,
		IsVerified: c.IsVerified,
		IsPublic:   c.IsPublic,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}

	if c.Email != nil {
		email := c.Email.Value()
		resp.Email = &email
	}

	if c.Phone != nil {
		resp.Phone = &PhoneResponse{
			CountryCode: c.Phone.CountryCode(),
			Number:      c.Phone.Value(),
			Formatted:   c.Phone.Formatted(),
		}
	}

	if c.Address != nil {
		apartment := c.Address.Street2
		var aptPtr *string
		if apartment != "" {
			aptPtr = &apartment
		}
		resp.Address = &AddressResponse{
			Street:     c.Address.Street,
			City:       c.Address.City,
			State:      c.Address.State,
			PostalCode: c.Address.PostalCode,
			Country:    c.Address.Country,
			Apartment:  aptPtr,
		}
	}

	return resp
}

// ToContactListResponse converts a slice of contacts to response DTOs
func ToContactListResponse(contacts []*contact.Contact) []ContactResponse {
	result := make([]ContactResponse, len(contacts))
	for i, c := range contacts {
		result[i] = ToContactResponse(c)
	}
	return result
}

// CreateContactFromRequest creates a Contact entity from CreateContactRequest
func CreateContactFromRequest(userID string, req CreateContactRequest) (*contact.Contact, error) {
	// Parse user ID
	userUUID, err := uuidv7.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	contactType := contact.ContactType(req.Type)

	switch contactType {
	case contact.ContactTypeEmail:
		if req.Email == nil {
			return nil, fmt.Errorf("email is required for email contact type")
		}
		return contact.NewEmailContact(userUUID, *req.Email, req.Label)

	case contact.ContactTypePhone:
		if req.Phone == nil {
			return nil, fmt.Errorf("phone is required for phone contact type")
		}
		// Phone number should already have + from request
		return contact.NewPhoneContact(userUUID, req.Phone.Number, req.Label)

	case contact.ContactTypeAddress:
		if req.Address == nil {
			return nil, fmt.Errorf("address is required for address contact type")
		}
		return contact.NewAddressContact(
			userUUID,
			req.Address.Street,
			req.Address.City,
			req.Address.Country,
			req.Address.PostalCode,
			req.Label,
		)

	default:
		return nil, fmt.Errorf("invalid contact type: %s", contactType)
	}
}
