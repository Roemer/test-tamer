package server

type tableComponent struct {
	Columns []tableComponentColumn
	Rows    []tableComponentRow
}

type tableComponentColumn struct {
	Header string
}

type tableComponentRow struct {
	Cells []tableComponentCell
}

type tableComponentCell struct {
	Value string
}
