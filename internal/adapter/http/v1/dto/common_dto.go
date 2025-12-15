package dto

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
    Page       int `json:"page" example:"1"`
    Limit      int `json:"limit" example:"10"`
    Total      int `json:"total" example:"100"`
    TotalPages int `json:"total_pages" example:"10"`
}
