package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/basilex/promenade/internal/contexts/warehouse/product"
	"github.com/basilex/promenade/pkg/response"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// ProductHandler handles HTTP requests for product operations.
type ProductHandler struct {
	usecase product.IUseCase
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(usecase product.IUseCase) *ProductHandler {
	return &ProductHandler{
		usecase: usecase,
	}
}

// Create creates a new product.
// @Summary Create product
// @Description Create a new product with SKU and name
// @Tags Products
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "Product creation request"
// @Success 201 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	p, err := h.usecase.CreateProduct(c.Request.Context(), req.SKU, req.Name)
	if err != nil {
		if errors.Is(err, product.ErrProductSKUDuplicate) {
			response.ErrorResponse(c, http.StatusConflict, "SKU_DUPLICATE", "Product with this SKU already exists")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create product")
		return
	}

	response.Created(c, ToProductResponse(p))
}

// GetByID retrieves a product by ID.
// @Summary Get product by ID
// @Description Get product details by product ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id} [get]
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// GetBySKU retrieves a product by SKU.
// @Summary Get product by SKU
// @Description Get product details by SKU code
// @Tags Products
// @Accept json
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/sku/{sku} [get]
func (h *ProductHandler) GetBySKU(c *gin.Context) {
	sku := c.Param("sku")

	p, err := h.usecase.GetProductBySKU(c.Request.Context(), sku)
	if err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// Update updates a product.
// @Summary Update product
// @Description Update product details
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param request body UpdateProductRequest true "Product update request"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Get existing product
	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	// Update fields
	p.Name = req.Name
	if req.Description != nil {
		p.SetDescription(*req.Description)
	}
	if req.Category != nil || req.Brand != nil {
		category := ""
		brand := ""
		var tags []string
		if req.Category != nil {
			category = *req.Category
		}
		if req.Brand != nil {
			brand = *req.Brand
		}
		p.SetClassification(category, brand, tags)
	}

	// Persist changes
	if err := h.usecase.UpdateProduct(c.Request.Context(), p); err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// Delete soft-deletes a product.
// @Summary Delete product
// @Description Soft-delete a product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 204
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	if err := h.usecase.DeleteProduct(c.Request.Context(), id); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete product")
		return
	}

	c.Status(http.StatusNoContent)
}

// List retrieves paginated list of products.
// @Summary List products
// @Description Get paginated list of products
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=ProductListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products [get]
func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, err := h.usecase.ListProducts(c.Request.Context(), page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list products")
		return
	}

	total, err := h.usecase.CountProducts(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count products")
		return
	}

	response.Success(c, ToProductListResponse(products, total, page, pageSize))
}

// ListByCategory retrieves products by category.
// @Summary List products by category
// @Description Get paginated list of products filtered by category
// @Tags Products
// @Accept json
// @Produce json
// @Param category path string true "Category name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=ProductListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/category/{category} [get]
func (h *ProductHandler) ListByCategory(c *gin.Context) {
	category := c.Param("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, err := h.usecase.ListProductsByCategory(c.Request.Context(), category, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list products")
		return
	}

	total, err := h.usecase.CountProducts(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count products")
		return
	}

	response.Success(c, ToProductListResponse(products, total, page, pageSize))
}

// ListByBrand retrieves products by brand.
// @Summary List products by brand
// @Description Get paginated list of products filtered by brand
// @Tags Products
// @Accept json
// @Produce json
// @Param brand path string true "Brand name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=ProductListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/brand/{brand} [get]
func (h *ProductHandler) ListByBrand(c *gin.Context) {
	brand := c.Param("brand")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, err := h.usecase.ListProductsByBrand(c.Request.Context(), brand, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list products")
		return
	}

	total, err := h.usecase.CountProducts(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count products")
		return
	}

	response.Success(c, ToProductListResponse(products, total, page, pageSize))
}

// ListByStatus retrieves products by status.
// @Summary List products by status
// @Description Get paginated list of products filtered by status
// @Tags Products
// @Accept json
// @Produce json
// @Param status path string true "Product status" Enums(active, draft, out_of_stock, discontinued)
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=ProductListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/status/{status} [get]
func (h *ProductHandler) ListByStatus(c *gin.Context) {
	status := product.ProductStatus(c.Param("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, err := h.usecase.ListProductsByStatus(c.Request.Context(), status, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list products")
		return
	}

	total, err := h.usecase.CountProducts(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count products")
		return
	}

	response.Success(c, ToProductListResponse(products, total, page, pageSize))
}

// Search searches products by query.
// @Summary Search products
// @Description Search products by name, SKU, or description
// @Tags Products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response{data=ProductListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/search [get]
func (h *ProductHandler) Search(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if query == "" {
		response.ErrorResponse(c, http.StatusBadRequest, "EMPTY_QUERY", "Search query is required")
		return
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, err := h.usecase.SearchProducts(c.Request.Context(), query, page, pageSize)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "SEARCH_FAILED", "Failed to search products")
		return
	}

	total, err := h.usecase.CountProducts(c.Request.Context())
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "COUNT_FAILED", "Failed to count products")
		return
	}

	response.Success(c, ToProductListResponse(products, total, page, pageSize))
}

// Activate activates a product.
// @Summary Activate product
// @Description Transition product to Active status
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id}/activate [post]
func (h *ProductHandler) Activate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	if err := h.usecase.ActivateProduct(c.Request.Context(), id); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		if errors.Is(err, product.ErrProductAlreadyActive) {
			response.ErrorResponse(c, http.StatusConflict, "ALREADY_ACTIVE", "Product is already active")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "ACTIVATION_FAILED", "Failed to activate product")
		return
	}

	// Retrieve updated product
	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// Deactivate deactivates a product.
// @Summary Deactivate product
// @Description Transition product to OutOfStock status
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id}/deactivate [post]
func (h *ProductHandler) Deactivate(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	if err := h.usecase.DeactivateProduct(c.Request.Context(), id); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		if errors.Is(err, product.ErrProductAlreadyInactive) {
			response.ErrorResponse(c, http.StatusConflict, "ALREADY_INACTIVE", "Product is already inactive")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DEACTIVATION_FAILED", "Failed to deactivate product")
		return
	}

	// Retrieve updated product
	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// Discontinue discontinues a product.
// @Summary Discontinue product
// @Description Mark product as discontinued
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id}/discontinue [post]
func (h *ProductHandler) Discontinue(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	if err := h.usecase.DiscontinueProduct(c.Request.Context(), id); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "DISCONTINUE_FAILED", "Failed to discontinue product")
		return
	}

	// Retrieve updated product
	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// UpdateInventorySettings updates inventory tracking settings.
// @Summary Update inventory settings
// @Description Update inventory tracking and backorder settings
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param request body UpdateInventorySettingsRequest true "Inventory settings"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id}/inventory-settings [put]
func (h *ProductHandler) UpdateInventorySettings(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	var req UpdateInventorySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.UpdateInventorySettings(c.Request.Context(), id, req.TrackInventory, req.AllowBackorder); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update product")
		return
	}

	// Retrieve updated product
	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// SetReorderPoint sets reorder point and quantity.
// @Summary Set reorder point
// @Description Set reorder point and reorder quantity for inventory management
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param request body SetReorderPointRequest true "Reorder settings"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id}/reorder-point [put]
func (h *ProductHandler) SetReorderPoint(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	var req SetReorderPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.usecase.SetReorderPoint(c.Request.Context(), id, req.ReorderPoint, req.ReorderQuantity); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		if errors.Is(err, product.ErrProductInvalidReorder) {
			response.ErrorResponse(c, http.StatusBadRequest, "INVALID_REORDER", "Invalid reorder settings")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update product")
		return
	}

	// Retrieve updated product
	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}

// SetPhysicalProperties sets physical properties.
// @Summary Set physical properties
// @Description Set weight and dimensions for physical product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param request body SetPhysicalPropertiesRequest true "Physical properties"
// @Success 200 {object} response.Response{data=ProductResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /warehouse/products/{id}/physical-properties [put]
func (h *ProductHandler) SetPhysicalProperties(c *gin.Context) {
	id, err := uuidv7.Parse(c.Param("id"))
	if err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid product ID format")
		return
	}

	var req SetPhysicalPropertiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Create dimensions struct
	dimensions := product.Dimensions{
		Length: req.Length,
		Width:  req.Width,
		Height: req.Height,
	}

	if err := h.usecase.SetPhysicalProperties(c.Request.Context(), id, req.Weight, dimensions); err != nil {
		if errors.Is(err, product.ErrProductNotFound) {
			response.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Product not found")
			return
		}
		if errors.Is(err, product.ErrProductInvalidWeight) || errors.Is(err, product.ErrProductInvalidDimension) {
			response.ErrorResponse(c, http.StatusBadRequest, "INVALID_PROPERTIES", "Invalid physical properties")
			return
		}
		response.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update product")
		return
	}

	// Retrieve updated product
	p, err := h.usecase.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(c, http.StatusInternalServerError, "RETRIEVAL_FAILED", "Failed to retrieve product")
		return
	}

	response.Success(c, ToProductResponse(p))
}
