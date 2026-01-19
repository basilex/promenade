package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cashregisteraggregate "github.com/basilex/promenade/internal/contexts/fiscal/cashregister/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister/repository"
	"github.com/basilex/promenade/pkg/fiscal/checkbox"
	"github.com/basilex/promenade/pkg/scheduler"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type mockCashRegisterRepository struct {
	ListFunc   func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error)
	UpdateFunc func(ctx context.Context, cr *cashregisteraggregate.CashRegister) error
}

func (m *mockCashRegisterRepository) Create(ctx context.Context, cr *cashregisteraggregate.CashRegister) error {
	return nil
}

func (m *mockCashRegisterRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *mockCashRegisterRepository) GetByFiscalNumber(ctx context.Context, fiscalNumber string) (*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *mockCashRegisterRepository) GetByLocation(ctx context.Context, locationID uuidv7.UUID) ([]*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *mockCashRegisterRepository) ListActive(ctx context.Context) ([]*cashregisteraggregate.CashRegister, error) {
	return nil, nil
}

func (m *mockCashRegisterRepository) List(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters)
	}
	return nil, nil
}

func (m *mockCashRegisterRepository) Update(ctx context.Context, cr *cashregisteraggregate.CashRegister) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, cr)
	}
	return nil
}

func (m *mockCashRegisterRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	return nil
}

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

func TestRegisterShiftJobs_DefaultCron(t *testing.T) {
	engine, err := scheduler.NewEngine(scheduler.DefaultConfig())
	require.NoError(t, err)
	require.NoError(t, engine.Start())
	defer func() {
		_ = engine.Stop()
	}()

	repo := &mockCashRegisterRepository{}
	client := &mockShiftClient{}

	require.NoError(t, RegisterShiftJobs(engine, repo, client, "", ""))

	jobs := engine.ListJobs()
	require.Len(t, jobs, 2)

	jobByName := map[string]*scheduler.Job{}
	for _, job := range jobs {
		jobByName[job.Name] = job
	}

	openJob := jobByName[shiftOpenJobName]
	require.NotNil(t, openJob)
	assert.Equal(t, defaultShiftOpenCron, openJob.CronExpression)
	assert.True(t, openJob.Enabled)

	closeJob := jobByName[shiftCloseJobName]
	require.NotNil(t, closeJob)
	assert.Equal(t, defaultShiftCloseCron, closeJob.CronExpression)
	assert.True(t, closeJob.Enabled)
}

func TestShiftOpenExecutor_Execute_OpensShiftAndUpdates(t *testing.T) {
	cr, err := cashregisteraggregate.NewCashRegister(uuidv7.New(), "FN-001", "Model", uuidv7.New())
	require.NoError(t, err)
	require.NoError(t, cr.Activate("LIC-001", uuidv7.New()))
	cr.ProviderCashRegisterID = "cr-1"

	updated := 0
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{cr}, nil
		},
		UpdateFunc: func(ctx context.Context, updatedCR *cashregisteraggregate.CashRegister) error {
			updated++
			assert.Equal(t, "shift-1", updatedCR.ActiveShiftID)
			assert.NotNil(t, updatedCR.ShiftOpenedAt)
			assert.Nil(t, updatedCR.ShiftClosedAt)
			assert.NotEqual(t, uuidv7.Nil, updatedCR.LastUpdatedBy)
			return nil
		},
	}

	client := &mockShiftClient{
		OpenShiftFunc: func(ctx context.Context, cashRegisterID string) (*checkbox.ShiftResponse, error) {
			require.Equal(t, "cr-1", cashRegisterID)
			return &checkbox.ShiftResponse{ID: "shift-1", Status: "open"}, nil
		},
	}

	executor := NewShiftOpenExecutor(repo, client)
	err = executor.Execute(context.Background(), &scheduler.Job{Name: shiftOpenJobName})
	require.NoError(t, err)
	assert.Equal(t, 1, updated)
}

func TestShiftOpenExecutor_Execute_SkipsAlreadyOpen(t *testing.T) {
	cr, err := cashregisteraggregate.NewCashRegister(uuidv7.New(), "FN-002", "Model", uuidv7.New())
	require.NoError(t, err)
	cr.ActiveShiftID = "shift-existing"

	updated := 0
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{cr}, nil
		},
		UpdateFunc: func(ctx context.Context, updatedCR *cashregisteraggregate.CashRegister) error {
			updated++
			return nil
		},
	}

	client := &mockShiftClient{}

	executor := NewShiftOpenExecutor(repo, client)
	err = executor.Execute(context.Background(), &scheduler.Job{Name: shiftOpenJobName})
	require.NoError(t, err)
	assert.Equal(t, 0, updated)
}

func TestShiftOpenExecutor_Execute_ListError(t *testing.T) {
	expectedErr := errors.New("list error")
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return nil, expectedErr
		},
	}

	client := &mockShiftClient{}

	executor := NewShiftOpenExecutor(repo, client)
	err := executor.Execute(context.Background(), &scheduler.Job{Name: shiftOpenJobName})
	require.ErrorIs(t, err, expectedErr)
}

func TestShiftOpenExecutor_Execute_OpenShiftErrorDoesNotFail(t *testing.T) {
	cr, err := cashregisteraggregate.NewCashRegister(uuidv7.New(), "FN-003", "Model", uuidv7.New())
	require.NoError(t, err)
	require.NoError(t, cr.Activate("LIC-003", uuidv7.New()))
	cr.ProviderCashRegisterID = "cr-3"

	updated := 0
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{cr}, nil
		},
		UpdateFunc: func(ctx context.Context, updatedCR *cashregisteraggregate.CashRegister) error {
			updated++
			return nil
		},
	}

	client := &mockShiftClient{
		OpenShiftFunc: func(ctx context.Context, cashRegisterID string) (*checkbox.ShiftResponse, error) {
			return nil, errors.New("shift open failed")
		},
	}

	executor := NewShiftOpenExecutor(repo, client)
	err = executor.Execute(context.Background(), &scheduler.Job{Name: shiftOpenJobName})
	require.NoError(t, err)
	assert.Equal(t, 0, updated)
}

func TestShiftCloseExecutor_Execute_ClosesShiftAndUpdates(t *testing.T) {
	cr, err := cashregisteraggregate.NewCashRegister(uuidv7.New(), "FN-004", "Model", uuidv7.New())
	require.NoError(t, err)
	require.NoError(t, cr.Activate("LIC-004", uuidv7.New()))
	cr.ActiveShiftID = "shift-1"

	updated := 0
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{cr}, nil
		},
		UpdateFunc: func(ctx context.Context, updatedCR *cashregisteraggregate.CashRegister) error {
			updated++
			assert.Equal(t, "", updatedCR.ActiveShiftID)
			assert.NotNil(t, updatedCR.ShiftClosedAt)
			assert.Equal(t, "z-1", updatedCR.LastZReportID)
			assert.NotNil(t, updatedCR.LastZReportAt)
			assert.NotEqual(t, uuidv7.Nil, updatedCR.LastUpdatedBy)
			return nil
		},
	}

	client := &mockShiftClient{
		CloseShiftFunc: func(ctx context.Context, shiftID string) (*checkbox.ZReport, error) {
			require.Equal(t, "shift-1", shiftID)
			return &checkbox.ZReport{ID: "z-1", ShiftID: "shift-1", Number: 1}, nil
		},
	}

	executor := NewShiftCloseExecutor(repo, client)
	err = executor.Execute(context.Background(), &scheduler.Job{Name: shiftCloseJobName})
	require.NoError(t, err)
	assert.Equal(t, 1, updated)
}

func TestShiftCloseExecutor_Execute_SkipsNoActiveShift(t *testing.T) {
	cr, err := cashregisteraggregate.NewCashRegister(uuidv7.New(), "FN-005", "Model", uuidv7.New())
	require.NoError(t, err)

	updated := 0
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{cr}, nil
		},
		UpdateFunc: func(ctx context.Context, updatedCR *cashregisteraggregate.CashRegister) error {
			updated++
			return nil
		},
	}

	client := &mockShiftClient{}

	executor := NewShiftCloseExecutor(repo, client)
	err = executor.Execute(context.Background(), &scheduler.Job{Name: shiftCloseJobName})
	require.NoError(t, err)
	assert.Equal(t, 0, updated)
}

func TestShiftCloseExecutor_Execute_ListError(t *testing.T) {
	expectedErr := errors.New("list error")
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return nil, expectedErr
		},
	}

	client := &mockShiftClient{}

	executor := NewShiftCloseExecutor(repo, client)
	err := executor.Execute(context.Background(), &scheduler.Job{Name: shiftCloseJobName})
	require.ErrorIs(t, err, expectedErr)
}

func TestShiftCloseExecutor_Execute_CloseShiftErrorDoesNotFail(t *testing.T) {
	cr, err := cashregisteraggregate.NewCashRegister(uuidv7.New(), "FN-006", "Model", uuidv7.New())
	require.NoError(t, err)
	require.NoError(t, cr.Activate("LIC-006", uuidv7.New()))
	cr.ActiveShiftID = "shift-6"

	updated := 0
	repo := &mockCashRegisterRepository{
		ListFunc: func(ctx context.Context, filters *repository.ListFilters) ([]*cashregisteraggregate.CashRegister, error) {
			return []*cashregisteraggregate.CashRegister{cr}, nil
		},
		UpdateFunc: func(ctx context.Context, updatedCR *cashregisteraggregate.CashRegister) error {
			updated++
			return nil
		},
	}

	client := &mockShiftClient{
		CloseShiftFunc: func(ctx context.Context, shiftID string) (*checkbox.ZReport, error) {
			return nil, errors.New("close shift failed")
		},
	}

	executor := NewShiftCloseExecutor(repo, client)
	err = executor.Execute(context.Background(), &scheduler.Job{Name: shiftCloseJobName})
	require.NoError(t, err)
	assert.Equal(t, 0, updated)
}
