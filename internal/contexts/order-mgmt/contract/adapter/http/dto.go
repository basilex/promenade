package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/order-mgmt/contract"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateContractRequest represents a request to create a contract
type CreateContractRequest struct {
	OrderID    uuidv7.UUID `json:"order_id" binding:"required"`
	CustomerID uuidv7.UUID `json:"customer_id" binding:"required"`
	Terms      string      `json:"terms" binding:"required,min=10"`
}

// SignContractRequest represents a request to sign a contract
type SignContractRequest struct {
	SignerName  string `json:"signer_name" binding:"required"`
	SignerEmail string `json:"signer_email" binding:"required,email"`
	SignatureID string `json:"signature_id" binding:"required"`
}

// TerminateContractRequest represents a request to terminate a contract
type TerminateContractRequest struct {
	Reason string `json:"reason" binding:"required,min=10"`
}

// SetExpirationDateRequest represents a request to set expiration date
type SetExpirationDateRequest struct {
	ExpirationDate time.Time `json:"expiration_date" binding:"required"`
}

// UpdateContractRequest represents a request to update contract terms
type UpdateContractRequest struct {
	Terms string `json:"terms" binding:"required,min=10"`
}

// ContractResponse represents a contract in responses
type ContractResponse struct {
	ID                 string     `json:"id"`
	OrderID            string     `json:"order_id"`
	CustomerID         string     `json:"customer_id"`
	Terms              string     `json:"terms"`
	Status             string     `json:"status"`
	Version            int        `json:"version"`
	SignerName         *string    `json:"signer_name,omitempty"`
	SignerEmail        *string    `json:"signer_email,omitempty"`
	SignatureID        *string    `json:"signature_id,omitempty"`
	SignedAt           *time.Time `json:"signed_at,omitempty"`
	TerminationReason  *string    `json:"termination_reason,omitempty"`
	TerminatedAt       *time.Time `json:"terminated_at,omitempty"`
	ExpirationDate     *time.Time `json:"expiration_date,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ContractListResponse represents a list of contracts
type ContractListResponse struct {
	Contracts []ContractResponse `json:"contracts"`
	Total     int                `json:"total"`
}

// ToContractResponse converts a contract entity to response DTO
func ToContractResponse(c *contract.Contract) ContractResponse {
	resp := ContractResponse{
		ID:         c.ID.String(),
		OrderID:    c.OrderID.String(),
		CustomerID: c.CustomerID.String(),
		Terms:      c.Terms,
		Status:     string(c.Status),
		Version:    c.Version,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}

	// Handle optional fields - only include if not empty
	if c.SignedByName != "" {
		resp.SignerName = &c.SignedByName
	}
	if c.SignedByEmail != "" {
		resp.SignerEmail = &c.SignedByEmail
	}
	if c.SignatureID != "" {
		resp.SignatureID = &c.SignatureID
	}
	if c.SignedAt != nil {
		resp.SignedAt = c.SignedAt
	}
	if c.TerminationReason != "" {
		resp.TerminationReason = &c.TerminationReason
	}
	if c.TerminatedAt != nil {
		resp.TerminatedAt = c.TerminatedAt
	}
	if c.ExpiresAt != nil {
		resp.ExpirationDate = c.ExpiresAt
	}

	return resp
}

// ToContractListResponse converts a slice of contracts to list response
func ToContractListResponse(contracts []*contract.Contract, total int) ContractListResponse {
	resp := ContractListResponse{
		Contracts: make([]ContractResponse, 0, len(contracts)),
		Total:     total,
	}

	for _, c := range contracts {
		resp.Contracts = append(resp.Contracts, ToContractResponse(c))
	}

	return resp
}
