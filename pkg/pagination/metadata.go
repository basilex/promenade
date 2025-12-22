package pagination

// Params represents pagination parameters
type Params struct {
	Page     int `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=100"`
	Limit    int `json:"limit" form:"limit" binding:"omitempty,min=1,max=100"`
	Offset   int `json:"offset" form:"offset" binding:"omitempty,min=0"`
}

// Metadata represents pagination metadata
type Metadata struct {
	Total       int  `json:"total"`
	Limit       int  `json:"limit"`
	Offset      int  `json:"offset"`
	TotalPages  int  `json:"total_pages"`
	CurrentPage int  `json:"current_page"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

// NewMetadata creates pagination metadata
func NewMetadata(total, limit, offset int) *Metadata {
	if limit == 0 {
		limit = 10 // default
	}

	currentPage := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	if totalPages == 0 {
		totalPages = 1
	}

	return &Metadata{
		Total:       total,
		Limit:       limit,
		Offset:      offset,
		TotalPages:  totalPages,
		CurrentPage: currentPage,
		HasNext:     currentPage < totalPages,
		HasPrev:     currentPage > 1,
	}
}

// GetOffset calculates offset from Page and PageSize
func (p Params) GetOffset() int {
	if p.Page < 1 {
		return 0
	}
	if p.PageSize < 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size or limit
func (p Params) GetLimit() int {
	if p.PageSize > 0 {
		return p.PageSize
	}
	if p.Limit > 0 {
		return p.Limit
	}
	return 20 // default
}
