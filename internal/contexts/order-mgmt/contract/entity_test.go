package contract

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewContract tests contract factory method
func TestNewContract(t *testing.T) {
	orderID := uuidv7.New()
	customerID := uuidv7.New()
	terms := "Standard service agreement terms"

	contract := NewContract(orderID, customerID, terms)

	assert.NotEqual(t, uuidv7.Nil, contract.GetID())
	assert.Equal(t, orderID, contract.OrderID)
	assert.Equal(t, customerID, contract.CustomerID)
	assert.Equal(t, terms, contract.Terms)
	assert.Equal(t, ContractStatusDraft, contract.Status)
	assert.Equal(t, 1, contract.Version)
	assert.Nil(t, contract.SignedAt)
	assert.Nil(t, contract.ActivatedAt)
	assert.NotEqual(t, time.Time{}, contract.GetCreatedAt())
}

// TestContract_Validate tests validation logic
func TestContract_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Contract
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid contract",
			setup: func() *Contract {
				return NewContract(uuidv7.New(), uuidv7.New(), "Terms")
			},
			wantErr: false,
		},
		{
			name: "missing order_id",
			setup: func() *Contract {
				c := NewContract(uuidv7.Nil, uuidv7.New(), "Terms")
				return c
			},
			wantErr: true,
			errMsg:  "order_id is required",
		},
		{
			name: "missing customer_id",
			setup: func() *Contract {
				c := NewContract(uuidv7.New(), uuidv7.Nil, "Terms")
				return c
			},
			wantErr: true,
			errMsg:  "customer_id is required",
		},
		{
			name: "missing terms",
			setup: func() *Contract {
				c := NewContract(uuidv7.New(), uuidv7.New(), "")
				return c
			},
			wantErr: true,
			errMsg:  "terms are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract := tt.setup()
			err := contract.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestContract_SubmitForSignature tests state transition from draft to pending_signature
func TestContract_SubmitForSignature(t *testing.T) {
	t.Run("success from draft", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		assert.Equal(t, ContractStatusDraft, contract.Status)

		err := contract.SubmitForSignature()

		assert.NoError(t, err)
		assert.Equal(t, ContractStatusPendingSignature, contract.Status)
	})

	t.Run("cannot submit non-draft contract", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive

		err := contract.SubmitForSignature()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only submit draft contracts")
	})
}

// TestContract_Sign tests signing the contract
func TestContract_Sign(t *testing.T) {
	t.Run("success from pending_signature", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusPendingSignature

		err := contract.Sign("John Doe", "john@example.com", "sig-123")

		require.NoError(t, err)
		assert.Equal(t, ContractStatusActive, contract.Status)
		assert.Equal(t, "John Doe", contract.SignedByName)
		assert.Equal(t, "john@example.com", contract.SignedByEmail)
		assert.Equal(t, "sig-123", contract.SignatureID)
		assert.NotNil(t, contract.SignedAt)
		assert.NotNil(t, contract.ActivatedAt)
	})

	t.Run("cannot sign non-pending contract", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		assert.Equal(t, ContractStatusDraft, contract.Status)

		err := contract.Sign("John Doe", "john@example.com", "sig-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only sign contracts in pending_signature status")
	})

	t.Run("missing signer name", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusPendingSignature

		err := contract.Sign("", "john@example.com", "sig-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signer name is required")
	})

	t.Run("missing signer email", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusPendingSignature

		err := contract.Sign("John Doe", "", "sig-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signer email is required")
	})
}

// TestContract_Complete tests completing active contract
func TestContract_Complete(t *testing.T) {
	t.Run("success from active", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive

		err := contract.Complete()

		require.NoError(t, err)
		assert.Equal(t, ContractStatusCompleted, contract.Status)
		assert.NotNil(t, contract.CompletedAt)
	})

	t.Run("cannot complete non-active contract", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		assert.Equal(t, ContractStatusDraft, contract.Status)

		err := contract.Complete()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only complete active contracts")
	})
}

// TestContract_Terminate tests terminating active contract
func TestContract_Terminate(t *testing.T) {
	t.Run("success from active with reason", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive

		err := contract.Terminate("Customer requested cancellation")

		require.NoError(t, err)
		assert.Equal(t, ContractStatusTerminated, contract.Status)
		assert.Equal(t, "Customer requested cancellation", contract.TerminationReason)
		assert.NotNil(t, contract.TerminatedAt)
	})

	t.Run("cannot terminate non-active contract", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		assert.Equal(t, ContractStatusDraft, contract.Status)

		err := contract.Terminate("Some reason")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only terminate active contracts")
	})

	t.Run("missing termination reason", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive

		err := contract.Terminate("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "termination reason is required")
	})
}

// TestContract_Renew tests contract renewal
func TestContract_Renew(t *testing.T) {
	t.Run("success from active", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive
		initialVersion := contract.Version

		err := contract.Renew()

		require.NoError(t, err)
		assert.Equal(t, initialVersion+1, contract.Version)
		assert.NotNil(t, contract.RenewedAt)
	})

	t.Run("success from completed", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusCompleted
		initialVersion := contract.Version

		err := contract.Renew()

		require.NoError(t, err)
		assert.Equal(t, initialVersion+1, contract.Version)
	})

	t.Run("cannot renew draft contract", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		assert.Equal(t, ContractStatusDraft, contract.Status)

		err := contract.Renew()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only renew active or completed contracts")
	})
}

// TestContract_SetExpirationDate tests expiration date management
func TestContract_SetExpirationDate(t *testing.T) {
	t.Run("success for draft contract", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		futureDate := time.Now().Add(30 * 24 * time.Hour)

		err := contract.SetExpirationDate(futureDate)

		assert.NoError(t, err)
		assert.NotNil(t, contract.ExpiresAt)
	})

	t.Run("cannot set expiration for active contract", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive
		futureDate := time.Now().Add(30 * 24 * time.Hour)

		err := contract.SetExpirationDate(futureDate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only set expiration date for draft or pending contracts")
	})

	t.Run("cannot set past expiration date", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		pastDate := time.Now().Add(-1 * time.Hour)

		err := contract.SetExpirationDate(pastDate)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expiration date must be in the future")
	})
}

// TestContract_IsExpired tests expiration checking
func TestContract_IsExpired(t *testing.T) {
	t.Run("not expired with no expiration date", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		assert.False(t, contract.IsExpired())
	})

	t.Run("not expired with future expiration date", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		futureDate := time.Now().Add(1 * time.Hour)
		contract.ExpiresAt = &futureDate

		assert.False(t, contract.IsExpired())
	})

	t.Run("expired with past expiration date", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		pastDate := time.Now().Add(-1 * time.Hour)
		contract.ExpiresAt = &pastDate

		assert.True(t, contract.IsExpired())
	})
}

// TestContract_IsActive tests active state checking
func TestContract_IsActive(t *testing.T) {
	t.Run("active with correct status and no expiration", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive

		assert.True(t, contract.IsActive())
	})

	t.Run("not active with non-active status", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		assert.Equal(t, ContractStatusDraft, contract.Status)

		assert.False(t, contract.IsActive())
	})

	t.Run("not active when expired", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		contract.Status = ContractStatusActive
		pastDate := time.Now().Add(-1 * time.Hour)
		contract.ExpiresAt = &pastDate

		assert.False(t, contract.IsActive())
	})
}

// TestContract_CanBeTerminated tests termination validation
func TestContract_CanBeTerminated(t *testing.T) {
	tests := []struct {
		name     string
		status   ContractStatus
		expected bool
	}{
		{"draft cannot be terminated", ContractStatusDraft, false},
		{"pending_signature cannot be terminated", ContractStatusPendingSignature, false},
		{"active can be terminated", ContractStatusActive, true},
		{"completed cannot be terminated", ContractStatusCompleted, false},
		{"terminated cannot be terminated", ContractStatusTerminated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
			contract.Status = tt.status

			assert.Equal(t, tt.expected, contract.CanBeTerminated())
		})
	}
}

// TestContract_CanBeRenewed tests renewal validation
func TestContract_CanBeRenewed(t *testing.T) {
	tests := []struct {
		name     string
		status   ContractStatus
		expected bool
	}{
		{"draft cannot be renewed", ContractStatusDraft, false},
		{"pending_signature cannot be renewed", ContractStatusPendingSignature, false},
		{"active can be renewed", ContractStatusActive, true},
		{"completed can be renewed", ContractStatusCompleted, true},
		{"terminated cannot be renewed", ContractStatusTerminated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
			contract.Status = tt.status

			assert.Equal(t, tt.expected, contract.CanBeRenewed())
		})
	}
}

// TestContract_Lifecycle tests complete contract workflows
func TestContract_Lifecycle(t *testing.T) {
	t.Run("complete lifecycle: draft -> pending -> active -> completed", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")

		// Step 1: Draft
		assert.Equal(t, ContractStatusDraft, contract.Status)

		// Step 2: Submit for signature
		err := contract.SubmitForSignature()
		require.NoError(t, err)
		assert.Equal(t, ContractStatusPendingSignature, contract.Status)

		// Step 3: Sign
		err = contract.Sign("John Doe", "john@example.com", "sig-123")
		require.NoError(t, err)
		assert.Equal(t, ContractStatusActive, contract.Status)
		assert.True(t, contract.IsActive())

		// Step 4: Complete
		err = contract.Complete()
		require.NoError(t, err)
		assert.Equal(t, ContractStatusCompleted, contract.Status)
		assert.False(t, contract.IsActive())
	})

	t.Run("termination lifecycle: draft -> pending -> active -> terminated", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")

		// Steps 1-3: Same as above
		_ = contract.SubmitForSignature()
		_ = contract.Sign("John Doe", "john@example.com", "sig-123")
		assert.Equal(t, ContractStatusActive, contract.Status)

		// Step 4: Terminate
		err := contract.Terminate("Customer cancelled order")
		require.NoError(t, err)
		assert.Equal(t, ContractStatusTerminated, contract.Status)
		assert.False(t, contract.IsActive())
		assert.False(t, contract.CanBeTerminated())
	})

	t.Run("renewal lifecycle: active -> renewed (version bump)", func(t *testing.T) {
		contract := NewContract(uuidv7.New(), uuidv7.New(), "Terms")
		_ = contract.SubmitForSignature()
		_ = contract.Sign("John Doe", "john@example.com", "sig-123")

		initialVersion := contract.Version
		assert.Equal(t, 1, initialVersion)

		// Renew multiple times
		err := contract.Renew()
		require.NoError(t, err)
		assert.Equal(t, 2, contract.Version)

		err = contract.Renew()
		require.NoError(t, err)
		assert.Equal(t, 3, contract.Version)

		assert.Equal(t, ContractStatusActive, contract.Status)
	})
}
