package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock ITimezoneRepository
type mockTimezoneRepository struct {
	mock.Mock
}

func (m *mockTimezoneRepository) Create(ctx context.Context, timezone *entity.Timezone) error {
	args := m.Called(ctx, timezone)
	return args.Error(0)
}

func (m *mockTimezoneRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Timezone, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Timezone), args.Error(1)
}

func (m *mockTimezoneRepository) GetByName(ctx context.Context, name string) (*entity.Timezone, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Timezone), args.Error(1)
}

func (m *mockTimezoneRepository) List(ctx context.Context, params pagination.Params) ([]*entity.Timezone, *pagination.Metadata, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Timezone), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *mockTimezoneRepository) ListActive(ctx context.Context) ([]*entity.Timezone, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Timezone), args.Error(1)
}

func (m *mockTimezoneRepository) Update(ctx context.Context, timezone *entity.Timezone) error {
	args := m.Called(ctx, timezone)
	return args.Error(0)
}

func (m *mockTimezoneRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTimezoneRepository) Exists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

// Test helpers
func setupTimezoneUseCase(t *testing.T) (*timezoneUseCase, *mockTimezoneRepository) {
	timezoneRepo := new(mockTimezoneRepository)
	uc := NewTimezoneUseCase(timezoneRepo).(*timezoneUseCase)
	return uc, timezoneRepo
}

func createTestTimezone(name, abbreviation string) *entity.Timezone {
	return &entity.Timezone{
		ID:           uuidv7.New(),
		Name:         name,
		Abbreviation: abbreviation,
		UtcOffset:    "+00:00",
		IsActive:     true,
	}
}

// Tests for Create
func TestTimezoneUseCase_Create_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone := createTestTimezone("UTC", "UTC")
	timezoneRepo.On("Exists", ctx, "UTC").Return(false, nil)
	timezoneRepo.On("Create", ctx, timezone).Return(nil)

	err := uc.Create(ctx, timezone)

	assert.NoError(t, err)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Create_GeneratesUUID(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone := createTestTimezone("UTC", "UTC")
	timezone.ID = uuidv7.UUID{} // Empty UUID

	timezoneRepo.On("Exists", ctx, "UTC").Return(false, nil)
	timezoneRepo.On("Create", ctx, mock.AnythingOfType("*entity.Timezone")).Return(nil)

	err := uc.Create(ctx, timezone)

	assert.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, timezone.ID)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Create_AlreadyExists(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone := createTestTimezone("UTC", "UTC")
	timezoneRepo.On("Exists", ctx, "UTC").Return(true, nil)

	err := uc.Create(ctx, timezone)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Create_ValidationError(t *testing.T) {
	uc, _ := setupTimezoneUseCase(t)
	ctx := context.Background()

	// Timezone with invalid data
	timezone := &entity.Timezone{
		ID:           uuidv7.New(),
		Name:         "", // Empty name should fail validation
		Abbreviation: "UTC",
		UtcOffset:    "+00:00",
	}

	err := uc.Create(ctx, timezone)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

// Tests for GetByID
func TestTimezoneUseCase_GetByID_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	expectedTimezone := createTestTimezone("UTC", "UTC")
	timezoneRepo.On("GetByID", ctx, expectedTimezone.ID).Return(expectedTimezone, nil)

	timezone, err := uc.GetByID(ctx, expectedTimezone.ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTimezone.ID, timezone.ID)
	assert.Equal(t, "UTC", timezone.Name)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_GetByID_NotFound(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezoneID := uuidv7.New()
	timezoneRepo.On("GetByID", ctx, timezoneID).Return(nil, entity.ErrNotFound)

	timezone, err := uc.GetByID(ctx, timezoneID)

	assert.Error(t, err)
	assert.Nil(t, timezone)
	timezoneRepo.AssertExpectations(t)
}

// Tests for GetByName
func TestTimezoneUseCase_GetByName_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	expectedTimezone := createTestTimezone("Europe/Moscow", "MSK")
	timezoneRepo.On("GetByName", ctx, "Europe/Moscow").Return(expectedTimezone, nil)

	timezone, err := uc.GetByName(ctx, "Europe/Moscow")

	assert.NoError(t, err)
	assert.Equal(t, "Europe/Moscow", timezone.Name)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_GetByName_NotFound(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezoneRepo.On("GetByName", ctx, "Invalid/Zone").Return(nil, entity.ErrNotFound)

	timezone, err := uc.GetByName(ctx, "Invalid/Zone")

	assert.Error(t, err)
	assert.Nil(t, timezone)
	timezoneRepo.AssertExpectations(t)
}

// Tests for List
func TestTimezoneUseCase_List_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone1 := createTestTimezone("UTC", "UTC")
	timezone2 := createTestTimezone("Europe/Moscow", "MSK")
	expectedTimezones := []*entity.Timezone{timezone1, timezone2}
	metadata := &pagination.Metadata{
		Total:       2,
		Limit:       20,
		Offset:      0,
		CurrentPage: 1,
		TotalPages:  1,
	}

	params := pagination.Params{Page: 1, PageSize: 20}
	timezoneRepo.On("List", ctx, params).Return(expectedTimezones, metadata, nil)

	timezones, meta, err := uc.List(ctx, params)

	assert.NoError(t, err)
	assert.Len(t, timezones, 2)
	assert.Equal(t, 2, meta.Total)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_List_DefaultPagination(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	expectedTimezones := []*entity.Timezone{}
	metadata := &pagination.Metadata{
		Total:       0,
		Limit:       20,
		Offset:      0,
		CurrentPage: 1,
		TotalPages:  0,
	}

	params := pagination.Params{Page: 0, PageSize: 0}
	normalizedParams := pagination.Params{Page: 1, PageSize: 20}
	timezoneRepo.On("List", ctx, normalizedParams).Return(expectedTimezones, metadata, nil)

	timezones, meta, err := uc.List(ctx, params)

	assert.NoError(t, err)
	assert.Empty(t, timezones)
	assert.Equal(t, 0, meta.Total)
	timezoneRepo.AssertExpectations(t)
}

// Tests for ListActive
func TestTimezoneUseCase_ListActive_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone1 := createTestTimezone("UTC", "UTC")
	timezone2 := createTestTimezone("Europe/Moscow", "MSK")
	expectedTimezones := []*entity.Timezone{timezone1, timezone2}

	timezoneRepo.On("ListActive", ctx).Return(expectedTimezones, nil)

	timezones, err := uc.ListActive(ctx)

	assert.NoError(t, err)
	assert.Len(t, timezones, 2)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_ListActive_Empty(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezoneRepo.On("ListActive", ctx).Return([]*entity.Timezone{}, nil)

	timezones, err := uc.ListActive(ctx)

	assert.NoError(t, err)
	assert.Empty(t, timezones)
	timezoneRepo.AssertExpectations(t)
}

// Tests for Update
func TestTimezoneUseCase_Update_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone := createTestTimezone("UTC", "UTC")
	timezone.Abbreviation = "UTC+0"

	timezoneRepo.On("GetByID", ctx, timezone.ID).Return(timezone, nil)
	timezoneRepo.On("Update", ctx, timezone).Return(nil)

	err := uc.Update(ctx, timezone)

	assert.NoError(t, err)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Update_NotFound(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone := createTestTimezone("UTC", "UTC")
	timezoneRepo.On("GetByID", ctx, timezone.ID).Return(nil, entity.ErrNotFound)

	err := uc.Update(ctx, timezone)

	assert.Error(t, err)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Update_NameChanged_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	existingTimezone := createTestTimezone("UTC", "UTC")
	updatedTimezone := createTestTimezone("Etc/UTC", "UTC")
	updatedTimezone.ID = existingTimezone.ID

	timezoneRepo.On("GetByID", ctx, updatedTimezone.ID).Return(existingTimezone, nil)
	timezoneRepo.On("Exists", ctx, "Etc/UTC").Return(false, nil)
	timezoneRepo.On("Update", ctx, updatedTimezone).Return(nil)

	err := uc.Update(ctx, updatedTimezone)

	assert.NoError(t, err)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Update_NameChanged_AlreadyExists(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	existingTimezone := createTestTimezone("UTC", "UTC")
	updatedTimezone := createTestTimezone("Europe/Moscow", "MSK")
	updatedTimezone.ID = existingTimezone.ID

	timezoneRepo.On("GetByID", ctx, updatedTimezone.ID).Return(existingTimezone, nil)
	timezoneRepo.On("Exists", ctx, "Europe/Moscow").Return(true, nil)

	err := uc.Update(ctx, updatedTimezone)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Update_ValidationError(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone := createTestTimezone("UTC", "UTC")
	timezone.Name = "" // Invalid name

	timezoneRepo.On("GetByID", ctx, timezone.ID).Return(timezone, nil)

	err := uc.Update(ctx, timezone)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	timezoneRepo.AssertExpectations(t)
}

// Tests for Delete
func TestTimezoneUseCase_Delete_Success(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezone := createTestTimezone("UTC", "UTC")
	timezoneRepo.On("GetByID", ctx, timezone.ID).Return(timezone, nil)
	timezoneRepo.On("Delete", ctx, timezone.ID).Return(nil)

	err := uc.Delete(ctx, timezone.ID)

	assert.NoError(t, err)
	timezoneRepo.AssertExpectations(t)
}

func TestTimezoneUseCase_Delete_NotFound(t *testing.T) {
	uc, timezoneRepo := setupTimezoneUseCase(t)
	ctx := context.Background()

	timezoneID := uuidv7.New()
	timezoneRepo.On("GetByID", ctx, timezoneID).Return(nil, entity.ErrNotFound)

	err := uc.Delete(ctx, timezoneID)

	assert.Error(t, err)
	timezoneRepo.AssertExpectations(t)
}
