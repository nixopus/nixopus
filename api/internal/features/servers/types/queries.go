package types

type ServerQueryParams struct {
	Page      int    `query:"page"`
	PageSize  int    `query:"page_size"`
	Search    string `query:"search"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

type ServerListResponse struct {
	Servers    interface{} `json:"servers"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	CurrentPage int  `json:"current_page"`
	PageSize    int  `json:"page_size"`
	TotalPages  int  `json:"total_pages"`
	TotalItems  int  `json:"total_items"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

func (q *ServerQueryParams) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 || q.PageSize > 100 {
		q.PageSize = 10
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortOrder != "asc" && q.SortOrder != "desc" {
		q.SortOrder = "desc"
	}
}

func (q *ServerQueryParams) GetOffset() int {
	return (q.Page - 1) * q.PageSize
}

func (q *ServerQueryParams) GetLimit() int {
	return q.PageSize
}

func ValidSortFields() []string {
	return []string{"name", "host", "port", "username", "created_at", "updated_at"}
}

func (q *ServerQueryParams) IsValidSortField() bool {
	validFields := ValidSortFields()
	for _, field := range validFields {
		if field == q.SortBy {
			return true
		}
	}
	return false
}
