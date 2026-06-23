package server

import "strconv"

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
	BaseUrl    string
	Target     string
	Page       int
	PageSize   int
	TotalItems int
	TotalPages int
	HasPrev    bool
	HasNext    bool
	PrevPage   int
	NextPage   int
}

type pagingDataPage struct {
	PageNum   int
	IsCurrent bool
}

func newPaging(baseUrl, target string, page, pageSize, totalItems int) *pagingData {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	pageSize = min(pageSize, maxPageSize)
	totalPages := max((totalItems+pageSize-1)/pageSize, 1)
	page = max(1, min(page, totalPages))
	return &pagingData{
		BaseUrl:    baseUrl,
		Target:     target,
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

func (p *pagingData) Pages() []pagingDataPage {
	pages := make([]pagingDataPage, 0, p.TotalPages)
	for page := 1; page <= p.TotalPages; page++ {
		pages = append(pages, pagingDataPage{
			PageNum:   page,
			IsCurrent: page == p.Page,
		})
	}
	return pages
}

func (p *pagingData) PageUrl(pageNum int) string {
	return p.BaseUrl + "?" + p.pageQueryPart(pageNum)
}

func (p *pagingData) pageQueryPart(pageNum int) string {
	query := "page=" + strconv.Itoa(pageNum)
	if p.PageSize != defaultPageSize {
		query = "page-size=" + strconv.Itoa(p.PageSize) + "&" + query
	}
	return query
}
