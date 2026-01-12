package location

import (
	"context"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// IUseCase defines business operations for Location aggregate
type IUseCase interface {
	// Core CRUD operations
	CreateLocation(ctx context.Context, code, name string, locationType LocationType, description *string, parentID *uuidv7.UUID) (*Location, error)
	GetLocation(ctx context.Context, id uuidv7.UUID) (*Location, error)
	GetLocationByCode(ctx context.Context, code string) (*Location, error)
	UpdateLocation(ctx context.Context, id uuidv7.UUID, name, description *string) (*Location, error)
	DeleteLocation(ctx context.Context, id uuidv7.UUID) error

	// List operations
	ListLocations(ctx context.Context, page, pageSize int) ([]*Location, int, error)
	ListLocationsByType(ctx context.Context, locationType LocationType, page, pageSize int) ([]*Location, int, error)
	ListLocationsByParent(ctx context.Context, parentID uuidv7.UUID) ([]*Location, error)
	ListLocationsByStatus(ctx context.Context, status LocationStatus, page, pageSize int) ([]*Location, int, error)
	ListAvailableLocations(ctx context.Context, page, pageSize int) ([]*Location, int, error)

	// Hierarchy operations
	GetLocationChildren(ctx context.Context, id uuidv7.UUID, recursive bool) ([]*Location, error)
	GetLocationHierarchy(ctx context.Context, id uuidv7.UUID) ([]*Location, error)

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

// useCase implements IUseCase interface
type useCase struct {
	repo IRepository
}

// NewUseCase creates a new location use case
func NewUseCase(repo IRepository) IUseCase {
	return &useCase{
		repo: repo,
	}
}

// CreateLocation creates a new location with validation
func (uc *useCase) CreateLocation(ctx context.Context, code, name string, locationType LocationType, description *string, parentID *uuidv7.UUID) (*Location, error) {
	// Check if code already exists
	existing, err := uc.repo.GetByCode(ctx, code)
	if err == nil && existing != nil {
		return nil, ErrLocationCodeExists
	}

	// Create new location
	location, err := NewLocation(code, name, locationType)
	if err != nil {
		return nil, ErrLocationCreateFailed
	}

	// Set optional fields
	if description != nil {
		location.Description = *description
	}

	// Handle parent relationship
	if parentID != nil {
		parent, err := uc.repo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, ErrParentLocationNotFound
		}
		if parent.DeletedAt != nil {
			return nil, ErrParentLocationDeleted
		}

		if err := location.SetParent(parent.ID, parent.Path, parent.Level); err != nil {
			return nil, ErrLocationUpdateFailed
		}
	}

	// Save to repository
	if err := uc.repo.Create(ctx, location); err != nil {
		return nil, ErrLocationUpdateFailed
	}

	return location, nil
}

// GetLocation retrieves a location by ID
func (uc *useCase) GetLocation(ctx context.Context, id uuidv7.UUID) (*Location, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLocationNotFound
	}
	return location, nil
}

// GetLocationByCode retrieves a location by code
func (uc *useCase) GetLocationByCode(ctx context.Context, code string) (*Location, error) {
	location, err := uc.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, ErrLocationNotFound
	}
	return location, nil
}

// UpdateLocation updates location details
func (uc *useCase) UpdateLocation(ctx context.Context, id uuidv7.UUID, name, description *string) (*Location, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return nil, ErrLocationAlreadyDeleted
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
		return nil, ErrLocationUpdateFailed
	}

	return location, nil
}

// DeleteLocation soft deletes a location
func (uc *useCase) DeleteLocation(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return ErrLocationAlreadyDeleted
	}

	// Check for children
	children, err := uc.repo.ListByParent(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return ErrLocationHasChildren
	}

	// Check occupancy
	if location.CurrentOccupancy > 0 {
		return ErrLocationHasItems
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return ErrLocationDeleteFailed
	}

	return nil
}

// ListLocations returns paginated list of all locations
func (uc *useCase) ListLocations(ctx context.Context, page, pageSize int) ([]*Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, ErrLocationListFailed
	}

	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, ErrLocationListFailed
	}

	return locations, total, nil
}

// ListLocationsByType returns locations filtered by type
func (uc *useCase) ListLocationsByType(ctx context.Context, locationType LocationType, page, pageSize int) ([]*Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.ListByType(ctx, locationType, page, pageSize)
	if err != nil {
		return nil, 0, ErrLocationListFailed
	}

	// Count by type (need to implement in repository)
	total := len(locations) // Simplified - should be separate count method
	if len(locations) == pageSize {
		total = pageSize * page // Estimate
	}

	return locations, total, nil
}

// ListLocationsByParent returns direct children of a parent
func (uc *useCase) ListLocationsByParent(ctx context.Context, parentID uuidv7.UUID) ([]*Location, error) {
	// Validate parent exists
	parent, err := uc.repo.GetByID(ctx, parentID)
	if err != nil {
		return nil, ErrParentLocationNotFound
	}
	if parent.DeletedAt != nil {
		return nil, ErrParentLocationDeleted
	}

	locations, err := uc.repo.ListByParent(ctx, parentID)
	if err != nil {
		return nil, ErrLocationListFailed
	}

	return locations, nil
}

// ListLocationsByStatus returns locations filtered by status
func (uc *useCase) ListLocationsByStatus(ctx context.Context, status LocationStatus, page, pageSize int) ([]*Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.ListByStatus(ctx, status, page, pageSize)
	if err != nil {
		return nil, 0, ErrLocationListFailed
	}

	// Count by status
	total := len(locations)
	if len(locations) == pageSize {
		total = pageSize * page
	}

	return locations, total, nil
}

// ListAvailableLocations returns locations available for operations
func (uc *useCase) ListAvailableLocations(ctx context.Context, page, pageSize int) ([]*Location, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	locations, err := uc.repo.ListAvailable(ctx, page, pageSize)
	if err != nil {
		return nil, 0, ErrLocationListFailed
	}

	total := len(locations)
	if len(locations) == pageSize {
		total = pageSize * page
	}

	return locations, total, nil
}

// GetLocationChildren returns children (direct or recursive)
func (uc *useCase) GetLocationChildren(ctx context.Context, id uuidv7.UUID, recursive bool) ([]*Location, error) {
	// Validate parent exists
	parent, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLocationNotFound
	}
	if parent.DeletedAt != nil {
		return nil, ErrLocationDeleted
	}

	if recursive {
		return uc.repo.ListChildren(ctx, id)
	}

	return uc.repo.ListByParent(ctx, id)
}

// GetLocationHierarchy returns full hierarchy path from root to location
func (uc *useCase) GetLocationHierarchy(ctx context.Context, id uuidv7.UUID) ([]*Location, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrLocationNotFound
	}

	// Parse path and get all ancestors
	hierarchy := []*Location{location}
	currentID := location.ParentID

	for currentID != nil {
		parent, err := uc.repo.GetByID(ctx, *currentID)
		if err != nil {
			break
		}
		// Prepend parent
		hierarchy = append([]*Location{parent}, hierarchy...)
		currentID = parent.ParentID
	}

	return hierarchy, nil
}

// ActivateLocation activates a location
func (uc *useCase) ActivateLocation(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return ErrLocationAlreadyDeleted
	}

	if err := location.Activate(); err != nil {
		return ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return ErrLocationUpdateFailed
	}

	return nil
}

// DeactivateLocation deactivates a location
func (uc *useCase) DeactivateLocation(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return ErrLocationAlreadyDeleted
	}

	if err := location.Deactivate(); err != nil {
		return ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return ErrLocationUpdateFailed
	}

	return nil
}

// SetLocationMaintenance sets location to maintenance status
func (uc *useCase) SetLocationMaintenance(ctx context.Context, id uuidv7.UUID) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return ErrLocationAlreadyDeleted
	}

	location.Status = LocationStatusMaintenance
	location.Touch()

	if err := uc.repo.Update(ctx, location); err != nil {
		return ErrLocationUpdateFailed
	}

	return nil
}

// UpdateLocationCapacity updates capacity settings
func (uc *useCase) UpdateLocationCapacity(ctx context.Context, id uuidv7.UUID, capacity int, isLimited bool) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return ErrLocationAlreadyDeleted
	}

	if err := location.SetCapacity(capacity, isLimited); err != nil {
		return ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return ErrLocationUpdateFailed
	}

	return nil
}

// UpdateLocationDimensions updates physical dimensions
func (uc *useCase) UpdateLocationDimensions(ctx context.Context, id uuidv7.UUID, width, height, depth float64) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return ErrLocationAlreadyDeleted
	}

	if err := location.SetDimensions(width, height, depth); err != nil {
		return ErrLocationUpdateFailed
	}

	if err := uc.repo.Update(ctx, location); err != nil {
		return ErrLocationUpdateFailed
	}

	return nil
}

// UpdateLocationFlags updates operational flags
func (uc *useCase) UpdateLocationFlags(ctx context.Context, id uuidv7.UUID, isPickable, isPutawayable bool) error {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return ErrLocationAlreadyDeleted
	}

	location.IsPickable = isPickable
	location.IsPutawayable = isPutawayable
	location.Touch()

	if err := uc.repo.Update(ctx, location); err != nil {
		return ErrLocationUpdateFailed
	}

	return nil
}

// CanReceiveItems checks if location can receive items
func (uc *useCase) CanReceiveItems(ctx context.Context, id uuidv7.UUID, quantity int) (bool, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return false, ErrLocationNotFound
	}

	if location.DeletedAt != nil {
		return false, ErrLocationDeleted
	}

	if !location.IsPutawayable {
		return false, nil
	}

	if location.Status != LocationStatusActive {
		return false, nil
	}

	if location.IsLimited {
		available := location.Capacity - location.CurrentOccupancy
		return available >= quantity, nil
	}

	return true, nil
}

// GetAvailableCapacity returns available capacity
func (uc *useCase) GetAvailableCapacity(ctx context.Context, id uuidv7.UUID) (int, error) {
	location, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return 0, ErrLocationNotFound
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
