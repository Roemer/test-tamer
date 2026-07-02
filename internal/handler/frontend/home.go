package frontend

import (
	"net/http"

	"github.com/roemer/test-tamer/internal/component"
)

func (h *FrontendHandler) Home(w http.ResponseWriter, r *http.Request) {
	type pageData struct {
		Breadcrumbs component.Breadcrumbs
	}
	data := pageData{
		Breadcrumbs: component.NewBreadcrumbs(),
	}
	h.renderPage(w, r, "page/home.html", data)
}
