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

func TestInvoiceUseCase_CreateInvoice(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	tests := []struct {
		name           string
		subscriptionID uuidv7.UUID
		subtotal       int64
		tax            int64
		currency       string
		dueDays        int
		mockSetup      func(*mocks.MockIInvoiceRepository)
		expectError    bool
		errorContains  string
	}{
		{
			name:           "successful invoice creation",
			subscriptionID: subscriptionID,
			subtotal:       1000,
			tax:            200,
			currency:       "USD",
			dueDays:        30,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Invoice")).Return(nil)
			},
			expectError: false,
		},
		{
			name:           "successful invoice - zero tax",
			subscriptionID: subscriptionID,
			subtotal:       5000,
			tax:            0,
			currency:       "EUR",
			dueDays:        14,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Invoice")).Return(nil)
			},
			expectError: false,
		},
		{
			name:           "invalid subtotal - negative",
			subscriptionID: subscriptionID,
			subtotal:       -100,
			tax:            0,
			currency:       "USD",
			dueDays:        30,
			mockSetup:      func(m *mocks.MockIInvoiceRepository) {},
			expectError:    true,
			errorContains:  "subtotal",
		},
		{
			name:           "invalid tax - negative",
			subscriptionID: subscriptionID,
			subtotal:       1000,
			tax:            -50,
			currency:       "USD",
			dueDays:        30,
			mockSetup:      func(m *mocks.MockIInvoiceRepository) {},
			expectError:    true,
			errorContains:  "tax",
		},
		{
			name:           "repository error",
			subscriptionID: subscriptionID,
			subtotal:       1000,
			tax:            100,
			currency:       "USD",
			dueDays:        30,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("Create", ctx, mock.AnythingOfType("*entity.Invoice")).
					Return(errors.New("db error"))
			},
			expectError:   true,
			errorContains: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			invoice, err := uc.CreateInvoice(ctx, tt.subscriptionID, tt.subtotal, tt.tax, tt.currency, tt.dueDays)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, invoice)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, invoice)
				assert.Equal(t, tt.subscriptionID, invoice.SubscriptionID)
				assert.Equal(t, tt.subtotal, invoice.SubtotalAmount)
				assert.Equal(t, tt.tax, invoice.TaxAmount)
				assert.Equal(t, tt.currency, invoice.Currency)
				assert.Equal(t, tt.subtotal+tt.tax, invoice.TotalAmount)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}

func TestInvoiceUseCase_GetInvoice(t *testing.T) {
	ctx := context.Background()
	invoiceID := uuidv7.New()

	tests := []struct {
		name        string
		invoiceID   uuidv7.UUID
		mockSetup   func(*mocks.MockIInvoiceRepository)
		expectError bool
		expectNil   bool
	}{
		{
			name:      "successful retrieval",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoice := &entity.Invoice{
					ID:             invoiceID,
					SubscriptionID: uuidv7.New(),
					SubtotalAmount: 1000,
					TaxAmount:      100,
					TotalAmount:    1100,
					Currency:       "USD",
					Status:         entity.InvoiceStatusDraft,
				}
				m.On("GetByID", ctx, invoiceID).Return(invoice, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name:      "invoice not found",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("GetByID", ctx, invoiceID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
		{
			name:      "repository error",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("GetByID", ctx, invoiceID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			invoice, err := uc.GetInvoice(ctx, tt.invoiceID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, invoice)
			} else {
				assert.NotNil(t, invoice)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}

func TestInvoiceUseCase_GetInvoiceByNumber(t *testing.T) {
	ctx := context.Background()
	invoiceNumber := "INV-2025-001"

	tests := []struct {
		name          string
		invoiceNumber string
		mockSetup     func(*mocks.MockIInvoiceRepository)
		expectError   bool
		expectNil     bool
	}{
		{
			name:          "successful retrieval",
			invoiceNumber: invoiceNumber,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoice := &entity.Invoice{
					ID:             uuidv7.New(),
					InvoiceNumber:  invoiceNumber,
					SubtotalAmount: 1000,
					Currency:       "USD",
				}
				m.On("GetByInvoiceNumber", ctx, invoiceNumber).Return(invoice, nil)
			},
			expectError: false,
			expectNil:   false,
		},
		{
			name:          "invoice not found",
			invoiceNumber: "INV-UNKNOWN",
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("GetByInvoiceNumber", ctx, "INV-UNKNOWN").Return(nil, entity.ErrNotFound)
			},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			invoice, err := uc.GetInvoiceByNumber(ctx, tt.invoiceNumber)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, invoice)
			} else {
				assert.NotNil(t, invoice)
				assert.Equal(t, tt.invoiceNumber, invoice.InvoiceNumber)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}

func TestInvoiceUseCase_GetSubscriptionInvoices(t *testing.T) {
	ctx := context.Background()
	subscriptionID := uuidv7.New()

	tests := []struct {
		name           string
		subscriptionID uuidv7.UUID
		mockSetup      func(*mocks.MockIInvoiceRepository)
		expectError    bool
		expectCount    int
	}{
		{
			name:           "successful retrieval - multiple invoices",
			subscriptionID: subscriptionID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoices := []*entity.Invoice{
					{ID: uuidv7.New(), SubscriptionID: subscriptionID, Status: entity.InvoiceStatusPaid},
					{ID: uuidv7.New(), SubscriptionID: subscriptionID, Status: entity.InvoiceStatusOpen},
				}
				m.On("GetBySubscriptionID", ctx, subscriptionID).Return(invoices, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:           "no invoices found",
			subscriptionID: subscriptionID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("GetBySubscriptionID", ctx, subscriptionID).Return([]*entity.Invoice{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:           "repository error",
			subscriptionID: subscriptionID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("GetBySubscriptionID", ctx, subscriptionID).Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			invoices, err := uc.GetSubscriptionInvoices(ctx, tt.subscriptionID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, invoices)
			} else {
				assert.NoError(t, err)
				assert.Len(t, invoices, tt.expectCount)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}

func TestInvoiceUseCase_FinalizeInvoice(t *testing.T) {
	ctx := context.Background()
	invoiceID := uuidv7.New()

	tests := []struct {
		name        string
		invoiceID   uuidv7.UUID
		mockSetup   func(*mocks.MockIInvoiceRepository)
		expectError bool
	}{
		{
			name:      "successful finalization",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoice := &entity.Invoice{
					ID:             invoiceID,
					Status:         entity.InvoiceStatusDraft,
					SubtotalAmount: 1000,
					TaxAmount:      100,
					TotalAmount:    1100,
				}
				m.On("GetByID", ctx, invoiceID).Return(invoice, nil)
				m.On("Update", ctx, mock.MatchedBy(func(inv *entity.Invoice) bool {
					return inv.Status == entity.InvoiceStatusOpen
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "invoice not found",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("GetByID", ctx, invoiceID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:      "update error",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoice := &entity.Invoice{
					ID:     invoiceID,
					Status: entity.InvoiceStatusDraft,
				}
				m.On("GetByID", ctx, invoiceID).Return(invoice, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			err := uc.FinalizeInvoice(ctx, tt.invoiceID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}

func TestInvoiceUseCase_VoidInvoice(t *testing.T) {
	ctx := context.Background()
	invoiceID := uuidv7.New()

	tests := []struct {
		name        string
		invoiceID   uuidv7.UUID
		mockSetup   func(*mocks.MockIInvoiceRepository)
		expectError bool
	}{
		{
			name:      "successful void",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoice := &entity.Invoice{
					ID:             invoiceID,
					Status:         entity.InvoiceStatusOpen,
					SubtotalAmount: 1000,
				}
				m.On("GetByID", ctx, invoiceID).Return(invoice, nil)
				m.On("Update", ctx, mock.MatchedBy(func(inv *entity.Invoice) bool {
					return inv.Status == entity.InvoiceStatusVoid
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "invoice not found",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("GetByID", ctx, invoiceID).Return(nil, entity.ErrNotFound)
			},
			expectError: true,
		},
		{
			name:      "update error",
			invoiceID: invoiceID,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoice := &entity.Invoice{
					ID:     invoiceID,
					Status: entity.InvoiceStatusOpen,
				}
				m.On("GetByID", ctx, invoiceID).Return(invoice, nil)
				m.On("Update", ctx, mock.Anything).Return(errors.New("update failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			err := uc.VoidInvoice(ctx, tt.invoiceID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}

func TestInvoiceUseCase_ListInvoices(t *testing.T) {
	ctx := context.Background()
	openStatus := entity.InvoiceStatusOpen
	paidStatus := entity.InvoiceStatusPaid

	tests := []struct {
		name        string
		status      *entity.InvoiceStatus
		limit       int
		offset      int
		mockSetup   func(*mocks.MockIInvoiceRepository)
		expectError bool
		expectCount int
	}{
		{
			name:   "list all invoices",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoices := []*entity.Invoice{
					{ID: uuidv7.New(), Status: entity.InvoiceStatusOpen},
					{ID: uuidv7.New(), Status: entity.InvoiceStatusPaid},
				}
				m.On("List", ctx, (*entity.InvoiceStatus)(nil), 10, 0).Return(invoices, nil)
			},
			expectError: false,
			expectCount: 2,
		},
		{
			name:   "list open invoices only",
			status: &openStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoices := []*entity.Invoice{
					{ID: uuidv7.New(), Status: entity.InvoiceStatusOpen},
				}
				m.On("List", ctx, &openStatus, 10, 0).Return(invoices, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "list with pagination",
			status: nil,
			limit:  5,
			offset: 10,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				invoices := []*entity.Invoice{
					{ID: uuidv7.New(), Status: entity.InvoiceStatusOpen},
				}
				m.On("List", ctx, (*entity.InvoiceStatus)(nil), 5, 10).Return(invoices, nil)
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name:   "empty result",
			status: &paidStatus,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("List", ctx, &paidStatus, 10, 0).Return([]*entity.Invoice{}, nil)
			},
			expectError: false,
			expectCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			limit:  10,
			offset: 0,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("List", ctx, (*entity.InvoiceStatus)(nil), 10, 0).
					Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			invoices, err := uc.ListInvoices(ctx, tt.status, tt.limit, tt.offset)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, invoices)
			} else {
				assert.NoError(t, err)
				assert.Len(t, invoices, tt.expectCount)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}

func TestInvoiceUseCase_CountInvoices(t *testing.T) {
	ctx := context.Background()
	openStatus := entity.InvoiceStatusOpen

	tests := []struct {
		name          string
		status        *entity.InvoiceStatus
		mockSetup     func(*mocks.MockIInvoiceRepository)
		expectError   bool
		expectedCount int
	}{
		{
			name:   "count all invoices",
			status: nil,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("Count", ctx, (*entity.InvoiceStatus)(nil)).Return(100, nil)
			},
			expectError:   false,
			expectedCount: 100,
		},
		{
			name:   "count open invoices",
			status: &openStatus,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("Count", ctx, &openStatus).Return(25, nil)
			},
			expectError:   false,
			expectedCount: 25,
		},
		{
			name:   "zero count",
			status: nil,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("Count", ctx, (*entity.InvoiceStatus)(nil)).Return(0, nil)
			},
			expectError:   false,
			expectedCount: 0,
		},
		{
			name:   "repository error",
			status: nil,
			mockSetup: func(m *mocks.MockIInvoiceRepository) {
				m.On("Count", ctx, (*entity.InvoiceStatus)(nil)).
					Return(0, errors.New("db error"))
			},
			expectError:   true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo := new(mocks.MockIInvoiceRepository)
			mockPaymentRepo := new(mocks.MockIPaymentRepository)
			mockSubRepo := new(mocks.MockISubscriptionRepository)
			tt.mockSetup(mockInvoiceRepo)

			uc := NewInvoiceUseCase(mockInvoiceRepo, mockPaymentRepo, mockSubRepo, nil)

			count, err := uc.CountInvoices(ctx, tt.status)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockInvoiceRepo.AssertExpectations(t)
		})
	}
}
