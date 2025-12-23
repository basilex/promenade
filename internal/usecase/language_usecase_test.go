package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/basilex/promenade/internal/domain/entity"
	"github.com/basilex/promenade/pkg/pagination"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// Mock LanguageRepository
type mockLanguageRepository struct {
	mock.Mock
}

func (m *mockLanguageRepository) Create(ctx context.Context, language *entity.Language) error {
	args := m.Called(ctx, language)
	return args.Error(0)
}

func (m *mockLanguageRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Language, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Language), args.Error(1)
}

func (m *mockLanguageRepository) GetByCode(ctx context.Context, code string) (*entity.Language, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Language), args.Error(1)
}

func (m *mockLanguageRepository) List(ctx context.Context, params pagination.Params) ([]*entity.Language, *pagination.Metadata, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*entity.Language), args.Get(1).(*pagination.Metadata), args.Error(2)
}

func (m *mockLanguageRepository) ListActive(ctx context.Context) ([]*entity.Language, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Language), args.Error(1)
}

func (m *mockLanguageRepository) Update(ctx context.Context, language *entity.Language) error {
	args := m.Called(ctx, language)
	return args.Error(0)
}

func (m *mockLanguageRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockLanguageRepository) Exists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

// Test helpers
func setupLanguageUseCase(t *testing.T) (*languageUseCase, *mockLanguageRepository) {
	languageRepo := new(mockLanguageRepository)
	uc := NewLanguageUseCase(languageRepo).(*languageUseCase)
	return uc, languageRepo
}

func createTestLanguage(name, code string) *entity.Language {
	return &entity.Language{
		ID:         uuidv7.New(),
		Name:       name,
		NativeName: name,
		Code:       code,
		ISO639_2:   code + "x",
		IsActive:   true,
	}
}

// Tests for Create
func TestLanguageUseCase_Create_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("Exists", ctx, "en").Return(false, nil)
	languageRepo.On("Create", ctx, language).Return(nil)

	err := uc.Create(ctx, language)

	assert.NoError(t, err)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Create_GeneratesUUID(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	language.ID = uuidv7.UUID{} // Empty UUID

	languageRepo.On("Exists", ctx, "en").Return(false, nil)
	languageRepo.On("Create", ctx, mock.AnythingOfType("*entity.Language")).Return(nil)

	err := uc.Create(ctx, language)

	assert.NoError(t, err)
	assert.NotEqual(t, uuidv7.UUID{}, language.ID) // UUID should be generated
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Create_AlreadyExists(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("Exists", ctx, "en").Return(true, nil)

	err := uc.Create(ctx, language)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Create_ValidationError(t *testing.T) {
	uc, _ := setupLanguageUseCase(t)
	ctx := context.Background()

	// Language with invalid data
	language := &entity.Language{
		ID:         uuidv7.New(),
		Name:       "", // Empty name should fail validation
		NativeName: "English",
		Code:       "en",
		ISO639_2:   "eng",
	}

	err := uc.Create(ctx, language)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestLanguageUseCase_Create_ExistsCheckError(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("Exists", ctx, "en").Return(false, errors.New("database error"))

	err := uc.Create(ctx, language)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check language existence")
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Create_RepositoryError(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("Exists", ctx, "en").Return(false, nil)
	languageRepo.On("Create", ctx, language).Return(errors.New("database error"))

	err := uc.Create(ctx, language)

	assert.Error(t, err)
	languageRepo.AssertExpectations(t)
}

// Tests for GetByID
func TestLanguageUseCase_GetByID_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	expectedLanguage := createTestLanguage("English", "en")
	languageRepo.On("GetByID", ctx, expectedLanguage.ID).Return(expectedLanguage, nil)

	language, err := uc.GetByID(ctx, expectedLanguage.ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedLanguage.ID, language.ID)
	assert.Equal(t, "English", language.Name)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_GetByID_NotFound(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	languageID := uuidv7.New()
	languageRepo.On("GetByID", ctx, languageID).Return(nil, entity.ErrNotFound)

	language, err := uc.GetByID(ctx, languageID)

	assert.Error(t, err)
	assert.Nil(t, language)
	languageRepo.AssertExpectations(t)
}

// Tests for GetByCode
func TestLanguageUseCase_GetByCode_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	expectedLanguage := createTestLanguage("English", "en")
	languageRepo.On("GetByCode", ctx, "en").Return(expectedLanguage, nil)

	language, err := uc.GetByCode(ctx, "en")

	assert.NoError(t, err)
	assert.Equal(t, "en", language.Code)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_GetByCode_NotFound(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	languageRepo.On("GetByCode", ctx, "xx").Return(nil, entity.ErrNotFound)

	language, err := uc.GetByCode(ctx, "xx")

	assert.Error(t, err)
	assert.Nil(t, language)
	languageRepo.AssertExpectations(t)
}

// Tests for List
func TestLanguageUseCase_List_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language1 := createTestLanguage("English", "en")
	language2 := createTestLanguage("Russian", "ru")
	expectedLanguages := []*entity.Language{language1, language2}
	metadata := &pagination.Metadata{
		Total:       2,
		Limit:       20,
		Offset:      0,
		CurrentPage: 1,
		TotalPages:  1,
	}

	params := pagination.Params{Page: 1, PageSize: 20}
	languageRepo.On("List", ctx, params).Return(expectedLanguages, metadata, nil)

	languages, meta, err := uc.List(ctx, params)

	assert.NoError(t, err)
	assert.Len(t, languages, 2)
	assert.Equal(t, 2, meta.Total)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_List_DefaultPagination(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	expectedLanguages := []*entity.Language{}
	metadata := &pagination.Metadata{
		Total:       0,
		Limit:       20,
		Offset:      0,
		CurrentPage: 1,
		TotalPages:  0,
	}

	// Invalid params should be normalized
	params := pagination.Params{Page: 0, PageSize: 0}
	normalizedParams := pagination.Params{Page: 1, PageSize: 20}
	languageRepo.On("List", ctx, normalizedParams).Return(expectedLanguages, metadata, nil)

	languages, meta, err := uc.List(ctx, params)

	assert.NoError(t, err)
	assert.Empty(t, languages)
	assert.Equal(t, 0, meta.Total)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_List_MaxPageSize(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	expectedLanguages := []*entity.Language{}
	metadata := &pagination.Metadata{
		Total:       0,
		Limit:       20,
		Offset:      0,
		CurrentPage: 1,
		TotalPages:  0,
	}

	// PageSize > 100 should be normalized to 20
	params := pagination.Params{Page: 1, PageSize: 200}
	normalizedParams := pagination.Params{Page: 1, PageSize: 20}
	languageRepo.On("List", ctx, normalizedParams).Return(expectedLanguages, metadata, nil)

	languages, meta, err := uc.List(ctx, params)

	assert.NoError(t, err)
	assert.Empty(t, languages)
	assert.Equal(t, 0, meta.Total)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_List_RepositoryError(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	params := pagination.Params{Page: 1, PageSize: 20}
	languageRepo.On("List", ctx, params).Return(nil, nil, errors.New("database error"))

	languages, meta, err := uc.List(ctx, params)

	assert.Error(t, err)
	assert.Nil(t, languages)
	assert.Nil(t, meta)
	languageRepo.AssertExpectations(t)
}

// Tests for ListActive
func TestLanguageUseCase_ListActive_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language1 := createTestLanguage("English", "en")
	language2 := createTestLanguage("Russian", "ru")
	expectedLanguages := []*entity.Language{language1, language2}

	languageRepo.On("ListActive", ctx).Return(expectedLanguages, nil)

	languages, err := uc.ListActive(ctx)

	assert.NoError(t, err)
	assert.Len(t, languages, 2)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_ListActive_Empty(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	languageRepo.On("ListActive", ctx).Return([]*entity.Language{}, nil)

	languages, err := uc.ListActive(ctx)

	assert.NoError(t, err)
	assert.Empty(t, languages)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_ListActive_RepositoryError(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	languageRepo.On("ListActive", ctx).Return(nil, errors.New("database error"))

	languages, err := uc.ListActive(ctx)

	assert.Error(t, err)
	assert.Nil(t, languages)
	languageRepo.AssertExpectations(t)
}

// Tests for Update
func TestLanguageUseCase_Update_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	language.Name = "English (US)"

	languageRepo.On("GetByID", ctx, language.ID).Return(language, nil)
	languageRepo.On("Update", ctx, language).Return(nil)

	err := uc.Update(ctx, language)

	assert.NoError(t, err)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Update_NotFound(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("GetByID", ctx, language.ID).Return(nil, entity.ErrNotFound)

	err := uc.Update(ctx, language)

	assert.Error(t, err)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Update_CodeChanged_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	existingLanguage := createTestLanguage("English", "en")
	updatedLanguage := &entity.Language{
		ID:         existingLanguage.ID,
		Name:       "English",
		NativeName: "English",
		Code:       "fr",  // 2-char ISO 639-1 code
		ISO639_2:   "fra", // 3-char ISO 639-2 code
		IsActive:   true,
	}

	languageRepo.On("GetByID", ctx, updatedLanguage.ID).Return(existingLanguage, nil)
	languageRepo.On("Exists", ctx, "fr").Return(false, nil)
	languageRepo.On("Update", ctx, updatedLanguage).Return(nil)

	err := uc.Update(ctx, updatedLanguage)

	assert.NoError(t, err)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Update_CodeChanged_AlreadyExists(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	existingLanguage := createTestLanguage("English", "en")
	updatedLanguage := createTestLanguage("English", "ru")
	updatedLanguage.ID = existingLanguage.ID

	languageRepo.On("GetByID", ctx, updatedLanguage.ID).Return(existingLanguage, nil)
	languageRepo.On("Exists", ctx, "ru").Return(true, nil)

	err := uc.Update(ctx, updatedLanguage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Update_ValidationError(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	language.Name = "" // Invalid name

	languageRepo.On("GetByID", ctx, language.ID).Return(language, nil)

	err := uc.Update(ctx, language)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Update_RepositoryError(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("GetByID", ctx, language.ID).Return(language, nil)
	languageRepo.On("Update", ctx, language).Return(errors.New("database error"))

	err := uc.Update(ctx, language)

	assert.Error(t, err)
	languageRepo.AssertExpectations(t)
}

// Tests for Delete
func TestLanguageUseCase_Delete_Success(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("GetByID", ctx, language.ID).Return(language, nil)
	languageRepo.On("Delete", ctx, language.ID).Return(nil)

	err := uc.Delete(ctx, language.ID)

	assert.NoError(t, err)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Delete_NotFound(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	languageID := uuidv7.New()
	languageRepo.On("GetByID", ctx, languageID).Return(nil, entity.ErrNotFound)

	err := uc.Delete(ctx, languageID)

	assert.Error(t, err)
	languageRepo.AssertExpectations(t)
}

func TestLanguageUseCase_Delete_RepositoryError(t *testing.T) {
	uc, languageRepo := setupLanguageUseCase(t)
	ctx := context.Background()

	language := createTestLanguage("English", "en")
	languageRepo.On("GetByID", ctx, language.ID).Return(language, nil)
	languageRepo.On("Delete", ctx, language.ID).Return(errors.New("database error"))

	err := uc.Delete(ctx, language.ID)

	assert.Error(t, err)
	languageRepo.AssertExpectations(t)
}
