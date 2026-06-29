package server

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

// HTMXAttrs generates a string of HTML attributes for HTMX requests.
func HTMXAttrs(attrs map[string]string) template.HTMLAttr {
	var b strings.Builder
	for k, v := range attrs {
		fmt.Fprintf(&b, `%s="%s" `,
			html.EscapeString(k),
			html.EscapeString(v),
		)
	}
	return template.HTMLAttr(b.String())
}
