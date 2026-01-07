package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/warehouse/location"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CreateLocationRequest represents request to create a location
type CreateLocationRequest struct {
	ParentID       *string `json:"parent_id,omitempty"`
	Code           string  `json:"code" binding:"required,max=50"`
	Name           string  `json:"name" binding:"required,max=200"`
	Type           string  `json:"type" binding:"required,oneof=warehouse zone aisle rack shelf bin"`
	Description    *string `json:"description,omitempty"`
	Width          float64 `json:"width,omitempty"`
	Height         float64 `json:"height,omitempty"`
	Depth          float64 `json:"depth,omitempty"`
	Capacity       int     `json:"capacity,omitempty"`
	IsLimited      bool    `json:"is_limited"`
	IsPickable     bool    `json:"is_pickable"`
	IsPutawayable  bool    `json:"is_putawayable"`
	Notes          *string `json:"notes,omitempty"`
}

// UpdateLocationRequest represents request to update a location
type UpdateLocationRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,max=200"`
	Description *string `json:"description,omitempty"`
}

// UpdateCapacityRequest represents request to update capacity
type UpdateCapacityRequest struct {
	Capacity  int  `json:"capacity" binding:"required,min=0"`
	IsLimited bool `json:"is_limited"`
}

// UpdateDimensionsRequest represents request to update dimensions
type UpdateDimensionsRequest struct {
	Width  float64 `json:"width" binding:"required,min=0"`
	Height float64 `json:"height" binding:"required,min=0"`
	Depth  float64 `json:"depth" binding:"required,min=0"`
}

// UpdateFlagsRequest represents request to update operational flags
type UpdateFlagsRequest struct {
	IsPickable    bool `json:"is_pickable"`
	IsPutawayable bool `json:"is_putawayable"`
}

// LocationResponse represents location data in API response
type LocationResponse struct {
	ID                string   `json:"id"`
	ParentID          *string  `json:"parent_id,omitempty"`
	Type              string   `json:"type"`
	Path              string   `json:"path"`
	Code              string   `json:"code"`
	Name              string   `json:"name"`
	Level             int      `json:"level"`
	Description       string   `json:"description"`
	Status            string   `json:"status"`
	Width             *float64 `json:"width,omitempty"`
	Height            *float64 `json:"height,omitempty"`
	Depth             *float64 `json:"depth,omitempty"`
	Volume            *float64 `json:"volume,omitempty"`
	Capacity          int      `json:"capacity"`
	CurrentOccupancy  int      `json:"current_occupancy"`
	AvailableCapacity *int     `json:"available_capacity,omitempty"`
	IsLimited         bool     `json:"is_limited"`
	IsPickable        bool     `json:"is_pickable"`
	IsPutawayable     bool     `json:"is_putawayable"`
	Notes             string   `json:"notes"`
	Version           int      `json:"version"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// LocationListResponse represents paginated list response
type LocationListResponse struct {
	Locations  []LocationResponse `json:"locations"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
}

// CapacityResponse represents capacity information
type CapacityResponse struct {
	LocationID        string `json:"location_id"`
	LocationCode      string `json:"location_code"`
	Capacity          int    `json:"capacity"`
	CurrentOccupancy  int    `json:"current_occupancy"`
	AvailableCapacity int    `json:"available_capacity"`
	IsLimited         bool   `json:"is_limited"`
	IsFull            bool   `json:"is_full"`
}

// ToLocationResponse converts Location entity to response DTO
func ToLocationResponse(loc *location.Location) LocationResponse {
	resp := LocationResponse{
		ID:               loc.ID.String(),
		Code:             loc.Code,
		Name:             loc.Name,
		Description:      loc.Description,
		Type:             string(loc.Type),
		Path:             loc.Path,
		Level:            loc.Level,
		Status:           string(loc.Status),
		Capacity:         loc.Capacity,
		CurrentOccupancy: loc.CurrentOccupancy,
		IsLimited:        loc.IsLimited,
		IsPickable:       loc.IsPickable,
		IsPutawayable:    loc.IsPutawayable,
		Notes:            loc.Notes,
		Version:          loc.Version,
		CreatedAt:        loc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        loc.UpdatedAt.Format(time.RFC3339),
	}

	if loc.ParentID != nil {
		parentIDStr := loc.ParentID.String()
		resp.ParentID = &parentIDStr
	}

	if loc.Width > 0 {
		resp.Width = &loc.Width
	}
	if loc.Height > 0 {
		resp.Height = &loc.Height
	}
	if loc.Depth > 0 {
		resp.Depth = &loc.Depth
	}

	// Calculate volume if dimensions are set
	if loc.Width > 0 && loc.Height > 0 && loc.Depth > 0 {
		volume := loc.Width * loc.Height * loc.Depth
		resp.Volume = &volume
	}

	if loc.IsLimited {
		available := loc.Capacity - loc.CurrentOccupancy
		if available < 0 {
			available = 0
		}
		resp.AvailableCapacity = &available
	}

	return resp
}

// ToLocationListResponse converts slice of locations to list response
func ToLocationListResponse(locations []*location.Location, total, page, pageSize int) LocationListResponse {
	resp := LocationListResponse{
		Locations:  make([]LocationResponse, 0, len(locations)),
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	for _, loc := range locations {
		resp.Locations = append(resp.Locations, ToLocationResponse(loc))
	}

	return resp
}

// ToCapacityResponse converts Location to capacity response
func ToCapacityResponse(loc *location.Location) CapacityResponse {
	available := max(loc.Capacity - loc.CurrentOccupancy, 0)

	return CapacityResponse{
		LocationID:        loc.ID.String(),
		LocationCode:      loc.Code,
		Capacity:          loc.Capacity,
		CurrentOccupancy:  loc.CurrentOccupancy,
		AvailableCapacity: available,
		IsLimited:         loc.IsLimited,
		IsFull:            loc.Status == location.LocationStatusFull,
	}
}

// ParseLocationType converts string to LocationType
func ParseLocationType(typeStr string) location.LocationType {
	return location.LocationType(typeStr)
}

// ParseLocationStatus converts string to LocationStatus
func ParseLocationStatus(statusStr string) location.LocationStatus {
	return location.LocationStatus(statusStr)
}

// ParseUUID safely parses UUID string
func ParseUUID(uuidStr string) (*uuidv7.UUID, error) {
	if uuidStr == "" {
		return nil, nil
	}
	id, err := uuidv7.Parse(uuidStr)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
