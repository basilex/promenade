package customer

import (
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
)

// CustomerStatus represents customer lifecycle stage
type CustomerStatus string

const (
	CustomerStatusLead     CustomerStatus = "lead"     // Initial contact, not qualified
	CustomerStatusProspect CustomerStatus = "prospect" // Qualified lead, potential customer
	CustomerStatusCustomer CustomerStatus = "customer" // Active paying customer
	CustomerStatusChurned  CustomerStatus = "churned"  // Lost customer
)

// CustomerTier represents customer subscription level
type CustomerTier string

const (
	CustomerTierFree       CustomerTier = "free"
	CustomerTierBasic      CustomerTier = "basic"
	CustomerTierPro        CustomerTier = "pro"
	CustomerTierEnterprise CustomerTier = "enterprise"
)

// Customer is an aggregate root for CRM customer management
type Customer struct {
	aggregate.BaseAggregate

	// Identity
	UserID    *uuidv7.UUID // Optional link to Identity.User
	CompanyID *uuidv7.UUID // Optional (B2B) or nil (B2C)

	// Basic Info
	Name  string
	Email valueobject.Email
	Phone *valueobject.Phone

	// Status
	Status CustomerStatus
	Tier   CustomerTier
	Source string // website, referral, cold-call, event

	// Relationship
	AssignedTo uuidv7.UUID // Sales rep (Identity.User)
	Tags       []string    // marketing, vip, high-value

	// Lifecycle (CreatedAt, UpdatedAt, DeletedAt from BaseAggregate)
	LastContactedAt *time.Time // Last interaction timestamp
	ConvertedAt     *time.Time // Timestamp when converted to paying customer
	ChurnedAt       *time.Time
	ChurnReason     string // Reason for churn
}

// NewCustomer creates a new customer (Lead status)
func NewCustomer(name, email, source string, assignedTo uuidv7.UUID) (*Customer, error) {
	if name == "" {
		return nil, ErrCustomerNameEmpty
	}

	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, ErrCustomerEmailInvalid
	}

	if source == "" {
		return nil, ErrCustomerSourceEmpty
	}

	if assignedTo == uuidv7.Nil {
		return nil, ErrCustomerSalesRepEmpty
	}

	return &Customer{
		BaseAggregate: aggregate.NewBaseAggregate(),
		Name:          name,
		Email:         emailVO,
		Status:        CustomerStatusLead,
		Tier:          CustomerTierFree,
		Source:        source,
		AssignedTo:    assignedTo,
		Tags:          []string{},
	}, nil
}

// NewB2BCustomer creates a B2B customer (linked to company)
func NewB2BCustomer(name, email, source string, companyID, assignedTo uuidv7.UUID) (*Customer, error) {
	if companyID == uuidv7.Nil {
		return nil, ErrCustomerCompanyIDEmpty
	}

	customer, err := NewCustomer(name, email, source, assignedTo)
	if err != nil {
		return nil, err
	}

	customer.CompanyID = &companyID
	return customer, nil
}

// SetPhone sets customer phone number
func (c *Customer) SetPhone(phone string) error {
	if phone == "" {
		c.Phone = nil
		c.Touch()
		return nil
	}

	phoneVO, err := valueobject.NewPhone(phone)
	if err != nil {
		return ErrCustomerPhoneInvalid
	}

	c.Phone = &phoneVO
	c.Touch()
	return nil
}

// LinkToUser links customer to Identity.User (when they register account)
func (c *Customer) LinkToUser(userID uuidv7.UUID) error {
	if userID == uuidv7.Nil {
		return ErrCustomerUserIDEmpty
	}

	if c.UserID != nil {
		return ErrCustomerUserAlreadyLinked
	}

	c.UserID = &userID
	c.Touch()
	return nil
}

// QualifyAsProspect moves lead to prospect status
func (c *Customer) QualifyAsProspect() error {
	if c.Status != CustomerStatusLead {
		return ErrInvalidStatusTransition
	}

	c.Status = CustomerStatusProspect
	c.Touch()
	return nil
}

// ConvertToCustomer moves prospect to customer status (paying customer)
func (c *Customer) ConvertToCustomer() error {
	if c.Status != CustomerStatusProspect {
		return ErrInvalidStatusTransition
	}

	now := time.Now()
	c.Status = CustomerStatusCustomer
	c.ConvertedAt = &now
	c.Touch()
	return nil
}

// Churn marks customer as churned (lost customer)
func (c *Customer) Churn(reason string) error {
	if c.Status != CustomerStatusCustomer {
		return ErrInvalidStatusTransition
	}

	if reason == "" {
		return ErrCustomerChurnReasonEmpty
	}

	now := time.Now()
	c.Status = CustomerStatusChurned
	c.ChurnedAt = &now
	c.ChurnReason = reason
	c.Touch()
	return nil
}

// Reactivate brings churned customer back to customer status
func (c *Customer) Reactivate() error {
	if c.Status != CustomerStatusChurned {
		return ErrInvalidStatusTransition
	}

	c.Status = CustomerStatusCustomer
	c.ChurnedAt = nil
	c.ChurnReason = ""
	c.Touch()
	return nil
}

// Reassign changes assigned sales rep
func (c *Customer) Reassign(newRepID uuidv7.UUID) error {
	if newRepID == uuidv7.Nil {
		return ErrCustomerSalesRepEmpty
	}

	if c.AssignedTo == newRepID {
		return ErrCustomerAssignedToSameRep
	}

	c.AssignedTo = newRepID
	c.Touch()
	return nil
}

// UpgradeTier changes customer subscription tier (only upgrades)
func (c *Customer) UpgradeTier(newTier CustomerTier) error {
	if !isValidTier(newTier) {
		return ErrInvalidTierTransition
	}

	if !isValidTierUpgrade(c.Tier, newTier) {
		return ErrInvalidTierTransition
	}

	c.Tier = newTier
	c.Touch()
	return nil
}

// DowngradeTier changes customer to lower tier
func (c *Customer) DowngradeTier(newTier CustomerTier) error {
	if !isValidTier(newTier) {
		return ErrInvalidTierTransition
	}

	// Validate downgrade path (opposite of upgrade)
	if isValidTierUpgrade(newTier, c.Tier) {
		c.Tier = newTier
		c.Touch()
		return nil
	}

	return ErrInvalidTierTransition
}

// AddTag adds a tag to customer
func (c *Customer) AddTag(tag string) error {
	if tag == "" {
		return ErrCustomerTagEmpty
	}

	for _, t := range c.Tags {
		if t == tag {
			return ErrCustomerTagAlreadyExists
		}
	}

	c.Tags = append(c.Tags, tag)
	c.Touch()
	return nil
}

// RemoveTag removes a tag from customer
func (c *Customer) RemoveTag(tag string) error {
	found := false
	newTags := make([]string, 0, len(c.Tags))

	for _, t := range c.Tags {
		if t != tag {
			newTags = append(newTags, t)
		} else {
			found = true
		}
	}

	if !found {
		return ErrCustomerTagNotFound
	}

	c.Tags = newTags
	c.Touch()
	return nil
}

// IsLead checks if customer is in lead status
func (c *Customer) IsLead() bool {
	return c.Status == CustomerStatusLead
}

// IsProspect checks if customer is in prospect status
func (c *Customer) IsProspect() bool {
	return c.Status == CustomerStatusProspect
}

// IsActiveCustomer checks if customer is active paying customer
func (c *Customer) IsActiveCustomer() bool {
	return c.Status == CustomerStatusCustomer
}

// IsChurned checks if customer is churned
func (c *Customer) IsChurned() bool {
	return c.Status == CustomerStatusChurned
}

// IsB2B checks if this is B2B customer (has company)
func (c *Customer) IsB2B() bool {
	return c.CompanyID != nil
}

// HasAccount checks if customer has user account
func (c *Customer) HasAccount() bool {
	return c.UserID != nil
}

// Validate validates customer data
func (c *Customer) Validate() error {
	if c.Name == "" {
		return ErrCustomerNameEmpty
	}

	if c.Email.Value() == "" {
		return ErrCustomerEmailInvalid
	}

	if !isValidStatus(c.Status) {
		return ErrInvalidStatusTransition
	}

	if !isValidTier(c.Tier) {
		return ErrInvalidTierTransition
	}

	if c.Source == "" {
		return ErrCustomerSourceEmpty
	}

	if c.AssignedTo == uuidv7.Nil {
		return ErrCustomerSalesRepEmpty
	}

	return nil
}

// Helper functions

func isValidStatus(status CustomerStatus) bool {
	switch status {
	case CustomerStatusLead, CustomerStatusProspect, CustomerStatusCustomer, CustomerStatusChurned:
		return true
	default:
		return false
	}
}

func isValidTier(tier CustomerTier) bool {
	switch tier {
	case CustomerTierFree, CustomerTierBasic, CustomerTierPro, CustomerTierEnterprise:
		return true
	default:
		return false
	}
}

func isValidTierUpgrade(from, to CustomerTier) bool {
	tierOrder := map[CustomerTier]int{
		CustomerTierFree:       0,
		CustomerTierBasic:      1,
		CustomerTierPro:        2,
		CustomerTierEnterprise: 3,
	}

	fromLevel, fromExists := tierOrder[from]
	toLevel, toExists := tierOrder[to]

	if !fromExists || !toExists {
		return false
	}

	return toLevel > fromLevel
}
