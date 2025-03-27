package dto

type Pagination struct {
	Page       int `json:"current_page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
	TotalData  int `json:"total_data,omitempty"`
}
