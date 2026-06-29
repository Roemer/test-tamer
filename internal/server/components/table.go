package components

import "html/template"

type Table struct {
	Columns []TableColumn
	Rows    []TableRow
}

func (t Table) HasAnyActions() bool {
	for _, row := range t.Rows {
		if len(row.Actions) > 0 {
			return true
		}
	}
	return false
}

type TableColumn struct {
	Header string
}

type TableRow struct {
	Cells   []TableCell
	Actions []TableAction
}

type TableCell struct {
	Value string
}

type TableAction struct {
	Label string
	Class string
	Attrs template.HTMLAttr
}
