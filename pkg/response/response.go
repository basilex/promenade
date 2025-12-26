package response

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp in seconds
}

type PaginatedResponse struct {
	Items      any   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int   `json:"total"`
	TotalPages int   `json:"total_pages"`
	Timestamp  int64 `json:"timestamp"` // Unix timestamp in seconds
}

func Success(c *gin.Context, code int, data any) {
	c.JSON(code, Response{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

type ErrorResponse struct {
	Success   bool   `json:"success" example:"false"`
	Message   string `json:"message" example:"Error message"`
	Error     string `json:"error" example:"Detailed error"`
	Timestamp int64  `json:"timestamp" example:"1766124866"`
}

type SuccessResponse struct {
	Success   bool   `json:"success" example:"true"`
	Message   string `json:"message" example:"Operation completed successfully"`
	Data      any    `json:"data,omitempty"`
	Timestamp int64  `json:"timestamp" example:"1766124866"`
}

func Error(c *gin.Context, code int, message string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	c.JSON(code, Response{
		Success:   false,
		Message:   message,
		Error:     errMsg,
		Timestamp: time.Now().Unix(),
	})
}

// GetPageFromQuery gets page number from query params
func GetPageFromQuery(c *gin.Context) int {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	return page
}

// GetPageSizeFromQuery gets page size from query params
func GetPageSizeFromQuery(c *gin.Context) int {
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageSize
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(items any, page, pageSize, total int) PaginatedResponse {
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return PaginatedResponse{
		Items:      items,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		Timestamp:  time.Now().Unix(),
	}
}


// SuccessWithPagination sends a paginated success response
func SuccessWithPagination(c *gin.Context, code int, items any, total, page, pageSize int) {
	response := NewPaginatedResponse(items, page, pageSize, total)
	c.JSON(code, response)
}
