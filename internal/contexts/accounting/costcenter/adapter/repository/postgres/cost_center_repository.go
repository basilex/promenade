package postgres

import (
    "context"
    "database/sql"
    "time"

    "github.com/jmoiron/sqlx"
	
    "github.com/basilex/promenade/internal/contexts/accounting/costcenter"
    "github.com/basilex/promenade/internal/contexts/accounting/costcenter/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/costcenter/repository"
    "github.com/basilex/promenade/internal/infrastructure/database"
    "github.com/basilex/promenade/pkg/uuidv7"
)

type costCenterRepository struct {
    db *sqlx.DB
}

// NewCostCenterRepository creates a new PostgreSQL cost center repository
func NewCostCenterRepository(db *sqlx.DB) repository.ICostCenterRepository {
    return &costCenterRepository{db: db}
}

func (r *costCenterRepository) getExecutor(ctx context.Context) sqlx.ExtContext {
    if tx, ok := database.GetTx(ctx); ok {
        return tx
    }
    return r.db
}

func (r *costCenterRepository) Create(ctx context.Context, cc *aggregate.CostCenter) error {
    query := `
        INSERT INTO accounting_cost_centers (
            id, version, organization_id, code, name, center_type,
            parent_id, level, manager_id, is_active, description,
            last_updated_by, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
        )`

    _, err := r.getExecutor(ctx).ExecContext(
        ctx, query,
        cc.ID.String(),
        cc.Version,
        cc.OrganizationID.String(),
        cc.Code,
        cc.Name,
        cc.CenterType,
        uuidPtrToString(cc.ParentID),
        cc.Level,
        uuidPtrToString(cc.ManagerID),
        cc.IsActive,
        cc.Description,
        cc.LastUpdatedBy.String(),
        cc.CreatedAt,
        cc.UpdatedAt,
    )

    return err
}

func (r *costCenterRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.CostCenter, error) {
    query := `
        SELECT id, version, organization_id, code, name, center_type,
               parent_id, level, manager_id, is_active, description,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_cost_centers
        WHERE id = $1 AND deleted_at IS NULL`

    var cc aggregate.CostCenter
    var orgID, lastUpdatedBy string
    var parentID, managerID sql.NullString

    err := r.getExecutor(ctx).QueryRowxContext(ctx, query, id.String()).Scan(
        &cc.ID,
        &cc.Version,
        &orgID,
        &cc.Code,
        &cc.Name,
        &cc.CenterType,
        &parentID,
        &cc.Level,
        &managerID,
        &cc.IsActive,
        &cc.Description,
        &lastUpdatedBy,
        &cc.CreatedAt,
        &cc.UpdatedAt,
        &cc.DeletedAt,
    )

    if err == sql.ErrNoRows {
        return nil, costcenter.ErrCostCenterNotFound
    }
    if err != nil {
        return nil, err
    }

    cc.OrganizationID, _ = uuidv7.Parse(orgID)
    cc.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

    if parentID.Valid {
        id, _ := uuidv7.Parse(parentID.String)
        cc.ParentID = &id
    }
    if managerID.Valid {
        id, _ := uuidv7.Parse(managerID.String)
        cc.ManagerID = &id
    }

    return &cc, nil
}

func (r *costCenterRepository) GetByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.CostCenter, error) {
    query := `
        SELECT id, version, organization_id, code, name, center_type,
               parent_id, level, manager_id, is_active, description,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_cost_centers
        WHERE organization_id = $1 AND code = $2 AND deleted_at IS NULL`

    var cc aggregate.CostCenter
    var orgID, lastUpdatedBy string
    var parentID, managerID sql.NullString

    err := r.getExecutor(ctx).QueryRowxContext(ctx, query, organizationID.String(), code).Scan(
        &cc.ID,
        &cc.Version,
        &orgID,
        &cc.Code,
        &cc.Name,
        &cc.CenterType,
        &parentID,
        &cc.Level,
        &managerID,
        &cc.IsActive,
        &cc.Description,
        &lastUpdatedBy,
        &cc.CreatedAt,
        &cc.UpdatedAt,
        &cc.DeletedAt,
    )

    if err == sql.ErrNoRows {
        return nil, costcenter.ErrCostCenterNotFound
    }
    if err != nil {
        return nil, err
    }

    cc.OrganizationID, _ = uuidv7.Parse(orgID)
    cc.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

    if parentID.Valid {
        id, _ := uuidv7.Parse(parentID.String)
        cc.ParentID = &id
    }
    if managerID.Valid {
        id, _ := uuidv7.Parse(managerID.String)
        cc.ManagerID = &id
    }

    return &cc, nil
}

func (r *costCenterRepository) Update(ctx context.Context, cc *aggregate.CostCenter) error {
    query := `
        UPDATE accounting_cost_centers
        SET version = $2,
            name = $3,
            parent_id = $4,
            level = $5,
            manager_id = $6,
            is_active = $7,
            description = $8,
            last_updated_by = $9,
            updated_at = $10
        WHERE id = $1 AND deleted_at IS NULL`

    result, err := r.getExecutor(ctx).ExecContext(
        ctx, query,
        cc.ID.String(),
        cc.Version,
        cc.Name,
        uuidPtrToString(cc.ParentID),
        cc.Level,
        uuidPtrToString(cc.ManagerID),
        cc.IsActive,
        cc.Description,
        cc.LastUpdatedBy.String(),
        cc.UpdatedAt,
    )

    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        return costcenter.ErrCostCenterNotFound
    }

    return nil
}

func (r *costCenterRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
    query := `
        UPDATE accounting_cost_centers
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
        return costcenter.ErrCostCenterNotFound
    }

    return nil
}

func (r *costCenterRepository) ListByOrganization(ctx context.Context, organizationID uuidv7.UUID, limit, offset int) ([]*aggregate.CostCenter, error) {
    query := `
        SELECT id, version, organization_id, code, name, center_type,
               parent_id, level, manager_id, is_active, description,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_cost_centers
        WHERE organization_id = $1 AND deleted_at IS NULL
        ORDER BY code
        LIMIT $2 OFFSET $3`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), limit, offset)
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    return r.scanCostCenters(rows)
}

func (r *costCenterRepository) ListChildren(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
    query := `
        SELECT id, version, organization_id, code, name, center_type,
               parent_id, level, manager_id, is_active, description,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_cost_centers
        WHERE parent_id = $1 AND deleted_at IS NULL
        ORDER BY code`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, parentID.String())
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    return r.scanCostCenters(rows)
}

func (r *costCenterRepository) ListByType(ctx context.Context, organizationID uuidv7.UUID, centerType aggregate.CenterType) ([]*aggregate.CostCenter, error) {
    query := `
        SELECT id, version, organization_id, code, name, center_type,
               parent_id, level, manager_id, is_active, description,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_cost_centers
        WHERE organization_id = $1 AND center_type = $2 AND deleted_at IS NULL
        ORDER BY code`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String(), centerType)
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    return r.scanCostCenters(rows)
}

func (r *costCenterRepository) ListActive(ctx context.Context, organizationID uuidv7.UUID) ([]*aggregate.CostCenter, error) {
    query := `
        SELECT id, version, organization_id, code, name, center_type,
               parent_id, level, manager_id, is_active, description,
               last_updated_by, created_at, updated_at, deleted_at
        FROM accounting_cost_centers
        WHERE organization_id = $1 AND is_active = true AND deleted_at IS NULL
        ORDER BY code`

    rows, err := r.getExecutor(ctx).QueryxContext(ctx, query, organizationID.String())
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    return r.scanCostCenters(rows)
}

func (r *costCenterRepository) scanCostCenters(rows *sqlx.Rows) ([]*aggregate.CostCenter, error) {
    var costCenters []*aggregate.CostCenter

    for rows.Next() {
        var cc aggregate.CostCenter
        var orgID, lastUpdatedBy string
        var parentID, managerID sql.NullString

        err := rows.Scan(
            &cc.ID,
            &cc.Version,
            &orgID,
            &cc.Code,
            &cc.Name,
            &cc.CenterType,
            &parentID,
            &cc.Level,
            &managerID,
            &cc.IsActive,
            &cc.Description,
            &lastUpdatedBy,
            &cc.CreatedAt,
            &cc.UpdatedAt,
            &cc.DeletedAt,
        )
        if err != nil {
            return nil, err
        }

        cc.OrganizationID, _ = uuidv7.Parse(orgID)
        cc.LastUpdatedBy, _ = uuidv7.Parse(lastUpdatedBy)

        if parentID.Valid {
            id, _ := uuidv7.Parse(parentID.String)
            cc.ParentID = &id
        }
        if managerID.Valid {
            id, _ := uuidv7.Parse(managerID.String)
            cc.ManagerID = &id
        }

        costCenters = append(costCenters, &cc)
    }

    return costCenters, nil
}

func uuidPtrToString(id *uuidv7.UUID) *string {
    if id == nil {
        return nil
    }
    s := id.String()
    return &s
}