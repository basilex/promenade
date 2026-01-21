package aggregate

import (
    "testing"

    "github.com/basilex/promenade/internal/contexts/accounting/costcenter"
    "github.com/basilex/promenade/pkg/uuidv7"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewCostCenter(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    tests := []struct {
        name        string
        code        string
        centerName  string
        centerType  CenterType
        expectError error
    }{
        {
            name:        "valid cost center",
            code:        "CC-001",
            centerName:  "Marketing Department",
            centerType:  CenterTypeCost,
            expectError: nil,
        },
        {
            name:        "valid profit center",
            code:        "PC-001",
            centerName:  "Sales Division",
            centerType:  CenterTypeProfit,
            expectError: nil,
        },
        {
            name:        "valid investment center",
            code:        "IC-001",
            centerName:  "R&D Division",
            centerType:  CenterTypeInvestment,
            expectError: nil,
        },
        {
            name:        "empty code",
            code:        "",
            centerName:  "Test Center",
            centerType:  CenterTypeCost,
            expectError: costcenter.ErrCodeRequired,
        },
        {
            name:        "empty name",
            code:        "CC-002",
            centerName:  "",
            centerType:  CenterTypeCost,
            expectError: costcenter.ErrNameRequired,
        },
        {
            name:        "invalid center type",
            code:        "CC-003",
            centerName:  "Test Center",
            centerType:  CenterType("invalid"),
            expectError: costcenter.ErrInvalidCenterType,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cc, err := NewCostCenter(orgID, tt.code, tt.centerName, tt.centerType, userID)

            if tt.expectError != nil {
                assert.ErrorIs(t, err, tt.expectError)
                assert.Nil(t, cc)
            } else {
                require.NoError(t, err)
                require.NotNil(t, cc)
                assert.Equal(t, orgID, cc.OrganizationID)
                assert.Equal(t, tt.code, cc.Code)
                assert.Equal(t, tt.centerName, cc.Name)
                assert.Equal(t, tt.centerType, cc.CenterType)
                assert.Nil(t, cc.ParentID)
                assert.Equal(t, 1, cc.Level)
                assert.True(t, cc.IsActive)
                assert.Nil(t, cc.ManagerID)
            }
        })
    }
}

func TestNewChildCostCenter(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    parentID := uuidv7.New()

    t.Run("valid child cost center", func(t *testing.T) {
        cc, err := NewChildCostCenter(orgID, "CC-002", "Marketing Operations", CenterTypeCost, parentID, 1, userID)
        require.NoError(t, err)
        require.NotNil(t, cc)
        assert.Equal(t, "CC-002", cc.Code)
        assert.Equal(t, "Marketing Operations", cc.Name)
        assert.NotNil(t, cc.ParentID)
        assert.Equal(t, parentID, *cc.ParentID)
        assert.Equal(t, 2, cc.Level)
    })

    t.Run("child with level 3", func(t *testing.T) {
        cc, err := NewChildCostCenter(orgID, "CC-003", "Digital Marketing", CenterTypeCost, parentID, 2, userID)
        require.NoError(t, err)
        assert.Equal(t, 3, cc.Level)
    })

    t.Run("child with invalid data", func(t *testing.T) {
        cc, err := NewChildCostCenter(orgID, "", "Invalid", CenterTypeCost, parentID, 1, userID)
        assert.ErrorIs(t, err, costcenter.ErrCodeRequired)
        assert.Nil(t, cc)
    })
}

func TestCostCenter_SetParent(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    parentID := uuidv7.New()

    cc, err := NewCostCenter(orgID, "CC-001", "Test Center", CenterTypeCost, userID)
    require.NoError(t, err)

    t.Run("set parent", func(t *testing.T) {
        err := cc.SetParent(parentID, 1)
        require.NoError(t, err)
        assert.NotNil(t, cc.ParentID)
        assert.Equal(t, parentID, *cc.ParentID)
        assert.Equal(t, 2, cc.Level)
    })

    t.Run("cannot be own parent", func(t *testing.T) {
        err := cc.SetParent(cc.ID, 1)
        assert.ErrorIs(t, err, costcenter.ErrCannotBeOwnParent)
    })

    t.Run("change parent", func(t *testing.T) {
        newParentID := uuidv7.New()
        err := cc.SetParent(newParentID, 2)
        require.NoError(t, err)
        assert.Equal(t, newParentID, *cc.ParentID)
        assert.Equal(t, 3, cc.Level)
    })
}

func TestCostCenter_SetManager(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()
    managerID := uuidv7.New()

    cc, err := NewCostCenter(orgID, "CC-001", "Test Center", CenterTypeCost, userID)
    require.NoError(t, err)
    assert.Nil(t, cc.ManagerID)

    t.Run("set manager", func(t *testing.T) {
        cc.SetManager(managerID)
        assert.NotNil(t, cc.ManagerID)
        assert.Equal(t, managerID, *cc.ManagerID)
    })

    t.Run("change manager", func(t *testing.T) {
        newManagerID := uuidv7.New()
        cc.SetManager(newManagerID)
        assert.Equal(t, newManagerID, *cc.ManagerID)
    })
}

func TestCostCenter_ActivateDeactivate(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    cc, err := NewCostCenter(orgID, "CC-001", "Test Center", CenterTypeCost, userID)
    require.NoError(t, err)
    require.True(t, cc.IsActive)

    t.Run("deactivate", func(t *testing.T) {
        cc.Deactivate()
        assert.False(t, cc.IsActive)
    })

    t.Run("deactivate already inactive", func(t *testing.T) {
        cc.Deactivate()
        assert.False(t, cc.IsActive)
    })

    t.Run("activate", func(t *testing.T) {
        cc.Activate()
        assert.True(t, cc.IsActive)
    })

    t.Run("activate already active", func(t *testing.T) {
        cc.Activate()
        assert.True(t, cc.IsActive)
    })
}

func TestCostCenter_UpdateDetails(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    cc, err := NewCostCenter(orgID, "CC-001", "Original Name", CenterTypeCost, userID)
    require.NoError(t, err)

    t.Run("update name and description", func(t *testing.T) {
        err := cc.UpdateDetails("New Name", "New description", userID)
        require.NoError(t, err)
        assert.Equal(t, "New Name", cc.Name)
        assert.Equal(t, "New description", cc.Description)
    })

    t.Run("update only name", func(t *testing.T) {
        err := cc.UpdateDetails("Another Name", "", userID)
        require.NoError(t, err)
        assert.Equal(t, "Another Name", cc.Name)
        assert.Equal(t, "", cc.Description)
    })

    t.Run("empty name", func(t *testing.T) {
        err := cc.UpdateDetails("", "Some description", userID)
        assert.ErrorIs(t, err, costcenter.ErrNameRequired)
    })
}

func TestCostCenter_CenterTypes(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    t.Run("cost center type", func(t *testing.T) {
        cc, err := NewCostCenter(orgID, "CC-001", "Cost Center", CenterTypeCost, userID)
        require.NoError(t, err)
        assert.Equal(t, CenterTypeCost, cc.CenterType)
    })

    t.Run("profit center type", func(t *testing.T) {
        cc, err := NewCostCenter(orgID, "PC-001", "Profit Center", CenterTypeProfit, userID)
        require.NoError(t, err)
        assert.Equal(t, CenterTypeProfit, cc.CenterType)
    })

    t.Run("investment center type", func(t *testing.T) {
        cc, err := NewCostCenter(orgID, "IC-001", "Investment Center", CenterTypeInvestment, userID)
        require.NoError(t, err)
        assert.Equal(t, CenterTypeInvestment, cc.CenterType)
    })
}

func TestCostCenter_Hierarchy(t *testing.T) {
    orgID := uuidv7.New()
    userID := uuidv7.New()

    // Create root center (level 1)
    root, err := NewCostCenter(orgID, "CC-ROOT", "Company", CenterTypeCost, userID)
    require.NoError(t, err)
    assert.Equal(t, 1, root.Level)
    assert.Nil(t, root.ParentID)

    // Create child center (level 2)
    child1, err := NewChildCostCenter(orgID, "CC-DEPT", "Department", CenterTypeCost, root.ID, root.Level, userID)
    require.NoError(t, err)
    assert.Equal(t, 2, child1.Level)
    assert.Equal(t, root.ID, *child1.ParentID)

    // Create grandchild center (level 3)
    child2, err := NewChildCostCenter(orgID, "CC-TEAM", "Team", CenterTypeCost, child1.ID, child1.Level, userID)
    require.NoError(t, err)
    assert.Equal(t, 3, child2.Level)
    assert.Equal(t, child1.ID, *child2.ParentID)

    // Verify hierarchy
    assert.Equal(t, 1, root.Level)
    assert.Equal(t, 2, child1.Level)
    assert.Equal(t, 3, child2.Level)
}