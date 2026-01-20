package banktransaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/banking/banktransaction/adapter/repository/postgres"
	txAggregate "github.com/basilex/promenade/internal/contexts/banking/banktransaction/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/integration"
)

// createTestAccount creates a test bank account and returns its ID
func createTestAccount(t *testing.T, ctx context.Context, tx *sqlx.Tx) uuidv7.UUID {
	accountID := uuidv7.New()
	organizationID := uuidv7.New()
	createdBy := uuidv7.New()
	_, err := tx.ExecContext(ctx, `
		INSERT INTO banking_bank_accounts (id, organization_id, name, bank_name, currency_code, balance_cents, status, last_updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, accountID, organizationID, "Test Account", "Test Bank", "UAH", 0, "active", createdBy)
	require.NoError(t, err)
	return accountID
}

func TestBankTransactionRepository_Create(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, err := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			100000,
			"UAH",
			time.Now(),
			"Test payment",
			uuidv7.New(),
		)
		require.NoError(t, err)

		err = repo.Create(ctx, transaction)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		assert.Equal(t, transaction.GetID(), found.GetID())
		assert.Equal(t, accountID, found.BankAccountID)
		assert.Equal(t, int64(100000), found.AmountCents)
	})
}

func TestBankTransactionRepository_GetByID_NotFound(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		nonExistent := uuidv7.New()
		_, err := repo.GetByID(ctx, nonExistent)
		assert.Error(t, err)
	})
}

func TestBankTransactionRepository_Update(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionDebit,
			50000,
			"UAH",
			time.Now(),
			"Initial",
			uuidv7.New(),
		)
		require.NoError(t, repo.Create(ctx, transaction))

		transaction.Description = "Updated description"
		err := repo.Update(ctx, transaction)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, transaction.GetID())
		assert.Equal(t, "Updated description", found.Description)
	})
}

func TestBankTransactionRepository_Delete(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			10000,
			"UAH",
			time.Now(),
			"To delete",
			uuidv7.New(),
		)
		require.NoError(t, repo.Create(ctx, transaction))

		err := repo.Delete(ctx, transaction.GetID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, transaction.GetID())
		assert.Error(t, err, "Soft deleted transaction should not be found")
	})
}

func TestBankTransactionRepository_ListByAccount(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)

		for i := 0; i < 5; i++ {
			transaction, _ := txAggregate.NewBankTransaction(
				accountID,
				txAggregate.DirectionCredit,
				int64(10000*(i+1)),
				"UAH",
				time.Now(),
				"Transaction",
				uuidv7.New(),
			)
			require.NoError(t, repo.Create(ctx, transaction))
		}

		transactions, err := repo.ListByAccount(ctx, accountID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, transactions, 2)

		transactions, err = repo.ListByAccount(ctx, accountID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, transactions, 5)
	})
}

func TestBankTransactionRepository_GetByExternalID(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			25000,
			"UAH",
			time.Now(),
			"External TX",
			uuidv7.New(),
		)
		transaction.ExternalID = "mono_tx_12345"
		require.NoError(t, repo.Create(ctx, transaction))

		found, err := repo.GetByExternalID(ctx, accountID, "mono_tx_12345")
		require.NoError(t, err)
		assert.Equal(t, transaction.GetID(), found.GetID())
		assert.Equal(t, "mono_tx_12345", found.ExternalID)
	})
}

func TestBankTransactionRepository_ListUnmatched(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)

		unmatchedTx, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			30000,
			"UAH",
			time.Now(),
			"Unmatched",
			uuidv7.New(),
		)
		_ = unmatchedTx.Book()
		require.NoError(t, repo.Create(ctx, unmatchedTx))

		matchedTx, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			40000,
			"UAH",
			time.Now(),
			"Matched",
			uuidv7.New(),
		)
		_ = matchedTx.Book()
		_ = matchedTx.Match(txAggregate.MatchedEntityInvoice, uuidv7.New())
		require.NoError(t, repo.Create(ctx, matchedTx))

		unmatched, err := repo.ListUnmatched(ctx, accountID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, unmatched, 1)
		assert.Equal(t, unmatchedTx.GetID(), unmatched[0].GetID())
	})
}

func TestBankTransactionRepository_CountByAccount(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)

		for i := 0; i < 3; i++ {
			transaction, _ := txAggregate.NewBankTransaction(
				accountID,
				txAggregate.DirectionCredit,
				10000,
				"UAH",
				time.Now(),
				"TX",
				uuidv7.New(),
			)
			require.NoError(t, repo.Create(ctx, transaction))
		}

		count, err := repo.CountByAccount(ctx, accountID)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

func TestBankTransactionRepository_AmountPrecision(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			123456,
			"UAH",
			time.Now(),
			"Precision test",
			uuidv7.New(),
		)
		require.NoError(t, repo.Create(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		assert.Equal(t, int64(123456), found.AmountCents,
			"Amount precision must be exact to the kopiyok! 1234.56 UAH = 123456 kopiyky")
	})
}

func TestBankTransactionRepository_MatchToInvoice(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		invoiceID := uuidv7.New()

		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			75000,
			"UAH",
			time.Now(),
			"Invoice payment",
			uuidv7.New(),
		)
		_ = transaction.Book()
		_ = transaction.Match(txAggregate.MatchedEntityInvoice, invoiceID)
		require.NoError(t, repo.Create(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		require.NotNil(t, found.MatchedEntityType)
		require.NotNil(t, found.MatchedEntityID)
		assert.Equal(t, txAggregate.MatchedEntityInvoice, *found.MatchedEntityType)
		assert.Equal(t, invoiceID, *found.MatchedEntityID)
		assert.Equal(t, txAggregate.TransactionStatusReconciled, found.Status)
	})
}

func TestBankTransactionRepository_MatchToOrder(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		orderID := uuidv7.New()

		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			85000,
			"UAH",
			time.Now(),
			"Order payment",
			uuidv7.New(),
		)
		_ = transaction.Book()
		_ = transaction.Match(txAggregate.MatchedEntityOrder, orderID)
		require.NoError(t, repo.Create(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		require.NotNil(t, found.MatchedEntityType)
		require.NotNil(t, found.MatchedEntityID)
		assert.Equal(t, txAggregate.MatchedEntityOrder, *found.MatchedEntityType)
		assert.Equal(t, orderID, *found.MatchedEntityID)
	})
}

func TestBankTransactionRepository_MatchToPayment(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		paymentID := uuidv7.New()

		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionDebit,
			95000,
			"UAH",
			time.Now(),
			"Payment to supplier",
			uuidv7.New(),
		)
		_ = transaction.Book()
		_ = transaction.Match(txAggregate.MatchedEntityPayment, paymentID)
		require.NoError(t, repo.Create(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		require.NotNil(t, found.MatchedEntityType)
		require.NotNil(t, found.MatchedEntityID)
		assert.Equal(t, txAggregate.MatchedEntityPayment, *found.MatchedEntityType)
		assert.Equal(t, paymentID, *found.MatchedEntityID)
	})
}

func TestBankTransactionRepository_Unmatch(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		invoiceID := uuidv7.New()

		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			55000,
			"UAH",
			time.Now(),
			"Matched then unmatched",
			uuidv7.New(),
		)
		_ = transaction.Book()
		_ = transaction.Match(txAggregate.MatchedEntityInvoice, invoiceID)
		require.NoError(t, repo.Create(ctx, transaction))

		_ = transaction.Unmatch()
		require.NoError(t, repo.Update(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		assert.Nil(t, found.MatchedEntityType)
		assert.Nil(t, found.MatchedEntityID)
	})
}

func TestBankTransactionRepository_StatusTransitions(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			65000,
			"UAH",
			time.Now(),
			"Status test",
			uuidv7.New(),
		)
		require.NoError(t, repo.Create(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		assert.Equal(t, txAggregate.TransactionStatusPending, found.Status)

		_ = transaction.Book()
		require.NoError(t, repo.Update(ctx, transaction))
		found, _ = repo.GetByID(ctx, transaction.GetID())
		assert.Equal(t, txAggregate.TransactionStatusBooked, found.Status)

		_ = transaction.Match(txAggregate.MatchedEntityInvoice, uuidv7.New())
		require.NoError(t, repo.Update(ctx, transaction))
		found, _ = repo.GetByID(ctx, transaction.GetID())
		assert.Equal(t, txAggregate.TransactionStatusReconciled, found.Status)
	})
}

func TestBankTransactionRepository_Cancel(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionDebit,
			35000,
			"UAH",
			time.Now(),
			"To cancel",
			uuidv7.New(),
		)
		require.NoError(t, repo.Create(ctx, transaction))

		_ = transaction.Cancel()
		require.NoError(t, repo.Update(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		assert.Equal(t, txAggregate.TransactionStatusCanceled, found.Status)
	})
}

func TestBankTransactionRepository_Counterparty(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)
		transaction, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			45000,
			"UAH",
			time.Now(),
			"From supplier",
			uuidv7.New(),
		)
		transaction.SetCounterparty("ТОВ Постачальник", "UA123456789012345678901234567")
		require.NoError(t, repo.Create(ctx, transaction))

		found, err := repo.GetByID(ctx, transaction.GetID())
		require.NoError(t, err)
		assert.Equal(t, "ТОВ Постачальник", found.CounterpartyName)
		assert.Equal(t, "UA123456789012345678901234567", found.CounterpartyIBAN)
	})
}

func TestBankTransactionRepository_ExternalIDUnique(t *testing.T) {
	db := integration.SetupTestDB(t)
	db.WithTransaction(t, func(ctx context.Context, tx *sqlx.Tx) {
		repo := postgres.NewBankTransactionRepository(db.DB)

		accountID := createTestAccount(t, ctx, tx)

		tx1, _ := txAggregate.NewBankTransaction(
			accountID,
			txAggregate.DirectionCredit,
			15000,
			"UAH",
			time.Now(),
			"First",
			uuidv7.New(),
		)
		tx1.ExternalID = "ext_unique_123"
		require.NoError(t, repo.Create(ctx, tx1))

		found, err := repo.GetByExternalID(ctx, accountID, "ext_unique_123")
		require.NoError(t, err)
		assert.Equal(t, tx1.GetID(), found.GetID())
	})
}
