package http

import (
	"time"

	"github.com/basilex/promenade/internal/contexts/warehouse/product"
)

// CreateProductRequest represents the request to create a new product.
type CreateProductRequest struct {
	SKU  string `json:"sku" binding:"required" example:"PROD-001"`
	Name string `json:"name" binding:"required" example:"Laptop Dell XPS 15"`
}

// UpdateProductRequest represents the request to update a product.
type UpdateProductRequest struct {
	Name        string  `json:"name" binding:"required" example:"Laptop Dell XPS 15"`
	Description *string `json:"description,omitempty" example:"High-performance laptop"`
	Category    *string `json:"category,omitempty" example:"Electronics"`
	Brand       *string `json:"brand,omitempty" example:"Dell"`
}

// SetPhysicalPropertiesRequest represents the request to set physical properties.
type SetPhysicalPropertiesRequest struct {
	Weight float64    `json:"weight" binding:"gte=0" example:"2.5"`
	Length float64    `json:"length" binding:"gte=0" example:"35.7"`
	Width  float64    `json:"width" binding:"gte=0" example:"23.5"`
	Height float64    `json:"height" binding:"gte=0" example:"1.8"`
	Unit   string     `json:"unit" binding:"required,oneof=kg lb" example:"kg"`
}

// UpdateInventorySettingsRequest represents the request to update inventory settings.
type UpdateInventorySettingsRequest struct {
	TrackInventory bool `json:"track_inventory" example:"true"`
	AllowBackorder bool `json:"allow_backorder" example:"false"`
}

// SetReorderPointRequest represents the request to set reorder point.
type SetReorderPointRequest struct {
	ReorderPoint    int `json:"reorder_point" binding:"required,gte=0" example:"10"`
	ReorderQuantity int `json:"reorder_quantity" binding:"required,gte=0" example:"50"`
}

// ProductResponse represents a product in API responses.
type ProductResponse struct {
	ID          string     `json:"id" example:"01JGABC123DEF456GHI789JKL0"`
	SKU         string     `json:"sku" example:"PROD-001"`
	Name        string     `json:"name" example:"Laptop Dell XPS 15"`
	Description *string    `json:"description,omitempty" example:"High-performance laptop"`
	Category    *string    `json:"category,omitempty" example:"Electronics"`
	Brand       *string    `json:"brand,omitempty" example:"Dell"`
	Status      string     `json:"status" example:"active"`
	IsActive    bool       `json:"is_active" example:"true"`
	
	// Inventory
	TrackInventory bool `json:"track_inventory" example:"true"`
	AllowBackorder bool `json:"allow_backorder" example:"false"`
	ReorderPoint   int  `json:"reorder_point" example:"10"`
	ReorderQuantity int `json:"reorder_quantity" example:"50"`
	
	// Serial/Lot tracking
	TrackSerial bool `json:"track_serial" example:"false"`
	TrackLot    bool `json:"track_lot" example:"false"`
	
	// Physical properties
	Weight     float64    `json:"weight" example:"2.5"`
	Dimensions Dimensions `json:"dimensions"`
	
	// Timestamps
	CreatedAt time.Time  `json:"created_at" example:"2026-01-06T12:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2026-01-06T12:00:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" example:"2026-01-06T12:00:00Z"`
}

// Dimensions represents product dimensions.
type Dimensions struct {
	Length float64 `json:"length" example:"35.7"`
	Width  float64 `json:"width" example:"23.5"`
	Height float64 `json:"height" example:"1.8"`
	Unit   string  `json:"unit" example:"cm"`
}

// ProductListResponse represents a paginated list of products.
type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	Total      int               `json:"total" example:"100"`
	Page       int               `json:"page" example:"1"`
	PageSize   int               `json:"page_size" example:"20"`
	TotalPages int               `json:"total_pages" example:"5"`
}

// ToProductResponse converts Product entity to ProductResponse DTO.
func ToProductResponse(p *product.Product) ProductResponse {
	var deletedAt *time.Time
	if p.IsDeleted {
		deletedAt = p.DeletedAt
	}

	return ProductResponse{
		ID:              p.GetID().String(),
		SKU:             p.SKU,
		Name:            p.Name,
		Description:     stringPtr(p.Description),
		Category:        stringPtr(p.Category),
		Brand:           stringPtr(p.Brand),
		Status:          string(p.Status),
		IsActive:        p.IsActive,
		TrackInventory:  p.TrackInventory,
		AllowBackorder:  p.AllowBackorder,
		ReorderPoint:    p.ReorderPoint,
		ReorderQuantity: p.ReorderQuantity,
		TrackSerial:     p.TrackSerialNumbers,
		TrackLot:        p.TrackLotNumbers,
		Weight:          p.Weight,
		Dimensions: Dimensions{
			Length: p.Dimensions.Length,
			Width:  p.Dimensions.Width,
			Height: p.Dimensions.Height,
			Unit:   "cm", // Default unit from entity
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// ToProductListResponse converts list of products to ProductListResponse.
func ToProductListResponse(products []*product.Product, total, page, pageSize int) ProductListResponse {
	responses := make([]ProductResponse, len(products))
	for i, p := range products {
		responses[i] = ToProductResponse(p)
	}

	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return ProductListResponse{
		Products:   responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// stringPtr returns nil for empty strings, otherwise pointer to string.
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
