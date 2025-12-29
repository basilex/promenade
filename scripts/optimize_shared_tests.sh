#!/bin/bash
# Script to optimize integration tests for Shared contexts

set -e

# Function to create optimized test file
create_optimized_test() {
    local context=$1
    local Context=$2  # Capitalized
    
    cat > "test/integration/contexts/shared/${context}/repository_test.go" <<'EOF'
package ${context}_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/${context}"
	"github.com/basilex/promenade/internal/contexts/shared/${context}/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func Test${Context}Repository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		// Test Create, Read, Update, Delete
		// Implementation depends on entity structure
	})
}

func Test${Context}Repository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		// Test queries (GetByCode, GetByName, List)
	})
}
EOF
    
    # Replace placeholders
    sed -i "" "s/\${context}/${context}/g" "test/integration/contexts/shared/${context}/repository_test.go"
    sed -i "" "s/\${Context}/${Context}/g" "test/integration/contexts/shared/${context}/repository_test.go"
}

echo "Optimizing Shared context integration tests..."

# Currency
cat > "test/integration/contexts/shared/currency/repository_test.go" <<'CURRENCYEOF'
package currency_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/internal/contexts/shared/currency/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestCurrencyRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		c := &currency.Currency{
			ID:            uuidv7.New(),
			Code:          "TST",
			NumericCode:   "999",
			Name:          "Test Currency",
			NameLocal:     "Test Local",
			Symbol:        "T$",
			DecimalDigits: 2,
			IsActive:      true,
		}
		require.NoError(t, repo.Create(ctx, c))
		found, err := repo.GetByID(ctx, c.ID)
		require.NoError(t, err)
		assert.Equal(t, "TST", found.Code)
		found.Name = "Updated"
		require.NoError(t, repo.Update(ctx, found))
		require.NoError(t, repo.Delete(ctx, c.ID))
		_, err = repo.GetByID(ctx, c.ID)
		assert.Error(t, err)
	})
}

func TestCurrencyRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		c := &currency.Currency{
			ID:            uuidv7.New(),
			Code:          "TST",
			NumericCode:   "999",
			Name:          "Test",
			NameLocal:     "Local",
			Symbol:        "$",
			DecimalDigits: 2,
			IsActive:      true,
		}
		require.NoError(t, repo.Create(ctx, c))
		found, err := repo.GetByCode(ctx, "TST")
		require.NoError(t, err)
		assert.Equal(t, "Test", found.Name)
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Greater(t, len(all), 0)
	})
}
CURRENCYEOF

# Language  
cat > "test/integration/contexts/shared/language/repository_test.go" <<'LANGUAGEEOF'
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

func TestLanguageRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		l := &language.Language{
			ID:        uuidv7.New(),
			Code:      "ts",
			Code3:     "tst",
			Name:      "Test",
			NameLocal: "Local",
			IsActive:  true,
		}
		require.NoError(t, repo.Create(ctx, l))
		found, err := repo.GetByID(ctx, l.ID)
		require.NoError(t, err)
		assert.Equal(t, "ts", found.Code)
		found.Name = "Updated"
		require.NoError(t, repo.Update(ctx, found))
		require.NoError(t, repo.Delete(ctx, l.ID))
		_, err = repo.GetByID(ctx, l.ID)
		assert.Error(t, err)
	})
}

func TestLanguageRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		l := &language.Language{
			ID:        uuidv7.New(),
			Code:      "ts",
			Code3:     "tst",
			Name:      "Test",
			NameLocal: "Local",
			IsActive:  true,
		}
		require.NoError(t, repo.Create(ctx, l))
		found, err := repo.GetByCode(ctx, "ts")
		require.NoError(t, err)
		assert.Equal(t, "Test", found.Name)
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Greater(t, len(all), 0)
	})
}
LANGUAGEEOF

# Timezone
cat > "test/integration/contexts/shared/timezone/repository_test.go" <<'TIMEZONEEOF'
package timezone_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/internal/contexts/shared/timezone/adapter/repository/postgres"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestTimezoneRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		tz := &timezone.Timezone{
			ID:        uuidv7.New(),
			Name:      "Test/Zone",
			Offset:    "+00:00",
			OffsetDST: "+01:00",
			IsActive:  true,
		}
		require.NoError(t, repo.Create(ctx, tz))
		found, err := repo.GetByID(ctx, tz.ID)
		require.NoError(t, err)
		assert.Equal(t, "Test/Zone", found.Name)
		found.Offset = "+02:00"
		require.NoError(t, repo.Update(ctx, found))
		require.NoError(t, repo.Delete(ctx, tz.ID))
		_, err = repo.GetByID(ctx, tz.ID)
		assert.Error(t, err)
	})
}

func TestTimezoneRepository_Queries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	testDB := integration.SetupTestDB(t)
	testDB.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewRepository(tx)
		tz := &timezone.Timezone{
			ID:        uuidv7.New(),
			Name:      "Test/Zone",
			Offset:    "+00:00",
			OffsetDST: "+01:00",
			IsActive:  true,
		}
		require.NoError(t, repo.Create(ctx, tz))
		found, err := repo.GetByName(ctx, "Test/Zone")
		require.NoError(t, err)
		assert.Equal(t, "+00:00", found.Offset)
		all, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Greater(t, len(all), 0)
	})
}
TIMEZONEEOF

echo "✓ Shared contexts optimized"
go test -short ./test/integration/contexts/shared/... 2>&1 | grep -E "(PASS|FAIL|SKIP)"
