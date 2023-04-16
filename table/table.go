package table

import (
	"fmt"
	"strconv"
	"strings"
)

type Table struct {
	Header []string
	Rows   []Row
	Width  int
	Fg     int
	Bg     int
}

type Style struct {
	Fg int
	Bg int
}

type Row struct {
	Cells []string
	Style Style
}

func New(header []string, defStyle Style) *Table {
	t := Table{
		Fg: defStyle.Fg,
		Bg: defStyle.Bg,
	}

	t.AddRow(Row{
		Cells: header,
		Style: Style{Fg: t.Fg, Bg: t.Bg},
	})

	return &t
}

func NewRow() *Row {
	return &Row{}
}

func (t *Table) AddRow(row Row) {
	t.Rows = append(t.Rows, row)
}

func (t *Table) Render() {

	colWidths := make([]int, len(t.Rows[0].Cells))
	for _, row := range t.Rows {
		for j, cell := range row.Cells {
			if colWidths[j] < len(cell) {
				colWidths[j] = len(cell)
			}
		}
	}

	// TODO: resizing columns to fit screen width

	for _, r := range t.Rows {
		if r.Style.Fg == 0 {
			r.Style.Fg = t.Fg
		}
		if r.Style.Bg == 0 {
			r.Style.Bg = t.Bg
		}

		cells := r.Cells
		for i, w := range colWidths {
			cells[i] = fmt.Sprintf("%-"+strconv.Itoa(w)+"v", cells[i])
		}

		line := strings.Join(r.Cells, strings.Repeat(" ", 2))
		fmt.Printf("\033[%d;38;5;%d;48;5;%dm%s\033[0m\n", 0, r.Style.Fg, r.Style.Bg, line)
	}
}
