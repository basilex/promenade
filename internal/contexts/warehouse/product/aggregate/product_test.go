package aggregate

import (
	producterrors "github.com/basilex/promenade/internal/contexts/warehouse/product"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name    string
		sku     string
		prodName string
		wantErr error
	}{
		{
			name:     "valid product",
			sku:      "PROD-001",
			prodName: "Test Product",
			wantErr:  nil,
		},
		{
			name:     "empty SKU",
			sku:      "",
			prodName: "Test Product",
			wantErr:  producterrors.ErrProductSKURequired,
		},
		{
			name:     "empty name",
			sku:      "PROD-001",
			prodName: "",
			wantErr:  producterrors.ErrProductNameRequired,
		},
		{
			name:     "both empty",
			sku:      "",
			prodName: "",
			wantErr:  producterrors.ErrProductSKURequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(tt.sku, tt.prodName)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, product)
			} else {
				require.NoError(t, err)
				require.NotNil(t, product)
				assert.Equal(t, tt.sku, product.SKU)
				assert.Equal(t, tt.prodName, product.Name)
				assert.Equal(t, ProductStatusDraft, product.Status)
				assert.False(t, product.IsActive)
				assert.True(t, product.TrackInventory) // Default
				assert.False(t, product.AllowBackorder) // Default
				assert.NotNil(t, product.Tags)
				assert.Empty(t, product.Tags)
			}
		})
	}
}

func TestProduct_SetDescription(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")
	product.SetDescription("Test description")

	assert.Equal(t, "Test description", product.Description)
}

func TestProduct_SetClassification(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")
	tags := []string{"electronics", "gadgets"}

	product.SetClassification("Electronics", "Apple", tags)

	assert.Equal(t, "Electronics", product.Category)
	assert.Equal(t, "Apple", product.Brand)
	assert.Equal(t, tags, product.Tags.Get())
}

func TestProduct_SetPhysicalProperties(t *testing.T) {
	tests := []struct {
		name       string
		weight     float64
		dimensions Dimensions
		wantErr    error
	}{
		{
			name:   "valid properties",
			weight: 1.5,
			dimensions: Dimensions{
				Length: 10.0,
				Width:  5.0,
				Height: 3.0,
			},
			wantErr: nil,
		},
		{
			name:   "zero values allowed",
			weight: 0,
			dimensions: Dimensions{
				Length: 0,
				Width:  0,
				Height: 0,
			},
			wantErr: nil,
		},
		{
			name:   "negative weight",
			weight: -1.0,
			dimensions: Dimensions{
				Length: 10.0,
				Width:  5.0,
				Height: 3.0,
			},
			wantErr: producterrors.ErrProductInvalidWeight,
		},
		{
			name:   "negative length",
			weight: 1.5,
			dimensions: Dimensions{
				Length: -10.0,
				Width:  5.0,
				Height: 3.0,
			},
			wantErr: producterrors.ErrProductInvalidDimension,
		},
		{
			name:   "negative width",
			weight: 1.5,
			dimensions: Dimensions{
				Length: 10.0,
				Width:  -5.0,
				Height: 3.0,
			},
			wantErr: producterrors.ErrProductInvalidDimension,
		},
		{
			name:   "negative height",
			weight: 1.5,
			dimensions: Dimensions{
				Length: 10.0,
				Width:  5.0,
				Height: -3.0,
			},
			wantErr: producterrors.ErrProductInvalidDimension,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			err := product.SetPhysicalProperties(tt.weight, tt.dimensions)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.weight, product.Weight)
				assert.Equal(t, tt.dimensions, product.Dimensions)
			}
		})
	}
}

func TestProduct_SetInventorySettings(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")

	product.SetInventorySettings(false, true)

	assert.False(t, product.TrackInventory)
	assert.True(t, product.AllowBackorder)
}

func TestProduct_SetReorderPoint(t *testing.T) {
	tests := []struct {
		name     string
		point    int
		quantity int
		wantErr  error
	}{
		{
			name:     "valid reorder point",
			point:    10,
			quantity: 50,
			wantErr:  nil,
		},
		{
			name:     "equal values",
			point:    50,
			quantity: 50,
			wantErr:  producterrors.ErrProductInvalidReorder,
		},
		{
			name:     "point greater than quantity",
			point:    100,
			quantity: 50,
			wantErr:  producterrors.ErrProductInvalidReorder,
		},
		{
			name:     "negative point",
			point:    -10,
			quantity: 50,
			wantErr:  producterrors.ErrProductInvalidReorder,
		},
		{
			name:     "negative quantity",
			point:    10,
			quantity: -50,
			wantErr:  producterrors.ErrProductInvalidReorder,
		},
		{
			name:     "zero quantity allowed",
			point:    0,
			quantity: 0,
			wantErr:  nil,
		},
		{
			name:     "point with zero quantity",
			point:    10,
			quantity: 0,
			wantErr:  nil, // Special case: quantity 0 means no reorder
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			err := product.SetReorderPoint(tt.point, tt.quantity)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.point, product.ReorderPoint)
				assert.Equal(t, tt.quantity, product.ReorderQuantity)
			}
		})
	}
}

func TestProduct_SetSerialTracking(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")

	product.SetSerialTracking(true)
	assert.True(t, product.TrackSerialNumbers)

	product.SetSerialTracking(false)
	assert.False(t, product.TrackSerialNumbers)
}

func TestProduct_SetLotTracking(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")

	product.SetLotTracking(true)
	assert.True(t, product.TrackLotNumbers)

	product.SetLotTracking(false)
	assert.False(t, product.TrackLotNumbers)
}

func TestProduct_Activate(t *testing.T) {
	tests := []struct {
		name        string
		setupStatus ProductStatus
		wantErr     error
	}{
		{
			name:        "activate from draft",
			setupStatus: ProductStatusDraft,
			wantErr:     nil,
		},
		{
			name:        "activate from out of stock",
			setupStatus: ProductStatusOutOfStock,
			wantErr:     nil,
		},
		{
			name:        "activate from discontinued",
			setupStatus: ProductStatusDiscontinued,
			wantErr:     nil,
		},
		{
			name:        "activate already active",
			setupStatus: ProductStatusActive,
			wantErr:     producterrors.ErrProductAlreadyActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			product.Status = tt.setupStatus
			if tt.setupStatus == ProductStatusActive {
				product.IsActive = true
			}

			err := product.Activate()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, ProductStatusActive, product.Status)
				assert.True(t, product.IsActive)
			}
		})
	}
}

func TestProduct_Deactivate(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		wantErr  error
	}{
		{
			name:     "deactivate active product",
			isActive: true,
			wantErr:  nil,
		},
		{
			name:     "deactivate inactive product",
			isActive: false,
			wantErr:  producterrors.ErrProductAlreadyInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			product.IsActive = tt.isActive
			if tt.isActive {
				product.Status = ProductStatusActive
			}

			err := product.Deactivate()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, ProductStatusOutOfStock, product.Status)
				assert.False(t, product.IsActive)
			}
		})
	}
}

func TestProduct_Discontinue(t *testing.T) {
	tests := []struct {
		name        string
		setupStatus ProductStatus
		wantErr     error
	}{
		{
			name:        "discontinue active product",
			setupStatus: ProductStatusActive,
			wantErr:     nil,
		},
		{
			name:        "discontinue draft product",
			setupStatus: ProductStatusDraft,
			wantErr:     nil,
		},
		{
			name:        "discontinue already discontinued",
			setupStatus: ProductStatusDiscontinued,
			wantErr:     producterrors.ErrProductDiscontinued,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			product.Status = tt.setupStatus

			err := product.Discontinue()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, ProductStatusDiscontinued, product.Status)
				assert.False(t, product.IsActive)
			}
		})
	}
}

func TestProduct_MarkAsOutOfStock(t *testing.T) {
	tests := []struct {
		name        string
		setupStatus ProductStatus
		wantErr     bool
	}{
		{
			name:        "mark active product out of stock",
			setupStatus: ProductStatusActive,
			wantErr:     false,
		},
		{
			name:        "mark draft product out of stock",
			setupStatus: ProductStatusDraft,
			wantErr:     true,
		},
		{
			name:        "mark discontinued product out of stock",
			setupStatus: ProductStatusDiscontinued,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			product.Status = tt.setupStatus

			err := product.MarkAsOutOfStock()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, ProductStatusOutOfStock, product.Status)
			}
		})
	}
}

func TestProduct_RestockFromOutOfStock(t *testing.T) {
	tests := []struct {
		name        string
		setupStatus ProductStatus
		wantErr     bool
	}{
		{
			name:        "restock out of stock product",
			setupStatus: ProductStatusOutOfStock,
			wantErr:     false,
		},
		{
			name:        "restock active product",
			setupStatus: ProductStatusActive,
			wantErr:     true,
		},
		{
			name:        "restock discontinued product",
			setupStatus: ProductStatusDiscontinued,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			product.Status = tt.setupStatus

			err := product.RestockFromOutOfStock()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, ProductStatusActive, product.Status)
			}
		})
	}
}

func TestProduct_IsPhysicalProduct(t *testing.T) {
	tests := []struct {
		name       string
		weight     float64
		dimensions Dimensions
		want       bool
	}{
		{
			name:   "product with weight",
			weight: 1.5,
			dimensions: Dimensions{
				Length: 0,
				Width:  0,
				Height: 0,
			},
			want: true,
		},
		{
			name:   "product with dimensions",
			weight: 0,
			dimensions: Dimensions{
				Length: 10.0,
				Width:  5.0,
				Height: 3.0,
			},
			want: true,
		},
		{
			name:   "product with weight and dimensions",
			weight: 1.5,
			dimensions: Dimensions{
				Length: 10.0,
				Width:  5.0,
				Height: 3.0,
			},
			want: true,
		},
		{
			name:   "digital product (no weight/dimensions)",
			weight: 0,
			dimensions: Dimensions{
				Length: 0,
				Width:  0,
				Height: 0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			_ = product.SetPhysicalProperties(tt.weight, tt.dimensions)

			assert.Equal(t, tt.want, product.IsPhysicalProduct())
		})
	}
}

func TestProduct_GetVolume(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")
	_ = product.SetPhysicalProperties(1.5, Dimensions{
		Length: 10.0,
		Width:  5.0,
		Height: 2.0,
	})

	volume := product.GetVolume()
	assert.Equal(t, 100.0, volume) // 10 * 5 * 2 = 100 cm³
}

func TestProduct_GetVolumeInCubicMeters(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")
	_ = product.SetPhysicalProperties(1.5, Dimensions{
		Length: 100.0, // 1 meter
		Width:  100.0, // 1 meter
		Height: 100.0, // 1 meter
	})

	volume := product.GetVolumeInCubicMeters()
	assert.Equal(t, 1.0, volume) // 1,000,000 cm³ = 1 m³
}

func TestProduct_RequiresReorder(t *testing.T) {
	tests := []struct {
		name             string
		trackInventory   bool
		reorderPoint     int
		currentStock     int
		want             bool
	}{
		{
			name:           "stock below reorder point",
			trackInventory: true,
			reorderPoint:   10,
			currentStock:   5,
			want:           true,
		},
		{
			name:           "stock at reorder point",
			trackInventory: true,
			reorderPoint:   10,
			currentStock:   10,
			want:           true,
		},
		{
			name:           "stock above reorder point",
			trackInventory: true,
			reorderPoint:   10,
			currentStock:   15,
			want:           false,
		},
		{
			name:           "inventory tracking disabled",
			trackInventory: false,
			reorderPoint:   10,
			currentStock:   5,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			product.SetInventorySettings(tt.trackInventory, false)
			_ = product.SetReorderPoint(tt.reorderPoint, 50)

			result := product.RequiresReorder(tt.currentStock)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestProduct_CanFulfillOrder(t *testing.T) {
	tests := []struct {
		name           string
		status         ProductStatus
		trackInventory bool
		allowBackorder bool
		want           bool
	}{
		{
			name:           "active product with tracking and backorders",
			status:         ProductStatusActive,
			trackInventory: true,
			allowBackorder: true,
			want:           true,
		},
		{
			name:           "active product without tracking",
			status:         ProductStatusActive,
			trackInventory: false,
			allowBackorder: false,
			want:           true,
		},
		{
			name:           "discontinued product",
			status:         ProductStatusDiscontinued,
			trackInventory: true,
			allowBackorder: true,
			want:           false,
		},
		{
			name:           "active product with tracking but no backorders",
			status:         ProductStatusActive,
			trackInventory: true,
			allowBackorder: false,
			want:           true, // Delegated to Inventory aggregate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			product.Status = tt.status
			product.SetInventorySettings(tt.trackInventory, tt.allowBackorder)

			result := product.CanFulfillOrder()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestProduct_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Product)
		wantErr error
	}{
		{
			name:    "valid product",
			setup:   func(p *Product) {},
			wantErr: nil,
		},
		{
			name: "empty SKU",
			setup: func(p *Product) {
				p.SKU = ""
			},
			wantErr: producterrors.ErrProductSKURequired,
		},
		{
			name: "empty name",
			setup: func(p *Product) {
				p.Name = ""
			},
			wantErr: producterrors.ErrProductNameRequired,
		},
		{
			name: "invalid reorder settings",
			setup: func(p *Product) {
				p.ReorderPoint = 100
				p.ReorderQuantity = 50
			},
			wantErr: producterrors.ErrProductInvalidReorder,
		},
		{
			name: "negative weight for physical product",
			setup: func(p *Product) {
				p.Weight = -1.0
			},
			wantErr: producterrors.ErrProductInvalidWeight,
		},
		{
			name: "negative dimension",
			setup: func(p *Product) {
				p.Dimensions = Dimensions{
					Length: -10.0,
					Width:  5.0,
					Height: 3.0,
				}
			},
			wantErr: producterrors.ErrProductInvalidDimension,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, _ := NewProduct("PROD-001", "Test Product")
			tt.setup(product)

			err := product.Validate()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProduct_SoftDelete(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")
	product.SoftDelete()

	assert.True(t, product.IsDeleted)
	assert.False(t, product.IsActive)
}

func TestProduct_Restore(t *testing.T) {
	product, _ := NewProduct("PROD-001", "Test Product")
	product.SoftDelete()

	product.Restore()

	assert.False(t, product.IsDeleted)
}
