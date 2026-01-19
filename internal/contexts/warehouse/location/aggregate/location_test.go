package aggregate

import (
	"testing"

	locationerrors "github.com/basilex/promenade/internal/contexts/warehouse/location"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocation(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		locName string
		locType LocationType
		wantErr error
	}{
		{
			name:    "valid warehouse location",
			code:    "WH01",
			locName: "Main Warehouse",
			locType: LocationTypeWarehouse,
			wantErr: nil,
		},
		{
			name:    "valid zone location",
			code:    "A",
			locName: "Zone A",
			locType: LocationTypeZone,
			wantErr: nil,
		},
		{
			name:    "code is normalized to uppercase",
			code:    "wh01",
			locName: "Warehouse",
			locType: LocationTypeWarehouse,
			wantErr: nil,
		},
		{
			name:    "empty code",
			code:    "",
			locName: "Test",
			locType: LocationTypeWarehouse,
			wantErr: locationerrors.ErrLocationCodeRequired,
		},
		{
			name:    "empty name",
			code:    "WH01",
			locName: "",
			locType: LocationTypeWarehouse,
			wantErr: locationerrors.ErrLocationNameRequired,
		},
		{
			name:    "invalid location type",
			code:    "WH01",
			locName: "Test",
			locType: "invalid",
			wantErr: locationerrors.ErrInvalidLocationType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := NewLocation(tt.code, tt.locName, tt.locType)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, location)
			assert.NotEqual(t, uuidv7.UUID{}, location.GetID())
			assert.Equal(t, tt.locType, location.Type)
			assert.Equal(t, LocationStatusActive, location.Status)
			assert.True(t, location.IsPickable)
			assert.True(t, location.IsPutawayable)
			assert.False(t, location.IsLimited)
			assert.Equal(t, 0, location.CurrentOccupancy)
			assert.Equal(t, 0, location.Level)
		})
	}
}

func TestLocation_SetParent(t *testing.T) {
	parent, _ := NewLocation("WH01", "Main Warehouse", LocationTypeWarehouse)
	child, _ := NewLocation("A", "Zone A", LocationTypeZone)

	err := child.SetParent(parent.GetID(), parent.Path, parent.Level)
	require.NoError(t, err)
	assert.Equal(t, parent.GetID(), *child.ParentID)
	assert.Equal(t, "/WH01/A", child.Path)
	assert.Equal(t, 1, child.Level)

	// Test cannot be own parent
	err = child.SetParent(child.GetID(), child.Path, child.Level)
	assert.Error(t, err)
	assert.ErrorIs(t, err, locationerrors.ErrCannotBeOwnParent)
}

func TestLocation_RemoveParent(t *testing.T) {
	parent, _ := NewLocation("WH01", "Main Warehouse", LocationTypeWarehouse)
	child, _ := NewLocation("A", "Zone A", LocationTypeZone)
	_ = child.SetParent(parent.GetID(), parent.Path, parent.Level)

	child.RemoveParent()

	assert.Nil(t, child.ParentID)
	assert.Equal(t, "/A", child.Path)
	assert.Equal(t, 0, child.Level)
}

func TestLocation_SetCapacity(t *testing.T) {
	tests := []struct {
		name       string
		capacity   int
		enforce    bool
		occupancy  int
		wantErr    error
		wantStatus LocationStatus
	}{
		{
			name:       "set capacity without enforcement",
			capacity:   100,
			enforce:    false,
			occupancy:  0,
			wantErr:    nil,
			wantStatus: LocationStatusActive,
		},
		{
			name:       "set capacity with enforcement",
			capacity:   100,
			enforce:    true,
			occupancy:  0,
			wantErr:    nil,
			wantStatus: LocationStatusActive,
		},
		{
			name:      "negative capacity",
			capacity:  -10,
			enforce:   true,
			occupancy: 0,
			wantErr:   locationerrors.ErrNegativeCapacity,
		},
		{
			name:      "capacity less than occupancy",
			capacity:  50,
			enforce:   true,
			occupancy: 75,
			wantErr:   locationerrors.ErrCapacityLessThanOccupancy,
		},
		{
			name:       "set capacity equal to occupancy marks as full",
			capacity:   100,
			enforce:    true,
			occupancy:  100,
			wantErr:    nil,
			wantStatus: LocationStatusFull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, _ := NewLocation("TEST", "Test", LocationTypeWarehouse)
			if tt.occupancy > 0 {
				loc.CurrentOccupancy = tt.occupancy
			}

			err := loc.SetCapacity(tt.capacity, tt.enforce)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.capacity, loc.Capacity)
			assert.Equal(t, tt.enforce, loc.IsLimited)
			if tt.wantStatus != "" {
				assert.Equal(t, tt.wantStatus, loc.Status)
			}
		})
	}
}

func TestLocation_SetDimensions(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)

	// Valid dimensions
	err := location.SetDimensions(10.5, 5.0, 3.5)
	require.NoError(t, err)
	assert.Equal(t, 10.5, location.Width)
	assert.Equal(t, 5.0, location.Height)
	assert.Equal(t, 3.5, location.Depth)

	// Negative dimensions
	err = location.SetDimensions(-1, 5, 3)
	assert.Error(t, err)
	assert.ErrorIs(t, err, locationerrors.ErrNegativeDimensions)
}

func TestLocation_AddOccupancy(t *testing.T) {
	tests := []struct {
		name          string
		capacity      int
		isLimited     bool
		current       int
		addAmount     int
		wantErr       error
		wantOccupancy int
		wantStatus    LocationStatus
	}{
		{
			name:          "add to unlimited location",
			capacity:      0,
			isLimited:     false,
			current:       10,
			addAmount:     5,
			wantErr:       nil,
			wantOccupancy: 15,
			wantStatus:    LocationStatusActive,
		},
		{
			name:          "add within capacity",
			capacity:      100,
			isLimited:     true,
			current:       50,
			addAmount:     30,
			wantErr:       nil,
			wantOccupancy: 80,
			wantStatus:    LocationStatusActive,
		},
		{
			name:      "add exceeds capacity",
			capacity:  100,
			isLimited: true,
			current:   80,
			addAmount: 30,
			wantErr:   locationerrors.ErrInsufficientCapacity,
		},
		{
			name:          "add fills to capacity",
			capacity:      100,
			isLimited:     true,
			current:       90,
			addAmount:     10,
			wantErr:       nil,
			wantOccupancy: 100,
			wantStatus:    LocationStatusFull,
		},
		{
			name:      "add negative amount",
			capacity:  100,
			isLimited: true,
			current:   50,
			addAmount: -10,
			wantErr:   locationerrors.ErrAmountMustBePositive,
		},
		{
			name:      "add zero amount",
			capacity:  100,
			isLimited: true,
			current:   50,
			addAmount: 0,
			wantErr:   locationerrors.ErrAmountMustBePositive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, _ := NewLocation("TEST", "Test", LocationTypeWarehouse)
			loc.Capacity = tt.capacity
			loc.IsLimited = tt.isLimited
			loc.CurrentOccupancy = tt.current

			err := loc.AddOccupancy(tt.addAmount)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOccupancy, loc.CurrentOccupancy)
			if tt.wantStatus != "" {
				assert.Equal(t, tt.wantStatus, loc.Status)
			}
		})
	}
}

func TestLocation_RemoveOccupancy(t *testing.T) {
	tests := []struct {
		name          string
		capacity      int
		isLimited     bool
		current       int
		removeAmount  int
		wantErr       error
		wantOccupancy int
		wantStatus    LocationStatus
	}{
		{
			name:          "remove from occupied location",
			capacity:      100,
			isLimited:     true,
			current:       50,
			removeAmount:  20,
			wantErr:       nil,
			wantOccupancy: 30,
			wantStatus:    LocationStatusActive,
		},
		{
			name:          "remove all occupancy",
			capacity:      100,
			isLimited:     true,
			current:       50,
			removeAmount:  50,
			wantErr:       nil,
			wantOccupancy: 0,
			wantStatus:    LocationStatusActive,
		},
		{
			name:         "remove more than occupied",
			capacity:     100,
			isLimited:    true,
			current:      30,
			removeAmount: 50,
			wantErr:      locationerrors.ErrCannotRemoveExcessOccupancy,
		},
		{
			name:         "remove negative amount",
			capacity:     100,
			isLimited:    true,
			current:      50,
			removeAmount: -10,
			wantErr:      locationerrors.ErrAmountMustBePositive,
		},
		{
			name:          "remove from full location unfills it",
			capacity:      100,
			isLimited:     true,
			current:       100,
			removeAmount:  10,
			wantErr:       nil,
			wantOccupancy: 90,
			wantStatus:    LocationStatusActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, _ := NewLocation("TEST", "Test", LocationTypeWarehouse)
			loc.Capacity = tt.capacity
			loc.IsLimited = tt.isLimited
			loc.CurrentOccupancy = tt.current
			if tt.isLimited && tt.current >= tt.capacity {
				loc.Status = LocationStatusFull
			}

			err := loc.RemoveOccupancy(tt.removeAmount)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOccupancy, loc.CurrentOccupancy)
			if tt.wantStatus != "" {
				assert.Equal(t, tt.wantStatus, loc.Status)
			}
		})
	}
}

func TestLocation_Activate(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)
	location.Status = LocationStatusInactive

	err := location.Activate()
	require.NoError(t, err)
	assert.Equal(t, LocationStatusActive, location.Status)

	// Already active
	err = location.Activate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, locationerrors.ErrLocationAlreadyActive)
}

func TestLocation_Deactivate(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)

	// Cannot deactivate with items
	location.CurrentOccupancy = 10
	err := location.Deactivate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, locationerrors.ErrCannotDeactivateWithItems)

	// Can deactivate when empty
	location.CurrentOccupancy = 0
	err = location.Deactivate()
	require.NoError(t, err)
	assert.Equal(t, LocationStatusInactive, location.Status)

	// Already inactive
	err = location.Deactivate()
	assert.Error(t, err)
	assert.ErrorIs(t, err, locationerrors.ErrLocationAlreadyInactive)
}

func TestLocation_SetMaintenance(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)

	err := location.SetMaintenance()
	require.NoError(t, err)
	assert.Equal(t, LocationStatusMaintenance, location.Status)

	// Already in maintenance
	err = location.SetMaintenance()
	assert.Error(t, err)
	assert.ErrorIs(t, err, locationerrors.ErrLocationAlreadyInMaintenance)
}

func TestLocation_PickingAndPutaway(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)

	// Initial state
	assert.True(t, location.IsPickable)
	assert.True(t, location.IsPutawayable)

	// Disable picking
	location.DisablePicking()
	assert.False(t, location.IsPickable)

	// Enable picking
	location.EnablePicking()
	assert.True(t, location.IsPickable)

	// Disable putaway
	location.DisablePutaway()
	assert.False(t, location.IsPutawayable)

	// Enable putaway
	location.EnablePutaway()
	assert.True(t, location.IsPutawayable)
}

func TestLocation_UpdateDescription(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)

	location.UpdateDescription("  Test Description  ")
	assert.Equal(t, "Test Description", location.Description)
}

func TestLocation_UpdateNotes(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)

	location.UpdateNotes("  Important notes  ")
	assert.Equal(t, "Important notes", location.Notes)
}

func TestLocation_GetAvailableCapacity(t *testing.T) {
	// Unlimited location
	loc1, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)
	assert.Equal(t, -1, loc1.GetAvailableCapacity())

	// Limited location
	loc2, _ := NewLocation("WH02", "Warehouse 2", LocationTypeWarehouse)
	_ = loc2.SetCapacity(100, true)
	loc2.CurrentOccupancy = 30
	assert.Equal(t, 70, loc2.GetAvailableCapacity())
}

func TestLocation_GetOccupancyPercentage(t *testing.T) {
	// Unlimited location
	loc1, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)
	assert.Equal(t, 0.0, loc1.GetOccupancyPercentage())

	// Limited location
	loc2, _ := NewLocation("WH02", "Warehouse 2", LocationTypeWarehouse)
	_ = loc2.SetCapacity(100, true)
	loc2.CurrentOccupancy = 30
	assert.Equal(t, 30.0, loc2.GetOccupancyPercentage())

	// Full location
	loc2.CurrentOccupancy = 100
	assert.Equal(t, 100.0, loc2.GetOccupancyPercentage())
}

func TestLocation_GetVolume(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)
	_ = location.SetDimensions(10, 5, 3)
	assert.Equal(t, 150.0, location.GetVolume())
}

func TestLocation_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Location)
		wantAvail bool
	}{
		{
			name:      "active and putawayable unlimited",
			setup:     func(l *Location) {},
			wantAvail: true,
		},
		{
			name: "active and putawayable with capacity",
			setup: func(l *Location) {
				_ = l.SetCapacity(100, true)
				l.CurrentOccupancy = 50
			},
			wantAvail: true,
		},
		{
			name: "inactive",
			setup: func(l *Location) {
				l.Status = LocationStatusInactive
			},
			wantAvail: false,
		},
		{
			name: "not putawayable",
			setup: func(l *Location) {
				l.DisablePutaway()
			},
			wantAvail: false,
		},
		{
			name: "at full capacity",
			setup: func(l *Location) {
				_ = l.SetCapacity(100, true)
				l.CurrentOccupancy = 100
			},
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, _ := NewLocation("TEST", "Test", LocationTypeWarehouse)
			tt.setup(loc)
			assert.Equal(t, tt.wantAvail, loc.IsAvailable())
		})
	}
}

func TestLocation_IsEmpty(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)
	assert.True(t, location.IsEmpty())

	location.CurrentOccupancy = 1
	assert.False(t, location.IsEmpty())
}

func TestLocation_IsFull(t *testing.T) {
	location, _ := NewLocation("WH01", "Warehouse", LocationTypeWarehouse)

	// Unlimited location is never full
	assert.False(t, location.IsFull())

	// Limited location at capacity is full
	_ = location.SetCapacity(100, true)
	location.CurrentOccupancy = 100
	assert.True(t, location.IsFull())

	// Below capacity is not full
	location.CurrentOccupancy = 99
	assert.False(t, location.IsFull())
}

func TestLocation_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Location)
		wantErr error
	}{
		{
			name:    "valid location",
			setup:   func(l *Location) {},
			wantErr: nil,
		},
		{
			name: "empty code",
			setup: func(l *Location) {
				l.Code = ""
			},
			wantErr: locationerrors.ErrLocationCodeRequired,
		},
		{
			name: "empty name",
			setup: func(l *Location) {
				l.Name = ""
			},
			wantErr: locationerrors.ErrLocationNameRequired,
		},
		{
			name: "invalid type",
			setup: func(l *Location) {
				l.Type = "invalid"
			},
			wantErr: locationerrors.ErrInvalidLocationType,
		},
		{
			name: "invalid status",
			setup: func(l *Location) {
				l.Status = "invalid"
			},
			wantErr: locationerrors.ErrInvalidLocationStatus,
		},
		{
			name: "negative occupancy",
			setup: func(l *Location) {
				l.CurrentOccupancy = -1
			},
			wantErr: locationerrors.ErrNegativeOccupancy,
		},
		{
			name: "negative capacity when limited",
			setup: func(l *Location) {
				l.IsLimited = true
				l.Capacity = -1
			},
			wantErr: locationerrors.ErrNegativeCapacity,
		},
		{
			name: "occupancy exceeds capacity",
			setup: func(l *Location) {
				l.IsLimited = true
				l.Capacity = 100
				l.CurrentOccupancy = 150
			},
			wantErr: locationerrors.ErrOccupancyExceedsCapacity,
		},
		{
			name: "negative dimensions",
			setup: func(l *Location) {
				l.Width = -1
			},
			wantErr: locationerrors.ErrNegativeDimensions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, _ := NewLocation("TEST", "Test", LocationTypeWarehouse)
			tt.setup(loc)
			err := loc.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
