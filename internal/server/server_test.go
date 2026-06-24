package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagingDefaultSize(t *testing.T) {
	assert := assert.New(t)

	paging := newPaging("/projects", "#project-list", 1, defaultPageSize, 50)

	assert.Equal("page=1", paging.pageQueryPart(paging.Page))
	assert.Equal("page=2", paging.pageQueryPart(paging.NextPage))
}

func TestPagingNoDefaultSize(t *testing.T) {
	assert := assert.New(t)

	paging := newPaging("/projects", "#project-list", 1, defaultPageSize*2, 50)

	assert.Equal("page-size=20&page=1", paging.pageQueryPart(paging.Page))
	assert.Equal("page-size=20&page=2", paging.pageQueryPart(paging.NextPage))
}

func TestPagingIndicatorFromTargetSelector(t *testing.T) {
	assert := assert.New(t)

	paging := newPaging("/projects", "#project-list", 1, defaultPageSize, 50)

	assert.Equal("project-list-loading", paging.IndicatorID)
	assert.Equal("#project-list-loading", paging.IndicatorSelector)
}

func TestPagingItems(t *testing.T) {
	assert := assert.New(t)

	// First page
	items := buildPagingItems(1, 20)
	assert.Equal(9, len(items))
	assert.Equal(1, items[0].PageNum)
	assert.Equal(2, items[1].PageNum)
	assert.Equal(3, items[2].PageNum)
	assert.Equal(4, items[3].PageNum)
	assert.Equal(5, items[4].PageNum)
	assert.Equal(6, items[5].PageNum)
	assert.Equal(7, items[6].PageNum)
	assert.Equal(true, items[7].IsGap)
	assert.Equal(20, items[8].PageNum)

	// Last page
	items = buildPagingItems(20, 20)
	assert.Equal(9, len(items))
	assert.Equal(1, items[0].PageNum)
	assert.Equal(true, items[1].IsGap)
	assert.Equal(14, items[2].PageNum)
	assert.Equal(15, items[3].PageNum)
	assert.Equal(16, items[4].PageNum)
	assert.Equal(17, items[5].PageNum)
	assert.Equal(18, items[6].PageNum)
	assert.Equal(19, items[7].PageNum)
	assert.Equal(20, items[8].PageNum)
}
