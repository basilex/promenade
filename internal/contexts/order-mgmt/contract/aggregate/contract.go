package aggregate

import (
	"time"

	contracterrors "github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ContractStatus represents the lifecycle state of a contract
type ContractStatus string

const (
	ContractStatusDraft            ContractStatus = "draft"
	ContractStatusPendingSignature ContractStatus = "pending_signature"
	ContractStatusActive           ContractStatus = "active"
	ContractStatusCompleted        ContractStatus = "completed"
	ContractStatusTerminated       ContractStatus = "terminated"
)

// Contract is an aggregate root representing a legal agreement for an order
type Contract struct {
	aggregate.BaseAggregate

	// IDs
	OrderID    uuidv7.UUID    `json:"order_id"`
	CustomerID uuidv7.UUID    `json:"customer_id"`
	Status     ContractStatus `json:"status"`

	// Content
	Terms       string `json:"terms"`        // Contract terms and conditions
	TermsURL    string `json:"terms_url"`    // Optional external document link
	Version     int    `json:"version"`      // Version tracking for renewals
	SignatureID string `json:"signature_id"` // External signature service ID (DocuSign, HelloSign)

	// Lifecycle timestamps
	SignedAt     *time.Time `json:"signed_at,omitempty"`
	ActivatedAt  *time.Time `json:"activated_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	TerminatedAt *time.Time `json:"terminated_at,omitempty"`
	RenewedAt    *time.Time `json:"renewed_at,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // Optional time-limited contracts

	// Termination
	TerminationReason string `json:"termination_reason,omitempty"` // Required when terminated

	// Metadata
	SignedByName  string `json:"signed_by_name,omitempty"`
	SignedByEmail string `json:"signed_by_email,omitempty"`
}

// NewContract creates a new contract in draft status
func NewContract(orderID, customerID uuidv7.UUID, terms string) *Contract {
	return &Contract{
		BaseAggregate: aggregate.NewBaseAggregate(),
		OrderID:       orderID,
		CustomerID:    customerID,
		Terms:         terms,
		Status:        ContractStatusDraft,
		Version:       1,
	}
}

// Validate validates the contract
func (c *Contract) Validate() error {
	if c.OrderID == uuidv7.Nil {
		return contracterrors.ErrOrderIDRequired
	}
	if c.CustomerID == uuidv7.Nil {
		return contracterrors.ErrCustomerIDRequired
	}
	if c.Terms == "" {
		return contracterrors.ErrTermsRequired
	}
	if c.Status == "" {
		return contracterrors.ErrStatusRequired
	}
	return nil
}

// SubmitForSignature submits the contract for customer signature
func (c *Contract) SubmitForSignature() error {
	if c.Status != ContractStatusDraft {
		return contracterrors.ErrCannotSubmitNonDraft
	}
	c.Status = ContractStatusPendingSignature
	c.Touch()
	return nil
}

// Sign signs the contract and activates it
func (c *Contract) Sign(signedByName, signedByEmail, signatureID string) error {
	if c.Status != ContractStatusPendingSignature {
		return contracterrors.ErrCannotSignNonPending
	}
	if signedByName == "" {
		return contracterrors.ErrSignerNameRequired
	}
	if signedByEmail == "" {
		return contracterrors.ErrSignerEmailRequired
	}

	now := time.Now()
	c.Status = ContractStatusActive
	c.SignedByName = signedByName
	c.SignedByEmail = signedByEmail
	c.SignatureID = signatureID
	c.SignedAt = &now
	c.ActivatedAt = &now
	c.Touch()
	return nil
}

// Complete marks the contract as completed
func (c *Contract) Complete() error {
	if c.Status != ContractStatusActive {
		return contracterrors.ErrCannotCompleteNonActive
	}
	now := time.Now()
	c.Status = ContractStatusCompleted
	c.CompletedAt = &now
	c.Touch()
	return nil
}

// Terminate terminates the contract
func (c *Contract) Terminate(reason string) error {
	if c.Status != ContractStatusActive {
		return contracterrors.ErrCannotTerminateNonActive
	}
	if reason == "" {
		return contracterrors.ErrTerminationReasonRequired
	}
	now := time.Now()
	c.Status = ContractStatusTerminated
	c.TerminationReason = reason
	c.TerminatedAt = &now
	c.Touch()
	return nil
}

// Renew renews the contract (for recurring contracts)
func (c *Contract) Renew() error {
	if c.Status != ContractStatusActive && c.Status != ContractStatusCompleted {
		return contracterrors.ErrCannotRenewInactiveContract
	}
	now := time.Now()
	c.Version++
	c.RenewedAt = &now
	c.Touch()
	return nil
}

// SetExpirationDate sets the contract expiration date
func (c *Contract) SetExpirationDate(expiresAt time.Time) error {
	if c.Status != ContractStatusDraft && c.Status != ContractStatusPendingSignature {
		return contracterrors.ErrCannotSetExpirationForActive
	}
	if expiresAt.Before(time.Now()) {
		return contracterrors.ErrExpirationInPast
	}
	c.ExpiresAt = &expiresAt
	c.Touch()
	return nil
}

// IsExpired checks if the contract has expired
func (c *Contract) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.ExpiresAt)
}

// IsActive checks if the contract is active and not expired
func (c *Contract) IsActive() bool {
	return c.Status == ContractStatusActive && !c.IsExpired()
}

// CanBeTerminated checks if the contract can be terminated
func (c *Contract) CanBeTerminated() bool {
	return c.Status == ContractStatusActive
}

// CanBeRenewed checks if the contract can be renewed
func (c *Contract) CanBeRenewed() bool {
	return c.Status == ContractStatusActive || c.Status == ContractStatusCompleted
}
