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
