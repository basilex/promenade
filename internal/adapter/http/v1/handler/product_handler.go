package handler

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/basilex/promenade/internal/adapter/http/shared/response"
    "github.com/basilex/promenade/internal/adapter/http/v1/dto"
    "github.com/basilex/promenade/internal/usecase"
)

type ProductHandler struct {
    productUseCase *usecase.ProductUseCase
}

func NewProductHandler(productUseCase *usecase.ProductUseCase) *ProductHandler {
    return &ProductHandler{productUseCase: productUseCase}
}

func (h *ProductHandler) Create(c *gin.Context) {
    var req dto.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid request", err)
        return
    }

    product, err := h.productUseCase.CreateProduct(c.Request.Context(), req.Name)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "failed to create product", err)
        return
    }

    response.Success(c, http.StatusCreated, dto.ToProductResponse(product))
}

func (h *ProductHandler) GetByID(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "invalid product id", err)
        return
    }

    product, err := h.productUseCase.GetProduct(c.Request.Context(), id)
    if err != nil {
        response.Error(c, http.StatusNotFound, "product not found", err)
        return
    }

    response.Success(c, http.StatusOK, dto.ToProductResponse(product))
}

func (h *ProductHandler) List(c *gin.Context) {
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    offset := (page - 1) * limit

    products, total, err := h.productUseCase.ListProducts(c.Request.Context(), limit, offset)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "failed to list products", err)
        return
    }

    productResponses := make([]*dto.ProductResponse, len(products))
    for i, product := range products {
        productResponses[i] = dto.ToProductResponse(product)
    }

    resp := &dto.ListProductsResponse{
        Products:   productResponses,
        Total:   total,
        Page:    page,
    }

    response.Success(c, http.StatusOK, resp)
}

func (h *ProductHandler) Update(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "invalid product id", err)
        return
    }

    var req dto.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid request", err)
        return
    }

    product, err := h.productUseCase.GetProduct(c.Request.Context(), id)
    if err != nil {
        response.Error(c, http.StatusNotFound, "product not found", err)
        return
    }

    if req.Name != "" {
        product.Name = req.Name
    }
    if req.Active != nil {
        product.Active = *req.Active
    }

    if err := h.productUseCase.UpdateProduct(c.Request.Context(), product); err != nil {
        response.Error(c, http.StatusInternalServerError, "failed to update product", err)
        return
    }

    response.Success(c, http.StatusOK, dto.ToProductResponse(product))
}

func (h *ProductHandler) Delete(c *gin.Context) {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        response.Error(c, http.StatusBadRequest, "invalid product id", err)
        return
    }

    if err := h.productUseCase.DeleteProduct(c.Request.Context(), id); err != nil {
        response.Error(c, http.StatusInternalServerError, "failed to delete product", err)
        return
    }

    c.Status(http.StatusNoContent)
}
