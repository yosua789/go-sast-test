package dto

type Pagination struct {
	TotalRecords int64  `json:"total_records"`
	MaxPage      int64  `json:"max_page"`
	CurrentPage  int64  `json:"current_page"`
	PrevPage     *int64 `json:"prev_page"`
	NextPage     *int64 `json:"next_page"`
}

type PaginationParam struct {
	TargetPage int64 `form:"page" validate:"omitempty,gte=1"`
}
