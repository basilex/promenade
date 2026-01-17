package cashregister

import (
	"time"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CashRegisterStatus represents the operational status
type CashRegisterStatus string

const (
	StatusInactive    CashRegisterStatus = "inactive"
	StatusActive      CashRegisterStatus = "active"
	StatusMaintenance CashRegisterStatus = "maintenance"
	StatusSuspended   CashRegisterStatus = "suspended"
)

// CashRegister is an aggregate root for fiscal cash register management
type CashRegister struct {
	aggregate.BaseAggregate

	OrganizationID         uuidv7.UUID        `db:"organization_id"`
	FiscalNumber           string             `db:"fiscal_number"`
	Model                  string             `db:"model"`
	Status                 CashRegisterStatus `db:"status"`
	LicenseKey             string             `db:"license_key"`
	LastSyncAt             *time.Time         `db:"last_sync_at"`
	ProviderCashRegisterID string             `db:"provider_cash_register_id"`
	ActiveShiftID          string             `db:"active_shift_id"`
	ShiftOpenedAt          *time.Time         `db:"shift_opened_at"`
	ShiftClosedAt          *time.Time         `db:"shift_closed_at"`
	LastZReportID          string             `db:"last_z_report_id"`
	LastZReportAt          *time.Time         `db:"last_z_report_at"`
	LastUpdatedBy          uuidv7.UUID        `db:"last_updated_by"`
}

// NewCashRegister creates a new cash register aggregate
func NewCashRegister(organizationID uuidv7.UUID, fiscalNumber, model string, createdBy uuidv7.UUID) (*CashRegister, error) {
	if organizationID == uuidv7.Nil {
		return nil, ErrOrganizationIDRequired
	}
	if fiscalNumber == "" {
		return nil, ErrFiscalNumberRequired
	}
	if model == "" {
		return nil, ErrModelRequired
	}
	if createdBy == uuidv7.Nil {
		return nil, ErrCreatedByRequired
	}

	return &CashRegister{
		BaseAggregate:  aggregate.NewBaseAggregate(),
		OrganizationID: organizationID,
		FiscalNumber:   fiscalNumber,
		Model:          model,
		Status:         StatusInactive,
		LastUpdatedBy:  createdBy,
	}, nil
}

// Activate activates the cash register with a license key
func (cr *CashRegister) Activate(licenseKey string, activatedBy uuidv7.UUID) error {
	if licenseKey == "" {
		return ErrLicenseKeyRequired
	}
	if cr.Status == StatusActive {
		return ErrCashRegisterAlreadyActive
	}

	cr.Status = StatusActive
	cr.LicenseKey = licenseKey
	cr.LastUpdatedBy = activatedBy
	cr.Touch()

	return nil
}

// Deactivate deactivates the cash register
func (cr *CashRegister) Deactivate(deactivatedBy uuidv7.UUID) error {
	if cr.Status == StatusInactive {
		return ErrCashRegisterAlreadyInactive
	}

	cr.Status = StatusInactive
	cr.LastUpdatedBy = deactivatedBy
	cr.Touch()

	return nil
}

// Suspend suspends the cash register
func (cr *CashRegister) Suspend(suspendedBy uuidv7.UUID) error {
	cr.Status = StatusSuspended
	cr.LastUpdatedBy = suspendedBy
	cr.Touch()

	return nil
}

// UpdateLastSync updates the last synchronization timestamp
func (cr *CashRegister) UpdateLastSync(syncedBy uuidv7.UUID) error {
	now := time.Now()
	cr.LastSyncAt = &now
	cr.LastUpdatedBy = syncedBy
	cr.Touch()

	return nil
}

// OpenShift marks a cash register shift as opened
func (cr *CashRegister) OpenShift(shiftID string, openedBy uuidv7.UUID) error {
	if shiftID == "" {
		return ErrShiftIDRequired
	}
	if cr.ActiveShiftID != "" {
		return ErrShiftAlreadyOpen
	}

	now := time.Now()
	cr.ActiveShiftID = shiftID
	cr.ShiftOpenedAt = &now
	cr.ShiftClosedAt = nil
	cr.LastUpdatedBy = openedBy
	cr.Touch()

	return nil
}

// CloseShift marks a cash register shift as closed and stores Z-report info
func (cr *CashRegister) CloseShift(zReportID string, closedBy uuidv7.UUID) error {
	if cr.ActiveShiftID == "" {
		return ErrNoActiveShift
	}

	now := time.Now()
	cr.ActiveShiftID = ""
	cr.ShiftClosedAt = &now
	cr.LastZReportID = zReportID
	cr.LastZReportAt = &now
	cr.LastUpdatedBy = closedBy
	cr.Touch()

	return nil
}

// SetMaintenance puts the cash register in maintenance mode
func (cr *CashRegister) SetMaintenance(maintainedBy uuidv7.UUID) error {
	cr.Status = StatusMaintenance
	cr.LastUpdatedBy = maintainedBy
	cr.Touch()

	return nil
}

// Validate performs business rule validation
func (cr *CashRegister) Validate() error {
	if cr.OrganizationID == uuidv7.Nil {
		return ErrOrganizationIDRequired
	}
	if cr.FiscalNumber == "" {
		return ErrFiscalNumberRequired
	}
	if cr.Model == "" {
		return ErrModelRequired
	}

	return nil
}
