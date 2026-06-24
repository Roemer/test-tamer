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
	const maxPagesToShow = 9 // Needs to be uneven!

	// Special case when all pages can be shown without gaps
	if totalPages <= maxPagesToShow {
		items := make([]pagingItem, 0, totalPages)
		for i := 1; i <= totalPages; i++ {
			items = append(items, pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page})
		}
		return items
	}

	// There are more pages that can be shown so we will always show the max pages in one of 3 ways:
	// Left:   1 2 3 4 5 6 7 ... last
	// Right:  1 ... last-6 last-5 last-4 last-3 last-2 last-1 last
	// Middle: 1 ... p-2 p-1 p p+1 p+2 ... last

	// Prepare a slice for the items
	items := make([]pagingItem, maxPagesToShow)

	// Always add the first and last page
	items[0] = pagingItem{PageNum: 1, Label: "1", IsCurrent: page == 1}
	items[maxPagesToShow-1] = pagingItem{PageNum: totalPages, Label: strconv.Itoa(totalPages), IsCurrent: page == totalPages}

	// Create a gap item which can be used left and/or right
	gap := pagingItem{Label: "...", IsGap: true}

	// Define the number of inner pages that are dynamic
	innerSlots := maxPagesToShow - 2

	// Define the thresholds for left or right mode
	leftThreshold := (innerSlots + 1) / 2 // Ceiling of half the inner slots
	rightThreshold := totalPages - leftThreshold + 1

	// Handle the different modes
	switch {
	case page <= leftThreshold:
		// Left mode: show pages 2..innerSlots, then gap.
		for i := 2; i <= innerSlots; i++ {
			items[i-1] = pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page}
		}
		items[maxPagesToShow-2] = gap

	case page >= rightThreshold:
		// Right mode: gap, then show the last pages until totalPages-1.
		items[1] = gap
		startPage := totalPages - (innerSlots - 1)
		for i := startPage; i <= totalPages-1; i++ {
			items[i-startPage+2] = pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page}
		}

	default:
		// Middle mode: gap, middleHalf pages left of current, current, middleHalf pages right, gap.
		items[1] = gap
		items[maxPagesToShow-2] = gap
		numItems := innerSlots - 2 // Number of items to show in the middle (excluding gaps)
		halfItems := numItems / 2
		for i := page - halfItems; i <= page+halfItems; i++ {
			items[i-(page-halfItems)+2] = pagingItem{PageNum: i, Label: strconv.Itoa(i), IsCurrent: i == page}
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
