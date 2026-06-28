package server

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

type tableComponent struct {
	Columns []tableComponentColumn
	Rows    []tableComponentRow
}

func (t tableComponent) HasAnyActions() bool {
	for _, row := range t.Rows {
		if len(row.Actions) > 0 {
			return true
		}
	}
	return false
}

type tableComponentColumn struct {
	Header string
}

type tableComponentRow struct {
	Cells   []tableComponentCell
	Actions []tableComponentAction
}

type tableComponentCell struct {
	Value string
}

type tableComponentAction struct {
	Label string
	Class string
	Attrs template.HTMLAttr
}

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
