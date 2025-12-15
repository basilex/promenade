package dto

import (
    "time"
    "github.com/google/uuid"
    "github.com/basilex/promenade/internal/domain/entity"
)

type ProductResponse struct {
    ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
    Name      string    `json:"name" example:"Sample Product"`
    Active    bool      `json:"active" example:"true"`
    CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

type CreateProductRequest struct {
    Name string `json:"name" binding:"required,min=2,max=255" example:"Sample Product"`
}

type UpdateProductRequest struct {
    Name   string `json:"name" binding:"omitempty,min=2,max=255" example:"Updated Product"`
    Active *bool  `json:"active" example:"true"`
}

type ListProductsResponse struct {
    Products []*ProductResponse `json:"products"`
    Total   int                  `json:"total" example:"100"`
    Page    int                  `json:"page" example:"1"`
}

func ToProductResponse(product *entity.Product) *ProductResponse {
    return &ProductResponse{
        ID:        product.ID,
        Name:      product.Name,
        Active:    product.Active,
        CreatedAt: product.CreatedAt,
    }
}
