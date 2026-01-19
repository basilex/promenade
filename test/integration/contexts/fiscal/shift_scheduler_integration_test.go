package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	crRepo "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/adapter/repository/postgres"
	cashregisterAggregate "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	fiscalIntegration "github.com/basilex/promenade/internal/contexts/fiscal/integration"
	"github.com/basilex/promenade/pkg/fiscal/checkbox"
	"github.com/basilex/promenade/pkg/scheduler"
	"github.com/basilex/promenade/pkg/uuidv7"
	testutils "github.com/basilex/promenade/test/integration"
)

type mockShiftClient struct {
	OpenShiftFunc  func(ctx context.Context, cashRegisterID string) (*checkbox.ShiftResponse, error)
	CloseShiftFunc func(ctx context.Context, shiftID string) (*checkbox.ZReport, error)
}

func (m *mockShiftClient) OpenShift(ctx context.Context, cashRegisterID string) (*checkbox.ShiftResponse, error) {
	if m.OpenShiftFunc != nil {
		return m.OpenShiftFunc(ctx, cashRegisterID)
	}
	return nil, nil
}

func (m *mockShiftClient) CloseShift(ctx context.Context, shiftID string) (*checkbox.ZReport, error) {
	if m.CloseShiftFunc != nil {
		return m.CloseShiftFunc(ctx, shiftID)
	}
	return nil, nil
}

func TestShiftOpenExecutor_Integration_OpensShift(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutils.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	repo := crRepo.NewCashRegisterRepository(db.DB)

	createdBy := uuidv7.New()
	cr, err := cashregisterAggregate.NewCashRegister(uuidv7.New(), "FN-OPEN", "Model", createdBy)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, cr))

	require.NoError(t, cr.Activate("LIC-OPEN", createdBy))
	cr.ProviderCashRegisterID = "provider-1"
	require.NoError(t, repo.Update(ctx, cr))

	client := &mockShiftClient{
		OpenShiftFunc: func(ctx context.Context, cashRegisterID string) (*checkbox.ShiftResponse, error) {
			require.Equal(t, "provider-1", cashRegisterID)
			return &checkbox.ShiftResponse{ID: "shift-1", Status: "open"}, nil
		},
	}

	executor := fiscalIntegration.NewShiftOpenExecutor(repo, client)
	require.NoError(t, executor.Execute(ctx, &scheduler.Job{Name: "test-open"}))

	updated, err := repo.GetByID(ctx, cr.GetID())
	require.NoError(t, err)
	require.Equal(t, "shift-1", updated.ActiveShiftID)
	require.NotNil(t, updated.ShiftOpenedAt)
	require.Nil(t, updated.ShiftClosedAt)
}

func TestShiftCloseExecutor_Integration_ClosesShift(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := testutils.SetupTestDBWithCleanTables(t)
	ctx := context.Background()

	repo := crRepo.NewCashRegisterRepository(db.DB)

	createdBy := uuidv7.New()
	cr, err := cashregisterAggregate.NewCashRegister(uuidv7.New(), "FN-CLOSE", "Model", createdBy)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, cr))

	require.NoError(t, cr.Activate("LIC-CLOSE", createdBy))
	require.NoError(t, cr.OpenShift("shift-2", createdBy))
	require.NoError(t, repo.Update(ctx, cr))

	client := &mockShiftClient{
		CloseShiftFunc: func(ctx context.Context, shiftID string) (*checkbox.ZReport, error) {
			require.Equal(t, "shift-2", shiftID)
			return &checkbox.ZReport{ID: "z-2", ShiftID: "shift-2", Number: 2}, nil
		},
	}

	executor := fiscalIntegration.NewShiftCloseExecutor(repo, client)
	require.NoError(t, executor.Execute(ctx, &scheduler.Job{Name: "test-close"}))

	updated, err := repo.GetByID(ctx, cr.GetID())
	require.NoError(t, err)
	require.Equal(t, "", updated.ActiveShiftID)
	require.NotNil(t, updated.ShiftClosedAt)
	require.Equal(t, "z-2", updated.LastZReportID)
	require.NotNil(t, updated.LastZReportAt)
}
