package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/customer-mgmt/interaction"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateInteractionRequest represents the request to create an interaction
type CreateInteractionRequest struct {
	CustomerID  string  `json:"customer_id" binding:"required"`
	CompanyID   *string `json:"company_id,omitempty"`
	Type        string  `json:"type" binding:"required"`
	Direction   string  `json:"direction" binding:"required"`
	Subject     string  `json:"subject" binding:"required"`
	Description string  `json:"description,omitempty"`
	CreatedBy   string  `json:"created_by" binding:"required"`
	StartedAt   string  `json:"started_at" binding:"required"`
}

// UpdateContentRequest represents the request to update interaction content
type UpdateContentRequest struct {
	Subject     string `json:"subject" binding:"required"`
	Description string `json:"description,omitempty"`
}

// SetOutcomeRequest represents the request to set interaction outcome
type SetOutcomeRequest struct {
	Outcome string `json:"outcome" binding:"required"`
}

// EndInteractionRequest represents the request to end an interaction
type EndInteractionRequest struct {
	EndedAt string `json:"ended_at" binding:"required"`
}

// SetFollowUpRequest represents the request to set follow-up
type SetFollowUpRequest struct {
	Required     bool    `json:"required"`
	FollowUpDate *string `json:"follow_up_date,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

// AddAttendeeRequest represents the request to add an attendee
type AddAttendeeRequest struct {
	AttendeeID string `json:"attendee_id" binding:"required"`
}

// InteractionResponse represents the response for an interaction
type InteractionResponse struct {
	ID               string     `json:"id"`
	CustomerID       string     `json:"customer_id"`
	CompanyID        *string    `json:"company_id,omitempty"`
	Type             string     `json:"type"`
	Direction        string     `json:"direction"`
	Outcome          *string    `json:"outcome,omitempty"`
	Subject          string     `json:"subject"`
	Description      string     `json:"description,omitempty"`
	CreatedBy        string     `json:"created_by"`
	Attendees        []string   `json:"attendees"`
	StartedAt        time.Time  `json:"started_at"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	DurationSec      *int       `json:"duration_sec,omitempty"`
	FollowUpRequired bool       `json:"follow_up_required"`
	FollowUpDate     *time.Time `json:"follow_up_date,omitempty"`
	FollowUpNotes    string     `json:"follow_up_notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ToInteractionResponse converts an interaction entity to response
func ToInteractionResponse(i *interaction.Interaction) *InteractionResponse {
	resp := &InteractionResponse{
		ID:               i.ID.String(),
		CustomerID:       i.CustomerID.String(),
		Type:             string(i.Type),
		Direction:        string(i.Direction),
		Subject:          i.Subject,
		Description:      i.Description,
		CreatedBy:        i.CreatedBy.String(),
		StartedAt:        i.StartedAt,
		EndedAt:          i.EndedAt,
		DurationSec:      i.DurationSec,
		FollowUpRequired: i.FollowUpRequired,
		FollowUpDate:     i.FollowUpDate,
		FollowUpNotes:    i.FollowUpNotes,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
		Attendees:        make([]string, 0),
	}

	if i.CompanyID != nil {
		companyIDStr := i.CompanyID.String()
		resp.CompanyID = &companyIDStr
	}

	if i.Outcome != nil {
		outcomeStr := string(*i.Outcome)
		resp.Outcome = &outcomeStr
	}

	for _, attendeeID := range i.Attendees {
		resp.Attendees = append(resp.Attendees, attendeeID.String())
	}

	return resp
}

// Helper functions

func parseUUID(s string) (uuidv7.UUID, error) {
	return uuidv7.Parse(s)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, err
	}

	return &t, nil
}
