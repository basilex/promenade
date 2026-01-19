package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	receipterrors "github.com/basilex/promenade/internal/contexts/fiscal/receipt"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func createReceiptForMock(t *testing.T) *aggregate.Receipt {
	rec, err := aggregate.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)
	return rec
}

func buildReceiptRow(t *testing.T) (*aggregate.Receipt, []string, []driver.Value) {
	rec := createReceiptForMock(t)
	now := time.Now().UTC()

	columns := []string{
		"id",
		"version",
		"cash_register_id",
		"order_id",
		"payment_type",
		"receipt_type",
		"currency",
		"total_amount",
		"tax_amount",
		"fiscal_number",
		"fiscal_url",
		"qr_code",
		"provider_receipt_id",
		"printed_at",
		"cancelled_at",
		"cancellation_reason",
		"lines",
		"created_by",
		"last_updated_by",
		"status",
		"created_at",
		"updated_at",
		"deleted_at",
	}

	values := []driver.Value{
		rec.GetID().String(),
		rec.GetVersion(),
		rec.CashRegisterID.String(),
		rec.OrderID.String(),
		string(rec.PaymentType),
		string(rec.ReceiptType),
		rec.Currency,
		rec.TotalAmount,
		rec.TaxAmount,
		sql.NullString{},
		sql.NullString{},
		sql.NullString{},
		sql.NullString{},
		sql.NullTime{},
		sql.NullTime{},
		sql.NullString{},
		rec.Lines,
		rec.CreatedBy.String(),
		rec.LastUpdatedBy.String(),
		string(rec.Status),
		now,
		now,
		sql.NullTime{},
	}

	return rec, columns, values
}

func TestReceiptRepository_Create_ReturnsCreateFailedOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec := createReceiptForMock(t)

	mock.ExpectExec("INSERT INTO fiscal_receipts").WillReturnError(sql.ErrConnDone)

	err = repo.Create(context.Background(), rec)
	require.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec := createReceiptForMock(t)

	mock.ExpectExec("INSERT INTO fiscal_receipts").WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), rec)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectQuery("SELECT id, version").WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectQuery("SELECT id, version").WillReturnError(sql.ErrConnDone)

	_, err = repo.GetByID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec, columns, values := buildReceiptRow(t)
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("SELECT id, version").WillReturnRows(rows)

	found, err := repo.GetByID(context.Background(), rec.GetID())
	require.NoError(t, err)
	require.Equal(t, rec.GetID(), found.GetID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByID_RowConversionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	_, columns, values := buildReceiptRow(t)
	values[0] = "invalid"
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("SELECT id, version").WillReturnRows(rows)

	_, err = repo.GetByID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByOrderID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectQuery("WHERE order_id").WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByOrderID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByOrderID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectQuery("WHERE order_id").WillReturnError(sql.ErrConnDone)

	_, err = repo.GetByOrderID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByOrderID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec, columns, values := buildReceiptRow(t)
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("WHERE order_id").WillReturnRows(rows)

	found, err := repo.GetByOrderID(context.Background(), rec.OrderID)
	require.NoError(t, err)
	require.Equal(t, rec.GetID(), found.GetID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_GetByOrderID_RowConversionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	_, columns, values := buildReceiptRow(t)
	values[0] = "invalid"
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("WHERE order_id").WillReturnRows(rows)

	_, err = repo.GetByOrderID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec := createReceiptForMock(t)

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), rec)
	require.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Update_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec := createReceiptForMock(t)

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnError(sql.ErrConnDone)

	err = repo.Update(context.Background(), rec)
	require.ErrorIs(t, err, receipterrors.ErrReceiptUpdateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec := createReceiptForMock(t)
	currentVersion := rec.GetVersion()

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), rec)
	require.NoError(t, err)
	require.Equal(t, currentVersion+1, rec.GetVersion())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Update_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec := createReceiptForMock(t)

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnResult(sqlmock.NewErrorResult(errors.New("rows error")))

	err = repo.Update(context.Background(), rec)
	require.ErrorIs(t, err, receipterrors.ErrReceiptUpdateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Delete_ReturnsDeleteFailedOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnError(sql.ErrConnDone)

	err = repo.Delete(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptDeleteFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), uuidv7.New())
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_Delete_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectExec("UPDATE fiscal_receipts").WillReturnResult(sqlmock.NewErrorResult(errors.New("rows error")))

	err = repo.Delete(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, receipterrors.ErrReceiptDeleteFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	_, columns, values := buildReceiptRow(t)
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("FROM fiscal_receipts").WillReturnRows(rows)

	list, err := repo.List(context.Background(), &repository.ListFilters{})
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	rec, columns, values := buildReceiptRow(t)
	rows := sqlmock.NewRows(columns).AddRow(values...)

	filters := &repository.ListFilters{
		CashRegisterID: &rec.CashRegisterID,
		OrderID:        &rec.OrderID,
	}
	status := aggregate.ReceiptStatusPrinted
	filters.Status = &status

	mock.ExpectQuery("FROM fiscal_receipts").WillReturnRows(rows)

	list, err := repo.List(context.Background(), filters)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_List_RowConversionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	_, columns, values := buildReceiptRow(t)
	values[0] = "invalid"
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("FROM fiscal_receipts").WillReturnRows(rows)

	list, err := repo.List(context.Background(), &repository.ListFilters{})
	require.ErrorIs(t, err, receipterrors.ErrReceiptCreateFailed)
	require.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReceiptRepository_List_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewReceiptRepository(sqlxDB)

	mock.ExpectQuery("FROM fiscal_receipts").WillReturnError(sql.ErrConnDone)

	list, err := repo.List(context.Background(), &repository.ListFilters{})
	require.ErrorIs(t, err, receipterrors.ErrReceiptListFailed)
	require.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}
