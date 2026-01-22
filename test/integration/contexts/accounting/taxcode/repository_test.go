package taxcode_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/accounting/taxcode"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/adapter/repository/postgres"
	"github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

func TestTaxCodeRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create VAT tax code with 20% rate (2000 basis points)
		taxCode, err := aggregate.NewTaxCode(
			orgID,
			"VAT-20",
			"VAT 20%",
			aggregate.TaxTypeVAT,
			2000,
			userID,
		)
		require.NoError(t, err)

		err = repo.Create(ctx, taxCode)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, taxCode.GetID())
		require.NoError(t, err)
		assert.Equal(t, taxCode.GetID(), found.GetID())
		assert.Equal(t, orgID, found.OrganizationID)
		assert.Equal(t, "VAT-20", found.Code)
		assert.Equal(t, "VAT 20%", found.Name)
		assert.Equal(t, aggregate.TaxTypeVAT, found.TaxType)
		assert.Equal(t, 2000, found.Rate)
		assert.True(t, found.IsActive)
	})
}

func TestTaxCodeRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.ErrorIs(t, err, taxcode.ErrTaxCodeNotFound)
	})
}

func TestTaxCodeRepository_GetByCode(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		taxCode, _ := aggregate.NewTaxCode(orgID, "VAT-7", "VAT 7%", aggregate.TaxTypeVAT, 700, userID)
		require.NoError(t, repo.Create(ctx, taxCode))

		found, err := repo.GetByCode(ctx, orgID, "VAT-7")
		require.NoError(t, err)
		assert.Equal(t, taxCode.GetID(), found.GetID())
		assert.Equal(t, "VAT-7", found.Code)
	})
}

func TestTaxCodeRepository_Update_ChangeRate(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		taxCode, _ := aggregate.NewTaxCode(orgID, "VAT-20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
		require.NoError(t, repo.Create(ctx, taxCode))

		// Change rate to 18%
		err := taxCode.UpdateRate(1800)
		require.NoError(t, err)

		err = repo.Update(ctx, taxCode)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, taxCode.GetID())
		require.NoError(t, err)
		assert.Equal(t, 1800, found.Rate)
	})
}

func TestTaxCodeRepository_Update_SetGLAccounts(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create GL accounts first (referential integrity!)
		payableAccountID, err := createTestGLAccount(ctx, tx, orgID, userID, "631", "Tax Payable", "liability")
		require.NoError(t, err)
		receivableAccountID, err := createTestGLAccount(ctx, tx, orgID, userID, "641", "Tax Receivable", "asset")
		require.NoError(t, err)

		taxCode, _ := aggregate.NewTaxCode(orgID, "VAT-20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
		require.NoError(t, repo.Create(ctx, taxCode))

		// Set GL accounts
		err = taxCode.SetTaxPayableAccount(payableAccountID)
		require.NoError(t, err)
		err = taxCode.SetTaxReceivableAccount(receivableAccountID)
		require.NoError(t, err)

		err = repo.Update(ctx, taxCode)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, taxCode.GetID())
		require.NoError(t, err)
		assert.NotNil(t, found.TaxPayableAccountID)
		assert.Equal(t, payableAccountID, *found.TaxPayableAccountID)
		assert.NotNil(t, found.TaxReceivableAccountID)
		assert.Equal(t, receivableAccountID, *found.TaxReceivableAccountID)
	})
}

func TestTaxCodeRepository_Update_ActivateDeactivate(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		taxCode, _ := aggregate.NewTaxCode(orgID, "VAT-OLD", "VAT Old Rate", aggregate.TaxTypeVAT, 1500, userID)
		require.NoError(t, repo.Create(ctx, taxCode))

		// Deactivate
		taxCode.Deactivate()
		err := repo.Update(ctx, taxCode)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, taxCode.GetID())
		require.NoError(t, err)
		assert.False(t, found.IsActive)

		// Reactivate
		taxCode.Activate()
		err = repo.Update(ctx, taxCode)
		require.NoError(t, err)

		found, err = repo.GetByID(ctx, taxCode.GetID())
		require.NoError(t, err)
		assert.True(t, found.IsActive)
	})
}

func TestTaxCodeRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		taxCode, _ := aggregate.NewTaxCode(orgID, "TEMP-TAX", "Temporary Tax", aggregate.TaxTypeOther, 500, userID)
		require.NoError(t, repo.Create(ctx, taxCode))

		err := repo.Delete(ctx, taxCode.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, taxCode.GetID())
		assert.ErrorIs(t, err, taxcode.ErrTaxCodeNotFound)
	})
}

func TestTaxCodeRepository_ListByOrganization(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create multiple tax codes
		codes := []struct {
			code    string
			name    string
			taxType aggregate.TaxType
			rate    int
		}{
			{"VAT-20", "VAT 20%", aggregate.TaxTypeVAT, 2000},
			{"VAT-7", "VAT 7%", aggregate.TaxTypeVAT, 700},
			{"INC-18", "Income Tax 18%", aggregate.TaxTypeIncomeTax, 1800},
		}

		for _, c := range codes {
			tc, _ := aggregate.NewTaxCode(orgID, c.code, c.name, c.taxType, c.rate, userID)
			require.NoError(t, repo.Create(ctx, tc))
		}

		found, err := repo.ListByOrganization(ctx, orgID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, found, 3)
	})
}

func TestTaxCodeRepository_ListByType(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create VAT and Income Tax codes
		vat1, _ := aggregate.NewTaxCode(orgID, "VAT-20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
		require.NoError(t, repo.Create(ctx, vat1))

		vat2, _ := aggregate.NewTaxCode(orgID, "VAT-7", "VAT 7%", aggregate.TaxTypeVAT, 700, userID)
		require.NoError(t, repo.Create(ctx, vat2))

		income, _ := aggregate.NewTaxCode(orgID, "INC-18", "Income Tax 18%", aggregate.TaxTypeIncomeTax, 1800, userID)
		require.NoError(t, repo.Create(ctx, income))

		found, err := repo.ListByType(ctx, orgID, aggregate.TaxTypeVAT)
		require.NoError(t, err)
		assert.Len(t, found, 2)
		for _, tc := range found {
			assert.Equal(t, aggregate.TaxTypeVAT, tc.TaxType)
		}
	})
}

func TestTaxCodeRepository_ListActive(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create active and inactive tax codes
		active1, _ := aggregate.NewTaxCode(orgID, "VAT-20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
		require.NoError(t, repo.Create(ctx, active1))

		inactive, _ := aggregate.NewTaxCode(orgID, "VAT-OLD", "VAT Old", aggregate.TaxTypeVAT, 1500, userID)
		require.NoError(t, repo.Create(ctx, inactive))
		inactive.Deactivate()
		require.NoError(t, repo.Update(ctx, inactive))

		active2, _ := aggregate.NewTaxCode(orgID, "VAT-7", "VAT 7%", aggregate.TaxTypeVAT, 700, userID)
		require.NoError(t, repo.Create(ctx, active2))

		found, err := repo.ListActive(ctx, orgID)
		require.NoError(t, err)
		assert.Len(t, found, 2)
		for _, tc := range found {
			assert.True(t, tc.IsActive)
		}
	})
}

func TestTaxCode_BusinessRules_CalculateTax(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		// 20% VAT (2000 basis points)
		taxCode, _ := aggregate.NewTaxCode(orgID, "VAT-20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)

		// 100.00 UAH taxable amount
		taxableAmount := int64(10000) // cents
		expectedTax := int64(2000)    // 20% of 100.00 = 20.00

		calculatedTax := taxCode.CalculateTax(taxableAmount)
		assert.Equal(t, expectedTax, calculatedTax)
	})
}

func TestTaxCode_BusinessRules_CalculateTaxFromGross(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		// 20% VAT (2000 basis points)
		taxCode, _ := aggregate.NewTaxCode(orgID, "VAT-20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)

		// 120.00 UAH gross amount (includes 20% VAT)
		grossAmount := int64(12000) // cents

		// Expected: base = 100.00, tax = 20.00
		expectedBase := int64(10000)
		expectedTax := int64(2000)

		calculatedBase := taxCode.CalculateTaxableBase(grossAmount)
		calculatedTax := taxCode.CalculateTaxFromGross(grossAmount)

		assert.Equal(t, expectedBase, calculatedBase)
		assert.Equal(t, expectedTax, calculatedTax)
	})
}

func TestTaxCode_BusinessRules_InvalidRate(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Negative rate
		_, err := aggregate.NewTaxCode(orgID, "INVALID", "Invalid Tax", aggregate.TaxTypeVAT, -100, userID)
		assert.ErrorIs(t, err, taxcode.ErrInvalidTaxRate)

		// Rate over 100% (10000 basis points)
		_, err = aggregate.NewTaxCode(orgID, "INVALID", "Invalid Tax", aggregate.TaxTypeVAT, 10001, userID)
		assert.ErrorIs(t, err, taxcode.ErrInvalidTaxRate)

		// Valid rate at boundary
		_, err = aggregate.NewTaxCode(orgID, "VALID", "Valid Tax", aggregate.TaxTypeVAT, 10000, userID)
		assert.NoError(t, err)
	})
}

func TestTaxCode_BusinessRules_EmptyCodeOrName(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Empty code
		_, err := aggregate.NewTaxCode(orgID, "", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
		assert.ErrorIs(t, err, taxcode.ErrTaxCodeEmpty)

		// Empty name
		_, err = aggregate.NewTaxCode(orgID, "VAT-20", "", aggregate.TaxTypeVAT, 2000, userID)
		assert.ErrorIs(t, err, taxcode.ErrTaxNameEmpty)
	})
}

func TestTaxCode_BusinessRules_DifferentTaxTypes(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewTaxCodeRepository(db.DB)

		orgID := uuidv7.New()
		userID := uuidv7.New()

		// Create tax codes of different types
		types := []aggregate.TaxType{
			aggregate.TaxTypeVAT,
			aggregate.TaxTypeIncomeTax,
			aggregate.TaxTypePayrollTax,
			aggregate.TaxTypeWithholding,
			aggregate.TaxTypeExcise,
			aggregate.TaxTypeCustoms,
			aggregate.TaxTypeProperty,
			aggregate.TaxTypeOther,
		}

		for i, taxType := range types {
			tc, err := aggregate.NewTaxCode(
				orgID,
				string(taxType),
				string(taxType)+" Tax",
				taxType,
				1000+i*100,
				userID,
			)
			require.NoError(t, err)
			require.NoError(t, repo.Create(ctx, tc))

			found, err := repo.GetByCode(ctx, orgID, string(taxType))
			require.NoError(t, err)
			assert.Equal(t, taxType, found.TaxType)
		}
	})
}
