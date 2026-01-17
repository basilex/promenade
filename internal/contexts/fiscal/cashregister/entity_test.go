package cashregister

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestNewCashRegister(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()

	tests := []struct {
		name           string
		organizationID uuidv7.UUID
		fiscalNumber   string
		model          string
		createdBy      uuidv7.UUID
		wantErr        error
	}{
		{
			name:           "valid cash register",
			organizationID: orgID,
			fiscalNumber:   "1234567890",
			model:          "Checkbox",
			createdBy:      createdBy,
			wantErr:        nil,
		},
		{
			name:           "missing organization ID",
			organizationID: uuidv7.Nil,
			fiscalNumber:   "1234567890",
			model:          "Checkbox",
			createdBy:      createdBy,
			wantErr:        ErrOrganizationIDRequired,
		},
		{
			name:           "missing fiscal number",
			organizationID: orgID,
			fiscalNumber:   "",
			model:          "Checkbox",
			createdBy:      createdBy,
			wantErr:        ErrFiscalNumberRequired,
		},
		{
			name:           "missing model",
			organizationID: orgID,
			fiscalNumber:   "1234567890",
			model:          "",
			createdBy:      createdBy,
			wantErr:        ErrModelRequired,
		},
		{
			name:           "missing created by",
			organizationID: orgID,
			fiscalNumber:   "1234567890",
			model:          "Checkbox",
			createdBy:      uuidv7.Nil,
			wantErr:        ErrCreatedByRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr, err := NewCashRegister(tt.organizationID, tt.fiscalNumber, tt.model, tt.createdBy)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, cr)
				assert.NotEqual(t, uuidv7.Nil, cr.GetID())
				assert.Equal(t, tt.organizationID, cr.OrganizationID)
				assert.Equal(t, tt.fiscalNumber, cr.FiscalNumber)
				assert.Equal(t, tt.model, cr.Model)
				assert.Equal(t, StatusInactive, cr.Status)
				assert.Equal(t, tt.createdBy, cr.LastUpdatedBy)
			}
		})
	}
}

func TestCashRegister_Activate(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()
	activatedBy := uuidv7.New()

	cr, err := NewCashRegister(orgID, "1234567890", "Checkbox", createdBy)
	require.NoError(t, err)

	tests := []struct {
		name       string
		licenseKey string
		setup      func()
		wantErr    error
	}{
		{
			name:       "successful activation",
			licenseKey: "LICENSE-KEY-123",
			setup:      func() {},
			wantErr:    nil,
		},
		{
			name:       "missing license key",
			licenseKey: "",
			setup:      func() {},
			wantErr:    ErrLicenseKeyRequired,
		},
		{
			name:       "already active",
			licenseKey: "LICENSE-KEY-456",
			setup: func() {
				cr.Status = StatusActive
			},
			wantErr: ErrCashRegisterAlreadyActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to inactive before each test
			cr.Status = StatusInactive
			tt.setup()

			err := cr.Activate(tt.licenseKey, activatedBy)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, StatusActive, cr.Status)
				assert.Equal(t, tt.licenseKey, cr.LicenseKey)
				assert.Equal(t, activatedBy, cr.LastUpdatedBy)
			}
		})
	}
}

func TestCashRegister_Deactivate(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()
	deactivatedBy := uuidv7.New()

	cr, err := NewCashRegister(orgID, "1234567890", "Checkbox", createdBy)
	require.NoError(t, err)

	// Activate first
	err = cr.Activate("LICENSE-KEY", createdBy)
	require.NoError(t, err)

	// Test deactivation
	err = cr.Deactivate(deactivatedBy)
	require.NoError(t, err)
	assert.Equal(t, StatusInactive, cr.Status)
	assert.Equal(t, deactivatedBy, cr.LastUpdatedBy)

	// Try deactivate again
	err = cr.Deactivate(deactivatedBy)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCashRegisterAlreadyInactive)
}

func TestCashRegister_UpdateLastSync(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()
	syncedBy := uuidv7.New()

	cr, err := NewCashRegister(orgID, "1234567890", "Checkbox", createdBy)
	require.NoError(t, err)

	assert.Nil(t, cr.LastSyncAt)

	err = cr.UpdateLastSync(syncedBy)
	require.NoError(t, err)
	assert.NotNil(t, cr.LastSyncAt)
	assert.Equal(t, syncedBy, cr.LastUpdatedBy)
}

func TestCashRegister_OpenShift(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()
	openedBy := uuidv7.New()

	cr, err := NewCashRegister(orgID, "1234567890", "Checkbox", createdBy)
	require.NoError(t, err)

	// Missing shift ID
	err = cr.OpenShift("", openedBy)
	assert.ErrorIs(t, err, ErrShiftIDRequired)

	// Success
	err = cr.OpenShift("shift-1", openedBy)
	require.NoError(t, err)
	assert.Equal(t, "shift-1", cr.ActiveShiftID)
	assert.NotNil(t, cr.ShiftOpenedAt)
	assert.Nil(t, cr.ShiftClosedAt)
	assert.Equal(t, openedBy, cr.LastUpdatedBy)

	// Already open
	err = cr.OpenShift("shift-2", openedBy)
	assert.ErrorIs(t, err, ErrShiftAlreadyOpen)
}

func TestCashRegister_CloseShift(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()
	closedBy := uuidv7.New()

	cr, err := NewCashRegister(orgID, "1234567890", "Checkbox", createdBy)
	require.NoError(t, err)

	// No active shift
	err = cr.CloseShift("z-1", closedBy)
	assert.ErrorIs(t, err, ErrNoActiveShift)

	// Open shift and close
	require.NoError(t, cr.OpenShift("shift-1", createdBy))
	err = cr.CloseShift("z-1", closedBy)
	require.NoError(t, err)
	assert.Equal(t, "", cr.ActiveShiftID)
	assert.NotNil(t, cr.ShiftClosedAt)
	assert.Equal(t, "z-1", cr.LastZReportID)
	assert.NotNil(t, cr.LastZReportAt)
	assert.Equal(t, closedBy, cr.LastUpdatedBy)
}

func TestCashRegister_SetMaintenance(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()
	maintainedBy := uuidv7.New()

	cr, err := NewCashRegister(orgID, "1234567890", "Checkbox", createdBy)
	require.NoError(t, err)

	err = cr.SetMaintenance(maintainedBy)
	require.NoError(t, err)
	assert.Equal(t, StatusMaintenance, cr.Status)
	assert.Equal(t, maintainedBy, cr.LastUpdatedBy)
}

func TestCashRegister_Suspend(t *testing.T) {
	orgID := uuidv7.New()
	createdBy := uuidv7.New()
	suspendedBy := uuidv7.New()

	cr, err := NewCashRegister(orgID, "1234567890", "Checkbox", createdBy)
	require.NoError(t, err)

	err = cr.Suspend(suspendedBy)
	require.NoError(t, err)
	assert.Equal(t, StatusSuspended, cr.Status)
	assert.Equal(t, suspendedBy, cr.LastUpdatedBy)
}

func TestCashRegister_Validate_Errors(t *testing.T) {
	cr := &CashRegister{}
	assert.ErrorIs(t, cr.Validate(), ErrOrganizationIDRequired)

	cr.OrganizationID = uuidv7.New()
	assert.ErrorIs(t, cr.Validate(), ErrFiscalNumberRequired)

	cr.FiscalNumber = "123"
	assert.ErrorIs(t, cr.Validate(), ErrModelRequired)
}

func TestCashRegister_Validate_Success(t *testing.T) {
	cr, err := NewCashRegister(uuidv7.New(), "FN-VALID", "Model", uuidv7.New())
	require.NoError(t, err)

	err = cr.Validate()
	require.NoError(t, err)
}
