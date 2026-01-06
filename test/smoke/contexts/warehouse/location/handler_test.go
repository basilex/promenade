package location_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/basilex/promenade/internal/contexts/warehouse/location"
	handler "github.com/basilex/promenade/internal/contexts/warehouse/location/adapter/http"
	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/test/smoke"
)

// MockLocationUseCase mocks IUseCase for smoke tests
type MockLocationUseCase struct {
	CreateLocationFunc           func(ctx context.Context, code, name string, locationType location.LocationType, description *string, parentID *uuidv7.UUID) (*location.Location, error)
	GetLocationFunc              func(ctx context.Context, id uuidv7.UUID) (*location.Location, error)
	GetLocationByCodeFunc        func(ctx context.Context, code string) (*location.Location, error)
	UpdateLocationFunc           func(ctx context.Context, id uuidv7.UUID, name, description *string) (*location.Location, error)
	DeleteLocationFunc           func(ctx context.Context, id uuidv7.UUID) error
	ListLocationsFunc            func(ctx context.Context, page, pageSize int) ([]*location.Location, int, error)
	ListLocationsByTypeFunc      func(ctx context.Context, locationType location.LocationType, page, pageSize int) ([]*location.Location, int, error)
	ListLocationsByParentFunc    func(ctx context.Context, parentID uuidv7.UUID) ([]*location.Location, error)
	ListLocationsByStatusFunc    func(ctx context.Context, status location.LocationStatus, page, pageSize int) ([]*location.Location, int, error)
	ListAvailableLocationsFunc   func(ctx context.Context, page, pageSize int) ([]*location.Location, int, error)
	GetLocationChildrenFunc      func(ctx context.Context, id uuidv7.UUID, recursive bool) ([]*location.Location, error)
	GetLocationHierarchyFunc     func(ctx context.Context, id uuidv7.UUID) ([]*location.Location, error)
	ActivateLocationFunc         func(ctx context.Context, id uuidv7.UUID) error
	DeactivateLocationFunc       func(ctx context.Context, id uuidv7.UUID) error
	SetLocationMaintenanceFunc   func(ctx context.Context, id uuidv7.UUID) error
	UpdateLocationCapacityFunc   func(ctx context.Context, id uuidv7.UUID, capacity int, isLimited bool) error
	UpdateLocationDimensionsFunc func(ctx context.Context, id uuidv7.UUID, width, height, depth float64) error
	UpdateLocationFlagsFunc      func(ctx context.Context, id uuidv7.UUID, isPickable, isPutawayable bool) error
	CanReceiveItemsFunc          func(ctx context.Context, id uuidv7.UUID, quantity int) (bool, error)
	GetAvailableCapacityFunc     func(ctx context.Context, id uuidv7.UUID) (int, error)
}

func (m *MockLocationUseCase) CreateLocation(ctx context.Context, code, name string, locationType location.LocationType, description *string, parentID *uuidv7.UUID) (*location.Location, error) {
	if m.CreateLocationFunc != nil {
		return m.CreateLocationFunc(ctx, code, name, locationType, description, parentID)
	}
	return nil, errors.New("CreateLocationFunc not implemented")
}

func (m *MockLocationUseCase) GetLocation(ctx context.Context, id uuidv7.UUID) (*location.Location, error) {
	if m.GetLocationFunc != nil {
		return m.GetLocationFunc(ctx, id)
	}
	return nil, errors.New("GetLocationFunc not implemented")
}

func (m *MockLocationUseCase) GetLocationByCode(ctx context.Context, code string) (*location.Location, error) {
	if m.GetLocationByCodeFunc != nil {
		return m.GetLocationByCodeFunc(ctx, code)
	}
	return nil, errors.New("GetLocationByCodeFunc not implemented")
}

func (m *MockLocationUseCase) UpdateLocation(ctx context.Context, id uuidv7.UUID, name, description *string) (*location.Location, error) {
	if m.UpdateLocationFunc != nil {
		return m.UpdateLocationFunc(ctx, id, name, description)
	}
	return nil, errors.New("UpdateLocationFunc not implemented")
}

func (m *MockLocationUseCase) DeleteLocation(ctx context.Context, id uuidv7.UUID) error {
	if m.DeleteLocationFunc != nil {
		return m.DeleteLocationFunc(ctx, id)
	}
	return errors.New("DeleteLocationFunc not implemented")
}

func (m *MockLocationUseCase) ListLocations(ctx context.Context, page, pageSize int) ([]*location.Location, int, error) {
	if m.ListLocationsFunc != nil {
		return m.ListLocationsFunc(ctx, page, pageSize)
	}
	return nil, 0, errors.New("ListLocationsFunc not implemented")
}

func (m *MockLocationUseCase) ListLocationsByType(ctx context.Context, locationType location.LocationType, page, pageSize int) ([]*location.Location, int, error) {
	if m.ListLocationsByTypeFunc != nil {
		return m.ListLocationsByTypeFunc(ctx, locationType, page, pageSize)
	}
	return nil, 0, errors.New("ListLocationsByTypeFunc not implemented")
}

func (m *MockLocationUseCase) ListLocationsByParent(ctx context.Context, parentID uuidv7.UUID) ([]*location.Location, error) {
	if m.ListLocationsByParentFunc != nil {
		return m.ListLocationsByParentFunc(ctx, parentID)
	}
	return nil, errors.New("ListLocationsByParentFunc not implemented")
}

func (m *MockLocationUseCase) ListLocationsByStatus(ctx context.Context, status location.LocationStatus, page, pageSize int) ([]*location.Location, int, error) {
	if m.ListLocationsByStatusFunc != nil {
		return m.ListLocationsByStatusFunc(ctx, status, page, pageSize)
	}
	return nil, 0, errors.New("ListLocationsByStatusFunc not implemented")
}

func (m *MockLocationUseCase) ListAvailableLocations(ctx context.Context, page, pageSize int) ([]*location.Location, int, error) {
	if m.ListAvailableLocationsFunc != nil {
		return m.ListAvailableLocationsFunc(ctx, page, pageSize)
	}
	return nil, 0, errors.New("ListAvailableLocationsFunc not implemented")
}

func (m *MockLocationUseCase) GetLocationChildren(ctx context.Context, id uuidv7.UUID, recursive bool) ([]*location.Location, error) {
	if m.GetLocationChildrenFunc != nil {
		return m.GetLocationChildrenFunc(ctx, id, recursive)
	}
	return nil, errors.New("GetLocationChildrenFunc not implemented")
}

func (m *MockLocationUseCase) GetLocationHierarchy(ctx context.Context, id uuidv7.UUID) ([]*location.Location, error) {
	if m.GetLocationHierarchyFunc != nil {
		return m.GetLocationHierarchyFunc(ctx, id)
	}
	return nil, errors.New("GetLocationHierarchyFunc not implemented")
}

func (m *MockLocationUseCase) ActivateLocation(ctx context.Context, id uuidv7.UUID) error {
	if m.ActivateLocationFunc != nil {
		return m.ActivateLocationFunc(ctx, id)
	}
	return errors.New("ActivateLocationFunc not implemented")
}

func (m *MockLocationUseCase) DeactivateLocation(ctx context.Context, id uuidv7.UUID) error {
	if m.DeactivateLocationFunc != nil {
		return m.DeactivateLocationFunc(ctx, id)
	}
	return errors.New("DeactivateLocationFunc not implemented")
}

func (m *MockLocationUseCase) SetLocationMaintenance(ctx context.Context, id uuidv7.UUID) error {
	if m.SetLocationMaintenanceFunc != nil {
		return m.SetLocationMaintenanceFunc(ctx, id)
	}
	return errors.New("SetLocationMaintenanceFunc not implemented")
}

func (m *MockLocationUseCase) UpdateLocationCapacity(ctx context.Context, id uuidv7.UUID, capacity int, isLimited bool) error {
	if m.UpdateLocationCapacityFunc != nil {
		return m.UpdateLocationCapacityFunc(ctx, id, capacity, isLimited)
	}
	return errors.New("UpdateLocationCapacityFunc not implemented")
}

func (m *MockLocationUseCase) UpdateLocationDimensions(ctx context.Context, id uuidv7.UUID, width, height, depth float64) error {
	if m.UpdateLocationDimensionsFunc != nil {
		return m.UpdateLocationDimensionsFunc(ctx, id, width, height, depth)
	}
	return errors.New("UpdateLocationDimensionsFunc not implemented")
}

func (m *MockLocationUseCase) UpdateLocationFlags(ctx context.Context, id uuidv7.UUID, isPickable, isPutawayable bool) error {
	if m.UpdateLocationFlagsFunc != nil {
		return m.UpdateLocationFlagsFunc(ctx, id, isPickable, isPutawayable)
	}
	return errors.New("UpdateLocationFlagsFunc not implemented")
}

func (m *MockLocationUseCase) CanReceiveItems(ctx context.Context, id uuidv7.UUID, quantity int) (bool, error) {
	if m.CanReceiveItemsFunc != nil {
		return m.CanReceiveItemsFunc(ctx, id, quantity)
	}
	return false, errors.New("CanReceiveItemsFunc not implemented")
}

func (m *MockLocationUseCase) GetAvailableCapacity(ctx context.Context, id uuidv7.UUID) (int, error) {
	if m.GetAvailableCapacityFunc != nil {
		return m.GetAvailableCapacityFunc(ctx, id)
	}
	return 0, errors.New("GetAvailableCapacityFunc not implemented")
}

// Helper: fake location
func fakeLocation() *location.Location {
	loc, _ := location.NewLocation("WH-MAIN", "Main Warehouse", location.LocationTypeWarehouse)
	return loc
}

// ============================================================================
// Smoke Tests
// ============================================================================

func TestLocationHandler_Create_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		CreateLocationFunc: func(ctx context.Context, code, name string, locationType location.LocationType, description *string, parentID *uuidv7.UUID) (*location.Location, error) {
			return fakeLocation(), nil
		},
		UpdateLocationDimensionsFunc: func(ctx context.Context, id uuidv7.UUID, width, height, depth float64) error {
			return nil
		},
		UpdateLocationCapacityFunc: func(ctx context.Context, id uuidv7.UUID, capacity int, isLimited bool) error {
			return nil
		},
		UpdateLocationFlagsFunc: func(ctx context.Context, id uuidv7.UUID, isPickable, isPutawayable bool) error {
			return nil
		},
		GetLocationFunc: func(ctx context.Context, id uuidv7.UUID) (*location.Location, error) {
			return fakeLocation(), nil
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.POST("/locations", h.Create)

	body := map[string]interface{}{
		"code": "WH-MAIN",
		"name": "Main Warehouse",
		"type": "warehouse",
	}

	w := smoke.MakeRequest(t, router, http.MethodPost, "/locations", body)
	smoke.AssertSuccessResponse(t, w, http.StatusCreated)
}

func TestLocationHandler_Create_ValidationError(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{}
	h := handler.NewLocationHandler(mockUC)
	router.POST("/locations", h.Create)

	body := map[string]interface{}{
		"name": "Main Warehouse",
		// Missing required "code" and "type"
	}

	w := smoke.MakeRequest(t, router, http.MethodPost, "/locations", body)
	smoke.AssertErrorResponse(t, w, http.StatusBadRequest, "BAD_REQUEST")
}

func TestLocationHandler_GetByID_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		GetLocationFunc: func(ctx context.Context, id uuidv7.UUID) (*location.Location, error) {
			return fakeLocation(), nil
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.GET("/locations/:id", h.GetByID)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/locations/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestLocationHandler_GetByID_NotFound(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		GetLocationFunc: func(ctx context.Context, id uuidv7.UUID) (*location.Location, error) {
			return nil, errors.New("location not found")
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.GET("/locations/:id", h.GetByID)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/locations/"+smoke.FakeUUID(), nil)
	smoke.AssertErrorResponse(t, w, http.StatusNotFound, "NOT_FOUND")
}

func TestLocationHandler_GetByCode_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		GetLocationByCodeFunc: func(ctx context.Context, code string) (*location.Location, error) {
			return fakeLocation(), nil
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.GET("/locations/code/:code", h.GetByCode)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/locations/code/WH-MAIN", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestLocationHandler_List_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		ListLocationsFunc: func(ctx context.Context, page, pageSize int) ([]*location.Location, int, error) {
			return []*location.Location{fakeLocation()}, 1, nil
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.GET("/locations", h.List)

	w := smoke.MakeRequest(t, router, http.MethodGet, "/locations?page=1&page_size=10", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestLocationHandler_Update_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		UpdateLocationFunc: func(ctx context.Context, id uuidv7.UUID, name, description *string) (*location.Location, error) {
			return fakeLocation(), nil
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.PUT("/locations/:id", h.Update)

	body := map[string]interface{}{
		"name": "Updated Warehouse",
	}

	w := smoke.MakeRequest(t, router, http.MethodPut, "/locations/"+smoke.FakeUUID(), body)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestLocationHandler_Delete_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		DeleteLocationFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.DELETE("/locations/:id", h.Delete)

	w := smoke.MakeRequest(t, router, http.MethodDelete, "/locations/"+smoke.FakeUUID(), nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}

func TestLocationHandler_Activate_Success(t *testing.T) {
	router := smoke.SetupRouter()

	mockUC := &MockLocationUseCase{
		ActivateLocationFunc: func(ctx context.Context, id uuidv7.UUID) error {
			return nil
		},
		GetLocationFunc: func(ctx context.Context, id uuidv7.UUID) (*location.Location, error) {
			return fakeLocation(), nil
		},
	}

	h := handler.NewLocationHandler(mockUC)
	router.PUT("/locations/:id/activate", h.Activate)

	w := smoke.MakeRequest(t, router, http.MethodPut, "/locations/"+smoke.FakeUUID()+"/activate", nil)
	smoke.AssertSuccessResponse(t, w, http.StatusOK)
}
