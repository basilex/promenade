package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/warehouse/location"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// LocationHandler handles HTTP requests for location operations
type LocationHandler struct {
	usecase location.IUseCase
}

// NewLocationHandler creates a new location handler
func NewLocationHandler(usecase location.IUseCase) *LocationHandler {
	return &LocationHandler{
		usecase: usecase,
	}
}

// Create creates a new location
// @Summary Create location
// @Description Create a new warehouse location
// @Tags Location
// @Accept json
// @Produce json
// @Param request body CreateLocationRequest true "Location creation request"
// @Success 201 {object} response.Response{data=LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations [post]
func (h *LocationHandler) Create(c *gin.Context) {
	var req CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Parse parent ID if provided
	var parentID *uuidv7.UUID
	if req.ParentID != nil {
		id, err := uuidv7.Parse(*req.ParentID)
		if err != nil {
			response.BadRequest(c, "Invalid parent ID format")
			return
		}
		parentID = &id
	}

	// Parse type
	locationType := ParseLocationType(req.Type)

	// Create location
	loc, err := h.usecase.CreateLocation(
		c.Request.Context(),
		req.Code,
		req.Name,
		locationType,
		req.Description,
		parentID,
	)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update dimensions if provided
	if req.Width > 0 && req.Height > 0 && req.Depth > 0 {
		if err := h.usecase.UpdateLocationDimensions(c.Request.Context(), loc.ID, req.Width, req.Height, req.Depth); err != nil {
			response.InternalError(c, err.Error())
			return
		}
		// Reload to get updated dimensions
		loc, _ = h.usecase.GetLocation(c.Request.Context(), loc.ID)
	}

	// Update capacity if provided
	if req.Capacity > 0 {
		if err := h.usecase.UpdateLocationCapacity(c.Request.Context(), loc.ID, req.Capacity, req.IsLimited); err != nil {
			response.InternalError(c, err.Error())
			return
		}
		// Reload to get updated capacity
		loc, _ = h.usecase.GetLocation(c.Request.Context(), loc.ID)
	}

	// Update flags
	if err := h.usecase.UpdateLocationFlags(c.Request.Context(), loc.ID, req.IsPickable, req.IsPutawayable); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// Reload final state
	loc, _ = h.usecase.GetLocation(c.Request.Context(), loc.ID)

	response.Created(c, ToLocationResponse(loc))
}

// GetByID retrieves a location by ID
// @Summary Get location by ID
// @Description Get location details by ID
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id} [get]
func (h *LocationHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	loc, err := h.usecase.GetLocation(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, ToLocationResponse(loc))
}

// GetByCode retrieves a location by code
// @Summary Get location by code
// @Description Get location details by unique code
// @Tags Location
// @Accept json
// @Produce json
// @Param code path string true "Location code"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/code/{code} [get]
func (h *LocationHandler) GetByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Location code is required")
		return
	}

	loc, err := h.usecase.GetLocationByCode(c.Request.Context(), code)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, ToLocationResponse(loc))
}

// Update updates a location
// @Summary Update location
// @Description Update location details
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Param request body UpdateLocationRequest true "Update request"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id} [put]
func (h *LocationHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	loc, err := h.usecase.UpdateLocation(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, ToLocationResponse(loc))
}

// Delete deletes a location
// @Summary Delete location
// @Description Soft delete a location
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id} [delete]
func (h *LocationHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	if err := h.usecase.DeleteLocation(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Location deleted successfully"})
}

// List retrieves locations with pagination and filters
// @Summary List locations
// @Description Get paginated list of locations with optional filters
// @Tags Location
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param type query string false "Location type filter"
// @Param status query string false "Location status filter"
// @Param available query bool false "Available locations only"
// @Success 200 {object} response.Response{data=LocationListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations [get]
func (h *LocationHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	typeFilter := c.Query("type")
	statusFilter := c.Query("status")
	availableOnly := c.Query("available") == "true"

	var locations []*location.Location
	var total int
	var err error

	// Apply filters
	if availableOnly {
		locations, total, err = h.usecase.ListAvailableLocations(c.Request.Context(), page, pageSize)
	} else if typeFilter != "" {
		locationType := ParseLocationType(typeFilter)
		locations, total, err = h.usecase.ListLocationsByType(c.Request.Context(), locationType, page, pageSize)
	} else if statusFilter != "" {
		status := ParseLocationStatus(statusFilter)
		locations, total, err = h.usecase.ListLocationsByStatus(c.Request.Context(), status, page, pageSize)
	} else {
		locations, total, err = h.usecase.ListLocations(c.Request.Context(), page, pageSize)
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, ToLocationListResponse(locations, total, page, pageSize))
}

// GetChildren retrieves child locations
// @Summary Get location children
// @Description Get direct or recursive child locations
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Param recursive query bool false "Include nested children" default(false)
// @Success 200 {object} response.Response{data=[]LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/children [get]
func (h *LocationHandler) GetChildren(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	recursive := c.Query("recursive") == "true"

	children, err := h.usecase.GetLocationChildren(c.Request.Context(), id, recursive)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	childrenResp := make([]LocationResponse, 0, len(children))
	for _, child := range children {
		childrenResp = append(childrenResp, ToLocationResponse(child))
	}

	response.Success(c, childrenResp)
}

// GetHierarchy retrieves location hierarchy path
// @Summary Get location hierarchy
// @Description Get full hierarchy path from root to location
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response{data=[]LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/hierarchy [get]
func (h *LocationHandler) GetHierarchy(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	hierarchy, err := h.usecase.GetLocationHierarchy(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	hierarchyResp := make([]LocationResponse, 0, len(hierarchy))
	for _, loc := range hierarchy {
		hierarchyResp = append(hierarchyResp, ToLocationResponse(loc))
	}

	response.Success(c, hierarchyResp)
}

// Activate activates a location
// @Summary Activate location
// @Description Set location status to active
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/activate [put]
func (h *LocationHandler) Activate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	if err := h.usecase.ActivateLocation(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	loc, _ := h.usecase.GetLocation(c.Request.Context(), id)
	response.Success(c, ToLocationResponse(loc))
}

// Deactivate deactivates a location
// @Summary Deactivate location
// @Description Set location status to inactive
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/deactivate [put]
func (h *LocationHandler) Deactivate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	if err := h.usecase.DeactivateLocation(c.Request.Context(), id); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	loc, _ := h.usecase.GetLocation(c.Request.Context(), id)
	response.Success(c, ToLocationResponse(loc))
}

// UpdateCapacity updates location capacity settings
// @Summary Update capacity
// @Description Update location capacity and limited flag
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Param request body UpdateCapacityRequest true "Capacity update request"
// @Success 200 {object} response.Response{data=CapacityResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/capacity [put]
func (h *LocationHandler) UpdateCapacity(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	var req UpdateCapacityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateLocationCapacity(c.Request.Context(), id, req.Capacity, req.IsLimited); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	loc, _ := h.usecase.GetLocation(c.Request.Context(), id)
	response.Success(c, ToCapacityResponse(loc))
}

// UpdateDimensions updates location dimensions (width, height, depth)
// @Summary Update location dimensions
// @Description Update physical dimensions of a location
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Param request body UpdateDimensionsRequest true "Dimensions update request"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/dimensions [put]
func (h *LocationHandler) UpdateDimensions(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	var req UpdateDimensionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateLocationDimensions(c.Request.Context(), id, req.Width, req.Height, req.Depth); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	loc, _ := h.usecase.GetLocation(c.Request.Context(), id)
	response.Success(c, ToLocationResponse(loc))
}

// UpdateFlags updates location operational flags (pickable, putawayable)
// @Summary Update location flags
// @Description Update operational flags for a location
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Param request body UpdateFlagsRequest true "Flags update request"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/flags [put]
func (h *LocationHandler) UpdateFlags(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	var req UpdateFlagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.usecase.UpdateLocationFlags(c.Request.Context(), id, req.IsPickable, req.IsPutawayable); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	loc, _ := h.usecase.GetLocation(c.Request.Context(), id)
	response.Success(c, ToLocationResponse(loc))
}

// SetMaintenance sets location to maintenance status
// @Summary Set location maintenance mode
// @Description Set a location to maintenance status (unavailable for operations)
// @Tags Location
// @Accept json
// @Produce json
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response{data=LocationResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/locations/{id}/maintenance [put]
func (h *LocationHandler) SetMaintenance(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid location ID format")
		return
	}

	if err := h.usecase.SetLocationMaintenance(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	loc, _ := h.usecase.GetLocation(c.Request.Context(), id)
	response.Success(c, ToLocationResponse(loc))
}
