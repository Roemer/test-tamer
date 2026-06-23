package server

import (
	"strconv"
	"strings"
)

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
	BaseUrl           string
	Target            string
	IndicatorID       string
	IndicatorSelector string
	Page              int
	PageSize          int
	TotalItems        int
	TotalPages        int
	Items             []pagingItem
	HasPrev           bool
	HasNext           bool
	PrevPage          int
	NextPage          int
}

type pagingItem struct {
	PageNum   int
	Label     string
	IsCurrent bool
	IsGap     bool
}

func newPaging(baseUrl, target string, page, pageSize, totalItems int) *pagingData {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	pageSize = min(pageSize, maxPageSize)
	indicatorID := strings.TrimPrefix(target, "#") + "-loading"
	totalPages := max((totalItems+pageSize-1)/pageSize, 1)
	page = max(1, min(page, totalPages))
	return &pagingData{
		BaseUrl:           baseUrl,
		Target:            target,
		IndicatorID:       indicatorID,
		IndicatorSelector: "#" + indicatorID,
		Page:              page,
		PageSize:          pageSize,
		TotalItems:        totalItems,
		TotalPages:        totalPages,
		Items:             buildPagingItems(page, totalPages),
		HasPrev:           page > 1,
		HasNext:           page < totalPages,
		PrevPage:          max(1, page-1),
		NextPage:          min(totalPages, page+1),
	}
}

func buildPagingItems(page, totalPages int) []pagingItem {
	const maxPagesToShow = 9

	// Special case: all pages fit without any ellipsis
	if totalPages <= maxPagesToShow {
		items := make([]pagingItem, 0, totalPages)
		for i := 1; i <= totalPages; i++ {
			items = append(items, pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page})
		}
		return items
	}

	// From here on totalPages > maxPagesToShow, so we always have exactly maxPagesToShow items:
	//   [0]        = first page (always 1)
	//   [1..7]     = dynamic inner region (7 slots)
	//   [8]        = last page (always totalPages)
	//
	// The inner 7 slots are filled in one of three ways:
	//   Left:   pages 2..7,  gap,  _          → 1 [2 3 4 5 6 7] [...] [last]
	//   Right:  _,  gap,  pages last-6..last-1 → [1] [...] [last-6 .. last-1] last
	//   Middle: gap, page-2..page+2, gap       → [1] [...] [p-2 p-1 p p+1 p+2] [...] [last]
	//
	// Derived constants (all from maxPagesToShow = 9):
	//   innerSlots      = maxPagesToShow - 2  = 7   (slots between first and last)
	//   sidePageCount   = innerSlots - 1      = 6   (pages shown in left/right mode)
	//   middlePageCount = innerSlots - 2      = 5   (pages shown in middle mode)
	//   middleHalf      = middlePageCount / 2 = 2   (pages on each side of current)
	//   leftThreshold   = sidePageCount / 2 + 1 = 4 (current <= this → left mode; current is within the left window)
	//   rightThreshold  = totalPages - sidePageCount/2 = totalPages-3 (current >= this → right mode)

	innerSlots := maxPagesToShow - 2
	sidePageCount := innerSlots - 1
	middleHalf := (innerSlots - 2) / 2

	leftThreshold := sidePageCount/2 + 1
	rightThreshold := totalPages - sidePageCount/2

	// Prepare the items slice
	items := make([]pagingItem, maxPagesToShow)

	// Add the first and last
	items[0] = pagingItem{PageNum: 1, Label: "1", IsCurrent: page == 1}
	items[maxPagesToShow-1] = pagingItem{PageNum: totalPages, Label: strconv.Itoa(totalPages), IsCurrent: page == totalPages}

	gap := pagingItem{Label: "...", IsGap: true}

	switch {
	case page <= leftThreshold:
		// Left mode: show pages 2..sidePageCount+1, then gap.
		for i := 2; i <= sidePageCount+1; i++ {
			items[i-1] = pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page}
		}
		items[maxPagesToShow-2] = gap

	case page >= rightThreshold:
		// Right mode: gap, then show pages totalPages-sidePageCount..totalPages-1.
		items[1] = gap
		for i := totalPages - sidePageCount; i <= totalPages-1; i++ {
			items[i-(totalPages-sidePageCount)+2] = pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page}
		}

	default:
		// Middle mode: gap, middleHalf pages left of current, current, middleHalf pages right, gap.
		items[1] = gap
		items[maxPagesToShow-2] = gap
		for i := page - middleHalf; i <= page+middleHalf; i++ {
			items[i-(page-middleHalf)+2] = pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page}
		}
	}

	return items
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
