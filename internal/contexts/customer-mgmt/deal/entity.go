package deal

import (
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
}

// NewDeal creates a new deal
func NewDeal(customerID uuidv7.UUID, name string, value valueobject.Money, assignedTo uuidv7.UUID, expectedCloseDate time.Time) (*Deal, error) {
	if name == "" {
		return nil, ErrDealNameEmpty
	}
	if customerID == uuidv7.Nil {
		return nil, ErrDealCustomerRequired
	}
	if assignedTo == uuidv7.Nil {
		return nil, ErrDealAssignedToRequired
	}
	if value.Amount < 0 {
		return nil, ErrDealValueNegative
	}
	if expectedCloseDate.Before(time.Now()) {
		return nil, ErrDealDateInPast
	}

	return &Deal{
		BaseAggregate:     aggregate.NewBaseAggregate(),
		CustomerID:        customerID,
		Name:              name,
		Value:             value,
		Currency:          value.Currency,
		Stage:             DealStageLead,
		Probability:       10,
		Source:            DealSourceInbound,
		ExpectedCloseDate: expectedCloseDate,
		AssignedTo:        assignedTo,
	}, nil
}

// UpdateBasicInfo updates deal name and description
func (d *Deal) UpdateBasicInfo(name, description string) error {
	if name == "" {
		return ErrDealNameEmpty
	}
	d.Name = name
	d.Description = description
	d.Touch()
	return nil
}

// UpdateValue updates deal value
func (d *Deal) UpdateValue(value valueobject.Money) error {
	if value.Amount < 0 {
		return ErrDealValueNegative
	}
	d.Value = value
	d.Currency = value.Currency
	d.Touch()
	return nil
}

// MoveTo advances deal to a new stage
func (d *Deal) MoveTo(stage DealStage) error {
	if d.Stage == DealStageClosedWon || d.Stage == DealStageClosedLost {
		return ErrDealTerminalStage
	}

	if !isValidStageTransition(d.Stage, stage) {
		return ErrDealInvalidStageTransition
	}

	d.Stage = stage
	d.Probability = getDefaultProbability(stage)
	d.Touch()
	return nil
}

// MarkAsWon marks deal as closed won
func (d *Deal) MarkAsWon(reason string) error {
	if d.Stage == DealStageClosedWon {
		return ErrDealAlreadyWon
	}
	if d.Stage == DealStageClosedLost {
		return ErrDealCannotMarkLostAsWon
	}

	now := time.Now()
	d.Stage = DealStageClosedWon
	d.Probability = 100
	d.ActualCloseDate = &now
	d.CloseReason = reason
	d.Touch()
	return nil
}

// MarkAsLost marks deal as closed lost
func (d *Deal) MarkAsLost(reason string) error {
	if d.Stage == DealStageClosedLost {
		return ErrDealAlreadyLost
	}
	if d.Stage == DealStageClosedWon {
		return ErrDealCannotMarkWonAsLost
	}
	if reason == "" {
		return ErrDealLossReasonRequired
	}

	now := time.Now()
	d.Stage = DealStageClosedLost
	d.Probability = 0
	d.ActualCloseDate = &now
	d.CloseReason = reason
	d.Touch()
	return nil
}

// UpdateProbability updates win probability
func (d *Deal) UpdateProbability(probability int) error {
	if probability < 0 || probability > 100 {
		return ErrDealProbabilityRange
	}
	if d.Stage == DealStageClosedWon || d.Stage == DealStageClosedLost {
		return ErrDealClosedMutation
	}
	d.Probability = probability
	d.Touch()
	return nil
}

// UpdateExpectedCloseDate updates expected close date
func (d *Deal) UpdateExpectedCloseDate(date time.Time) error {
	if date.Before(time.Now()) {
		return ErrDealDateInPast
	}
	if d.Stage == DealStageClosedWon || d.Stage == DealStageClosedLost {
		return ErrDealClosedMutation
	}
	d.ExpectedCloseDate = date
	d.Touch()
	return nil
}

// AssignToSalesRep assigns deal to a sales rep
func (d *Deal) AssignToSalesRep(userID uuidv7.UUID) error {
	if userID == uuidv7.Nil {
		return ErrDealSalesRepRequired
	}
	d.AssignedTo = userID
	d.Touch()
	return nil
}

// SetCompany links deal to a company
func (d *Deal) SetCompany(companyID *uuidv7.UUID) {
	d.CompanyID = companyID
	d.Touch()
}

// SetSource sets deal source
func (d *Deal) SetSource(source DealSource) {
	d.Source = source
	d.Touch()
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
