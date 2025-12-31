package deal

import (
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// DealStage represents deal pipeline stage
type DealStage string

const (
	DealStageLead        DealStage = "lead"
	DealStageQualified   DealStage = "qualified"
	DealStageProposal    DealStage = "proposal"
	DealStageNegotiation DealStage = "negotiation"
	DealStageClosedWon   DealStage = "closed_won"
	DealStageClosedLost  DealStage = "closed_lost"
)

// DealSource represents how the deal originated
type DealSource string

const (
	DealSourceInbound     DealSource = "inbound"
	DealSourceOutbound    DealSource = "outbound"
	DealSourceReferral    DealSource = "referral"
	DealSourcePartner     DealSource = "partner"
	DealSourceEvent       DealSource = "event"
	DealSourceAdvertising DealSource = "advertising"
)

// Deal is an aggregate root for sales opportunity management
type Deal struct {
	aggregate.BaseAggregate

	ID         uuidv7.UUID
	CustomerID uuidv7.UUID
	CompanyID  *uuidv7.UUID

	Name        string
	Description string
	Value       valueobject.Money
	Currency    string

	Stage       DealStage
	Probability int
	Source      DealSource

	ExpectedCloseDate time.Time
	ActualCloseDate   *time.Time

	AssignedTo uuidv7.UUID

	CloseReason string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewDeal creates a new deal
func NewDeal(customerID uuidv7.UUID, name string, value valueobject.Money, assignedTo uuidv7.UUID, expectedCloseDate time.Time) (*Deal, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if customerID == uuidv7.Nil {
		return nil, fmt.Errorf("customer_id is required")
	}
	if assignedTo == uuidv7.Nil {
		return nil, fmt.Errorf("assigned_to is required")
	}
	if value.Amount < 0 {
		return nil, fmt.Errorf("value must be non-negative")
	}
	if expectedCloseDate.Before(time.Now()) {
		return nil, fmt.Errorf("expected_close_date cannot be in the past")
	}

	now := time.Now()
	return &Deal{
		BaseAggregate:     aggregate.NewBaseAggregate(),
		ID:                uuidv7.New(),
		CustomerID:        customerID,
		Name:              name,
		Value:             value,
		Currency:          value.Currency,
		Stage:             DealStageLead,
		Probability:       10,
		Source:            DealSourceInbound,
		ExpectedCloseDate: expectedCloseDate,
		AssignedTo:        assignedTo,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

// UpdateBasicInfo updates deal name and description
func (d *Deal) UpdateBasicInfo(name, description string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	d.Name = name
	d.Description = description
	d.UpdatedAt = time.Now()
	return nil
}

// UpdateValue updates deal value
func (d *Deal) UpdateValue(value valueobject.Money) error {
	if value.Amount < 0 {
		return fmt.Errorf("value must be non-negative")
	}
	d.Value = value
	d.Currency = value.Currency
	d.UpdatedAt = time.Now()
	return nil
}

// MoveTo advances deal to a new stage
func (d *Deal) MoveTo(stage DealStage) error {
	if d.Stage == DealStageClosedWon || d.Stage == DealStageClosedLost {
		return fmt.Errorf("cannot move deal from terminal stage %s", d.Stage)
	}

	if !isValidStageTransition(d.Stage, stage) {
		return fmt.Errorf("invalid stage transition from %s to %s", d.Stage, stage)
	}

	d.Stage = stage
	d.Probability = getDefaultProbability(stage)
	d.UpdatedAt = time.Now()
	return nil
}

// MarkAsWon marks deal as closed won
func (d *Deal) MarkAsWon(reason string) error {
	if d.Stage == DealStageClosedWon {
		return fmt.Errorf("deal is already marked as won")
	}
	if d.Stage == DealStageClosedLost {
		return fmt.Errorf("cannot mark lost deal as won")
	}

	now := time.Now()
	d.Stage = DealStageClosedWon
	d.Probability = 100
	d.ActualCloseDate = &now
	d.CloseReason = reason
	d.UpdatedAt = now
	return nil
}

// MarkAsLost marks deal as closed lost
func (d *Deal) MarkAsLost(reason string) error {
	if d.Stage == DealStageClosedLost {
		return fmt.Errorf("deal is already marked as lost")
	}
	if d.Stage == DealStageClosedWon {
		return fmt.Errorf("cannot mark won deal as lost")
	}
	if reason == "" {
		return fmt.Errorf("loss reason is required")
	}

	now := time.Now()
	d.Stage = DealStageClosedLost
	d.Probability = 0
	d.ActualCloseDate = &now
	d.CloseReason = reason
	d.UpdatedAt = now
	return nil
}

// UpdateProbability updates win probability
func (d *Deal) UpdateProbability(probability int) error {
	if probability < 0 || probability > 100 {
		return fmt.Errorf("probability must be between 0 and 100")
	}
	if d.Stage == DealStageClosedWon || d.Stage == DealStageClosedLost {
		return fmt.Errorf("cannot change probability for closed deal")
	}
	d.Probability = probability
	d.UpdatedAt = time.Now()
	return nil
}

// UpdateExpectedCloseDate updates expected close date
func (d *Deal) UpdateExpectedCloseDate(date time.Time) error {
	if date.Before(time.Now()) {
		return fmt.Errorf("expected_close_date cannot be in the past")
	}
	if d.Stage == DealStageClosedWon || d.Stage == DealStageClosedLost {
		return fmt.Errorf("cannot change expected close date for closed deal")
	}
	d.ExpectedCloseDate = date
	d.UpdatedAt = time.Now()
	return nil
}

// AssignToSalesRep assigns deal to a sales rep
func (d *Deal) AssignToSalesRep(userID uuidv7.UUID) error {
	if userID == uuidv7.Nil {
		return fmt.Errorf("sales rep ID is required")
	}
	d.AssignedTo = userID
	d.UpdatedAt = time.Now()
	return nil
}

// SetCompany links deal to a company
func (d *Deal) SetCompany(companyID *uuidv7.UUID) {
	d.CompanyID = companyID
	d.UpdatedAt = time.Now()
}

// SetSource sets deal source
func (d *Deal) SetSource(source DealSource) {
	d.Source = source
	d.UpdatedAt = time.Now()
}

// IsClosed returns true if deal is in terminal stage
func (d *Deal) IsClosed() bool {
	return d.Stage == DealStageClosedWon || d.Stage == DealStageClosedLost
}

// IsWon returns true if deal is won
func (d *Deal) IsWon() bool {
	return d.Stage == DealStageClosedWon
}

// IsLost returns true if deal is lost
func (d *Deal) IsLost() bool {
	return d.Stage == DealStageClosedLost
}

// Helper functions

func isValidStageTransition(from, to DealStage) bool {
	if to == DealStageClosedWon || to == DealStageClosedLost {
		return true
	}

	validTransitions := map[DealStage][]DealStage{
		DealStageLead:        {DealStageQualified, DealStageClosedLost},
		DealStageQualified:   {DealStageProposal, DealStageClosedLost},
		DealStageProposal:    {DealStageNegotiation, DealStageClosedLost},
		DealStageNegotiation: {DealStageClosedWon, DealStageClosedLost},
	}

	allowedStages, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, stage := range allowedStages {
		if stage == to {
			return true
		}
	}
	return false
}

func getDefaultProbability(stage DealStage) int {
	probabilities := map[DealStage]int{
		DealStageLead:        10,
		DealStageQualified:   25,
		DealStageProposal:    50,
		DealStageNegotiation: 75,
		DealStageClosedWon:   100,
		DealStageClosedLost:  0,
	}
	return probabilities[stage]
}

// String methods
func (s DealStage) String() string {
	return string(s)
}

func (s DealSource) String() string {
	return string(s)
}
