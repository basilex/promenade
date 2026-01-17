package aggregate

import (
	producterrors "github.com/basilex/promenade/internal/contexts/warehouse/product"
	"github.com/basilex/promenade/pkg/aggregate"
	"github.com/basilex/promenade/pkg/jsonstore"
)

// Product is the aggregate root for product catalog and specifications.
//
// Business Rules:
// - SKU must be unique across all products
// - Weight and dimensions required for physical products
// - Cannot discontinue product with active inventory
// - ReorderPoint must be less than ReorderQuantity
// - Name and SKU are required fields
type Product struct {
	aggregate.BaseAggregate

	// Identity
	SKU         string // Stock Keeping Unit (unique identifier)
	Name        string
	Description string

	// Classification
	Category string
	Brand    string
	Tags     jsonstore.Field[[]string] // JSONB tags (cross-database compatible)

	// Physical Properties
	Weight     float64    // Weight in kg
	Dimensions Dimensions // Length, width, height in cm

	// Inventory Settings
	TrackInventory  bool // If false, product has unlimited stock
	AllowBackorder  bool // Allow orders when out of stock
	ReorderPoint    int  // Low stock threshold (triggers reorder alert)
	ReorderQuantity int  // Quantity to order when below reorder point

	// Serial/Lot Tracking
	TrackSerialNumbers bool // Track individual units by serial number
	TrackLotNumbers    bool // Track batches by lot number

	// Status
	Status    ProductStatus
	IsActive  bool
	IsDeleted bool
}

// Dimensions represents product physical dimensions.
type Dimensions struct {
	Length float64 // Length in cm
	Width  float64 // Width in cm
	Height float64 // Height in cm
}

// ProductStatus represents the lifecycle state of a product.
type ProductStatus string

const (
	ProductStatusActive       ProductStatus = "active"       // Available for sale
	ProductStatusDiscontinued ProductStatus = "discontinued" // No longer produced/sold
	ProductStatusOutOfStock   ProductStatus = "out_of_stock" // Temporarily unavailable
	ProductStatusDraft        ProductStatus = "draft"        // Not yet published
)
// NewProduct creates a new Product aggregate with required fields.
//
// Business rules enforced:
// - SKU and Name are required
// - Product starts in Draft status
// - Inventory tracking enabled by default
// - Backorders disabled by default
func NewProduct(sku, name string) (*Product, error) {
	if sku == "" {
		return nil, producterrors.ErrProductSKURequired
	}
	if name == "" {
		return nil, producterrors.ErrProductNameRequired
	}

	return &Product{
		BaseAggregate: aggregate.NewBaseAggregate(),
		SKU:           sku,
		Name:          name,
		Status:        ProductStatusDraft,
		IsActive:      false,
		TrackInventory: true, // Default: track inventory
		AllowBackorder: false, // Default: no backorders
	}, nil
}

// SetDescription updates the product description.
func (p *Product) SetDescription(description string) {
	p.Description = description
	p.Touch()
}

// SetClassification updates product classification (category, brand, tags).
func (p *Product) SetClassification(category, brand string, tags []string) {
	p.Category = category
	p.Brand = brand
	p.Tags.Set(tags)
	p.Touch()
}

// SetPhysicalProperties sets weight and dimensions for physical products.
//
// Business rule: Weight and dimensions must be positive values.
func (p *Product) SetPhysicalProperties(weight float64, dimensions Dimensions) error {
	if weight < 0 {
		return producterrors.ErrProductInvalidWeight
	}
	if dimensions.Length < 0 || dimensions.Width < 0 || dimensions.Height < 0 {
		return producterrors.ErrProductInvalidDimension
	}

	p.Weight = weight
	p.Dimensions = dimensions
	p.Touch()
	return nil
}

// SetInventorySettings configures inventory tracking and backorder settings.
func (p *Product) SetInventorySettings(trackInventory, allowBackorder bool) {
	p.TrackInventory = trackInventory
	p.AllowBackorder = allowBackorder
	p.Touch()
}

// SetReorderPoint sets the low stock threshold and reorder quantity.
//
// Business rule: ReorderPoint must be less than ReorderQuantity.
func (p *Product) SetReorderPoint(point, quantity int) error {
	if point < 0 || quantity < 0 {
		return producterrors.ErrProductInvalidReorder
	}
	if point >= quantity && quantity > 0 {
		return producterrors.ErrProductInvalidReorder
	}

	p.ReorderPoint = point
	p.ReorderQuantity = quantity
	p.Touch()
	return nil
}

// SetSerialTracking enables/disables serial number tracking.
func (p *Product) SetSerialTracking(enabled bool) {
	p.TrackSerialNumbers = enabled
	p.Touch()
}

// SetLotTracking enables/disables lot number tracking.
func (p *Product) SetLotTracking(enabled bool) {
	p.TrackLotNumbers = enabled
	p.Touch()
}

// Activate transitions product to Active status.
//
// Business rule: Product must be in Draft, OutOfStock, or Discontinued status.
func (p *Product) Activate() error {
	if p.Status == ProductStatusActive {
		return producterrors.ErrProductAlreadyActive
	}

	p.Status = ProductStatusActive
	p.IsActive = true
	p.Touch()
	return nil
}

// Deactivate transitions product to OutOfStock status.
//
// Business rule: Product must be Active.
func (p *Product) Deactivate() error {
	if !p.IsActive {
		return producterrors.ErrProductAlreadyInactive
	}

	p.Status = ProductStatusOutOfStock
	p.IsActive = false
	p.Touch()
	return nil
}

// Discontinue marks product as discontinued (no longer produced/sold).
//
// Business rule: Once discontinued, product cannot be reactivated.
// Use case: Product end-of-life, replaced by newer version.
func (p *Product) Discontinue() error {
	if p.Status == ProductStatusDiscontinued {
		return producterrors.ErrProductDiscontinued
	}

	p.Status = ProductStatusDiscontinued
	p.IsActive = false
	p.Touch()
	return nil
}

// MarkAsOutOfStock transitions product to OutOfStock status.
//
// Business rule: Only Active products can be marked out of stock.
func (p *Product) MarkAsOutOfStock() error {
	if p.Status != ProductStatusActive {
		return producterrors.ErrProductNotActive
	}

	p.Status = ProductStatusOutOfStock
	p.Touch()
	return nil
}

// RestockFromOutOfStock transitions product back to Active status.
//
// Business rule: Only OutOfStock products can be restocked.
func (p *Product) RestockFromOutOfStock() error {
	if p.Status != ProductStatusOutOfStock {
		return producterrors.ErrProductNotOutOfStock
	}

	p.Status = ProductStatusActive
	p.Touch()
	return nil
}

// IsPhysicalProduct returns true if product has weight or dimensions.
func (p *Product) IsPhysicalProduct() bool {
	return p.Weight > 0 || p.Dimensions.Length > 0 || p.Dimensions.Width > 0 || p.Dimensions.Height > 0
}

// GetVolume calculates product volume in cubic centimeters.
func (p *Product) GetVolume() float64 {
	return p.Dimensions.Length * p.Dimensions.Width * p.Dimensions.Height
}

// GetVolumeInCubicMeters calculates product volume in cubic meters.
func (p *Product) GetVolumeInCubicMeters() float64 {
	return p.GetVolume() / 1000000 // Convert cm³ to m³
}

// RequiresReorder checks if product inventory is below reorder point.
// This is a helper for integration with Inventory aggregate.
func (p *Product) RequiresReorder(currentStock int) bool {
	if !p.TrackInventory {
		return false
	}
	return currentStock <= p.ReorderPoint
}

// CanFulfillOrder checks if product can fulfill an order.
//
// Business rules:
// - Discontinued products cannot fulfill orders
// - If inventory tracking disabled, always can fulfill
// - If backorders allowed, always can fulfill
// - Otherwise, check stock availability (handled by Inventory aggregate)
func (p *Product) CanFulfillOrder() bool {
	if p.Status == ProductStatusDiscontinued {
		return false
	}
	if !p.TrackInventory {
		return true // Unlimited stock
	}
	if p.AllowBackorder {
		return true // Backorders allowed
	}
	// Stock check delegated to Inventory aggregate
	return true
}

// Validate ensures product data integrity.
//
// Business rules enforced:
// - SKU and Name are required
// - ReorderPoint < ReorderQuantity (if set)
// - Weight cannot be negative (if set)
// - Dimensions cannot be negative (if set)
func (p *Product) Validate() error {
	if p.SKU == "" {
		return producterrors.ErrProductSKURequired
	}
	if p.Name == "" {
		return producterrors.ErrProductNameRequired
	}
	if p.ReorderQuantity > 0 && p.ReorderPoint >= p.ReorderQuantity {
		return producterrors.ErrProductInvalidReorder
	}
	if p.Weight < 0 {
		return producterrors.ErrProductInvalidWeight
	}
	if p.Dimensions.Length < 0 || p.Dimensions.Width < 0 || p.Dimensions.Height < 0 {
		return producterrors.ErrProductInvalidDimension
	}
	return nil
}

// SoftDelete marks product as deleted (soft delete pattern).
func (p *Product) SoftDelete() {
	p.IsDeleted = true
	p.IsActive = false
	p.Touch()
}

// Restore undeletes a soft-deleted product.
func (p *Product) Restore() {
	p.IsDeleted = false
	p.Touch()
}
