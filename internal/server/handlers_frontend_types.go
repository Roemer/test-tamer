package server

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

type breadcrumbItem struct {
	Name   string
	Link   string
	Active bool
}

type pagingData struct {
	Page       int
	PageSize   int
	TotalItems int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	PrevPage   int
	NextPage   int
}

func newPaging(page, pageSize, totalItems int) *pagingData {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	pageSize = min(pageSize, maxPageSize)
	totalPages := max((totalItems+pageSize-1)/pageSize, 1)
	page = max(1, min(page, totalPages))
	return &pagingData{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
		PrevPage:   max(1, page-1),
		NextPage:   min(totalPages, page+1),
	}
}

func (p *pagingData) Pages() []int {
	pages := make([]int, 0, p.TotalPages)
	for page := 1; page <= p.TotalPages; page++ {
		pages = append(pages, page)
	}
	return pages
}
