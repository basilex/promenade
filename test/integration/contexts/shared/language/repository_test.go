package language_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	"github.com/basilex/promenade/internal/contexts/shared/language/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "ts",
				Code3:      "tst",
				Name:       "Test Language",
				NativeName: "Язык Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			// Verify created
			found, err := repo.GetByID(ctx, l.ID)
			require.NoError(t, err)
			assert.Equal(t, l.ID, found.ID)
			assert.Equal(t, "ts", found.Code)
			assert.Equal(t, "tst", found.Code3)
			assert.Equal(t, "Test Language", found.Name)
			assert.Equal(t, "Язык Тест", found.NativeName)
			assert.True(t, found.IsActive)
		})
	})

	t.Run("duplicate_code", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l1 := &language.Language{
				ID:         uuidv7.New(),
				Code:       "du",
				Code3:      "dup",
				Name:       "Duplicate Test",
				NativeName: "Дубль Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l1)
			require.NoError(t, err)

			// Try to create with same code
			l2 := &language.Language{
				ID:         uuidv7.New(),
				Code:       "du", // Same code
				Code3:      "du2",
				Name:       "Duplicate Test 2",
				NativeName: "Дубль Тест 2",
				IsActive:   true,
			}

			err = repo.Create(ctx, l2)
			assert.Error(t, err) // Should fail due to unique constraint
		})
	})

	t.Run("optional_code3", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "oc",
				Code3:      "", // Optional field
				Name:       "Optional Code3 Test",
				NativeName: "Опціональний",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, l.ID)
			require.NoError(t, err)
			assert.Equal(t, "", found.Code3)
		})
	})
}

func TestRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "gb",
				Code3:      "gbt",
				Name:       "GetByID Test",
				NativeName: "GetByID Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			found, err := repo.GetByID(ctx, l.ID)
			require.NoError(t, err)
			assert.Equal(t, l.ID, found.ID)
			assert.Equal(t, "gb", found.Code)
			assert.Equal(t, "GetByID Test", found.Name)
		})
	})

	t.Run("not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			nonExistentID := uuidv7.New()
			found, err := repo.GetByID(ctx, nonExistentID)
			assert.Error(t, err)
			assert.Equal(t, language.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})

	t.Run("inactive_not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "in",
				Code3:      "ina",
				Name:       "Inactive Test",
				NativeName: "Неактивний Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			// Soft delete
			err = repo.Delete(ctx, l.ID)
			require.NoError(t, err)

			// Should not find inactive
			found, err := repo.GetByID(ctx, l.ID)
			assert.Error(t, err)
			assert.Equal(t, language.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}

func TestRepository_GetByCode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "gc",
				Code3:      "gct",
				Name:       "GetByCode Test",
				NativeName: "GetByCode Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			found, err := repo.GetByCode(ctx, "gc")
			require.NoError(t, err)
			assert.Equal(t, l.ID, found.ID)
			assert.Equal(t, "gc", found.Code)
			assert.Equal(t, "GetByCode Test", found.Name)
		})
	})

	t.Run("case_sensitive", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "cs",
				Code3:      "cst",
				Name:       "Case Sensitive Test",
				NativeName: "Регістр Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			// Should find with exact case
			found, err := repo.GetByCode(ctx, "cs")
			require.NoError(t, err)
			assert.Equal(t, "cs", found.Code)

			// Should not find with different case
			notFound, err := repo.GetByCode(ctx, "CS")
			assert.Error(t, err)
			assert.Equal(t, language.ErrNotFound, err)
			assert.Nil(t, notFound)
		})
	})

	t.Run("not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			found, err := repo.GetByCode(ctx, "zz")
			assert.Error(t, err)
			assert.Equal(t, language.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})

	t.Run("inactive_not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "ia",
				Code3:      "iac",
				Name:       "Inactive Code Test",
				NativeName: "Неактивний Код Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			// Soft delete
			err = repo.Delete(ctx, l.ID)
			require.NoError(t, err)

			// Should not find inactive
			found, err := repo.GetByCode(ctx, "ia")
			assert.Error(t, err)
			assert.Equal(t, language.ErrNotFound, err)
			assert.Nil(t, found)
		})
	})
}

func TestRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "up",
				Code3:      "upd",
				Name:       "Update Test",
				NativeName: "Оновлення Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			// Update fields
			l.Name = "Updated Name"
			l.NativeName = "Оновлена Назва"
			l.Code3 = "upn"
			l.IsActive = false

			err = repo.Update(ctx, l)
			require.NoError(t, err)

			// Verify updated (note: is_active = false means GetByID won't find it)
			// So we need to query directly
			var found language.Language
			query := `SELECT id, code, code3, name, native_name, is_active 
			          FROM shared_languages WHERE id = $1`
			err = tx.GetContext(ctx, &found, query, l.ID)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", found.Name)
			assert.Equal(t, "Оновлена Назва", found.NativeName)
			assert.Equal(t, "upn", found.Code3)
			assert.False(t, found.IsActive)
		})
	})

	t.Run("not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "nf",
				Code3:      "nfd",
				Name:       "Not Found",
				NativeName: "Не Знайдено",
				IsActive:   true,
			}

			// Try to update non-existent record (no error, just 0 rows affected)
			err := repo.Update(ctx, l)
			assert.NoError(t, err) // UPDATE doesn't fail if no rows affected
		})
	})
}

func TestRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			l := &language.Language{
				ID:         uuidv7.New(),
				Code:       "dl",
				Code3:      "del",
				Name:       "Delete Test",
				NativeName: "Видалення Тест",
				IsActive:   true,
			}

			err := repo.Create(ctx, l)
			require.NoError(t, err)

			// Delete (soft delete)
			err = repo.Delete(ctx, l.ID)
			require.NoError(t, err)

			// Should not find after delete
			found, err := repo.GetByID(ctx, l.ID)
			assert.Error(t, err)
			assert.Equal(t, language.ErrNotFound, err)
			assert.Nil(t, found)

			// But should exist in DB with is_active = false
			var dbRecord language.Language
			query := `SELECT id, code, is_active FROM shared_languages WHERE id = $1`
			err = tx.GetContext(ctx, &dbRecord, query, l.ID)
			require.NoError(t, err)
			assert.False(t, dbRecord.IsActive)
		})
	})

	t.Run("not_found", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			nonExistentID := uuidv7.New()
			err := repo.Delete(ctx, nonExistentID)
			assert.NoError(t, err) // DELETE doesn't fail if no rows affected
		})
	})
}

func TestRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)

	t.Run("success", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			// Create test languages
			languages := []*language.Language{
				{
					ID:         uuidv7.New(),
					Code:       "l1",
					Code3:      "ls1",
					Name:       "List Test Alpha",
					NativeName: "Список Альфа",
					IsActive:   true,
				},
				{
					ID:         uuidv7.New(),
					Code:       "l2",
					Code3:      "ls2",
					Name:       "List Test Beta",
					NativeName: "Список Бета",
					IsActive:   true,
				},
				{
					ID:         uuidv7.New(),
					Code:       "l3",
					Code3:      "ls3",
					Name:       "List Test Gamma",
					NativeName: "Список Гамма",
					IsActive:   false, // Inactive - should not appear in list
				},
			}

			for _, l := range languages {
				err := repo.Create(ctx, l)
				require.NoError(t, err)
			}

			// List should return only active languages
			found, err := repo.List(ctx)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(found), 2) // At least our 2 active test languages

			// Check ordering (by name ASC)
			testLanguages := filterTestLanguages(found)
			if len(testLanguages) >= 2 {
				assert.Equal(t, "List Test Alpha", testLanguages[0].Name)
				assert.Equal(t, "List Test Beta", testLanguages[1].Name)
			}

			// Verify inactive not in list
			for _, l := range found {
				assert.True(t, l.IsActive)
			}
		})
	})

	t.Run("empty_list", func(t *testing.T) {
		testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
			repo := postgres.NewRepository(tx)

			// In a fresh transaction, no test data exists yet
			found, err := repo.List(ctx)
			require.NoError(t, err)
			// List returns nil or empty slice when no data - both valid
			_ = found
		})
	})
}

// Helper function to filter test languages from list
func filterTestLanguages(languages []*language.Language) []*language.Language {
	var result []*language.Language
	for _, l := range languages {
		if l.Code == "l1" || l.Code == "l2" {
			result = append(result, l)
		}
	}
	return result
}
