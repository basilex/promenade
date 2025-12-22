package dto

import (
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
)

// LanguageResponse represents language response
type LanguageResponse struct {
	ID         uuidv7.UUID `json:"id"`
	Name       string      `json:"name" example:"English"`
	NativeName string      `json:"native_name" example:"English"`
	Code       string      `json:"code" example:"en"`
	ISO639_2   string      `json:"iso639_2" example:"eng"`
	IsRtl      bool        `json:"is_rtl" example:"false"`
	IsActive   bool        `json:"is_active" example:"true"`
	SortOrder  int         `json:"sort_order" example:"1"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// CreateLanguageRequest represents create language request
type CreateLanguageRequest struct {
	Name       string `json:"name" binding:"required,min=2,max=100" example:"English"`
	NativeName string `json:"native_name" binding:"required,min=2,max=100" example:"English"`
	Code       string `json:"code" binding:"required,len=2,alpha,lowercase" example:"en"`
	ISO639_2   string `json:"iso639_2" binding:"required,len=3,alpha,lowercase" example:"eng"`
	IsRtl      bool   `json:"is_rtl" example:"false"`
	IsActive   bool   `json:"is_active" example:"true"`
	SortOrder  int    `json:"sort_order" binding:"omitempty,min=0" example:"1"`
}

// UpdateLanguageRequest represents update language request
type UpdateLanguageRequest struct {
	Name       string `json:"name" binding:"required,min=2,max=100" example:"English"`
	NativeName string `json:"native_name" binding:"required,min=2,max=100" example:"English"`
	Code       string `json:"code" binding:"required,len=2,alpha,lowercase" example:"en"`
	ISO639_2   string `json:"iso639_2" binding:"required,len=3,alpha,lowercase" example:"eng"`
	IsRtl      bool   `json:"is_rtl" example:"false"`
	IsActive   bool   `json:"is_active" example:"true"`
	SortOrder  int    `json:"sort_order" binding:"omitempty,min=0" example:"1"`
}

// LanguageListResponse represents paginated languages response
type LanguageListResponse struct {
	Languages []*LanguageResponse `json:"languages"`
	Total     int                 `json:"total"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
}
