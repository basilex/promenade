package interaction

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// InteractionType represents the type of customer interaction
type InteractionType string

const (
	InteractionTypeCall    InteractionType = "call"
	InteractionTypeEmail   InteractionType = "email"
	InteractionTypeMeeting InteractionType = "meeting"
	InteractionTypeNote    InteractionType = "note"
	InteractionTypeSMS     InteractionType = "sms"
	InteractionTypeChat    InteractionType = "chat"
)

// InteractionDirection represents the direction of communication
type InteractionDirection string

const (
	InteractionDirectionInbound  InteractionDirection = "inbound"
	InteractionDirectionOutbound InteractionDirection = "outbound"
)

// InteractionOutcome represents the outcome of the interaction
type InteractionOutcome string

const (
	InteractionOutcomeSuccessful    InteractionOutcome = "successful"
	InteractionOutcomeNoAnswer      InteractionOutcome = "no_answer"
	InteractionOutcomeVoicemail     InteractionOutcome = "voicemail"
	InteractionOutcomeBusy          InteractionOutcome = "busy"
	InteractionOutcomeScheduled     InteractionOutcome = "scheduled"
	InteractionOutcomeNotInterested InteractionOutcome = "not_interested"
)

// Interaction is an aggregate root for customer interaction tracking
type Interaction struct {
	aggregate.BaseAggregate

	ID         uuidv7.UUID
	CustomerID uuidv7.UUID
	CompanyID  *uuidv7.UUID

	Type      InteractionType
	Direction InteractionDirection
	Outcome   *InteractionOutcome

	Subject     string
	Description string
	CreatedBy   uuidv7.UUID
	Attendees   []uuidv7.UUID

	StartedAt   time.Time
	EndedAt     *time.Time
	DurationSec *int

	FollowUpRequired bool
	FollowUpDate     *time.Time
	FollowUpNotes    string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewInteraction creates a new interaction
func NewInteraction(
	customerID uuidv7.UUID,
	companyID *uuidv7.UUID,
	interactionType InteractionType,
	direction InteractionDirection,
	subject string,
	description string,
	createdBy uuidv7.UUID,
	startedAt time.Time,
) (*Interaction, error) {
	if !isValidInteractionType(interactionType) {
		return nil, fmt.Errorf("invalid interaction type: %s", interactionType)
	}
	if !isValidDirection(direction) {
		return nil, fmt.Errorf("invalid direction: %s", direction)
	}
	if subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if customerID == uuidv7.Nil {
		return nil, fmt.Errorf("customer_id is required")
	}
	if createdBy == uuidv7.Nil {
		return nil, fmt.Errorf("created_by is required")
	}

	now := time.Now()
	return &Interaction{
		BaseAggregate:    aggregate.NewBaseAggregate(),
		ID:               uuidv7.New(),
		CustomerID:       customerID,
		CompanyID:        companyID,
		Type:             interactionType,
		Direction:        direction,
		Subject:          subject,
		Description:      description,
		CreatedBy:        createdBy,
		Attendees:        make([]uuidv7.UUID, 0),
		StartedAt:        startedAt,
		FollowUpRequired: false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// UpdateContent updates the subject and description
func (i *Interaction) UpdateContent(subject, description string) error {
	if subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	i.Subject = subject
	i.Description = description
	i.UpdatedAt = time.Now()
	return nil
}

// SetCompany sets or updates the company association
func (i *Interaction) SetCompany(companyID *uuidv7.UUID) {
	i.CompanyID = companyID
	i.UpdatedAt = time.Now()
}

// SetOutcome sets the interaction outcome
func (i *Interaction) SetOutcome(outcome InteractionOutcome) error {
	if !isValidOutcome(outcome) {
		return fmt.Errorf("invalid outcome: %s", outcome)
	}

	i.Outcome = &outcome
	i.UpdatedAt = time.Now()
	return nil
}

// EndInteraction marks the interaction as ended
func (i *Interaction) EndInteraction(endedAt time.Time) error {
	if i.EndedAt != nil {
		return ErrInteractionAlreadyEnded
	}
	if endedAt.Before(i.StartedAt) {
		return fmt.Errorf("ended_at cannot be before started_at")
	}

	duration := int(endedAt.Sub(i.StartedAt).Seconds())
	i.EndedAt = &endedAt
	i.DurationSec = &duration
	i.UpdatedAt = time.Now()
	return nil
}

// SetFollowUp configures follow-up requirements
func (i *Interaction) SetFollowUp(required bool, followUpDate *time.Time, notes string) error {
	if required && followUpDate == nil {
		return fmt.Errorf("follow_up_date required when follow-up is enabled")
	}

	i.FollowUpRequired = required
	i.FollowUpDate = followUpDate
	i.FollowUpNotes = notes

	if !required {
		i.FollowUpDate = nil
		i.FollowUpNotes = ""
	}

	i.UpdatedAt = time.Now()
	return nil
}

// AddAttendee adds a participant to the interaction
func (i *Interaction) AddAttendee(attendeeID uuidv7.UUID) {
	i.Attendees = append(i.Attendees, attendeeID)
	i.UpdatedAt = time.Now()
}

// RemoveAttendee removes a participant from the interaction
func (i *Interaction) RemoveAttendee(attendeeID uuidv7.UUID) {
	for idx, id := range i.Attendees {
		if id == attendeeID {
			i.Attendees = append(i.Attendees[:idx], i.Attendees[idx+1:]...)
			break
		}
	}
	i.UpdatedAt = time.Now()
}

// Delete soft deletes the interaction
func (i *Interaction) Delete() {
	now := time.Now()
	i.DeletedAt = &now
	i.UpdatedAt = now
}

// Validate performs comprehensive validation
func (i *Interaction) Validate() error {
	if !isValidInteractionType(i.Type) {
		return fmt.Errorf("invalid interaction type: %s", i.Type)
	}
	if !isValidDirection(i.Direction) {
		return fmt.Errorf("invalid direction: %s", i.Direction)
	}
	if i.Outcome != nil && !isValidOutcome(*i.Outcome) {
		return fmt.Errorf("invalid outcome: %s", *i.Outcome)
	}
	if i.Subject == "" {
		return fmt.Errorf("subject cannot be empty")
	}
	if i.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if i.DurationSec != nil && *i.DurationSec < 0 {
		return fmt.Errorf("duration_sec cannot be negative")
	}
	if i.EndedAt != nil && i.EndedAt.Before(i.StartedAt) {
		return fmt.Errorf("ended_at cannot be before started_at")
	}
	if i.FollowUpRequired && i.FollowUpDate == nil {
		return fmt.Errorf("follow_up_date required when follow-up is enabled")
	}

	return nil
}

// Helper functions

func isValidInteractionType(t InteractionType) bool {
	validTypes := []InteractionType{
		InteractionTypeCall,
		InteractionTypeEmail,
		InteractionTypeMeeting,
		InteractionTypeNote,
		InteractionTypeSMS,
		InteractionTypeChat,
	}
	for _, valid := range validTypes {
		if t == valid {
			return true
		}
	}
	return false
}

func isValidDirection(d InteractionDirection) bool {
	return d == InteractionDirectionInbound || d == InteractionDirectionOutbound
}

func isValidOutcome(o InteractionOutcome) bool {
	validOutcomes := []InteractionOutcome{
		InteractionOutcomeSuccessful,
		InteractionOutcomeNoAnswer,
		InteractionOutcomeVoicemail,
		InteractionOutcomeBusy,
		InteractionOutcomeScheduled,
		InteractionOutcomeNotInterested,
	}
	for _, valid := range validOutcomes {
		if o == valid {
			return true
		}
	}
	return false
}
