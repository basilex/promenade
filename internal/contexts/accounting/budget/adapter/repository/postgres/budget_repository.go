package postgres

import (
    "context"
    "database/sql"
    "time"

    "github.com/basilex/promenade/internal/contexts/accounting/budget"
    "github.com/basilex/promenade/internal/contexts/accounting/budget/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/budget/repository"
    "github.com/basilex/promenade/internal/infrastructure/database"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/jmoiron/sqlx"
)

type budgetRepository struct {
    db *sqlx.DB
}

// NewBudgetRepository creates a new PostgreSQL budget repository
func NewBudgetRepository(db *sqlx.DB) repository.IBudgetRepository {
    return &budgetRepository{db: db}
}

func (r *budgetRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
    if tx, ok := database.GetTx(ctx); ok {
        return tx
    }
    return r.db
}

func (r *budgetRepository) Create(ctx context.Context, budg *aggregate.Budget) error {
    // Insert budget
    query := `
        INSERT INTO accounting_budgets (
            id, version, organization_id, name, fiscal_year, status,
            total_budget_cents, total_actual_cents, approved_by, approved_at,
            last_updated_by, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
        )`

    _, err := r.getExecutor(ctx).ExecContext(
        ctx, query,
        budg.ID.String(),
        budg.Version,
        budg.OrganizationID.String(),
        budg.Name,
        budg.FiscalYear,
        budg.Status,
        budg.TotalBudget,
        budg.TotalActual,
        uuidPtrToString(budg.ApprovedBy),
        budg.ApprovedAt,
        budg.LastUpdatedBy.String(),
        budg.CreatedAt,
        budg.UpdatedAt,
    )
    if err != nil {
        return err
    }

    // Insert budget lines
    if len(budg.Lines) > 0 {
        lineQuery := `
            INSERT INTO accounting_budget_lines (
                id, budget_id, account_id, budget_amount_cents,
                actual_amount_cents, variance_amount_cents, variance_percent,
                description, created_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

        for _, line := range budg.Lines {
            _, err := r.getExecutor(ctx).ExecContext(
                ctx, lineQuery,
                line.ID.String(),
                budg.ID.String(),
                line.AccountID.String(),
                line.BudgetAmount,
                line.ActualAmount,
                line.VarianceAmount,
                line.VariancePercent,
                line.Description,
                time.Now(),
            )
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func (r *budgetRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Budget, error) {
    // Get budget
    query := `
        SELECT id, version, organization_id, name, fiscal_year, status,
               total_budget_cents, total_actual_cents, approved_by, approved_at,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_budgets
        WHERE id = $1 AND deleted_at IS NULL`

    var budg aggregate.Budget
    var orgID, lastUpdatedBy string
    var approvedBy sql.NullString
    var approvedAt sql.NullString

    err := r.getExecutor(ctx).QueryRowxContext(ctx, query, id.String()).Scan(
        &budg.ID,
        &budg.Version,
        &orgID,
        &budg.Name,
        &budg.FiscalYear,
        &budg.Status,
        &budg.TotalBudget,
        &budg.TotalActual,
        &approvedBy,
        &approvedAt,
        &lastUpdatedBy,
        &budg.CreatedAt,
        &budg.UpdatedAt,
        &budg.DeletedAt,
    )

    if err == sql.ErrNoRows {
        return nil, budget.ErrBudgetNotFound
    }
    if err != nil {
        return nil, err
    }

    budg.OrganizationID, _ = uuidv7.Parse(orgID)
    budg.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

    if approvedBy.Valid {
        id, _ := uuidv7.Parse(approvedBy.String)
        budg.ApprovedBy = &id
    }
    if approvedAt.Valid {
        budg.ApprovedAt = &approvedAt.String
    }

    // Get budget lines
    lines, err := r.getBudgetLines(ctx, budg.ID)
    if err != nil {
        return nil, err
    }
    budg.Lines = lines

    return &budg, nil
}

func (r *budgetRepository) GetByName(ctx context.Context, organizationID uuidv7.UUID, name string) (*aggregate.Budget, error) {
    query := `
        SELECT id, version, organization_id, name, fiscal_year, status,
               total_budget_cents, total_actual_cents, approved_by, approved_at,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_budgets
        WHERE organization_id = $1 AND name = $2 AND deleted_at IS NULL`

    var budg aggregate.Budget
    var orgID, lastUpdatedBy string
    var approvedBy sql.NullString
    var approvedAt sql.NullString

    err := r.getExecutor(ctx).QueryRowxContext(ctx, query, organizationID.String(), name).Scan(
        &budg.ID,
        &budg.Version,
        &orgID,
        &budg.Name,
        &budg.FiscalYear,
        &budg.Status,
        &budg.TotalBudget,
        &budg.TotalActual,
        &approvedBy,
        &approvedAt,
        &lastUpdatedBy,
        &budg.CreatedAt,
        &budg.UpdatedAt,
        &budg.DeletedAt,
    )

    if err == sql.ErrNoRows {
        return nil, budget.ErrBudgetNotFound
    }
    if err != nil {
        return nil, err
    }

    budg.OrganizationID, _ = uuidv7.Parse(orgID)
    budg.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

    if approvedBy.Valid {
        id, _ := uuidv7.Parse(approvedBy.String)
        budg.ApprovedBy = &id
    }
    if approvedAt.Valid {
        budg.ApprovedAt = &approvedAt.String
    }

    // Get budget lines
    lines, err := r.getBudgetLines(ctx, budg.ID)
    if err != nil {
        return nil, err
    }
    budg.Lines = lines

    return &budg, nil
}

func (r *budgetRepository) Update(ctx context.Context, budg *aggregate.Budget) error {
    // Update budget
    query := `
        UPDATE accounting_budgets
        SET version = $2,
            status = $3,
            total_budget_cents = $4,
            total_actual_cents = $5,
            approved_by = $6,
            approved_at = $7,
            last_updated_by = $8,
            updated_at = $9
        WHERE id = $1 AND deleted_at IS NULL`

    result, err := r.getExecutor(ctx).ExecContext(
        ctx, query,
        budg.ID.String(),
        budg.Version,
        budg.Status,
        budg.TotalBudget,
        budg.TotalActual,
        uuidPtrToString(budg.ApprovedBy),
        budg.ApprovedAt,
        budg.LastUpdatedBy.String(),
        budg.UpdatedAt,
    )

    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return budget.ErrBudgetNotFound
    }

    // Delete existing lines
    deleteQuery := `DELETE FROM accounting_budget_lines WHERE budget_id = $1`
    _, err = r.getExecutor(ctx).ExecContext(ctx, deleteQuery, budg.ID.String())
    if err != nil {
        return err
    }

    // Insert updated lines
    if len(budg.Lines) > 0 {
        lineQuery := `
            INSERT INTO accounting_budget_lines (
                id, budget_id, account_id, budget_amount_cents,
                actual_amount_cents, variance_amount_cents, variance_percent,
                description, created_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

        for _, line := range budg.Lines {
            _, err := r.getExecutor(ctx).ExecContext(
                ctx, lineQuery,
                line.ID.String(),
                budg.ID.String(),
                line.AccountID.String(),
                line.BudgetAmount,
                line.ActualAmount,
                line.VarianceAmount,
                line.VariancePercent,
                line.Description,
                time.Now(),
            )
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func (r *budgetRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
    query := `
        UPDATE accounting_budgets
        SET deleted_at = $2
        WHERE id = $1 AND deleted_at IS NULL`

    result, err := r.getExecutor(ctx).ExecContext(ctx, query, id.String(), time.Now())
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return budget.ErrBudgetNotFound
    }

    return nil
}

func (r *budgetRepository) ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.Budget, error) {
    query := `
        SELECT id, version, organization_id, name, fiscal_year, status,
               total_budget_cents, total_actual_cents, approved_by, approved_at,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_budgets
        WHERE organization_id = $1 AND deleted_at IS NULL
        ORDER BY fiscal_year DESC, name
        LIMIT $2 OFFSET $3`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), limit, offset)
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    return r.scanBudgets(ctx, rows)
}

func (r *budgetRepository) ListByStatus(ctx context.Context, organizationID uuidv7.UUID, status aggregate.BudgetStatus) ([]*aggregate.Budget, error) {
    query := `
        SELECT id, version, organization_id, name, fiscal_year, status,
               total_budget_cents, total_actual_cents, approved_by, approved_at,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_budgets
        WHERE organization_id = $1 AND status = $2 AND deleted_at IS NULL
        ORDER BY fiscal_year DESC, name`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), status)
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    return r.scanBudgets(ctx, rows)
}

func (r *budgetRepository) ListActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.Budget, error) {
    query := `
        SELECT id, version, organization_id, name, fiscal_year, status,
               total_budget_cents, total_actual_cents, approved_by, approved_at,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_budgets
        WHERE organization_id = $1 AND status = 'active' AND deleted_at IS NULL
        ORDER BY fiscal_year DESC, name`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String())
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    return r.scanBudgets(ctx, rows)
}

func (r *budgetRepository) scanBudgets(ctx context.Context, rows *sqlx.Rows) ([]*aggregate.Budget, error) {
    var budgets []*aggregate.Budget

    for rows.Next() {
        var budg aggregate.Budget
        var orgID, lastUpdatedBy string
        var approvedBy sql.NullString
        var approvedAt sql.NullString

        err := rows.Scan(
            &budg.ID,
            &budg.Version,
            &orgID,
            &budg.Name,
            &budg.FiscalYear,
            &budg.Status,
            &budg.TotalBudget,
            &budg.TotalActual,
            &approvedBy,
            &approvedAt,
            &lastUpdatedBy,
            &budg.CreatedAt,
            &budg.UpdatedAt,
            &budg.DeletedAt,
        )
        if err != nil {
            return nil, err
        }

        budg.OrganizationID, _ = uuidv7.Parse(orgID)
        budg.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

        if approvedBy.Valid {
            id, _ := uuidv7.Parse(approvedBy.String)
            budg.ApprovedBy = &id
        }
        if approvedAt.Valid {
            budg.ApprovedAt = &approvedAt.String
        }

        // Get budget lines
        lines, err := r.getBudgetLines(ctx, budg.ID)
        if err != nil {
            return nil, err
        }
        budg.Lines = lines

        budgets = append(budgets, &budg)
    }

    return budgets, nil
}

func (r *budgetRepository) getBudgetLines(ctx context.Context, budgetID uuidv7.UUID) ([]*aggregate.BudgetLine, error) {
    query := `
        SELECT id, account_id, budget_amount_cents, actual_amount_cents,
               variance_amount_cents, variance_percent, description
        FROM accounting_budget_lines
        WHERE budget_id = $1
        ORDER BY id`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, budgetID.String())
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    var lines []*aggregate.BudgetLine

    for rows.Next() {
        var line aggregate.BudgetLine
        var accountID string

        err := rows.Scan(
            &line.ID,
            &accountID,
            &line.BudgetAmount,
            &line.ActualAmount,
            &line.VarianceAmount,
            &line.VariancePercent,
            &line.Description,
        )
        if err != nil {
            return nil, err
        }

        line.AccountID, _ = uuidv7.Parse(accountID)
        lines = append(lines, &line)
    }

    return lines, nil
}

func uuidPtrToString(id *uuidv7.UUID) *string {
    if id == nil {
        return nil
    }
    s := id.String()
    return &s
}