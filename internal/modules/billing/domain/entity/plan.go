package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// PlanStatus represents the status of a billing plan
type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "active"
	PlanStatusInactive PlanStatus = "inactive"
	PlanStatusArchived PlanStatus = "archived"
)

// IsValid checks if status is valid
func (s PlanStatus) IsValid() bool {
	switch s {
	case PlanStatusActive, PlanStatusInactive, PlanStatusArchived:
		return true
	}
	return false
}

// PlanInterval represents billing interval
type PlanInterval string

const (
	PlanIntervalMonthly PlanInterval = "monthly"
	PlanIntervalYearly  PlanInterval = "yearly"
)

// IsValid checks if interval is valid
func (i PlanInterval) IsValid() bool {
	switch i {
	case PlanIntervalMonthly, PlanIntervalYearly:
		return true
	}
	return false
}

// Plan represents a subscription plan with pricing and features
type Plan struct {
	ID          uuidv7.UUID  `db:"id" json:"id"`
	Slug        string       `db:"slug" json:"slug" validate:"required,min=3,max=100"`
	Name        string       `db:"name" json:"name" validate:"required,min=3,max=255"`
	Description string       `db:"description" json:"description"`
	Status      PlanStatus   `db:"status" json:"status" validate:"required"`
	Interval    PlanInterval `db:"interval" json:"interval" validate:"required"`
	Amount      int64        `db:"amount" json:"amount" validate:"required,min=0"` // Price in cents/kopecks
	Currency    string       `db:"currency" json:"currency" validate:"required,len=3"`
	TrialDays   int          `db:"trial_days" json:"trial_days" validate:"min=0"`

	// Features and limits
	Features            []string `db:"features" json:"features"` // JSONB array
	MaxUsers            *int     `db:"max_users" json:"max_users"`
	MaxProjects         *int     `db:"max_projects" json:"max_projects"`
	MaxStorage          *int64   `db:"max_storage" json:"max_storage"` // Bytes
	IsUnlimitedUsers    bool     `db:"is_unlimited_users" json:"is_unlimited_users"`
	IsUnlimitedProjects bool     `db:"is_unlimited_projects" json:"is_unlimited_projects"`
	IsUnlimitedStorage  bool     `db:"is_unlimited_storage" json:"is_unlimited_storage"`

	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// NewPlan creates a new plan
func NewPlan(name, slug, description string, amount int64, currency string, interval PlanInterval) (*Plan, error) {
	now := time.Now()

	plan := &Plan{
		ID:          uuidv7.New(),
		Slug:        slug,
		Name:        name,
		Description: description,
		Status:      PlanStatusActive,
		Interval:    interval,
		Amount:      amount,
		Currency:    currency,
		TrialDays:   0,
		Features:    []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := plan.Validate(); err != nil {
		return nil, err
	}

	return plan, nil
}

// Validate validates plan data
func (p *Plan) Validate() error {
	if p.ID == uuidv7.Nil {
		return errors.New("plan ID is required")
	}

	if p.Slug == "" {
		return errors.New("slug is required")
	}

	if p.Name == "" {
		return errors.New("name is required")
	}

	if p.Amount < 0 {
		return errors.New("amount must be non-negative")
	}

	if p.Currency == "" {
		return errors.New("currency is required")
	}

	if p.Interval != PlanIntervalMonthly && p.Interval != PlanIntervalYearly {
		return errors.New("invalid plan interval")
	}

	if p.TrialDays < 0 {
		return errors.New("trial days must be non-negative")
	}

	return nil
}

// IsActive checks if plan is active
func (p *Plan) IsActive() bool {
	return p.Status == PlanStatusActive
}

// Activate activates the plan
func (p *Plan) Activate() {
	p.Status = PlanStatusActive
	p.UpdatedAt = time.Now()
}

// Deactivate deactivates the plan
func (p *Plan) Deactivate() {
	p.Status = PlanStatusInactive
	p.UpdatedAt = time.Now()
}

// Archive archives the plan
func (p *Plan) Archive() {
	p.Status = PlanStatusArchived
	p.UpdatedAt = time.Now()
}

// HasUnlimitedUsers checks if plan has unlimited users
func (p *Plan) HasUnlimitedUsers() bool {
	return p.IsUnlimitedUsers
}

// HasUnlimitedProjects checks if plan has unlimited projects
func (p *Plan) HasUnlimitedProjects() bool {
	return p.IsUnlimitedProjects
}

// HasUnlimitedStorage checks if plan has unlimited storage
func (p *Plan) HasUnlimitedStorage() bool {
	return p.IsUnlimitedStorage
}

// GetMonthlyPrice returns monthly equivalent price in cents
func (p *Plan) GetMonthlyPrice() int64 {
	if p.Interval == PlanIntervalMonthly {
		return p.Amount
	}
	// Yearly plan - divide by 12
	return p.Amount / 12
}

// FormatAmount formats the plan amount with currency symbol
func (p *Plan) FormatAmount() string {
	symbols := map[string]string{
		"USD": "$",
		"EUR": "€",
		"GBP": "£",
		"UAH": "₴",
		"RUB": "₽",
	}
	symbol, ok := symbols[p.Currency]
	if !ok {
		symbol = p.Currency + " "
	}
	return symbol + fmt.Sprintf("%.2f", float64(p.Amount)/100.0)
}

// String returns string representation
func (p *Plan) String() string {
	return fmt.Sprintf("Plan{ID: %s, Name: %s, Amount: %d %s/%s}",
		p.ID.String(), p.Name, p.Amount, p.Currency, p.Interval)
}
