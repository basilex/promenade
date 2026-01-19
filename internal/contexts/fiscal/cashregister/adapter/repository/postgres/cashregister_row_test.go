package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestCashRegisterRow_ToEntityAndFromEntity(t *testing.T) {
	now := time.Now().UTC()
	lastSync := now.Add(-time.Hour)

	row := &cashRegisterRow{
		ID:                     uuidv7.New().String(),
		Version:                2,
		OrganizationID:         uuidv7.New().String(),
		FiscalNumber:           "FN-123",
		Model:                  "Checkbox",
		Status:                 string(aggregate.StatusActive),
		LicenseKey:             sql.NullString{String: "LIC-123", Valid: true},
		LastSyncAt:             sql.NullTime{Time: lastSync, Valid: true},
		ProviderCashRegisterID: sql.NullString{String: "provider-1", Valid: true},
		ActiveShiftID:          sql.NullString{String: "shift-1", Valid: true},
		ShiftOpenedAt:          sql.NullTime{Time: lastSync, Valid: true},
		ShiftClosedAt:          sql.NullTime{},
		LastZReportID:          sql.NullString{String: "z-1", Valid: true},
		LastZReportAt:          sql.NullTime{Time: now, Valid: true},
		LastUpdatedBy:          uuidv7.New().String(),
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	entity, err := row.toEntity()
	require.NoError(t, err)
	require.Equal(t, aggregate.StatusActive, entity.Status)
	require.Equal(t, "LIC-123", entity.LicenseKey)
	require.NotNil(t, entity.LastSyncAt)
	require.Equal(t, "provider-1", entity.ProviderCashRegisterID)
	require.Equal(t, "shift-1", entity.ActiveShiftID)
	require.NotNil(t, entity.ShiftOpenedAt)
	require.Equal(t, "z-1", entity.LastZReportID)

	back := fromEntity(entity)
	require.Equal(t, row.ID, back.ID)
	require.Equal(t, row.OrganizationID, back.OrganizationID)
	require.Equal(t, row.FiscalNumber, back.FiscalNumber)
}
