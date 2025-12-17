package dto

import (
	"time"

	"github.com/basilex/promenade/internal/domain/entity"
)

// CreateUserContactRequest represents a request to create a new user contact
type CreateUserContactRequest struct {
	ContactType   string   `json:"contact_type" binding:"required,oneof=email phone telegram whatsapp viber signal skype discord linkedin other"`
	ContactValue  string   `json:"contact_value" binding:"required,min=1,max=255"`
	Label         *string  `json:"label" binding:"omitempty,max=100"`
	IsPublic      bool     `json:"is_public"`
	AvailableFrom *string  `json:"available_from" binding:"omitempty"` // Format: "09:00:00"
	AvailableTo   *string  `json:"available_to" binding:"omitempty"`   // Format: "18:00:00"
	AvailableDays []string `json:"available_days" binding:"omitempty,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	Timezone      *string  `json:"timezone" binding:"omitempty,max=50"`
	Notes         *string  `json:"notes" binding:"omitempty,max=1000"`
}

// UpdateUserContactRequest represents a request to update an existing contact
type UpdateUserContactRequest struct {
	ContactType   string   `json:"contact_type" binding:"required,oneof=email phone telegram whatsapp viber signal skype discord linkedin other"`
	ContactValue  string   `json:"contact_value" binding:"required,min=1,max=255"`
	Label         *string  `json:"label" binding:"omitempty,max=100"`
	IsPublic      bool     `json:"is_public"`
	AvailableFrom *string  `json:"available_from" binding:"omitempty"` // Format: "09:00:00"
	AvailableTo   *string  `json:"available_to" binding:"omitempty"`   // Format: "18:00:00"
	AvailableDays []string `json:"available_days" binding:"omitempty,dive,oneof=monday tuesday wednesday thursday friday saturday sunday"`
	Timezone      *string  `json:"timezone" binding:"omitempty,max=50"`
	Notes         *string  `json:"notes" binding:"omitempty,max=1000"`
}

// UserContactResponse represents a user contact in API responses
type UserContactResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	ContactType   string    `json:"contact_type"`
	ContactValue  string    `json:"contact_value"`
	Label         *string   `json:"label,omitempty"`
	IsVerified    bool      `json:"is_verified"`
	IsPrimary     bool      `json:"is_primary"`
	IsActive      bool      `json:"is_active"`
	IsPublic      bool      `json:"is_public"`
	AvailableFrom *string   `json:"available_from,omitempty"` // Format: "09:00:00"
	AvailableTo   *string   `json:"available_to,omitempty"`   // Format: "18:00:00"
	AvailableDays []string  `json:"available_days,omitempty"`
	Timezone      *string   `json:"timezone,omitempty"`
	Notes         *string   `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToUserContactResponse converts entity.UserContact to UserContactResponse
func ToUserContactResponse(contact *entity.UserContact) UserContactResponse {
	response := UserContactResponse{
		ID:            contact.ID.String(),
		UserID:        contact.UserID.String(),
		ContactType:   string(contact.ContactType),
		ContactValue:  contact.ContactValue,
		Label:         contact.Label,
		IsVerified:    contact.IsVerified,
		IsPrimary:     contact.IsPrimary,
		IsActive:      contact.IsActive,
		IsPublic:      contact.IsPublic,
		AvailableDays: contact.AvailableDays,
		Timezone:      contact.Timezone,
		Notes:         contact.Notes,
		CreatedAt:     contact.CreatedAt,
		UpdatedAt:     contact.UpdatedAt,
	}

	// Format time.Time to string
	if contact.AvailableFrom != nil {
		timeStr := contact.AvailableFrom.Format("15:04:05")
		response.AvailableFrom = &timeStr
	}

	if contact.AvailableTo != nil {
		timeStr := contact.AvailableTo.Format("15:04:05")
		response.AvailableTo = &timeStr
	}

	return response
}

// ToUserContactListResponse converts a slice of entity.UserContact to a slice of UserContactResponse
func ToUserContactListResponse(contacts []*entity.UserContact) []UserContactResponse {
	responses := make([]UserContactResponse, 0, len(contacts))
	for _, contact := range contacts {
		responses = append(responses, ToUserContactResponse(contact))
	}
	return responses
}

// ParseTimeString converts a time string (HH:MM:SS) to *time.Time
func ParseTimeString(timeStr *string) (*time.Time, error) {
	if timeStr == nil || *timeStr == "" {
		return nil, nil
	}

	t, err := time.Parse("15:04:05", *timeStr)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
