package entity

type Pagination struct {
	TotalRecords    int64
	Page            int64
	TotalPage       int64
	NextPage        int64
	PreviousPage    int64
	HasPreviousPage bool
	HasNextPage     bool
	SortBy          string
	Order           string
}
