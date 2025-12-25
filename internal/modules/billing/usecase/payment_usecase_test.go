package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/modules/billing/domain/entity"
	"github.com/basilex/promenade/internal/modules/billing/domain/repository/mocks"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestPaymentUseCase_CreatePayment(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()
	invoiceID := uuidv7.New()
	transactionID := "txn_123456"

	tests := []struct {
		name          string
		userID        uuidv7.UUID
		invoiceID     uuidv7.UUID
		amount        int64
		currency      string
		method        entity.PaymentMethod
		transactionID string
		mockSetup     func(*mocks.MockIPaymentRepository)
		expectError   bool
		errorContains string
	}{
		{
			name:          "successful payment creation - credit card",
			userID:        userID,
			invoiceID:     invoiceID,
			amount:        1000,
			currency:      "USD",
			method:        entity.PaymentMethodCard,
			transactionID: transactionID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Payment")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "successful payment - paypal",
			userID:        userID,
			invoiceID:     invoiceID,
			amount:        5000,
			currency:      "EUR",
			method:        entity.PaymentMethodPayPal,
			transactionID: "paypal_tx_789",
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Payment")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "successful payment - bank transfer",
			userID:        userID,
			invoiceID:     invoiceID,
			amount:        2500,
			currency:      "GBP",
			method:        entity.PaymentMethodBankTransfer,
			transactionID: "bank_tx_xyz",
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Payment")).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "invalid amount - negative",
			userID:        userID,
			invoiceID:     invoiceID,
			amount:        -100,
			currency:      "USD",
			method:        entity.PaymentMethodCard,
			transactionID: transactionID,
			mockSetup:     func(m *mocks.MockIPaymentRepository) {},
			expectError:   true,
			errorContains: "amount",
		},
		{
			name:          "invalid amount - zero",
			userID:        userID,
			invoiceID:     invoiceID,
			amount:        0,
			currency:      "USD",
			method:        entity.PaymentMethodCard,
			transactionID: transactionID,
			mockSetup:     func(m *mocks.MockIPaymentRepository) {},
			expectError:   true,
			errorContains: "amount",
		},
		{
			name:          "repository error",
			userID:        userID,
			invoiceID:     invoiceID,
			amount:        1000,
			currency:      "USD",
			method:        entity.PaymentMethodCard,
			transactionID: transactionID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Payment")).
					Return(errors.New("db error"))
			},
			expectError:   true,
			errorContains: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			payment, err := uc.CreatePayment(ctx, tt.userID, tt.invoiceID, tt.amount, tt.currency, tt.method, tt.transactionID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, payment)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, tt.userID, payment.UserID)
				assert.Equal(t, tt.invoiceID, payment.InvoiceID)
				assert.Equal(t, tt.amount, payment.Amount)
				assert.Equal(t, tt.currency, payment.Currency)
				assert.Equal(t, tt.method, payment.Method)
				assert.Equal(t, entity.PaymentStatusPending, payment.Status)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_GetPayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()

	tests := []struct {
		name        string
		paymentID   uuidv7.UUID
		mockSetup   func(*mocks.MockIPaymentRepository)
		expectError bool
		expectNil   bool
	}{
		{
			name:      "successful retrieval",
			paymentID: paymentID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payment := &entity.Payment{
					ID:            paymentID,
					UserID:        uuidv7.New(),
					InvoiceID:     uuidv7.New(),
					Amount:        1000,
					Currency:      "USD",
					Status:        entity.PaymentStatusCompleted,
					Method: entity.PaymentMethodCard,
				}
				m.On("GetByID", ctx, paymentID).Return(payment, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name:      "payment not found",
			paymentID: paymentID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByID", ctx, paymentID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
		{
			name:      "repository error",
			paymentID: paymentID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByID", ctx, paymentID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			payment, err := uc.GetPayment(ctx, tt.paymentID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, payment)
			} else {
				assert.NotNil(t, payment)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_GetPaymentByTransactionID(t *testing.T) {
	ctx := context.Background()
	transactionID := "txn_123456"

	tests := []struct {
		name          string
		transactionID string
		mockSetup     func(*mocks.MockIPaymentRepository)
		expectError   bool
		expectNil     bool
	}{
		{
			name:          "successful retrieval",
			transactionID: transactionID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payment := &entity.Payment{
					ID:            uuidv7.New(),
					TransactionID: transactionID,
					Amount:        1000,
					Currency:      "USD",
				}
				m.On("GetByTransactionID", ctx, transactionID).Return(payment, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name:          "payment not found",
			transactionID: "unknown_tx",
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByTransactionID", ctx, "unknown_tx").Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			payment, err := uc.GetPaymentByTransactionID(ctx, tt.transactionID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, payment)
			} else {
				assert.NotNil(t, payment)
				assert.Equal(t, tt.transactionID, payment.TransactionID)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_GetInvoicePayments(t *testing.T) {
	ctx := context.Background()
	invoiceID := uuidv7.New()

	tests := []struct {
		name        string
		invoiceID   uuidv7.UUID
		mockSetup   func(*mocks.MockIPaymentRepository)
		expectError bool
		expectCount int
	}{
		{
			name:      "successful retrieval - multiple payments",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payments := []*entity.Payment{
					{ID: uuidv7.New(), InvoiceID: invoiceID, Status: entity.PaymentStatusCompleted},
					{ID: uuidv7.New(), InvoiceID: invoiceID, Status: entity.PaymentStatusPending},
				}
				m.On("GetByInvoiceID", ctx, invoiceID).Return(payments, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:      "no payments found",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByInvoiceID", ctx, invoiceID).Return([]*entity.Payment{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:      "repository error",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByInvoiceID", ctx, invoiceID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			payments, err := uc.GetInvoicePayments(ctx, tt.invoiceID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, payments)
			} else {
				assert.NoError(t, err)
				assert.Len(t, payments, tt.expectCount)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_GetUserPayments(t *testing.T) {
	ctx := context.Background()
	userID := uuidv7.New()

	tests := []struct {
		name        string
		userID      uuidv7.UUID
		mockSetup   func(*mocks.MockIPaymentRepository)
		expectError bool
		expectCount int
	}{
		{
			name:   "successful retrieval - multiple payments",
			userID: userID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payments := []*entity.Payment{
					{ID: uuidv7.New(), UserID: userID, Amount: 1000},
					{ID: uuidv7.New(), UserID: userID, Amount: 2000},
					{ID: uuidv7.New(), UserID: userID, Amount: 500},
				}
				m.On("GetByUserID", ctx, userID).Return(payments, nil)
			},
			expectError: false,
			expectCount: 3,
		},
		{
			name:   "no payments found",
			userID: userID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByUserID", ctx, userID).Return([]*entity.Payment{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "repository error",
			userID: userID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByUserID", ctx, userID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			payments, err := uc.GetUserPayments(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, payments)
			} else {
				assert.NoError(t, err)
				assert.Len(t, payments, tt.expectCount)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_CompletePayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()
	transactionID := "txn_completed_123"

	tests := []struct {
		name          string
		paymentID     uuidv7.UUID
		transactionID string
		mockSetup     func(*mocks.MockIPaymentRepository)
		expectError   bool
	}{
		{
			name:          "successful completion",
			paymentID:     paymentID,
			transactionID: transactionID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payment := &entity.Payment{
					ID:       paymentID,
					Status:   entity.PaymentStatusPending,
					Amount:   1000,
					Currency: "USD",
				}
				m.On("GetByID", ctx, paymentID).Return(payment, nil)
				m.On("Update", ctx, mock.MatchedBy(func(p *entity.Payment) bool {
					return p.Status == entity.PaymentStatusCompleted &&
						p.TransactionID == transactionID &&
						p.CompletedAt != nil
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:          "payment not found",
			paymentID:     paymentID,
			transactionID: transactionID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByID", ctx, paymentID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:          "update error",
			paymentID:     paymentID,
			transactionID: transactionID,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payment := &entity.Payment{
					ID:     paymentID,
					Status: entity.PaymentStatusPending,
				}
				m.On("GetByID", ctx, paymentID).Return(payment, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			err := uc.CompletePayment(ctx, tt.paymentID, tt.transactionID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_RefundPayment(t *testing.T) {
	ctx := context.Background()
	paymentID := uuidv7.New()

	tests := []struct {
		name        string
		paymentID   uuidv7.UUID
		reason      string
		mockSetup   func(*mocks.MockIPaymentRepository)
		expectError bool
	}{
		{
			name:      "successful refund",
			paymentID: paymentID,
			reason:    "Customer requested refund",
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payment := &entity.Payment{
					ID:       paymentID,
					Status:   entity.PaymentStatusCompleted,
					Amount:   1000,
					Currency: "USD",
				}
				m.On("GetByID", ctx, paymentID).Return(payment, nil)
				m.On("Update", ctx, mock.MatchedBy(func(p *entity.Payment) bool {
					return p.Status == entity.PaymentStatusRefunded &&
						p.RefundedAt != nil &&
						p.RefundReason == "Customer requested refund"
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "successful refund - different reason",
			paymentID: paymentID,
			reason:    "Duplicate payment",
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payment := &entity.Payment{
					ID:     paymentID,
					Status: entity.PaymentStatusCompleted,
				}
				m.On("GetByID", ctx, paymentID).Return(payment, nil)
				m.On("Update", ctx, mock.MatchedBy(func(p *entity.Payment) bool {
					return p.Status == entity.PaymentStatusRefunded &&
						p.RefundReason == "Duplicate payment"
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "payment not found",
			paymentID: paymentID,
			reason:    "Refund reason",
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("GetByID", ctx, paymentID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:      "update error",
			paymentID: paymentID,
			reason:    "Refund reason",
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payment := &entity.Payment{
					ID:     paymentID,
					Status: entity.PaymentStatusCompleted,
				}
				m.On("GetByID", ctx, paymentID).Return(payment, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			err := uc.RefundPayment(ctx, tt.paymentID, tt.reason)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_ListPayments(t *testing.T) {
	ctx := context.Background()
	completedStatus := entity.PaymentStatusCompleted
	pendingStatus := entity.PaymentStatusPending

	tests := []struct {
		name        string
		status      *entity.PaymentStatus
		limit       int
		offset      int
		mockSetup   func(*mocks.MockIPaymentRepository)
		expectError bool
		expectCount int
	}{
		{
			name:   "list all payments",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payments := []*entity.Payment{
					{ID: uuidv7.New(), Status: entity.PaymentStatusCompleted},
					{ID: uuidv7.New(), Status: entity.PaymentStatusPending},
				}
				m.On("List", ctx, (*entity.PaymentStatus)(nil), 10, 0).Return(payments, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:   "list completed payments only",
			status: &completedStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payments := []*entity.Payment{
					{ID: uuidv7.New(), Status: entity.PaymentStatusCompleted},
				}
				m.On("List", ctx, &completedStatus, 10, 0).Return(payments, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "list with pagination",
			status: nil,
			limit:  5,
			offset: 10,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				payments := []*entity.Payment{
					{ID: uuidv7.New(), Status: entity.PaymentStatusCompleted},
				}
				m.On("List", ctx, (*entity.PaymentStatus)(nil), 5, 10).Return(payments, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "empty result",
			status: &pendingStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("List", ctx, &pendingStatus, 10, 0).Return([]*entity.Payment{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("List", ctx, (*entity.PaymentStatus)(nil), 10, 0).
					Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			payments, err := uc.ListPayments(ctx, tt.status, tt.limit, tt.offset)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, payments)
			} else {
				assert.NoError(t, err)
				assert.Len(t, payments, tt.expectCount)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentUseCase_CountPayments(t *testing.T) {
	ctx := context.Background()
	completedStatus := entity.PaymentStatusCompleted

	tests := []struct {
		name          string
		status        *entity.PaymentStatus
		mockSetup     func(*mocks.MockIPaymentRepository)
		expectError   bool
		expectedCount int
	}{
		{
			name:   "count all payments",
			status: nil,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Count", ctx, (*entity.PaymentStatus)(nil)).Return(150, nil)
			},
			expectError:   false,
			expectedCount: 150,
		},
		{
			name:   "count completed payments",
			status: &completedStatus,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Count", ctx, &completedStatus).Return(75, nil)
			},
			expectError:   false,
			expectedCount: 75,
		},
		{
			name:   "zero count",
			status: nil,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Count", ctx, (*entity.PaymentStatus)(nil)).Return(0, nil)
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			mockSetup: func(m *mocks.MockIPaymentRepository) {
				m.On("Count", ctx, (*entity.PaymentStatus)(nil)).
					Return(0, errors.New("db error"))
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			tt.mockSetup(mockPaymentRepo)

			uc := NewPaymentUseCase(mockPaymentRepo, mockInvoiceRepo, nil)

			count, err := uc.CountPayments(ctx, tt.status)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}
