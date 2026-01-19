package usecase

import (
	"context"

	locationerrors "github.com/basilex/promenade/internal/contexts/warehouse/location"
	"github.com/basilex/promenade/internal/contexts/warehouse/location/aggregate"
	"github.com/basilex/promenade/internal/contexts/warehouse/location/repository"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ILocationUseCase defines business operations for aggregate.Location aggregate
type ILocationUseCase interface {
	// Core CRUD operations
	CreateLocation(ctx context.Context, code, name string, locationType aggregate.LocationType, description *string, parentID *uuidv7.UUID) (*aggregate.Location, error)
	GetLocation(ctx context.Context, id uuidv7.UUID) (*aggregate.Location, error)
	GetLocationByCode(ctx context.Context, code string) (*aggregate.Location, error)
	UpdateLocation(ctx context.Context, id uuidv7.UUID, name, description *string) (*aggregate.Location, error)
	DeleteLocation(ctx context.Context, id uuidv7.UUID) error

	// List operations
	ListLocations(ctx context.Context, page, pageSize int) ([]*aggregate.Location, int, error)
	ListLocationsByType(ctx context.Context, locationType aggregate.LocationType, page, pageSize int) ([]*aggregate.Location, int, error)
	ListLocationsByParent(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Location, error)
	ListLocationsByStatus(ctx context.Context, status aggregate.LocationStatus, page, pageSize int) ([]*aggregate.Location, int, error)
	ListAvailableLocations(ctx context.Context, page, pageSize int) ([]*aggregate.Location, int, error)

	// Hierarchy operations
	GetLocationChildren(ctx context.Context, id uuidv7.UUID, recursive bool) ([]*aggregate.Location, error)
	GetLocationHierarchy(ctx context.Context, id uuidv7.UUID) ([]*aggregate.Location, error)

	// Status management
	ActivateLocation(ctx context.Context, id uuidv7.UUID) error
	DeactivateLocation(ctx context.Context, id uuidv7.UUID) error
	SetLocationMaintenance(ctx context.Context, id uuidv7.UUID) error

	// Capacity management
	UpdateLocationCapacity(ctx context.Context, id uuidv7.UUID, capacity int, isLimited bool) error
	UpdateLocationDimensions(ctx context.Context, id uuidv7.UUID, width, height, depth float64) error
	UpdateLocationFlags(ctx context.Context, id uuidv7.UUID, isPickable, isPutawayable bool) error

	// Business validation
	CanReceiveItems(ctx context.Context, id uuidv7.UUID, quantity int) (bool, error)
	GetAvailableCapacity(ctx context.Context, id uuidv7.UUID) (int, error)
}

// LocationUseCase implements ILocationUseCase interface
type LocationUseCase struct {
	repo repository.ILocationRepository
}

// NewLocationUseCase creates a new location use case
func NewLocationUseCase(repo repository.ILocationRepository) ILocationUseCase {
	return &LocationUseCase{
		repo: repo,
	}
}

// CreateLocation creates a new location with validation
func (uc *LocationUseCase) CreateLocation(ctx context.Context, code, name string, locationType aggregate.LocationType, description *string, parentID *uuidv7.UUID) (*aggregate.Location, error) {
	// Check if code already exists
	existing, err := uc.repo.GetByCode(ctx, code)
	if err == nil && existing != nil {
		return nil, locationerrors.ErrLocationCodeExists
	}

	// Create new location
	location, err := aggregate.NewLocation(code, name, locationType)
	if err != nil {
		return nil, locationerrors.ErrLocationCreateFailed
	}

	// Set optional fields
	if description != nil {
		location.Description = *description
	}

	// Handle parent relationship
	if parentID != nil {
		parent, err := uc.repo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, locationerrors.ErrParentLocationNotFound
		}
		if parent.DeletedAt != nil {
			return nil, locationerrors.ErrParentLocationDeleted
		}

		if err := location.SetParent(parent.ID, parent.Path, parent.Level); err != nil {
			return nil, locationerrors.ErrLocationUpdateFailed
		}
	}

	// Save to repository
	if err := uc.repo.Create(ctx, location); err != nil {
		return nil, locationerrors.ErrLocationUpdateFailed
	}

	return location, nil
}

// GetLocation retrieves a location by ID
func (uc *LocationUseCase) GetLocation(ctx context.Context, id uuidv7.UUID) (*aggregate.Location, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, locationerrors.ErrLocationNotFound
	}
	return location, nil
}

// GetLocationByCode retrieves a location by code
func (uc *LocationUseCase) GetLocationByCode(ctx context.Context, code string) (*aggregate.Location, error) {
	location, err := uc.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, locationerrors.ErrLocationNotFound
	}
	return location, nil
}

// UpdateLocation updates location details
func (uc *LocationUseCase) UpdateLocation(ctx context.Context, id uuidv7.UUID, name, description *string) (*aggregate.Location, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return nil, locationerrors.ErrLocationAlreadyDeleted
	}

	// Update fields
	if name != nil && *name != "" {
		location.Name = *name
	}
	if description != nil {
		location.Description = *description
	}

	location.Touch()

	if err := uc.repo.Update(ctx, location); err != nil {
		return nil, locationerrors.ErrLocationUpdateFailed
	}

	return location, nil
}

// DeleteLocation soft deletes a location
func (uc *LocationUseCase) DeleteLocation(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return locationerrors.ErrLocationAlreadyDeleted
	}

	// Check for children
	children, err := uc.repo.ListByParent(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return locationerrors.ErrLocationHasChildren
	}

	// Check occupancy
	if location.CurrentOccupancy > 0 {
		return locationerrors.ErrLocationHasItems
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return locationerrors.ErrLocationDeleteFailed
	}

	return nil
}

// ListLocations returns paginated list of all locations
func (uc *LocationUseCase) ListLocations(ctx context.Context, page, pageSize int) ([]*aggregate.Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, locationerrors.ErrLocationListFailed
	}

	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, locationerrors.ErrLocationListFailed
	}

	return locations, total, nil
}

// ListLocationsByType returns locations filtered by type
func (uc *LocationUseCase) ListLocationsByType(ctx context.Context, locationType aggregate.LocationType, page, pageSize int) ([]*aggregate.Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.ListByType(ctx, locationType, page, pageSize)
	if err != nil {
		return nil, 0, locationerrors.ErrLocationListFailed
	}

	// Count by type (need to implement in repository)
	total := len(locations) // Simplified - should be separate count method
	if len(locations) == pageSize {
		total = pageSize * page // Estimate
	}

	return locations, total, nil
}

// ListLocationsByParent returns direct children of a parent
func (uc *LocationUseCase) ListLocationsByParent(ctx context.Context, parentID uuidv7.UUID) ([]*aggregate.Location, error) {
	// Validate parent exists
	parent, err := uc.repo.GetByID(ctx, parentID)
	if err != nil {
		return nil, locationerrors.ErrParentLocationNotFound
	}
	if parent.DeletedAt != nil {
		return nil, locationerrors.ErrParentLocationDeleted
	}

	locations, err := uc.repo.ListByParent(ctx, parentID)
	if err != nil {
		return nil, locationerrors.ErrLocationListFailed
	}

	return locations, nil
}

// ListLocationsByStatus returns locations filtered by status
func (uc *LocationUseCase) ListLocationsByStatus(ctx context.Context, status aggregate.LocationStatus, page, pageSize int) ([]*aggregate.Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, locationerrors.ErrLocationListFailed
	}

	// Count by status
	total := len(locations)
	if len(locations) == pageSize {
		total = pageSize * page
	}

	return locations, total, nil
}

// ListAvailableLocations returns locations available for operations
func (uc *LocationUseCase) ListAvailableLocations(ctx context.Context, page, pageSize int) ([]*aggregate.Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.ListAvailable(ctx, page, pageSize)
	if err != nil {
		return nil, 0, locationerrors.ErrLocationListFailed
	}

	total := len(locations)
	if len(locations) == pageSize {
		total = pageSize * page
	}

	return locations, total, nil
}

// GetLocationChildren returns children (direct or recursive)
func (uc *LocationUseCase) GetLocationChildren(ctx context.Context, id uuidv7.UUID, recursive bool) ([]*aggregate.Location, error) {
	// Validate parent exists
	parent, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, locationerrors.ErrLocationNotFound
	}
	if parent.DeletedAt != nil {
		return nil, locationerrors.ErrLocationDeleted
	}

	if recursive {
		return uc.repo.ListChildren(ctx, id)
	}

	return uc.repo.ListByParent(ctx, id)
}

// GetLocationHierarchy returns full hierarchy path from root to location
func (uc *LocationUseCase) GetLocationHierarchy(ctx context.Context, id uuidv7.UUID) ([]*aggregate.Location, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, locationerrors.ErrLocationNotFound
	}

	// Parse path and get all ancestors
	hierarchy := []*aggregate.Location{location}
	currentID := location.ParentID

	for currentID != nil {
		parent, err := uc.repo.GetByID(ctx, *currentID)
		if err != nil {
			break
		}
		// Prepend parent
		hierarchy = append([]*aggregate.Location{parent}, hierarchy...)
		currentID = parent.ParentID
	}

	return hierarchy, nil
}

// ActivateLocation activates a location
func (uc *LocationUseCase) ActivateLocation(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return locationerrors.ErrLocationAlreadyDeleted
	}

	if err := location.Activate(); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	return nil
}

// DeactivateLocation deactivates a location
func (uc *LocationUseCase) DeactivateLocation(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return locationerrors.ErrLocationAlreadyDeleted
	}

	if err := location.Deactivate(); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	return nil
}

// SetLocationMaintenance sets location to maintenance status
func (uc *LocationUseCase) SetLocationMaintenance(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return locationerrors.ErrLocationAlreadyDeleted
	}

	location.Status = aggregate.LocationStatusMaintenance
	location.Touch()

	if err := uc.repo.Update(ctx, location); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	return nil
}

// UpdateLocationCapacity updates capacity settings
func (uc *LocationUseCase) UpdateLocationCapacity(ctx context.Context, id uuidv7.UUID, capacity int, isLimited bool) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return locationerrors.ErrLocationAlreadyDeleted
	}

	if err := location.SetCapacity(capacity, isLimited); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	return nil
}

// UpdateLocationDimensions updates physical dimensions
func (uc *LocationUseCase) UpdateLocationDimensions(ctx context.Context, id uuidv7.UUID, width, height, depth float64) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return locationerrors.ErrLocationAlreadyDeleted
	}

	if err := location.SetDimensions(width, height, depth); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	return nil
}

// UpdateLocationFlags updates operational flags
func (uc *LocationUseCase) UpdateLocationFlags(ctx context.Context, id uuidv7.UUID, isPickable, isPutawayable bool) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return locationerrors.ErrLocationAlreadyDeleted
	}

	location.IsPickable = isPickable
	location.IsPutawayable = isPutawayable
	location.Touch()

	if err := uc.repo.Update(ctx, location); err != nil {
		return locationerrors.ErrLocationUpdateFailed
	}

	return nil
}

// CanReceiveItems checks if location can receive items
func (uc *LocationUseCase) CanReceiveItems(ctx context.Context, id uuidv7.UUID, quantity int) (bool, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return false, locationerrors.ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return false, locationerrors.ErrLocationDeleted
	}

	if !location.IsPutawayable {
		return false, nil
	}

	if location.Status != aggregate.LocationStatusActive {
		return false, nil
	}

	if location.IsLimited {
		available := location.Capacity - location.CurrentOccupancy
		return available >= quantity, nil
	}

	return true, nil
}

// GetAvailableCapacity returns available capacity
func (uc *LocationUseCase) GetAvailableCapacity(ctx context.Context, id uuidv7.UUID) (int, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return 0, locationerrors.ErrLocationNotFound
	}

	if !location.IsLimited {
		return -1, nil // Unlimited
	}

	available := location.Capacity - location.CurrentOccupancy
	if available < 0 {
		available = 0
	}

	return available, nil
}
