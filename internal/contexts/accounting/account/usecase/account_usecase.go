package usecase

import (
    "context"

    "github.com/basilex/promenade/internal/contexts/accounting/account"
    "github.com/basilex/promenade/internal/contexts/accounting/account/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/account/cache"
    "github.com/basilex/promenade/internal/contexts/accounting/account/repository"
    "github.com/basilex/promenade/internal/contexts/accounting/audit"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// IAccountUseCase defines the interface for account business logic
type IAccountUseCase interface {
    CreateAccount(ctx context.Context, organizationID uuidv7.UUID, code, name string, accountType aggregate.AccountType, parentID *uuidv7.UUID, createdBy uuidv7.UUID) (*aggregate.Account, error)
    GetAccountByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error)
    GetAccountByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.Account, error)
    UpdateAccount(ctx context.Context, id uuidv7.UUID, name string, updatedBy uuidv7.UUID) (*aggregate.Account, error)
    SetParent(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) error
    ActivateAccount(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error
    DeactivateAccount(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error
    DeleteAccount(ctx context.Context, id uuidv7.UUID) error
    ListAccountsByOrganization(ctx context.Context, organizationID uuidv7.UUID, includeInactive bool) ([]*aggregate.Account, error)
    ListAccountsByType(ctx context.Context, organizationID uuidv7.UUID, accountType aggregate.AccountType) ([]*aggregate.Account, error)
    ListChildAccounts(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error)
}

type accountUseCase struct {
    accountRepo  repository.IAccountRepository
    accountCache *cache.AccountCache
    auditLogger  *audit.AuditLogger
}

// NewAccountUseCase creates a new account use case
func NewAccountUseCase(
    accountRepo repository.IAccountRepository,
    accountCache *cache.AccountCache,
    auditLogger *audit.AuditLogger,
) IAccountUseCase {
    return &accountUseCase{
        accountRepo:  accountRepo,
        accountCache: accountCache,
        auditLogger:  auditLogger,
    }
}

func (uc *accountUseCase) CreateAccount(
    ctx context.Context,
    organizationID uuidv7.UUID,
    code, name string,
    accountType aggregate.AccountType,
    parentID *uuidv7.UUID,
    createdBy uuidv7.UUID,
) (*aggregate.Account, error) {
    // Check if account with this code already exists (check cache first)
    existing, _ := uc.accountCache.GetByCode(ctx, organizationID, code)
    if existing != nil {
        return nil, account.ErrAccountCodeAlreadyExists
    }

    // Double-check in repository
    existing, err := uc.accountRepo.GetByCode(ctx, organizationID, code)
    if err == nil && existing != nil {
        return nil, account.ErrAccountCodeAlreadyExists
    }

    // Create new account
    var acc *aggregate.Account
    var createErr error

    if parentID != nil {
        // Get parent to determine level
        parent, err := uc.accountRepo.GetByID(ctx, *parentID)
        if err != nil {
            return nil, err
        }
        acc, createErr = aggregate.NewChildAccount(organizationID, code, name, accountType, "UAH", *parentID, parent.Level, createdBy)
    } else {
        acc, createErr = aggregate.NewAccount(organizationID, code, name, accountType, "UAH", createdBy)
    }

    if createErr != nil {
        return nil, createErr
    }

    // Save to repository
    if err := uc.accountRepo.Create(ctx, acc); err != nil {
        return nil, err
    }

    // Log audit
    if uc.auditLogger != nil {
        _ = uc.auditLogger.LogCreate(ctx, audit.AuditRecord{
            EntityType:     "account",
            EntityID:       acc.ID,
            Action:         "create",
            OrganizationID: organizationID,
            UserID:         createdBy,
            Details:        map[string]interface{}{"code": code, "name": name, "type": accountType},
        })
    }

    // Invalidate cache
    uc.accountCache.Invalidate(organizationID)

    return acc, nil
}

func (uc *accountUseCase) GetAccountByID(ctx context.Context, id uuidv7.UUID) (*aggregate.Account, error) {
    // Try cache first
    if acc, err := uc.accountCache.GetByID(ctx, id); err == nil && acc != nil {
        return acc, nil
    }

    // Fallback to repository
    return uc.accountRepo.GetByID(ctx, id)
}

func (uc *accountUseCase) GetAccountByCode(ctx context.Context, organizationID uuidv7.UUID, code string) (*aggregate.Account, error) {
    // Try cache first
    if acc, err := uc.accountCache.GetByCode(ctx, organizationID, code); err == nil && acc != nil {
        return acc, nil
    }

    // Fallback to repository
    return uc.accountRepo.GetByCode(ctx, organizationID, code)
}

func (uc *accountUseCase) UpdateAccount(ctx context.Context, id uuidv7.UUID, name string, updatedBy uuidv7.UUID) (*aggregate.Account, error) {
    // Get account
    acc, err := uc.accountRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Capture before state for audit
    beforeName := acc.Name

    // Update name
    if err := acc.UpdateName(name); err != nil {
        return nil, err
    }

    // Save
    if err := uc.accountRepo.Update(ctx, acc); err != nil {
        return nil, err
    }

    // Log audit with before/after
    if uc.auditLogger != nil {
        _ = uc.auditLogger.LogUpdate(ctx, audit.AuditRecord{
            EntityType:     "account",
            EntityID:       acc.ID,
            Action:         "update",
            OrganizationID: acc.OrganizationID,
            UserID:         updatedBy,
            ChangesBefore:  map[string]interface{}{"name": beforeName},
            ChangesAfter:   map[string]interface{}{"name": name},
        })
    }

    // Invalidate cache
    uc.accountCache.Invalidate(acc.OrganizationID)

    return acc, nil
}

func (uc *accountUseCase) SetParent(ctx context.Context, id, parentID uuidv7.UUID, updatedBy uuidv7.UUID) error {
    // Get account
    acc, err := uc.accountRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    // Get parent to determine level
    parent, err := uc.accountRepo.GetByID(ctx, parentID)
    if err != nil {
        return err
    }

    // Set parent manually
    acc.ParentID = &parentID
    acc.Level = parent.Level + 1
    acc.Touch()

    // Save
    return uc.accountRepo.Update(ctx, acc)
}

func (uc *accountUseCase) ActivateAccount(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error {
    // Get account
    acc, err := uc.accountRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    // Activate
    if err := acc.Activate(); err != nil {
        return err
    }

    // Save
    if err := uc.accountRepo.Update(ctx, acc); err != nil {
        return err
    }

    // Log audit
    if uc.auditLogger != nil {
        _ = uc.auditLogger.LogUpdate(ctx, audit.AuditRecord{
            EntityType:     "account",
            EntityID:       acc.ID,
            Action:         "activate",
            OrganizationID: acc.OrganizationID,
            UserID:         updatedBy,
            Details:        map[string]interface{}{"is_active": true},
        })
    }

    // Invalidate cache
    uc.accountCache.Invalidate(acc.OrganizationID)

    return nil
}

func (uc *accountUseCase) DeactivateAccount(ctx context.Context, id uuidv7.UUID, updatedBy uuidv7.UUID) error {
    // Get account
    acc, err := uc.accountRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    // Deactivate
    if err := acc.Deactivate(); err != nil {
        return err
    }

    // Save
    if err := uc.accountRepo.Update(ctx, acc); err != nil {
        return err
    }

    // Log audit
    if uc.auditLogger != nil {
        _ = uc.auditLogger.LogUpdate(ctx, audit.AuditRecord{
            EntityType:     "account",
            EntityID:       acc.ID,
            Action:         "deactivate",
            OrganizationID: acc.OrganizationID,
            UserID:         updatedBy,
            Details:        map[string]interface{}{"is_active": false},
        })
    }

    // Invalidate cache
    uc.accountCache.Invalidate(acc.OrganizationID)

    return nil
}

func (uc *accountUseCase) DeleteAccount(ctx context.Context, id uuidv7.UUID) error {
    // Get account for audit
    acc, err := uc.accountRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    // Check if account has children
    children, err := uc.accountRepo.ListChildren(ctx, id)
    if err != nil {
        return err
    }
    if len(children) > 0 {
        return account.ErrAccountHasChildren
    }

    // Delete
    if err := uc.accountRepo.Delete(ctx, id); err != nil {
        return err
    }

    // Log audit
    if uc.auditLogger != nil {
        _ = uc.auditLogger.LogCreate(ctx, audit.AuditRecord{
            EntityType:     "account",
            EntityID:       acc.ID,
            Action:         "delete",
            OrganizationID: acc.OrganizationID,
            UserID:         uuidv7.Nil, // no user context in delete
            Details:        map[string]interface{}{"code": acc.Code, "name": acc.Name},
        })
    }

    // Invalidate cache
    uc.accountCache.Invalidate(acc.OrganizationID)

    return nil
}

func (uc *accountUseCase) ListAccountsByOrganization(ctx context.Context, organizationID uuidv7.UUID, includeInactive bool) ([]*aggregate.Account, error) {
    if includeInactive {
        return uc.accountRepo.ListByOrganization(ctx, organizationID, includeInactive)
    }

    // For active accounts, try cache first
    accounts, err := uc.accountCache.GetAllActive(ctx, organizationID)
    if err == nil && len(accounts) > 0 {
        return accounts, nil
    }

    // Fallback to repository
    return uc.accountRepo.ListByOrganization(ctx, organizationID, includeInactive)
}

func (uc *accountUseCase) ListAccountsByType(ctx context.Context, organizationID uuidv7.UUID, accountType aggregate.AccountType) ([]*aggregate.Account, error) {
    return uc.accountRepo.ListByType(ctx, organizationID, accountType)
}

func (uc *accountUseCase) ListChildAccounts(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Account, error) {
    return uc.accountRepo.ListChildren(ctx, parentID)
}