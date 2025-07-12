package domain

type PaginationParam struct {
	TargetPage int64 // Target page
	// SortBy     string
	Order string // ASC | DESC
}
