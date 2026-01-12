package location

import "errors"

// Repository Errors - Data access failures
var (
	// ErrLocationNotFound is returned when location doesn't exist
	ErrLocationNotFound = errors.New("location not found")

	// ErrVersionMismatch is returned when optimistic locking fails (concurrent update)
	ErrVersionMismatch = errors.New("version mismatch during concurrent update")
)

// Business Logic Errors - Domain rule violations
var (
	// ErrLocationCodeExists is returned when code already exists
	ErrLocationCodeExists = errors.New("location code already exists")

	// ErrLocationAlreadyDeleted is returned when operating on deleted location
	ErrLocationAlreadyDeleted = errors.New("location already deleted")

	// ErrParentLocationDeleted is returned when parent is deleted
	ErrParentLocationDeleted = errors.New("parent location is deleted")

	// ErrParentLocationNotFound is returned when parent doesn't exist
	ErrParentLocationNotFound = errors.New("parent location not found")

	// ErrLocationHasChildren is returned when deleting location with children
	ErrLocationHasChildren = errors.New("cannot delete location with children")

	// ErrLocationHasItems is returned when deleting location with items
	ErrLocationHasItems = errors.New("cannot delete location with items")

	// ErrLocationDeleted is returned when operating on deleted location
	ErrLocationDeleted = errors.New("location is deleted")

	// ErrInvalidPagination is returned when page/pageSize invalid
	ErrInvalidPagination = errors.New("invalid pagination parameters")

	// Entity validation errors (used in tests)
	// ErrLocationCodeRequired is returned when location code is empty
	ErrLocationCodeRequired = errors.New("location code is required")

	// ErrLocationNameRequired is returned when location name is empty
	ErrLocationNameRequired = errors.New("location name is required")

	// ErrInvalidLocationType is returned when location type is invalid
	ErrInvalidLocationType = errors.New("invalid location type")

	// ErrInvalidLocationStatus is returned when location status is invalid
	ErrInvalidLocationStatus = errors.New("invalid location status")

	// ErrCannotBeOwnParent is returned when trying to set location as its own parent
	ErrCannotBeOwnParent = errors.New("location cannot be its own parent")

	// ErrNegativeCapacity is returned when capacity is negative
	ErrNegativeCapacity = errors.New("capacity cannot be negative")

	// ErrNegativeOccupancy is returned when occupancy is negative
	ErrNegativeOccupancy = errors.New("occupancy cannot be negative")

	// ErrCapacityLessThanOccupancy is returned when capacity < occupancy
	ErrCapacityLessThanOccupancy = errors.New("capacity cannot be less than current occupancy")

	// ErrNegativeDimensions is returned when dimensions are negative
	ErrNegativeDimensions = errors.New("dimensions cannot be negative")

	// ErrInsufficientCapacity is returned when adding exceeds capacity
	ErrInsufficientCapacity = errors.New("insufficient capacity")

	// ErrAmountMustBePositive is returned when amount is zero or negative
	ErrAmountMustBePositive = errors.New("amount must be positive")

	// ErrCannotRemoveExcessOccupancy is returned when removing more than current
	ErrCannotRemoveExcessOccupancy = errors.New("cannot remove more than current occupancy")

	// ErrLocationAlreadyActive is returned when activating active location
	ErrLocationAlreadyActive = errors.New("location is already active")

	// ErrCannotDeactivateWithItems is returned when deactivating with items
	ErrCannotDeactivateWithItems = errors.New("cannot deactivate location with items")

	// ErrLocationAlreadyInactive is returned when deactivating inactive location
	ErrLocationAlreadyInactive = errors.New("location is already inactive")

	// ErrLocationAlreadyInMaintenance is returned when already in maintenance
	ErrLocationAlreadyInMaintenance = errors.New("location is already in maintenance")

	// ErrOccupancyExceedsCapacity is returned when occupancy > capacity (validation)
	ErrOccupancyExceedsCapacity = errors.New("occupancy exceeds capacity")

	// ErrLocationCreateFailed is returned when creation fails
	ErrLocationCreateFailed = errors.New("failed to create location")

	// ErrLocationUpdateFailed is returned when update fails
	ErrLocationUpdateFailed = errors.New("failed to update location")

	// ErrLocationDeleteFailed is returned when deletion fails
	ErrLocationDeleteFailed = errors.New("failed to delete location")

	// ErrLocationListFailed is returned when listing fails
	ErrLocationListFailed = errors.New("failed to list locations")
)
