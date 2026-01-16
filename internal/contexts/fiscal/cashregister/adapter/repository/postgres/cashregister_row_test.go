package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func TestCashRegisterRow_ToEntityAndFromEntity(t *testing.T) {
	now := time.Now().UTC()
	lastSync := now.Add(-time.Hour)

	row := &cashRegisterRow{
		ID:             uuidv7.New().String(),
		Version:        2,
		OrganizationID: uuidv7.New().String(),
		FiscalNumber:   "FN-123",
		Model:          "Checkbox",
		Status:         string(cashregister.StatusActive),
		LicenseKey:     sql.NullString{String: "LIC-123", Valid: true},
		LastSyncAt:     sql.NullTime{Time: lastSync, Valid: true},
		LastUpdatedBy:  uuidv7.New().String(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	entity, err := row.toEntity()
	require.NoError(t, err)
	require.Equal(t, cashregister.StatusActive, entity.Status)
	require.Equal(t, "LIC-123", entity.LicenseKey)
	require.NotNil(t, entity.LastSyncAt)

	back := fromEntity(entity)
	require.Equal(t, row.ID, back.ID)
	require.Equal(t, row.OrganizationID, back.OrganizationID)
	require.Equal(t, row.FiscalNumber, back.FiscalNumber)
}
