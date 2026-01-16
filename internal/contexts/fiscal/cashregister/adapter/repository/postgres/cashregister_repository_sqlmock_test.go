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

	"github.com/basilex/promenade/internal/contexts/fiscal/cashregister"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func buildCashRegisterRow() (cashregister.CashRegister, []string, []driver.Value) {
	orgID := uuidv7.New()
	userID := uuidv7.New()
	cr, _ := cashregister.NewCashRegister(orgID, "FN-ROW", "Model", userID)
	_ = cr.Activate("LIC-ROW", userID)

	columns := []string{
		"id",
		"version",
		"organization_id",
		"fiscal_number",
		"model",
		"status",
		"license_key",
		"last_sync_at",
		"last_updated_by",
		"created_at",
		"updated_at",
		"deleted_at",
	}

	values := []driver.Value{
		cr.GetID().String(),
		cr.GetVersion(),
		cr.OrganizationID.String(),
		cr.FiscalNumber,
		cr.Model,
		string(cr.Status),
		sql.NullString{String: cr.LicenseKey, Valid: true},
		sql.NullTime{Time: time.Now(), Valid: true},
		cr.LastUpdatedBy.String(),
		cr.CreatedAt,
		cr.UpdatedAt,
		sql.NullTime{},
	}

	return *cr, columns, values
}

func TestCashRegisterRepository_Create_ReturnsCreateFailedOnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, err := cashregister.NewCashRegister(uuidv7.New(), "FN-TEST", "Model", uuidv7.New())
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO fiscal_cash_registers").WillReturnError(sql.ErrConnDone)

	err = repo.Create(context.Background(), cr)
	require.ErrorIs(t, err, cashregister.ErrCashRegisterCreateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, err := cashregister.NewCashRegister(uuidv7.New(), "FN-SUCCESS", "Model", uuidv7.New())
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO fiscal_cash_registers").WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(context.Background(), cr)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectQuery("SELECT id, version").WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, cashregister.ErrCashRegisterNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, columns, values := buildCashRegisterRow()
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("SELECT id, version").WillReturnRows(rows)

	found, err := repo.GetByID(context.Background(), cr.GetID())
	require.NoError(t, err)
	require.Equal(t, cr.GetID(), found.GetID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_GetByID_RowConversionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	_, columns, values := buildCashRegisterRow()
	values[0] = "invalid"
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("SELECT id, version").WillReturnRows(rows)

	_, err = repo.GetByID(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, cashregister.ErrCashRegisterCreateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_GetByFiscalNumber_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectQuery("WHERE fiscal_number").WillReturnError(sql.ErrNoRows)

	_, err = repo.GetByFiscalNumber(context.Background(), "FN-NOT-FOUND")
	require.ErrorIs(t, err, cashregister.ErrCashRegisterNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_GetByFiscalNumber_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, columns, values := buildCashRegisterRow()
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("WHERE fiscal_number").WillReturnRows(rows)

	found, err := repo.GetByFiscalNumber(context.Background(), cr.FiscalNumber)
	require.NoError(t, err)
	require.Equal(t, cr.GetID(), found.GetID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, err := cashregister.NewCashRegister(uuidv7.New(), "FN-UPDATE", "Model", uuidv7.New())
	require.NoError(t, err)

	mock.ExpectExec("UPDATE fiscal_cash_registers").WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), cr)
	require.ErrorIs(t, err, cashregister.ErrCashRegisterNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, err := cashregister.NewCashRegister(uuidv7.New(), "FN-UPD", "Model", uuidv7.New())
	require.NoError(t, err)
	currentVersion := cr.GetVersion()

	mock.ExpectExec("UPDATE fiscal_cash_registers").WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), cr)
	require.NoError(t, err)
	require.Equal(t, currentVersion+1, cr.GetVersion())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_Update_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, err := cashregister.NewCashRegister(uuidv7.New(), "FN-ROWS", "Model", uuidv7.New())
	require.NoError(t, err)

	mock.ExpectExec("UPDATE fiscal_cash_registers").WillReturnResult(sqlmock.NewErrorResult(errors.New("rows error")))

	err = repo.Update(context.Background(), cr)
	require.ErrorIs(t, err, cashregister.ErrCashRegisterUpdateFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectExec("UPDATE fiscal_cash_registers").WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, cashregister.ErrCashRegisterNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectExec("UPDATE fiscal_cash_registers").WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), uuidv7.New())
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_Delete_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectExec("UPDATE fiscal_cash_registers").WillReturnResult(sqlmock.NewErrorResult(errors.New("rows error")))

	err = repo.Delete(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, cashregister.ErrCashRegisterDeleteFailed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_ListActive_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, columns, values := buildCashRegisterRow()
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("FROM fiscal_cash_registers").WillReturnRows(rows)

	list, err := repo.ListActive(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, cr.GetID(), list[0].GetID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_ListActive_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectQuery("FROM fiscal_cash_registers").WillReturnError(sql.ErrConnDone)

	list, err := repo.ListActive(context.Background())
	require.ErrorIs(t, err, cashregister.ErrCashRegisterCreateFailed)
	require.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_List_WithFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, columns, values := buildCashRegisterRow()
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("FROM fiscal_cash_registers").WillReturnRows(rows)

	filters := &cashregister.ListFilters{OrganizationID: &cr.OrganizationID}
	list, err := repo.List(context.Background(), filters)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, cr.GetID(), list[0].GetID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_List_RowConversionError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	_, columns, values := buildCashRegisterRow()
	values[0] = "invalid"
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("FROM fiscal_cash_registers").WillReturnRows(rows)

	list, err := repo.List(context.Background(), &cashregister.ListFilters{})
	require.ErrorIs(t, err, cashregister.ErrCashRegisterCreateFailed)
	require.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_List_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectQuery("FROM fiscal_cash_registers").WillReturnError(sql.ErrConnDone)

	list, err := repo.List(context.Background(), &cashregister.ListFilters{})
	require.ErrorIs(t, err, cashregister.ErrCashRegisterCreateFailed)
	require.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_GetByLocation_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	cr, columns, values := buildCashRegisterRow()
	rows := sqlmock.NewRows(columns).AddRow(values...)

	mock.ExpectQuery("WHERE organization_id").WillReturnRows(rows)

	list, err := repo.GetByLocation(context.Background(), cr.OrganizationID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, cr.GetID(), list[0].GetID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCashRegisterRepository_GetByLocation_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("db close: %v", err)
		}
	}()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewCashRegisterRepository(sqlxDB)

	mock.ExpectQuery("WHERE organization_id").WillReturnError(sql.ErrConnDone)

	list, err := repo.GetByLocation(context.Background(), uuidv7.New())
	require.ErrorIs(t, err, cashregister.ErrCashRegisterCreateFailed)
	require.Nil(t, list)
	require.NoError(t, mock.ExpectationsWereMet())
}
