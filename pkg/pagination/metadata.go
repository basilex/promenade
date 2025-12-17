package pagination

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
