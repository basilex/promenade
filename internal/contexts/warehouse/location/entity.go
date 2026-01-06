package location

import (
	"fmt"
	"strings"

	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// LocationType represents the type of warehouse location
type LocationType string

const (
	LocationTypeWarehouse LocationType = "warehouse" // Top-level warehouse
	LocationTypeZone      LocationType = "zone"      // Zone within warehouse
	LocationTypeAisle     LocationType = "aisle"     // Aisle within zone
	LocationTypeRack      LocationType = "rack"      // Rack within aisle
	LocationTypeShelf     LocationType = "shelf"     // Shelf within rack
	LocationTypeBin       LocationType = "bin"       // Bin (smallest unit)
)

// LocationStatus represents the operational status of a location
type LocationStatus string

const (
	LocationStatusActive      LocationStatus = "active"      // Available for use
	LocationStatusInactive    LocationStatus = "inactive"    // Temporarily unavailable
	LocationStatusMaintenance LocationStatus = "maintenance" // Under maintenance
	LocationStatusFull        LocationStatus = "full"        // At capacity
)

// Location represents a physical storage location in a warehouse
type Location struct {
	aggregate.BaseAggregate

	// Identification
	Code        string       // Unique location code (e.g., "WH01-A-01-02")
	Name        string       // Human-readable name
	Type        LocationType // Type of location
	Description string       // Optional description

	// Hierarchy
	ParentID *uuidv7.UUID // Parent location ID (nil for top-level)
	Path     string       // Full hierarchical path (e.g., "/WH01/A/01/02")
	Level    int          // Hierarchy level (0 = warehouse, 1 = zone, etc.)

	// Capacity
	Capacity         int  // Maximum capacity (units/pallets/boxes)
	CurrentOccupancy int  // Current occupancy
	IsLimited        bool // Whether capacity is enforced

	// Physical Properties
	Width  float64 // Width in meters
	Height float64 // Height in meters
	Depth  float64 // Depth in meters

	// Status
	Status        LocationStatus // Current status
	IsPickable    bool           // Can pick items from this location
	IsPutawayable bool           // Can store items in this location

	// Metadata
	Notes string // Additional notes
}

// NewLocation creates a new location
func NewLocation(code, name string, locType LocationType) (*Location, error) {
	if code == "" {
		return nil, fmt.Errorf("location code is required")
	}
	if name == "" {
		return nil, fmt.Errorf("location name is required")
	}
	if !isValidLocationType(locType) {
		return nil, fmt.Errorf("invalid location type: %s", locType)
	}

	code = strings.TrimSpace(strings.ToUpper(code))

	location := &Location{
		BaseAggregate:    aggregate.NewBaseAggregate(),
		Code:             code,
		Name:             strings.TrimSpace(name),
		Type:             locType,
		Path:             "/" + code,
		Level:            0,
		Status:           LocationStatusActive,
		IsPickable:       true,
		IsPutawayable:    true,
		IsLimited:        false,
		CurrentOccupancy: 0,
	}

	return location, nil
}

// SetParent sets the parent location and updates hierarchy
func (l *Location) SetParent(parentID uuidv7.UUID, parentPath string, parentLevel int) error {
	if parentID == l.GetID() {
		return fmt.Errorf("location cannot be its own parent")
	}

	l.ParentID = &parentID
	l.Path = parentPath + "/" + l.Code
	l.Level = parentLevel + 1
	l.Touch()

	return nil
}

// RemoveParent removes parent relationship (makes location top-level)
func (l *Location) RemoveParent() {
	l.ParentID = nil
	l.Path = "/" + l.Code
	l.Level = 0
	l.Touch()
}

// SetCapacity sets the location capacity
func (l *Location) SetCapacity(capacity int, enforce bool) error {
	if capacity < 0 {
		return fmt.Errorf("capacity cannot be negative")
	}
	if capacity < l.CurrentOccupancy {
		return fmt.Errorf("capacity (%d) cannot be less than current occupancy (%d)", capacity, l.CurrentOccupancy)
	}

	l.Capacity = capacity
	l.IsLimited = enforce
	l.Touch()

	// Update status if now full
	if enforce && l.CurrentOccupancy >= l.Capacity && l.Status == LocationStatusActive {
		l.Status = LocationStatusFull
	}

	return nil
}

// SetDimensions sets the physical dimensions
func (l *Location) SetDimensions(width, height, depth float64) error {
	if width < 0 || height < 0 || depth < 0 {
		return fmt.Errorf("dimensions cannot be negative")
	}

	l.Width = width
	l.Height = height
	l.Depth = depth
	l.Touch()

	return nil
}

// AddOccupancy increases current occupancy
func (l *Location) AddOccupancy(amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	newOccupancy := l.CurrentOccupancy + amount

	if l.IsLimited && newOccupancy > l.Capacity {
		return fmt.Errorf("insufficient capacity: available %d, requested %d",
			l.Capacity-l.CurrentOccupancy, amount)
	}

	l.CurrentOccupancy = newOccupancy
	l.Touch()

	// Update status if now full
	if l.IsLimited && l.CurrentOccupancy >= l.Capacity && l.Status == LocationStatusActive {
		l.Status = LocationStatusFull
	}

	return nil
}

// RemoveOccupancy decreases current occupancy
func (l *Location) RemoveOccupancy(amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if amount > l.CurrentOccupancy {
		return fmt.Errorf("cannot remove %d units, only %d occupied", amount, l.CurrentOccupancy)
	}

	l.CurrentOccupancy -= amount
	l.Touch()

	// Update status if no longer full
	if l.Status == LocationStatusFull && l.CurrentOccupancy < l.Capacity {
		l.Status = LocationStatusActive
	}

	return nil
}

// Activate activates the location
func (l *Location) Activate() error {
	if l.Status == LocationStatusActive {
		return fmt.Errorf("location already active")
	}

	l.Status = LocationStatusActive
	l.Touch()

	return nil
}

// Deactivate deactivates the location
func (l *Location) Deactivate() error {
	if l.Status == LocationStatusInactive {
		return fmt.Errorf("location already inactive")
	}
	if l.CurrentOccupancy > 0 {
		return fmt.Errorf("cannot deactivate location with %d items", l.CurrentOccupancy)
	}

	l.Status = LocationStatusInactive
	l.Touch()

	return nil
}

// SetMaintenance marks location as under maintenance
func (l *Location) SetMaintenance() error {
	if l.Status == LocationStatusMaintenance {
		return fmt.Errorf("location already in maintenance")
	}

	l.Status = LocationStatusMaintenance
	l.Touch()

	return nil
}

// EnablePicking enables picking from this location
func (l *Location) EnablePicking() {
	l.IsPickable = true
	l.Touch()
}

// DisablePicking disables picking from this location
func (l *Location) DisablePicking() {
	l.IsPickable = false
	l.Touch()
}

// EnablePutaway enables storing items in this location
func (l *Location) EnablePutaway() {
	l.IsPutawayable = true
	l.Touch()
}

// DisablePutaway disables storing items in this location
func (l *Location) DisablePutaway() {
	l.IsPutawayable = false
	l.Touch()
}

// UpdateDescription updates the location description
func (l *Location) UpdateDescription(description string) {
	l.Description = strings.TrimSpace(description)
	l.Touch()
}

// UpdateNotes updates the location notes
func (l *Location) UpdateNotes(notes string) {
	l.Notes = strings.TrimSpace(notes)
	l.Touch()
}

// GetAvailableCapacity returns available capacity
func (l *Location) GetAvailableCapacity() int {
	if !l.IsLimited {
		return -1 // Unlimited
	}
	return l.Capacity - l.CurrentOccupancy
}

// GetOccupancyPercentage returns occupancy as percentage
func (l *Location) GetOccupancyPercentage() float64 {
	if !l.IsLimited || l.Capacity == 0 {
		return 0.0
	}
	return float64(l.CurrentOccupancy) / float64(l.Capacity) * 100.0
}

// GetVolume returns the volume in cubic meters
func (l *Location) GetVolume() float64 {
	return l.Width * l.Height * l.Depth
}

// IsAvailable checks if location can accept items
func (l *Location) IsAvailable() bool {
	return l.Status == LocationStatusActive &&
		l.IsPutawayable &&
		(!l.IsLimited || l.CurrentOccupancy < l.Capacity)
}

// IsEmpty checks if location has no items
func (l *Location) IsEmpty() bool {
	return l.CurrentOccupancy == 0
}

// IsFull checks if location is at capacity
func (l *Location) IsFull() bool {
	return l.IsLimited && l.CurrentOccupancy >= l.Capacity
}

// Validate validates the location entity
func (l *Location) Validate() error {
	if l.Code == "" {
		return fmt.Errorf("location code is required")
	}
	if l.Name == "" {
		return fmt.Errorf("location name is required")
	}
	if !isValidLocationType(l.Type) {
		return fmt.Errorf("invalid location type: %s", l.Type)
	}
	if !isValidLocationStatus(l.Status) {
		return fmt.Errorf("invalid location status: %s", l.Status)
	}
	if l.CurrentOccupancy < 0 {
		return fmt.Errorf("current occupancy cannot be negative")
	}
	if l.IsLimited && l.Capacity < 0 {
		return fmt.Errorf("capacity cannot be negative when limited")
	}
	if l.IsLimited && l.CurrentOccupancy > l.Capacity {
		return fmt.Errorf("current occupancy (%d) exceeds capacity (%d)", l.CurrentOccupancy, l.Capacity)
	}
	if l.Width < 0 || l.Height < 0 || l.Depth < 0 {
		return fmt.Errorf("dimensions cannot be negative")
	}

	return nil
}

// Helper functions

func isValidLocationType(t LocationType) bool {
	switch t {
	case LocationTypeWarehouse, LocationTypeZone, LocationTypeAisle,
		LocationTypeRack, LocationTypeShelf, LocationTypeBin:
		return true
	default:
		return false
	}
}

func isValidLocationStatus(s LocationStatus) bool {
	switch s {
	case LocationStatusActive, LocationStatusInactive,
		LocationStatusMaintenance, LocationStatusFull:
		return true
	default:
		return false
	}
}
