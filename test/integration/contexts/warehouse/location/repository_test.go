package location_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/warehouse/location"
	locationRepo "github.com/basilex/promenade/internal/contexts/warehouse/location/adapter/repository/postgres"
	"github.com/basilex/promenade/test/integration"
)

func TestLocationRepository_Create(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	loc, err := location.NewLocation("WH01", "Main Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, err)

	err = repo.Create(ctx, loc)
	require.NoError(t, err)

	// Verify created
	retrieved, err := repo.GetByID(ctx, loc.GetID())
	require.NoError(t, err)
	assert.Equal(t, loc.GetID(), retrieved.GetID())
	assert.Equal(t, "WH01", retrieved.Code)
	assert.Equal(t, "Main Warehouse", retrieved.Name)
	assert.Equal(t, location.LocationTypeWarehouse, retrieved.Type)
}

func TestLocationRepository_Update(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create location
	loc, _ := location.NewLocation("WH01", "Main Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, loc))

	// Update
	loc.Name = "Updated Warehouse"
	loc.Description = "New description"
	err := repo.Update(ctx, loc)
	require.NoError(t, err)

	// Verify updated
	retrieved, err := repo.GetByID(ctx, loc.GetID())
	require.NoError(t, err)
	assert.Equal(t, "Updated Warehouse", retrieved.Name)
	assert.Equal(t, "New description", retrieved.Description)
	assert.Equal(t, 2, retrieved.GetVersion()) // Version incremented
}

func TestLocationRepository_Update_OptimisticLocking(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create location
	loc, _ := location.NewLocation("WH01", "Main Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, loc))

	// Get two copies
	loc1, _ := repo.GetByID(ctx, loc.GetID())
	loc2, _ := repo.GetByID(ctx, loc.GetID())

	// Update first copy
	loc1.Name = "Update 1"
	require.NoError(t, repo.Update(ctx, loc1))

	// Update second copy should fail (version mismatch)
	loc2.Name = "Update 2"
	err := repo.Update(ctx, loc2)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, location.ErrVersionMismatch))
}

func TestLocationRepository_Delete(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create location
	loc, _ := location.NewLocation("WH01", "Main Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, loc))

	// Delete
	err := repo.Delete(ctx, loc.GetID())
	require.NoError(t, err)

	// Verify soft deleted (not found)
	_, err = repo.GetByID(ctx, loc.GetID())
	assert.Error(t, err)
}

func TestLocationRepository_GetByCode(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create location
	loc, _ := location.NewLocation("WH01", "Main Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, loc))

	// Get by code
	retrieved, err := repo.GetByCode(ctx, "WH01")
	require.NoError(t, err)
	assert.Equal(t, loc.GetID(), retrieved.GetID())
	assert.Equal(t, "WH01", retrieved.Code)

	// Not found
	_, err = repo.GetByCode(ctx, "NONEXISTENT")
	assert.Error(t, err)
}

func TestLocationRepository_List(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create multiple locations
	loc1, _ := location.NewLocation("WH01", "Warehouse 1", location.LocationTypeWarehouse)
	loc2, _ := location.NewLocation("WH02", "Warehouse 2", location.LocationTypeWarehouse)
	loc3, _ := location.NewLocation("WH03", "Warehouse 3", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, loc1))
	require.NoError(t, repo.Create(ctx, loc2))
	require.NoError(t, repo.Create(ctx, loc3))

	// List all
	locations, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, locations, 3)

	// List with pagination
	locations, err = repo.List(ctx, 2, 0)
	require.NoError(t, err)
	assert.Len(t, locations, 2)

	locations, err = repo.List(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, locations, 1)
}

func TestLocationRepository_Count(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Initially zero
	count, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create locations
	loc1, _ := location.NewLocation("WH01", "Warehouse 1", location.LocationTypeWarehouse)
	loc2, _ := location.NewLocation("WH02", "Warehouse 2", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, loc1))
	require.NoError(t, repo.Create(ctx, loc2))

	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Delete one
	require.NoError(t, repo.Delete(ctx, loc1.GetID()))
	count, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count) // Soft deleted not counted
}

func TestLocationRepository_ListByType(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create locations of different types
	wh, _ := location.NewLocation("WH01", "Warehouse", location.LocationTypeWarehouse)
	zone1, _ := location.NewLocation("A", "Zone A", location.LocationTypeZone)
	zone2, _ := location.NewLocation("B", "Zone B", location.LocationTypeZone)
	aisle, _ := location.NewLocation("01", "Aisle 01", location.LocationTypeAisle)

	require.NoError(t, repo.Create(ctx, wh))
	require.NoError(t, repo.Create(ctx, zone1))
	require.NoError(t, repo.Create(ctx, zone2))
	require.NoError(t, repo.Create(ctx, aisle))

	// List zones only
	zones, err := repo.ListByType(ctx, location.LocationTypeZone, 10, 0)
	require.NoError(t, err)
	assert.Len(t, zones, 2)
	for _, z := range zones {
		assert.Equal(t, location.LocationTypeZone, z.Type)
	}

	// Count warehouses
	count, err := repo.CountByType(ctx, location.LocationTypeWarehouse)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLocationRepository_ListByParent(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create hierarchy: WH01 -> A, B
	wh, _ := location.NewLocation("WH01", "Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, wh))

	zoneA, _ := location.NewLocation("A", "Zone A", location.LocationTypeZone)
	require.NoError(t, zoneA.SetParent(wh.GetID(), wh.Path, wh.Level))
	require.NoError(t, repo.Create(ctx, zoneA))

	zoneB, _ := location.NewLocation("B", "Zone B", location.LocationTypeZone)
	require.NoError(t, zoneB.SetParent(wh.GetID(), wh.Path, wh.Level))
	require.NoError(t, repo.Create(ctx, zoneB))

	// Get direct children of warehouse
	children, err := repo.ListByParent(ctx, wh.GetID())
	require.NoError(t, err)
	assert.Len(t, children, 2)
}

func TestLocationRepository_ListChildren(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create hierarchy: WH01 -> A -> 01 -> 02
	wh, _ := location.NewLocation("WH01", "Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, wh))

	zoneA, _ := location.NewLocation("A", "Zone A", location.LocationTypeZone)
	require.NoError(t, zoneA.SetParent(wh.GetID(), wh.Path, wh.Level))
	require.NoError(t, repo.Create(ctx, zoneA))

	aisle, _ := location.NewLocation("01", "Aisle 01", location.LocationTypeAisle)
	require.NoError(t, aisle.SetParent(zoneA.GetID(), zoneA.Path, zoneA.Level))
	require.NoError(t, repo.Create(ctx, aisle))

	rack, _ := location.NewLocation("02", "Rack 02", location.LocationTypeRack)
	require.NoError(t, rack.SetParent(aisle.GetID(), aisle.Path, aisle.Level))
	require.NoError(t, repo.Create(ctx, rack))

	// Get ALL descendants of warehouse (should get 3: zone, aisle, rack)
	descendants, err := repo.ListChildren(ctx, wh.GetID())
	require.NoError(t, err)
	assert.Len(t, descendants, 3)

	// Get descendants of zone (should get 2: aisle, rack)
	descendants, err = repo.ListChildren(ctx, zoneA.GetID())
	require.NoError(t, err)
	assert.Len(t, descendants, 2)
}

func TestLocationRepository_ListByStatus(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create locations with different statuses
	active1, _ := location.NewLocation("L1", "Location 1", location.LocationTypeBin)
	active2, _ := location.NewLocation("L2", "Location 2", location.LocationTypeBin)
	inactive, _ := location.NewLocation("L3", "Location 3", location.LocationTypeBin)

	require.NoError(t, repo.Create(ctx, active1))
	require.NoError(t, repo.Create(ctx, active2))
	require.NoError(t, repo.Create(ctx, inactive))

	// Deactivate one
	require.NoError(t, inactive.Deactivate())
	require.NoError(t, repo.Update(ctx, inactive))

	// List active
	activeLocations, err := repo.ListByStatus(ctx, location.LocationStatusActive, 10, 0)
	require.NoError(t, err)
	assert.Len(t, activeLocations, 2)

	// Count inactive
	count, err := repo.CountByStatus(ctx, location.LocationStatusInactive)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLocationRepository_ListAvailable(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create locations with different availability
	available, _ := location.NewLocation("L1", "Available", location.LocationTypeBin)
	require.NoError(t, available.SetCapacity(100, true))
	require.NoError(t, repo.Create(ctx, available))

	full, _ := location.NewLocation("L2", "Full", location.LocationTypeBin)
	require.NoError(t, full.SetCapacity(100, true))
	require.NoError(t, full.AddOccupancy(100)) // Fill to capacity
	require.NoError(t, repo.Create(ctx, full))

	inactive, _ := location.NewLocation("L3", "Inactive", location.LocationTypeBin)
	require.NoError(t, inactive.Deactivate())
	require.NoError(t, repo.Create(ctx, inactive))

	notPutawayable, _ := location.NewLocation("L4", "Not Putawayable", location.LocationTypeBin)
	notPutawayable.DisablePutaway()
	require.NoError(t, repo.Create(ctx, notPutawayable))

	// List available (only L1 should be available)
	availableLocations, err := repo.ListAvailable(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, availableLocations, 1)
	assert.Equal(t, "L1", availableLocations[0].Code)

	// Count available
	count, err := repo.CountAvailable(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestLocationRepository_HierarchyIntegrity(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create full hierarchy
	wh, _ := location.NewLocation("WH01", "Warehouse", location.LocationTypeWarehouse)
	require.NoError(t, repo.Create(ctx, wh))

	zone, _ := location.NewLocation("A", "Zone A", location.LocationTypeZone)
	require.NoError(t, zone.SetParent(wh.GetID(), wh.Path, wh.Level))
	require.NoError(t, repo.Create(ctx, zone))

	aisle, _ := location.NewLocation("01", "Aisle 01", location.LocationTypeAisle)
	require.NoError(t, aisle.SetParent(zone.GetID(), zone.Path, zone.Level))
	require.NoError(t, repo.Create(ctx, aisle))

	// Verify path integrity
	retrieved, err := repo.GetByID(ctx, aisle.GetID())
	require.NoError(t, err)
	assert.Equal(t, "/WH01/A/01", retrieved.Path)
	assert.Equal(t, 2, retrieved.Level)
	assert.NotNil(t, retrieved.ParentID)
	assert.Equal(t, zone.GetID(), *retrieved.ParentID)
}

func TestLocationRepository_CapacityPersistence(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create location with capacity
	loc, _ := location.NewLocation("BIN01", "Bin 01", location.LocationTypeBin)
	require.NoError(t, loc.SetCapacity(500, true))
	require.NoError(t, loc.AddOccupancy(150))
	require.NoError(t, repo.Create(ctx, loc))

	// Retrieve and verify
	retrieved, err := repo.GetByID(ctx, loc.GetID())
	require.NoError(t, err)
	assert.Equal(t, 500, retrieved.Capacity)
	assert.Equal(t, 150, retrieved.CurrentOccupancy)
	assert.True(t, retrieved.IsLimited)
	assert.Equal(t, 350, retrieved.GetAvailableCapacity())
}

func TestLocationRepository_DimensionsPersistence(t *testing.T) {
	db := integration.SetupTestDBWithCleanTables(t)
	repo := locationRepo.NewLocationRepository(db.DB)
	ctx := context.Background()

	// Create location with dimensions
	loc, _ := location.NewLocation("RACK01", "Rack 01", location.LocationTypeRack)
	require.NoError(t, loc.SetDimensions(2.5, 3.0, 1.2))
	require.NoError(t, repo.Create(ctx, loc))

	// Retrieve and verify
	retrieved, err := repo.GetByID(ctx, loc.GetID())
	require.NoError(t, err)
	assert.Equal(t, 2.5, retrieved.Width)
	assert.Equal(t, 3.0, retrieved.Height)
	assert.Equal(t, 1.2, retrieved.Depth)
	assert.Equal(t, 9.0, retrieved.GetVolume())
}
