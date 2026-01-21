package usecase

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/basilex/promenade/internal/contexts/accounting/audit"
    "github.com/basilex/promenade/internal/contexts/accounting/taxcode"
    "github.com/basilex/promenade/internal/contexts/accounting/taxcode/aggregate"
    "github.com/basilex/promenade/internal/contexts/accounting/taxcode/cache"
    "github.com/basilex/promenade/pkg/uuidv7"
)

// ============================================================================
// Mock Repository
// ============================================================================

type MockTaxCodeRepository struct {
    CreateFunc                 func(ctx context.Context, tc *aggregate.TaxCode) error
    GetByIDFunc                func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error)
    GetByCodeFunc              func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.TaxCode, error)
    UpdateFunc                 func(ctx context.Context, tc *aggregate.TaxCode) error
    DeleteFunc                 func(ctx context.Context, id uuidv7.UUID) error
    ListByOrganizationFunc     func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.TaxCode, error)
    ListActiveFunc             func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.TaxCode, error)
    ListByTypeFunc             func(ctx context.Context, orgID uuidv7.UUID, taxType aggregate.TaxType) ([]*aggregate.TaxCode, error)
}

func (m *MockTaxCodeRepository) Create(ctx context.Context, tc *aggregate.TaxCode) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, tc)
    }
    return errors.New("CreateFunc not implemented")
}

func (m *MockTaxCodeRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, errors.New("GetByIDFunc not implemented")
}

func (m *MockTaxCodeRepository) GetByCode(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.TaxCode, error) {
    if m.GetByCodeFunc != nil {
        return m.GetByCodeFunc(ctx, orgID, code)
    }
    return nil, errors.New("GetByCodeFunc not implemented")
}

func (m *MockTaxCodeRepository) Update(ctx context.Context, tc *aggregate.TaxCode) error {
    if m.UpdateFunc != nil {
        return m.UpdateFunc(ctx, tc)
    }
    return errors.New("UpdateFunc not implemented")
}

func (m *MockTaxCodeRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
    if m.DeleteFunc != nil {
        return m.DeleteFunc(ctx, id)
    }
    return errors.New("DeleteFunc not implemented")
}

func (m *MockTaxCodeRepository) ListByOrganization(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.TaxCode, error) {
    if m.ListByOrganizationFunc != nil {
        return m.ListByOrganizationFunc(ctx, orgID, limit, offset)
    }
    return nil, errors.New("ListByOrganizationFunc not implemented")
}

func (m *MockTaxCodeRepository) ListActive(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.TaxCode, error) {
    if m.ListActiveFunc != nil {
        return m.ListActiveFunc(ctx, orgID)
    }
    return nil, errors.New("ListActiveFunc not implemented")
}

func (m *MockTaxCodeRepository) ListByType(ctx context.Context, orgID uuidv7.UUID, taxType aggregate.TaxType) ([]*aggregate.TaxCode, error) {
    if m.ListByTypeFunc != nil {
        return m.ListByTypeFunc(ctx, orgID, taxType)
    }
    return nil, errors.New("ListByTypeFunc not implemented")
}

// ============================================================================
// Tests
// ============================================================================

func TestCreateTaxCode_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.TaxCode, error) {
            return nil, taxcode.ErrTaxCodeNotFound
        },
        CreateFunc: func(ctx context.Context, tc *aggregate.TaxCode) error {
            return nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    tc, err := uc.CreateTaxCode(context.Background(), orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    require.NoError(t, err)
    assert.NotEqual(t, uuidv7.Nil, tc.ID)
    assert.Equal(t, "VAT20", tc.Code)
    assert.Equal(t, "VAT 20%", tc.Name)
    assert.Equal(t, aggregate.TaxTypeVAT, tc.TaxType)
    assert.Equal(t, 2000, tc.Rate)
    assert.True(t, tc.IsActive)
}

func TestCreateTaxCode_DuplicateCode(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    existing, _ := aggregate.NewTaxCode(orgID, "VAT20", "Existing", aggregate.TaxTypeVAT, 2000, userID)

    mockRepo := &MockTaxCodeRepository{
        GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.TaxCode, error) {
            return existing, nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    _, err := uc.CreateTaxCode(context.Background(), orgID, "VAT20", "Duplicate", aggregate.TaxTypeVAT, 2000, userID)
    assert.Error(t, err)
    assert.Equal(t, taxcode.ErrTaxCodeDuplicate, err)
}

func TestGetTaxCodeByID_Success(t *testing.T) {
    tc, _ := aggregate.NewTaxCode(uuidv7.New(), "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, uuidv7.New())
    tc.ID = uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    retrieved, err := uc.GetTaxCodeByID(context.Background(), tc.ID)
    require.NoError(t, err)
    assert.Equal(t, tc.ID, retrieved.ID)
    assert.Equal(t, "VAT20", retrieved.Code)
}

func TestGetTaxCodeByID_NotFound(t *testing.T) {
    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            return nil, taxcode.ErrTaxCodeNotFound
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    _, err := uc.GetTaxCodeByID(context.Background(), uuidv7.New())
    assert.Error(t, err)
    assert.Equal(t, taxcode.ErrTaxCodeNotFound, err)
}

func TestGetTaxCodeByCode_Success(t *testing.T) {
    orgID := uuidv7.New()
    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, uuidv7.New())

    mockRepo := &MockTaxCodeRepository{
        GetByCodeFunc: func(ctx context.Context, orgID uuidv7.UUID, code string) (*aggregate.TaxCode, error) {
            if code == "VAT20" {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    retrieved, err := uc.GetTaxCodeByCode(context.Background(), orgID, "VAT20")
    require.NoError(t, err)
    assert.Equal(t, "VAT20", retrieved.Code)
}

func TestUpdateTaxCode_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "Old Name", aggregate.TaxTypeVAT, 2000, userID)
    tc.ID = uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
        UpdateFunc: func(ctx context.Context, updated *aggregate.TaxCode) error {
            return nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    updated, err := uc.UpdateTaxCode(context.Background(), tc.ID, 2500, userID)
    require.NoError(t, err)
    assert.Equal(t, 2500, updated.Rate)
}

func TestSetTaxPayableAccount_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID := uuidv7.New()

    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc.ID = uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
        UpdateFunc: func(ctx context.Context, updated *aggregate.TaxCode) error {
            return nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    updated, err := uc.SetTaxPayableAccount(context.Background(), tc.ID, accountID, userID)
    require.NoError(t, err)
    assert.NotNil(t, updated.TaxPayableAccountID)
    assert.Equal(t, accountID, *updated.TaxPayableAccountID)
}

func TestSetTaxReceivableAccount_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    accountID := uuidv7.New()

    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc.ID = uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
        UpdateFunc: func(ctx context.Context, updated *aggregate.TaxCode) error {
            return nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    updated, err := uc.SetTaxReceivableAccount(context.Background(), tc.ID, accountID, userID)
    require.NoError(t, err)
    assert.NotNil(t, updated.TaxReceivableAccountID)
    assert.Equal(t, accountID, *updated.TaxReceivableAccountID)
}

func TestActivateTaxCode_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc.ID = uuidv7.New()
    tc.Deactivate()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
        UpdateFunc: func(ctx context.Context, updated *aggregate.TaxCode) error {
            return nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    activated, err := uc.ActivateTaxCode(context.Background(), tc.ID, userID)
    require.NoError(t, err)
    assert.True(t, activated.IsActive)
}

func TestDeactivateTaxCode_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc.ID = uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
        UpdateFunc: func(ctx context.Context, updated *aggregate.TaxCode) error {
            return nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    deactivated, err := uc.DeactivateTaxCode(context.Background(), tc.ID, userID)
    require.NoError(t, err)
    assert.False(t, deactivated.IsActive)
}

func TestDeleteTaxCode_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc.ID = uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
        DeleteFunc: func(ctx context.Context, id uuidv7.UUID) error {
            return nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    err := uc.DeleteTaxCode(context.Background(), tc.ID)
    require.NoError(t, err)
}

func TestListTaxCodesByOrganization_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    tc1, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc2, _ := aggregate.NewTaxCode(orgID, "VAT5", "VAT 5%", aggregate.TaxTypeVAT, 500, userID)

    mockRepo := &MockTaxCodeRepository{
        ListByOrganizationFunc: func(ctx context.Context, orgID uuidv7.UUID, limit, offset int) ([]*aggregate.TaxCode, error) {
            return []*aggregate.TaxCode{tc1, tc2}, nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    taxCodes, err := uc.ListTaxCodesByOrganization(context.Background(), orgID, 10, 0)
    require.NoError(t, err)
    assert.Len(t, taxCodes, 2)
}

func TestListActiveTaxCodes_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    tc1, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc2, _ := aggregate.NewTaxCode(orgID, "VAT5", "VAT 5%", aggregate.TaxTypeVAT, 500, userID)

    mockRepo := &MockTaxCodeRepository{
        ListActiveFunc: func(ctx context.Context, orgID uuidv7.UUID) ([]*aggregate.TaxCode, error) {
            return []*aggregate.TaxCode{tc1, tc2}, nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    activeTaxCodes, err := uc.ListActiveTaxCodes(context.Background(), orgID)
    require.NoError(t, err)
    assert.Len(t, activeTaxCodes, 2)
    for _, tc := range activeTaxCodes {
        assert.True(t, tc.IsActive)
    }
}

func TestListTaxCodesByType_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    salesTax1, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    salesTax2, _ := aggregate.NewTaxCode(orgID, "VAT5", "VAT 5%", aggregate.TaxTypeVAT, 500, userID)

    mockRepo := &MockTaxCodeRepository{
        ListByTypeFunc: func(ctx context.Context, orgID uuidv7.UUID, taxType aggregate.TaxType) ([]*aggregate.TaxCode, error) {
            if taxType == aggregate.TaxTypeVAT {
                return []*aggregate.TaxCode{salesTax1, salesTax2}, nil
            }
            return []*aggregate.TaxCode{}, nil
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    salesTaxCodes, err := uc.ListTaxCodesByType(context.Background(), orgID, aggregate.TaxTypeVAT)
    require.NoError(t, err)
    assert.Len(t, salesTaxCodes, 2)
    for _, tc := range salesTaxCodes {
        assert.Equal(t, aggregate.TaxTypeVAT, tc.TaxType)
    }
}

func TestGetTaxCodeAndCalculateTax_Success(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    tc, _ := aggregate.NewTaxCode(orgID, "VAT20", "VAT 20%", aggregate.TaxTypeVAT, 2000, userID)
    tc.ID = uuidv7.New()

    mockRepo := &MockTaxCodeRepository{
        GetByIDFunc: func(ctx context.Context, id uuidv7.UUID) (*aggregate.TaxCode, error) {
            if id == tc.ID {
                return tc, nil
            }
            return nil, taxcode.ErrTaxCodeNotFound
        },
    }

    var auditLogger *audit.AuditLogger = nil
    uc := NewTaxCodeUseCase(mockRepo, cache.NewTaxCodeCache(mockRepo), auditLogger)

    // Get tax code through usecase
    retrieved, err := uc.GetTaxCodeByID(context.Background(), tc.ID)
    require.NoError(t, err)
    
    // Calculate tax using aggregate method
    taxAmount := retrieved.CalculateTax(100000)
    assert.Equal(t, int64(20000), taxAmount) // 2000 basis points = 20%, so 20% of 100000 = 20000
}
