package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/jmoiron/sqlx"

    "github.com/basilex/promenade/internal/contexts/accounting/account"
    "github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/account/repository"
    "github.com/basilex/promenade/internal/infrastructure/database"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type AccountRepository struct {
    db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) repository.IAccountRepository {
    return &AccountRepository{db: db}
}

func (r *AccountRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
    if tx, ok := database.GetTx(ctx); ok {
        return tx
    }
    return r.db
}

type accountRow struct {
    ID             string         `db:"id"`
    OrganizationID string         `db:"organization_id"`
    Code           string         `db:"code"`
    Name           string         `db:"name"`
    Type           string         `db:"type"`
    ParentID       sql.NullString `db:"parent_id"`
    Level          int            `db:"level"`
    CurrencyCode   string         `db:"currency_code"`
    IsActive       bool           `db:"is_active"`
    CreatedBy      string         `db:"created_by"`
    LastUpdatedBy  string         `db:"last_updated_by"`
    CreatedAt      time.Time      `db:"created_at"`
    UpdatedAt      time.Time      `db:"updated_at"`
    DeletedAt      sql.NullTime   `db:"deleted_at"`
}

func (r *accountRow) toEntity() (*aggregate.Account, error) {
    id, err := uuidv7.Parse(r.ID)
    if err != nil {
        return nil, fmt.Errorf("invalid account ID: %w", err)
    }
    a := &aggregate.Account{
        Code:         r.Code,
        Name:         r.Name,
        Type:         aggregate.AccountType(r.Type),
        Level:        r.Level,
        CurrencyCode: r.CurrencyCode,
        IsActive:     r.IsActive,
    }
    a.ID = id
    a.CreatedAt = r.CreatedAt
    a.UpdatedAt = r.UpdatedAt
    if r.ParentID.Valid {
        parentID, err := uuidv7.Parse(r.ParentID.String)
        if err != nil {
            return nil, fmt.Errorf("invalid parent_id: %w", err)
        }
        a.ParentID = &parentID
    }
    return a, nil
}

func (r *AccountRepository) Create(ctx context.Context, acc *aggregate.Account) error {
    q := "INSERT INTO accounting_chart_of_accounts (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, created_by, last_updated_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)"
    var pid *string
    if acc.ParentID != nil {
        s := acc.ParentID.String()
        pid = &s
    }
    nilUUID := "00000000-0000-0000-0000-000000000000"
    _, err := r.getExecutor(ctx).ExecContext(ctx, q, acc.ID.String(), nilUUID, acc.Code, acc.Name, string(acc.Type), pid, acc.Level, acc.CurrencyCode, acc.IsActive, nilUUID, nilUUID, acc.CreatedAt, acc.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to create account: %w", err)
    }
    return nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
    q := "SELECT id,organization_id,code,name,type,parent_id,level,currency_code,is_active,created_by,last_updated_by,created_at,updated_at,deleted_at FROM accounting_chart_of_accounts WHERE id=$1 AND deleted_at IS NULL"
    var row accountRow
    err := sqlx.GetContext(ctx, r.getExecutor(ctx), &row, q, id.String())
    if err == sql.ErrNoRows {
        return nil, account.ErrAccountNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get account: %w", err)
    }
    return row.toEntity()
}

func (r *AccountRepository) GetByCode(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.Account, error) {
    q := "SELECT id,organization_id,code,name,type,parent_id,level,currency_code,is_active,created_by,last_updated_by,created_at,updated_at,deleted_at FROM accounting_chart_of_accounts WHERE organization_id=$1 AND code=$2 AND deleted_at IS NULL"
    var row accountRow
    err := sqlx.GetContext(ctx, r.getExecutor(ctx), &row, q, orgID.String(), code)
    if err == sql.ErrNoRows {
        return nil, account.ErrAccountNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get account by code: %w", err)
    }
    return row.toEntity()
}

func (r *AccountRepository) Update(ctx context.Context, acc *aggregate.Account) error {
    q := "UPDATE accounting_chart_of_accounts SET name=$1,is_active=$2,last_updated_by=$3,updated_at=$4 WHERE id=$5 AND deleted_at IS NULL"
    nilUUID := "00000000-0000-0000-0000-000000000000"
    res, err := r.getExecutor(ctx).ExecContext(ctx, q, acc.Name, acc.IsActive, nilUUID, acc.UpdatedAt, acc.ID.String())
    if err != nil {
        return fmt.Errorf("failed to update account: %w", err)
    }
    rows, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rows == 0 {
        return account.ErrAccountNotFound
    }
    return nil
}

func (r *AccountRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
    q := "UPDATE accounting_chart_of_accounts SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL"
    res, err := r.getExecutor(ctx).ExecContext(ctx, q, time.Now(), id.String())
    if err != nil {
        return fmt.Errorf("failed to delete account: %w", err)
    }
    rows, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rows == 0 {
        return account.ErrAccountNotFound
    }
    return nil
}

func (r *AccountRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, includeInactive bool) ([]*aggregate.Account, error) {
    q := "SELECT id,organization_id,code,name,type,parent_id,level,currency_code,is_active,created_by,last_updated_by,created_at,updated_at,deleted_at FROM accounting_chart_of_accounts WHERE organization_id=$1 AND deleted_at IS NULL"
    if !includeInactive {
        q += " AND is_active=true"
    }
    q += " ORDER BY code"
    var rows []accountRow
    err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, q, orgID.String())
    if err != nil {
        return nil, fmt.Errorf("failed to list accounts: %w", err)
    }
    accounts := make([]*aggregate.Account, 0, len(rows))
    for _, row := range rows {
        acc, err := row.toEntity()
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, acc)
    }
    return accounts, nil
}

func (r *AccountRepository) ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error) {
    q := "SELECT id,organization_id,code,name,type,parent_id,level,currency_code,is_active,created_by,last_updated_by,created_at,updated_at,deleted_at FROM accounting_chart_of_accounts WHERE parent_id=$1 AND deleted_at IS NULL ORDER BY code"
    var rows []accountRow
    err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, q, parentID.String())
    if err != nil {
        return nil, fmt.Errorf("failed to list child accounts: %w", err)
    }
    accounts := make([]*aggregate.Account, 0, len(rows))
    for _, row := range rows {
        acc, err := row.toEntity()
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, acc)
    }
    return accounts, nil
}

func (r *AccountRepository) ListByType(ctx context.Context, orgID uuidv7.UUID, accountType aggregate.AccountType) ([]*aggregate.Account, error) {
    q := "SELECT id,organization_id,code,name,type,parent_id,level,currency_code,is_active,created_by,last_updated_by,created_at,updated_at,deleted_at FROM accounting_chart_of_accounts WHERE organization_id=$1 AND type=$2 AND deleted_at IS NULL AND is_active=true ORDER BY code"
    var rows []accountRow
    err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, q, orgID.String(), string(accountType))
    if err != nil {
        return nil, fmt.Errorf("failed to list accounts by type: %w", err)
    }
    accounts := make([]*aggregate.Account, 0, len(rows))
    for _, row := range rows {
        acc, err := row.toEntity()
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, acc)
    }
    return accounts, nil
}

// CreateMany creates multiple accounts in a single transaction (for seeding/imports)
func (r *AccountRepository) CreateMany(ctx context.Context, accounts []*aggregate.Account) error {
    if len(accounts) == 0 {
        return nil
    }

    q := `INSERT INTO accounting_chart_of_accounts 
        (id, organization_id, code, name, type, parent_id, level, currency_code, is_active, created_by, last_updated_by, created_at, updated_at) 
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

    for _, acc := range accounts {
        var parentIDStr *string
        if acc.ParentID != nil {
            s := acc.ParentID.String()
            parentIDStr = &s
        }

        _, err := r.getExecutor(ctx).ExecContext(ctx, q,
            acc.ID.String(),
            acc.OrganizationID.String(),
            acc.Code,
            acc.Name,
            string(acc.Type),
            parentIDStr,
            acc.Level,
            acc.CurrencyCode,
            acc.IsActive,
            acc.LastUpdatedBy.String(),
            acc.LastUpdatedBy.String(),
            acc.CreatedAt,
            acc.UpdatedAt,
        )
        if err != nil {
            return fmt.Errorf("failed to create account %s: %w", acc.Code, err)
        }
    }

    return nil
}

// UpdateMany updates multiple accounts in a single transaction
func (r *AccountRepository) UpdateMany(ctx context.Context, accounts []*aggregate.Account) error {
    if len(accounts) == 0 {
        return nil
    }

    q := `UPDATE accounting_chart_of_accounts 
        SET name=$1, is_active=$2, last_updated_by=$3, updated_at=$4 
        WHERE id=$5 AND deleted_at IS NULL`

    for _, acc := range accounts {
        _, err := r.getExecutor(ctx).ExecContext(ctx, q,
            acc.Name,
            acc.IsActive,
            acc.LastUpdatedBy.String(),
            acc.UpdatedAt,
            acc.ID.String(),
        )
        if err != nil {
            return fmt.Errorf("failed to update account %s: %w", acc.Code, err)
        }
    }

    return nil
}

// GetAllActive retrieves all active accounts for an organization (for caching)
func (r *AccountRepository) GetAllActive(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.Account, error) {
    q := `SELECT id,organization_id,code,name,type,parent_id,level,currency_code,is_active,created_by,last_updated_by,created_at,updated_at,deleted_at 
        FROM accounting_chart_of_accounts 
        WHERE organization_id=$1 AND is_active=true AND deleted_at IS NULL 
        ORDER BY code`

    var rows []accountRow
    err := sqlx.SelectContext(ctx, r.getExecutor(ctx), &rows, q, orgID.String())
    if err != nil {
        return nil, fmt.Errorf("failed to get all active accounts: %w", err)
    }

    accounts := make([]*aggregate.Account, 0, len(rows))
    for _, row := range rows {
        acc, err := row.toEntity()
        if err != nil {
            return nil, err
        }
        accounts = append(accounts, acc)
    }

    return accounts, nil
}
