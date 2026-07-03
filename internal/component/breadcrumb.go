package component

type Breadcrumbs struct {
	Items []BreadcrumbItem
}

func NewBreadcrumbs(items ...BreadcrumbItem) Breadcrumbs {
	lastIndex := len(items) - 1
	for i := range items {
		items[i].Active = i == lastIndex
	}
	return Breadcrumbs{
		Items: items,
	}
}

type BreadcrumbItem struct {
	Name   string
	Link   string
	Active bool
}

func NewBreadcrumbItem(name, link string) BreadcrumbItem {
	return BreadcrumbItem{
		Name: name,
		Link: link,
	}
}
